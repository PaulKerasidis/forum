package utils

import (
	"net/http"
	"strings"
)

// SortOptions defines the sorting options for both posts and comments
type SortOptions struct {
	SortBy string // "newest", "oldest", "likes", "comments"
}

// ContentType represents the type of content being sorted
type ContentType string

const (
	ContentTypePosts    ContentType = "posts"
	ContentTypeComments ContentType = "comments"
)

// IsValidSortOption checks if the sort option is valid for the given content type
var validSortOptions = map[ContentType]map[string]bool{
	ContentTypeComments: {"oldest": true, "newest": true, "likes": true},
	ContentTypePosts:    {"newest": true, "oldest": true, "likes": true, "comments": true},
}

// ParseSortOptions extracts and validates sort options from HTTP request query parameters
func ParseSortOptions(r *http.Request, contentType ContentType) SortOptions {
	options := defaultSortOptions(contentType)

	// Parse sort parameter
	sort := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
	if isValidSortOption(sort, contentType) {
		options.SortBy = sort
	}

	return options
}

// BuildOrderClause returns the SQL ORDER BY clause based on sort option and content type
func BuildOrderClause(sortBy string, contentType ContentType) string {
	switch contentType {
	case ContentTypeComments:
		return buildCommentOrderClause(sortBy)
	case ContentTypePosts:
		fallthrough
	default:
		return buildPostOrderClause(sortBy)
	}
}

// defaultSortOptions returns the default sorting options based on content type
func defaultSortOptions(contentType ContentType) SortOptions {
	if contentType == ContentTypeComments {
		return SortOptions{SortBy: "oldest"}
	}
	// Everything else (ContentTypePosts or unknown) defaults to newest
	return SortOptions{SortBy: "newest"}
}
func isValidSortOption(sort string, contentType ContentType) bool {
	if options, exists := validSortOptions[contentType]; exists {
		return options[sort]
	}
	return validSortOptions[ContentTypePosts][sort] // Default fallback
}

// buildPostOrderClause builds ORDER BY clause for posts
func buildPostOrderClause(sortBy string) string {
	switch sortBy {
	case "likes":
		return "ORDER BY like_count DESC, p.created_at DESC"
	case "comments":
		return "ORDER BY comment_count DESC, p.created_at DESC"
	case "oldest":
		return "ORDER BY p.created_at ASC"
	case "newest":
		fallthrough
	default:
		return "ORDER BY p.created_at DESC"
	}
}

// buildCommentOrderClause builds ORDER BY clause for comments
func buildCommentOrderClause(sortBy string) string {
	switch sortBy {
	case "newest":
		return "ORDER BY c.created_at DESC"
	case "likes":
		return "ORDER BY like_count DESC, c.created_at ASC"
	case "oldest":
		fallthrough
	default:
		return "ORDER BY c.created_at ASC" // Default: conversation order
	}
}

// Legacy function names for backward compatibility (if needed)

// ParsePostSortOptions - wrapper for posts
func ParsePostSortOptions(r *http.Request) SortOptions {
	return ParseSortOptions(r, ContentTypePosts)
}

// ParseCommentSortOptions - wrapper for comments
func ParseCommentSortOptions(r *http.Request) SortOptions {
	return ParseSortOptions(r, ContentTypeComments)
}
