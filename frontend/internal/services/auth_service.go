package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"frontend-service/config"
	"frontend-service/internal/models"
	"frontend-service/internal/validations"
)

type AuthService struct {
	*BaseClient
	SessionName string // Store session cookie name from config
}

// NewAuthService creates a new auth service
func NewAuthService(baseClient *BaseClient, cfg *config.Config) *AuthService { // Accept config
	return &AuthService{
		BaseClient:  baseClient,
		SessionName: cfg.SessionName, // Store session name from config
	}
}

// RegisterUser registers a new user via the backend API
func (s *AuthService) RegisterUser(formData models.RegisterFormData) error {
	err := validations.ValidateUserInput(formData.Username, formData.Email, formData.Password)
	if err != nil {
		return fmt.Errorf("invalid user input: %w", err)
	}

	// Convert form data to JSON
	jsonData, err := json.Marshal(formData)
	if err != nil {
		return fmt.Errorf("failed to marshal form data: %w", err)
	}

	//api from URL
	registerURL := s.BaseURL + "/auth/register"

	// Make HTTP POST request
	resp, err := s.HTTPClient.Post(registerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response to get better error messages
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err == nil {
		if !apiResponse.Success {
			return fmt.Errorf("%s", apiResponse.Error)
		}
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LoginUser logs in a user via the backend API
func (s *AuthService) LoginUser(formData models.LoginFormData) (*models.User, string, error) {
	err := validations.ValidateEmail(formData.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email format: %w", err)
	}
	err = validations.ValidatePassword(formData.Password)
	if err != nil {
		return nil, "", fmt.Errorf("invalid password format: %w", err)
	}
	// Convert form data to JSON
	jsonData, err := json.Marshal(formData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal form data: %w", err)
	}

	// FIXED: Remove duplicate /api from URL
	loginURL := s.BaseURL + "/auth/login"

	// Make HTTP POST request
	resp, err := s.HTTPClient.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to login user: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response first to get better error messages
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return nil, "", fmt.Errorf("%s", apiResponse.Error)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Convert data to LoginResponse (which contains User and SessionID)
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal data: %w", err)
	}

	var loginResponse struct {
		User      models.User `json:"user"`
		SessionID string      `json:"session_id"`
	}
	if err := json.Unmarshal(dataBytes, &loginResponse); err != nil {
		return nil, "", fmt.Errorf("failed to parse login data: %w", err)
	}

	return &loginResponse.User, loginResponse.SessionID, nil
}

// LogoutUser logs out a user via the backend API
func (s *AuthService) LogoutUser(sessionID string) error {
	// FIXED: Remove duplicate /api from URL
	logoutURL := s.BaseURL + "/auth/logout"

	// Create request with session cookie
	req, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie to request using config value
	req.AddCookie(&http.Cookie{
		Name:  s.SessionName, // CHANGED: Use config value instead of hardcoded "session_id"
		Value: sessionID,
	})

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logout user: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ValidateSession validates a session ID with the backend API
func (s *AuthService) ValidateSession(sessionID string) (*models.User, error) {
	// FIXED: Remove duplicate /api from URL
	validateURL := s.BaseURL + "/auth/me"

	// Create request with session cookie
	req, err := http.NewRequest("POST", validateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add session cookie to request using config value
	req.AddCookie(&http.Cookie{
		Name:  s.SessionName, // CHANGED: Use config value instead of hardcoded "session_id"
		Value: sessionID,
	})

	// Make HTTP request
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors (401 means invalid session)
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid session")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("session validation failed: %s", string(body))
	}

	// Parse JSON response
	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check API success
	if !apiResponse.Success {
		return nil, fmt.Errorf("session validation failed: %s", apiResponse.Error)
	}

	// Convert data to User
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var user models.User
	if err := json.Unmarshal(dataBytes, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

// InitiateGoogleOAuth initiates Google OAuth flow via backend
func (s *AuthService) InitiateGoogleOAuth() (string, error) {
	loginURL := s.BaseURL + "/auth/google/login"

	resp, err := s.HTTPClient.Get(loginURL)
	if err != nil {
		return "", fmt.Errorf("failed to initiate Google OAuth: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OAuth initiation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !apiResponse.Success {
		return "", fmt.Errorf("%s", apiResponse.Error)
	}

	// Extract auth_url from response
	dataMap, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response data format")
	}

	authURL, ok := dataMap["auth_url"].(string)
	if !ok {
		return "", fmt.Errorf("auth_url not found in response")
	}

	return authURL, nil
}

// HandleGoogleCallback handles Google OAuth callback via backend
func (s *AuthService) HandleGoogleCallback(code, state string) (*models.User, string, bool, error) {
	callbackURL := fmt.Sprintf("%s/auth/google/callback?code=%s&state=%s", s.BaseURL, code, state)

	resp, err := s.HTTPClient.Get(callbackURL)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to handle Google callback: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("OAuth callback failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, "", false, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !apiResponse.Success {
		return nil, "", false, fmt.Errorf("%s", apiResponse.Error)
	}

	// Convert data to OAuthLoginResponse
	dataBytes, err := json.Marshal(apiResponse.Data)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to marshal data: %w", err)
	}

	var oauthResponse struct {
		User      models.User `json:"user"`
		SessionID string      `json:"session_id"`
		IsNewUser bool        `json:"is_new_user"`
	}
	if err := json.Unmarshal(dataBytes, &oauthResponse); err != nil {
		return nil, "", false, fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	return &oauthResponse.User, oauthResponse.SessionID, oauthResponse.IsNewUser, nil
}

// GetOAuthStatus gets OAuth configuration from backend
func (s *AuthService) GetOAuthStatus() (map[string]interface{}, error) {
	statusURL := s.BaseURL + "/auth/oauth/status"

	resp, err := s.HTTPClient.Get(statusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("%s", apiResponse.Error)
	}

	status, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid status data format")
	}

	return status, nil
}
