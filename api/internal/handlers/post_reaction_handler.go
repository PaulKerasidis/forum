package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/PaulKerasidis/forum/internal/middleware"
	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// TogglePostReactionHandler handles like/dislike toggle for posts
func TogglePostReactionHandler(prr *repository.PostReactionRepository) http.HandlerFunc {
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

		// Parse request body
		var req models.PostReactionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate the request
		if err := req.Validate(); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Call repository method to toggle post reaction
		result, err := prr.TogglePostReaction(user.ID, req.PostID, req.ReactionType)
		if err != nil {
			// Handle specific errors from repository
			if err.Error() == "post not found" {
				utils.RespondWithError(w, http.StatusNotFound, "Post not found")
				return
			}
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to toggle reaction")
			return
		}

		// Return the detailed reaction result
		utils.RespondWithSuccess(w, http.StatusOK, result)
	}
}
