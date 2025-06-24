package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AppConfig holds all API backend configuration
type AppConfig struct {
	// API Server configuration
	ServerHost string
	ServerPort string

	// Database configuration
	DBPath           string
	DBMaxConnections int

	// Security configuration (Session-based)
	SessionName string // ADDED: Session cookie name
	Environment string

	// Authentication configuration
	SessionDuration time.Duration
	BCryptCost      int
	MaxPasswordLen  int
	MinPasswordLen  int
	MaxUsernameLen  int
	MinUsernameLen  int

	// Content configuration
	MaxPostContentLength int
	MinPostContentLength int
	MaxCommentLength     int
	MinCommentLength     int

	// Rate limiting configuration
	RateLimitRequests int
	RateLimitWindow   int // in minutes

	// Pagination configuration
	DefaultPageSize int
	MaxPageSize     int
	MinPageSize     int

	// Database query configuration
	MaxCategories int
	MinCategories int

	// CORS configuration (for frontend communication)
	AllowedOrigins string // comma-separated origins
	AllowedMethods string
	AllowedHeaders string

	// OAuth configuration
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	GoogleOAuthRedirectURL  string

	// File/Image configuration (for future use)

}

// Global configuration instance
var Config AppConfig

// LoadConfig initializes and validates configuration with values from environment variables or defaults
func LoadConfig() error {
	err := LoadEnv(".env")
	if err != nil {
		return err
	}

	// API Server configuration
	Config.ServerHost = getEnv("SERVER_HOST", "localhost")
	Config.ServerPort = getEnv("SERVER_PORT", "8080")

	// Database configuration
	Config.DBPath = getEnv("DB_PATH", "./DBPath/forum.db")
	Config.DBMaxConnections = getEnvAsInt("DB_MAX_CONNECTIONS", 10)

	// Security configuration (Session-based)

	Config.SessionName = getEnv("SESSION_NAME", "forum_session") // ADDED: Load session name from env
	Config.Environment = getEnv("ENVIRONMENT", "development")

	// Authentication configuration
	Config.SessionDuration = getEnvAsDuration("SESSION_DURATION", 24*time.Hour)
	Config.BCryptCost = getEnvAsInt("BCRYPT_COST", 10) // 10 for dev, 12+ for production
	Config.MaxUsernameLen = getEnvAsInt("MAX_USERNAME_LENGTH", 15)
	Config.MinUsernameLen = getEnvAsInt("MIN_USERNAME_LENGTH", 5)
	Config.MaxPasswordLen = getEnvAsInt("MAX_PASSWORD_LENGTH", 15)
	Config.MinPasswordLen = getEnvAsInt("MIN_PASSWORD_LENGTH", 3)

	// Content configuration - Posts
	Config.MaxPostContentLength = getEnvAsInt("MAX_POST_CONTENT_LENGTH", 500)
	Config.MinPostContentLength = getEnvAsInt("MIN_POST_CONTENT_LENGTH", 10)

	// Content configuration - Comments
	Config.MaxCommentLength = getEnvAsInt("MAX_COMMENT_LENGTH", 150)
	Config.MinCommentLength = getEnvAsInt("MIN_COMMENT_LENGTH", 5)

	// Rate limiting configuration
	Config.RateLimitRequests = getEnvAsInt("RATE_LIMIT_REQUESTS", 100000)
	Config.RateLimitWindow = getEnvAsInt("RATE_LIMIT_WINDOW", 60) // minutes

	// Pagination configuration
	Config.DefaultPageSize = getEnvAsInt("DEFAULT_PAGE_SIZE", 20)
	Config.MaxPageSize = getEnvAsInt("MAX_PAGE_SIZE", 50)
	Config.MinPageSize = getEnvAsInt("MIN_PAGE_SIZE", 1)

	// Database query configuration
	Config.MaxCategories = getEnvAsInt("MAX_CATEGORIES_PER_POST", 5)
	Config.MinCategories = getEnvAsInt("MIN_CATEGORIES_PER_POST", 1)

	// CORS configuration (for frontend communication)
	Config.AllowedOrigins = getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	Config.AllowedMethods = getEnv("ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	Config.AllowedHeaders = getEnv("ALLOWED_HEADERS", "Content-Type,Authorization,X-Requested-With")

	// OAuth configuration
	Config.GoogleOAuthClientID = getEnv("GOOGLE_OAUTH_CLIENT_ID", "79779164831-4obtdfsgqnjakbd1n8lj1ma28lg1i7dl.apps.googleusercontent.com")
	Config.GoogleOAuthClientSecret = getEnv("GOOGLE_OAUTH_CLIENT_SECRET", "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")
	Config.GoogleOAuthRedirectURL = getEnv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback")

	return nil

}

// LoadEnv function manually loads .env file into environment variables
func LoadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// Don't fail if .env doesn't exist (production scenario)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines and lines starting with #
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the line into key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format on line %d: %s", lineNumber, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Only set if not already set (environment variables take precedence)
		if os.Getenv(key) == "" {
			err := os.Setenv(key, value)
			if err != nil {
				return fmt.Errorf("failed to set environment variable %s: %v", key, err)
			}
		}
	}

	return scanner.Err()
}

// Helper functions to get environment variables with fallbacks
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
