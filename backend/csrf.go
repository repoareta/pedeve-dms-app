package main

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
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

// CSRFMiddleware memvalidasi token CSRF untuk request yang mengubah state (untuk Fiber)
func CSRFMiddleware(c *fiber.Ctx) error {
	// Skip pengecekan CSRF untuk method yang aman (GET, HEAD, OPTIONS)
	method := c.Method()
	if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
		return c.Next()
	}

	// Ambil token CSRF dari header
	csrfToken := c.Get(csrfTokenHeader)
	if csrfToken == "" {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error:   "csrf_token_missing",
			Message: "CSRF token is required",
		})
	}

	// Validasi token CSRF
	if !ValidateCSRFToken(csrfToken) {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error:   "csrf_token_invalid",
			Message: "Invalid or expired CSRF token",
		})
	}

	return c.Next()
}

// GetCSRFTokenHandler mengembalikan token CSRF (untuk Fiber)
// @Summary      Get CSRF token
// @Description  Get a new CSRF token for form submissions
// @Tags         Security
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/csrf-token [get]
func GetCSRFTokenHandler(c *fiber.Ctx) error {
	token, err := GenerateCSRFToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate CSRF token",
		})
	}

	// Simpan token
	StoreCSRFToken(token)

	// Set cookie dengan CSRF token (opsional, untuk double submit cookie pattern)
	isHTTPS := c.Protocol() == "https" || c.Get("X-Forwarded-Proto") == "https"
	c.Cookie(&fiber.Cookie{
		Name:     csrfTokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(csrfTokenExpiry.Seconds()),
		HTTPOnly: true,
		Secure:   isHTTPS, // Hanya set flag Secure jika HTTPS
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(map[string]string{
		"csrf_token": token,
	})
}

