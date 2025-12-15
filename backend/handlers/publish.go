package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shipshipship/constants"
	"shipshipship/database"
	"shipshipship/email"
	"shipshipship/models"
	"shipshipship/services"

	"github.com/gin-gonic/gin"
)

// GetEventPublishStatus gets the publication status of an event
func GetEventPublishStatus(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	db := database.GetDB()

	var event models.Event
	if err := db.Preload("Publication").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Initialize publication if it doesn't exist
	if event.Publication == nil {
		publication := &models.EventPublication{
			EventID:   uint(eventID),
			EmailSent: false,
		}
		if err := db.Create(publication).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize publication"})
			return
		}
		event.Publication = publication
	}

	c.JSON(http.StatusOK, gin.H{
		"is_public":        event.IsPublic,
		"has_public_url":   event.HasPublicUrl,
		"email_sent":       event.Publication.EmailSent,
		"email_sent_at":    event.Publication.EmailSentAt,
		"email_subject":    event.Publication.EmailSubject,
		"email_template":   event.Publication.EmailTemplate,
		"subscriber_count": event.Publication.SubscriberCount,
	})
}

// UpdateEventPublicStatus updates the public visibility of an event
func UpdateEventPublicStatus(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.EventPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	db := database.GetDB()

	// Prepare updates map
	updates := make(map[string]interface{})

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if req.HasPublicUrl != nil {
		updates["has_public_url"] = *req.HasPublicUrl
	}

	// Update the event's status
	if len(updates) > 0 {
		if err := db.Model(&models.Event{}).Where("id = ?", eventID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event status updated successfully",
		"updates": updates,
	})
}

// GetEventNewsletterPreview generates a preview of the newsletter for an event
func GetEventNewsletterPreview(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Use generic event template for all statuses
	templateType := "event"

	db := database.GetDB()

	// Get the event with tags and status definition preloaded
	var event models.Event
	if err := db.Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Get status definition for color
	var statusDef models.EventStatusDefinition
	if err := db.Where("display_name = ?", event.Status).First(&statusDef).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status definition not found"})
		return
	}

	// Get the email template or use defaults if not found
	template, err := models.GetEmailTemplate(db, templateType)
	if err != nil {
		// If template doesn't exist, use default content directly
		log.Printf("Template not found in DB, using defaults for type: %s", templateType)
		defaultTemplate := constants.GetTemplateByType(templateType)
		if defaultTemplate == nil {
			log.Printf("ERROR: No default template found for type: %s", templateType)
			// List available templates
			allTemplates := constants.GetDefaultTemplates()
			log.Printf("Available templates: %v", allTemplates)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template type"})
			return
		}

		log.Printf("Using default template: %s", defaultTemplate.Type)
		// Create a temporary template object with defaults
		template = &models.EmailTemplate{
			Type:    defaultTemplate.Type,
			Subject: defaultTemplate.Subject,
			Content: defaultTemplate.Content,
		}
	}

	// Get branding settings for project info
	// Get the base URL from request or BASE_URL env
	baseURL := getBaseURL(c, db)
	branding, err := models.GetBrandingSettingsWithBaseURL(db, baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branding settings"})
		return
	}

	// Generate the preview with variable replacements
	subject, content, err := email.GenerateEmailContent(db, template, &event, &statusDef, branding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate email content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subject": subject,
		"content": content,
	})
}

// SendEventNewsletter sends a newsletter for an event
func SendEventNewsletter(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.EventNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Use generic event template
	req.Template = "event"

	db := database.GetDB()

	// Get the event with tags and publication preloaded
	var event models.Event
	if err := db.Preload("Publication").Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Note: We allow resending emails, but track the history

	// Get newsletter subscribers
	subscribers, err := models.GetActiveNewsletterSubscribers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter subscribers"})
		return
	}

	if len(subscribers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active newsletter subscribers found"})
		return
	}

	// Get branding settings
	// Get the base URL from request or BASE_URL env
	baseURL := getBaseURL(c, db)
	branding, err := models.GetBrandingSettingsWithBaseURL(db, baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branding settings"})
		return
	}

	// Send emails to all subscribers
	emailService := services.NewEmailService()
	sentCount := 0

	for _, subscriber := range subscribers {
		// Replace unsubscribe URL in content (use BaseURL, not ProjectURL)
		unsubscribeURL := fmt.Sprintf("%s/unsubscribe?email=%s", branding.BaseURL, subscriber.Email)
		if branding.BaseURL == "" {
			unsubscribeURL = fmt.Sprintf("/unsubscribe?email=%s", subscriber.Email)
		}
		personalizedContent := strings.ReplaceAll(req.Content, "{{unsubscribe_url}}", unsubscribeURL)

		err := emailService.SendEmail(subscriber.Email, req.Subject, personalizedContent)
		if err != nil {
			// Log the error but continue sending to other subscribers
			fmt.Printf("Failed to send email to %s: %v\n", subscriber.Email, err)
			continue
		}
		sentCount++
	}

	// Create email history record
	now := time.Now()
	historyRecord := &models.EventEmailHistory{
		EventID:         uint(eventID),
		EventStatus:     string(event.Status),
		EmailSubject:    req.Subject,
		EmailTemplate:   req.Template,
		SubscriberCount: sentCount,
		SentAt:          now,
	}
	if err := db.Create(historyRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save email history"})
		return
	}

	// Newsletter history now uses EventEmailHistory directly (see newsletter.go handler)

	// Update or create publication record (for backward compatibility)
	if event.Publication == nil {
		publication := &models.EventPublication{
			EventID:         uint(eventID),
			EmailSent:       true,
			EmailSubject:    req.Subject,
			EmailContent:    req.Content,
			EmailTemplate:   req.Template,
			EmailSentAt:     &now,
			SubscriberCount: sentCount,
		}
		if err := db.Create(publication).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save publication record"})
			return
		}
	} else {
		updates := map[string]interface{}{
			"email_sent":       true,
			"email_subject":    req.Subject,
			"email_content":    req.Content,
			"email_template":   req.Template,
			"email_sent_at":    &now,
			"subscriber_count": sentCount,
		}
		if err := db.Model(event.Publication).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update publication record"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Newsletter sent successfully",
		"subscribers_sent":  sentCount,
		"total_subscribers": len(subscribers),
	})
}

// GetEventEmailHistory returns the email sending history for an event
func GetEventEmailHistory(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	db := database.GetDB()

	// Get email history for the event
	var history []models.EventEmailHistory
	if err := db.Where("event_id = ?", eventID).Order("sent_at DESC").Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}
