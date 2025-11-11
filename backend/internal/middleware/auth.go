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
					c.Set("user_id", user.ID)
					c.Next()
					return
				}
			}
		}

		// Try to authenticate via API token (Bearer token)
		token := extractToken(c)
		if token != "" {
			// Check if it's a user token
			var userToken models.UserToken
			if err := database.DB.Where("token = ?", token).First(&userToken).Error; err == nil {
				// Valid user token, get user from database
				var user models.User
				if err := database.DB.First(&user, userToken.UserID).Error; err == nil {
					// Update last used timestamp
					database.DB.Model(&userToken).Update("last_used", database.DB.NowFunc())
					c.Set("user", &user)
					c.Set("user_id", user.ID)
					c.Next()
					return
				}
			}

			// Check if it's a project token
			var projectToken models.ProjectToken
			if err := database.DB.Preload("Project").Where("token = ?", token).First(&projectToken).Error; err == nil {
				// Valid project token, get user from project
				var user models.User
				if err := database.DB.First(&user, projectToken.Project.UserID).Error; err == nil {
					// Update last used timestamp
					database.DB.Model(&projectToken).Update("last_used", database.DB.NowFunc())
					c.Set("user", &user)
					c.Set("user_id", user.ID)
					c.Set("project_id", projectToken.ProjectID)
					c.Next()
					return
				}
			}

			// Fall back to old token-based authentication (user.token field)
			var user models.User
			if err := database.DB.Where("token = ?", token).First(&user).Error; err == nil {
				c.Set("user", &user)
				c.Set("user_id", user.ID)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
		c.Abort()
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
