package handlers

import (
	"encoding/json"
	"net/http"

	"forum/models"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	UserRepo    *models.UserRepository
	SessionRepo *models.SessionRepository
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *models.UserRepository, sessionRepo *models.SessionRepository) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reg models.UserRegistration
	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if reg.Username == "" || reg.Email == "" || reg.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	if len(reg.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Create user
	user, err := h.UserRepo.Create(reg)
	if err != nil {
		switch err {
		case models.ErrEmailTaken:
			http.Error(w, "Email is already taken", http.StatusConflict)
		case models.ErrUsernameTaken:
			http.Error(w, "Username is already taken", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var login models.UserLogin
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if login.Email == "" || login.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.UserRepo.Authenticate(login)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Create a new session
	session, err := h.SessionRepo.Create(user.ID, r.RemoteAddr)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.SessionID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set Secure flag if TLS is enabled
		SameSite: http.SameSiteStrictMode,
	})

	// Return response
	w.Header().Set("Content-Type", "application/json")
	response := models.LoginResponse{
		User:      *user,
		SessionID: session.SessionID,
	}
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// If no cookie, nothing to do
		w.WriteHeader(http.StatusOK)
		return
	}

	// Delete the session
	err = h.SessionRepo.Delete(cookie.Value)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}