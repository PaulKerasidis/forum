package handlers

import (
	"net/http"
	"strings"

	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/validations"
)

type CreatePostHandler struct {
	authService     *services.AuthService
	postService     *services.PostService
	categoryService *services.CategoryService
	templateService *services.TemplateService
}

// NewCreatePostHandler creates a new create post handler
func NewCreatePostHandler(authService *services.AuthService, postService *services.PostService, categoryService *services.CategoryService, templateService *services.TemplateService) *CreatePostHandler {
	return &CreatePostHandler{
		authService:     authService,
		postService:     postService,
		categoryService: categoryService,
		templateService: templateService,
	}
}

// ServeCreatePost handles both GET (show form) and POST (submit form) for creating posts
func (h *CreatePostHandler) ServeCreatePost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showCreatePostForm(w, r)
	case http.MethodPost:
		h.handleCreatePostForm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// showCreatePostForm displays the create post form (GET request)
func (h *CreatePostHandler) showCreatePostForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
		"Categories": categories,
		"FormData":   map[string]interface{}{}, // Empty form data for initial load
	}

	// Render the template
	if err := h.templateService.Render(w, "create-post.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// handleCreatePostForm processes the create post form submission (POST request)
func (h *CreatePostHandler) handleCreatePostForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.showCreatePostError(w, r, "Invalid form data", nil)
		return
	}

	// Get form values
	content := strings.TrimSpace(r.FormValue("content"))
	selectedCategories := r.Form["categories"] // This will be a slice of category names

	// Basic validation
	if content == "" {
		h.showCreatePostError(w, r, "Post content is required", map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	if len(selectedCategories) == 0 {
		h.showCreatePostError(w, r, "At least one category must be selected", map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	// Validate post content using existing validation
	if err := validations.ValidatePostContent(content); err != nil {
		h.showCreatePostError(w, r, err.Error(), map[string]interface{}{
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

	// Call backend API to create post
	createResponse, err := h.postService.CreatePost(selectedCategories, content, sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		h.showCreatePostError(w, r, err.Error(), map[string]interface{}{
			"content":    content,
			"categories": selectedCategories,
		})
		return
	}

	// Redirect to the newly created post
	http.Redirect(w, r, "/post/"+createResponse.PostID, http.StatusSeeOther)
}

// showCreatePostError displays the create post form with error message and preserved form data
func (h *CreatePostHandler) showCreatePostError(w http.ResponseWriter, r *http.Request, errorMsg string, formData map[string]interface{}) {
	// Get user (should be logged in if we reach this point)
	user := session.GetUserFromSession(r, h.authService)

	// Get all categories for the form
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":       user,
		"Categories": categories,
		"Error":      errorMsg,
		"FormData":   formData,
	}

	// Render the template with error
	if err := h.templateService.Render(w, "create-post.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
