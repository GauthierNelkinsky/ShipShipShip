package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"shipshipship/database"
	"shipshipship/models"
	"shipshipship/utils"
)

type EmailService struct {
	mailSettings *models.MailSettings
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendEmail sends an email to a single recipient
func (es *EmailService) SendEmail(to, subject, htmlContent string) error {
	// Get mail settings
	if es.mailSettings == nil {
		db := database.GetDB()
		settings, err := models.GetOrCreateMailSettings(db)
		if err != nil {
			return fmt.Errorf("failed to get mail settings: %v", err)
		}
		es.mailSettings = settings
	}

	// Validate settings
	if es.mailSettings.SMTPHost == "" || es.mailSettings.FromEmail == "" {
		return fmt.Errorf("SMTP host and from email must be configured")
	}

	// Prepare email content
	fromName := es.mailSettings.FromName
	if fromName == "" {
		fromName = "ShipShipShip"
	}

	from := fmt.Sprintf("%s <%s>", fromName, es.mailSettings.FromEmail)

	// Create email message
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += htmlContent

	// Determine authentication
	var auth smtp.Auth
	if es.mailSettings.SMTPUsername != "" {
		auth = smtp.PlainAuth("", es.mailSettings.SMTPUsername, es.mailSettings.SMTPPassword, es.mailSettings.SMTPHost)
	}

	// Send email based on encryption type
	addr := fmt.Sprintf("%s:%d", es.mailSettings.SMTPHost, es.mailSettings.SMTPPort)

	switch strings.ToLower(es.mailSettings.SMTPEncryption) {
	case "ssl":
		return utils.SendMailWithSSL(addr, auth, es.mailSettings.FromEmail, []string{to}, []byte(message))
	case "tls":
		return utils.SendMailWithTLS(addr, auth, es.mailSettings.FromEmail, []string{to}, []byte(message))
	default:
		// No encryption
		return smtp.SendMail(addr, auth, es.mailSettings.FromEmail, []string{to}, []byte(message))
	}
}
