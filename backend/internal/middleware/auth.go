package middleware

import (
	"net/http"
	"strings"

	"github.com/Frantche/Librecov/backend/internal/database"
	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/Frantche/Librecov/backend/internal/session"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the authentication token or session
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, try to authenticate via session cookie
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			sessionStore := session.GetStore()
			sess, err := sessionStore.GetSession(sessionID)
			if err == nil {
				// Valid session, get user from database
				var user models.User
				if err := database.DB.First(&user, sess.UserID).Error; err == nil {
					// Set user in context
					c.Set("user", &user)
					c.Next()
					return
				}
			}
		}

		// Fall back to token-based authentication
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// Find user by token
		var user models.User
		if err := database.DB.Where("token = ?", token).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", &user)
		c.Next()
	}
}

// OptionalAuthMiddleware checks for authentication but doesn't require it
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, try to authenticate via session cookie
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			sessionStore := session.GetStore()
			sess, err := sessionStore.GetSession(sessionID)
			if err == nil {
				// Valid session, get user from database
				var user models.User
				if err := database.DB.First(&user, sess.UserID).Error; err == nil {
					c.Set("user", &user)
					c.Next()
					return
				}
			}
		}

		// Fall back to token-based authentication
		token := extractToken(c)
		if token != "" {
			var user models.User
			if err := database.DB.Where("token = ?", token).First(&user).Error; err == nil {
				c.Set("user", &user)
			}
		}
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok || !user.Admin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts the token from the request
func extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
		return authHeader
	}

	// Check query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Check form parameter
	return c.PostForm("token")
}

// GetCurrentUser returns the current authenticated user from context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*models.User)
	return user, ok
}
