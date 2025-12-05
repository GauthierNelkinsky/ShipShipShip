package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StatusCategoryMapping maps user-defined statuses to theme categories
type StatusCategoryMapping struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	StatusDefinitionID uint      `json:"status_definition_id" gorm:"not null;index"`
	ThemeID            string    `json:"theme_id" gorm:"not null;index"`
	CategoryID         string    `json:"category_id" gorm:"not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ThemeSetting defines a configurable setting for the theme
type ThemeSetting struct {
	ID          string               `json:"id"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Type        string               `json:"type"`
	Default     interface{}          `json:"default"`
	Options     []ThemeSettingOption `json:"options,omitempty"`
	Items       []ThemeSetting       `json:"items,omitempty"`
}

// ThemeSettingOption defines an option for select-type settings
type ThemeSettingOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ThemeSettingGroup groups related settings together
type ThemeSettingGroup struct {
	Group       string         `json:"group"`
	Description string         `json:"description"`
	Settings    []ThemeSetting `json:"settings"`
}

// ThemeSettingValue stores the user-configured value for a theme setting
type ThemeSettingValue struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ThemeID   string    `json:"theme_id" gorm:"not null;index"`
	SettingID string    `json:"setting_id" gorm:"not null;index"`
	Value     string    `json:"value" gorm:"not null"` // Stored as JSON string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ThemeManifest represents the structure of theme.json
type ThemeManifest struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Author      string              `json:"author"`
	Settings    []ThemeSettingGroup `json:"settings"`
	Categories  []ThemeCategory     `json:"categories"`
}

// ThemeCategory defines a category that events can be mapped to
type ThemeCategory struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Multiple    bool   `json:"multiple"`
	Order       int    `json:"order"`
}

// LoadThemeManifest reads and parses the theme.json file
func LoadThemeManifest(themePath string) (*ThemeManifest, error) {
	manifestPath := fmt.Sprintf("%s/theme.json", themePath)

	// Check if file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("theme.json not found at %s", manifestPath)
	}

	// Read file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme.json: %w", err)
	}

	// Parse JSON
	var manifest ThemeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse theme.json: %w", err)
	}

	// Validate manifest
	if err := validateManifest(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validateManifest ensures the manifest has all required fields
func validateManifest(manifest *ThemeManifest) error {
	if manifest.ID == "" {
		return fmt.Errorf("theme ID is required")
	}
	if manifest.Name == "" {
		return fmt.Errorf("theme name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("theme version is required")
	}
	if len(manifest.Categories) == 0 {
		return fmt.Errorf("at least one category is required")
	}

	// Validate each category
	categoryIDs := make(map[string]bool)
	for i, cat := range manifest.Categories {
		if cat.ID == "" {
			return fmt.Errorf("category %d: ID is required", i)
		}
		if categoryIDs[cat.ID] {
			return fmt.Errorf("duplicate category ID: %s", cat.ID)
		}
		categoryIDs[cat.ID] = true

		if cat.Label == "" {
			return fmt.Errorf("category %s: label is required", cat.ID)
		}
		if cat.Description == "" {
			return fmt.Errorf("category %s: description is required", cat.ID)
		}
	}

	return nil
}

// GetOrCreateMapping gets or creates a mapping for a status
func GetOrCreateMapping(db *gorm.DB, statusDefID uint, themeID string, defaultCategoryID string) (*StatusCategoryMapping, error) {
	var mapping StatusCategoryMapping

	err := db.Where("status_definition_id = ? AND theme_id = ?", statusDefID, themeID).First(&mapping).Error
	if err == nil {
		return &mapping, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new mapping with default category
	mapping = StatusCategoryMapping{
		StatusDefinitionID: statusDefID,
		ThemeID:            themeID,
		CategoryID:         defaultCategoryID,
	}

	if err := db.Create(&mapping).Error; err != nil {
		return nil, err
	}

	return &mapping, nil
}

// SuggestCategoryForStatus suggests a category based on status name
func SuggestCategoryForStatus(statusName string, categories []ThemeCategory) string {
	lower := strings.ToLower(statusName)

	// Define keyword mappings
	keywordMappings := map[string][]string{
		"upcoming": {"doing", "progress", "wip", "dev", "development", "building",
			"cours", "actuel", "en cours", "current", "in progress"},
		"released": {"done", "released", "shipped", "live", "deployed", "completed",
			"terminé", "publié", "fini", "sortie", "launch"},
		"proposed": {"vote", "voting", "proposed", "idea", "suggestion", "feedback",
			"proposition", "idée", "request"},
		"feedback": {"feedback", "suggestion", "suggestions", "user feedback", "feature request"},
	}

	// Try to match keywords to category IDs
	for categoryID, keywords := range keywordMappings {
		// Check if this category exists in the theme
		categoryExists := false
		for _, cat := range categories {
			if cat.ID == categoryID {
				categoryExists = true
				break
			}
		}

		if !categoryExists {
			continue
		}

		// Check if status name contains any keyword
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return categoryID
			}
		}
	}

	// Default: return first category
	if len(categories) > 0 {
		return categories[0].ID
	}

	return "feedback"
}

// CreateDefaultMappings creates mappings for all statuses when a theme is applied
func CreateDefaultMappings(db *gorm.DB, themeID string, manifest *ThemeManifest) error {
	// Get all status definitions
	var statuses []EventStatusDefinition
	if err := db.Find(&statuses).Error; err != nil {
		return fmt.Errorf("failed to fetch statuses: %w", err)
	}

	for _, status := range statuses {
		// Check if mapping already exists
		var existing StatusCategoryMapping
		err := db.Where("status_definition_id = ? AND theme_id = ?", status.ID, themeID).First(&existing).Error

		if err == nil {
			// Mapping already exists, skip
			continue
		}

		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing mapping: %w", err)
		}

		// Suggest a category
		suggestedCategory := SuggestCategoryForStatus(status.DisplayName, manifest.Categories)

		// Create mapping
		mapping := StatusCategoryMapping{
			StatusDefinitionID: status.ID,
			ThemeID:            themeID,
			CategoryID:         suggestedCategory,
		}

		if err := db.Create(&mapping).Error; err != nil {
			return fmt.Errorf("failed to create mapping for status %s: %w", status.DisplayName, err)
		}

		fmt.Printf("Created mapping: %s -> %s\n", status.DisplayName, suggestedCategory)
	}

	return nil
}

// CreateDefaultStatusesFromTheme creates default statuses based on theme categories if no statuses exist
func CreateDefaultStatusesFromTheme(db *gorm.DB, themeID string, manifest *ThemeManifest) error {
	// Check if any statuses exist
	var count int64
	if err := db.Model(&EventStatusDefinition{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing statuses: %w", err)
	}

	// If statuses already exist, don't create defaults
	if count > 0 {
		fmt.Printf("Statuses already exist (%d), skipping default creation\n", count)
		return nil
	}

	fmt.Printf("No statuses found, creating defaults from theme categories\n")

	// Create a status for each category in the theme
	for i, category := range manifest.Categories {
		statusName := category.Label
		slug := strings.ToLower(strings.ReplaceAll(statusName, " ", "-"))

		// Check if status already exists (shouldn't, but be safe)
		var existing EventStatusDefinition
		err := db.Where("LOWER(display_name) = ?", strings.ToLower(statusName)).First(&existing).Error
		if err == nil {
			fmt.Printf("Status %s already exists, skipping\n", statusName)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing status: %w", err)
		}

		// Create the status
		status := EventStatusDefinition{
			DisplayName: statusName,
			Slug:        slug,
			Order:       i,
			IsReserved:  false,
		}

		if err := db.Create(&status).Error; err != nil {
			return fmt.Errorf("failed to create status %s: %w", statusName, err)
		}

		fmt.Printf("Created status: %s (order: %d)\n", statusName, i)

		// Create mapping for this status
		mapping := StatusCategoryMapping{
			StatusDefinitionID: status.ID,
			ThemeID:            themeID,
			CategoryID:         category.ID,
		}

		if err := db.Create(&mapping).Error; err != nil {
			return fmt.Errorf("failed to create mapping for status %s: %w", statusName, err)
		}

		fmt.Printf("Created mapping: %s -> %s\n", statusName, category.ID)
	}

	return nil
}
