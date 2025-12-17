package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStatuses returns all status definitions ordered by `order` then `display_name`
func GetStatuses(c *gin.Context) {
	db := database.GetDB()

	var statuses []models.EventStatusDefinition
	if err := db.Order("`order` ASC, display_name ASC").Find(&statuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statuses"})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// GetStatus returns a single status definition by ID
func GetStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing status id"})
		return
	}

	db := database.GetDB()
	var status models.EventStatusDefinition
	if err := db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status"})
		}
		return
	}

	c.JSON(http.StatusOK, status)
}

// CreateStatus creates a new status definition
func CreateStatus(c *gin.Context) {
	var req models.CreateStatusDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nameTrimmed := strings.TrimSpace(req.DisplayName)
	if nameTrimmed == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_name cannot be empty"})
		return
	}

	db := database.GetDB()

	// Check uniqueness
	var count int64
	db.Model(&models.EventStatusDefinition{}).Where("LOWER(display_name) = ?", strings.ToLower(nameTrimmed)).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Status with same name already exists"})
		return
	}

	// Determine next order
	var maxOrder int
	db.Model(&models.EventStatusDefinition{}).Select("COALESCE(MAX(`order`),0)").Scan(&maxOrder)

	order := maxOrder + 1
	if req.Order != nil {
		order = *req.Order
	}

	// Generate slug from display name
	slug := utils.GenerateUniqueSlug(db, nameTrimmed, "event_status_definitions")

	status := models.EventStatusDefinition{
		DisplayName: nameTrimmed,
		Slug:        slug,
		Order:       order,
		IsReserved:  false,
	}

	if err := db.Create(&status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create status"})
		return
	}

	// Create category mapping if category_id is provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		// Get current theme ID from settings
		var settings models.ProjectSettings
		if err := db.First(&settings).Error; err == nil && settings.CurrentThemeID != "" {
			mapping := models.StatusCategoryMapping{
				StatusDefinitionID: status.ID,
				ThemeID:            settings.CurrentThemeID,
				CategoryID:         *req.CategoryID,
			}
			if err := db.Create(&mapping).Error; err != nil {
				// Log error but don't fail the status creation
				c.JSON(http.StatusCreated, gin.H{
					"status":  status,
					"warning": "Status created but category mapping failed",
				})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, status)
}

// UpdateStatus updates a status definition
func UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing status id"})
		return
	}

	var req models.UpdateStatusDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var status models.EventStatusDefinition
	if err := db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status"})
		}
		return
	}

	originalName := status.DisplayName

	// Apply changes
	if req.DisplayName != nil {
		newName := strings.TrimSpace(*req.DisplayName)
		if newName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "display_name cannot be empty"})
			return
		}

		// Ensure uniqueness of new display name
		var count int64
		db.Model(&models.EventStatusDefinition{}).
			Where("id != ? AND LOWER(display_name) = ?", status.ID, strings.ToLower(newName)).
			Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Another status with this name already exists"})
			return
		}

		status.DisplayName = newName
		// Regenerate slug when display name changes
		status.Slug = utils.GenerateUniqueSlug(db, newName, "event_status_definitions", status.ID)
	}

	if req.Order != nil {
		status.Order = *req.Order
	}

	if err := db.Save(&status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// If display name changed, update events referencing old name
	if req.DisplayName != nil && originalName != status.DisplayName {
		db.Model(&models.Event{}).Where("status = ?", originalName).Update("status", status.DisplayName)

		// Also update newsletter automation trigger statuses
		automationSettings, err := models.GetOrCreateAutomationSettings(db)
		if err == nil && automationSettings.TriggerStatuses != "" {
			var triggerStatuses []string
			if err := json.Unmarshal([]byte(automationSettings.TriggerStatuses), &triggerStatuses); err == nil {
				// Replace old status name with new status name
				updated := false
				for i, ts := range triggerStatuses {
					if ts == originalName {
						triggerStatuses[i] = status.DisplayName
						updated = true
					}
				}

				// Save updated trigger statuses if changed
				if updated {
					statusesJSON, err := json.Marshal(triggerStatuses)
					if err == nil {
						db.Model(&automationSettings).Update("trigger_statuses", string(statusesJSON))
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, status)
}

// DeleteStatus deletes a status definition (blocked if in use)
func DeleteStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing status id"})
		return
	}

	db := database.GetDB()
	var status models.EventStatusDefinition
	if err := db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch status"})
		}
		return
	}

	// Check usage
	var eventCount int64
	db.Model(&models.Event{}).Where("status = ?", status.DisplayName).Count(&eventCount)
	if eventCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete status while it is used by events"})
		return
	}

	// Delete any status category mappings first
	if err := db.Where("status_definition_id = ?", status.ID).Delete(&models.StatusCategoryMapping{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete status mappings"})
		return
	}

	// Remove this status from newsletter automation trigger statuses
	automationSettings, err := models.GetOrCreateAutomationSettings(db)
	if err == nil && automationSettings.TriggerStatuses != "" {
		var triggerStatuses []string
		if err := json.Unmarshal([]byte(automationSettings.TriggerStatuses), &triggerStatuses); err == nil {
			// Filter out the deleted status
			updatedStatuses := []string{}
			for _, ts := range triggerStatuses {
				if ts != status.DisplayName {
					updatedStatuses = append(updatedStatuses, ts)
				}
			}

			// Save updated trigger statuses if changed
			if len(updatedStatuses) != len(triggerStatuses) {
				statusesJSON, err := json.Marshal(updatedStatuses)
				if err == nil {
					db.Model(&automationSettings).Update("trigger_statuses", string(statusesJSON))
				}
			}
		}
	}

	if err := db.Delete(&status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status deleted"})
}

// ReorderStatuses sets ordering based on provided list.
// Request body: { "order": [ { "id": 1, "order": 0 }, { "id": 2, "order": 1 } ] }
func ReorderStatuses(c *gin.Context) {
	var req struct {
		Order []struct {
			ID    uint `json:"id"`
			Order int  `json:"order"`
		} `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	for _, item := range req.Order {
		if err := db.Model(&models.EventStatusDefinition{}).
			Where("id = ?", item.ID).
			Update("order", item.Order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order for status id " + string(rune(item.ID))})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Statuses reordered"})
}
