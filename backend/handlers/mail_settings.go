package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/utils"

	"github.com/gin-gonic/gin"
)

func GetMailSettings(c *gin.Context) {
	db := database.GetDB()
	settings, err := models.GetOrCreateMailSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mail settings"})
		return
	}

	// Don't return the password in the response for security
	settings.SMTPPassword = ""

	c.JSON(http.StatusOK, settings)
}

func UpdateMailSettings(c *gin.Context) {
	var req models.UpdateMailSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	settings, err := models.GetOrCreateMailSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mail settings"})
		return
	}

	// Update fields if provided
	if req.SMTPHost != nil {
		settings.SMTPHost = *req.SMTPHost
	}
	if req.SMTPPort != nil {
		settings.SMTPPort = *req.SMTPPort
	}
	if req.SMTPUsername != nil {
		settings.SMTPUsername = *req.SMTPUsername
	}
	if req.SMTPPassword != nil {
		settings.SMTPPassword = *req.SMTPPassword
	}
	if req.SMTPEncryption != nil {
		settings.SMTPEncryption = *req.SMTPEncryption
	}
	if req.FromEmail != nil {
		settings.FromEmail = *req.FromEmail
	}
	if req.FromName != nil {
		settings.FromName = *req.FromName
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mail settings"})
		return
	}

	// Don't return the password in the response for security
	settings.SMTPPassword = ""

	c.JSON(http.StatusOK, settings)
}

func TestMailSettings(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email address is required"})
		return
	}

	db := database.GetDB()
	settings, err := models.GetOrCreateMailSettings(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mail settings"})
		return
	}

	// Validate that required settings are configured
	if settings.SMTPHost == "" || settings.FromEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP host and from email must be configured"})
		return
	}

	// Send test email
	err = sendTestEmail(settings, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send test email: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}

func sendTestEmail(settings *models.MailSettings, toEmail string) error {
	// Prepare email content
	fromName := settings.FromName
	if fromName == "" {
		fromName = "ShipShipShip"
	}

	from := fmt.Sprintf("%s <%s>", fromName, settings.FromEmail)
	to := toEmail
	subject := "ShipShipShip Test Email"
	body := `This is a test email from ShipShipShip to verify your SMTP configuration.

If you received this email, your mail settings are working correctly!

Best regards,
ShipShipShip Team`

	// Prepare message
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Determine authentication
	var auth smtp.Auth
	if settings.SMTPUsername != "" {
		auth = smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
	}

	// Send email based on encryption type
	addr := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)

	switch strings.ToLower(settings.SMTPEncryption) {
	case "ssl":
		return utils.SendMailWithSSL(addr, auth, settings.FromEmail, []string{toEmail}, []byte(message))
	case "tls":
		return utils.SendMailWithTLS(addr, auth, settings.FromEmail, []string{toEmail}, []byte(message))
	default:
		// No encryption
		return smtp.SendMail(addr, auth, settings.FromEmail, []string{toEmail}, []byte(message))
	}
}
