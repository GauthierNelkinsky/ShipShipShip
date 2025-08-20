package models

import (
	"time"

	"shipshipship/constants"

	"gorm.io/gorm"
)

type NewsletterSubscriber struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	SubscribedAt time.Time      `json:"subscribed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type SubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UnsubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// GetActiveSubscriberCount returns the number of active newsletter subscribers
func GetActiveSubscriberCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&NewsletterSubscriber{}).Count(&count).Error
	return count, err
}

// FindSubscriberByEmail finds a subscriber by email address
func FindSubscriberByEmail(db *gorm.DB, email string) (*NewsletterSubscriber, error) {
	var subscriber NewsletterSubscriber
	err := db.Where("email = ?", email).First(&subscriber).Error
	if err != nil {
		return nil, err
	}
	return &subscriber, nil
}

// Subscribe creates a new newsletter subscription or reactivates a soft-deleted one
func Subscribe(db *gorm.DB, email string) (*NewsletterSubscriber, error) {
	var subscriber NewsletterSubscriber

	// Check if subscriber already exists (including soft-deleted records)
	err := db.Unscoped().Where("email = ?", email).First(&subscriber).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new subscriber
			subscriber = NewsletterSubscriber{
				Email:        email,
				IsActive:     true,
				SubscribedAt: time.Now(),
			}
			err = db.Create(&subscriber).Error
			if err != nil {
				return nil, err
			}
			return &subscriber, nil
		}
		return nil, err
	}

	// If record was soft-deleted, restore it
	if subscriber.DeletedAt.Valid {
		subscriber.DeletedAt = gorm.DeletedAt{}
		subscriber.IsActive = true
		subscriber.SubscribedAt = time.Now()
		err = db.Unscoped().Save(&subscriber).Error
		if err != nil {
			return nil, err
		}
		return &subscriber, nil
	}

	// Already active subscriber
	return &subscriber, nil
}

// Unsubscribe removes a newsletter subscription using soft delete
func Unsubscribe(db *gorm.DB, email string) error {
	return db.Where("email = ?", email).Delete(&NewsletterSubscriber{}).Error
}

type NewsletterHistory struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Subject        string         `json:"subject" gorm:"not null"`
	Content        string         `json:"content" gorm:"type:text;not null"`
	Status         string         `json:"status" gorm:"not null;default:'draft'"` // draft, sending, sent, failed
	RecipientCount int            `json:"recipient_count" gorm:"default:0"`
	OpenCount      int            `json:"open_count" gorm:"default:0"`
	ClickCount     int            `json:"click_count" gorm:"default:0"`
	SentAt         *time.Time     `json:"sent_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetNewsletterHistory returns paginated newsletter history
func GetNewsletterHistory(db *gorm.DB, page, limit int) ([]NewsletterHistory, int64, error) {
	var newsletters []NewsletterHistory
	var total int64

	// Count total records
	if err := db.Model(&NewsletterHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * limit
	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&newsletters).Error
	return newsletters, total, err
}

// GetSubscribersPaginated returns paginated newsletter subscribers
func GetSubscribersPaginated(db *gorm.DB, page, limit int) ([]NewsletterSubscriber, int64, error) {
	var subscribers []NewsletterSubscriber
	var total int64

	// Count total records
	if err := db.Model(&NewsletterSubscriber{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * limit
	err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&subscribers).Error
	return subscribers, total, err
}

type EmailTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"not null;uniqueIndex"` // newsletter, welcome
	Subject   string         `json:"subject" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetEmailTemplate returns an email template by type
func GetEmailTemplate(db *gorm.DB, templateType string) (*EmailTemplate, error) {
	var template EmailTemplate
	err := db.Where("type = ?", templateType).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// SaveEmailTemplate creates or updates an email template
func SaveEmailTemplate(db *gorm.DB, templateType, subject, content string) error {
	var template EmailTemplate

	err := db.Where("type = ?", templateType).First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new template
			template = EmailTemplate{
				Type:    templateType,
				Subject: subject,
				Content: content,
			}
			return db.Create(&template).Error
		}
		return err
	}

	// Update existing template
	template.Subject = subject
	template.Content = content
	return db.Save(&template).Error
}

// GetAllEmailTemplates returns all email templates
func GetAllEmailTemplates(db *gorm.DB) (map[string]EmailTemplate, error) {
	var templates []EmailTemplate
	err := db.Find(&templates).Error
	if err != nil {
		return nil, err
	}

	templateMap := make(map[string]EmailTemplate)
	for _, template := range templates {
		templateMap[template.Type] = template
	}

	return templateMap, nil
}

// GetActiveNewsletterSubscribers returns all active newsletter subscribers
func GetActiveNewsletterSubscribers(db *gorm.DB) ([]NewsletterSubscriber, error) {
	var subscribers []NewsletterSubscriber
	err := db.Find(&subscribers).Error
	return subscribers, err
}

// InitializeDefaultEmailTemplates ensures default email templates exist in the database
func InitializeDefaultEmailTemplates(db *gorm.DB) error {
	templates := constants.GetDefaultTemplates()

	for _, template := range templates {
		// Check if template already exists
		var existingTemplate EmailTemplate
		err := db.Where("type = ?", template.Type).First(&existingTemplate).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			// Template doesn't exist, create it
			err = SaveEmailTemplate(db, template.Type, template.Subject, template.Content)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}

// UpdateEmailTemplatesToMobileFriendly updates existing email templates to mobile-friendly layout
func UpdateEmailTemplatesToMobileFriendly(db *gorm.DB) error {
	templates := constants.GetDefaultTemplates()

	for _, template := range templates {
		// Update existing template with new content
		err := SaveEmailTemplate(db, template.Type, template.Subject, template.Content)
		if err != nil {
			return err
		}
	}

	return nil
}
