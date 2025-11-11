package models

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestUserModel(t *testing.T) {
	user := User{
		Email: "test@example.com",
		Name:  "Test User",
		Admin: false,
		Token: "test-token",
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name to be Test User, got %s", user.Name)
	}

	if user.Admin {
		t.Error("Expected admin to be false")
	}
}

func TestProjectModel(t *testing.T) {
	project := Project{
		Name:          "Test Project",
		Token:         "project-token",
		CurrentBranch: "main",
		BaseURL:       "https://github.com/test/repo",
		CoverageRate:  85.5,
		UserID:        1,
	}

	if project.Name != "Test Project" {
		t.Errorf("Expected name to be Test Project, got %s", project.Name)
	}

	if project.CoverageRate != 85.5 {
		t.Errorf("Expected coverage rate to be 85.5, got %f", project.CoverageRate)
	}
}

func TestBuildModel(t *testing.T) {
	build := Build{
		ProjectID:    "test-project-uuid",
		BuildNum:     42,
		Branch:       "main",
		CommitSHA:    "abc123",
		CommitMsg:    "Test commit",
		CoverageRate: 90.0,
	}

	if build.BuildNum != 42 {
		t.Errorf("Expected build num to be 42, got %d", build.BuildNum)
	}

	if build.Branch != "main" {
		t.Errorf("Expected branch to be main, got %s", build.Branch)
	}
}

func TestJobModel(t *testing.T) {
	job := Job{
		BuildID:      1,
		JobNumber:    "1.1",
		CoverageRate: 88.5,
		Data:         `{"test": "data"}`,
	}

	if job.JobNumber != "1.1" {
		t.Errorf("Expected job number to be 1.1, got %s", job.JobNumber)
	}

	if job.CoverageRate != 88.5 {
		t.Errorf("Expected coverage rate to be 88.5, got %f", job.CoverageRate)
	}
}

func TestJobFileModel(t *testing.T) {
	file := JobFile{
		JobID:        1,
		Name:         "main.go",
		Coverage:     "[1, 2, 3]",
		Source:       "package main\n\nfunc main() {}",
		CoverageRate: 100.0,
	}

	if file.Name != "main.go" {
		t.Errorf("Expected name to be main.go, got %s", file.Name)
	}

	if file.CoverageRate != 100.0 {
		t.Errorf("Expected coverage rate to be 100.0, got %f", file.CoverageRate)
	}
}

func TestUserTimestamps(t *testing.T) {
	now := time.Now()
	user := User{
		Email:     "test@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestProjectRelationships(t *testing.T) {
	project := Project{
		Name:   "Test Project",
		UserID: 1,
		User: User{
			ID:    1,
			Email: "test@example.com",
		},
	}

	if project.User.ID != 1 {
		t.Errorf("Expected user ID to be 1, got %d", project.User.ID)
	}
}

func TestBuildRelationships(t *testing.T) {
	build := Build{
		ProjectID: "test-project-uuid",
		Project: Project{
			ID:   "test-project-uuid",
			Name: "Test Project",
		},
	}

	if build.Project.ID != "test-project-uuid" {
		t.Errorf("Expected project ID to be test-project-uuid, got %s", build.Project.ID)
	}
}

func TestSoftDelete(t *testing.T) {
	user := User{
		Email:     "test@example.com",
		DeletedAt: gorm.DeletedAt{},
	}

	if user.DeletedAt.Valid {
		t.Error("Expected DeletedAt to be invalid for non-deleted record")
	}
}
