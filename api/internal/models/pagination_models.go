package models

// PaginationInfo contains metadata about paginated results
type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
	PerPage     int  `json:"per_page"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// PaginatedPostsResponse is the response for paginated posts
type PaginatedPostsResponse struct {
	Posts      []*Post        `json:"posts"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginatedCommentsResponse is the response for paginated comments
type PaginatedCommentsResponse struct {
	Comments   []*Comment     `json:"comments"`
	Pagination PaginationInfo `json:"pagination"`
}

// NewPaginationInfo creates pagination metadata from basic parameters
func NewPaginationInfo(totalCount, limit, offset int) PaginationInfo {
	// Calculate pagination values
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
	currentPage := (offset / limit) + 1

	// Handle edge cases
	if totalPages == 0 {
		totalPages = 1
		currentPage = 1
	}

	return PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		PerPage:     limit,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}
}

// NewPaginatedPostsResponse creates a paginated posts response
func NewPaginatedPostsResponse(posts []*Post, totalCount, limit, offset int) *PaginatedPostsResponse {
	return &PaginatedPostsResponse{
		Posts:      posts,
		Pagination: NewPaginationInfo(totalCount, limit, offset),
	}
}

// NewPaginatedCommentsResponse creates a paginated comments response
func NewPaginatedCommentsResponse(comments []*Comment, totalCount, limit, offset int) *PaginatedCommentsResponse {
	return &PaginatedCommentsResponse{
		Comments:   comments,
		Pagination: NewPaginationInfo(totalCount, limit, offset),
	}
}
