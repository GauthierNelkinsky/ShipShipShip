package handlers

import (
	"fmt"
	"net/http"

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

	c.JSON(http.StatusOK, settings)
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

	if req.LogoURL != nil {
		// Clean up old logo file if it's being replaced or removed
		if settings.LogoURL != "" && isImageURL(settings.LogoURL) && settings.LogoURL != *req.LogoURL {
			if err := deleteImageFromURL(settings.LogoURL); err != nil {
				fmt.Printf("Warning: Failed to cleanup old logo file: %v\n", err)
			}
		}
		settings.LogoURL = *req.LogoURL
	}
	if req.DarkLogoURL != nil {
		// Clean up old dark logo file if it's being replaced or removed
		if settings.DarkLogoURL != "" && isImageURL(settings.DarkLogoURL) && settings.DarkLogoURL != *req.DarkLogoURL {
			if err := deleteImageFromURL(settings.DarkLogoURL); err != nil {
				fmt.Printf("Warning: Failed to cleanup old dark logo file: %v\n", err)
			}
		}
		settings.DarkLogoURL = *req.DarkLogoURL
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
	if req.PrimaryColor != nil {
		settings.PrimaryColor = *req.PrimaryColor
	}
	if req.NewsletterEnabled != nil {
		settings.NewsletterEnabled = *req.NewsletterEnabled
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}
