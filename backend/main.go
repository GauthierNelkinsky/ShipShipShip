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

// getAdminIndexPath returns the correct path to the admin index.html file
func getAdminIndexPath() string {
	// Get the current working directory
	wd, _ := os.Getwd()

	// Check if we're running from the backend subdirectory or project root
	var projectRoot string
	if filepath.Base(wd) == "backend" {
		// Running from backend/ subdirectory
		projectRoot = filepath.Dir(wd)
	} else {
		// Running from project root
		projectRoot = wd
	}

	return filepath.Join(projectRoot, "admin", "build", "index.html")
}

// getAdminBuildPath returns the correct path to the admin build directory
func getAdminBuildPath() string {
	// Get the current working directory
	wd, _ := os.Getwd()

	// Check if we're running from the backend subdirectory or project root
	var projectRoot string
	if filepath.Base(wd) == "backend" {
		// Running from backend/ subdirectory
		projectRoot = filepath.Dir(wd)
	} else {
		// Running from project root
		projectRoot = wd
	}

	return filepath.Join(projectRoot, "admin", "build")
}

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

	// Initialize default theme if none is applied
	if err := handlers.InitializeDefaultTheme(); err != nil {
		log.Printf("Warning: Failed to initialize default theme: %v", err)
	}

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
		api.GET("/auth/demo-mode", handlers.CheckDemoMode)
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

		// Theme routes (public read access for admin interface)
		api.GET("/themes/info", handlers.GetThemeInfo)
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

		// Theme admin routes
		admin.POST("/themes/apply", handlers.ApplyTheme)
		admin.GET("/themes/current", handlers.GetCurrentTheme)
		admin.GET("/themes/info", handlers.GetThemeInfo)
	}

	// Public file serving route
	api.GET("/uploads/:filename", handlers.ServeUploadedFile)

	// Admin interface routes (register these BEFORE wildcard routes)
	r.GET("/admin", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(getAdminIndexPath())
	})

	// Admin SPA routes - handle all admin sub-routes
	r.GET("/admin/*any", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(getAdminIndexPath())
	})

	// Public theme static files - try theme first, fallback to admin
	r.GET("/_app/*filepath", func(c *gin.Context) {
		filePath := c.Param("filepath")
		themePath := filepath.Join("./data/themes/current", "_app", filePath)
		if _, err := os.Stat(themePath); err == nil {
			c.File(themePath)
			return
		}
		// Fallback to admin build for admin interface
		serveStaticFile(getAdminBuildPath())(c)
	})

	r.GET("/assets/*filepath", func(c *gin.Context) {
		filePath := c.Param("filepath")
		themePath := filepath.Join("./data/themes/current", "assets", filePath)
		if _, err := os.Stat(themePath); err == nil {
			c.File(themePath)
			return
		}
		// Fallback to admin build
		serveStaticFile(getAdminBuildPath())(c)
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		// Try to get favicon from database settings
		settings, err := models.GetOrCreateSettings(database.GetDB())
		if err == nil && settings.FaviconURL != "" {
			// Redirect to the configured favicon
			c.Redirect(http.StatusTemporaryRedirect, settings.FaviconURL)
			return
		}

		// Try theme favicon first
		if _, err := os.Stat("./data/themes/current/favicon.ico"); err == nil {
			c.Header("Content-Type", "image/x-icon")
			c.File("./data/themes/current/favicon.ico")
			return
		}

		// Fallback to admin favicon
		c.Header("Content-Type", "image/x-icon")
		c.File(filepath.Join(getAdminBuildPath(), "favicon.ico"))
	})

	// Public changelog routes - serve theme if available
	r.GET("/", func(c *gin.Context) {
		// Check if theme exists
		if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File("./data/themes/current/index.html")
			return
		}
		// Fallback to admin SPA for setup
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(getAdminIndexPath())
	})

	// Handle slug routes for public changelog (admin is handled by dedicated routes above)
	r.GET("/:slug", func(c *gin.Context) {
		slug := c.Param("slug")

		// Handle admin routes specifically
		if slug == "admin" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File(getAdminIndexPath())
			return
		}

		// Check if theme exists for other slugs
		if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File("./data/themes/current/index.html")
			return
		}

		// No theme available, return 404 for public slugs
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})

	// Fallback for unmatched routes
	r.NoRoute(func(c *gin.Context) {
		// Check if it's an API route
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}

		// Check if it's an admin route
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File(getAdminIndexPath())
			return
		}

		// For other routes, check if theme exists
		if _, err := os.Stat("./data/themes/current/index.html"); err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File("./data/themes/current/index.html")
			return
		}

		// No theme available
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
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
