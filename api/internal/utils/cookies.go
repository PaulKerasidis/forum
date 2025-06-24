package utils

import (
	"net/http"
	"time"

	"github.com/PaulKerasidis/forum/config"
)

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.Config.SessionName, // CHANGED: Use config session name
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,                // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // CHANGED: Consistent with frontend
		Domain:   "localhost",          // Allow cookie to be shared across ports
	})
}

func SetSessionCookie(value string, w http.ResponseWriter, r *http.Request, expiresAt time.Time) {
	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     config.Config.SessionName, // CHANGED: Use config session name
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,                // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode, // CHANGED: Consistent with frontend
		Domain:   "localhost",          // Allow cookie to be shared across ports
	})
}
