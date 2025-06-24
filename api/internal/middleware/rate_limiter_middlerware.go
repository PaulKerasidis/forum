package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/PaulKerasidis/forum/internal/utils"
)

// RateLimiter tracks request rates per IP and limits excessive requests
type RateLimiter struct {
	requests    map[string][]time.Time
	windowSize  time.Duration
	maxRequests int
	mutex       sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified window size and max requests
func NewRateLimiter(windowSize time.Duration, maxRequests int) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		windowSize:  windowSize,
		maxRequests: maxRequests,
		mutex:       sync.Mutex{},
	}
}

// cleanupOldRequests removes requests outside the current time window
func (rl *RateLimiter) cleanupOldRequests(ip string) {
	now := time.Now()
	keepFrom := now.Add(-rl.windowSize)

	var validRequests []time.Time
	for _, reqTime := range rl.requests[ip] {
		if reqTime.After(keepFrom) {
			validRequests = append(validRequests, reqTime)
		}
	}

	rl.requests[ip] = validRequests
}

// checkRateLimit checks if an IP has exceeded the rate limit
func (rl *RateLimiter) checkRateLimit(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// First cleanup old requests
	rl.cleanupOldRequests(ip)

	// Check current count against limit
	return len(rl.requests[ip]) >= rl.maxRequests
}

// addRequest records a new request for an IP
func (rl *RateLimiter) addRequest(ip string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// First cleanup old requests
	rl.cleanupOldRequests(ip)

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], time.Now())
}

// Limit is the middleware handler for rate limiting
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		// Check if rate limited
		if rl.checkRateLimit(ip) {
			utils.RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			return
		}

		// Record this request
		rl.addRequest(ip)

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}
