package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/google/uuid"
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
	// Check if we need to rename o_id_c_subject to oidc_subject
	var count int64
	err := db.Raw("SELECT count(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'o_id_c_subject' AND table_schema = CURRENT_SCHEMA()").Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check for o_id_c_subject column: %w", err)
	}

	if count > 0 {
		log.Println("Renaming o_id_c_subject column to oidc_subject...")

		// First drop the unique index if it exists
		err = db.Exec("DROP INDEX IF EXISTS idx_users_o_id_c_subject").Error
		if err != nil {
			log.Printf("Warning: failed to drop index idx_users_o_id_c_subject: %v", err)
		}

		// Rename the column
		err = db.Exec("ALTER TABLE users RENAME COLUMN o_id_c_subject TO oidc_subject").Error
		if err != nil {
			return fmt.Errorf("failed to rename o_id_c_subject column: %w", err)
		}

		log.Println("Successfully renamed o_id_c_subject to oidc_subject")
	}

	// Migrate project IDs from integer to UUID
	if err := migrateProjectIDsToUUID(db); err != nil {
		return fmt.Errorf("failed to migrate project IDs to UUID: %w", err)
	}

	return nil
}

// migrateProjectIDsToUUID migrates existing projects from integer IDs to UUID strings
func migrateProjectIDsToUUID(db *gorm.DB) error {
	// Check if projects table exists and if id is integer type
	var columnType string
	err := db.Raw("SELECT data_type FROM information_schema.columns WHERE table_name = 'projects' AND column_name = 'id' AND table_schema = CURRENT_SCHEMA()").Scan(&columnType).Error
	if err != nil {
		log.Printf("Could not check projects.id column type: %v", err)
		return nil // Table might not exist yet
	}

	// If already varchar, skip migration
	if columnType == "character varying" {
		log.Println("Project IDs already migrated to UUID format")
		return nil
	}

	log.Println("Migrating project IDs from integer to UUID...")

	// Start a transaction for the migration
	return db.Transaction(func(tx *gorm.DB) error {
		// Create temporary tables for mapping
		tx.Exec("CREATE TEMP TABLE project_id_mapping (old_id INTEGER, new_id VARCHAR(36))")

		// Get all existing projects and generate UUIDs for them
		type TempProject struct {
			ID uint
		}
		var projects []TempProject
		if err := tx.Raw("SELECT id FROM projects").Scan(&projects).Error; err != nil {
			log.Printf("No existing projects to migrate")
			return nil
		}

		// Generate UUID mappings
		for _, p := range projects {
			newID := uuid.NewString()
			tx.Exec("INSERT INTO project_id_mapping (old_id, new_id) VALUES (?, ?)", p.ID, newID)
		}

		// Create new tables with correct schema
		tx.Exec(`CREATE TABLE projects_new (
			id VARCHAR(36) PRIMARY KEY,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			current_branch VARCHAR(255),
			base_url VARCHAR(255),
			coverage_rate DOUBLE PRECISION,
			user_id INTEGER
		)`)

		tx.Exec(`CREATE TABLE builds_new (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			project_id VARCHAR(36) NOT NULL,
			build_num INTEGER NOT NULL,
			branch VARCHAR(255),
			commit_sha VARCHAR(255),
			commit_msg TEXT,
			coverage_rate DOUBLE PRECISION
		)`)

		tx.Exec(`CREATE TABLE project_tokens_new (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			project_id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			last_used TIMESTAMP
		)`)

		tx.Exec(`CREATE TABLE project_shares_new (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			project_id VARCHAR(36) NOT NULL,
			group_name VARCHAR(255) NOT NULL
		)`)

		// Copy data with UUID mapping
		tx.Exec(`INSERT INTO projects_new SELECT m.new_id, p.created_at, p.updated_at, p.deleted_at, p.name, p.token, p.current_branch, p.base_url, p.coverage_rate, p.user_id
			FROM projects p
			JOIN project_id_mapping m ON p.id = m.old_id`)

		if len(projects) > 0 {
			tx.Exec(`INSERT INTO builds_new SELECT b.id, b.created_at, b.updated_at, b.deleted_at, m.new_id, b.build_num, b.branch, b.commit_sha, b.commit_msg, b.coverage_rate
				FROM builds b
				JOIN project_id_mapping m ON b.project_id = m.old_id`)

			tx.Exec(`INSERT INTO project_tokens_new SELECT pt.id, pt.created_at, pt.updated_at, pt.deleted_at, m.new_id, pt.name, pt.token, pt.last_used
				FROM project_tokens pt
				JOIN project_id_mapping m ON pt.project_id = m.old_id`)

			tx.Exec(`INSERT INTO project_shares_new SELECT ps.id, ps.created_at, ps.updated_at, ps.deleted_at, m.new_id, ps.group_name
				FROM project_shares ps
				JOIN project_id_mapping m ON ps.project_id = m.old_id`)
		}

		// Drop old tables
		tx.Exec("DROP TABLE IF EXISTS project_shares")
		tx.Exec("DROP TABLE IF EXISTS project_tokens")
		tx.Exec("DROP TABLE IF EXISTS builds")
		tx.Exec("DROP TABLE IF EXISTS projects")

		// Rename new tables
		tx.Exec("ALTER TABLE projects_new RENAME TO projects")
		tx.Exec("ALTER TABLE builds_new RENAME TO builds")
		tx.Exec("ALTER TABLE project_tokens_new RENAME TO project_tokens")
		tx.Exec("ALTER TABLE project_shares_new RENAME TO project_shares")

		// Recreate indexes
		tx.Exec("CREATE INDEX idx_projects_deleted_at ON projects(deleted_at)")
		tx.Exec("CREATE INDEX idx_projects_user_id ON projects(user_id)")
		tx.Exec("CREATE INDEX idx_builds_deleted_at ON builds(deleted_at)")
		tx.Exec("CREATE INDEX idx_builds_project_id ON builds(project_id)")
		tx.Exec("CREATE INDEX idx_project_tokens_deleted_at ON project_tokens(deleted_at)")
		tx.Exec("CREATE INDEX idx_project_tokens_project_id ON project_tokens(project_id)")
		tx.Exec("CREATE INDEX idx_project_shares_deleted_at ON project_shares(deleted_at)")
		tx.Exec("CREATE INDEX idx_project_shares_project_id ON project_shares(project_id)")
		tx.Exec("CREATE INDEX idx_project_shares_group_name ON project_shares(group_name)")

		log.Println("Successfully migrated project IDs to UUID format")
		return nil
	})
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
