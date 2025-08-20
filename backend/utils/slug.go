package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateSlug creates a URL-friendly slug from a title
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 50 characters
	if len(slug) > 50 {
		slug = slug[:50]
		// Remove trailing hyphen if we cut in the middle of a word
		slug = strings.TrimRight(slug, "-")
	}

	// If slug is empty (title had no alphanumeric characters), use UUID
	if slug == "" {
		return uuid.New().String()[:8]
	}

	return slug
}

// GenerateUniqueSlug creates a unique slug by checking the database
func GenerateUniqueSlug(db *gorm.DB, title string, tableName string, excludeID ...uint) string {
	baseSlug := GenerateSlug(title)
	slug := baseSlug
	counter := 1

	for {
		// Check if slug exists in database
		var count int64
		query := db.Table(tableName).Where("slug = ?", slug)

		// Exclude current record if updating
		if len(excludeID) > 0 && excludeID[0] > 0 {
			query = query.Where("id != ?", excludeID[0])
		}

		query.Count(&count)

		// If slug is unique, return it
		if count == 0 {
			return slug
		}

		// Generate new slug with counter
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++

		// Safety check to prevent infinite loop
		if counter > 1000 {
			// Use UUID as fallback
			return uuid.New().String()[:8]
		}
	}
}
