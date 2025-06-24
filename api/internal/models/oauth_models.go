package models

import "time"

// OAuthState represents the state parameter for OAuth flow
type OAuthState struct {
	State     string    `json:"state"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GoogleUserInfo represents the user info returned by Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GoogleTokenResponse represents the token response from Google
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope"`
}

// OAuthUser represents a user from OAuth provider
type OAuthUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Provider      string `json:"provider"`
	ProviderID    string `json:"provider_id"`
	ProviderEmail string `json:"provider_email"`
	CreatedAt     time.Time `json:"created_at"`
}

// OAuthLoginResponse represents the response after successful OAuth login
type OAuthLoginResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
	IsNewUser bool   `json:"is_new_user"`
}