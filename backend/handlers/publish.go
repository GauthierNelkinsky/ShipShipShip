package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"shipshipship/constants"
	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// formatDate formats a date string to match the public page format (e.g., "10 Aug. 2025")
func formatDate(dateString string) string {
	if dateString == "" {
		return ""
	}

	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return dateString // Return original if parsing fails
	}

	// Format as "2 Jan. 2006"
	formatted := date.Format("2 Jan 2006")
	// Add period after month abbreviation
	return strings.Replace(formatted, " "+date.Format("Jan")+" ", " "+date.Format("Jan")+". ", 1)
}

// generateTagsHTML generates HTML for tags
func generateTagsHTML(db *gorm.DB, tags []models.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	var tagHTML strings.Builder
	for i, tag := range tags {
		if i > 0 {
			tagHTML.WriteString(" ")
		}

		// Generate badge HTML with tag color
		tagHTML.WriteString(fmt.Sprintf(
			`<span style="display: inline-flex; align-items: center; border-radius: 12px; border: 1px solid %s; background-color: %s20; color: %s; padding: 2px 8px; font-size: 11px; font-weight: 600; margin-right: 6px;">%s</span>`,
			tag.Color, tag.Color, tag.Color, tag.Name,
		))
	}

	return tagHTML.String()
}

// GetEventPublishStatus gets the publication status of an event
func GetEventPublishStatus(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	db := database.GetDB()

	var event models.Event
	if err := db.Preload("Publication").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Initialize publication if it doesn't exist
	if event.Publication == nil {
		publication := &models.EventPublication{
			EventID:   uint(eventID),
			EmailSent: false,
		}
		if err := db.Create(publication).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize publication"})
			return
		}
		event.Publication = publication
	}

	c.JSON(http.StatusOK, gin.H{
		"is_public":        event.IsPublic,
		"has_public_url":   event.HasPublicUrl,
		"email_sent":       event.Publication.EmailSent,
		"email_sent_at":    event.Publication.EmailSentAt,
		"email_subject":    event.Publication.EmailSubject,
		"email_template":   event.Publication.EmailTemplate,
		"subscriber_count": event.Publication.SubscriberCount,
	})
}

// UpdateEventPublicStatus updates the public visibility of an event
func UpdateEventPublicStatus(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.EventPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	db := database.GetDB()

	// Prepare updates map
	updates := make(map[string]interface{})

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if req.HasPublicUrl != nil {
		updates["has_public_url"] = *req.HasPublicUrl
	}

	// Update the event's status
	if len(updates) > 0 {
		if err := db.Model(&models.Event{}).Where("id = ?", eventID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event status updated successfully",
		"updates": updates,
	})
}

// GetEventNewsletterPreview generates a preview of the newsletter for an event
func GetEventNewsletterPreview(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Use generic event template for all statuses
	templateType := "event"

	db := database.GetDB()

	// Get the event with tags and status definition preloaded
	var event models.Event
	if err := db.Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Get status definition for color
	var statusDef models.EventStatusDefinition
	if err := db.Where("display_name = ?", event.Status).First(&statusDef).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status definition not found"})
		return
	}

	// Get the email template or use defaults if not found
	template, err := models.GetEmailTemplate(db, templateType)
	if err != nil {
		// If template doesn't exist, use default content directly
		log.Printf("Template not found in DB, using defaults for type: %s", templateType)
		defaultTemplate := constants.GetTemplateByType(templateType)
		if defaultTemplate == nil {
			log.Printf("ERROR: No default template found for type: %s", templateType)
			// List available templates
			allTemplates := constants.GetDefaultTemplates()
			log.Printf("Available templates: %v", allTemplates)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template type"})
			return
		}

		log.Printf("Using default template: %s", defaultTemplate.Type)
		// Create a temporary template object with defaults
		template = &models.EmailTemplate{
			Type:    defaultTemplate.Type,
			Subject: defaultTemplate.Subject,
			Content: defaultTemplate.Content,
		}
	}

	// Get branding settings for project info
	branding, err := models.GetBrandingSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branding settings"})
		return
	}

	// Generate the preview with variable replacements
	subject, content, err := generateEmailContent(db, template, &event, &statusDef, branding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate email content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subject": subject,
		"content": content,
	})
}

// SendEventNewsletter sends a newsletter for an event
func SendEventNewsletter(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.EventNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Use generic event template
	req.Template = "event"

	db := database.GetDB()

	// Get the event with tags and publication preloaded
	var event models.Event
	if err := db.Preload("Publication").Preload("Tags").First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Note: We allow resending emails, but track the history

	// Get newsletter subscribers
	subscribers, err := models.GetActiveNewsletterSubscribers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter subscribers"})
		return
	}

	if len(subscribers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active newsletter subscribers found"})
		return
	}

	// Get branding settings
	branding, err := models.GetBrandingSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branding settings"})
		return
	}

	// Send emails to all subscribers
	emailService := services.NewEmailService()
	sentCount := 0

	for _, subscriber := range subscribers {
		// Replace unsubscribe URL in content
		personalizedContent := strings.ReplaceAll(req.Content, "{{unsubscribe_url}}",
			fmt.Sprintf("%s/unsubscribe?email=%s", branding.ProjectURL, subscriber.Email))

		err := emailService.SendEmail(subscriber.Email, req.Subject, personalizedContent)
		if err != nil {
			// Log the error but continue sending to other subscribers
			fmt.Printf("Failed to send email to %s: %v\n", subscriber.Email, err)
			continue
		}
		sentCount++
	}

	// Create email history record
	now := time.Now()
	historyRecord := &models.EventEmailHistory{
		EventID:         uint(eventID),
		EventStatus:     string(event.Status),
		EmailSubject:    req.Subject,
		EmailTemplate:   req.Template,
		SubscriberCount: sentCount,
		SentAt:          now,
	}
	if err := db.Create(historyRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save email history"})
		return
	}

	// Newsletter history now uses EventEmailHistory directly (see newsletter.go handler)

	// Update or create publication record (for backward compatibility)
	if event.Publication == nil {
		publication := &models.EventPublication{
			EventID:         uint(eventID),
			EmailSent:       true,
			EmailSubject:    req.Subject,
			EmailContent:    req.Content,
			EmailTemplate:   req.Template,
			EmailSentAt:     &now,
			SubscriberCount: sentCount,
		}
		if err := db.Create(publication).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save publication record"})
			return
		}
	} else {
		updates := map[string]interface{}{
			"email_sent":       true,
			"email_subject":    req.Subject,
			"email_content":    req.Content,
			"email_template":   req.Template,
			"email_sent_at":    &now,
			"subscriber_count": sentCount,
		}
		if err := db.Model(event.Publication).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update publication record"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Newsletter sent successfully",
		"subscribers_sent":  sentCount,
		"total_subscribers": len(subscribers),
	})
}

// generateEmailContent generates email subject and content with variable replacements
func generateEmailContent(db *gorm.DB, template *models.EmailTemplate, event *models.Event, statusDef *models.EventStatusDefinition, branding *models.BrandingSettings) (string, string, error) {
	subject := template.Subject
	content := template.Content

	// Convert relative image URLs to absolute URLs in event content
	eventContent := convertRelativeUrlsToAbsolute(event.Content, branding.ProjectURL)

	// Generate tags HTML
	tagsHTML := generateTagsHTML(db, event.Tags)

	// Get primary color for buttons
	settings, err := models.GetOrCreateSettings(db)
	if err != nil {
		return "", "", err
	}
	primaryColor := settings.PrimaryColor

	// Format date with "Estimated" badge for upcoming events
	formattedDate := formatDate(event.Date)
	formattedDateHTML := ""
	if formattedDate != "" {
		isUpcoming := event.Status == "Upcoming"

		if isUpcoming {
			formattedDateHTML = `<span style="color: #d97706; background: #fef3c7; font-weight: 500; font-size: 11px; padding: 4px 8px; border-radius: 4px; display: inline-flex; align-items: center; margin-right: 8px;">Estimated</span><span style="color: #6b7280; font-size: 14px; font-weight: 500;">` + formattedDate + `</span>`
		} else {
			formattedDateHTML = `<span style="color: #6b7280; font-size: 14px; font-weight: 500;">` + formattedDate + `</span>`
		}
	}

	// Replace common variables
	replacements := map[string]string{
		"{{project_name}}":    branding.ProjectName,
		"{{project_url}}":     branding.ProjectURL,
		"{{event_name}}":      event.Title,
		"{{event_url}}":       fmt.Sprintf("%s/%s", branding.ProjectURL, event.Slug),
		"{{event_content}}":   eventContent,
		"{{event_date}}":      formattedDateHTML,
		"{{event_tags}}":      tagsHTML,
		"{{primary_color}}":   primaryColor,
		"{{status}}":          statusDef.DisplayName,
		"{{unsubscribe_url}}": fmt.Sprintf("%s/unsubscribe", branding.ProjectURL),
	}

	// Apply replacements
	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	return subject, content, nil
}

// GetEventEmailHistory returns the email sending history for an event
func GetEventEmailHistory(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	db := database.GetDB()

	// Get email history for the event
	var history []models.EventEmailHistory
	if err := db.Where("event_id = ?", eventID).Order("sent_at DESC").Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// convertRelativeUrlsToAbsolute converts relative image URLs to absolute URLs for email compatibility
func convertRelativeUrlsToAbsolute(content, baseURL string) string {
	// Replace relative image URLs like /api/uploads/... with absolute URLs
	re := regexp.MustCompile(`src="(/api/uploads/[^"]*)"`)
	return re.ReplaceAllString(content, fmt.Sprintf(`src="%s$1"`, baseURL))
}
