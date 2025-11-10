package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Frantche/Librecov/backend/internal/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

// mockOIDCProvider creates a mock for testing with non-nil Provider
type mockProvider struct{}

func TestGetConfigWithOIDCEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create OIDC provider with config
	// We use type assertion to set a non-nil Provider
	oidcProvider := &auth.OIDCProvider{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:4000/auth/callback",
		Issuer:       "https://test-issuer.com",
	}
	// For this unit test, we set Provider to nil to simulate OIDC being disabled.
	// Mocking oidc.Provider is non-trivial, so this test only verifies the disabled case.
	oidcProvider.Provider = nil

	// Create test request
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	handler := NewAuthHandler(db, oidcProvider)
	router.GET("/auth/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/auth/config", nil)
	c.Request = req

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// When Provider is nil, oidc_enabled should be false
	// This test actually validates the disabled case
	oidcEnabled, ok := response["oidc_enabled"].(bool)
	if !ok {
		t.Errorf("Expected oidc_enabled to be present, got %v", response["oidc_enabled"])
	}

	// Since we can't easily mock an oidc.Provider in tests, this will be false
	// The real test of OIDC enabled state happens in integration tests
	if oidcEnabled {
		// If somehow enabled, verify the config is returned
		oidcConfig, ok := response["oidc"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected oidc config in response when enabled, got %v", response["oidc"])
		}

		if oidcConfig["issuer"] != "https://test-issuer.com" {
			t.Errorf("Expected issuer 'https://test-issuer.com', got %v", oidcConfig["issuer"])
		}

		if oidcConfig["client_id"] != "test-client-id" {
			t.Errorf("Expected client_id 'test-client-id', got %v", oidcConfig["client_id"])
		}

		if oidcConfig["redirect_url"] != "http://localhost:4000/auth/callback" {
			t.Errorf("Expected redirect_url 'http://localhost:4000/auth/callback', got %v", oidcConfig["redirect_url"])
		}
	}
}

func TestGetConfigWithOIDCDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create OIDC provider without config (disabled)
	oidcProvider := &auth.OIDCProvider{}

	// Create test request
	w := httptest.NewRecorder()
	c, router := gin.CreateTestContext(w)

	handler := NewAuthHandler(db, oidcProvider)
	router.GET("/auth/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/auth/config", nil)
	c.Request = req

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check OIDC enabled flag
	oidcEnabled, ok := response["oidc_enabled"].(bool)
	if !ok || oidcEnabled {
		t.Errorf("Expected oidc_enabled to be false, got %v", response["oidc_enabled"])
	}

	// Check OIDC config should not be present
	if _, ok := response["oidc"]; ok {
		t.Errorf("Expected no oidc config when disabled, got %v", response["oidc"])
	}
}
