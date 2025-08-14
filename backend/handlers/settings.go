package handlers

import (
	"net/http"

	"chessload-changelog/database"
	"chessload-changelog/models"

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
		settings.LogoURL = *req.LogoURL
	}
	if req.DarkLogoURL != nil {
		settings.DarkLogoURL = *req.DarkLogoURL
	}
	if req.FaviconURL != nil {
		settings.FaviconURL = *req.FaviconURL
	}
	if req.WebsiteURL != nil {
		settings.WebsiteURL = *req.WebsiteURL
	}
	if req.PrimaryColor != nil {
		settings.PrimaryColor = *req.PrimaryColor
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}
