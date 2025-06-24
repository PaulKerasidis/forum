package models

// UserRegistration - Registration form data (matches backend exactly)
type UserRegistration struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"` // Added confirm password field
}

// UserLogin - Login form data (matches backend exactly)
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreatePostRequest - Post creation form data (matches backend exactly)
type CreatePostRequest struct {
	CategoryNames []string `json:"category_names"`
	Content       string   `json:"content"`
}

// UpdatePostRequest - Post update form data (matches backend exactly)
type UpdatePostRequest struct {
	CategoryNames []string `json:"category_names"`
	Content       string   `json:"content"`
}

// CreateCommentRequest - Comment creation form data (matches backend exactly)
type CreateCommentRequest struct {
	PostID  string `json:"post_id"`
	Content string `json:"content"`
}

// UpdateCommentRequest - Comment update form data (matches backend exactly)
type UpdateCommentRequest struct {
	Content string `json:"content"`
}
