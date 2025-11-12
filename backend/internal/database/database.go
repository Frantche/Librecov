package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Frantche/Librecov/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Initialize sets up the database connection and runs migrations
func Initialize() (*gorm.DB, error) {
	// Get database configuration from environment
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "librecov_dev")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set global DB instance
	DB = db

	log.Println("Database connection established")

	// Run auto migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Run custom migrations first
	if err := runCustomMigrations(db); err != nil {
		return fmt.Errorf("failed to run custom migrations: %w", err)
	}

	// Run auto migrations
	return db.AutoMigrate(
		&models.User{},
		&models.UserToken{},
		&models.Project{},
		&models.ProjectShare{},
		&models.ProjectToken{},
		&models.Build{},
		&models.Job{},
		&models.JobFile{},
	)
}

// runCustomMigrations handles specific migration cases
func runCustomMigrations(db *gorm.DB) error {
	

	return nil
}


// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
