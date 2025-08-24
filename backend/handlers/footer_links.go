package handlers

import (
	"net/http"
	"strconv"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

// GetFooterLinks returns all footer links
func GetFooterLinks(c *gin.Context) {
	db := database.GetDB()
	links, err := models.GetFooterLinks(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch footer links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": links})
}

// GetFooterLinksByColumn returns footer links grouped by column
func GetFooterLinksByColumn(c *gin.Context) {
	db := database.GetDB()
	linksByColumn, err := models.GetFooterLinksByColumn(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch footer links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": linksByColumn})
}

// GetFooterLink returns a specific footer link by ID
func GetFooterLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	db := database.GetDB()
	var link models.FooterLink
	if err := db.First(&link, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Footer link not found"})
		return
	}

	c.JSON(http.StatusOK, link)
}

// CreateFooterLink creates a new footer link
func CreateFooterLink(c *gin.Context) {
	var req models.CreateFooterLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate column type
	if req.Column != models.FooterColumnLeft && req.Column != models.FooterColumnMiddle && req.Column != models.FooterColumnRight {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column type. Must be 'left', 'middle', or 'right'"})
		return
	}

	db := database.GetDB()

	// Get next order for the column
	nextOrder, err := models.GetNextOrder(db, req.Column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate order"})
		return
	}

	link := models.FooterLink{
		Name:            req.Name,
		URL:             req.URL,
		Column:          req.Column,
		Order:           nextOrder,
		OpenInNewWindow: req.OpenInNewWindow,
	}

	if err := db.Create(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create footer link"})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// UpdateFooterLink updates an existing footer link
func UpdateFooterLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.UpdateFooterLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var link models.FooterLink
	if err := db.First(&link, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Footer link not found"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		link.Name = *req.Name
	}
	if req.URL != nil {
		link.URL = *req.URL
	}
	if req.Column != nil {
		// Validate column type
		if *req.Column != models.FooterColumnLeft && *req.Column != models.FooterColumnMiddle && *req.Column != models.FooterColumnRight {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column type. Must be 'left', 'middle', or 'right'"})
			return
		}

		// If column is changing, get next order for the new column
		if link.Column != *req.Column {
			nextOrder, err := models.GetNextOrder(db, *req.Column)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate order"})
				return
			}
			link.Order = nextOrder
		}
		link.Column = *req.Column
	}
	if req.OpenInNewWindow != nil {
		link.OpenInNewWindow = *req.OpenInNewWindow
	}

	if err := db.Save(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update footer link"})
		return
	}

	c.JSON(http.StatusOK, link)
}

// DeleteFooterLink deletes a footer link
func DeleteFooterLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	db := database.GetDB()
	var link models.FooterLink
	if err := db.First(&link, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Footer link not found"})
		return
	}

	if err := db.Delete(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete footer link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Footer link deleted successfully"})
}

// ReorderFooterLinks updates the order of footer links
func ReorderFooterLinks(c *gin.Context) {
	var req models.ReorderFooterLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Start a transaction for atomic updates
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update each link's order
	for _, linkUpdate := range req.Links {
		if err := tx.Model(&models.FooterLink{}).Where("id = ?", linkUpdate.ID).Update("order", linkUpdate.Order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update link order"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit order changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Footer links reordered successfully"})
}
