package models

import "time"

// Comment - Forum comment with all metadata (matches backend exactly)
type Comment struct {
	ID        string     `json:"comment_id"`
	UserID    string     `json:"user_id"`
	Username  string     `json:"username"`
	PostID    string     `json:"post_id"`
	Content   string     `json:"comment_content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// Aggregated metrics
	LikeCount    int `json:"like_count"`
	DislikeCount int `json:"dislike_count"`

	// User context
	UserReaction *int `json:"user_reaction,omitempty"` // nil, 1=like, 2=dislike
	IsOwner      bool `json:"is_owner,omitempty"`      // can current user edit/delete
}
