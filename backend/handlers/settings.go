package handlers

import (
	"fmt"
	"net/http"
	"os"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

func GetSettings(c *gin.Context) {
	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	// Get environment mode
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "production"
	}

	// Sanitize URLs to remove localhost references
	faviconURL := sanitizeImageURL(settings.FaviconURL)
	websiteURL := sanitizeImageURL(settings.WebsiteURL)

	// Return settings with environment info
	response := gin.H{
		"id":                    settings.ID,
		"title":                 settings.Title,
		"favicon_url":           faviconURL,
		"website_url":           websiteURL,
		"current_theme_id":      settings.CurrentThemeID,
		"current_theme_version": settings.CurrentThemeVersion,
		"created_at":            settings.CreatedAt,
		"updated_at":            settings.UpdatedAt,
		"environment":           environment,
	}

	c.JSON(http.StatusOK, response)
}

func UpdateSettings(c *gin.Context) {
	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	// Update fields if provided
	if req.Title != nil {
		settings.Title = *req.Title
	}

	if req.FaviconURL != nil {
		// Clean up old favicon file if it's being replaced or removed
		if settings.FaviconURL != "" && isImageURL(settings.FaviconURL) && settings.FaviconURL != *req.FaviconURL {
			if err := deleteImageFromURL(settings.FaviconURL); err != nil {
				fmt.Printf("Warning: Failed to cleanup old favicon file: %v\n", err)
			}
		}
		settings.FaviconURL = *req.FaviconURL
	}

	if req.WebsiteURL != nil {
		settings.WebsiteURL = *req.WebsiteURL
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}
