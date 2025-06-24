package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/utils"
)

type HomeHandler struct {
	authService     *services.AuthService
	postService     *services.PostService
	categoryService *services.CategoryService
	templateService *services.TemplateService
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(authService *services.AuthService, postService *services.PostService, categoryService *services.CategoryService, templateService *services.TemplateService) *HomeHandler {
	return &HomeHandler{
		authService:     authService,
		postService:     postService,
		categoryService: categoryService,
		templateService: templateService,
	}
}

// ServeHome handles the main page request
func (h *HomeHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	// Handle OAuth success - set session cookie if session_id is provided
	if sessionID := r.URL.Query().Get("session_id"); sessionID != "" {
		log.Printf("OAuth callback detected! Setting session cookie: %s", sessionID)
		// Set session cookie with 24 hour expiration
		expiresAt := time.Now().Add(24 * time.Hour)
		utils.SetSessionCookie("forum_session", sessionID, w, r, expiresAt)
		log.Printf("Session cookie set, redirecting to clean URL")
		
		// Redirect to clean URL without session_id parameter
		http.Redirect(w, r, "/?oauth=success", http.StatusTemporaryRedirect)
		return
	}

	// Parse pagination parameters from query string
	limit := 10 // Default posts per page
	offset := 0 // Default starting point

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// 🔧 FIX: Parse sort parameter (this was missing!)
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "newest" // Default sort
	}

	// Check if user is logged in and get session cookie
	user := session.GetUserFromSession(r, h.authService)
	var sessionCookie *http.Cookie
	if user != nil {
		log.Printf("User is logged in: %s (%s)", user.Username, user.Email)
		sessionCookie, _ = session.GetSessionCookie(r, h.authService)
	} else {
		log.Printf("No user found in session")
		// Debug: check if there's a session cookie at all
		if cookie, err := r.Cookie("forum_session"); err == nil {
			log.Printf("Session cookie exists but user validation failed: %s", cookie.Value)
		} else {
			log.Printf("No session cookie found")
		}
	}

	// 🔧 FIX: Get posts from backend API WITH sort parameter
	postsResponse, err := h.postService.GetAllPosts(limit, offset, sortBy, sessionCookie)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Get categories from backend API
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	// 🔧 FIX: Prepare data for template WITH sort information
	data := models.HomePageData{
		Posts:      postsResponse.Posts,
		Categories: categories,
		Pagination: postsResponse.Pagination,
		User:       user,
		Sort:       sortBy, // 🔧 ADD: Pass sort parameter to template
	}

	// Render the template
	if err := h.templateService.Render(w, "home.html", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
