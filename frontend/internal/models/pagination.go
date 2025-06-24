package models

// PaginationInfo - Pagination metadata (matches backend exactly)
type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	TotalCount  int  `json:"total_count"`
	PerPage     int  `json:"per_page"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// PaginatedPostsResponse - Paginated posts response (matches backend exactly)
type PaginatedPostsResponse struct {
	Posts      []*Post        `json:"posts"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginatedCommentsResponse - Paginated comments response (matches backend exactly)
type PaginatedCommentsResponse struct {
	Comments   []*Comment     `json:"comments"`
	Pagination PaginationInfo `json:"pagination"`
}
