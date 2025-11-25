package handlers

import (
	"net/http"
	"strconv"

	"shipshipship/database"
	"shipshipship/models"

	"github.com/gin-gonic/gin"
)

// AddOrRemoveReaction handles adding or removing a reaction (toggle behavior)
func AddOrRemoveReaction(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		ReactionType models.ReactionType `json:"reaction_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate reaction type
	if !models.IsValidReactionType(req.ReactionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reaction type"})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()

	db := database.GetDB()

	// Check if event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if this IP has already reacted with this type
	var existingReaction models.EventReaction
	err = db.Where("event_id = ? AND ip_address = ? AND reaction_type = ?", eventID, clientIP, req.ReactionType).
		First(&existingReaction).Error

	if err == nil {
		// Reaction exists, remove it (toggle off)
		if err := db.Delete(&existingReaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
			return
		}

		// Get updated reaction summary
		summary := getReactionSummary(db, uint(eventID), clientIP)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Reaction removed successfully",
			"removed":  true,
			"reaction": req.ReactionType,
			"summary":  summary,
		})
		return
	}

	// Reaction doesn't exist, create it
	reaction := models.EventReaction{
		EventID:      uint(eventID),
		ReactionType: req.ReactionType,
		IPAddress:    clientIP,
	}

	if err := db.Create(&reaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	// Get updated reaction summary
	summary := getReactionSummary(db, uint(eventID), clientIP)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Reaction added successfully",
		"added":    true,
		"reaction": req.ReactionType,
		"summary":  summary,
	})
}

// GetEventReactions returns all reactions for an event
func GetEventReactions(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()

	db := database.GetDB()

	// Check if event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	summary := getReactionSummary(db, uint(eventID), clientIP)

	c.JSON(http.StatusOK, summary)
}

// GetMyReactions returns the current user/IP's reactions for an event
func GetMyReactions(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()

	db := database.GetDB()

	// Get user's reactions
	var reactions []models.EventReaction
	db.Where("event_id = ? AND ip_address = ?", eventID, clientIP).Find(&reactions)

	reactionTypes := make([]models.ReactionType, len(reactions))
	for i, r := range reactions {
		reactionTypes[i] = r.ReactionType
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":  eventID,
		"reactions": reactionTypes,
	})
}

// GetAllEventReactionsCount returns reaction counts for all events
func GetAllEventReactionsCount(c *gin.Context) {
	db := database.GetDB()

	// Get all reaction counts grouped by event_id
	var results []struct {
		EventID uint  `json:"event_id"`
		Count   int64 `json:"count"`
	}

	db.Model(&models.EventReaction{}).
		Select("event_id, COUNT(*) as count").
		Group("event_id").
		Find(&results)

	// Convert to map for easier lookup
	reactionCounts := make(map[uint]int64)
	for _, r := range results {
		reactionCounts[r.EventID] = r.Count
	}

	c.JSON(http.StatusOK, reactionCounts)
}

// MigrateVotesToReactions migrates old votes to thumbs_up reactions
// This is a one-time migration function that can be called via an admin endpoint
func MigrateVotesToReactions(c *gin.Context) {
	db := database.GetDB()

	// Get all existing votes
	var votes []models.Vote
	if err := db.Find(&votes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch votes"})
		return
	}

	migratedCount := 0
	errorCount := 0

	// Convert each vote to a thumbs_up reaction
	for _, vote := range votes {
		// Check if reaction already exists
		var existingReaction models.EventReaction
		err := db.Where("event_id = ? AND ip_address = ? AND reaction_type = ?",
			vote.EventID, vote.IPAddress, models.ReactionThumbsUp).
			First(&existingReaction).Error

		if err == nil {
			// Reaction already exists, skip
			continue
		}

		// Create new reaction
		reaction := models.EventReaction{
			EventID:      vote.EventID,
			ReactionType: models.ReactionThumbsUp,
			IPAddress:    vote.IPAddress,
			CreatedAt:    vote.CreatedAt,
			UpdatedAt:    vote.UpdatedAt,
		}

		if err := db.Create(&reaction).Error; err != nil {
			errorCount++
			continue
		}

		migratedCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Migration completed",
		"migrated_count": migratedCount,
		"error_count":    errorCount,
		"total_votes":    len(votes),
	})
}

// GetReactionTypes returns all available reaction types with metadata
func GetReactionTypes(c *gin.Context) {
	reactionTypes := models.ValidReactionTypes()

	type ReactionTypeInfo struct {
		Type  models.ReactionType `json:"type"`
		Emoji string              `json:"emoji"`
		Label string              `json:"label"`
	}

	reactionInfo := make([]ReactionTypeInfo, len(reactionTypes))
	for i, rt := range reactionTypes {
		reactionInfo[i] = ReactionTypeInfo{
			Type:  rt,
			Emoji: models.GetReactionEmoji(rt),
			Label: models.GetReactionLabel(rt),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reactions": reactionInfo,
	})
}
