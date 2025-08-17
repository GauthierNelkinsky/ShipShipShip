package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	clients map[string]*ClientData
	mutex   sync.RWMutex
}

type ClientData struct {
	lastSubmission  time.Time
	submissionCount int
	resetTime       time.Time
}

var feedbackLimiter = &RateLimiter{
	clients: make(map[string]*ClientData),
}

// CleanupOldEntries removes old entries from the rate limiter
func (rl *RateLimiter) cleanupOldEntries() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for ip, data := range rl.clients {
		// Remove entries older than 24 hours
		if now.Sub(data.lastSubmission) > 24*time.Hour {
			delete(rl.clients, ip)
		}
	}
}

// Start cleanup goroutine
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				feedbackLimiter.cleanupOldEntries()
			}
		}
	}()
}

func FeedbackRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		feedbackLimiter.mutex.Lock()
		defer feedbackLimiter.mutex.Unlock()

		// Get or create client data
		clientData, exists := feedbackLimiter.clients[clientIP]
		if !exists {
			clientData = &ClientData{
				lastSubmission:  time.Time{},
				submissionCount: 0,
				resetTime:       now.Add(24 * time.Hour),
			}
			feedbackLimiter.clients[clientIP] = clientData
		}

		// Reset count if 24 hours have passed
		if now.After(clientData.resetTime) {
			clientData.submissionCount = 0
			clientData.resetTime = now.Add(24 * time.Hour)
		}

		// Basic rate limiting: 1 submission per minute
		rateLimitWindow := 60 * time.Second
		if !clientData.lastSubmission.IsZero() {
			timeSinceLastSubmission := now.Sub(clientData.lastSubmission)
			if timeSinceLastSubmission < rateLimitWindow {
				remainingSeconds := int((rateLimitWindow - timeSinceLastSubmission).Seconds())
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Rate limit exceeded. Please wait before submitting again.",
					"retry_after": remainingSeconds,
				})
				c.Abort()
				return
			}
		}

		// Progressive rate limiting for frequent submitters
		if clientData.submissionCount >= 3 {
			extendedLimit := 5 * time.Minute
			if !clientData.lastSubmission.IsZero() {
				timeSinceLastSubmission := now.Sub(clientData.lastSubmission)
				if timeSinceLastSubmission < extendedLimit {
					remainingSeconds := int((extendedLimit - timeSinceLastSubmission).Seconds())
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":       "Too many submissions. Please wait longer before submitting again.",
						"retry_after": remainingSeconds,
					})
					c.Abort()
					return
				}
			}
		}

		// Update client data
		clientData.lastSubmission = now
		clientData.submissionCount++

		c.Next()
	}
}
