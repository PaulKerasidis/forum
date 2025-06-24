package handlers

import (
	"net/http"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/utils"
)

type ProfileHandler struct {
	authService     *services.AuthService
	userService     *services.UserService
	templateService *services.TemplateService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(authService *services.AuthService, userService *services.UserService, templateService *services.TemplateService) *ProfileHandler {
	return &ProfileHandler{
		authService:     authService,
		userService:     userService,
		templateService: templateService,
	}
}

// ServeProfile handles the main profile page request (shows stats and navigation)
func (h *ProfileHandler) ServeProfile(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get session cookie for API calls
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user profile with stats from backend API
	userProfile, err := h.userService.GetUserProfile(user.ID, sessionCookie)
	if err != nil {
		if err.Error() == "unauthorized: please log in" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err.Error() == "forbidden: you can only view your own profile" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := models.ProfilePageData{
		Profile: userProfile,
		User:    user,
	}

	// Render the profile template
	if err := h.templateService.Render(w, "profile.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// ServeUserPosts handles the "My Posts" page - shows posts created by the user
func (h *ProfileHandler) ServeUserPosts(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse pagination parameters
	paginationParams := utils.ParsePaginationFromRequest(r)
	sortBy := utils.ParseSortFromRequest(r, "newest")

	// Get session cookie for API calls
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's posts from backend API
	postsResponse, err := h.userService.GetUserPosts(user.ID, paginationParams.Limit, paginationParams.Offset, sortBy, sessionCookie)
	if err != nil {
		if err.Error() == "unauthorized: please log in" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to load your posts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":         user,
		"Posts":        postsResponse.Posts,
		"Pagination":   postsResponse.Pagination,
		"PageTitle":    "My Posts",
		"PageType":     "user-posts",
		"EmptyMessage": "You haven't created any posts yet.",
		"CreateLink":   "/create-post",
	}

	// Render the user posts template (we'll create this)
	if err := h.templateService.Render(w, "user-posts.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// ServeUserLikedPosts handles the "Liked Posts" page - shows posts liked by the user
func (h *ProfileHandler) ServeUserLikedPosts(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse pagination parameters
	paginationParams := utils.ParsePaginationFromRequest(r)
	sortBy := utils.ParseSortFromRequest(r, "newest")

	// Get session cookie for API calls
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's liked posts from backend API
	postsResponse, err := h.userService.GetUserLikedPosts(user.ID, paginationParams.Limit, paginationParams.Offset, sortBy, sessionCookie)
	if err != nil {
		if err.Error() == "unauthorized: please log in" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to load your liked posts", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":         user,
		"Posts":        postsResponse.Posts,
		"Pagination":   postsResponse.Pagination,
		"PageTitle":    "Posts I Liked",
		"PageType":     "liked-posts",
		"EmptyMessage": "You haven't liked any posts yet.",
		"CreateLink":   "/",
	}

	// Render the user posts template (same template, different data)
	if err := h.templateService.Render(w, "user-posts.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// ServeUserCommentedPosts handles the "Commented Posts" page - shows posts the user commented on
func (h *ProfileHandler) ServeUserCommentedPosts(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is logged in
	user := session.GetUserFromSession(r, h.authService)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse pagination parameters
	paginationParams := utils.ParsePaginationFromRequest(r)
	sortBy := utils.ParseSortFromRequest(r, "newest")

	// Get session cookie for API calls
	sessionCookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function instead of hardcoded "session_id"
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get posts user commented on from backend API
	postsResponse, err := h.userService.GetUserCommentedPosts(user.ID, paginationParams.Limit, paginationParams.Offset, sortBy, sessionCookie)
	if err != nil {
		if err.Error() == "unauthorized: please log in" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to load posts you commented on", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := map[string]interface{}{
		"User":         user,
		"Posts":        postsResponse.Posts,
		"Pagination":   postsResponse.Pagination,
		"PageTitle":    "Posts I Commented On",
		"PageType":     "commented-posts",
		"EmptyMessage": "You haven't commented on any posts yet.",
		"CreateLink":   "/",
	}

	// Render the user posts template (same template, different data)
	if err := h.templateService.Render(w, "user-posts.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
