package models

import (
	"time"

	"gorm.io/gorm"
)

type ProjectSettings struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	Title               string         `json:"title" gorm:"not null;default:'Changelog'"`
	LogoURL             string         `json:"logo_url" gorm:"column:logo_url"`
	DarkLogoURL         string         `json:"dark_logo_url" gorm:"column:dark_logo_url"`
	FaviconURL          string         `json:"favicon_url" gorm:"column:favicon_url"`
	WebsiteURL          string         `json:"website_url" gorm:"column:website_url"`
	PrimaryColor        string         `json:"primary_color" gorm:"not null;default:'#3b82f6'"`
	NewsletterEnabled   bool           `json:"newsletter_enabled" gorm:"column:newsletter_enabled;default:false"`
	CurrentThemeID      string         `json:"current_theme_id" gorm:"column:current_theme_id"`
	CurrentThemeVersion string         `json:"current_theme_version" gorm:"column:current_theme_version"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

type UpdateSettingsRequest struct {
	Title               *string `json:"title"`
	LogoURL             *string `json:"logo_url"`
	DarkLogoURL         *string `json:"dark_logo_url"`
	FaviconURL          *string `json:"favicon_url"`
	WebsiteURL          *string `json:"website_url"`
	PrimaryColor        *string `json:"primary_color"`
	NewsletterEnabled   *bool   `json:"newsletter_enabled"`
	CurrentThemeID      *string `json:"current_theme_id"`
	CurrentThemeVersion *string `json:"current_theme_version"`
}

// GetOrCreateSettings ensures there's always a settings record
func GetOrCreateSettings(db *gorm.DB) (*ProjectSettings, error) {
	var settings ProjectSettings

	// Try to get existing settings
	result := db.First(&settings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create default settings if none exist
			settings = ProjectSettings{
				Title:               "Changelog",
				LogoURL:             "",
				DarkLogoURL:         "",
				FaviconURL:          "",
				WebsiteURL:          "",
				PrimaryColor:        "#3b82f6",
				NewsletterEnabled:   false,
				CurrentThemeID:      "",
				CurrentThemeVersion: "",
			}
			if err := db.Create(&settings).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	}

	return &settings, nil
}

type MailSettings struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	SMTPHost       string         `json:"smtp_host" gorm:"column:smtp_host"`
	SMTPPort       int            `json:"smtp_port" gorm:"column:smtp_port;default:587"`
	SMTPUsername   string         `json:"smtp_username" gorm:"column:smtp_username"`
	SMTPPassword   string         `json:"smtp_password" gorm:"column:smtp_password"`
	SMTPEncryption string         `json:"smtp_encryption" gorm:"column:smtp_encryption;default:'tls'"`
	FromEmail      string         `json:"from_email" gorm:"column:from_email"`
	FromName       string         `json:"from_name" gorm:"column:from_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type UpdateMailSettingsRequest struct {
	SMTPHost       *string `json:"smtp_host"`
	SMTPPort       *int    `json:"smtp_port"`
	SMTPUsername   *string `json:"smtp_username"`
	SMTPPassword   *string `json:"smtp_password"`
	SMTPEncryption *string `json:"smtp_encryption"`
	FromEmail      *string `json:"from_email"`
	FromName       *string `json:"from_name"`
}

// GetOrCreateMailSettings ensures there's always a mail settings record
func GetOrCreateMailSettings(db *gorm.DB) (*MailSettings, error) {
	var settings MailSettings
	var count int64

	// Check if any mail settings exist to avoid logging "record not found"
	db.Model(&MailSettings{}).Count(&count)

	if count == 0 {
		// Create default settings if none exist
		settings = MailSettings{
			SMTPHost:       "",
			SMTPPort:       587,
			SMTPUsername:   "",
			SMTPPassword:   "",
			SMTPEncryption: "tls",
			FromEmail:      "",
			FromName:       "",
		}
		if err := db.Create(&settings).Error; err != nil {
			return nil, err
		}
	} else {
		// Get existing settings
		if err := db.First(&settings).Error; err != nil {
			return nil, err
		}
	}

	return &settings, nil
}

// BrandingSettings represents settings used for email branding
type BrandingSettings struct {
	ProjectName string
	ProjectURL  string
}

// GetBrandingSettings returns branding settings for email generation
func GetBrandingSettings(db *gorm.DB) (*BrandingSettings, error) {
	settings, err := GetOrCreateSettings(db)
	if err != nil {
		return nil, err
	}

	return &BrandingSettings{
		ProjectName: settings.Title,
		ProjectURL:  settings.WebsiteURL,
	}, nil
}
