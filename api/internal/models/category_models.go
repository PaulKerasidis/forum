package models

// Category represents a forum category with post count for UI sidebar
type Category struct {
	ID    string `json:"category_id"`
	Name  string `json:"category_name"`
	Count int    `json:"post_count"` // Number of posts in the category
}

// The category data used in the post list
type PostCategory struct {
	ID   string `json:"category_id"`
	Name string `json:"category_name"`
}
