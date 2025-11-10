package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/Frantche/Librecov/backend/internal/auth"
	"github.com/Frantche/Librecov/backend/internal/middleware"
	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db           *gorm.DB
	oidcProvider *auth.OIDCProvider
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, oidcProvider *auth.OIDCProvider) *AuthHandler {
	return &AuthHandler{
		db:           db,
		oidcProvider: oidcProvider,
	}
}

// Login initiates the OIDC login flow
func (h *AuthHandler) Login(c *gin.Context) {
	if !h.oidcProvider.IsEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC not configured"})
		return
	}

	// Generate state token
	state := generateRandomString(32)

	// Store state in session (simplified, should use proper session management)
	c.SetCookie("oidc_state", state, 600, "/", "", false, true)

	// Redirect to OIDC provider
	authURL := h.oidcProvider.GetAuthURL(state)
	c.Redirect(http.StatusFound, authURL)
}

// Callback handles the OIDC callback
func (h *AuthHandler) Callback(c *gin.Context) {
	if !h.oidcProvider.IsEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC not configured"})
		return
	}

	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oidc_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Exchange code for token
	code := c.Query("code")
	oauth2Token, err := h.oidcProvider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token in response"})
		return
	}

	// Verify ID token
	idToken, err := h.oidcProvider.VerifyIDToken(c.Request.Context(), rawIDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID token"})
		return
	}

	// Extract claims
	claims, err := auth.ExtractClaims(idToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract claims"})
		return
	}

	// Find or create user
	var user models.User
	result := h.db.Where("oidc_subject = ?", claims.Sub).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:         claims.Email,
			Name:          claims.Name,
			OIDCSubject:   claims.Sub,
			EmailVerified: claims.EmailVerified,
			Token:         generateRandomString(32),
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return user with token
	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": user.Token,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// Me returns the current user info
func (h *AuthHandler) Me(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetConfig returns authentication configuration including OIDC info
func (h *AuthHandler) GetConfig(c *gin.Context) {
	response := gin.H{
		"oidc_enabled": h.oidcProvider.IsEnabled(),
	}

	if h.oidcProvider.IsEnabled() {
		response["oidc"] = gin.H{
			"issuer":       h.oidcProvider.Issuer,
			"client_id":    h.oidcProvider.ClientID,
			"redirect_url": h.oidcProvider.RedirectURL,
		}
	}

	c.JSON(http.StatusOK, response)
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
