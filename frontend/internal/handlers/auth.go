package handlers

import (
	"log"
	"net/http"
	"time"

	"frontend-service/config"
	"frontend-service/internal/models"
	"frontend-service/internal/services"
	"frontend-service/internal/session"
	"frontend-service/internal/utils"
	"frontend-service/internal/validations"
)

type AuthHandler struct {
	authService     *services.AuthService
	templateService *services.TemplateService
	config          *config.Config // ADDED: Store config for session cookie name
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, templateService *services.TemplateService, cfg *config.Config) *AuthHandler { // CHANGED: Accept config
	return &AuthHandler{
		authService:     authService,
		templateService: templateService,
		config:          cfg, // ADDED: Store config
	}
}

// ServeRegister handles GET and POST requests for registration
func (h *AuthHandler) ServeRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showRegisterForm(w, r)
	case http.MethodPost:
		h.handleRegisterForm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ServeLogin handles GET and POST requests for login
func (h *AuthHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showLoginForm(w, r)
	case http.MethodPost:
		h.handleLoginForm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ServeLogout handles logout requests
func (h *AuthHandler) ServeLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Logout request received")

	// Get session cookie using config value
	cookie, err := session.GetSessionCookie(r, h.authService) // CHANGED: Use utility function with config
	if err != nil {
		log.Printf("No session cookie found during logout")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("Found session cookie, calling backend logout")

	// Call backend to logout
	if err := h.authService.LogoutUser(cookie.Value); err != nil {
		log.Printf("Error during backend logout: %v", err)
	} else {
		log.Printf("Backend logout successful")
	}

	// Clear the session cookie on frontend using config value
	utils.ClearSessionCookie(h.config.SessionName, w) // CHANGED: Use config session name

	log.Printf("Session cookie cleared, redirecting to home")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// showRegisterForm displays the registration form (GET request)
func (h *AuthHandler) showRegisterForm(w http.ResponseWriter, _ *http.Request) {
	data := models.RegisterPageData{
		FormData: &models.UserRegistration{}, // Empty form data for initial load
	}

	if err := h.templateService.Render(w, "register.html", data); err != nil {
		log.Printf("Error rendering register template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// handleRegisterForm processes the registration form submission (POST request)
func (h *AuthHandler) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		h.showRegisterError(w, "Invalid form data", &models.UserRegistration{})
		return
	}

	// Get form values - ADD confirm_password here
	formData := models.UserRegistration{
		Username:        r.FormValue("username"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"), // NEW: Add this line
	}

	// Basic validation (check if all fields are provided)
	if formData.Username == "" || formData.Email == "" || formData.Password == "" {
		h.showRegisterError(w, "All fields are required", &formData)
		return
	}

	// NEW: Check password confirmation
	if formData.ConfirmPassword == "" {
		h.showRegisterError(w, "Password confirmation is required", &formData)
		return
	}

	// NEW: Validate passwords match (frontend validation)
	if formData.Password != formData.ConfirmPassword {
		h.showRegisterError(w, "Passwords do not match", &formData)
		return
	}

	// FRONTEND VALIDATION using your validation functions
	if err := validations.ValidateUserInput(formData.Username, formData.Email, formData.Password); err != nil {
		h.showRegisterError(w, err.Error(), &formData)
		return
	}

	// Call backend API to register user (backend will also validate)
	if err := h.authService.RegisterUser(formData); err != nil {
		log.Printf("Registration error: %v", err)
		h.showRegisterError(w, err.Error(), &formData)
		return
	}

	// Registration successful - show success message with empty form
	data := models.RegisterPageData{
		Success:  "Registration successful! You can now login.",
		FormData: &models.UserRegistration{}, // Clear form on success
	}

	if err := h.templateService.Render(w, "register.html", data); err != nil {
		log.Printf("Error rendering register template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// showRegisterError displays registration form with error message AND preserved form data
func (h *AuthHandler) showRegisterError(w http.ResponseWriter, errorMsg string, formData *models.UserRegistration) {
	// Clear password for security - user will need to retype it
	if formData != nil {
		formData.Password = ""
	}

	data := models.RegisterPageData{
		Error:    errorMsg,
		FormData: formData, // Pass back the form data so fields stay populated
	}

	if err := h.templateService.Render(w, "register.html", data); err != nil {
		log.Printf("Error rendering register template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// showLoginForm displays the login form (GET request)
func (h *AuthHandler) showLoginForm(w http.ResponseWriter, _ *http.Request) {
	data := models.LoginPageData{
		FormData: &models.UserLogin{}, // Empty form data for initial load
	}

	if err := h.templateService.Render(w, "login.html", data); err != nil {
		log.Printf("Error rendering login template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// handleLoginForm processes the login form submission (POST request)
func (h *AuthHandler) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		h.showLoginError(w, "Invalid form data", &models.UserLogin{})
		return
	}

	// Get form values
	formData := models.UserLogin{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	// Basic validation (check if all fields are provided)
	if formData.Email == "" || formData.Password == "" {
		h.showLoginError(w, "Email and password are required", &formData)
		return
	}

	// FRONTEND VALIDATION using your validation functions
	err := validations.ValidateEmail(formData.Email)
	if err != nil {
		h.showLoginError(w, err.Error(), &formData)
		return
	}
	err = validations.ValidatePassword(formData.Password)
	if err != nil {
		h.showLoginError(w, err.Error(), &formData)
		return
	}

	// Call backend API to login user
	user, sessionID, err := h.authService.LoginUser(formData)
	if err != nil {
		log.Printf("Login error: %v", err)
		h.showLoginError(w, err.Error(), &formData)
		return
	}

	// Set session cookie using config value and utility function
	expiresAt := time.Now().Add(24 * time.Hour)
	utils.SetSessionCookie(h.config.SessionName, sessionID, w, r, expiresAt) // CHANGED: Use utility with config session name

	// Login successful - redirect to home page
	log.Printf("User %s logged in successfully", user.Username)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// showLoginError displays login form with error message AND preserved form data
func (h *AuthHandler) showLoginError(w http.ResponseWriter, errorMsg string, formData *models.UserLogin) {
	// Clear password for security - user will need to retype it
	if formData != nil {
		formData.Password = ""
	}

	data := models.LoginPageData{
		Error:    errorMsg,
		FormData: formData, // Pass back the form data so fields stay populated
	}

	if err := h.templateService.Render(w, "login.html", data); err != nil {
		log.Printf("Error rendering login template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
