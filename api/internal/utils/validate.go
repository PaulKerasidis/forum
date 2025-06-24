package utils

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/PaulKerasidis/forum/config"
)

// more secure way to validate email, password, and username inputs
// Used net/mail package for email validation
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func ValidatePassword(password string) error {
	// Password validation using configuration
	if len(password) < config.Config.MinPasswordLen || len(password) > config.Config.MaxPasswordLen {
		return errors.New("password must be between " +
			string(rune(config.Config.MinPasswordLen)) + " and " +
			string(rune(config.Config.MaxPasswordLen)) + " characters")
	}

	// Check password complexity
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	if !(hasUppercase && hasLowercase && hasNumber && hasSpecial) {
		return errors.New("password must contain uppercase, lowercase, number, and special character")
	}

	return nil
}

func ValidateUsername(username string) error {
	// Username validation using configuration
	if len(username) < config.Config.MinUsernameLen || len(username) > config.Config.MaxUsernameLen {
		return errors.New("username must be between " +
			string(rune(config.Config.MinUsernameLen)) + " and " +
			string(rune(config.Config.MaxUsernameLen)) + " characters")
	}

	// Username character validation (alphanumeric and underscore only)
	usernameRegex := regexp.MustCompile("^[a-zA-Z0-9_]+$")
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain alphanumeric characters and underscores")
	}

	return nil
}

func ValidateUserInput(username, email, password string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}
	return nil
}

func ValidatePostContent(content string) error {
	// Content validation using configuration
	if len(content) < config.Config.MinPostContentLength || len(content) > config.Config.MaxPostContentLength {
		return errors.New("post content must be between " +
			string(rune(config.Config.MinPostContentLength)) + " and " +
			string(rune(config.Config.MaxPostContentLength)) + " characters")
	}

	// Check for prohibited words (example)
	prohibitedWords := []string{"fuck", "bitch", "asshole"}
	for _, word := range prohibitedWords {
		if strings.Contains(strings.ToLower(content), word) {
			return errors.New("post content contains prohibited words")
		}
	}

	return nil
}

func ValidateCommentContent(content string) error {
	// Content validation using configuration
	if len(content) < config.Config.MinCommentLength || len(content) > config.Config.MaxCommentLength {
		return errors.New("comment content must be between " +
			string(rune(config.Config.MinCommentLength)) + " and " +
			string(rune(config.Config.MaxCommentLength)) + " characters")
	}

	// Check for prohibited words (example)
	prohibitedWords := []string{"fuck", "bitch", "asshole"}
	for _, word := range prohibitedWords {
		if strings.Contains(strings.ToLower(content), word) {
			return errors.New("comment content contains prohibited words")
		}
	}

	return nil
}
