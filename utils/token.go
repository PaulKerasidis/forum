package utils

import (
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateSessionToken creates a new session token/ID
func GenerateSessionToken() string {
	return GenerateUUID()
}

// CalculateSessionExpiry calculates the expiry time for a session
// Default session lifetime is 24 hours
func CalculateSessionExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}