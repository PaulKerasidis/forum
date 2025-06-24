package middleware

import "net/http"

// SecurityHeaders is a middleware that sets security-related HTTP headers
// to protect against common web vulnerabilities.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ✅ ALWAYS NEEDED: Prevent MIME sniffing attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// ✅ MINIMAL CSP for JSON API only
		csp := "default-src 'none'; " + // Block everything by default
			"connect-src 'self' http://localhost:3000; " + // Allow API calls
			"frame-ancestors 'none'; " + // Prevent iframe embedding
			"base-uri 'self'; " + // Prevent base tag hijacking
			"form-action 'self'" // Prevent form hijacking

		w.Header().Set("Content-Security-Policy", csp)

		// ✅ GOOD TO HAVE: Privacy protection
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
