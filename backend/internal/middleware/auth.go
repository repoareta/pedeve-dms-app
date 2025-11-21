package middleware

import (
	"strings"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/cookie"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/jwt"
	"github.com/gofiber/fiber/v2"
)

// JWTAuthMiddleware memvalidasi token JWT dan menambahkan info user ke locals (untuk Fiber)
func JWTAuthMiddleware(c *fiber.Ctx) error {
	var tokenString string

	// Coba ambil token dari cookie terlebih dahulu (metode yang diutamakan)
	cookieToken, err := cookie.GetSecureCookie(c, cookie.GetAuthTokenCookieName())
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
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required. Please login.",
		})
	}

	// Validasi token
	claims, err := jwt.ValidateJWT(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
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

