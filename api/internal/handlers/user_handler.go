package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PaulKerasidis/forum/config"
	"github.com/PaulKerasidis/forum/internal/middleware"
	"github.com/PaulKerasidis/forum/internal/models"
	"github.com/PaulKerasidis/forum/internal/repository"
	"github.com/PaulKerasidis/forum/internal/utils"
)

// Handle user registration logic here
func RegisterHandler(ur *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var reg models.UserRegistration
		err := json.NewDecoder(r.Body).Decode(&reg)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate request - basic required fields check
		if reg.Email == "" || reg.Username == "" || reg.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email, username and password are required")
			return
		}

		// NEW: Check if ConfirmPassword field exists and validate
		if reg.ConfirmPassword == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Password confirmation is required")
			return
		}

		// NEW: Validate that passwords match
		if reg.Password != reg.ConfirmPassword {
			utils.RespondWithError(w, http.StatusBadRequest, "Passwords do not match")
			return
		}

		err = utils.ValidateUserInput(reg.Username, reg.Email, reg.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := ur.CreateUser(reg)
		if err != nil {
			if err.Error() == "email already taken" {
				utils.RespondWithError(w, http.StatusConflict, err.Error())
			} else if err.Error() == "username already taken" {
				utils.RespondWithError(w, http.StatusConflict, err.Error())
			} else {
				utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, user)
	}
}

// LoginHandler handles user login
func LoginHandler(ur *repository.UserRepository, sr *repository.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, errors.New("method not allowed").Error())
			return
		}

		// Parse request body
		var login models.UserLogin
		err := json.NewDecoder(r.Body).Decode(&login)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, errors.New("invalid request payload").Error())
			return
		}

		// Validate request - basic required fields check
		if login.Email == "" || login.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, errors.New("email and password are required").Error())
			return
		}

		err = utils.ValidateEmail(login.Email)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = utils.ValidatePassword(login.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Authenticate user
		user, err := ur.Authenticate(login)
		if err != nil {
			switch err {
			case errors.New("invalid credentials"), errors.New("email not found"):
				utils.RespondWithError(w, http.StatusUnauthorized, errors.New("invalid credentials").Error())
			default:
				utils.RespondWithError(w, http.StatusInternalServerError, errors.New("authentication failed").Error())
			}
			return
		}

		// Create a new session
		session, err := sr.CreateSession(user.ID, r.RemoteAddr)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, errors.New("failed to create session").Error())
			return
		}

		utils.SetSessionCookie(session.SessionID, w, r, session.ExpiresAt) // CHANGED: Use simplified call

		// Return JSON response
		response := models.LoginResponse{
			User:      *user,
			SessionID: session.SessionID,
		}
		utils.RespondWithSuccess(w, http.StatusOK, response)
	}
}

// LogoutHandler handles user logout
func LogoutHandler(ur *repository.UserRepository, sr *repository.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, errors.New("method not allowed").Error())
			return
		}

		// Get the session cookie using config session name
		cookie, err := r.Cookie(config.Config.SessionName) // CHANGED: Use config session name
		if err != nil {
			// this will never happen because of requireAuth function in middleware
			// sending error response in case authentication fails
			utils.RespondWithError(w, http.StatusUnauthorized, errors.New("unauthorized access").Error())
			return
		}

		// Delete the session
		err = sr.DeleteSession(cookie.Value)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, errors.New("failed to logout").Error())
			return
		}

		// Clear the cookie
		utils.ClearSessionCookie(w) // CHANGED: Use simplified call

		utils.RespondWithSuccess(w, http.StatusOK, nil)
	}
}

func GetCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow only post method
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, errors.New("method not allowed").Error())
			return
		}

		// The user will already be in the context thanks to the RequireAuth middleware
		user := middleware.GetCurrentUser(r)
		if user == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, errors.New("unauthorized access").Error())
			return
		}

		// Return the user data
		utils.RespondWithSuccess(w, http.StatusOK, user)
	}
}

func GetUserProfileHandler(ur *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get authenticated user (since only users can view their own profile)
		currentUser := middleware.GetCurrentUser(r)
		if currentUser == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Extract user ID from URL path
		userID := r.PathValue("id")
		// Ensure user can only view their own profile
		if currentUser.ID != userID {
			utils.RespondWithError(w, http.StatusForbidden, "You can only view your own profile")
			return
		}

		// Get user profile with statistics
		profile, err := ur.GetUserProfile(userID)
		if err != nil {
			if err.Error() == "user not found" {
				utils.RespondWithError(w, http.StatusNotFound, "User not found")
			} else {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user profile")
			}
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, profile)
	}
}
