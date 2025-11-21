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

// SanitizeString membersihkan input string
func SanitizeString(input string) string {
	// Trim whitespace
	trimmed := strings.TrimSpace(input)

	// Hapus tag HTML
	cleaned := sanitizer.Sanitize(trimmed)

	// Escape entity HTML
	escaped := html.EscapeString(cleaned)

	return escaped
}

// SanitizeEmail membersihkan dan memvalidasi email
func SanitizeEmail(email string) (string, error) {
	sanitized := strings.TrimSpace(strings.ToLower(email))

	if !govalidator.IsEmail(sanitized) {
		return "", fmt.Errorf("invalid email format")
	}

	return sanitized, nil
}

// SanitizeUsername membersihkan dan memvalidasi username
func SanitizeUsername(username string) (string, error) {
	// Trim dan konversi ke lowercase
	sanitized := strings.TrimSpace(strings.ToLower(username))

	// Validasi panjang
	if len(sanitized) < 3 {
		return "", fmt.Errorf("username must be at least 3 characters")
	}
	if len(sanitized) > 50 {
		return "", fmt.Errorf("username must be less than 50 characters")
	}

	// Validasi format: hanya alphanumeric dan underscore
	matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, sanitized)
	if !matched {
		return "", fmt.Errorf("username can only contain lowercase letters, numbers, and underscores")
	}

	return sanitized, nil
}

// SanitizePassword memvalidasi kekuatan password
func SanitizePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	// Cek minimal satu huruf dan satu angka
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

// ValidateInput memvalidasi dan membersihkan RegisterRequest
func ValidateRegisterInput(req *RegisterRequest) error {
	// Bersihkan username
	username, err := SanitizeUsername(req.Username)
	if err != nil {
		return fmt.Errorf("username: %v", err)
	}
	req.Username = username

	// Bersihkan email
	email, err := SanitizeEmail(req.Email)
	if err != nil {
		return fmt.Errorf("email: %v", err)
	}
	req.Email = email

	// Validasi password
	if err := SanitizePassword(req.Password); err != nil {
		return fmt.Errorf("password: %v", err)
	}

	return nil
}

// ValidateLoginInput memvalidasi dan membersihkan LoginRequest
func ValidateLoginInput(req *LoginRequest) error {
	// Trim whitespace
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Cek apakah kosong
	if req.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if req.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Jika username terlihat seperti email, validasi format email
	if strings.Contains(req.Username, "@") {
		if !govalidator.IsEmail(req.Username) {
			return fmt.Errorf("invalid email format")
		}
		req.Username = strings.ToLower(req.Username)
	} else {
		// Validasi format username
		if len(req.Username) < 3 {
			return fmt.Errorf("username must be at least 3 characters")
		}
	}

	return nil
}

// SanitizeSQLInput membersihkan input untuk mencegah SQL injection (lapisan tambahan)
func SanitizeSQLInput(input string) string {
	// Hapus pola SQL injection
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

