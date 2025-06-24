package utils

import (
	"encoding/json"
	"net/http"

	"github.com/PaulKerasidis/forum/internal/models"
)

// RespondWithError sends a standardized error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespondWithPaginatedPosts sends a standardized paginated posts response
func RespondWithPaginatedPosts(w http.ResponseWriter, posts []*models.Post, totalCount, limit, offset int) {
	response := models.NewPaginatedPostsResponse(posts, totalCount, limit, offset)
	RespondWithSuccess(w, http.StatusOK, response)
}

// RespondWithPaginatedComments sends a standardized paginated comments response
func RespondWithPaginatedComments(w http.ResponseWriter, comments []*models.Comment, totalCount, limit, offset int) {
	response := models.NewPaginatedCommentsResponse(comments, totalCount, limit, offset)
	RespondWithSuccess(w, http.StatusOK, response)
}
