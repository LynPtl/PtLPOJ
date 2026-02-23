package storage

import (
	"fmt"
	"log"
	
	"pt_lpoj/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the SQLite database connection using GORM.
func InitDB(dsn string) error {
	var err error

	// We append the WAL mode PRAGMAs to the DSN for concurrent writing performance
	// This is critical for Judge Workers reporting statuses back simultaneously.
	walDSN := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", dsn)

	DB, err = gorm.Open(sqlite.Open(walDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("SQLite database connected successfully with WAL mode.")

	// Auto-migrate our schemas
	err = DB.AutoMigrate(&models.User{}, &models.Submission{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	log.Println("Database schemas auto-migrated successfully.")
	return nil
}
