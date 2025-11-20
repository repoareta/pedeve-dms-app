package main

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = bluemonday.UGCPolicy()

// SanitizeString sanitizes a string input
func SanitizeString(input string) string {
	// Trim whitespace
	trimmed := strings.TrimSpace(input)

	// Remove HTML tags
	cleaned := sanitizer.Sanitize(trimmed)

	// Escape HTML entities
	escaped := html.EscapeString(cleaned)

	return escaped
}

// SanitizeEmail sanitizes and validates email
func SanitizeEmail(email string) (string, error) {
	sanitized := strings.TrimSpace(strings.ToLower(email))

	if !govalidator.IsEmail(sanitized) {
		return "", fmt.Errorf("invalid email format")
	}

	return sanitized, nil
}

// SanitizeUsername sanitizes and validates username
func SanitizeUsername(username string) (string, error) {
	// Trim and convert to lowercase
	sanitized := strings.TrimSpace(strings.ToLower(username))

	// Validate length
	if len(sanitized) < 3 {
		return "", fmt.Errorf("username must be at least 3 characters")
	}
	if len(sanitized) > 50 {
		return "", fmt.Errorf("username must be less than 50 characters")
	}

	// Validate format: alphanumeric and underscore only
	matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, sanitized)
	if !matched {
		return "", fmt.Errorf("username can only contain lowercase letters, numbers, and underscores")
	}

	return sanitized, nil
}

// SanitizePassword validates password strength
func SanitizePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	// Check for at least one letter and one number
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	return nil
}

// ValidateInput validates and sanitizes RegisterRequest
func ValidateRegisterInput(req *RegisterRequest) error {
	// Sanitize username
	username, err := SanitizeUsername(req.Username)
	if err != nil {
		return fmt.Errorf("username: %v", err)
	}
	req.Username = username

	// Sanitize email
	email, err := SanitizeEmail(req.Email)
	if err != nil {
		return fmt.Errorf("email: %v", err)
	}
	req.Email = email

	// Validate password
	if err := SanitizePassword(req.Password); err != nil {
		return fmt.Errorf("password: %v", err)
	}

	return nil
}

// ValidateLoginInput validates and sanitizes LoginRequest
func ValidateLoginInput(req *LoginRequest) error {
	// Trim whitespace
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Check if empty
	if req.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if req.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// If username looks like email, validate email format
	if strings.Contains(req.Username, "@") {
		if !govalidator.IsEmail(req.Username) {
			return fmt.Errorf("invalid email format")
		}
		req.Username = strings.ToLower(req.Username)
	} else {
		// Validate username format
		if len(req.Username) < 3 {
			return fmt.Errorf("username must be at least 3 characters")
		}
	}

	return nil
}

// SanitizeSQLInput sanitizes input to prevent SQL injection (additional layer)
func SanitizeSQLInput(input string) string {
	// Remove SQL injection patterns
	dangerous := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"DROP", "DELETE", "INSERT", "UPDATE", "SELECT",
		"EXEC", "EXECUTE", "UNION", "SCRIPT",
	}

	sanitized := input
	for _, pattern := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}

	return strings.TrimSpace(sanitized)
}

