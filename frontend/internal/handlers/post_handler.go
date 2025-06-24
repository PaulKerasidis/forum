package handlers

import (
	"net/http"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/utils"
)

type PostHandler struct {
	authService     *services.AuthService
	postService     *services.PostService
	templateService *services.TemplateService
}

// NewPostHandler creates a new post handler
func NewPostHandler(authService *services.AuthService, postService *services.PostService, templateService *services.TemplateService) *PostHandler {
	return &PostHandler{
		authService:     authService,
		postService:     postService,
		templateService: templateService,
	}
}

// ServePostView handles the single post view page request
func (h *PostHandler) ServePostView(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	postID := r.PathValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters for comments
	paginationParams := utils.ParsePaginationFromRequest(r)

	// Parse sort parameter for comments (default to newest for better UX)
	sortBy := utils.ParseSortFromRequest(r, "newest")

	// Check if user is logged in and get session cookie for reaction data
	user := session.GetUserFromSession(r, h.authService)
	var sessionCookie *http.Cookie
	if user != nil {
		sessionCookie, _ = session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	}

	// Get post and comments from backend API (with session for reaction data)
	post, comments, err := h.postService.GetSinglePostWithComments(postID, paginationParams.Limit, paginationParams.Offset, sortBy, sessionCookie)
	if err != nil {
		if err.Error() == "post not found" {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load post", http.StatusInternalServerError)
		return
	}

	// Convert []*Comment to []Comment to match PostPageData struct
	var commentsSlice []models.Comment
	for _, comment := range comments {
		if comment != nil {
			commentsSlice = append(commentsSlice, *comment)
		}
	}

	// Prepare data for template using existing PostPageData model
	data := models.PostPageData{
		Post:     post,
		Comments: commentsSlice,
		User:     user, // Pass user for reaction buttons
	}

	// Render the template
	if err := h.templateService.Render(w, "post.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
