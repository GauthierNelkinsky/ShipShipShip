package models

import (
	"time"

	"gorm.io/gorm"
)

type ProjectSettings struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Title        string         `json:"title" gorm:"not null;default:'Changelog'"`
	LogoURL      string         `json:"logo_url" gorm:"column:logo_url"`
	DarkLogoURL  string         `json:"dark_logo_url" gorm:"column:dark_logo_url"`
	FaviconURL   string         `json:"favicon_url" gorm:"column:favicon_url"`
	WebsiteURL   string         `json:"website_url" gorm:"column:website_url"`
	PrimaryColor string         `json:"primary_color" gorm:"not null;default:'#3b82f6'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type UpdateSettingsRequest struct {
	Title        *string `json:"title"`
	LogoURL      *string `json:"logo_url"`
	DarkLogoURL  *string `json:"dark_logo_url"`
	FaviconURL   *string `json:"favicon_url"`
	WebsiteURL   *string `json:"website_url"`
	PrimaryColor *string `json:"primary_color"`
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
				Title:        "Changelog",
				LogoURL:      "",
				DarkLogoURL:  "",
				FaviconURL:   "",
				WebsiteURL:   "",
				PrimaryColor: "#3b82f6",
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
