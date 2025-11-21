package main

import (
	"net/http"
	"os"
)

const (
	authTokenCookie = "auth_token"
	cookieMaxAge    = 24 * 60 * 60 // 24 hours in seconds
)

// SetSecureCookie sets a secure cookie with proper flags
func SetSecureCookie(w http.ResponseWriter, name, value string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,              // Prevent XSS attacks
		Secure:   isHTTPS,           // Only send over HTTPS in production
		SameSite: http.SameSiteStrictMode, // CSRF protection
	})
}

// GetSecureCookie gets a secure cookie value
func GetSecureCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteSecureCookie deletes a secure cookie
func DeleteSecureCookie(w http.ResponseWriter, name string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   isHTTPS,
		SameSite: http.SameSiteStrictMode,
	})
}

