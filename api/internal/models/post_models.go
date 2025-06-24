package models

import "time"

type Post struct {
	ID         string         `json:"post_id"`
	UserID     string         `json:"user_id"`
	Username   string         `json:"username"`
	Categories []PostCategory `json:"categories"`
	Content    string         `json:"post_content"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  *time.Time     `json:"updated_at,omitempty"`

	// Aggregated metrics
	LikeCount    int `json:"like_count"`
	DislikeCount int `json:"dislike_count"`
	CommentCount int `json:"comment_count"`

	// NEW FIELDS for two-call approach:
	UserReaction *int `json:"user_reaction,omitempty"` // nil, 1=like, 2=dislike
	IsOwner      bool `json:"is_owner,omitempty"`      // can current user edit/delete
}

// CreatePostRequest - Post creation payload
type CreatePostRequest struct {
	CategoryNames []string `json:"category_names" binding:"required,min=1,max=5"`
	Content       string   `json:"content" binding:"required,min=10,max=5000"`
}

// UpdatePostRequest - Post update payload
type UpdatePostRequest struct {
	CategoryNames []string `json:"category_names" binding:"required,min=1,max=5"`
	Content       string   `json:"content" binding:"required,min=10,max=5000"`
}
