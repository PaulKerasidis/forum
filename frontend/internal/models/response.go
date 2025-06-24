package models

import "time"

// APIResponse - Standard API response wrapper (matches backend exactly)
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginResponse - Response after successful login (matches backend exactly)
type LoginResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
}

// CreatePostResponse - Response after successful post creation (matches backend exactly)
type CreatePostResponse struct {
	PostID    string    `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
