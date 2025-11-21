package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/render"
)

// CSRF token store (in-memory, bisa diganti dengan Redis untuk production)
var csrfTokens = make(map[string]time.Time)
var csrfMutex sync.RWMutex

const (
	csrfTokenHeader    = "X-CSRF-Token"
	csrfTokenCookie    = "csrf_token"
	csrfTokenExpiry    = 24 * time.Hour
	csrfTokenCleanupInterval = 1 * time.Hour
)

// GenerateCSRFToken menghasilkan token CSRF baru
func GenerateCSRFToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// StoreCSRFToken menyimpan token CSRF dengan masa kedaluwarsa
func StoreCSRFToken(token string) {
	csrfMutex.Lock()
	defer csrfMutex.Unlock()
	csrfTokens[token] = time.Now().Add(csrfTokenExpiry)
}

// ValidateCSRFToken memvalidasi token CSRF
func ValidateCSRFToken(token string) bool {
	if token == "" {
		return false
	}

	csrfMutex.RLock()
	defer csrfMutex.RUnlock()

	expiry, exists := csrfTokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		// Token expired, hapus
		csrfMutex.RUnlock()
		csrfMutex.Lock()
		delete(csrfTokens, token)
		csrfMutex.Unlock()
		csrfMutex.RLock()
		return false
	}

	return true
}

// CleanupExpiredCSRFTokens menghapus token CSRF yang expired
func CleanupExpiredCSRFTokens() {
	csrfMutex.Lock()
	defer csrfMutex.Unlock()

	now := time.Now()
	for token, expiry := range csrfTokens {
		if now.After(expiry) {
			delete(csrfTokens, token)
		}
	}
}

// StartCSRFTokenCleanup memulai cleanup background untuk token yang expired
func StartCSRFTokenCleanup() {
	go func() {
		ticker := time.NewTicker(csrfTokenCleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			CleanupExpiredCSRFTokens()
		}
	}()
}

// CSRFMiddleware memvalidasi token CSRF untuk request yang mengubah state
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip pengecekan CSRF untuk method yang aman (GET, HEAD, OPTIONS)
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Ambil token CSRF dari header
		csrfToken := r.Header.Get(csrfTokenHeader)
		if csrfToken == "" {
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, ErrorResponse{
				Error:   "csrf_token_missing",
				Message: "CSRF token is required",
			})
			return
		}

		// Validasi token CSRF
		if !ValidateCSRFToken(csrfToken) {
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, ErrorResponse{
				Error:   "csrf_token_invalid",
				Message: "Invalid or expired CSRF token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFTokenHandler mengembalikan token CSRF
// @Summary      Get CSRF token
// @Description  Get a new CSRF token for form submissions
// @Tags         Security
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/csrf-token [get]
func GetCSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := GenerateCSRFToken()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate CSRF token",
		})
		return
	}

	// Simpan token
	StoreCSRFToken(token)

	// Set cookie with CSRF token (optional, untuk double submit cookie pattern)
	isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(csrfTokenExpiry.Seconds()),
		HttpOnly: true,
		Secure:   isHTTPS, // Hanya set flag Secure jika HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"csrf_token": token,
	})
}

