package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
)

type PostReactionHandler struct {
	authService            *services.AuthService
	postReactionService    *services.PostReactionService
	commentReactionService *services.CommentReactionService
}

// NewPostReactionHandler creates a new post reaction handler (now handles both post and comment reactions)
func NewPostReactionHandler(authService *services.AuthService, postReactionService *services.PostReactionService, commentReactionService *services.CommentReactionService) *PostReactionHandler {
	return &PostReactionHandler{
		authService:            authService,
		postReactionService:    postReactionService,
		commentReactionService: commentReactionService,
	}
}

// ServeTogglePostReaction handles post reaction toggle (form submission from any page)
func (h *PostReactionHandler) ServeTogglePostReaction(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in - REQUIRED for reactions
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.redirectWithError(w, r, "Invalid form data")
		return
	}

	// Get form values
	postID := strings.TrimSpace(r.FormValue("post_id"))
	reactionTypeStr := strings.TrimSpace(r.FormValue("reaction_type"))

	// Validate post ID
	if postID == "" {
		h.redirectWithError(w, r, "Post ID is required")
		return
	}

	// Parse and validate reaction type
	reactionType, err := strconv.Atoi(reactionTypeStr)
	if err != nil || (reactionType != models.ReactionTypeLike && reactionType != models.ReactionTypeDislike) {
		h.redirectWithError(w, r, "Invalid reaction type")
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Determine reaction type for logging
	reactionName := "like"
	if reactionType == models.ReactionTypeDislike {
		reactionName = "dislike"
	}

	log.Printf("Toggling post reaction: user=%s, post=%s, reaction=%s", user.Username, postID, reactionName)

	// Toggle reaction via API
	result, err := h.postReactionService.TogglePostReaction(postID, reactionType, sessionCookie)
	if err != nil {
		// Handle authentication errors
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// For other errors, redirect back with error message
		h.redirectWithError(w, r, "Failed to update reaction")
		return
	}

	log.Printf("Post reaction toggled successfully: %s - %s", result.Action, result.Message)

	// Always redirect back to the referring page to show updated reaction state
	h.redirectBack(w, r, postID)
}

// ServeToggleCommentReaction handles comment reaction toggle (form submission from post page)
func (h *PostReactionHandler) ServeToggleCommentReaction(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in - REQUIRED for reactions
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.redirectWithError(w, r, "Invalid form data")
		return
	}

	// Get form values
	commentID := strings.TrimSpace(r.FormValue("comment_id"))
	reactionTypeStr := strings.TrimSpace(r.FormValue("reaction_type"))

	// Validate comment ID
	if commentID == "" {
		h.redirectWithError(w, r, "Comment ID is required")
		return
	}

	// Parse and validate reaction type
	reactionType, err := strconv.Atoi(reactionTypeStr)
	if err != nil || (reactionType != models.ReactionTypeLike && reactionType != models.ReactionTypeDislike) {
		h.redirectWithError(w, r, "Invalid reaction type")
		return
	}

	// Get session cookie for API call
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Determine reaction type for logging
	reactionName := "like"
	if reactionType == models.ReactionTypeDislike {
		reactionName = "dislike"
	}

	log.Printf("Toggling comment reaction: user=%s, comment=%s, reaction=%s", user.Username, commentID, reactionName)

	// Toggle reaction via API
	result, err := h.commentReactionService.ToggleCommentReaction(commentID, reactionType, sessionCookie)
	if err != nil {
		// Handle authentication errors
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// For other errors, redirect back with error message
		h.redirectWithError(w, r, "Failed to update comment reaction")
		return
	}

	log.Printf("Comment reaction toggled successfully: %s - %s", result.Action, result.Message)

	// Always redirect back to the referring page (should be the post page)
	h.redirectBackForComment(w, r)
}

// redirectBack redirects the user back to where they came from after post reaction toggle
func (h *PostReactionHandler) redirectBack(w http.ResponseWriter, r *http.Request, postID string) {
	// Try to get the referring URL to redirect back
	referer := r.Header.Get("Referer")

	// Check if there's a custom redirect_to parameter in the form
	if redirectTo := r.FormValue("redirect_to"); redirectTo != "" {
		// Validate that it's a safe internal redirect
		if strings.HasPrefix(redirectTo, "/") && !strings.HasPrefix(redirectTo, "//") {
			http.Redirect(w, r, redirectTo, http.StatusSeeOther)
			return
		}
	}

	// If we have a referer and it's valid, redirect back to it
	if referer != "" {
		// Make sure it's a safe internal URL
		if strings.Contains(referer, "/post/") || strings.Contains(referer, "/category/") || strings.Contains(referer, "/profile") || referer == "/" {
			http.Redirect(w, r, referer, http.StatusSeeOther)
			return
		}
	}

	// Default fallback: redirect to the post page
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

// redirectBackForComment redirects back after comment reaction toggle
func (h *PostReactionHandler) redirectBackForComment(w http.ResponseWriter, r *http.Request) {
	// For comment reactions, always try to go back to the referring page (post page)
	referer := r.Header.Get("Referer")
	if referer != "" && strings.Contains(referer, "/post/") {
		http.Redirect(w, r, referer, http.StatusSeeOther)
		return
	}

	// Fallback: redirect to home if we can't determine the post
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// redirectWithError redirects back with an error message
func (h *PostReactionHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	referer := r.Header.Get("Referer")
	if referer != "" {
		// Add error parameter to the referer URL
		separator := "?"
		if strings.Contains(referer, "?") {
			separator = "&"
		}
		http.Redirect(w, r, referer+separator+"error="+errorMsg, http.StatusSeeOther)
		return
	}

	// Fallback: redirect to home with error
	http.Redirect(w, r, "/?error="+errorMsg, http.StatusSeeOther)
}
