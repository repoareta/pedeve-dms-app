package utils

import "os"

// IsProduction returns true when running in production environment.
func IsProduction() bool {
	env := os.Getenv("ENV")
	return env == "production" || env == "prod"
}
