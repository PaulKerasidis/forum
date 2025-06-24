package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"frontend-service/internal/models"
)

type PostService struct {
	*BaseClient
}

// NewPostService creates a new post service
func NewPostService(baseClient *BaseClient) *PostService {
	return &PostService{
		BaseClient: baseClient,
	}
}

// GetAllPosts retrieves posts from the backend API (updated to accept sort parameter)
func (s *PostService) GetAllPosts(limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.PaginatedPostsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/posts")
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	// 🔧 FIX: Add sort parameter if provided
	if sortBy != "" {
		params.Add("sort", sortBy)
	}
	u.RawQuery = params.Encode()

	// Create request (instead of using GET directly)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie if provided (for user reaction data)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
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

	// Convert data to PaginatedPostsResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var postsResponse models.PaginatedPostsResponse
	if err := json.Unmarshal(dataBytes, &postsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse posts data: %w", err)
	}

	return &postsResponse, nil
}

// GetPostsByCategory retrieves posts filtered by category from the backend API
func (s *PostService) GetPostsByCategory(categoryID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.CategoryPostsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/posts/by-category/" + categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	if sortBy != "" {
		params.Add("sort", sortBy)
	}
	u.RawQuery = params.Encode()

	// Create request instead of using GET directly
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie if provided (for user reaction data)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category posts: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
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

	// Convert data to CategoryPostsResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var categoryResponse models.CategoryPostsResponse
	if err := json.Unmarshal(dataBytes, &categoryResponse); err != nil {
		return nil, fmt.Errorf("failed to parse category posts data: %w", err)
	}

	return &categoryResponse, nil
}

// GetSinglePostWithComments retrieves a single post and its comments from the backend API
// Returns data formatted for the existing PostPageData struct
func (s *PostService) GetSinglePostWithComments(postID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.Post, []*models.Comment, error) {
	// First, get the post
	post, err := s.getSinglePost(postID, sessionCookie)
	if err != nil {
		return nil, nil, err
	}

	// Then, get the comments (we'll extract just the comments array)
	commentsResponse, err := s.getPostComments(postID, limit, offset, sortBy, sessionCookie)
	if err != nil {
		return nil, nil, err
	}

	// Return post and comments array (not the pagination wrapper)
	return post, commentsResponse.Comments, nil
}

// Helper method to get a single post (updated to accept session cookie)
func (s *PostService) getSinglePost(postID string, sessionCookie *http.Cookie) (*models.Post, error) {
	// Build URL for single post
	u, err := url.Parse(s.BaseURL + "/posts/view/" + postID)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create request instead of using GET directly
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie if provided (for user reaction data)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("post not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
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

	// Convert data to Post
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var post models.Post
	if err := json.Unmarshal(dataBytes, &post); err != nil {
		return nil, fmt.Errorf("failed to parse post data: %w", err)
	}

	return &post, nil
}

// Helper method to get post comments (updated to accept session cookie)
func (s *PostService) getPostComments(postID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.PaginatedCommentsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/comments/for-post/" + postID)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	if sortBy != "" {
		params.Add("sort", sortBy)
	}
	u.RawQuery = params.Encode()

	// Create request instead of using GET directly
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie if provided (for user reaction data on comments)
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
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

	// Convert data to PaginatedCommentsResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var commentsResponse models.PaginatedCommentsResponse
	if err := json.Unmarshal(dataBytes, &commentsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse comments data: %w", err)
	}

	return &commentsResponse, nil
}

// Add this method to your existing post_service.go file

// CreatePost submits a new post to the backend API
func (s *PostService) CreatePost(categoryNames []string, content string, sessionCookie *http.Cookie) (*models.CreatePostResponse, error) {
	// Prepare request data
	requestData := map[string]interface{}{
		"category_names": categoryNames,
		"content":        content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for create post
	createURL := s.BaseURL + "/posts/create"

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
		return nil, fmt.Errorf("failed to create post: %w", err)
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
		return nil, fmt.Errorf("create post failed with status %d: %s", resp.StatusCode, string(body))
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

	// Convert data to CreatePostResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var createResponse models.CreatePostResponse
	if err := json.Unmarshal(dataBytes, &createResponse); err != nil {
		return nil, fmt.Errorf("failed to parse create post response: %w", err)
	}

	return &createResponse, nil
}

// Add these methods to your existing post_service.go file

// UpdatePost updates an existing post via the backend API
func (s *PostService) UpdatePost(postID string, categoryNames []string, content string, sessionCookie *http.Cookie) error {
	// Prepare request data
	requestData := map[string]interface{}{
		"category_names": categoryNames,
		"content":        content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Build URL for update post
	updateURL := s.BaseURL + "/posts/edit/" + postID

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
		return fmt.Errorf("failed to update post: %w", err)
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
		return fmt.Errorf("forbidden: you can only edit your own posts")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("post not found")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return fmt.Errorf("update post failed with status %d: %s", resp.StatusCode, string(body))
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

// DeletePost deletes a post via the backend API
func (s *PostService) DeletePost(postID string, sessionCookie *http.Cookie) error {
	// Build URL for delete post
	deleteURL := s.BaseURL + "/posts/remove/" + postID

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
		return fmt.Errorf("failed to delete post: %w", err)
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
		return fmt.Errorf("forbidden: you can only delete your own posts")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("post not found")
	}
	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var apiResponse models.APIResponse
		if json.Unmarshal(body, &apiResponse) == nil && !apiResponse.Success {
			return fmt.Errorf("API error: %s", apiResponse.Error)
		}
		return fmt.Errorf("delete post failed with status %d: %s", resp.StatusCode, string(body))
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
