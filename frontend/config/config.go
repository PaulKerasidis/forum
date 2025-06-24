package config

import "os"

// Config holds all configuration for the frontend service
type Config struct {
	// Server configuration
	Port string

	// Backend API configuration
	APIBaseURL string

	// Template and static files
	TemplatesDir string
	StaticDir    string

	// Session configuration (only what's needed)
	SessionName string // For cookie name consistency with backend
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("FRONTEND_PORT", "3000"),
		APIBaseURL:   getEnv("API_BASE_URL", "http://localhost:8080/api"),
		TemplatesDir: getEnv("TEMPLATES_DIR", "./web/templates"),
		StaticDir:    getEnv("STATIC_DIR", "./web/static"),
		SessionName:  getEnv("SESSION_NAME", "forum_session"),
	}
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
