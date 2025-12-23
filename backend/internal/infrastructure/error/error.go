package error

import (
	"os"
	"strings"
)

// SanitizeErrorMessage mengembalikan error message yang aman untuk production
// Di production, error message akan digeneralisasi untuk mencegah information disclosure
// Di development, error message detail tetap ditampilkan untuk debugging
func SanitizeErrorMessage(err error, defaultMessage string) string {
	if err == nil {
		return defaultMessage
	}

	env := os.Getenv("ENV")
	isProduction := env == "production" || env == "prod"

	// Di production, gunakan generic message
	if isProduction {
		return defaultMessage
	}

	// Di development, tampilkan error detail untuk debugging
	return err.Error()
}

// SanitizeErrorString mengembalikan error message string yang aman untuk production
func SanitizeErrorString(errMsg string, defaultMessage string) string {
	if errMsg == "" {
		return defaultMessage
	}

	env := os.Getenv("ENV")
	isProduction := env == "production" || env == "prod"

	// Di production, gunakan generic message
	if isProduction {
		return defaultMessage
	}

	// Di development, tampilkan error detail untuk debugging
	return errMsg
}

// GetGenericErrorMessage mengembalikan generic error message berdasarkan error type
func GetGenericErrorMessage(err error) string {
	if err == nil {
		return "An error occurred"
	}

	errMsg := strings.ToLower(err.Error())

	// Map common error patterns to generic messages
	switch {
	case strings.Contains(errMsg, "database") || strings.Contains(errMsg, "sql"):
		return "Database operation failed. Please try again later."
	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "network"):
		return "Connection error. Please check your network connection."
	case strings.Contains(errMsg, "timeout"):
		return "Request timeout. Please try again."
	case strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "access denied"):
		return "You don't have permission to perform this action."
	case strings.Contains(errMsg, "not found"):
		return "The requested resource was not found."
	case strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate"):
		return "The resource already exists."
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		return "Invalid input. Please check your data and try again."
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "authentication"):
		return "Authentication failed. Please login again."
	case strings.Contains(errMsg, "forbidden"):
		return "Access denied. You don't have permission to access this resource."
	default:
		return "An unexpected error occurred. Please try again later."
	}
}

// IsProduction checks if the application is running in production environment
func IsProduction() bool {
	env := os.Getenv("ENV")
	return env == "production" || env == "prod"
}
