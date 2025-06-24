package models

import (
	"time"
)

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

	// User context
	UserReaction *int `json:"user_reaction,omitempty"` // nil, 1=like, 2=dislike
	IsOwner      bool `json:"is_owner,omitempty"`      // can current user edit/delete
}

// CategoryPostsResponse represents the response from the backend for category posts
type CategoryPostsResponse struct {
	Category   Category               `json:"category"`
	Posts      []*Post                `json:"posts"`
	Pagination PaginationInfo         `json:"pagination"`
	Sort       map[string]interface{} `json:"sort"`
}
