package database

import (
	"log"
	"os"

	"shipshipship/models"

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
	// Fix corrupted project_settings table if it exists
	if err := fixCorruptedProjectSettings(DB); err != nil {
		log.Printf("Warning: Failed to fix corrupted project_settings: %v", err)
		// If we can't fix it, drop the table entirely and let AutoMigrate recreate it
		log.Println("Attempting to drop project_settings table to allow clean recreation...")
		DB.Exec("DROP TABLE IF EXISTS project_settings")
	}

	// Clean up removed columns and tables BEFORE auto-migration
	if err := cleanupRemovedColumnsAndTables(DB); err != nil {
		log.Printf("Warning: Failed to cleanup removed columns and tables: %v", err)
	}

	// Auto-migrate the schema
	if err := DB.AutoMigrate(
		&models.Tag{},
		&models.EventStatusDefinition{},
		&models.Event{},
		&models.EventPublication{},
		&models.EventEmailHistory{},
		&models.ProjectSettings{},
		&models.Vote{},
		&models.EventReaction{},
		&models.MailSettings{},
		&models.NewsletterSubscriber{},
		&models.NewsletterHistory{},
		&models.EmailTemplate{},
		&models.NewsletterAutomationSettings{},
		&models.StatusCategoryMapping{},
		&models.ThemeSettingValue{},
	); err != nil {
		// If AutoMigrate fails on project_settings, it's likely corrupted
		log.Printf("AutoMigrate failed: %v", err)
		log.Println("Dropping and recreating project_settings table...")
		DB.Exec("DROP TABLE IF EXISTS project_settings")
		// Try AutoMigrate again
		if err := DB.AutoMigrate(&models.ProjectSettings{}); err != nil {
			return err
		}
	}

	// Initialize default email templates
	if err := models.InitializeDefaultEmailTemplates(DB); err != nil {
		log.Printf("Warning: Failed to initialize default email templates: %v", err)
	} else {
		log.Println("Successfully initialized default email templates")
	}

	// Seed status definitions (reserved + legacy)
	if err := models.SeedStatusDefinitions(DB); err != nil {
		log.Printf("Warning: Failed to seed status definitions: %v", err)
	}

	// Ensure newsletter automation settings table exists (manual fallback)
	if err := createNewsletterAutomationTableIfNotExists(DB); err != nil {
		log.Printf("Warning: Failed to create newsletter automation table: %v", err)
	}

	return nil
}

// fixCorruptedProjectSettings checks for and fixes corrupted project_settings table
func fixCorruptedProjectSettings(db *gorm.DB) error {
	// Check if project_settings table exists
	var tableCount int64
	err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='project_settings'").Scan(&tableCount).Error
	if err != nil {
		return err
	}

	if tableCount == 0 {
		// Table doesn't exist yet, nothing to fix
		return nil
	}

	// Try to get the CREATE statement to check for corruption
	var createStmt string
	err = db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='project_settings'").Scan(&createStmt).Error
	if err != nil {
		log.Println("Cannot read project_settings schema, dropping table...")
		db.Exec("DROP TABLE IF EXISTS project_settings")
		return nil
	}

	// Check if the schema looks corrupted (has backticks or duplicate title)
	if len(createStmt) > 500 { // Unreasonably long schema indicates corruption

		log.Println("Detected potentially corrupted project_settings table schema")
		log.Println("Dropping and will recreate with clean schema...")

		if err := db.Exec("DROP TABLE IF EXISTS project_settings").Error; err != nil {
			return err
		}

		log.Println("✓ Dropped corrupted project_settings table")
		return nil
	}

	// Try to query the table to ensure it's readable
	var testQuery int64
	err = db.Raw("SELECT count(*) FROM project_settings").Scan(&testQuery).Error
	if err != nil {
		log.Printf("project_settings table exists but is unreadable: %v", err)
		log.Println("Dropping corrupted table...")
		db.Exec("DROP TABLE IF EXISTS project_settings")
		return nil
	}

	return nil
}

// createNewsletterAutomationTableIfNotExists ensures the newsletter automation settings table exists
func createNewsletterAutomationTableIfNotExists(db *gorm.DB) error {
	var count int64
	err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='newsletter_automation_settings'").Scan(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		// Table doesn't exist, create it
		err = db.Exec(`CREATE TABLE newsletter_automation_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enabled BOOLEAN DEFAULT FALSE,
			trigger_statuses TEXT DEFAULT '[]',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`).Error
		if err != nil {
			return err
		}
		log.Println("Successfully created newsletter_automation_settings table")
	}

	return nil
}

// cleanupRemovedColumnsAndTables removes deprecated tables and columns
func cleanupRemovedColumnsAndTables(db *gorm.DB) error {
	// Drop footer_links table if it exists
	var footerLinksCount int64
	err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='footer_links'").Scan(&footerLinksCount).Error
	if err != nil {
		return err
	}

	if footerLinksCount > 0 {
		log.Println("Dropping footer_links table (no longer used)...")
		if err := db.Exec("DROP TABLE footer_links").Error; err != nil {
			log.Printf("Warning: Failed to drop footer_links table: %v", err)
		} else {
			log.Println("✓ Dropped footer_links table")
		}
	}

	// Check if any of the deprecated columns exist in project_settings
	columnsToCheck := []string{"logo_url", "dark_logo_url", "primary_color", "newsletter_enabled"}
	hasDeprecatedColumns := false

	for _, column := range columnsToCheck {
		var columnCount int64
		err := db.Raw("SELECT count(*) FROM pragma_table_info('project_settings') WHERE name = ?", column).Scan(&columnCount).Error
		if err != nil {
			log.Printf("Warning: Failed to check column %s: %v", column, err)
			continue
		}
		if columnCount > 0 {
			hasDeprecatedColumns = true
			break
		}
	}

	// If deprecated columns exist, recreate the table without them
	if hasDeprecatedColumns {
		log.Println("Removing deprecated columns from project_settings (logo_url, dark_logo_url, primary_color, newsletter_enabled)...")

		// SQLite requires recreating the table to drop columns
		err := db.Transaction(func(tx *gorm.DB) error {
			// Create new table with correct schema
			if err := tx.Exec(`
				CREATE TABLE project_settings_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					title TEXT NOT NULL DEFAULT 'Changelog',
					favicon_url TEXT,
					website_url TEXT,
					current_theme_id TEXT,
					current_theme_version TEXT,
					created_at DATETIME,
					updated_at DATETIME,
					deleted_at DATETIME
				)
			`).Error; err != nil {
				return err
			}

			// Copy data from old table to new table (only the columns we want to keep)
			if err := tx.Exec(`
				INSERT INTO project_settings_new (
					id, title, favicon_url, website_url,
					current_theme_id, current_theme_version, created_at, updated_at, deleted_at
				)
				SELECT
					id, title, favicon_url, website_url,
					current_theme_id, current_theme_version, created_at, updated_at, deleted_at
				FROM project_settings
			`).Error; err != nil {
				return err
			}

			// Drop old table
			if err := tx.Exec("DROP TABLE project_settings").Error; err != nil {
				return err
			}

			// Rename new table to original name
			if err := tx.Exec("ALTER TABLE project_settings_new RENAME TO project_settings").Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("Warning: Failed to remove deprecated columns: %v", err)
		} else {
			log.Println("✓ Successfully removed deprecated columns from project_settings")
		}
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
