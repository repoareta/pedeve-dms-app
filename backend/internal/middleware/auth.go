package middleware

import (
	"strings"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/cookie"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/jwt"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// JWTAuthMiddleware memvalidasi token JWT dan menambahkan info user ke locals (untuk Fiber)
func JWTAuthMiddleware(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()
	var tokenString string

	// Coba ambil token dari cookie terlebih dahulu (metode yang diutamakan)
	cookieToken, err := cookie.GetSecureCookie(c, cookie.GetAuthTokenCookieName())
	if err == nil && cookieToken != "" {
		tokenString = cookieToken
		zapLog.Debug("JWT token found in cookie",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)
	} else {
		// Fallback ke Authorization header (untuk kompatibilitas ke belakang)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
				zapLog.Debug("JWT token found in Authorization header",
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
				)
			}
		}
		
		// Log jika token tidak ditemukan (untuk debugging)
		if tokenString == "" {
			// Log cookie auth_token yang diterima untuk debugging
			authCookieValue := c.Cookies(cookie.GetAuthTokenCookieName())
			zapLog.Warn("JWT token not found",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("ip", c.IP()),
				zap.String("cookie_name", cookie.GetAuthTokenCookieName()),
				zap.String("cookie_value", authCookieValue),
				zap.Error(err),
			)
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
		zapLog.Warn("JWT token validation failed",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Error(err),
		)
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

