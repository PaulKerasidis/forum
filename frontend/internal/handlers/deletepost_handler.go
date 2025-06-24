package handlers

import (
	"net/http"
	"strings"

	"frontend-service/internal/services"
	"frontend-service/internal/session"
)

type DeletePostHandler struct {
	authService *services.AuthService
	postService *services.PostService
}

// NewDeletePostHandler creates a new delete post handler
func NewDeletePostHandler(authService *services.AuthService, postService *services.PostService) *DeletePostHandler {
	return &DeletePostHandler{
		authService: authService,
		postService: postService,
	}
}

// ServeDeletePost handles post deletion (POST request only for security)
func (h *DeletePostHandler) ServeDeletePost(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method for security (prevent accidental deletion via GET)
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
	postID := r.PathValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Get the post to verify ownership
	sessionCookie, _ := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	post, _, err := h.postService.GetSinglePostWithComments(postID, 1, 0, "oldest", sessionCookie)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify post ownership", http.StatusInternalServerError)
		return
	}

	// Check if user owns the post
	if post.UserID != user.ID {
		http.Error(w, "You can only delete your own posts", http.StatusForbidden)
		return
	}

	// Call backend API to delete post
	err = h.postService.DeletePost(postID, sessionCookie)
	if err != nil {
		// Handle different error types
		if strings.Contains(err.Error(), "unauthorized") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, "You can only delete your own posts", http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		// For other errors, redirect back to post with error (we could implement flash messages later)
		http.Redirect(w, r, "/post/"+postID+"?error=delete_failed", http.StatusSeeOther)
		return
	}

	// Determine where to redirect after successful deletion
	redirectURL := h.getRedirectAfterDeletion(r)

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// getRedirectAfterDeletion determines where to redirect user after successful deletion
func (h *DeletePostHandler) getRedirectAfterDeletion(r *http.Request) string {
	// Check if there's a specific redirect URL in the form or query parameter
	if redirectTo := r.FormValue("redirect_to"); redirectTo != "" {
		// Validate that it's a safe internal redirect
		if strings.HasPrefix(redirectTo, "/") && !strings.HasPrefix(redirectTo, "//") {
			return redirectTo
		}
	}

	// Check if there's a referer header to go back to
	if referer := r.Header.Get("Referer"); referer != "" {
		// If coming from the post page itself, redirect to home
		if strings.Contains(referer, "/post/") {
			return "/"
		}
		// If coming from a category page, try to extract and redirect there
		if strings.Contains(referer, "/category/") {
			// Extract category ID from referer if possible
			parts := strings.Split(referer, "/category/")
			if len(parts) > 1 {
				categoryID := strings.Split(parts[1], "?")[0]  // Remove query parameters
				categoryID = strings.Split(categoryID, "#")[0] // Remove hash
				if categoryID != "" {
					return "/category/" + categoryID
				}
			}
		}
		// If coming from home page or other safe page, redirect there
		if strings.Contains(referer, "/") && !strings.Contains(referer, "/post/") {
			// Validate it's a safe internal URL
			if strings.HasPrefix(referer, "http://localhost") || strings.HasPrefix(referer, "https://") {
				// Extract just the path part
				if strings.Contains(referer, "://") {
					parts := strings.SplitN(referer, "://", 2)
					if len(parts) > 1 {
						pathParts := strings.SplitN(parts[1], "/", 2)
						if len(pathParts) > 1 {
							return "/" + pathParts[1]
						}
					}
				}
			}
		}
	}

	// Default fallback: redirect to home page
	return "/"
}
