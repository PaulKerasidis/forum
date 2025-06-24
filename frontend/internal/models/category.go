package models

// Category - Forum category with post count (matches backend exactly)
type Category struct {
	ID    string `json:"category_id"`
	Name  string `json:"category_name"`
	Count int    `json:"post_count"` // Number of posts in the category
}

// PostCategory - Category info for posts (matches backend exactly)
type PostCategory struct {
	ID   string `json:"category_id"`
	Name string `json:"category_name"`
}
