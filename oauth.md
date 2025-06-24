# Google OAuth Implementation Guide

This document explains how Google OAuth 2.0 authentication was implemented in the Forum application using only Go standard library (no external OAuth libraries).

## Table of Contents

1. [Overview](#overview)
2. [Google Cloud Console Setup](#google-cloud-console-setup)
3. [Backend Implementation](#backend-implementation)
4. [Frontend Implementation](#frontend-implementation)
5. [Database Schema](#database-schema)
6. [OAuth Flow Explanation](#oauth-flow-explanation)
7. [Security Considerations](#security-considerations)
8. [Troubleshooting](#troubleshooting)

## Overview

The OAuth implementation allows users to register and login using their Google accounts instead of creating traditional username/password accounts. The system supports both regular users and OAuth users seamlessly.

### Key Features
- ✅ Google OAuth 2.0 integration using only Go standard library
- ✅ Secure state parameter validation
- ✅ Automatic user creation for new OAuth users
- ✅ Session management compatible with existing auth system
- ✅ Cross-port cookie handling (API: 8080, Frontend: 3000)
- ✅ IP normalization for localhost (IPv4/IPv6 compatibility)

## Google Cloud Console Setup

### Step 1: Create a Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Give it a name like "Forum OAuth"

### Step 2: Enable Required APIs
1. Navigate to **APIs & Services** → **Library**
2. Search for and enable **Google+ API**
3. Also enable **People API** as backup

### Step 3: Configure OAuth Consent Screen
1. Go to **APIs & Services** → **OAuth consent screen**
2. Fill in required fields:
   - **App name**: `Forum404NotFound`
   - **User support email**: Your email
   - **Developer contact email**: Your email
3. Add scopes:
   - `../auth/userinfo.email`
   - `../auth/userinfo.profile`
   - `openid`

### Step 4: Create OAuth Client ID
1. Go to **APIs & Services** → **Credentials**
2. Click **+ CREATE CREDENTIALS** → **OAuth client ID**
3. Select **Web application**
4. Configure URIs:
   - **Authorized JavaScript origins**: `http://localhost:3000`, `http://localhost:8080`
   - **Authorized redirect URIs**: `http://localhost:8080/auth/google/callback`

### Step 5: Configure Environment
Add credentials to `api/.env`:
```env
GOOGLE_OAUTH_CLIENT_ID=your_client_id_here
GOOGLE_OAUTH_CLIENT_SECRET=your_client_secret_here
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

## Backend Implementation

### 1. Configuration Setup

**File**: `api/config/config.go`

Added OAuth configuration fields to `AppConfig` struct:
```go
type AppConfig struct {
    // ... existing fields ...
    
    // OAuth configuration
    GoogleOAuthClientID     string
    GoogleOAuthClientSecret string
    GoogleOAuthRedirectURL  string
}
```

Configuration loading in `LoadConfig()`:
```go
// OAuth configuration
Config.GoogleOAuthClientID = getEnv("GOOGLE_OAUTH_CLIENT_ID", "")
Config.GoogleOAuthClientSecret = getEnv("GOOGLE_OAUTH_CLIENT_SECRET", "")
Config.GoogleOAuthRedirectURL = getEnv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback")
```

### 2. OAuth Models

**File**: `api/internal/models/oauth_models.go`

Created models for OAuth data handling:
```go
// OAuth state for CSRF protection
type OAuthState struct {
    State     string    `json:"state"`
    Nonce     string    `json:"nonce"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
}

// Google user information
type GoogleUserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    // ... more fields
}

// OAuth login response
type OAuthLoginResponse struct {
    User      User   `json:"user"`
    SessionID string `json:"session_id"`
    IsNewUser bool   `json:"is_new_user"`
}
```

Updated `User` model to support OAuth:
```go
type User struct {
    ID            string    `json:"id"`
    Username      string    `json:"username"`
    Email         string    `json:"email"`
    Provider      string    `json:"provider,omitempty"`      // oauth provider
    ProviderID    string    `json:"provider_id,omitempty"`   // user id from provider
    ProviderEmail string    `json:"provider_email,omitempty"` // email from provider
    CreatedAt     time.Time `json:"created_at"`
}
```

### 3. OAuth Utilities

**File**: `api/internal/utils/oauth.go`

Implemented OAuth flow utilities using only Go standard library:

#### State Generation (CSRF Protection)
```go
func GenerateOAuthState() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

#### Google Auth URL Generation
```go
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
```

#### Token Exchange
```go
func ExchangeGoogleCode(code string) (*models.GoogleTokenResponse, error) {
    data := url.Values{}
    data.Set("client_id", config.Config.GoogleOAuthClientID)
    data.Set("client_secret", config.Config.GoogleOAuthClientSecret)
    data.Set("code", code)
    data.Set("grant_type", "authorization_code")
    data.Set("redirect_uri", config.Config.GoogleOAuthRedirectURL)

    req, err := http.NewRequest("POST", googleTokenURL, strings.NewReader(data.Encode()))
    // ... HTTP request handling
}
```

#### User Info Retrieval
```go
func GetGoogleUserInfo(accessToken string) (*models.GoogleUserInfo, error) {
    req, err := http.NewRequest("GET", googleUserURL, nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)
    // ... HTTP request handling
}
```

### 4. OAuth Handlers

**File**: `api/internal/handlers/oauth_handler.go`

#### Login Initiation Handler
```go
func GoogleLoginHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Generate secure state
        state, err := utils.GenerateOAuthState()
        
        // Store state with expiration (10 minutes)
        oauthStates[state] = &models.OAuthState{
            State:     state,
            CreatedAt: time.Now(),
            ExpiresAt: time.Now().Add(10 * time.Minute),
        }

        // Generate Google auth URL
        authURL := utils.GenerateGoogleAuthURL(state)

        // Return JSON response with auth URL
        response := map[string]string{
            "auth_url": authURL,
            "state":    state,
        }
        utils.RespondWithSuccess(w, http.StatusOK, response)
    }
}
```

#### Callback Handler
```go
func GoogleCallbackHandler(ur *repository.UserRepository, sr *repository.SessionRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get and validate parameters
        code := r.URL.Query().Get("code")
        state := r.URL.Query().Get("state")
        
        // Validate state (CSRF protection)
        storedState, exists := oauthStates[state]
        if !exists || time.Now().After(storedState.ExpiresAt) {
            // Redirect with error
            return
        }

        // Exchange code for token
        tokenResp, err := utils.ExchangeGoogleCode(code)
        
        // Get user info from Google
        userInfo, err := utils.GetGoogleUserInfo(tokenResp.AccessToken)
        
        // Create or get user
        user, isNewUser, err := ur.CreateOrGetOAuthUser(userInfo)
        
        // Create session
        session, err := sr.CreateSession(user.ID, r.RemoteAddr)
        
        // Redirect to frontend with session ID
        redirectURL := fmt.Sprintf("http://localhost:3000/?oauth=success&session_id=%s", session.SessionID)
        http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
    }
}
```

### 5. User Repository Updates

**File**: `api/internal/repository/user_repository.go`

Added OAuth user handling:
```go
func (ur *UserRepository) CreateOrGetOAuthUser(googleUser *models.GoogleUserInfo) (*models.User, bool, error) {
    return utils.ExecuteInTransactionWithResult(ur.DB, func(tx *sql.Tx) (OAuthUserResult, error) {
        // Try to find existing user by OAuth provider and provider ID
        var existingUser models.User
        err := tx.QueryRow(`
            SELECT user_id, username, email, provider, provider_id, provider_email, created_at 
            FROM users 
            WHERE provider = 'google' AND provider_id = ?
        `, googleUser.ID).Scan(...)
        
        if err == nil {
            // User exists, return it
            return OAuthUserResult{User: &existingUser, IsNewUser: false}, nil
        }
        
        // Check if email exists for regular account
        // Generate unique username
        // Create new OAuth user
        // Return new user
    })
}
```

### 6. Routes Configuration

**File**: `api/internal/routes/routes.go`

Added OAuth routes with both API and direct patterns:
```go
// API routes (for frontend AJAX calls)
mux.Handle("/api/auth/google/login", http.HandlerFunc(handlers.GoogleLoginHandler()))
mux.Handle("/api/auth/google/callback", http.HandlerFunc(handlers.GoogleCallbackHandler(UserRepo, SessionRepo)))
mux.Handle("/api/auth/oauth/status", http.HandlerFunc(handlers.GoogleLoginStatusHandler()))

// Direct routes (for Google callbacks)
mux.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler())
mux.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler(UserRepo, SessionRepo))
```

### 7. Security Middleware Fix

**File**: `api/internal/middleware/middleware.go`

Fixed IP normalization to handle localhost properly:
```go
func cleanIP(ipWithPossiblePort string) string {
    host, _, err := net.SplitHostPort(ipWithPossiblePort)
    if err != nil {
        host = ipWithPossiblePort
    }
    
    // Normalize localhost - treat IPv4 and IPv6 localhost as the same
    if host == "127.0.0.1" || host == "::1" || host == "localhost" {
        return "localhost"
    }
    
    return host
}
```

## Frontend Implementation

### 1. OAuth Handler

**File**: `frontend/internal/handlers/oauth_handler.go`

Created OAuth handler for frontend:
```go
func (h *OAuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
    // Call backend to get Google auth URL
    authURL, err := h.authService.InitiateGoogleOAuth()
    
    // Redirect to Google OAuth
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
    // Get query parameters
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Forward to backend for processing
    user, sessionID, isNewUser, err := h.authService.HandleGoogleCallback(code, state)
    
    // Set session cookie and redirect
    http.Redirect(w, r, "/?oauth=success", http.StatusSeeOther)
}
```

### 2. Auth Service Updates

**File**: `frontend/internal/services/auth_service.go`

Added OAuth methods:
```go
func (s *AuthService) InitiateGoogleOAuth() (string, error) {
    loginURL := s.BaseURL + "/auth/google/login"
    resp, err := s.HTTPClient.Get(loginURL)
    // ... parse response and return auth URL
}

func (s *AuthService) HandleGoogleCallback(code, state string) (*models.User, string, bool, error) {
    callbackURL := fmt.Sprintf("%s/auth/google/callback?code=%s&state=%s", s.BaseURL, code, state)
    resp, err := s.HTTPClient.Get(callbackURL)
    // ... parse response and return user info
}
```

### 3. Template Updates

**File**: `frontend/web/templates/login.html`

Added Google OAuth button:
```html
<!-- OAuth Divider -->
<div class="oauth-divider">
    <span>OR</span>
</div>

<!-- OAuth Login -->
<div class="oauth-container">
    <a href="/auth/google/login" class="oauth-btn google-btn">
        <i class="fab fa-google"></i>
        Continue with Google
    </a>
</div>
```

### 4. Home Handler Session Handling

**File**: `frontend/internal/handlers/home_handler.go`

Added OAuth session handling:
```go
func (h *HomeHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
    // Handle OAuth success - set session cookie if session_id is provided
    if sessionID := r.URL.Query().Get("session_id"); sessionID != "" {
        log.Printf("OAuth callback detected! Setting session cookie: %s", sessionID)
        expiresAt := time.Now().Add(24 * time.Hour)
        utils.SetSessionCookie("forum_session", sessionID, w, r, expiresAt)
        
        // Redirect to clean URL
        http.Redirect(w, r, "/?oauth=success", http.StatusTemporaryRedirect)
        return
    }
    
    // ... rest of home handler logic
}
```

## Database Schema

### Updated Users Table

**File**: `api/database/sql_statements.go`

Updated users table to support OAuth:
```sql
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT, -- Made optional for OAuth users
    provider TEXT, -- OAuth provider (google, github, etc.)
    provider_id TEXT, -- User ID from OAuth provider
    provider_email TEXT, -- Email from OAuth provider
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Added Indexes

```sql
-- OAuth indexes
CREATE INDEX IF NOT EXISTS idx_users_provider_id ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_provider_email ON users(provider, provider_email);
```

## OAuth Flow Explanation

### Complete OAuth 2.0 Flow

```
1. User clicks "Continue with Google"
   ↓
2. Frontend → GET /auth/google/login
   ↓
3. API generates state, stores it, returns Google auth URL
   ↓
4. Frontend redirects user → Google OAuth consent screen
   ↓
5. User authorizes application
   ↓
6. Google redirects → /auth/google/callback?code=ABC&state=XYZ
   ↓
7. API validates state, exchanges code for token
   ↓
8. API gets user info from Google
   ↓
9. API creates or finds user in database
   ↓
10. API creates session
    ↓
11. API redirects → Frontend with session_id
    ↓
12. Frontend sets session cookie and redirects to clean URL
    ↓
13. User is logged in
```

### Key Security Measures

1. **State Parameter**: Prevents CSRF attacks
2. **State Expiration**: States expire after 10 minutes
3. **Session Validation**: Sessions are validated on each request
4. **IP Normalization**: Handles localhost IPv4/IPv6 properly
5. **HTTPS in Production**: Ensure secure cookies in production

## Security Considerations

### 1. State Management
- States are generated using cryptographically secure random numbers
- States expire after 10 minutes to prevent replay attacks
- States are deleted after use (one-time use)

### 2. Session Security
- Sessions are created with proper IP tracking
- IP changes are monitored for security
- Sessions expire after 24 hours by default

### 3. OAuth Token Handling
- Access tokens are never stored in database
- Tokens are only used to fetch user information once
- No refresh tokens are stored (stateless approach)

### 4. Cross-Origin Considerations
- CORS is properly configured for localhost development
- Cookie domain is set to allow cross-port sharing
- SameSite policy is set to Lax for OAuth flows

## Troubleshooting

### Common Issues

#### 1. "Missing required parameter: client_id"
**Cause**: OAuth credentials not loaded
**Solution**: 
- Ensure `.env` file exists in `api/` directory (not just `.env.example`)
- Check that `GOOGLE_OAUTH_CLIENT_ID` is set correctly
- Restart API server after updating `.env`

#### 2. "404 page not found" on callback
**Cause**: OAuth callback route not registered
**Solution**:
- Ensure both `/api/auth/google/callback` and `/auth/google/callback` routes are registered
- Verify Google Console redirect URI matches exactly: `http://localhost:8080/auth/google/callback`

#### 3. "Suspicious IP change detected"
**Cause**: IPv4/IPv6 localhost mismatch
**Solution**:
- IP normalization fix implemented in middleware
- Both `127.0.0.1` and `::1` are treated as `localhost`

#### 4. Not logged in after OAuth success
**Cause**: Session cookie not set properly
**Solution**:
- Check that session ID is passed in URL
- Verify cookie domain settings
- Ensure frontend handles session ID parameter

### Debug Tips

1. **Check API logs** for OAuth flow progress
2. **Check frontend logs** for session handling
3. **Inspect browser cookies** for session cookie
4. **Verify Google Console configuration** matches code
5. **Test individual endpoints** with curl/Postman

## Production Deployment

### Environment Variables
Update for production:
```env
GOOGLE_OAUTH_REDIRECT_URL=https://yourdomain.com/auth/google/callback
ALLOWED_ORIGINS=https://yourdomain.com
ENVIRONMENT=production
```

### Google Console Updates
- Update authorized domains to production domain
- Update redirect URIs to production URLs
- Consider app verification for public release

### Security Enhancements
- Enable HTTPS everywhere
- Set secure cookie flags
- Implement proper rate limiting
- Add comprehensive logging
- Consider storing OAuth states in Redis instead of memory

---

This implementation provides a complete, secure OAuth 2.0 integration using only Go standard library, demonstrating proper security practices and error handling throughout the authentication flow.