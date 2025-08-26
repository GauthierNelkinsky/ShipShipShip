package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"shipshipship/constants"
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

// StatusTemplateMapping maps event status to email template type
var StatusTemplateMapping = map[models.EventStatus]string{
	models.StatusProposed: constants.TemplateTypeProposedFeature,
	models.StatusUpcoming: constants.TemplateTypeUpcomingFeature,
	models.StatusRelease:  constants.TemplateTypeNewRelease,
}

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

	// Determine template type based on status
	templateType, exists := StatusTemplateMapping[status]
	if !exists {
		// Fallback to proposed feature template for unmapped statuses
		log.Printf("Warning: No template mapping for status %s, using proposed feature template", status)
		templateType = constants.TemplateTypeProposedFeature
	}

	// Get email template or use default
	template, err := models.GetEmailTemplate(nas.db, templateType)
	if err != nil {
		// Use default template if custom one doesn't exist
		defaultTemplate := constants.GetTemplateByType(templateType)
		if defaultTemplate == nil {
			return fmt.Errorf("no default template found for type: %s", templateType)
		}

		template = &models.EmailTemplate{
			Type:    defaultTemplate.Type,
			Subject: defaultTemplate.Subject,
			Content: defaultTemplate.Content,
		}
	}

	// Get branding settings
	branding, err := models.GetBrandingSettings(nas.db)
	if err != nil {
		return fmt.Errorf("failed to get branding settings: %v", err)
	}

	// Generate email content with variable replacements
	subject, content, err := nas.generateEmailContent(template, &event, branding)
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
		EmailTemplate:   templateType,
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
				EmailTemplate:   templateType,
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
			"email_template":   templateType,
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

// generateEmailContent generates email subject and content with variable replacements
func (nas *NewsletterAutomationService) generateEmailContent(template *models.EmailTemplate, event *models.Event, branding *models.BrandingSettings) (string, string, error) {
	subject := template.Subject
	content := template.Content

	// Convert relative image URLs to absolute URLs in event content
	eventContent := nas.convertRelativeUrlsToAbsolute(event.Content, branding.ProjectURL)

	// Generate tags HTML
	tagsHTML := nas.generateTagsHTML(event.Tags)

	// Get primary color for buttons
	settings, err := models.GetOrCreateSettings(nas.db)
	if err != nil {
		return "", "", err
	}
	primaryColor := settings.PrimaryColor

	// Format date with "Estimated" badge for upcoming events
	formattedDate := nas.formatDate(event.Date)
	formattedDateHTML := ""
	if formattedDate != "" {
		isUpcoming := event.Status == models.StatusUpcoming

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
		"{{unsubscribe_url}}": "{{unsubscribe_url}}", // This gets replaced per subscriber
	}

	// Apply replacements
	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
		content = strings.ReplaceAll(content, placeholder, value)
	}

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
