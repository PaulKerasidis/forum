package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"frontend-service/internal/models"
)

type UserService struct {
	*BaseClient
}

// NewUserService creates a new user service
func NewUserService(baseClient *BaseClient) *UserService {
	return &UserService{
		BaseClient: baseClient,
	}
}

// GetUserProfile retrieves user profile with stats from the backend API
func (s *UserService) GetUserProfile(userID string, sessionCookie *http.Cookie) (*models.UserProfile, error) {
	// Build URL for user profile
	profileURL := s.BaseURL + "/users/profile/" + userID

	// Create HTTP request
	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Create HTTP request
	req, err = http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication AND reaction data
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Create HTTP request
	req, err = http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication AND reaction data
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
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
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden: you can only view your own profile")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
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

	// Convert data to UserProfile
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var userProfile models.UserProfile
	if err := json.Unmarshal(dataBytes, &userProfile); err != nil {
		return nil, fmt.Errorf("failed to parse user profile data: %w", err)
	}

	return &userProfile, nil
}

// GetUserPosts retrieves posts created by a specific user
func (s *UserService) GetUserPosts(userID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.PaginatedPostsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/users/posts/" + userID)
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

	// Create HTTP request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication AND reaction data
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user posts: %w", err)
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
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden: you can only view your own posts")
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

// GetUserLikedPosts retrieves posts liked by a specific user
func (s *UserService) GetUserLikedPosts(userID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.PaginatedPostsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/users/liked-posts/" + userID)
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

	// Create HTTP request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication AND reaction data
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user liked posts: %w", err)
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
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden: you can only view your own liked posts")
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

	// Convert data to PaginatedPostsResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var postsResponse models.PaginatedPostsResponse
	if err := json.Unmarshal(dataBytes, &postsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse liked posts data: %w", err)
	}

	return &postsResponse, nil
}

// GetUserCommentedPosts retrieves posts that a specific user has commented on
func (s *UserService) GetUserCommentedPosts(userID string, limit, offset int, sortBy string, sessionCookie *http.Cookie) (*models.PaginatedPostsResponse, error) {
	// Build URL with query parameters
	u, err := url.Parse(s.BaseURL + "/users/commented-posts/" + userID)
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

	// Create HTTP request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie for authentication AND reaction data
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user commented posts: %w", err)
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
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden: you can only view your own commented posts")
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

	// Convert data to PaginatedPostsResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var postsResponse models.PaginatedPostsResponse
	if err := json.Unmarshal(dataBytes, &postsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse commented posts data: %w", err)
	}

	return &postsResponse, nil
}
