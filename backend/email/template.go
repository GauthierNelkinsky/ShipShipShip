package email

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"shipshipship/models"

	"gorm.io/gorm"
)

// FormatDate formats a date string to match the public page format (e.g., "10 Aug. 2025")
func FormatDate(dateString string) string {
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

// GenerateTagsHTML generates HTML for tags
func GenerateTagsHTML(tags []models.Tag) string {
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

// ConvertRelativeUrlsToAbsolute converts relative image URLs to absolute URLs for email compatibility
func ConvertRelativeUrlsToAbsolute(content, baseURL string) string {
	// If baseURL is empty, return content unchanged (keep relative URLs)
	if baseURL == "" {
		return content
	}
	// Replace relative image URLs like /api/uploads/... with absolute URLs
	re := regexp.MustCompile(`src="(/api/uploads/[^"]*)"`)
	return re.ReplaceAllString(content, fmt.Sprintf(`src="%s$1"`, baseURL))
}

// GenerateEmailContent generates email subject and content with variable replacements
func GenerateEmailContent(db *gorm.DB, template *models.EmailTemplate, event *models.Event, statusDef *models.EventStatusDefinition, branding *models.BrandingSettings) (string, string, error) {
	subject := template.Subject
	content := template.Content

	// Convert relative image URLs to absolute URLs in event content
	eventContent := ConvertRelativeUrlsToAbsolute(event.Content, branding.BaseURL)

	// Generate tags HTML
	tagsHTML := GenerateTagsHTML(event.Tags)

	// Format date
	formattedDate := FormatDate(event.Date)
	formattedDateHTML := ""
	if formattedDate != "" {
		formattedDateHTML = `<span style="color: #6b7280; font-size: 14px; font-weight: 500;">` + formattedDate + `</span>`
	}

	// Generate URLs (use BaseURL for event links, or relative URLs if empty)
	eventURL := fmt.Sprintf("%s/%s", branding.BaseURL, event.Slug)
	if branding.BaseURL == "" {
		eventURL = fmt.Sprintf("/%s", event.Slug)
	}

	// Use BaseURL for unsubscribe (not ProjectURL which is the external website)
	unsubscribeURL := fmt.Sprintf("%s/unsubscribe", branding.BaseURL)
	if branding.BaseURL == "" {
		unsubscribeURL = "/unsubscribe"
	}

	// Replace common variables
	replacements := map[string]string{
		"{{project_name}}":    branding.ProjectName,
		"{{project_url}}":     branding.ProjectURL,
		"{{event_name}}":      event.Title,
		"{{event_url}}":       eventURL,
		"{{event_content}}":   eventContent,
		"{{event_date}}":      formattedDateHTML,
		"{{event_tags}}":      tagsHTML,
		"{{status}}":          statusDef.DisplayName,
		"{{unsubscribe_url}}": unsubscribeURL,
	}

	// Apply replacements
	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	return subject, content, nil
}
