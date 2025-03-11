package routes

import (
	"database/sql"
	"net/http"

	"forum/handlers"
	"forum/middleware"
	"forum/models"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(db *sql.DB) http.Handler {
	// Create repositories
	userRepo := models.NewUserRepository(db)
	sessionRepo := models.NewSessionRepository(db)

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo)

	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

	// Create router (using standard net/http for simplicity)
	mux := http.NewServeMux()

	// Define auth routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	// Apply middleware to all routes
	return authMiddleware.Authenticate(mux)
}