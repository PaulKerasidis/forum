package session

import (
	"net/http"

	"frontend-service/internal/models"
	"frontend-service/internal/services"
)

// GetUserFromSession gets user from session cookie and validates with backend
func GetUserFromSession(r *http.Request, authService *services.AuthService) *models.User {
	// Get session cookie using the session name from auth service
	cookie, err := r.Cookie(authService.SessionName) // CHANGED: Use SessionName from authService instead of hardcoded "session_id"
	if err != nil {
		// No session cookie found
		return nil
	}

	// Validate session with backend
	user, err := authService.ValidateSession(cookie.Value)
	if err != nil {
		// Session is invalid or expired
		return nil
	}

	return user
}

// GetSessionCookie is a helper function to get the session cookie by name from auth service
func GetSessionCookie(r *http.Request, authService *services.AuthService) (*http.Cookie, error) {
	// ADDED: Helper function to get session cookie using the correct name from config
	return r.Cookie(authService.SessionName)
}
