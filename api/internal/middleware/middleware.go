package middleware

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

type AuthMiddleware struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewMiddleware(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

type contextKey string

// Define constants using this type
const (
	userContextKey contextKey = "user"
)

// Authenticate middleware verifies authentication and sets user in context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session cookie using config session name
		cookie, err := r.Cookie(config.Config.SessionName)
		if err != nil {
			// No cookie means not authenticated - just continue
			next.ServeHTTP(w, r)
			return
		}

		// Validate the session
		session, err := m.sessionRepo.GetBySessionID(cookie.Value)
		if err != nil {
			// Session invalid or expired - clear the cookie
			utils.ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Get current client IP and clean it (remove port if present)
		currentIP := getClientIP(r)

		// Clean both IPs to ensure consistent comparison
		storedIP := cleanIP(session.IPAddress)
		cleanCurrentIP := cleanIP(currentIP)

		// Smart IP validation using cleaned IPs
		if shouldLogoutDueToIPChange(storedIP, cleanCurrentIP) {
			log.Printf("🛡️ SECURITY: Suspicious IP change detected - forcing logout")
			log.Printf("   User: %s", session.UserID)
			log.Printf("   Original IP: %s", storedIP)
			log.Printf("   New IP: %s", cleanCurrentIP)
			log.Printf("   Reason: Major network/location change detected")

			// Delete the session for security
			err = m.sessionRepo.DeleteSession(cookie.Value)
			if err != nil {
				log.Printf("Error deleting session: %v", err)
			}

			// Clear the cookie
			utils.ClearSessionCookie(w)

			// Continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// If IP changed but it's safe, update the session
		if storedIP != cleanCurrentIP {
			log.Printf("🔄 INFO: Updating session IP for user %s: %s -> %s",
				session.UserID, storedIP, cleanCurrentIP)

			// Update session with new CLEAN IP (no port)
			if updateErr := m.sessionRepo.UpdateSessionIP(session.SessionID, cleanCurrentIP); updateErr != nil {
				log.Printf("Warning: Failed to update session IP: %v", updateErr)
			}
		}

		// Get the user
		user, err := m.userRepo.GetUserBySessionID(session.UserID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Set user in context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth middleware ensures the user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextKey)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, errors.New("unauthorized access").Error())
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetCurrentUser returns the authenticated user from the context
func GetCurrentUser(r *http.Request) *models.User {
	userValue := r.Context().Value(userContextKey)

	if userValue == nil {
		return nil
	}

	user, ok := userValue.(*models.User)
	if !ok {
		return nil
	}

	return user
}

// Clean IP address - remove port if present and normalize localhost
func cleanIP(ipWithPossiblePort string) string {
	// Handle the case where IP might include port (like "127.0.0.1:55394")
	host, _, err := net.SplitHostPort(ipWithPossiblePort)
	if err != nil {
		// If SplitHostPort fails, it's probably just an IP without port
		host = ipWithPossiblePort
	}
	
	// Normalize localhost - treat IPv4 and IPv6 localhost as the same
	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return "localhost"
	}
	
	return host
}

// Smart IP validation - returns true if user should be logged out
func shouldLogoutDueToIPChange(originalIP, currentIP string) bool {
	if originalIP == currentIP {
		return false // No change - all good
	}

	// Parse both IP addresses
	origParsed := net.ParseIP(originalIP)
	currParsed := net.ParseIP(currentIP)

	if origParsed == nil || currParsed == nil {
		return true // Invalid IP format - logout for safety
	}

	// Both are private IPs (home networks) - check if same subnet
	if isPrivateIP(origParsed) && isPrivateIP(currParsed) {
		// Allow changes within same home network (DHCP renewals)
		return !isSameSubnet(originalIP, currentIP, "/24")
	}

	// Both are public IPs - check if same ISP range
	if !isPrivateIP(origParsed) && !isPrivateIP(currParsed) {
		// Allow changes within same ISP (dynamic IP assignments)
		return !isSameISPRange(originalIP, currentIP)
	}

	// One private, one public = major change (home to mobile)
	return true
}

// Check if IP is in private network range (RFC 1918)
func isPrivateIP(ip net.IP) bool {
	// Handle localhost normalization first
	ipStr := ip.String()
	if ipStr == "127.0.0.1" || ipStr == "::1" || ipStr == "localhost" {
		return true
	}
	
	// Private IPv4 ranges: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	privateRanges := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
		"127.0.0.0/8",    // IPv4 Localhost
		"::1/128",        // IPv6 Localhost
		"fc00::/7",       // IPv6 private range
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

// Check if two IPs are in the same subnet
func isSameSubnet(ip1, ip2, mask string) bool {
	// Create subnet from first IP
	_, subnet, err := net.ParseCIDR(ip1 + mask)
	if err != nil {
		return false
	}

	// Check if second IP is in that subnet
	ip2Parsed := net.ParseIP(ip2)
	if ip2Parsed == nil {
		return false
	}

	return subnet.Contains(ip2Parsed)
}

// Check if two public IPs are from same ISP (broad range check)
func isSameISPRange(ip1, ip2 string) bool {
	// For public IPs, check /16 range (usually same ISP/region)
	return isSameSubnet(ip1, ip2, "/16")
}

// Get real client IP address (handles proxies and load balancers)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common proxy header)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Can contain multiple IPs, take the first (original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx reverse proxy)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fall back to direct connection IP
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return as-is if can't parse
	}
	return host
}
