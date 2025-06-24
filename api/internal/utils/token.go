package utils

import (
	"encoding/base64"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/PaulKerasidis/forum/config"
)

// GenerateUUID generates a new UUID string
func GenerateUUIDToken() string {
	return uuid.New().String()
}

// GenerateSessionToken creates a new session token/ID
func GenerateSessionToken() (string, error) {
	// 32 bytes gives you 256 bits of entropy
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b), nil
}

// CalculateSessionExpiry calculates the expiry time for a session
// Default session lifetime is 24 hours
func CalculateSessionExpiry() time.Time {
	return time.Now().Add(config.Config.SessionDuration) // fix later for not having magic numbers
}
