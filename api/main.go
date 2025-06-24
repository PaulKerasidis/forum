package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/database"
	"github.com/PaulKerasidis/forum/internal/routes"
)

func main() {
	// Load configuration
	err := config.LoadConfig()
	if err != nil {
		log.Panic(err)
	}

	// Debug: Check if OAuth config is loaded
	fmt.Printf("OAuth Client ID loaded: %s\n", config.Config.GoogleOAuthClientID)
	if config.Config.GoogleOAuthClientID == "" {
		log.Println("WARNING: Google OAuth Client ID is empty!")
	}

	// Initialize the database
	db, err := database.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Setup API routes
	apiRoutes := routes.SetupRoutes(db)

	// Create server using config values
	serverAddr := fmt.Sprintf("%s:%s", config.Config.ServerHost, config.Config.ServerPort)

	fmt.Printf("OAuth routes registered:\n")
	fmt.Printf("- /auth/google/login\n")
	fmt.Printf("- /auth/google/callback\n")
	fmt.Printf("- /api/auth/google/login\n")
	fmt.Printf("- /api/auth/google/callback\n")
	fmt.Printf("Starting API server on %s\n", serverAddr)

	log.Fatal(http.ListenAndServe(serverAddr, apiRoutes))
}
