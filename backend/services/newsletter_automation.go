package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"shipshipship/constants"
	"shipshipship/database"
	"shipshipship/email"
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

// getBaseURL returns the base URL from BASE_URL env var
func (nas *NewsletterAutomationService) getBaseURL() string {
	// Try environment variable
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}

	// Fallback to empty (will use relative URLs)
	return ""
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

	// Get branding settings with base URL
	baseURL := nas.getBaseURL()
	branding, err := models.GetBrandingSettingsWithBaseURL(nas.db, baseURL)
	if err != nil {
		return fmt.Errorf("failed to get branding settings: %v", err)
	}

	// Get the email template from database or use default
	template, err := models.GetEmailTemplate(nas.db, "event")
	if err != nil {
		// If template doesn't exist, use default content
		log.Printf("Template not found in DB, using defaults for type: event")
		defaultTemplate := constants.GetTemplateByType("event")
		if defaultTemplate == nil {
			log.Printf("ERROR: No default template found for type: event")
			return fmt.Errorf("no template found for automated newsletter")
		}

		log.Printf("Using default template for automated newsletter")
		// Create a temporary template object with defaults
		template = &models.EmailTemplate{
			Type:    defaultTemplate.Type,
			Subject: defaultTemplate.Subject,
			Content: defaultTemplate.Content,
		}
	}

	// Generate the email content with variable replacements
	subject, content, err := email.GenerateEmailContent(nas.db, template, &event, &statusDef, branding)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %v", err)
	}

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
		// Personalize unsubscribe URL for each subscriber (use BaseURL, not ProjectURL)
		unsubscribeURL := fmt.Sprintf("%s/unsubscribe?email=%s", branding.BaseURL, subscriber.Email)
		if branding.BaseURL == "" {
			unsubscribeURL = fmt.Sprintf("/unsubscribe?email=%s", subscriber.Email)
		}
		personalizedContent := strings.ReplaceAll(content, "{{unsubscribe_url}}", unsubscribeURL)

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
		EmailTemplate:   template.Type,
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
				EmailTemplate:   template.Type,
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
			"email_template":   template.Type,
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

// Legacy helper functions below are kept for backward compatibility
// but are now available as shared utilities in the utils package
