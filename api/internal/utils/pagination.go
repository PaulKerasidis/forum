package utils

import (
	"net/http"
	"strconv"

	"github.com/PaulKerasidis/forum/config"
)

// ParsePaginationParams extracts and validates pagination parameters from HTTP request
func ParsePaginationParams(r *http.Request) (limit, offset int) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Convert to integers
	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	// Validate and normalize using the model helper
	return ValidatePaginationParams(limit, offset)
}

// ValidatePaginationParams validates and normalizes pagination parameters
func ValidatePaginationParams(limit, offset int) (int, int) {
	// Set defaults and validate limit using config
	if limit <= 0 {
		limit = config.Config.DefaultPageSize
	}
	if limit > config.Config.MaxPageSize {
		limit = config.Config.MaxPageSize
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
