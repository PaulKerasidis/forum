package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserAuth contains authentication information for a user
type UserAuth struct {
	UserID       string `json:"user_id"`
	PasswordHash string `json:"-"` // Never expose password hash in JSON
}

// Session represents a user login session
type Session struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Post represents a forum post
type Post struct {
	PostID     string     `json:"post_id"`
	UserID     string     `json:"user_id"`
	CategoryID int        `json:"category_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`

	// Not stored in database directly - populated by joins
	Username string `json:"username,omitempty"`
}

// Comment represents a user comment on a post
type Comment struct {
	CommentID string     `json:"comment_id"`
	PostID    string     `json:"post_id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Not stored in database directly - populated by joins
	Username string `json:"username,omitempty"`
}

// Category represents a post category
type Category struct {
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
}

// ReactionType defines the type of reaction (like or dislike)
type ReactionType int

const (
	Like    ReactionType = 1
	Dislike ReactionType = 2
)

// Reaction represents a user's reaction to a post or comment
type Reaction struct {
	UserID       string       `json:"user_id"`
	ReactionType ReactionType `json:"reaction_type"`
	CommentID    *string      `json:"comment_id,omitempty"`
	PostID       *string      `json:"post_id,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}
