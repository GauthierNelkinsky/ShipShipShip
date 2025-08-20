package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	"shipshipship/constants"
	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SubscribeToNewsletter handles newsletter subscription requests
func SubscribeToNewsletter(c *gin.Context) {
	var req models.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	db := database.GetDB()

	// Check if user is already subscribed
	existingSubscriber, err := models.FindSubscriberByEmail(db, req.Email)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":            "You are already subscribed to our newsletter",
			"email":              existingSubscriber.Email,
			"already_subscribed": true,
		})
		return
	}

	subscriber, err := models.Subscribe(db, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe to newsletter"})
		return
	}

	// Send welcome email (don't fail subscription if email fails)
	go func() {
		if err := sendWelcomeEmail(db, subscriber.Email); err != nil {
			fmt.Printf("Failed to send welcome email to %s: %v\n", subscriber.Email, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":            "Successfully subscribed to newsletter",
		"email":              subscriber.Email,
		"already_subscribed": false,
	})
}

// UnsubscribeFromNewsletter handles newsletter unsubscription requests
func UnsubscribeFromNewsletter(c *gin.Context) {
	var req models.UnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	db := database.GetDB()
	err := models.Unsubscribe(db, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe from newsletter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unsubscribed from newsletter",
	})
}

// GetNewsletterStats returns newsletter subscription statistics
func GetNewsletterStats(c *gin.Context) {
	db := database.GetDB()
	count, err := models.GetActiveSubscriberCount(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_subscribers": count,
	})
}

// GetNewsletterSubscribers returns all newsletter subscribers (admin only)
func GetNewsletterSubscribers(c *gin.Context) {
	db := database.GetDB()
	var subscribers []models.NewsletterSubscriber

	if err := db.Order("created_at DESC").Find(&subscribers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter subscribers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       len(subscribers),
	})
}

// CheckSubscriptionStatus checks if an email is subscribed to the newsletter
func CheckSubscriptionStatus(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	db := database.GetDB()
	_, err := models.FindSubscriberByEmail(db, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"subscribed": false,
				"active":     false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribed": true,
		"active":     true,
	})
}

// GetNewsletterHistory returns newsletter history with pagination (admin only)
// Now returns event email history to show all newsletters sent by events
func GetNewsletterHistory(c *gin.Context) {
	db := database.GetDB()

	// Parse pagination parameters
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get event email history instead of newsletter history
	var eventHistory []models.EventEmailHistory
	var total int64

	// Count total records
	if err := db.Model(&models.EventEmailHistory{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count newsletter history"})
		return
	}

	// Get paginated records
	offset := (page - 1) * limit
	if err := db.Order("sent_at DESC").Offset(offset).Limit(limit).Find(&eventHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter history"})
		return
	}

	// Transform event email history to match expected newsletter format
	newsletters := make([]map[string]interface{}, len(eventHistory))
	for i, email := range eventHistory {
		newsletters[i] = map[string]interface{}{
			"id":              email.ID,
			"subject":         email.EmailSubject,
			"content":         "", // Don't expose full content in list
			"status":          "sent",
			"recipient_count": email.SubscriberCount,
			"open_count":      0, // Event emails don't track opens yet
			"click_count":     0, // Event emails don't track clicks yet
			"sent_at":         email.SentAt,
			"created_at":      email.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"newsletters": newsletters,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetNewsletterSubscribersPaginated returns paginated newsletter subscribers (admin only)
func GetNewsletterSubscribersPaginated(c *gin.Context) {
	db := database.GetDB()

	// Parse pagination parameters
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	subscribers, total, err := models.GetSubscribersPaginated(db, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get newsletter subscribers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers": subscribers,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// sendWelcomeEmail sends a welcome email to new newsletter subscribers
func sendWelcomeEmail(db *gorm.DB, email string) error {
	// Get mail settings
	mailSettings, err := models.GetOrCreateMailSettings(db)
	if err != nil || mailSettings.SMTPHost == "" || mailSettings.FromEmail == "" {
		return fmt.Errorf("mail settings not configured")
	}

	// Get project settings for project name
	projectSettings, err := models.GetOrCreateSettings(db)
	if err != nil {
		return fmt.Errorf("failed to get project settings: %v", err)
	}

	// Replace variables in template
	projectName := projectSettings.Title
	if projectName == "" {
		projectName = "ShipShipShip"
	}

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/unsubscribe?email=%s", email)

	// Get welcome email template and subject (check for custom template first)
	welcomeTemplate := getWelcomeEmailTemplate()
	welcomeSubject := fmt.Sprintf("Welcome to %s!", projectName)

	if customTemplate, err := models.GetEmailTemplate(db, "welcome"); err == nil {
		welcomeTemplate = customTemplate.Content
		welcomeSubject = strings.ReplaceAll(customTemplate.Subject, "{{project_name}}", projectName)
	} else if err != gorm.ErrRecordNotFound {
		// Log only unexpected errors, not "record not found"
		fmt.Printf("Warning: Failed to load custom welcome template: %v\n", err)
	}

	projectURL := projectSettings.WebsiteURL
	if projectURL == "" {
		projectURL = "http://localhost:8080"
	}

	content := strings.ReplaceAll(welcomeTemplate, "{{project_name}}", projectName)
	content = strings.ReplaceAll(content, "{{project_url}}", projectURL)
	content = strings.ReplaceAll(content, "{{unsubscribe_url}}", unsubscribeURL)

	// Prepare email
	fromName := mailSettings.FromName
	if fromName == "" {
		fromName = projectName
	}

	from := fmt.Sprintf("%s <%s>", fromName, mailSettings.FromEmail)

	// Prepare message
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", email)
	message += fmt.Sprintf("Subject: %s\r\n", welcomeSubject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += content

	// Determine authentication
	var auth smtp.Auth
	if mailSettings.SMTPUsername != "" {
		auth = smtp.PlainAuth("", mailSettings.SMTPUsername, mailSettings.SMTPPassword, mailSettings.SMTPHost)
	}

	// Send email based on encryption type
	addr := fmt.Sprintf("%s:%d", mailSettings.SMTPHost, mailSettings.SMTPPort)

	switch strings.ToLower(mailSettings.SMTPEncryption) {
	case "ssl":
		return utils.SendMailWithSSL(addr, auth, mailSettings.FromEmail, []string{email}, []byte(message))
	case "tls":
		return utils.SendMailWithTLS(addr, auth, mailSettings.FromEmail, []string{email}, []byte(message))
	default:
		// No encryption
		return smtp.SendMail(addr, auth, mailSettings.FromEmail, []string{email}, []byte(message))
	}
}

// getWelcomeEmailTemplate returns the default welcome email template
func getWelcomeEmailTemplate() string {
	return constants.TemplateWelcome
}

// GetEmailTemplates returns all email templates (admin only)
func GetEmailTemplates(c *gin.Context) {
	db := database.GetDB()

	templates, err := models.GetAllEmailTemplates(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email templates"})
		return
	}

	// If no templates exist, return defaults
	if len(templates) == 0 {
		defaultTemplates := constants.GetDefaultTemplates()
		result := make(map[string]interface{})
		for _, template := range defaultTemplates {
			result[template.Type] = map[string]string{
				"subject": template.Subject,
				"content": template.Content,
			}
		}
		c.JSON(http.StatusOK, gin.H{"templates": result})
		return
	}

	result := make(map[string]interface{})
	for templateType, template := range templates {
		result[templateType] = map[string]string{
			"subject": template.Subject,
			"content": template.Content,
		}
	}

	c.JSON(http.StatusOK, gin.H{"templates": result})
}

// UpdateEmailTemplates updates email templates (admin only)
func UpdateEmailTemplates(c *gin.Context) {
	var req struct {
		Templates map[string]struct {
			Subject string `json:"subject" binding:"required"`
			Content string `json:"content" binding:"required"`
		} `json:"templates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	db := database.GetDB()

	// Save each template
	for templateType, template := range req.Templates {
		if templateType != constants.TemplateTypeUpcomingFeature &&
			templateType != constants.TemplateTypeNewRelease &&
			templateType != constants.TemplateTypeProposedFeature &&
			templateType != constants.TemplateTypeWelcome {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template type: " + templateType})
			return
		}

		err := models.SaveEmailTemplate(db, templateType, template.Subject, template.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save " + templateType + " template"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email templates updated successfully"})
}

// getDefaultUpcomingTemplate returns the default upcoming feature template
func getDefaultUpcomingTemplate() string {
	return constants.TemplateUpcomingFeature
}

// getDefaultReleaseTemplate returns the default new release template
func getDefaultReleaseTemplate() string {
	return constants.TemplateNewRelease
}

// getDefaultProposedTemplate returns the default proposed feature email template
func getDefaultProposedTemplate() string {
	return constants.TemplateProposedFeature
}

// DeleteNewsletterSubscriber removes a subscriber using soft delete (admin only)
func DeleteNewsletterSubscriber(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	db := database.GetDB()

	// Use soft delete for consistency with GORM's soft delete behavior
	err := db.Where("email = ?", email).Delete(&models.NewsletterSubscriber{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscriber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscriber deleted successfully",
	})
}
