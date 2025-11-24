package http

import (
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// GetCSRFTokenHandler mengembalikan token CSRF (untuk Fiber)
// @Summary      Get CSRF token
// @Description  Get a new CSRF token for form submissions
// @Tags         Security
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/csrf-token [get]
func GetCSRFTokenHandler(c *fiber.Ctx) error {
	token, err := middleware.GenerateCSRFToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate CSRF token",
		})
	}

	// Simpan token
	middleware.StoreCSRFToken(token)

	// Set cookie with CSRF token (optional, untuk double submit cookie pattern)
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true" ||
	           c.Protocol() == "https" || 
	           c.Get("X-Forwarded-Proto") == "https"
	
	// SameSite: "Lax" untuk development (memungkinkan cookie terkirim dari cross-site navigation)
	// "Strict" untuk production (lebih aman, tapi bisa memblokir beberapa use case)
	sameSite := "Lax"
	if isHTTPS {
		sameSite = "Strict" // Production: gunakan Strict untuk keamanan maksimal
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(24 * 60 * 60), // 24 jam
		HTTPOnly: true,
		Secure:   isHTTPS, // Hanya set flag Secure jika HTTPS
		SameSite: sameSite, // Lax untuk development, Strict untuk production
	})

	return c.Status(fiber.StatusOK).JSON(map[string]string{
		"csrf_token": token,
	})
}

