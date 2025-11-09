package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Build{},
		&models.Job{},
		&models.JobFile{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUploadCoverageSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create a test user and project
	user := models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Token: "user-token",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	project := models.Project{
		Name:   "Test Project",
		Token:  "test-project-token",
		UserID: user.ID,
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Prepare test coverage data
	coverageData := map[string]interface{}{
		"repo_token":     "test-project-token",
		"service_name":   "github-actions",
		"service_number": "123",
		"service_job_id": "job-123",
		"git": map[string]interface{}{
			"head": map[string]interface{}{
				"id":      "abc123def456",
				"message": "Test commit",
			},
			"branch": "main",
		},
		"source_files": []map[string]interface{}{
			{
				"name":     "main.go",
				"source":   "package main\n\nfunc main() {}\n",
				"coverage": []interface{}{nil, nil, 1, nil, 1, nil},
			},
			{
				"name":     "helper.go",
				"source":   "package main\n\nfunc helper() {}\n",
				"coverage": []interface{}{nil, nil, 1, nil, 0, nil},
			},
		},
	}

	jsonData, err := json.Marshal(coverageData)
	if err != nil {
		t.Fatalf("Failed to marshal coverage data: %v", err)
	}

	// Create test request
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	handler := NewJobHandler(db)
	router.POST("/upload/v2", handler.Upload)

	req := httptest.NewRequest("POST", "/upload/v2", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify database records
	var builds []models.Build
	db.Find(&builds)
	if len(builds) != 1 {
		t.Errorf("Expected 1 build, got %d", len(builds))
	}

	var jobs []models.Job
	db.Find(&jobs)
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	var files []models.JobFile
	db.Find(&files)
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check coverage calculation
	if jobs[0].CoverageRate == 0 {
		t.Error("Expected coverage rate to be calculated")
	}

	// Verify project coverage was updated
	var updatedProject models.Project
	db.First(&updatedProject, project.ID)
	if updatedProject.CoverageRate == 0 {
		t.Error("Expected project coverage rate to be updated")
	}
}

func TestUploadCoverageInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Prepare test coverage data with invalid token
	coverageData := map[string]interface{}{
		"repo_token":   "invalid-token",
		"service_name": "github-actions",
		"source_files": []map[string]interface{}{
			{
				"name":     "main.go",
				"source":   "package main\n",
				"coverage": []interface{}{nil, 1},
			},
		},
	}

	jsonData, err := json.Marshal(coverageData)
	if err != nil {
		t.Fatalf("Failed to marshal coverage data: %v", err)
	}

	// Create test request
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	handler := NewJobHandler(db)
	router.POST("/upload/v2", handler.Upload)

	req := httptest.NewRequest("POST", "/upload/v2", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUploadCoverageInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create test request with invalid JSON
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	handler := NewJobHandler(db)
	router.POST("/upload/v2", handler.Upload)

	req := httptest.NewRequest("POST", "/upload/v2", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
