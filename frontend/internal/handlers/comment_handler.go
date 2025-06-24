package handlers

import (
	"log"
	"net/http"
	"strings"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/validations"
)

type CommentHandler struct {
	authService     *services.AuthService
	commentService  *services.CommentService
	postService     *services.PostService
	templateService *services.TemplateService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(authService *services.AuthService, commentService *services.CommentService, postService *services.PostService, templateService *services.TemplateService) *CommentHandler {
	return &CommentHandler{
		authService:     authService,
		commentService:  commentService,
		postService:     postService,
		templateService: templateService,
	}
}

// ServeCreateComment handles comment creation (form submission)
func (h *CommentHandler) ServeCreateComment(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract post ID from URL path
	postID := r.PathValue("post_id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/post/"+postID+"?error=invalid_form", http.StatusSeeOther)
		return
	}

	// Get and validate comment content
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Redirect(w, r, "/post/"+postID+"?error=empty_content", http.StatusSeeOther)
		return
	}

	if err := validations.ValidateCommentContent(content); err != nil {
		http.Redirect(w, r, "/post/"+postID+"?error=validation_failed", http.StatusSeeOther)
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Create comment via API
	_, err = h.commentService.CreateComment(postID, content, sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/post/"+postID+"?error=create_failed", http.StatusSeeOther)
		return
	}

	// Redirect back to the post (comment will appear after page refresh)
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

// ServeEditComment handles both GET (show edit form) and POST (save changes)
func (h *CommentHandler) ServeEditComment(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract comment ID from URL path
	commentID := r.PathValue("comment_id")
	if commentID == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.showEditCommentForm(w, commentID)
	case http.MethodPost:
		h.handleEditCommentForm(w, r, commentID, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ServeDeleteComment handles comment deletion (form submission)
func (h *CommentHandler) ServeDeleteComment(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract comment ID from URL path
	commentID := r.PathValue("comment_id")
	if commentID == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Delete comment via API
	err = h.commentService.DeleteComment(commentID, sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// We don't know the post ID here, so redirect to home with error
		http.Redirect(w, r, "/?error=delete_failed", http.StatusSeeOther)
		return
	}

	// Try to get the referring post URL to redirect back
	referer := r.Header.Get("Referer")
	if referer != "" && strings.Contains(referer, "/post/") {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}

	// Fallback: redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// showEditCommentForm displays the edit comment form (GET request)
func (h *CommentHandler) showEditCommentForm(w http.ResponseWriter, commentID string) {
	// For now, we'll create a simple edit comment template
	// This is a placeholder - we'll create the template in the next step

	// TODO: We need to fetch the comment data and show an edit form
	// For now, just redirect back - we'll implement this in the template step
	http.Error(w, "Edit comment form not implemented yet", http.StatusNotImplemented)
}

// handleEditCommentForm processes the edit comment form submission (POST request)
func (h *CommentHandler) handleEditCommentForm(w http.ResponseWriter, r *http.Request, commentID string, user *models.User) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get and validate comment content
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	if err := validations.ValidateCommentContent(content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Update comment via API
	err = h.commentService.UpdateComment(commentID, content, sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, "You can only edit your own comments", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	// Try to redirect back to the referring post
	referer := r.Header.Get("Referer")
	if referer != "" && strings.Contains(referer, "/post/") {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}

	// Fallback: redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ServeEditCommentForm shows the edit comment form (GET)
func (h *CommentHandler) ServeEditCommentForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get comment ID from URL
	commentID := r.PathValue("id")
	if commentID == "" {
		http.Error(w, "Comment ID required", http.StatusBadRequest)
		return
	}

	// Get session cookie
	sessionCookie, _ := session.GetSessionCookie(r, h.authService)

	// Get the comment
	comment, err := h.commentService.GetCommentByID(commentID, sessionCookie)
	if err != nil {
		log.Printf("Error getting comment: %v", err)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if user owns the comment
	if comment.UserID != user.ID {
		http.Error(w, "Forbidden: You can only edit your own comments", http.StatusForbidden)
		return
	}

	// Prepare template data
	data := struct {
		Comment *models.Comment
		User    *models.User
		Error   string
	}{
		Comment: comment,
		User:    user,
	}

	// Render edit comment template
	if err := h.templateService.Render(w, "edit-comment.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// ServeEditCommentSubmit processes the edit comment form (POST)
func (h *CommentHandler) ServeEditCommentSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get comment ID from URL
	commentID := r.PathValue("id")
	if commentID == "" {
		http.Error(w, "Comment ID required", http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get form data
	content := strings.TrimSpace(r.FormValue("content"))
	redirectTo := r.FormValue("redirect_to")

	// Validate content
	if err := validations.ValidateCommentContent(content); err != nil {
		// Show form again with error
		h.showEditCommentError(w, r, commentID, content, err.Error())
		return
	}

	// Get session cookie
	sessionCookie, _ := session.GetSessionCookie(r, h.authService)

	// Update the comment
	if err := h.commentService.UpdateComment(commentID, content, sessionCookie); err != nil {
		log.Printf("Error updating comment: %v", err)
		h.showEditCommentError(w, r, commentID, content, err.Error())
		return
	}

	// Redirect back to post or default location
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Helper method to show edit form with error
func (h *CommentHandler) showEditCommentError(w http.ResponseWriter, r *http.Request, commentID, content, errorMsg string) {
	user := session.GetUserFromSession(r, h.authService)

	// Create a comment object with the form data
	comment := &models.Comment{
		ID:      commentID,
		Content: content,
	}

	data := struct {
		Comment *models.Comment
		User    *models.User
		Error   string
	}{
		Comment: comment,
		User:    user,
		Error:   errorMsg,
	}

	if err := h.templateService.Render(w, "edit-comment.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
