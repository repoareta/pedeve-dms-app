package cookie

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

const (
	authTokenCookie = "auth_token"
	cookieMaxAge    = 24 * 60 * 60 // 24 jam dalam detik
)

// SetSecureCookie mengatur cookie aman dengan flag yang sesuai (untuk Fiber)
func SetSecureCookie(c *fiber.Ctx, name, value string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	// SameSite: "Lax" untuk development (memungkinkan cookie terkirim dari cross-site navigation)
	// "Strict" untuk production (lebih aman, tapi bisa memblokir beberapa use case)
	sameSite := "Lax"
	if isHTTPS {
		sameSite = "Strict" // Production: gunakan Strict untuk keamanan maksimal
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HTTPOnly: true,              // Cegah serangan XSS
		Secure:   isHTTPS,           // Hanya kirim melalui HTTPS di production
		SameSite: sameSite,          // Lax untuk development, Strict untuk production
	})
}

// GetSecureCookie mengambil nilai cookie aman (untuk Fiber)
func GetSecureCookie(c *fiber.Ctx, name string) (string, error) {
	cookieValue := c.Cookies(name)
	if cookieValue == "" {
		return "", fiber.ErrUnauthorized
	}
	return cookieValue, nil
}

// DeleteSecureCookie menghapus cookie aman (untuk Fiber)
func DeleteSecureCookie(c *fiber.Ctx, name string) {
	isHTTPS := os.Getenv("ENV") == "production" || 
	           os.Getenv("HTTPS") == "true" ||
	           os.Getenv("FORCE_HTTPS") == "true"
	
	// SameSite: "Lax" untuk development, "Strict" untuk production
	sameSite := "Lax"
	if isHTTPS {
		sameSite = "Strict"
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Hapus cookie
		HTTPOnly: true,
		Secure:   isHTTPS,
		SameSite: sameSite,
	})
}

// GetAuthTokenCookieName mengembalikan nama cookie untuk auth token
func GetAuthTokenCookieName() string {
	return authTokenCookie
}

