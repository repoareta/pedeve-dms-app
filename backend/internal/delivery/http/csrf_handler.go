package http

import (
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
	isHTTPS := c.Protocol() == "https" || c.Get("X-Forwarded-Proto") == "https"
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(24 * 60 * 60), // 24 jam
		HTTPOnly: true,
		Secure:   isHTTPS, // Hanya set flag Secure jika HTTPS
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(map[string]string{
		"csrf_token": token,
	})
}

