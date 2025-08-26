package main

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"shipshipship/database"
	"shipshipship/handlers"
	"shipshipship/middleware"
	"shipshipship/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Custom static file handler with proper MIME types
func serveStaticFile(buildDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		fullPath := filepath.Join(buildDir, path)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.Status(404)
			return
		}

		// Set proper MIME type based on file extension
		ext := filepath.Ext(fullPath)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			switch ext {
			case ".js":
				contentType = "application/javascript"
			case ".css":
				contentType = "text/css"
			case ".html":
				contentType = "text/html; charset=utf-8"
			case ".json":
				contentType = "application/json"
			case ".svg":
				contentType = "image/svg+xml"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			default:
				contentType = "application/octet-stream"
			}
		}

		c.Header("Content-Type", contentType)
		c.File(fullPath)
	}
}

func main() {
	// Initialize database
	database.InitDatabase()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	r.Use(cors.New(config))

	// Public routes
	api := r.Group("/api")
	{
		api.GET("/events", handlers.GetEvents)
		api.GET("/events/:id", handlers.GetEvent)
		api.GET("/events/slug/:slug", handlers.GetEventBySlug)
		api.POST("/events/:id/vote", handlers.VoteEvent)
		api.GET("/events/:id/vote-status", handlers.CheckVoteStatus)
		api.POST("/feedback", middleware.FeedbackRateLimit(), handlers.SubmitFeedback)
		api.POST("/auth/login", handlers.Login)
		api.GET("/settings", handlers.GetSettings)

		// Tag routes (public)
		api.GET("/tags", handlers.GetTags)

		// Newsletter routes
		api.POST("/newsletter/subscribe", handlers.SubscribeToNewsletter)
		api.POST("/newsletter/unsubscribe", handlers.UnsubscribeFromNewsletter)
		api.GET("/newsletter/status", handlers.CheckSubscriptionStatus)

		// Footer links routes (public read access)
		api.GET("/footer-links", handlers.GetFooterLinks)
		api.GET("/footer-links/by-column", handlers.GetFooterLinksByColumn)
	}

	// Protected admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/validate", handlers.ValidateToken)
		admin.GET("/events", handlers.GetAllEvents)
		admin.POST("/events", handlers.CreateEvent)
		admin.PUT("/events/:id", handlers.UpdateEvent)
		admin.DELETE("/events/:id", handlers.DeleteEvent)
		admin.POST("/events/reorder", handlers.ReorderEvents)
		admin.PUT("/settings", handlers.UpdateSettings)
		admin.POST("/upload/image", handlers.UploadImage)

		// Tag admin routes
		admin.GET("/tags", handlers.GetTags)
		admin.GET("/tags/usage", handlers.GetTagUsage)
		admin.GET("/tags/:id", handlers.GetTag)
		admin.POST("/tags", handlers.CreateTag)
		admin.PUT("/tags/:id", handlers.UpdateTag)
		admin.DELETE("/tags/:id", handlers.DeleteTag)

		// Mail settings routes
		admin.GET("/settings/mail", handlers.GetMailSettings)
		admin.POST("/settings/mail", handlers.UpdateMailSettings)
		admin.POST("/settings/mail/test", handlers.TestMailSettings)

		// Newsletter admin routes
		admin.GET("/newsletter/stats", handlers.GetNewsletterStats)
		admin.GET("/newsletter/subscribers", handlers.GetNewsletterSubscribers)
		admin.GET("/newsletter/subscribers/paginated", handlers.GetNewsletterSubscribersPaginated)
		admin.DELETE("/newsletter/subscribers/:email", handlers.DeleteNewsletterSubscriber)
		admin.GET("/newsletter/history", handlers.GetNewsletterHistory)
		admin.GET("/newsletter/templates", handlers.GetEmailTemplates)
		admin.PUT("/newsletter/templates", handlers.UpdateEmailTemplates)
		admin.GET("/newsletter/automation", handlers.GetNewsletterAutomationSettings)
		admin.PUT("/newsletter/automation", handlers.UpdateNewsletterAutomationSettings)

		// Event publishing routes
		admin.GET("/events/:id/publish", handlers.GetEventPublishStatus)
		admin.PUT("/events/:id/publish", handlers.UpdateEventPublicStatus)
		admin.GET("/events/:id/newsletter/preview", handlers.GetEventNewsletterPreview)
		admin.POST("/events/:id/newsletter/send", handlers.SendEventNewsletter)
		admin.GET("/events/:id/newsletter/history", handlers.GetEventEmailHistory)

		// Footer links admin routes
		admin.GET("/footer-links", handlers.GetFooterLinks)
		admin.GET("/footer-links/:id", handlers.GetFooterLink)
		admin.POST("/footer-links", handlers.CreateFooterLink)
		admin.PUT("/footer-links/:id", handlers.UpdateFooterLink)
		admin.DELETE("/footer-links/:id", handlers.DeleteFooterLink)
		admin.POST("/footer-links/reorder", handlers.ReorderFooterLinks)
	}

	// Public file serving route
	api.GET("/uploads/:filename", handlers.ServeUploadedFile)

	// Serve static files with proper MIME types
	r.GET("/_app/*filepath", serveStaticFile("./frontend/build"))
	r.GET("/assets/*filepath", serveStaticFile("./frontend/build"))

	r.GET("/favicon.ico", func(c *gin.Context) {
		// Try to get favicon from database settings
		settings, err := models.GetOrCreateSettings(database.GetDB())
		if err == nil && settings.FaviconURL != "" {
			// Redirect to the configured favicon
			c.Redirect(http.StatusTemporaryRedirect, settings.FaviconURL)
			return
		}

		// Fallback to static favicon if no custom one is set
		c.Header("Content-Type", "image/x-icon")
		c.File("./frontend/build/favicon.ico")
	})

	// Serve SPA routes
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/build/index.html")
	})
	r.GET("/admin", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/build/index.html")
	})
	r.GET("/admin/*path", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/build/index.html")
	})

	// Fallback for SPA routing
	r.NoRoute(func(c *gin.Context) {
		// Check if it's an API route
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		// Serve the SPA for all other routes
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/build/index.html")
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
