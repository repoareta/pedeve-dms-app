package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTAuthMiddleware memvalidasi token JWT dan menambahkan info user ke locals (untuk Fiber)
func JWTAuthMiddleware(c *fiber.Ctx) error {
	var tokenString string

	// Coba ambil token dari cookie terlebih dahulu (metode yang diutamakan)
	cookieToken, err := GetSecureCookie(c, authTokenCookie)
	if err == nil && cookieToken != "" {
		tokenString = cookieToken
	} else {
		// Fallback ke Authorization header (untuk kompatibilitas ke belakang)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
	}

	// Jika token tidak ditemukan, return unauthorized
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required. Please login.",
		})
	}

	// Validasi token
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or expired token",
		})
	}

	// Tambahkan info user ke locals (Fiber equivalent dari context)
	c.Locals("userID", claims.UserID)
	c.Locals("username", claims.Username)

	// Panggil handler berikutnya
	return c.Next()
}

// SecurityHeadersMiddleware menambahkan security headers (untuk Fiber)
func SecurityHeadersMiddleware(c *fiber.Ctx) error {
	// Header keamanan
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	// Cek apakah ini route Swagger
	if strings.HasPrefix(c.Path(), "/swagger") {
		// Header yang lebih permisif untuk Swagger UI
		c.Set("X-Frame-Options", "SAMEORIGIN")
		// Izinkan inline scripts dan styles untuk Swagger UI
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")
	} else {
		// Header ketat untuk route API
		c.Set("X-Frame-Options", "DENY")
		c.Set("Content-Security-Policy", "default-src 'self'")
	}

	// Panggil handler berikutnya
	return c.Next()
}

