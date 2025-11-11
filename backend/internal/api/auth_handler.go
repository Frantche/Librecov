package api

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Frantche/Librecov/backend/internal/auth"
	"github.com/Frantche/Librecov/backend/internal/middleware"
	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/Frantche/Librecov/backend/internal/session"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db           *gorm.DB
	oidcProvider *auth.OIDCProvider
	sessionStore *session.Store
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, oidcProvider *auth.OIDCProvider) *AuthHandler {
	return &AuthHandler{
		db:           db,
		oidcProvider: oidcProvider,
		sessionStore: session.GetStore(),
	}
}

// Login initiates the OIDC login flow with PKCE
func (h *AuthHandler) Login(c *gin.Context) {
	if !h.oidcProvider.IsEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC not configured"})
		return
	}

	// Generate state token (CSRF protection)
	state := generateRandomString(32)

	// Generate PKCE verifier and challenge
	pkce, err := auth.GeneratePKCE()
	if err != nil {
		log.Printf("Failed to generate PKCE: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate login"})
		return
	}

	// Store state and PKCE verifier in session store (server-side)
	h.sessionStore.StoreState(state, pkce.Verifier)

	// Get cookie domain from environment or use empty string for current domain
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	isSecure := os.Getenv("COOKIE_SECURE") == "true" // Set to true in production with HTTPS

	// Set state cookie (for CSRF validation)
	c.SetCookie(
		"oidc_state",           // name
		state,                  // value
		600,                    // maxAge (10 minutes)
		"/",                    // path
		cookieDomain,           // domain
		isSecure,               // secure (true for HTTPS)
		true,                   // httpOnly
		// SameSite=Lax allows the cookie to be sent during OIDC redirects
	)
	c.SetSameSite(http.SameSiteLaxMode)

	// Redirect to OIDC provider with PKCE challenge
	authURL := h.oidcProvider.GetAuthURL(state, pkce.Challenge)
	log.Printf("Redirecting to OIDC provider: %s", authURL)
	c.Redirect(http.StatusFound, authURL)
}

// Callback handles the OIDC callback with PKCE verification
func (h *AuthHandler) Callback(c *gin.Context) {
	if !h.oidcProvider.IsEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC not configured"})
		return
	}

	// Verify state (CSRF protection)
	state := c.Query("state")
	storedState, err := c.Cookie("oidc_state")
	if err != nil {
		log.Printf("Error retrieving state cookie: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "State cookie not found"})
		return
	}

	if state != storedState {
		log.Printf("State mismatch: query=%s, cookie=%s", state, storedState)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Clear state cookie
	c.SetCookie("oidc_state", "", -1, "/", os.Getenv("COOKIE_DOMAIN"), os.Getenv("COOKIE_SECURE") == "true", true)

	// Retrieve PKCE verifier from session store
	stateData, err := h.sessionStore.GetState(state)
	if err != nil {
		log.Printf("Failed to retrieve state data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired state"})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		log.Printf("Missing authorization code")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Exchange code for token with PKCE verifier
	oauth2Token, err := h.oidcProvider.ExchangeCode(c.Request.Context(), code, stateData.PKCEVerifier)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange authorization code"})
		return
	}

	// Extract and verify ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Printf("No id_token in response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No ID token in response"})
		return
	}

	idToken, err := h.oidcProvider.VerifyIDToken(c.Request.Context(), rawIDToken)
	if err != nil {
		log.Printf("ID token verification failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID token"})
		return
	}

	// Extract claims
	claims, err := auth.ExtractClaims(idToken)
	if err != nil {
		log.Printf("Failed to extract claims: %v", err)
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

		// If FIRST_ADMIN_EMAIL is set and matches this user's email, mark as admin
		firstAdmin := os.Getenv("FIRST_ADMIN_EMAIL")
		if firstAdmin != "" && claims.Email != "" {
			if strings.EqualFold(strings.TrimSpace(claims.Email), strings.TrimSpace(firstAdmin)) {
				user.Admin = true
			}
		}
		if err := h.db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		log.Printf("Created new user: %s (ID: %d)", user.Email, user.ID)
	} else if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		log.Printf("User found: %s (ID: %d)", user.Email, user.ID)
	}

	// Create session
	sessionID := h.sessionStore.CreateSession(user.ID, oauth2Token)

	// Get cookie settings
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	isSecure := os.Getenv("COOKIE_SECURE") == "true"

	// Set session cookie (HttpOnly, Secure, SameSite=Strict)
	c.SetCookie(
		"session_id",      // name
		sessionID,         // value
		3600*24,           // maxAge (24 hours)
		"/",               // path
		cookieDomain,      // domain
		isSecure,          // secure
		true,              // httpOnly
	)
	c.SetSameSite(http.SameSiteStrictMode)

	log.Printf("Session created for user %d: %s", user.ID, sessionID)

	// Redirect to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "/"
	}
	c.Redirect(http.StatusFound, frontendURL)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		// Delete session from store
		h.sessionStore.DeleteSession(sessionID)
	}

	// Clear session cookie
	c.SetCookie(
		"session_id",
		"",
		-1,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		os.Getenv("COOKIE_SECURE") == "true",
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshSession refreshes the user's session
func (h *AuthHandler) RefreshSession(c *gin.Context) {
	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	// Retrieve session
	sess, err := h.sessionStore.GetSession(sessionID)
	if err != nil {
		log.Printf("Session retrieval failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	// Get user from database
	var user models.User
	if err := h.db.First(&user, sess.UserID).Error; err != nil {
		log.Printf("User not found: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Return user info and session status
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":             user.ID,
			"email":          user.Email,
			"name":           user.Name,
			"admin":          user.Admin,
			"email_verified": user.EmailVerified,
		},
		"session_valid": true,
	})
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
