package middleware

import (
	"net/http"
	"strings"

	"github.com/PaulKerasidis/forum/config"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Since ALLOWED_ORIGINS=http://localhost:3000 (no wildcard)
		// Check if origin is in allowed list
		allowedOrigins := strings.Split(config.Config.AllowedOrigins, ",")
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			trimmedOrigin := strings.TrimSpace(allowedOrigin)
			if trimmedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				originAllowed = true
				break
			}
		}

		// Only set credentials if origin is explicitly allowed
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", config.Config.AllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", config.Config.AllowedHeaders)
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
