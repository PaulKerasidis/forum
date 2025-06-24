package models

import "time"

// Standard api response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginResponse is the data response after successful login
type LoginResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
}

// CreatePostResponse - Lightweight response for post creation
type CreatePostResponse struct {
	PostID    string    `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentResponse - Lightweight response for comment creation
type CreateCommentResponse struct {
	CommentID string    `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}
