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
	// Auto-migrate the schema
	if err := DB.AutoMigrate(
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
		&models.FooterLink{},
	); err != nil {
		return err
	}

	// Initialize default email templates
	if err := models.InitializeDefaultEmailTemplates(DB); err != nil {
		log.Printf("Warning: Failed to initialize default email templates: %v", err)
	} else {
		log.Println("Successfully initialized default email templates")
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
