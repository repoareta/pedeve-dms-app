package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

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

