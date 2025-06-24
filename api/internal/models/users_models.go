package models

import "time"

// User represents public user information
type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Provider      string    `json:"provider,omitempty"`      // oauth provider (google, github, etc.)
	ProviderID    string    `json:"provider_id,omitempty"`   // user id from oauth provider
	ProviderEmail string    `json:"provider_email,omitempty"` // email from oauth provider
	CreatedAt     time.Time `json:"created_at"`
}

// UserAuth represents internal user data including authentication
type UserPassword struct {
	UserID       string `json:"user_id"`
	PasswordHash string `json:"-"` // Never expose in JSON
}

// UserRegistration - Registration form data (updated with password confirmation)
type UserRegistration struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"` // NEW: Password confirmation field
}

// UserLogin is used for login requests
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
