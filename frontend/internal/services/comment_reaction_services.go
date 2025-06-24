package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"frontend-service/internal/models"
)

type CommentReactionService struct {
	*BaseClient
}

// NewCommentReactionService creates a new comment reaction service
func NewCommentReactionService(baseClient *BaseClient) *CommentReactionService {
	return &CommentReactionService{
		BaseClient: baseClient,
	}
}

// ToggleCommentReaction toggles a like/dislike reaction on a comment
func (s *CommentReactionService) ToggleCommentReaction(commentID string, reactionType int, sessionCookie *http.Cookie) (*models.ReactionResult, error) {
	// Prepare request data
	requestData := models.CommentReactionRequest{
		CommentID:    commentID,
		ReactionType: reactionType,
	}

	// Basic validation
	if requestData.CommentID == "" {
		return nil, fmt.Errorf("comment ID is required")
	}
	if requestData.ReactionType != models.ReactionTypeLike && requestData.ReactionType != models.ReactionTypeDislike {
		return nil, fmt.Errorf("invalid reaction type: must be 1 (like) or 2 (dislike)")
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for toggle comment reaction
	toggleURL := s.BaseURL + "/reactions/comments/toggle"

	// Create HTTP request
	req, err := http.NewRequest("POST", toggleURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add session cookie for authentication
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle comment reaction: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: please log in")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return nil, fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return nil, fmt.Errorf("toggle comment reaction failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	// Convert data to ReactionResult
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var reactionResult models.ReactionResult
	if err := json.Unmarshal(dataBytes, &reactionResult); err != nil {
		return nil, fmt.Errorf("failed to parse reaction result: %w", err)
	}

	return &reactionResult, nil
}

// GetCommentReactionStatus gets the current user's reaction status on a comment
func (s *CommentReactionService) GetCommentReactionStatus(commentID string, sessionCookie *http.Cookie) (*int, error) {
	// Build URL for get comment reaction status
	statusURL := s.BaseURL + "/reactions/comments/status/" + commentID

	// Create HTTP request
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment reaction status: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: please log in")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return nil, fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return nil, fmt.Errorf("get comment reaction status failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	// Convert data to reaction status
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var statusResponse struct {
		CommentID    string `json:"comment_id"`
		UserReaction *int   `json:"user_reaction"` // nil, 1=like, 2=dislike
	}
	if err := json.Unmarshal(dataBytes, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse reaction status: %w", err)
	}

	return statusResponse.UserReaction, nil
}
