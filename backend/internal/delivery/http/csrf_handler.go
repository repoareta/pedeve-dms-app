package http

import (
	"os"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// GetCSRFTokenHandler mengembalikan token CSRF (untuk Fiber)
// @Summary      Ambil Token CSRF
// @Description  Mengambil token CSRF baru untuk digunakan pada request yang mengubah state (POST, PUT, DELETE, PATCH). Token CSRF juga disimpan dalam cookie (csrf_token) untuk double-submit cookie pattern. Endpoint ini public dan tidak memerlukan authentication.
// @Tags         Security
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Token CSRF berhasil diambil. Response berisi csrf_token. Token juga disimpan dalam cookie csrf_token."
// @Failure      500  {object}  domain.ErrorResponse  "Gagal generate token CSRF"
// @Router       /api/v1/csrf-token [get]
// @note         Catatan Teknis:
// @note         1. Public Endpoint: Endpoint ini tidak memerlukan authentication
// @note         2. CSRF Token: Token CSRF berlaku selama 24 jam dan disimpan dalam memory (bisa diganti dengan Redis untuk production)
// @note         3. Double-Submit Cookie: Token dikembalikan dalam response body dan juga disimpan dalam cookie untuk keamanan
// @note         4. Usage: Token harus dikirim dalam header X-CSRF-Token untuk request POST, PUT, DELETE, PATCH
// @note         5. Cookie: Cookie csrf_token menggunakan SameSite: Lax untuk development, Strict untuk production
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

