package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/services"
	"shipshipship/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper function to get reaction summary for an event
func getReactionSummary(db *gorm.DB, eventID uint, clientIP string) models.ReactionSummary {
	var reactions []models.ReactionCount

	// Get count for each reaction type
	db.Model(&models.EventReaction{}).
		Select("reaction_type, COUNT(*) as count").
		Where("event_id = ?", eventID).
		Group("reaction_type").
		Scan(&reactions)

	// Calculate total count
	var totalCount int64
	for _, r := range reactions {
		totalCount += r.Count
	}

	// Get user's reactions
	var userReactions []models.EventReaction
	db.Where("event_id = ? AND ip_address = ?", eventID, clientIP).Find(&userReactions)

	userReactionTypes := make([]models.ReactionType, len(userReactions))
	for i, r := range userReactions {
		userReactionTypes[i] = r.ReactionType
	}

	return models.ReactionSummary{
		EventID:       eventID,
		TotalCount:    totalCount,
		Reactions:     reactions,
		UserReactions: userReactionTypes,
	}
}

func GetEvents(c *gin.Context) {
	var events []models.Event

	db := database.GetDB()

	if err := db.Preload("Tags").Where("is_public = ?", true).Order("created_at ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	// Get client IP for user-specific reaction data
	clientIP := c.ClientIP()

	// Build response with reaction summaries
	type EventWithReactions struct {
		models.Event
		ReactionSummary models.ReactionSummary `json:"reaction_summary"`
	}

	eventsWithReactions := make([]EventWithReactions, len(events))
	for i, event := range events {
		summary := getReactionSummary(db, event.ID, clientIP)
		eventsWithReactions[i] = EventWithReactions{
			Event:           event,
			ReactionSummary: summary,
		}
	}

	c.JSON(http.StatusOK, eventsWithReactions)
}

func GetAllEvents(c *gin.Context) {
	var events []models.Event

	db := database.GetDB()
	if err := db.Preload("Tags").Order("created_at ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	// Get client IP for user-specific reaction data
	clientIP := c.ClientIP()

	// Build response with reaction summaries
	type EventWithReactions struct {
		models.Event
		ReactionSummary models.ReactionSummary `json:"reaction_summary"`
	}

	eventsWithReactions := make([]EventWithReactions, len(events))
	for i, event := range events {
		summary := getReactionSummary(db, event.ID, clientIP)
		eventsWithReactions[i] = EventWithReactions{
			Event:           event,
			ReactionSummary: summary,
		}
	}

	c.JSON(http.StatusOK, eventsWithReactions)
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
	if err := db.Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Get client IP for user-specific reaction data
	clientIP := c.ClientIP()

	// Add reaction summary
	type EventWithReactions struct {
		models.Event
		ReactionSummary models.ReactionSummary `json:"reaction_summary"`
	}

	summary := getReactionSummary(db, event.ID, clientIP)
	response := EventWithReactions{
		Event:           event,
		ReactionSummary: summary,
	}

	c.JSON(http.StatusOK, response)
}

func GetEventBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slug"})
		return
	}

	var event models.Event
	db := database.GetDB()
	if err := db.Preload("Tags").Where("slug = ?", slug).First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if the event has a public URL enabled
	if !event.HasPublicUrl {
		c.JSON(http.StatusNotFound, gin.H{"error": "This event is not publicly accessible"})
		return
	}

	// Get client IP for user-specific reaction data
	clientIP := c.ClientIP()

	// Add reaction summary
	type EventWithReactions struct {
		models.Event
		ReactionSummary models.ReactionSummary `json:"reaction_summary"`
	}

	summary := getReactionSummary(db, event.ID, clientIP)
	response := EventWithReactions{
		Event:           event,
		ReactionSummary: summary,
	}

	c.JSON(http.StatusOK, response)
}

func CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert media array to JSON string
	mediaJSON, _ := json.Marshal(req.Media)

	// Database connection
	db := database.GetDB()

	// Ensure status definition exists (auto-creates if needed)
	if _, err := models.GetOrCreateStatusDefinition(db, string(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure status definition"})
		return
	}

	// Generate unique slug
	slug := utils.GenerateUniqueSlug(db, req.Title, "events")
	if slug == "" {
		slug = fmt.Sprintf("event-%d", time.Now().Unix())
	}

	event := models.Event{
		Title:   req.Title,
		Slug:    slug,
		Media:   string(mediaJSON),
		Status:  req.Status,
		Date:    req.Date,
		Content: req.Content,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	// Associate tags with the event
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		if err := db.Find(&tags, req.TagIDs).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag IDs"})
			return
		}
		if err := db.Model(&event).Association("Tags").Replace(tags); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate tags"})
			return
		}
	}

	// Reload event with tags for response
	if err := db.Preload("Tags").First(&event, event.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload event"})
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
	if err := db.Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Store the original status to detect changes
	originalStatus := event.Status

	// Update fields if provided
	if req.Title != nil {
		event.Title = *req.Title
		// Regenerate slug when title changes
		newSlug := utils.GenerateUniqueSlug(db, *req.Title, "events", event.ID)
		if newSlug == "" {
			newSlug = fmt.Sprintf("event-%d", time.Now().Unix())
		}
		event.Slug = newSlug
	}
	if req.TagIDs != nil {
		var tags []models.Tag
		// Only query for tags if we have IDs to find
		if len(*req.TagIDs) > 0 {
			if err := db.Find(&tags, *req.TagIDs).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag IDs"})
				return
			}
		}
		// Replace with the tags (empty array if no tag IDs provided)
		if err := db.Model(&event).Association("Tags").Replace(tags); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags"})
			return
		}
	}
	if req.Media != nil {
		// Clean up old media files that are no longer referenced
		if event.Media != "" {
			// Parse old media URLs
			var oldMediaURLs []string
			if err := json.Unmarshal([]byte(event.Media), &oldMediaURLs); err == nil {
				// Find URLs that are in old media but not in new media
				newMediaURLs := req.Media
				removedURLs := []string{}

				for _, oldURL := range oldMediaURLs {
					found := false
					for _, newURL := range newMediaURLs {
						if oldURL == newURL {
							found = true
							break
						}
					}
					if !found {
						removedURLs = append(removedURLs, oldURL)
					}
				}

				// Clean up only the removed URLs
				for _, removedURL := range removedURLs {
					if err := deleteImageFromURL(removedURL); err != nil {
						fmt.Printf("Warning: Failed to cleanup removed media file %s for event %d: %v\n", removedURL, eventID, err)
					}
				}
			}
		}
		mediaJSON, _ := json.Marshal(req.Media)
		event.Media = string(mediaJSON)
	}
	if req.Status != nil {
		event.Status = *req.Status
		// Ensure status definition exists for new status value
		if _, err := models.GetOrCreateStatusDefinition(db, string(event.Status)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure status definition"})
			return
		}
	}
	if req.Date != nil {
		event.Date = *req.Date
	}
	if req.Content != nil {
		// Clean up images that were removed from content
		if event.Content != "" && event.Content != *req.Content {
			oldImages := extractImagesFromContent(event.Content)
			newImages := extractImagesFromContent(*req.Content)

			// Find images that are in old content but not in new content
			for _, oldURL := range oldImages {
				found := false
				for _, newURL := range newImages {
					if oldURL == newURL {
						found = true
						break
					}
				}
				if !found {
					if err := deleteImageFromURL(oldURL); err != nil {
						fmt.Printf("Warning: Failed to cleanup removed content image %s for event %d: %v\n", oldURL, eventID, err)
					}
				}
			}
		}
		event.Content = *req.Content
	}
	// Order field removed

	if err := db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	// Trigger newsletter automation if status changed
	if req.Status != nil && originalStatus != event.Status {
		go func() {
			automationService := services.NewNewsletterAutomationService()
			if err := automationService.ProcessStatusChange(event.ID, originalStatus, event.Status); err != nil {
				fmt.Printf("Newsletter automation error for event %d: %v\n", event.ID, err)
			}
		}()
	}

	// Reload event with tags for response
	if err := db.Preload("Tags").First(&event, event.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload event"})
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

	// First, get the event to access its media files before deletion
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Clean up associated media files
	if event.Media != "" {
		if err := cleanupMediaFiles(event.Media); err != nil {
			// Log the error but don't fail the deletion
			fmt.Printf("Warning: Failed to cleanup media files for event %d: %v\n", eventID, err)
		}
	}

	// Clean up images in content (from TipTap editor)
	if event.Content != "" {
		if err := cleanupContentImages(event.Content); err != nil {
			// Log the error but don't fail the deletion
			fmt.Printf("Warning: Failed to cleanup content images for event %d: %v\n", eventID, err)
		}
	}

	// Delete the event from database
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

	// Voting now allowed for all statuses (restriction removed)

	// Check if this IP has already voted for this event using count to avoid error logging
	var voteCount int64
	db.Model(&models.Vote{}).Where("event_id = ? AND ip_address = ?", eventID, clientIP).Count(&voteCount)
	if voteCount > 0 {
		// User has already voted, get the existing vote for deletion
		var existingVote models.Vote
		db.Where("event_id = ? AND ip_address = ?", eventID, clientIP).First(&existingVote)
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

	// Check if this IP has voted for this event using count to avoid error logging
	var voteCount int64
	db.Model(&models.Vote{}).Where("event_id = ? AND ip_address = ?", eventID, clientIP).Count(&voteCount)
	hasVoted := voteCount > 0

	c.JSON(http.StatusOK, gin.H{
		"voted": hasVoted,
		"votes": event.Votes,
	})
}

// getStatusForCategory finds the first status mapped to a given category
func getStatusForCategory(db *gorm.DB, categoryID string, themeID string) (string, error) {
	var mapping models.StatusCategoryMapping
	err := db.Where("category_id = ? AND theme_id = ?", categoryID, themeID).
		First(&mapping).Error

	if err != nil {
		return "", err
	}

	// Get the status definition to get the display name
	var statusDef models.EventStatusDefinition
	err = db.First(&statusDef, mapping.StatusDefinitionID).Error
	if err != nil {
		return "", err
	}

	return statusDef.DisplayName, nil
}

// SubmitFeedback allows public users to submit feedback
func SubmitFeedback(c *gin.Context) {
	var req struct {
		Title         string `json:"title" binding:"required"`
		Content       string `json:"content" binding:"required"`
		FormStartTime int64  `json:"form_start_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Time validation
	now := time.Now().UnixMilli()
	formDuration := now - req.FormStartTime

	// Minimum time check (3 seconds)
	if formDuration < 3000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please take your time to fill out the form properly.",
		})
		return
	}

	// Maximum time check (30 minutes)
	if formDuration > 30*60*1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Form session expired. Please refresh and try again.",
		})
		return
	}

	// Create feedback event
	mediaJSON, _ := json.Marshal([]string{})

	// Get the next order value for backlog status
	db := database.GetDB()

	// Generate unique slug
	slug := utils.GenerateUniqueSlug(db, req.Title, "events")
	if slug == "" {
		slug = fmt.Sprintf("feedback-%d", time.Now().Unix())
	}

	// Get current theme and find status mapped to feedback category
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	// Determine which status to use for feedback
	feedbackStatus := models.EventStatus("Feedback") // default status for feedback submissions
	if settings.CurrentThemeID != "" {
		if status, err := getStatusForCategory(db, "feedback", settings.CurrentThemeID); err == nil {
			feedbackStatus = models.EventStatus(status)
		}
	}

	event := models.Event{
		Title:   req.Title,
		Slug:    slug,
		Media:   string(mediaJSON),
		Status:  feedbackStatus,
		Date:    "",
		Content: req.Content,
	}

	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	// Associate "Feedback" tag with the event
	var feedbackTag models.Tag
	if err := db.Where("name = ?", "Feedback").First(&feedbackTag).Error; err != nil {
		// Create feedback tag if it doesn't exist
		feedbackTag = models.Tag{
			Name:  "Feedback",
			Color: "#F59E0B", // Yellow color
		}
		if err := db.Create(&feedbackTag).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback tag"})
			return
		}
	}

	if err := db.Model(&event).Association("Tags").Append(&feedbackTag); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to associate feedback tag with event %d: %v\n", event.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feedback submitted successfully",
		"id":      event.ID,
	})
}
