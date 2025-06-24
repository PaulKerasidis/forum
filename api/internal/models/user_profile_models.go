package models

import "time"

// UserProfile - Complete user profile with activity statistics
type UserProfile struct {
	ID        string       `json:"user_id"`
	Username  string       `json:"username"`
	Email     string       `json:"email"`
	CreatedAt time.Time    `json:"created_at"`
	Stats     ProfileStats `json:"stats"`
}

type ProfileStats struct {
	TotalPosts       int `json:"total_posts"`
	TotalComments    int `json:"total_comments"`
	PostsLiked       int `json:"posts_liked"`        // Posts this user has liked
	PostsCommentedOn int `json:"posts_commented_on"` // Posts this user has commented on
	LikesReceived    int `json:"likes_received"`     // Total likes on user's posts/comments
	DislikesReceived int `json:"dislikes_received"`  // Total dislikes on user's posts/comments
}
