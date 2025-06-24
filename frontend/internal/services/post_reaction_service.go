package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"frontend-service/internal/models"
)

type PostReactionService struct {
	*BaseClient
}

// NewPostReactionService creates a new post reaction service
func NewPostReactionService(baseClient *BaseClient) *PostReactionService {
	return &PostReactionService{
		BaseClient: baseClient,
	}
}

// TogglePostReaction toggles a like/dislike reaction on a post
func (s *PostReactionService) TogglePostReaction(postID string, reactionType int, sessionCookie *http.Cookie) (*models.ReactionResult, error) {
	// Prepare request data
	requestData := models.PostReactionRequest{
		PostID:       postID,
		ReactionType: reactionType,
	}

	// Basic validation
	if requestData.PostID == "" {
		return nil, fmt.Errorf("post ID is required")
	}
	if requestData.ReactionType != models.ReactionTypeLike && requestData.ReactionType != models.ReactionTypeDislike {
		return nil, fmt.Errorf("invalid reaction type: must be 1 (like) or 2 (dislike)")
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for toggle post reaction
	toggleURL := s.BaseURL + "/reactions/posts/toggle"

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
		return nil, fmt.Errorf("failed to toggle post reaction: %w", err)
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
		return nil, fmt.Errorf("toggle post reaction failed with status %d: %s", resp.StatusCode, string(body))
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

// GetPostReactionStatus gets the current user's reaction status on a post
func (s *PostReactionService) GetPostReactionStatus(postID string, sessionCookie *http.Cookie) (*int, error) {
	// Build URL for get post reaction status
	statusURL := s.BaseURL + "/reactions/posts/status/" + postID

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
		return nil, fmt.Errorf("failed to get post reaction status: %w", err)
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
		return nil, fmt.Errorf("get post reaction status failed with status %d: %s", resp.StatusCode, string(body))
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
		PostID       string `json:"post_id"`
		UserReaction *int   `json:"user_reaction"` // nil, 1=like, 2=dislike
	}
	if err := json.Unmarshal(dataBytes, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse reaction status: %w", err)
	}

	return statusResponse.UserReaction, nil
}
