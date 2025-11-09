package api

import (
	"github.com/Frantche/Librecov/backend/internal/auth"
	"github.com/Frantche/Librecov/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, db *gorm.DB, oidcProvider *auth.OIDCProvider) {
	// CORS middleware
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Authentication routes
	authGroup := router.Group("/auth")
	{
		authHandler := NewAuthHandler(db, oidcProvider)
		authGroup.GET("/login", authHandler.Login)
		authGroup.GET("/callback", authHandler.Callback)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.POST("/jobs", NewJobHandler(db).CreateJob)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Projects
			projectHandler := NewProjectHandler(db)
			protected.GET("/projects", projectHandler.List)
			protected.POST("/projects", projectHandler.Create)
			protected.GET("/projects/:id", projectHandler.Get)
			protected.PUT("/projects/:id", projectHandler.Update)
			protected.DELETE("/projects/:id", projectHandler.Delete)

			// Builds
			buildHandler := NewBuildHandler(db)
			protected.GET("/projects/:id/builds", buildHandler.List)
			protected.GET("/builds/:id", buildHandler.Get)

			// Jobs
			jobHandler := NewJobHandler(db)
			protected.GET("/jobs/:id", jobHandler.Get)
			protected.GET("/builds/:id/jobs", jobHandler.ListByBuild)

			// Files
			fileHandler := NewFileHandler(db)
			protected.GET("/jobs/:id/files", fileHandler.List)
			protected.GET("/files/:id", fileHandler.Get)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			userHandler := NewUserHandler(db)
			admin.GET("/users", userHandler.List)
			admin.GET("/users/:id", userHandler.Get)
			admin.PUT("/users/:id", userHandler.Update)
			admin.DELETE("/users/:id", userHandler.Delete)
		}
	}

	// Coveralls-compatible upload endpoint
	router.POST("/upload/v2", NewJobHandler(db).Upload)

	// Webhook endpoint
	router.POST("/webhook", NewWebhookHandler(db).HandleWebhook)

	// Badge endpoint
	router.GET("/projects/:id/badge.svg", NewBadgeHandler(db).GetBadge)

	// Serve frontend in production
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Not found"})
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
