package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"frontend-service/internal/models"
)

type CommentService struct {
	*BaseClient
}

// NewCommentService creates a new comment service
func NewCommentService(baseClient *BaseClient) *CommentService {
	return &CommentService{
		BaseClient: baseClient,
	}
}

// CreateComment creates a new comment on a post
func (s *CommentService) CreateComment(postID, content string, sessionCookie *http.Cookie) (*models.Comment, error) {
	// Prepare request data
	requestData := map[string]interface{}{
		"content": content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for create comment
	createURL := s.BaseURL + "/comments/create-on-post/" + postID

	// Create HTTP request
	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("failed to create comment: %w", err)
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
	if resp.StatusCode != http.StatusCreated {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return nil, fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return nil, fmt.Errorf("create comment failed with status %d: %s", resp.StatusCode, string(body))
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

	// Convert data to Comment (the API returns the created comment)
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var comment models.Comment
	if err := json.Unmarshal(dataBytes, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment data: %w", err)
	}

	return &comment, nil
}

// UpdateComment updates an existing comment
func (s *CommentService) UpdateComment(commentID, content string, sessionCookie *http.Cookie) error {
	// Prepare request data
	requestData := map[string]interface{}{
		"content": content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for update comment
	updateURL := s.BaseURL + "/comments/edit/" + commentID

	// Create HTTP request
	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
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
		return fmt.Errorf("failed to update comment: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: please log in")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden: you can only edit your own comments")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("comment not found")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return fmt.Errorf("update comment failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response to check for API errors
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return fmt.Errorf("API error: %s", apiResponse.Error)
	}

	return nil
}

// DeleteComment deletes a comment
func (s *CommentService) DeleteComment(commentID string, sessionCookie *http.Cookie) error {
	// Build URL for delete comment
	deleteURL := s.BaseURL + "/comments/remove/" + commentID

	// Create HTTP request
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: please log in")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden: you can only delete your own comments")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("comment not found")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return fmt.Errorf("delete comment failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response to check for API errors
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return fmt.Errorf("API error: %s", apiResponse.Error)
	}

	return nil
}

// GetCommentByID retrieves a single comment by ID
func (s *CommentService) GetCommentByID(commentID string, sessionCookie *http.Cookie) (*models.Comment, error) {
	// Build URL for get single comment
	commentURL := s.BaseURL + "/comments/view/" + commentID

	// Create HTTP request
	req, err := http.NewRequest("GET", commentURL, nil)
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
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("comment not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get comment: %s", string(body))
	}

	// Parse JSON response
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	// Convert data to Comment
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var comment models.Comment
	if err := json.Unmarshal(dataBytes, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment data: %w", err)
	}

	return &comment, nil
}
