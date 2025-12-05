package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// Create a set of valid status IDs for quick lookup
	validStatusIDs := make(map[uint]bool)
	for _, status := range statuses {
		validStatusIDs[status.ID] = true
	}

	// Clean up orphaned mappings (mappings for deleted statuses)
	for i := range mappings {
		if !validStatusIDs[mappings[i].StatusDefinitionID] {
			// Delete orphaned mapping
			db.Delete(&mappings[i])
		}
	}

	// Create mapping lookup (only for valid statuses)
	mappingLookup := make(map[uint]*models.StatusCategoryMapping)
	for i := range mappings {
		if validStatusIDs[mappings[i].StatusDefinitionID] {
			mappingLookup[mappings[i].StatusDefinitionID] = &mappings[i]
		}
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

	var targetCategory *models.ThemeCategory
	for i, cat := range manifest.Categories {
		if cat.ID == req.CategoryID {
			targetCategory = &manifest.Categories[i]
			break
		}
	}

	if targetCategory == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Category '%s' does not exist in current theme", req.CategoryID),
		})
		return
	}

	// Check if category allows multiple statuses
	if !targetCategory.Multiple {
		// Check if another status is already mapped to this category
		var existingMappings []models.StatusCategoryMapping
		err = db.Where("theme_id = ? AND category_id = ? AND status_definition_id != ?",
			settings.CurrentThemeID, req.CategoryID, statusID).Find(&existingMappings).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing mappings"})
			return
		}

		if len(existingMappings) > 0 {
			// Check if the existing status still exists, clean up orphaned mappings
			var existingStatus models.EventStatusDefinition
			if err := db.First(&existingStatus, existingMappings[0].StatusDefinitionID).Error; err == nil {
				// Status exists, return error
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Category '%s' does not allow multiple statuses. Status '%s' is already mapped to this category.", targetCategory.Label, existingStatus.DisplayName),
				})
				return
			} else if err == gorm.ErrRecordNotFound {
				// Orphaned mapping - clean it up and continue
				db.Delete(&existingMappings[0])
				// Don't return error, allow the new mapping to be created
			} else {
				// Database error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify existing status"})
				return
			}
		}
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

// GetThemeSettings returns the current values for theme settings
func GetThemeSettings(c *gin.Context) {
	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"settings": []interface{}{},
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

	// Get all setting values for this theme
	var settingValues []models.ThemeSettingValue
	if err := db.Where("theme_id = ?", settings.CurrentThemeID).Find(&settingValues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting values"})
		return
	}

	// Create a map of setting values
	valueMap := make(map[string]string)
	for _, sv := range settingValues {
		valueMap[sv.SettingID] = sv.Value
	}

	// Build response with settings and their current values
	type SettingResponse struct {
		ID          string      `json:"id"`
		Label       string      `json:"label"`
		Description string      `json:"description"`
		Type        string      `json:"type"`
		Default     interface{} `json:"default"`
		Value       interface{} `json:"value"`
		Options     interface{} `json:"options,omitempty"`
	}

	settingsResponse := []SettingResponse{}
	// Iterate over setting groups
	for _, group := range manifest.Settings {
		// Iterate over settings within each group
		for _, setting := range group.Settings {
			response := SettingResponse{
				ID:          setting.ID,
				Label:       setting.Label,
				Description: setting.Description,
				Type:        setting.Type,
				Default:     setting.Default,
				Value:       setting.Default, // Default to the default value
			}

			// Include options for select type
			if setting.Type == "select" && len(setting.Options) > 0 {
				response.Options = setting.Options
			}

			// If user has set a value, use that instead
			if val, exists := valueMap[setting.ID]; exists {
				// Parse the stored JSON value based on type
				if setting.Type == "boolean" {
					response.Value = val == "true"
				} else if setting.Type == "number" {
					// Parse as number
					var num float64
					fmt.Sscanf(val, "%f", &num)
					response.Value = num
				} else {
					response.Value = val
				}
			}

			settingsResponse = append(settingsResponse, response)
		}
	}

	// Get all status definitions
	var statusDefs []models.EventStatusDefinition
	if err := db.Find(&statusDefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status definitions"})
		return
	}

	// Fetch all status mappings for the current theme at once
	var mappings []models.StatusCategoryMapping
	if err := db.Where("theme_id = ?", settings.CurrentThemeID).Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status mappings"})
		return
	}

	// Build a map of status ID to category ID for quick lookup
	statusToCategoryMap := make(map[uint]string)
	for _, mapping := range mappings {
		statusToCategoryMap[mapping.StatusDefinitionID] = mapping.CategoryID
	}

	// Build a map of statuses grouped by category
	statusesByCategory := make(map[string][]string)

	for _, statusDef := range statusDefs {
		if categoryID, exists := statusToCategoryMap[statusDef.ID]; exists {
			// Status is mapped to a category
			if statusesByCategory[categoryID] == nil {
				statusesByCategory[categoryID] = []string{}
			}
			statusesByCategory[categoryID] = append(statusesByCategory[categoryID], statusDef.DisplayName)
		}
	}

	// Add statuses as additional settings entries
	for categoryID, statuses := range statusesByCategory {
		settingsResponse = append(settingsResponse, SettingResponse{
			ID:    categoryID + "-statuses",
			Label: categoryID + " Statuses",
			Type:  "array",
			Value: statuses,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"theme_id": settings.CurrentThemeID,
		"settings": settingsResponse,
	})
}

// UpdateThemeSettings updates theme setting values
func UpdateThemeSettings(c *gin.Context) {
	var req map[string]interface{}

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

	// Load theme manifest to validate settings
	manifest, err := models.LoadThemeManifest("./data/themes/current")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load theme manifest"})
		return
	}

	// Create a map of valid settings
	validSettings := make(map[string]models.ThemeSetting)
	for _, group := range manifest.Settings {
		for _, setting := range group.Settings {
			validSettings[setting.ID] = setting
		}
	}

	// Update each setting value
	for settingID, value := range req {
		// Validate that this setting exists in the theme
		setting, exists := validSettings[settingID]
		if !exists {
			continue // Skip invalid settings
		}

		// Convert value to string for storage
		var valueStr string
		switch v := value.(type) {
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		case float64:
			valueStr = fmt.Sprintf("%v", v)
		case string:
			valueStr = v
		default:
			// For arrays and objects, serialize as JSON
			if setting.Type == "array" || setting.Type == "object" {
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid value for %s", settingID)})
					return
				}
				valueStr = string(jsonBytes)
			} else {
				valueStr = fmt.Sprintf("%v", v)
			}
		}

		// Update or create setting value
		var settingValue models.ThemeSettingValue
		err := db.Where("theme_id = ? AND setting_id = ?", settings.CurrentThemeID, settingID).
			First(&settingValue).Error

		if err == nil {
			// Update existing
			settingValue.Value = valueStr
			if err := db.Save(&settingValue).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
				return
			}
		} else {
			// Create new
			settingValue = models.ThemeSettingValue{
				ThemeID:   settings.CurrentThemeID,
				SettingID: settingID,
				Value:     valueStr,
			}
			if err := db.Create(&settingValue).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// GetPublicThemeSettings returns theme settings for public access (no authentication required)
func GetPublicThemeSettings(c *gin.Context) {
	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"theme_id": "",
			"settings": map[string]interface{}{},
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

	// Get all setting values for this theme
	var settingValues []models.ThemeSettingValue
	if err := db.Where("theme_id = ?", settings.CurrentThemeID).Find(&settingValues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting values"})
		return
	}

	// Create a map of setting values
	valueMap := make(map[string]string)
	for _, sv := range settingValues {
		valueMap[sv.SettingID] = sv.Value
	}

	// Build simplified response with just setting IDs and values
	settingsResponse := make(map[string]interface{})
	for _, group := range manifest.Settings {
		for _, setting := range group.Settings {
			var value interface{} = setting.Default

			// If user has set a value, use that instead
			if val, exists := valueMap[setting.ID]; exists {
				// Parse the stored JSON value based on type
				if setting.Type == "boolean" {
					value = val == "true"
				} else if setting.Type == "number" {
					// Parse as number
					var num float64
					fmt.Sscanf(val, "%f", &num)
					value = num
				} else if setting.Type == "array" || setting.Type == "object" {
					// Parse JSON for arrays and objects
					var parsed interface{}
					if err := json.Unmarshal([]byte(val), &parsed); err == nil {
						value = parsed
					} else {
						value = setting.Default
					}
				} else {
					value = val
				}
			}

			settingsResponse[setting.ID] = value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"theme_id": settings.CurrentThemeID,
		"settings": settingsResponse,
	})
}

// GetPublicStatusMappings returns status mappings grouped by category for public access
func GetPublicStatusMappings(c *gin.Context) {
	db := database.GetDB()

	// Get current theme ID
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.CurrentThemeID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"theme_id":   "",
			"categories": map[string]interface{}{},
		})
		return
	}

	// Get all status definitions with order
	var statusDefs []models.EventStatusDefinition
	if err := db.Order("`order` ASC").Find(&statusDefs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status definitions"})
		return
	}

	// Build a map of statuses grouped by category
	type StatusDetail struct {
		ID          uint   `json:"id"`
		DisplayName string `json:"display_name"`
		Slug        string `json:"slug"`
		Order       int    `json:"order"`
		IsReserved  bool   `json:"is_reserved"`
	}

	statusesByCategory := make(map[string][]StatusDetail)

	for _, statusDef := range statusDefs {
		var mapping models.StatusCategoryMapping
		err := db.Where("status_definition_id = ? AND theme_id = ?", statusDef.ID, settings.CurrentThemeID).
			First(&mapping).Error

		if err == nil {
			// Status is mapped to a category
			if statusesByCategory[mapping.CategoryID] == nil {
				statusesByCategory[mapping.CategoryID] = []StatusDetail{}
			}
			statusesByCategory[mapping.CategoryID] = append(statusesByCategory[mapping.CategoryID], StatusDetail{
				ID:          statusDef.ID,
				DisplayName: statusDef.DisplayName,
				Slug:        statusDef.Slug,
				Order:       statusDef.Order,
				IsReserved:  statusDef.IsReserved,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"theme_id":   settings.CurrentThemeID,
		"categories": statusesByCategory,
	})
}
