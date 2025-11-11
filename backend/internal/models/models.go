package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
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
	OIDCSubject   string `gorm:"column:oidc_subject;uniqueIndex" json:"-"` // OIDC subject identifier
	EmailVerified bool   `gorm:"default:false" json:"email_verified"`
	Groups        string `gorm:"type:text" json:"groups"` // JSON array of group names from OIDC token

	// Relationships
	Projects   []Project   `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	UserTokens []UserToken `gorm:"foreignKey:UserID" json:"tokens,omitempty"`
}

// UserToken represents an API token for a user
type UserToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID   uint       `gorm:"not null;index" json:"user_id"`
	Name     string     `gorm:"not null" json:"name"`
	Token    string     `gorm:"uniqueIndex;not null" json:"token,omitempty"`
	LastUsed *time.Time `json:"last_used,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// GenerateToken generates a secure random token
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Project represents a code coverage project
type Project struct {
	ID        string         `gorm:"type:varchar(36);primarykey" json:"id"`
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
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Builds        []Build        `gorm:"foreignKey:ProjectID" json:"builds,omitempty"`
	ProjectTokens []ProjectToken `gorm:"foreignKey:ProjectID" json:"tokens,omitempty"`
	ProjectShares []ProjectShare `gorm:"foreignKey:ProjectID" json:"shares,omitempty"`
}

// BeforeCreate hook to generate UUID v7 for new projects
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

// ProjectShare represents group-based sharing of a project
type ProjectShare struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID string `gorm:"type:varchar(36);not null;index" json:"project_id"`
	GroupName string `gorm:"not null;index" json:"group_name"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

// ProjectToken represents an API token for a project
type ProjectToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID string     `gorm:"type:varchar(36);not null;index" json:"project_id"`
	Name      string     `gorm:"not null" json:"name"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token,omitempty"`
	LastUsed  *time.Time `json:"last_used,omitempty"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

// Build represents a coverage build for a project
type Build struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID    string  `gorm:"type:varchar(36);not null;index" json:"project_id"`
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
	Build Build     `gorm:"foreignKey:BuildID" json:"build,omitempty"`
	Files []JobFile `gorm:"foreignKey:JobID" json:"files,omitempty"`
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
