package api

import (
"bytes"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/Frantche/Librecov/backend/internal/models"
"github.com/gin-gonic/gin"
)

func TestRefreshProjectToken(t *testing.T) {
gin.SetMode(gin.TestMode)

// Setup test database
db, err := setupTestDB()
if err != nil {
t.Fatalf("Failed to setup test database: %v", err)
}

// Create a test user
user := models.User{
Email: "test@example.com",
Name:  "Test User",
Token: "user-token",
}
if err := db.Create(&user).Error; err != nil {
t.Fatalf("Failed to create test user: %v", err)
}

// Create a test project
project := models.Project{
Name:   "Test Project",
Token:  "old-token-value",
UserID: user.ID,
}
if err := db.Create(&project).Error; err != nil {
t.Fatalf("Failed to create test project: %v", err)
}

// Create test request
w := httptest.NewRecorder()
c, router := gin.CreateTestContext(w)

server := &Server{db: db}
router.POST("/projects/:id/refresh-token", func(c *gin.Context) {
c.Set("user_id", user.ID)
server.RefreshProjectToken(c)
})

req := httptest.NewRequest("POST", "/projects/"+project.ID+"/refresh-token", bytes.NewBuffer([]byte("{}")))
req.Header.Set("Content-Type", "application/json")
c.Request = req

router.ServeHTTP(w, req)

// Check response
if w.Code != http.StatusOK {
t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
}

// Verify the response contains a new token
var response map[string]interface{}
err = json.Unmarshal(w.Body.Bytes(), &response)
if err != nil {
t.Fatalf("Failed to unmarshal response: %v", err)
}

newToken, ok := response["token"].(string)
if !ok || newToken == "" {
t.Error("Expected response to contain a new token")
}

if newToken == "old-token-value" {
t.Error("Expected new token to be different from old token")
}

// Verify database was updated
var updatedProject models.Project
db.First(&updatedProject, "id = ?", project.ID)
if updatedProject.Token == "old-token-value" {
t.Error("Expected project token to be updated in database")
}

if updatedProject.Token != newToken {
t.Errorf("Expected project token to match response token, got %s and %s", updatedProject.Token, newToken)
}
}
