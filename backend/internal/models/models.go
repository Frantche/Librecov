package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email         string `gorm:"uniqueIndex;not null" json:"email"`
	Name          string `json:"name"`
	Admin         bool   `gorm:"default:false" json:"admin"`
	Token         string `gorm:"uniqueIndex" json:"token,omitempty"`
	OIDCSubject   string `gorm:"uniqueIndex" json:"-"` // OIDC subject identifier
	EmailVerified bool   `gorm:"default:false" json:"email_verified"`

	// Relationships
	Projects []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"`
}

// Project represents a code coverage project
type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name          string  `gorm:"not null" json:"name"`
	Token         string  `gorm:"uniqueIndex;not null" json:"token"`
	CurrentBranch string  `json:"current_branch"`
	BaseURL       string  `json:"base_url"`
	CoverageRate  float64 `json:"coverage_rate"`
	UserID        uint    `json:"user_id"`

	// Relationships
	User   User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Builds []Build `gorm:"foreignKey:ProjectID" json:"builds,omitempty"`
}

// Build represents a coverage build for a project
type Build struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID    uint    `gorm:"not null;index" json:"project_id"`
	BuildNum     int     `gorm:"not null" json:"build_num"`
	Branch       string  `json:"branch"`
	CommitSHA    string  `json:"commit_sha"`
	CommitMsg    string  `json:"commit_msg"`
	CoverageRate float64 `json:"coverage_rate"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Jobs    []Job   `gorm:"foreignKey:BuildID" json:"jobs,omitempty"`
}

// Job represents a coverage job within a build
type Job struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BuildID      uint    `gorm:"not null;index" json:"build_id"`
	JobNumber    string  `json:"job_number"`
	CoverageRate float64 `json:"coverage_rate"`
	Data         string  `gorm:"type:text" json:"data"` // JSON data

	// Relationships
	Build Build      `gorm:"foreignKey:BuildID" json:"build,omitempty"`
	Files []JobFile  `gorm:"foreignKey:JobID" json:"files,omitempty"`
}

// JobFile represents a file in a coverage job
type JobFile struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	JobID        uint    `gorm:"not null;index" json:"job_id"`
	Name         string  `gorm:"not null" json:"name"`
	Coverage     string  `gorm:"type:text" json:"coverage"` // JSON array
	Source       string  `gorm:"type:text" json:"source"`
	CoverageRate float64 `json:"coverage_rate"`

	// Relationships
	Job Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}
