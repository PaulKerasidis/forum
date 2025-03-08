// Package handlers provides HTTP request handlers and error handling for the application
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
)

// ErrorType represents a basic HTTP error with status code and message
type ErrorType struct {
	Status  int
	Message string
}

// ErrorData combines an ErrorType with additional description for template rendering
type ErrorData struct {
	ErrorType
	Description string
}

// Predefined application errors
var (
	ErrBadRequest = ErrorType{
		Status:  http.StatusBadRequest,
		Message: "Bad Request",
	}
	ErrNotFound = ErrorType{
		Status:  http.StatusNotFound,
		Message: "Page Not Found",
	}
	ErrInternalServer = ErrorType{
		Status:  http.StatusInternalServerError,
		Message: "Internal Server Error",
	}
	ErrInvalidID = ErrorType{
		Status:  http.StatusBadRequest,
		Message: "Invalid ID Format",
	}
)

// ErrorHandler renders the error page template with provided error information
// If template processing fails, falls back to basic HTTP error response
func ErrorHandler(w http.ResponseWriter, errType ErrorType, description string) {
	w.WriteHeader(errType.Status)

	data := ErrorData{
		ErrorType:   errType,
		Description: description,
	}

	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// sendJSONError sends an error response in JSON format
func sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Create error response structure
	errorResponse := struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		Status:  statusCode,
		Message: message,
	}

	// Encode and send the error
	json.NewEncoder(w).Encode(errorResponse)
}
