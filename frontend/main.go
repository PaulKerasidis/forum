package main

import (
	"fmt"
	"log"
	"net/http"

	"frontend-service/config"
	"frontend-service/internal/routes"
	"frontend-service/internal/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create base client
	baseClient := services.NewBaseClient(cfg.APIBaseURL)

	// Create specialized services with config access
	authService := services.NewAuthService(baseClient, cfg) // CHANGED: Pass config
	postService := services.NewPostService(baseClient)
	categoryService := services.NewCategoryService(baseClient)
	userService := services.NewUserService(baseClient)
	commentService := services.NewCommentService(baseClient)
	postReactionService := services.NewPostReactionService(baseClient)
	commentReactionService := services.NewCommentReactionService(baseClient)

	// Create template service
	templateService, err := services.NewTemplateService(cfg.TemplatesDir)
	if err != nil {
		log.Fatal("Failed to create template service:", err)
	}

	// Setup routes with all services including config for session name
	mux := routes.SetupRoutes(authService, postService, categoryService,
		userService, commentService, postReactionService,
		commentReactionService, templateService, cfg) // CHANGED: Pass config

	// Create server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	fmt.Printf("Frontend server starting on port %s\n", cfg.Port)
	fmt.Printf("Backend API URL: %s\n", cfg.APIBaseURL)
	fmt.Printf("Session Cookie Name: %s\n", cfg.SessionName) // ADDED: Debug info

	// Start server
	log.Fatal(server.ListenAndServe())
}
