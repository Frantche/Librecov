package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func TestExtractTokenFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{
			name:          "Bearer token",
			authHeader:    "Bearer test-token-123",
			expectedToken: "test-token-123",
		},
		{
			name:          "Simple token",
			authHeader:    "test-token-456",
			expectedToken: "test-token-456",
		},
		{
			name:          "Empty header",
			authHeader:    "",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			c.Request = req

			token := extractToken(c)

			if token != tt.expectedToken {
				t.Errorf("Expected token %s, got %s", tt.expectedToken, token)
			}
		})
	}
}

func TestExtractTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/test?token=query-token", nil)
	c.Request = req

	token := extractToken(c)

	if token != "query-token" {
		t.Errorf("Expected token query-token, got %s", token)
	}
}

func TestGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("User exists in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		expectedUser := &models.User{
			ID:    1,
			Email: "test@example.com",
			Admin: true,
		}

		c.Set("user", expectedUser)

		user, exists := GetCurrentUser(c)

		if !exists {
			t.Error("Expected user to exist")
		}

		if user.ID != expectedUser.ID {
			t.Errorf("Expected user ID %d, got %d", expectedUser.ID, user.ID)
		}

		if user.Email != expectedUser.Email {
			t.Errorf("Expected email %s, got %s", expectedUser.Email, user.Email)
		}
	})

	t.Run("User does not exist in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		user, exists := GetCurrentUser(c)

		if exists {
			t.Error("Expected user to not exist")
		}

		if user != nil {
			t.Error("Expected user to be nil")
		}
	})
}

func TestAuthMiddlewareWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", AuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestOptionalAuthMiddlewareWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.GET("/test", OptionalAuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	r.ServeHTTP(w, req)

	// Should pass through without authentication
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
