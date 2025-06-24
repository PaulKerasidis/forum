package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"frontend-service/internal/services"
	"frontend-service/internal/utils"
)

type OAuthHandler struct {
	authService     *services.AuthService
	templateService *services.TemplateService
	baseURL         string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(authService *services.AuthService, templateService *services.TemplateService, baseURL string) *OAuthHandler {
	return &OAuthHandler{
		authService:     authService,
		templateService: templateService,
		baseURL:         baseURL,
	}
}

// GoogleLoginHandler initiates Google OAuth login
func (h *OAuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Call backend to get Google auth URL
	authURL, err := h.authService.InitiateGoogleOAuth()
	if err != nil {
		log.Printf("Failed to initiate Google OAuth: %v", err)
		http.Error(w, "Failed to initiate Google login", http.StatusInternalServerError)
		return
	}

	// Redirect to Google OAuth
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles Google OAuth callback
func (h *OAuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		log.Printf("OAuth error: %s", errorParam)
		h.showLoginError("Google login was cancelled or failed")
		return
	}

	// Validate required parameters
	if code == "" || state == "" {
		h.showLoginError("Invalid OAuth response")
		return
	}

	// Forward to backend for processing
	user, sessionID, isNewUser, err := h.authService.HandleGoogleCallback(code, state)
	if err != nil {
		log.Printf("Google OAuth callback error: %v", err)
		h.showLoginError(err.Error())
		return
	}

	// Set session cookie
	expiresAt := time.Now().Add(24 * time.Hour)
	utils.SetSessionCookie("forum_session", sessionID, w, r, expiresAt)

	log.Printf("Google OAuth successful for user: %s (new user: %t)", user.Username, isNewUser)

	// Redirect to home page with success
	http.Redirect(w, r, "/?oauth=success", http.StatusSeeOther)
}

// showLoginError redirects to login page with error
func (h *OAuthHandler) showLoginError(errorMsg string) {
	// In a real implementation, you might store the error in a session
	// and redirect to login page to display it
	log.Printf("OAuth login error: %s", errorMsg)
}

// OAuthStatusHandler returns OAuth configuration for frontend
func (h *OAuthHandler) OAuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get OAuth status from backend
	status, err := h.authService.GetOAuthStatus()
	if err != nil {
		log.Printf("Failed to get OAuth status: %v", err)
		http.Error(w, "Failed to get OAuth status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
