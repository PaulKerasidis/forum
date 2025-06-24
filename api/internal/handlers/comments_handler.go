package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/PaulKerasidis/forum/internal/middleware"
	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// Create comment handler.
func CreateCommentHandler(cor *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Get postID from URL path, not request body
		postID := r.PathValue("id")
		if postID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
			return
		}

		// Parse request body (only content, no postID)
		var req models.CreateCommentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate comment content
		if err := utils.ValidateCommentContent(req.Content); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create comment - now returns lightweight response
		createResponse, err := cor.CreateComment(postID, user.ID, req.Content)
		if err != nil {
			if err.Error() == "post not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Post not found")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create comment")
			return
		}

		// Return lightweight response
		utils.RespondWithSuccess(w, http.StatusCreated, createResponse)
	}
}

// update comment handler.
func UpdateCommentHandler(cor *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Get comment ID from URL path
		commentID := r.PathValue("id")
		if commentID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Comment ID is required")
			return
		}

		// Parse request body
		var req models.UpdateCommentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate comment content
		if err := utils.ValidateCommentContent(req.Content); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Update the comment
		err = cor.UpdateComment(commentID, user.ID, req.Content)
		if err != nil {
			if err.Error() == "comment not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Comment not found")
				return
			}
			if err.Error() == "unauthorized: you can only update your own comments" {
				utils.RespondWithError(w, http.StatusForbidden, "You can only update your own comments")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update comment")
			return
		}

		// Respond with success
		utils.RespondWithSuccess(w, http.StatusOK, "Comment updated successfully")
	}
}

// delete comment handler.
func DeleteCommentHandler(cor *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Get comment ID from URL path
		commentID := r.PathValue("id")
		if commentID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Comment ID is required")
			return
		}

		// Delete comment base on userID
		err := cor.DeleteComment(commentID, user.ID)
		if err != nil {
			if err.Error() == "comment not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Comment not found")
				return
			}
			if err.Error() == "unauthorized: you can only delete your own comments" {
				utils.RespondWithError(w, http.StatusForbidden, "You can only delete your own comments")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, "Comment deleted successfully")
	}
}

// Get ALL comments by post ID handler.
func GetCommentsByPostIDHandler(cor *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get user context
		currentUser := middleware.GetCurrentUser(r)
		var userID *string = nil
		if currentUser != nil {
			userID = &currentUser.ID
		}

		// Get post ID from URL
		postID := r.PathValue("id")
		if postID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Post ID is required")
			return
		}

		// Parse pagination and sort parameters using unified system
		limit, offset := utils.ParsePaginationParams(r)
		// Parse sort options using unified system
		sortOptions := utils.ParseCommentSortOptions(r)

		// Get total count
		totalCount, err := cor.GetCommentCountByPost(postID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve comment count")
			return
		}

		// Get comments with sorting
		comments, err := cor.GetCommentsByPostID(postID, limit, offset, userID, sortOptions)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve comments")
			return
		}

		// Respond with paginated comments
		utils.RespondWithPaginatedComments(w, comments, totalCount, limit, offset)
	}
}

// GetSingleCommentHandler retrieves a single comment by ID
func GetSingleCommentHandler(cor *repository.CommentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get comment ID from URL
		commentID := r.PathValue("id")
		if commentID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Comment ID is required")
			return
		}

		// Get user context (optional for this endpoint)
		currentUser := middleware.GetCurrentUser(r)
		var userID *string = nil
		if currentUser != nil {
			userID = &currentUser.ID
		}

		// Get the comment - you'll need to add this method to your repository
		comment, err := cor.GetCommentByID(commentID, userID)
		if err != nil {
			if err.Error() == "comment not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Comment not found")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve comment")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, comment)
	}
}
