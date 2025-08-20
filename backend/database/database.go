package database

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"shipshipship/models"
	"shipshipship/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Migration represents a database migration
type Migration struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"unique;not null"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

var DB *gorm.DB

func InitDatabase() {
	var err error

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/changelog.db"
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Configure GORM logger
	gormLogger := logger.Default
	if os.Getenv("GIN_MODE") == "release" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

func migrate() error {
	// Auto-migrate the schema
	if err := DB.AutoMigrate(
		&Migration{},
		&models.Tag{},
		&models.Event{},
		&models.EventPublication{},
		&models.EventEmailHistory{},
		&models.ProjectSettings{},
		&models.Vote{},
		&models.MailSettings{},
		&models.NewsletterSubscriber{},
		&models.NewsletterHistory{},
		&models.EmailTemplate{},
	); err != nil {
		return err
	}

	// Run data migrations
	return runDataMigrations()
}

// hasAppliedMigration checks if a migration has already been applied
func hasAppliedMigration(name string) bool {
	var count int64
	DB.Model(&Migration{}).Where("name = ?", name).Count(&count)
	return count > 0
}

// markMigrationApplied marks a migration as applied
func markMigrationApplied(name string) error {
	migration := Migration{Name: name, AppliedAt: time.Now()}
	return DB.Create(&migration).Error
}

func runDataMigrations() error {
	log.Println("=== Starting Data Migrations ===")

	// Migration 1: Convert "Vote" status to "Upcoming"
	migrationName1 := "vote_status_to_upcoming"
	if !hasAppliedMigration(migrationName1) {
		var voteCount int64
		DB.Model(&models.Event{}).Where("status = ?", "Vote").Count(&voteCount)
		log.Printf("Migration 1: Found %d events with 'Vote' status", voteCount)

		if voteCount > 0 {
			if err := DB.Model(&models.Event{}).Where("status = ?", "Vote").Update("status", "Upcoming").Error; err != nil {
				log.Printf("Warning: Failed to migrate Vote status to Upcoming: %v", err)
				// Don't fail completely, just log the warning
			} else {
				log.Printf("Successfully migrated %d events from Vote to Upcoming", voteCount)
			}
		}

		// Mark migration as applied
		if err := markMigrationApplied(migrationName1); err != nil {
			log.Printf("Warning: Failed to mark migration as applied: %v", err)
		}
	} else {
		log.Println("Migration 1 (Vote to Upcoming) already applied, skipping")
	}

	// Migration 2: Drop description column from events table
	if DB.Migrator().HasColumn(&models.Event{}, "description") {
		if err := DB.Migrator().DropColumn(&models.Event{}, "description"); err != nil {
			log.Printf("Warning: Failed to drop description column: %v", err)
			// Don't fail completely, just log the warning
		} else {
			log.Println("Successfully dropped description column from events table")
		}
	}

	// Migration 3: Merge ReleasedDate and EstimatedDate into Date field
	if DB.Migrator().HasColumn(&models.Event{}, "released_date") || DB.Migrator().HasColumn(&models.Event{}, "estimated_date") {
		// Add the new Date column if it doesn't exist
		if !DB.Migrator().HasColumn(&models.Event{}, "date") {
			if err := DB.Migrator().AddColumn(&models.Event{}, "date"); err != nil {
				log.Printf("Warning: Failed to add date column: %v", err)
			}
		}

		// Migrate data from old columns to new Date column
		var events []struct {
			ID            uint
			ReleasedDate  *string `gorm:"column:released_date"`
			EstimatedDate *string `gorm:"column:estimated_date"`
		}

		if err := DB.Raw("SELECT id, released_date, estimated_date FROM events").Scan(&events).Error; err != nil {
			log.Printf("Warning: Failed to fetch events for date migration: %v", err)
		} else {
			for _, event := range events {
				var dateValue string

				// Prioritize released_date if it exists and is not null
				if event.ReleasedDate != nil && *event.ReleasedDate != "" {
					// Parse the released date and format it
					if parsedTime, err := time.Parse("2006-01-02T15:04:05Z", *event.ReleasedDate); err == nil {
						dateValue = parsedTime.Format("2006-01-02")
					} else if parsedTime, err := time.Parse("2006-01-02", *event.ReleasedDate); err == nil {
						dateValue = parsedTime.Format("2006-01-02")
					} else {
						dateValue = *event.ReleasedDate
					}
				} else if event.EstimatedDate != nil && *event.EstimatedDate != "" {
					// Use estimated_date as is
					dateValue = *event.EstimatedDate
				}

				if dateValue != "" {
					if err := DB.Exec("UPDATE events SET date = ? WHERE id = ?", dateValue, event.ID).Error; err != nil {
						log.Printf("Warning: Failed to migrate date for event %d: %v", event.ID, err)
					}
				}
			}
			log.Println("Successfully migrated date fields")
		}

		// Drop the old columns
		if DB.Migrator().HasColumn(&models.Event{}, "released_date") {
			if err := DB.Migrator().DropColumn(&models.Event{}, "released_date"); err != nil {
				log.Printf("Warning: Failed to drop released_date column: %v", err)
			} else {
				log.Println("Successfully dropped released_date column")
			}
		}

		if DB.Migrator().HasColumn(&models.Event{}, "estimated_date") {
			if err := DB.Migrator().DropColumn(&models.Event{}, "estimated_date"); err != nil {
				log.Printf("Warning: Failed to drop estimated_date column: %v", err)
			} else {
				log.Println("Successfully dropped estimated_date column")
			}
		}
	}

	// Migration 4: Add sort_order column to events table
	if !DB.Migrator().HasColumn(&models.Event{}, "sort_order") {
		if err := DB.Migrator().AddColumn(&models.Event{}, "sort_order"); err != nil {
			log.Printf("Warning: Failed to add sort_order column: %v", err)
		} else {
			log.Println("Successfully added sort_order column to events table")

			// Set initial order values based on created_at for existing events
			var events []models.Event
			if err := DB.Order("created_at ASC").Find(&events).Error; err == nil {
				for i, event := range events {
					if err := DB.Model(&event).Update("sort_order", i).Error; err != nil {
						log.Printf("Warning: Failed to set initial sort_order for event %d: %v", event.ID, err)
					}
				}
				log.Println("Successfully set initial sort_order values for existing events")
			}
		}
	}

	// Migration 5: Initialize default email templates
	if err := models.InitializeDefaultEmailTemplates(DB); err != nil {
		log.Printf("Warning: Failed to initialize default email templates: %v", err)
		// Don't fail completely, just log the warning
	} else {
		log.Println("Successfully initialized default email templates")
	}

	// Migration 6: Rename event statuses - Upcoming->Proposed, Doing->Upcoming
	migrationName := "status_rename_upcoming_to_proposed"

	// Log current status distribution
	var statusCounts struct {
		Proposed int64
		Upcoming int64
		Doing    int64
	}
	DB.Model(&models.Event{}).Where("status = ?", "Proposed").Count(&statusCounts.Proposed)
	DB.Model(&models.Event{}).Where("status = ?", "Upcoming").Count(&statusCounts.Upcoming)
	DB.Model(&models.Event{}).Where("status = ?", "Doing").Count(&statusCounts.Doing)

	log.Printf("Migration 6: Current status distribution - Proposed: %d, Upcoming: %d, Doing: %d",
		statusCounts.Proposed, statusCounts.Upcoming, statusCounts.Doing)

	// Additional check: if we already have events with "Proposed" status,
	// this migration likely already ran
	if !hasAppliedMigration(migrationName) && statusCounts.Proposed == 0 {
		if statusCounts.Upcoming > 0 || statusCounts.Doing > 0 {
			log.Println("Running Migration 6: status rename (Upcoming->Proposed, Doing->Upcoming)")

			if statusCounts.Upcoming > 0 {
				if err := DB.Model(&models.Event{}).Where("status = ?", "Upcoming").Update("status", "Proposed").Error; err != nil {
					log.Printf("Warning: Failed to migrate Upcoming status to Proposed: %v", err)
				} else {
					log.Printf("Successfully migrated %d events from Upcoming to Proposed", statusCounts.Upcoming)
				}
			}

			if statusCounts.Doing > 0 {
				if err := DB.Model(&models.Event{}).Where("status = ?", "Doing").Update("status", "Upcoming").Error; err != nil {
					log.Printf("Warning: Failed to migrate Doing status to Upcoming: %v", err)
				} else {
					log.Printf("Successfully migrated %d events from Doing to Upcoming", statusCounts.Doing)
				}
			}
		} else {
			log.Println("Migration 6: No events to migrate")
		}

		// Mark migration as applied
		if err := markMigrationApplied(migrationName); err != nil {
			log.Printf("Warning: Failed to mark migration as applied: %v", err)
		}
	} else if statusCounts.Proposed > 0 {
		log.Printf("Migration 6 (status rename) already applied (detected %d Proposed events), skipping", statusCounts.Proposed)
		// Mark as applied if not already marked
		if !hasAppliedMigration(migrationName) {
			if err := markMigrationApplied(migrationName); err != nil {
				log.Printf("Warning: Failed to mark migration as applied: %v", err)
			}
		}
	} else {
		log.Println("Migration 6 (status rename) already applied, skipping")
	}

	// Migration 7: Migrate string tags to Tag model
	if err := migrateStringTagsToTagModel(); err != nil {
		log.Printf("Warning: Failed to migrate string tags to Tag model: %v", err)
		// Don't fail completely, just log the warning
	} else {
		log.Println("Successfully migrated string tags to Tag model")
	}

	// Migration 8: Update email templates to mobile-friendly layout
	migrationName8 := "email_templates_mobile_layout"
	if !hasAppliedMigration(migrationName8) {
		log.Println("Running Migration 8: Update email templates to mobile-friendly layout")

		if err := models.UpdateEmailTemplatesToMobileFriendly(DB); err != nil {
			log.Printf("Warning: Failed to update email templates to mobile-friendly layout: %v", err)
		} else {
			log.Println("Successfully updated email templates to mobile-friendly layout")
		}

		// Mark migration as applied
		if err := markMigrationApplied(migrationName8); err != nil {
			log.Printf("Warning: Failed to mark migration as applied: %v", err)
		}
	} else {
		log.Println("Migration 8 (email templates mobile layout) already applied, skipping")
	}

	// Migration 9: Add slug field to events table and populate existing events
	migrationName9 := "add_slug_to_events"
	if !hasAppliedMigration(migrationName9) {
		log.Println("Running Migration 9: Add slug field to events table")

		// Check if slug column already exists
		if !DB.Migrator().HasColumn(&models.Event{}, "slug") {
			// Add the slug column
			if err := DB.Migrator().AddColumn(&models.Event{}, "slug"); err != nil {
				log.Printf("Warning: Failed to add slug column: %v", err)
			} else {
				log.Println("Successfully added slug column to events table")
			}
		}

		// Populate slugs for existing events
		var events []models.Event
		if err := DB.Find(&events).Error; err == nil {
			for _, event := range events {
				if event.Slug == "" {
					// Generate unique slug for this event
					slug := utils.GenerateUniqueSlug(DB, event.Title, "events", event.ID)
					if err := DB.Model(&event).Update("slug", slug).Error; err != nil {
						log.Printf("Warning: Failed to set slug for event %d: %v", event.ID, err)
					}
				}
			}
			log.Println("Successfully populated slugs for existing events")
		} else {
			log.Printf("Warning: Failed to fetch events for slug population: %v", err)
		}

		// Mark migration as applied
		if err := markMigrationApplied(migrationName9); err != nil {
			log.Printf("Warning: Failed to mark migration as applied: %v", err)
		}
	} else {
		log.Println("Migration 9 (add slug to events) already applied, skipping")
	}

	log.Println("=== Data Migrations Complete ===")
	return nil
}

func migrateStringTagsToTagModel() error {
	// Check if the events table still has the old tags column as text
	if !DB.Migrator().HasColumn(&models.Event{}, "tags") {
		return nil // Migration already completed
	}

	// Get all events with their old string tags
	var events []struct {
		ID   uint
		Tags string
	}

	if err := DB.Raw("SELECT id, tags FROM events WHERE tags != '' AND tags != '[]' AND tags IS NOT NULL").Scan(&events).Error; err != nil {
		return err
	}

	// Create a map to store unique tags
	tagMap := make(map[string]*models.Tag)

	// Process each event's tags
	for _, event := range events {
		if event.Tags == "" || event.Tags == "[]" {
			continue
		}

		// Parse JSON string tags
		var tagNames []string
		if err := json.Unmarshal([]byte(event.Tags), &tagNames); err != nil {
			log.Printf("Warning: Failed to parse tags for event %d: %v", event.ID, err)
			continue
		}

		// Create or find tags and associate with event
		for _, tagName := range tagNames {
			if tagName == "" {
				continue
			}

			var tag *models.Tag
			if existingTag, exists := tagMap[tagName]; exists {
				tag = existingTag
			} else {
				// Check if tag already exists in database
				var dbTag models.Tag
				if err := DB.Where("name = ?", tagName).First(&dbTag).Error; err != nil {
					// Tag doesn't exist, create it
					dbTag = models.Tag{
						Name:  tagName,
						Color: getDefaultColorForTag(tagName),
					}
					if err := DB.Create(&dbTag).Error; err != nil {
						log.Printf("Warning: Failed to create tag %s: %v", tagName, err)
						continue
					}
				}
				tag = &dbTag
				tagMap[tagName] = tag
			}

			// Associate tag with event using raw SQL to avoid GORM issues
			DB.Exec("INSERT OR IGNORE INTO event_tags (event_id, tag_id) VALUES (?, ?)", event.ID, tag.ID)
		}
	}

	// Drop the old tags column after successful migration
	if err := DB.Migrator().DropColumn(&models.Event{}, "tags"); err != nil {
		log.Printf("Warning: Failed to drop old tags column: %v", err)
	}

	return nil
}

func getDefaultColorForTag(tagName string) string {
	// Assign different colors based on tag name
	colors := map[string]string{
		"Feature":     "#10B981", // Green
		"Bug":         "#EF4444", // Red
		"Enhancement": "#8B5CF6", // Purple
		"Feedback":    "#F59E0B", // Yellow
		"UI/UX":       "#EC4899", // Pink
		"API":         "#06B6D4", // Cyan
		"Security":    "#DC2626", // Dark red
		"Performance": "#059669", // Dark green
	}

	if color, exists := colors[tagName]; exists {
		return color
	}

	// Default blue color
	return "#3B82F6"
}

func GetDB() *gorm.DB {
	return DB
}
