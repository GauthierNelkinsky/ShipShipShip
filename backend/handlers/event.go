package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chessload-changelog/database"
	"chessload-changelog/models"

	"github.com/gin-gonic/gin"
)

func GetEvents(c *gin.Context) {
	var events []models.Event

	db := database.GetDB()
	if err := db.Where("status != ?", models.StatusArchived).Order("sort_order ASC, created_at ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

func GetAllEvents(c *gin.Context) {
	var events []models.Event

	db := database.GetDB()
	if err := db.Order("sort_order ASC, created_at ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

func GetEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	db := database.GetDB()
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert tags and media arrays to JSON strings
	tagsJSON, _ := json.Marshal(req.Tags)
	mediaJSON, _ := json.Marshal(req.Media)

	// Get the next order value for this status
	var maxOrder int
	db := database.GetDB()
	db.Model(&models.Event{}).Where("status = ?", req.Status).Select("COALESCE(MAX(sort_order), -1) + 1").Scan(&maxOrder)

	order := maxOrder
	if req.Order != nil {
		order = *req.Order
	}

	event := models.Event{
		Title:   req.Title,
		Tags:    string(tagsJSON),
		Media:   string(mediaJSON),
		Status:  req.Status,
		Date:    req.Date,
		Content: req.Content,
		Order:   order,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Update fields if provided
	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		event.Tags = string(tagsJSON)
	}
	if req.Media != nil {
		mediaJSON, _ := json.Marshal(req.Media)
		event.Media = string(mediaJSON)
	}
	if req.Status != nil {
		event.Status = *req.Status
	}
	if req.Date != nil {
		event.Date = *req.Date
	}
	if req.Content != nil {
		event.Content = *req.Content
	}
	if req.Order != nil {
		event.Order = *req.Order
	}

	if err := db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	db := database.GetDB()
	if err := db.Delete(&models.Event{}, eventID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func VoteEvent(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()

	db := database.GetDB()
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Only allow voting on events with "Upcoming" status
	if event.Status != models.StatusUpcoming {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event is not available for voting"})
		return
	}

	// Check if this IP has already voted for this event
	var existingVote models.Vote
	if err := db.Where("event_id = ? AND ip_address = ?", eventID, clientIP).First(&existingVote).Error; err == nil {
		// User has already voted, so remove the vote (toggle functionality)
		if err := db.Delete(&existingVote).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}

		// Decrement vote count
		if event.Votes > 0 {
			event.Votes--
		}
		if err := db.Save(&event).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Vote removed successfully",
			"votes":   event.Votes,
			"voted":   false,
		})
		return
	}

	// Create vote record
	vote := models.Vote{
		EventID:   uint(eventID),
		IPAddress: clientIP,
	}

	if err := db.Create(&vote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	// Increment vote count
	event.Votes++
	if err := db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
		"votes":   event.Votes,
		"voted":   true,
	})
}

func CheckVoteStatus(c *gin.Context) {
	id := c.Param("id")
	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()

	db := database.GetDB()
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if this IP has voted for this event
	var existingVote models.Vote
	hasVoted := db.Where("event_id = ? AND ip_address = ?", eventID, clientIP).First(&existingVote).Error == nil

	c.JSON(http.StatusOK, gin.H{
		"voted": hasVoted,
		"votes": event.Votes,
	})
}

// ReorderEvents handles reordering of events within the same status
func ReorderEvents(c *gin.Context) {
	var req struct {
		EventID  uint   `json:"event_id" binding:"required"`
		NewOrder int    `json:"new_order" binding:"min=0"`
		Status   string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the event to reorder
	var event models.Event
	if err := tx.First(&event, req.EventID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Get all events in the same status ordered by current order
	var events []models.Event
	if err := tx.Where("status = ?", req.Status).Order("sort_order ASC").Find(&events).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	// Find current position of the event
	currentIndex := -1
	for i, e := range events {
		if e.ID == req.EventID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found in status"})
		return
	}

	// Ensure new order is within bounds
	if req.NewOrder < 0 {
		req.NewOrder = 0
	}
	if req.NewOrder >= len(events) {
		req.NewOrder = len(events) - 1
	}

	// Remove event from current position
	events = append(events[:currentIndex], events[currentIndex+1:]...)

	// Insert event at new position
	if req.NewOrder >= len(events) {
		events = append(events, event)
	} else {
		events = append(events[:req.NewOrder], append([]models.Event{event}, events[req.NewOrder:]...)...)
	}

	// Update order values for all events in this status
	for i, e := range events {
		if err := tx.Model(&e).Update("sort_order", i).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event order"})
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Events reordered successfully"})
}

// SubmitFeedback allows public users to submit feedback
func SubmitFeedback(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create feedback event
	tagsJSON, _ := json.Marshal([]string{"Feedback"})
	mediaJSON, _ := json.Marshal([]string{})

	// Get the next order value for backlog status
	var maxOrder int
	db := database.GetDB()
	db.Model(&models.Event{}).Where("status = ?", models.StatusBacklogs).Select("COALESCE(MAX(sort_order), -1) + 1").Scan(&maxOrder)

	event := models.Event{
		Title:   req.Title,
		Tags:    string(tagsJSON),
		Media:   string(mediaJSON),
		Status:  models.StatusBacklogs,
		Date:    "",
		Content: req.Content,
		Order:   maxOrder,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feedback submitted successfully",
		"id":      event.ID,
	})
}
