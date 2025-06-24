package utils

import (
	"net/http"
	"time"
)

// ClearSessionCookie clears the session cookie by setting it to expire
func ClearSessionCookie(name string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,                // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Consistent with frontend
	})
}

// SetSessionCookie sets a session cookie with the provided name, value, and expiration
func SetSessionCookie(name, value string, w http.ResponseWriter, r *http.Request, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,                // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // Consistent with frontend
	})
}
