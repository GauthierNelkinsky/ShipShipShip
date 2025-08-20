package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

// GetTags returns all tags
func GetTags(c *gin.Context) {
	var tags []models.Tag

	db := database.GetDB()
	if err := db.Order("name ASC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// GetTag returns a single tag by ID
func GetTag(c *gin.Context) {
	id := c.Param("id")
	tagID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var tag models.Tag
	db := database.GetDB()
	if err := db.First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// CreateTag creates a new tag
func CreateTag(c *gin.Context) {
	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate color format (should be hex color)
	if len(req.Color) != 7 || req.Color[0] != '#' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Color must be in hex format (e.g., #FF0000)"})
		return
	}

	tag := models.Tag{
		Name:  req.Name,
		Color: req.Color,
	}

	db := database.GetDB()
	if err := db.Create(&tag).Error; err != nil {
		// Check if it's a unique constraint violation
		if err.Error() == "UNIQUE constraint failed: tags.name" {
			c.JSON(http.StatusConflict, gin.H{"error": "Tag with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// UpdateTag updates an existing tag
func UpdateTag(c *gin.Context) {
	id := c.Param("id")
	tagID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var tag models.Tag
	if err := db.First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.Color != nil {
		// Validate color format
		if len(*req.Color) != 7 || (*req.Color)[0] != '#' {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Color must be in hex format (e.g., #FF0000)"})
			return
		}
		tag.Color = *req.Color
	}

	if err := db.Save(&tag).Error; err != nil {
		// Check if it's a unique constraint violation
		if err.Error() == "UNIQUE constraint failed: tags.name" {
			c.JSON(http.StatusConflict, gin.H{"error": "Tag with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// DeleteTag deletes a tag
func DeleteTag(c *gin.Context) {
	id := c.Param("id")
	tagID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	db := database.GetDB()
	var tag models.Tag
	if err := db.First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	// Protect the Feedback tag from deletion
	if strings.ToLower(tag.Name) == "feedback" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "The 'Feedback' tag cannot be deleted as it's used by the system",
		})
		return
	}

	// Start a transaction to ensure atomicity
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove all associations between this tag and events from the junction table
	if err := tx.Exec("DELETE FROM event_tags WHERE tag_id = ?", tagID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tag associations"})
		return
	}

	// Delete the tag itself
	if err := tx.Delete(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag"})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// GetTagUsage returns usage statistics for all tags
func GetTagUsage(c *gin.Context) {
	type TagUsage struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
		Count int64  `json:"count"`
	}

	var tagUsage []TagUsage
	db := database.GetDB()

	// Get all tags with their usage count
	if err := db.Raw(`
		SELECT
			t.id,
			t.name,
			t.color,
			COALESCE(COUNT(et.event_id), 0) as count
		FROM tags t
		LEFT JOIN event_tags et ON t.id = et.tag_id
		LEFT JOIN events e ON et.event_id = e.id AND e.deleted_at IS NULL
		GROUP BY t.id, t.name, t.color
		ORDER BY count DESC, t.name ASC
	`).Scan(&tagUsage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag usage"})
		return
	}

	c.JSON(http.StatusOK, tagUsage)
}
