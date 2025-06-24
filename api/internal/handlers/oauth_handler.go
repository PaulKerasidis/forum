package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// In-memory store for OAuth states (in production, use Redis or database)
var (
	oauthStates = make(map[string]*models.OAuthState)
	statesMutex = sync.RWMutex{}
)

// CleanupExpiredStates removes expired OAuth states
func CleanupExpiredStates() {
	statesMutex.Lock()
	defer statesMutex.Unlock()
	
	now := time.Now()
	for state, stateData := range oauthStates {
		if now.After(stateData.ExpiresAt) {
			delete(oauthStates, state)
		}
	}
}

// GoogleLoginHandler initiates Google OAuth login
func GoogleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Clean up expired states
		go CleanupExpiredStates()

		// Generate secure state
		state, err := utils.GenerateOAuthState()
		if err != nil {
			log.Printf("Failed to generate OAuth state: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to initiate OAuth")
			return
		}

		// Store state with expiration (10 minutes)
		statesMutex.Lock()
		oauthStates[state] = &models.OAuthState{
			State:     state,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		statesMutex.Unlock()

		// Generate Google auth URL
		authURL := utils.GenerateGoogleAuthURL(state)

		// Return JSON response with the auth URL
		response := map[string]string{
			"auth_url": authURL,
			"state":    state,
		}

		utils.RespondWithSuccess(w, http.StatusOK, response)
	}
}

// GoogleCallbackHandler handles Google OAuth callback
func GoogleCallbackHandler(ur *repository.UserRepository, sr *repository.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get query parameters
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		// Check for OAuth errors
		if errorParam != "" {
			log.Printf("OAuth error: %s", errorParam)
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_cancelled", http.StatusTemporaryRedirect)
			return
		}

		// Validate required parameters
		if code == "" {
			log.Printf("OAuth error: missing authorization code")
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_invalid", http.StatusTemporaryRedirect)
			return
		}

		if state == "" {
			log.Printf("OAuth error: missing state parameter")
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_invalid", http.StatusTemporaryRedirect)
			return
		}

		// Validate state
		statesMutex.RLock()
		storedState, exists := oauthStates[state]
		statesMutex.RUnlock()

		if !exists || time.Now().After(storedState.ExpiresAt) {
			log.Printf("OAuth error: invalid or expired state")
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_expired", http.StatusTemporaryRedirect)
			return
		}

		// Clean up used state
		statesMutex.Lock()
		delete(oauthStates, state)
		statesMutex.Unlock()

		// Exchange code for token
		tokenResp, err := utils.ExchangeGoogleCode(code)
		if err != nil {
			log.Printf("Failed to exchange code for token: %v", err)
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_token_failed", http.StatusTemporaryRedirect)
			return
		}

		// Get user info from Google
		userInfo, err := utils.GetGoogleUserInfo(tokenResp.AccessToken)
		if err != nil {
			log.Printf("Failed to get user info: %v", err)
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_userinfo_failed", http.StatusTemporaryRedirect)
			return
		}

		// Check if user exists or create new user
		user, isNewUser, err := ur.CreateOrGetOAuthUser(userInfo)
		if err != nil {
			log.Printf("Failed to create/get OAuth user: %v", err)
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_user_failed", http.StatusTemporaryRedirect)
			return
		}

		// Create session
		session, err := sr.CreateSession(user.ID, r.RemoteAddr)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Redirect(w, r, "http://localhost:3000/login?error=oauth_session_failed", http.StatusTemporaryRedirect)
			return
		}

		log.Printf("OAuth session created successfully: UserID=%s, SessionID=%s", user.ID, session.SessionID)

		// Don't set session cookie here, let frontend handle it
		// utils.SetSessionCookie(session.SessionID, w, r, session.ExpiresAt)

		// Redirect to frontend with success and session ID
		var redirectURL string
		if isNewUser {
			redirectURL = fmt.Sprintf("http://localhost:3000/?oauth=success&new_user=true&session_id=%s", session.SessionID)
		} else {
			redirectURL = fmt.Sprintf("http://localhost:3000/?oauth=success&session_id=%s", session.SessionID)
		}

		log.Printf("OAuth success, redirecting to: %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// GoogleLoginStatusHandler returns OAuth login status for frontend
func GoogleLoginStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		response := map[string]interface{}{
			"oauth_enabled": true,
			"providers":     []string{"google"},
		}

		utils.RespondWithSuccess(w, http.StatusOK, response)
	}
}