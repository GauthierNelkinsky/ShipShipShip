package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

// GetThemeManifest returns the current theme's manifest
func GetThemeManifest(c *gin.Context) {
	themePath := "./data/themes/current"

	manifest, err := models.LoadThemeManifest(themePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load theme manifest",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"manifest": manifest,
	})
}

// GetStatusMappings returns all status-to-category mappings for the current theme
func GetStatusMappings(c *gin.Context) {
	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":           true,
			"mappings":          []interface{}{},
			"unmapped_statuses": []interface{}{},
		})
		return
	}

	// Load theme manifest
	manifest, err := models.LoadThemeManifest("./data/themes/current")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load theme manifest",
			"details": err.Error(),
		})
		return
	}

	// Get all status definitions
	var statuses []models.EventStatusDefinition
	if err := db.Order("`order` ASC").Find(&statuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statuses"})
		return
	}

	// Get all mappings
	var mappings []models.StatusCategoryMapping
	if err := db.Where("theme_id = ?", settings.CurrentThemeID).Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mappings"})
		return
	}

	// Create mapping lookup
	mappingLookup := make(map[uint]*models.StatusCategoryMapping)
	for i := range mappings {
		mappingLookup[mappings[i].StatusDefinitionID] = &mappings[i]
	}

	// Create category lookup
	categoryLookup := make(map[string]*models.ThemeCategory)
	for i := range manifest.Categories {
		categoryLookup[manifest.Categories[i].ID] = &manifest.Categories[i]
	}

	// Build response
	type MappingResponse struct {
		StatusID      uint   `json:"status_id"`
		StatusName    string `json:"status_name"`
		CategoryID    string `json:"category_id"`
		CategoryLabel string `json:"category_label"`
		ThemeID       string `json:"theme_id"`
	}

	type UnmappedStatusResponse struct {
		StatusID          uint   `json:"status_id"`
		StatusName        string `json:"status_name"`
		SuggestedCategory string `json:"suggested_category"`
	}

	mappedStatuses := []MappingResponse{}
	unmappedStatuses := []UnmappedStatusResponse{}

	for _, status := range statuses {
		if mapping, exists := mappingLookup[status.ID]; exists {
			response := MappingResponse{
				StatusID:   status.ID,
				StatusName: status.DisplayName,
				CategoryID: mapping.CategoryID,
				ThemeID:    mapping.ThemeID,
			}

			if category, found := categoryLookup[mapping.CategoryID]; found {
				response.CategoryLabel = category.Label
			}

			mappedStatuses = append(mappedStatuses, response)
		} else {
			// Status not mapped yet
			suggested := models.SuggestCategoryForStatus(status.DisplayName, manifest.Categories)
			unmappedStatuses = append(unmappedStatuses, UnmappedStatusResponse{
				StatusID:          status.ID,
				StatusName:        status.DisplayName,
				SuggestedCategory: suggested,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"theme_id":          settings.CurrentThemeID,
		"theme_name":        manifest.Name,
		"mappings":          mappedStatuses,
		"unmapped_statuses": unmappedStatuses,
	})
}

// UpdateStatusMapping updates the category mapping for a status
func UpdateStatusMapping(c *gin.Context) {
	statusIDStr := c.Param("statusId")
	statusID, err := strconv.ParseUint(statusIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status ID"})
		return
	}

	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No theme is currently applied"})
		return
	}

	// Verify status exists
	var status models.EventStatusDefinition
	if err := db.First(&status, statusID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		return
	}

	// Verify category exists in theme
	manifest, err := models.LoadThemeManifest("./data/themes/current")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theme manifest"})
		return
	}

	categoryExists := false
	for _, cat := range manifest.Categories {
		if cat.ID == req.CategoryID {
			categoryExists = true
			break
		}
	}

	if !categoryExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Category '%s' does not exist in current theme", req.CategoryID),
		})
		return
	}

	// Update or create mapping
	var mapping models.StatusCategoryMapping
	err = db.Where("status_definition_id = ? AND theme_id = ?", statusID, settings.CurrentThemeID).First(&mapping).Error

	if err == nil {
		// Update existing mapping
		mapping.CategoryID = req.CategoryID
		if err := db.Save(&mapping).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mapping"})
			return
		}
	} else {
		// Create new mapping
		mapping = models.StatusCategoryMapping{
			StatusDefinitionID: uint(statusID),
			ThemeID:            settings.CurrentThemeID,
			CategoryID:         req.CategoryID,
		}
		if err := db.Create(&mapping).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create mapping"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"mapping": mapping,
	})
}

// DeleteStatusMapping removes the category mapping for a status
func DeleteStatusMapping(c *gin.Context) {
	statusIDStr := c.Param("statusId")
	statusID, err := strconv.ParseUint(statusIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status ID"})
		return
	}

	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No theme is currently applied"})
		return
	}

	// Delete the mapping
	result := db.Where("status_definition_id = ? AND theme_id = ?", statusID, settings.CurrentThemeID).
		Delete(&models.StatusCategoryMapping{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mapping deleted successfully",
	})
}

// GetPublicEventsByCategory returns events grouped by theme category
func GetPublicEventsByCategory(c *gin.Context) {
	db := database.GetDB()

	// Get current theme
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	// Load theme manifest
	manifest, err := models.LoadThemeManifest("./data/themes/current")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theme manifest"})
		return
	}

	// Get all public events
	var events []models.Event
	if err := db.Preload("Tags").
		Where("is_public = ?", true).
		Where("status != ?", models.StatusArchived).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	// Get all status definitions
	var statusDefs []models.EventStatusDefinition
	if err := db.Find(&statusDefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status definitions"})
		return
	}

	// Create status -> category lookup
	statusCategoryMap := make(map[string]string)
	for _, statusDef := range statusDefs {
		var mapping models.StatusCategoryMapping
		err := db.Where("status_definition_id = ? AND theme_id = ?", statusDef.ID, settings.CurrentThemeID).
			First(&mapping).Error

		if err == nil {
			statusCategoryMap[statusDef.DisplayName] = mapping.CategoryID
		}
		// Skip unmapped statuses - they won't appear in any category
	}

	// Group events by category
	categorizedEvents := make(map[string][]models.Event)

	// Initialize all categories from manifest
	for _, category := range manifest.Categories {
		categorizedEvents[category.ID] = []models.Event{}
	}

	// Populate events
	for _, event := range events {
		categoryID, exists := statusCategoryMap[string(event.Status)]
		if exists {
			categorizedEvents[categoryID] = append(categorizedEvents[categoryID], event)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"theme_id":   settings.CurrentThemeID,
		"theme_name": manifest.Name,
		"categories": categorizedEvents,
	})
}
