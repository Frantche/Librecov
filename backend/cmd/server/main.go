package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Frantche/Librecov/backend/internal/api"
	"github.com/Frantche/Librecov/backend/internal/auth"
	"github.com/Frantche/Librecov/backend/internal/database"
	"github.com/Frantche/Librecov/backend/internal/models"
	"github.com/Frantche/Librecov/backend/internal/session"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize OIDC provider
	oidcProvider, err := auth.NewOIDCProvider()
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	// If FIRST_ADMIN_EMAIL is set, ensure that user exists and is marked as admin
	firstAdmin := os.Getenv("FIRST_ADMIN_EMAIL")
	if firstAdmin != "" {
		var user models.User
		if err := db.Where("email = ?", firstAdmin).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("FIRST_ADMIN_EMAIL=%s set but user not found. Create the user via OIDC first.", firstAdmin)
			} else {
				log.Printf("Error looking up FIRST_ADMIN_EMAIL user: %v", err)
			}
		} else {
			if !user.Admin {
				user.Admin = true
				if err := db.Save(&user).Error; err != nil {
					log.Printf("Failed to set user %s as admin: %v", firstAdmin, err)
				} else {
					log.Printf("User %s marked as admin (FIRST_ADMIN_EMAIL)", firstAdmin)
				}
			} else {
				log.Printf("User %s already an admin", firstAdmin)
			}
		}
	}

	// Start session cleanup routine
	sessionStore := session.GetStore()
	sessionStore.StartCleanupRoutine()

	// Create Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router, db, oidcProvider)

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
