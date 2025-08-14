package database

import (
	"log"
	"os"
	"time"

	"chessload-changelog/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
		&models.Event{},
		&models.ProjectSettings{},
		&models.Vote{},
	); err != nil {
		return err
	}

	// Run data migrations
	return runDataMigrations()
}

func runDataMigrations() error {
	// Migration 1: Convert "Vote" status to "Upcoming"
	if err := DB.Model(&models.Event{}).Where("status = ?", "Vote").Update("status", "Upcoming").Error; err != nil {
		log.Printf("Warning: Failed to migrate Vote status to Upcoming: %v", err)
		// Don't fail completely, just log the warning
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

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
