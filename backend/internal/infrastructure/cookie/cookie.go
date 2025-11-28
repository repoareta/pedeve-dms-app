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
	
	// SameSite: "None" untuk cross-site requests (frontend dan backend di subdomain berbeda)
	// "None" memerlukan Secure: true (HTTPS)
	// Ini diperlukan karena frontend di pedeve-dev.aretaamany.com dan backend di api-pedeve-dev.aretaamany.com
	sameSite := "Lax"
	if isHTTPS {
		sameSite = "None" // Production: gunakan None untuk cross-site requests dengan Secure: true
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HTTPOnly: true,              // Cegah serangan XSS
		Secure:   isHTTPS,           // Hanya kirim melalui HTTPS di production (required untuk SameSite=None)
		SameSite: sameSite,          // Lax untuk development, None untuk production cross-site
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
	
	// SameSite: "None" untuk cross-site requests (harus match dengan SetSecureCookie)
	sameSite := "Lax"
	if isHTTPS {
		sameSite = "None" // Production: gunakan None untuk cross-site requests
	}
	
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Hapus cookie
		HTTPOnly: true,
		Secure:   isHTTPS, // Required untuk SameSite=None
		SameSite: sameSite,
	})
}

// GetAuthTokenCookieName mengembalikan nama cookie untuk auth token
func GetAuthTokenCookieName() string {
	return authTokenCookie
}

