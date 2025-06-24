package utils

import (
	"net/http"
	"strconv"
)

// PaginationParams holds the pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
	Page   int
}

// ParsePaginationFromRequest extracts pagination parameters from HTTP request
func ParsePaginationFromRequest(r *http.Request) PaginationParams {
	// Default values
	limit := 20
	offset := 0
	page := 1

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Calculate page from offset
	if limit > 0 {
		page = (offset / limit) + 1
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}

// ParseSortFromRequest extracts sort parameter from HTTP request
func ParseSortFromRequest(r *http.Request, defaultSort string) string {
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		return defaultSort
	}
	return sortBy
}

// BuildPaginationURL creates a URL with pagination parameters
func BuildPaginationURL(baseURL string, limit, offset int) string {
	if baseURL == "" {
		baseURL = "?"
	} else if baseURL[len(baseURL)-1] != '?' && baseURL[len(baseURL)-1] != '&' {
		baseURL += "?"
	}

	return baseURL + "limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
}

// CalculateOffset calculates offset from page number
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}
