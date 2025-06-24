package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/internal/models"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GenerateOAuthState generates a secure random state for OAuth flow
func GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateGoogleAuthURL generates the Google OAuth authorization URL
func GenerateGoogleAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", config.Config.GoogleOAuthClientID)
	params.Set("redirect_uri", config.Config.GoogleOAuthRedirectURL)
	params.Set("scope", "openid email profile")
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")

	return googleAuthURL + "?" + params.Encode()
}

// ExchangeGoogleCode exchanges authorization code for access token
func ExchangeGoogleCode(code string) (*models.GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", config.Config.GoogleOAuthClientID)
	data.Set("client_secret", config.Config.GoogleOAuthClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.Config.GoogleOAuthRedirectURL)

	req, err := http.NewRequest("POST", googleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp models.GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	return &tokenResp, nil
}

// GetGoogleUserInfo fetches user information from Google using access token
func GetGoogleUserInfo(accessToken string) (*models.GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", googleUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo models.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	return &userInfo, nil
}

// ValidateOAuthState validates the OAuth state parameter
func ValidateOAuthState(receivedState, expectedState string) error {
	if receivedState == "" {
		return errors.New("missing state parameter")
	}
	if receivedState != expectedState {
		return errors.New("invalid state parameter")
	}
	return nil
}

// GenerateUsernameFromEmail generates a username from email for OAuth users
func GenerateUsernameFromEmail(email string) string {
	// Extract the part before @ and clean it
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return "user"
	}
	
	username := parts[0]
	// Remove any non-alphanumeric characters except underscore
	var cleaned strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}
	
	result := cleaned.String()
	if result == "" {
		return "user"
	}
	
	// Ensure it's within length limits
	if len(result) > 15 {
		result = result[:15]
	}
	if len(result) < 5 {
		result = result + "user"
		if len(result) > 15 {
			result = result[:15]
		}
	}
	
	return result
}

// HashEmail creates a hash of the email for consistent user identification
func HashEmail(email string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(email)))
	return base64.URLEncoding.EncodeToString(hash[:])[:12] // Use first 12 chars for readability
}