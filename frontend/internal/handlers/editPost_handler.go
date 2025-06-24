package handlers

import (
	"net/http"
	"strings"

	"frontend-service/internal/services"
	"frontend-service/internal/session"

	"frontend-service/internal/validations"
)

type EditPostHandler struct {
	authService     *services.AuthService
	postService     *services.PostService
	categoryService *services.CategoryService
	templateService *services.TemplateService
}

// NewEditPostHandler creates a new edit post handler
func NewEditPostHandler(authService *services.AuthService, postService *services.PostService, categoryService *services.CategoryService, templateService *services.TemplateService) *EditPostHandler {
	return &EditPostHandler{
		authService:     authService,
		postService:     postService,
		categoryService: categoryService,
		templateService: templateService,
	}
}

// ServeEditPost handles both GET (show form) and POST (submit form) for editing posts
func (h *EditPostHandler) ServeEditPost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showEditPostForm(w, r)
	case http.MethodPost:
		h.handleEditPostForm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// showEditPostForm displays the edit post form (GET request)
func (h *EditPostHandler) showEditPostForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract post ID from URL path
	postID := r.PathValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Get the post to edit
	sessionCookie, _ := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	post, _, err := h.postService.GetSinglePostWithComments(postID, 1, 0, "oldest", sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load post", http.StatusInternalServerError)
		return
	}

	// Check if user owns the post
	if post.UserID != user.ID {
		http.Error(w, "You can only edit your own posts", http.StatusForbidden)
		return
	}

	// Get all categories for the form
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// Extract category names from the post
	var postCategoryNames []string
	for _, cat := range post.Categories {
		postCategoryNames = append(postCategoryNames, cat.Name)
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":       user,
		"Post":       post,
		"Categories": categories,
		"FormData": map[string]interface{}{
			"content":    post.Content,
			"categories": postCategoryNames,
		},
		"IsEdit": true, // Flag to indicate this is an edit form
	}

	// Render the template (we'll reuse create-post.html with edit mode)
	if err := h.templateService.Render(w, "edit-post.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// handleEditPostForm processes the edit post form submission (POST request)
func (h *EditPostHandler) handleEditPostForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract post ID from URL path
	postID := r.PathValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.showEditPostError(w, r, postID, "Invalid form data", nil)
		return
	}

	// Get form values
	content := strings.TrimSpace(r.FormValue("content"))
	selectedCategories := r.Form["categories"] // This will be a slice of category names

	// Basic validation
	if content == "" {
		h.showEditPostError(w, r, postID, "Post content is required", map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	if len(selectedCategories) == 0 {
		h.showEditPostError(w, r, postID, "At least one category must be selected", map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	// Validate post content using existing validation
	if err := validations.ValidatePostContent(content); err != nil {
		h.showEditPostError(w, r, postID, err.Error(), map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Call backend API to update post
	err = h.postService.UpdatePost(postID, selectedCategories, content, sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, "You can only edit your own posts", http.StatusForbidden)
			return
		}
		h.showEditPostError(w, r, postID, err.Error(), map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	// Redirect to the updated post
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

// showEditPostError displays the edit post form with error message and preserved form data
func (h *EditPostHandler) showEditPostError(w http.ResponseWriter, r *http.Request, postID, errorMsg string, formData map[string]interface{}) {
	// Get user (should be logged in if we reach this point)
	user := session.GetUserFromSession(r, h.authService)

	// Get the original post for reference
	sessionCookie, _ := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	post, _, err := h.postService.GetSinglePostWithComments(postID, 1, 0, "oldest", sessionCookie)
	if err != nil {
		http.Error(w, "Failed to load post", http.StatusInternalServerError)
		return
	}

	// Get all categories for the form
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":       user,
		"Post":       post,
		"Categories": categories,
		"Error":      errorMsg,
		"FormData":   formData,
		"IsEdit":     true,
	}

	// Render the template with error
	if err := h.templateService.Render(w, "edit-post.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
