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
	// @Summary		Health check
	// @Description	Check if the API is running
	// @Tags			system
	// @Accept			json
	// @Produce		json
	// @Success		200	{object}	map[string]string
	// @Router			/health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Authentication routes
	authGroup := router.Group("/auth")
	{
		authHandler := NewAuthHandler(db, oidcProvider)
		authGroup.GET("/config", authHandler.GetConfig)        // Public endpoint for auth config
		authGroup.GET("/login", authHandler.Login)             // Initiate OIDC login
		authGroup.GET("/callback", authHandler.Callback)       // OIDC callback handler
		authGroup.POST("/logout", authHandler.Logout)          // Logout
		authGroup.POST("/refresh", authHandler.RefreshSession) // Refresh session
		authGroup.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
		authGroup.GET("/groups", middleware.AuthMiddleware(), authHandler.GetUserGroups)
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
			// User tokens
			server := &Server{db: db}
			protected.GET("/user/tokens", server.GetUserTokens)
			protected.POST("/user/tokens", server.CreateUserToken)
			protected.DELETE("/user/tokens/:id", server.DeleteUserToken)

			// Projects
			projectHandler := NewProjectHandler(db)
			protected.GET("/projects", projectHandler.List)
			protected.POST("/projects", projectHandler.Create)
			protected.GET("/projects/:id", projectHandler.Get)
			protected.PUT("/projects/:id", projectHandler.Update)
			protected.DELETE("/projects/:id", projectHandler.Delete)

			// Project tokens
			protected.GET("/projects/:id/tokens", server.GetProjectTokens)
			protected.POST("/projects/:id/tokens", server.CreateProjectToken)
			protected.DELETE("/projects/:id/tokens/:tokenId", server.DeleteProjectToken)
			protected.POST("/projects/:id/refresh-token", server.RefreshProjectToken)

			// Project shares
			protected.GET("/projects/:id/shares", projectHandler.GetShares)
			protected.POST("/projects/:id/shares", projectHandler.CreateShare)
			protected.DELETE("/projects/:id/shares/:shareId", projectHandler.DeleteShare)

			// Project ownership transfer
			protected.POST("/projects/:id/transfer-ownership", projectHandler.TransferOwnership)

			// Users for ownership transfer
			userHandler := NewUserHandler(db)
			protected.GET("/users", userHandler.ListForOwnershipTransfer)

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

			projectHandler := NewProjectHandler(db)
			admin.GET("/projects", projectHandler.ListAll)
		}
	}

	// Coveralls-compatible upload endpoint
	router.POST("/upload/v2", NewJobHandler(db).Upload)

	// Webhook endpoint
	router.POST("/webhook", NewWebhookHandler(db).HandleWebhook)

	// Badge endpoint
	router.GET("/projects/:id/badge.svg", NewBadgeHandler(db).GetBadge)

	// Serve frontend static files in production
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// Serve index.html for all unmatched routes (SPA fallback)
	router.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
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
