package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"shipshipship/database"
	"shipshipship/models"

	"gorm.io/gorm"
)

// NewsletterAutomationService handles automated newsletter sending
type NewsletterAutomationService struct {
	db           *gorm.DB
	emailService *EmailService
}

// NewNewsletterAutomationService creates a new newsletter automation service
func NewNewsletterAutomationService() *NewsletterAutomationService {
	return &NewsletterAutomationService{
		db:           database.GetDB(),
		emailService: NewEmailService(),
	}
}

// StatusTemplateMapping removed: generic template used for all statuses
// (legacy constants still referenced elsewhere for backward compatibility)
// Placeholder kept to preserve line numbers.

// ProcessStatusChange checks if automation should be triggered and sends newsletters
func (nas *NewsletterAutomationService) ProcessStatusChange(eventID uint, oldStatus, newStatus models.EventStatus) error {
	// Skip if status hasn't actually changed
	if oldStatus == newStatus {
		return nil
	}

	log.Printf("Processing status change for event %d: %s -> %s", eventID, oldStatus, newStatus)

	// Safety check: prevent automation for rapid successive changes
	// Check if an email was sent for this event in the last 30 seconds
	var recentEmailCount int64
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)
	nas.db.Model(&models.EventEmailHistory{}).
		Where("event_id = ? AND sent_at > ?", eventID, thirtySecondsAgo).
		Count(&recentEmailCount)

	if recentEmailCount > 0 {
		log.Printf("Skipping automation for event %d - email sent recently (within 30 seconds)", eventID)
		return nil
	}

	// Get automation settings
	automationSettings, err := models.GetOrCreateAutomationSettings(nas.db)
	if err != nil {
		return fmt.Errorf("failed to get automation settings: %v", err)
	}

	// Check if automation is enabled
	if !automationSettings.Enabled {
		log.Printf("Newsletter automation is disabled, skipping event %d", eventID)
		return nil
	}

	// Parse trigger statuses
	var triggerStatuses []string
	if err := json.Unmarshal([]byte(automationSettings.TriggerStatuses), &triggerStatuses); err != nil {
		return fmt.Errorf("failed to parse trigger statuses: %v", err)
	}

	// Check if new status is in trigger list
	shouldTrigger := false
	for _, status := range triggerStatuses {
		if string(newStatus) == status {
			shouldTrigger = true
			break
		}
	}

	if !shouldTrigger {
		log.Printf("Status %s is not in trigger list for event %d, skipping", newStatus, eventID)
		return nil
	}

	log.Printf("Triggering automated newsletter for event %d with status %s", eventID, newStatus)

	// Send automated newsletter
	return nas.sendAutomatedNewsletter(eventID, newStatus)
}

// sendAutomatedNewsletter sends a newsletter for an event based on its status
func (nas *NewsletterAutomationService) sendAutomatedNewsletter(eventID uint, status models.EventStatus) error {
	// Get the event with tags
	var event models.Event
	if err := nas.db.Preload("Tags").First(&event, eventID).Error; err != nil {
		return fmt.Errorf("failed to get event: %v", err)
	}

	// Get status definition for color
	var statusDef models.EventStatusDefinition
	if err := nas.db.Where("display_name = ?", event.Status).First(&statusDef).Error; err != nil {
		return fmt.Errorf("failed to get status definition: %v", err)
	}

	// Generic automated newsletter (no per-status template mapping)
	// Get branding settings
	branding, err := models.GetBrandingSettings(nas.db)
	if err != nil {
		return fmt.Errorf("failed to get branding settings: %v", err)
	}

	// Build subject: {status} : {title} - {project_name}
	subject := fmt.Sprintf("%s: %s - %s", statusDef.DisplayName, event.Title, branding.ProjectName)

	// Format date (no 'Estimated' badge logic)
	formattedDate := nas.formatDate(event.Date)
	dateHTML := ""
	if formattedDate != "" {
		dateHTML = fmt.Sprintf(`<div style="margin-bottom:8px;color:#6b7280;font-size:14px;font-weight:500;">%s</div>`, formattedDate)
	}

	// Tags HTML
	tagsHTML := nas.generateTagsHTML(event.Tags)

	// Event content with absolute URLs
	eventContent := nas.convertRelativeUrlsToAbsolute(event.Content, branding.ProjectURL)

	// Use default primary color since it's no longer in settings
	primaryColor := "#3b82f6"

	// Email content with status name
	content := fmt.Sprintf(`<body style="font-family:Arial,sans-serif;line-height:1.6;color:#333;max-width:600px;margin:0 auto;padding:20px;">
<h1 style="color:#3B82F6;text-align:center;font-size:28px;font-weight:bold;margin:20px 0;">%s</h1>
<div style="padding:20px;margin-bottom:20px;">
  <h2 style="color:#000;margin-top:0;font-size:48px;font-weight:bold;margin-bottom:15px;text-align:center;">%s</h2>
  <div style="margin-bottom:20px;">
    %s
    <div style="display:flex;flex-wrap:wrap;gap:6px;align-items:center;">%s</div>
  </div>
  <div style="margin:15px 0;font-size:16px;line-height:1.6;">%s</div>
  <div style="text-align:center;margin-top:30px;">
    <a href="%s/%s" style="background:%s;color:white;padding:14px 28px;text-decoration:none;border-radius:6px;display:inline-block;font-weight:bold;font-size:16px;">See Details</a>
  </div>
</div>
<hr style="border:none;border-top:1px solid #eee;margin:30px 0;">
<div style="text-align:center;font-size:12px;color:#666;">
  <p style="margin:5px 0;">
    <a href="%s" style="color:#2563eb;text-decoration:none;">%s</a>
    <br><a href="{{unsubscribe_url}}" style="color:#2563eb;text-decoration:none;">Unsubscribe</a>
  </p>
</div>
</body>`,
		statusDef.DisplayName,
		event.Title,
		dateHTML,
		tagsHTML,
		eventContent,
		branding.ProjectURL,
		event.Slug,
		primaryColor,
		branding.ProjectURL,
		branding.ProjectName,
	)

	// Get active newsletter subscribers
	subscribers, err := models.GetActiveNewsletterSubscribers(nas.db)
	if err != nil {
		return fmt.Errorf("failed to get newsletter subscribers: %v", err)
	}

	if len(subscribers) == 0 {
		log.Printf("No active newsletter subscribers found for event %d", eventID)
		return nil
	}

	// Send emails to all subscribers
	sentCount := 0
	var sendErrors []string

	for _, subscriber := range subscribers {
		// Personalize unsubscribe URL for each subscriber
		personalizedContent := strings.ReplaceAll(content, "{{unsubscribe_url}}",
			fmt.Sprintf("%s/unsubscribe?email=%s", branding.ProjectURL, subscriber.Email))

		err := nas.emailService.SendEmail(subscriber.Email, subject, personalizedContent)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to send to %s: %v", subscriber.Email, err)
			sendErrors = append(sendErrors, errorMsg)
			log.Printf("Newsletter automation error: %s", errorMsg)
			continue
		}
		sentCount++
	}

	// Create email history record
	now := time.Now()
	historyRecord := &models.EventEmailHistory{
		EventID:         eventID,
		EventStatus:     string(status),
		EmailSubject:    subject,
		EmailTemplate:   "generic",
		SubscriberCount: sentCount,
		SentAt:          now,
	}

	if err := nas.db.Create(historyRecord).Error; err != nil {
		log.Printf("Failed to save email history for automated newsletter: %v", err)
		// Don't return error as emails were already sent
	}

	// Update or create publication record for backward compatibility
	var publication models.EventPublication
	err = nas.db.Where("event_id = ?", eventID).First(&publication).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new publication record
			publication = models.EventPublication{
				EventID:         eventID,
				EmailSent:       true,
				EmailSubject:    subject,
				EmailContent:    content,
				EmailTemplate:   "generic",
				EmailSentAt:     &now,
				SubscriberCount: sentCount,
			}
			if err := nas.db.Create(&publication).Error; err != nil {
				log.Printf("Failed to create publication record for automated newsletter: %v", err)
			}
		} else {
			log.Printf("Failed to query publication record: %v", err)
		}
	} else {
		// Update existing publication record
		updates := map[string]interface{}{
			"email_sent":       true,
			"email_subject":    subject,
			"email_content":    content,
			"email_template":   "generic",
			"email_sent_at":    &now,
			"subscriber_count": sentCount,
		}
		if err := nas.db.Model(&publication).Updates(updates).Error; err != nil {
			log.Printf("Failed to update publication record for automated newsletter: %v", err)
		}
	}

	log.Printf("Automated newsletter sent successfully for event %d: %d/%d emails sent",
		eventID, sentCount, len(subscribers))

	// Log any send errors but don't fail the operation
	if len(sendErrors) > 0 {
		log.Printf("Some emails failed to send: %v", sendErrors)
	}

	return nil
}

// generateGenericEmailContent kept for backward compatibility (not used by automation anymore)
// Can be removed later after confirming no manual callers rely on it.
func (nas *NewsletterAutomationService) generateGenericEmailContent(event *models.Event, branding *models.BrandingSettings) (string, string, error) {
	// Get status definition for color
	var statusDef models.EventStatusDefinition
	if err := nas.db.Where("display_name = ?", event.Status).First(&statusDef).Error; err != nil {
		return "", "", fmt.Errorf("failed to get status definition: %v", err)
	}

	// Use default primary color since it's no longer in settings
	primaryColor := "#3b82f6"
	subject := fmt.Sprintf("%s: %s - %s", statusDef.DisplayName, event.Title, branding.ProjectName)
	formattedDate := nas.formatDate(event.Date)
	dateHTML := ""
	if formattedDate != "" {
		dateHTML = fmt.Sprintf(`<div style="margin-bottom:8px;color:#6b7280;font-size:14px;">%s</div>`, formattedDate)
	}
	tagsHTML := nas.generateTagsHTML(event.Tags)
	eventContent := nas.convertRelativeUrlsToAbsolute(event.Content, branding.ProjectURL)
	// Generic email content with status name
	content := fmt.Sprintf(`<body style="font-family:Arial,sans-serif;line-height:1.6;color:#333;max-width:600px;margin:0 auto;padding:20px;">
<h1 style="color:#3B82F6;text-align:center;font-size:28px;font-weight:bold;margin:20px 0;">%s</h1>
<div style="padding:20px;margin-bottom:20px;">
  <h2 style="color:#000;margin-top:0;font-size:48px;font-weight:bold;margin-bottom:15px;text-align:center;">%s</h2>
  <div style="margin-bottom:20px;">
    %s
    <div style="display:flex;flex-wrap:wrap;gap:6px;align-items:center;">%s</div>
  </div>
  <div style="margin:15px 0;font-size:16px;line-height:1.6;">%s</div>
  <div style="text-align:center;margin-top:30px;">
    <a href="%s/%s" style="background:%s;color:white;padding:14px 28px;text-decoration:none;border-radius:6px;display:inline-block;font-weight:bold;font-size:16px;">See Details</a>
  </div>
</div>
<hr style="border:none;border-top:1px solid #eee;margin:30px 0;">
<div style="text-align:center;font-size:12px;color:#666;">
  <p style="margin:5px 0;">
    <a href="%s" style="color:#2563eb;text-decoration:none;">%s</a>
    <br><a href="{{unsubscribe_url}}" style="color:#2563eb;text-decoration:none;">Unsubscribe</a>
  </p>
</div>
</body>`,
		statusDef.DisplayName,
		event.Title,
		dateHTML,
		tagsHTML,
		eventContent,
		branding.ProjectURL,
		event.Slug,
		primaryColor,
		branding.ProjectURL,
		branding.ProjectName,
	)
	return subject, content, nil
}

// formatDate formats a date string to match the public page format (e.g., "10 Aug. 2025")
func (nas *NewsletterAutomationService) formatDate(dateString string) string {
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
func (nas *NewsletterAutomationService) generateTagsHTML(tags []models.Tag) string {
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

// convertRelativeUrlsToAbsolute converts relative image URLs to absolute URLs for email compatibility
func (nas *NewsletterAutomationService) convertRelativeUrlsToAbsolute(content, baseURL string) string {
	// Replace relative image URLs like /api/uploads/... with absolute URLs
	return strings.ReplaceAll(content, `src="/api/uploads/`, fmt.Sprintf(`src="%s/api/uploads/`, baseURL))
}
