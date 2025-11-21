package http

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/audit"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/cookie"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/jwt"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/validation"
	"github.com/Fajarriswandi/dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Login handles user login (untuk Fiber)
// @Summary      User login
// @Description  Authentikasi user dan kembalikan JWT token. Mendukung login dengan username atau email. Jika 2FA aktif, akan memerlukan kode verifikasi tambahan.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials  body      domain.LoginRequest  true  "Login credentials (username/email dan password, opsional: code untuk 2FA)"
// @Success      200          {object}  domain.AuthResponse  "Login berhasil, token JWT dikembalikan dalam response body dan httpOnly cookie"
// @Success      200          {object}  map[string]interface{}  "2FA diperlukan: requires_2fa: true, kirim code pada request berikutnya"
// @Failure      400          {object}  domain.ErrorResponse  "Request body tidak valid atau validation error"
// @Failure      401          {object}  domain.ErrorResponse  "Kredensial tidak valid atau 2FA code salah"
// @Failure      429          {object}  domain.ErrorResponse  "Terlalu banyak request, rate limit terlampaui"
// @Router       /api/v1/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate and sanitize input
	if err := validation.ValidateLoginInput(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	// Ambil IP client dan User Agent untuk audit logging
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")

	// Cari user di database (bisa login dengan username atau email)
	var userModel domain.UserModel
	result := database.GetDB().Where("username = ? OR email = ?", req.Username, req.Username).First(&userModel)
	if result.Error == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	// Cek password
	if !password.CheckPasswordHash(req.Password, userModel.Password) {
		// Log percobaan login yang gagal
		audit.LogAction(userModel.ID, userModel.Username, audit.ActionFailedLogin, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusFailure, map[string]interface{}{
			"reason": "invalid_password",
		})

		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	// Cek apakah 2FA diaktifkan
	var twoFA domain.TwoFactorAuth
	result2FA := database.GetDB().Where("user_id = ? AND enabled = ?", userModel.ID, true).First(&twoFA)
	is2FAEnabled := result2FA.Error == nil

	// Jika 2FA diaktifkan, verifikasi kode
	if is2FAEnabled {
		if req.Code == "" {
			// 2FA diperlukan tapi kode tidak diberikan
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"requires_2fa": true,
				"message":      "2FA verification required. Please enter your 2FA code.",
			})
		}

		// Verifikasi kode 2FA
		valid, err := usecase.Verify2FALogin(userModel.ID, req.Code)
		if err != nil || !valid {
			// Log percobaan 2FA yang gagal
			audit.LogAction(userModel.ID, userModel.Username, audit.ActionFailedLogin, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusFailure, map[string]interface{}{
				"reason": "invalid_2fa_code",
			})

			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "invalid_2fa_code",
				Message: "Invalid 2FA code",
			})
		}
	}

	// Log login yang berhasil
	audit.LogAction(userModel.ID, userModel.Username, audit.ActionLogin, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusSuccess, nil)

	// Generate JWT token
	token, err := jwt.GenerateJWT(userModel.ID, userModel.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
	}

	// Set cookie aman dengan token JWT (httpOnly cookie untuk keamanan yang lebih baik)
	cookie.SetSecureCookie(c, cookie.GetAuthTokenCookieName(), token)

	// Kembalikan response
	return c.Status(fiber.StatusOK).JSON(domain.AuthResponse{
		Token: token,
		User: domain.User{
			ID:        userModel.ID,
			Username:  userModel.Username,
			Email:     userModel.Email,
			Role:      userModel.Role,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		},
	})
}

// GetProfile returns current user profile (untuk Fiber)
// @Summary      Get user profile
// @Description  Mengambil profil user yang sedang terautentikasi. Data user diambil dari JWT token yang tersimpan dalam cookie.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.User  "Profil user berhasil diambil (tanpa informasi password)"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau expired, user tidak terautentikasi"
// @Failure      404  {object}  domain.ErrorResponse  "User tidak ditemukan di database"
// @Router       /api/v1/auth/profile [get]
func GetProfile(c *fiber.Ctx) error {
	// Ambil user dari locals (diset oleh JWT middleware)
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	_ = c.Locals("username") // Tersedia tapi tidak diperlukan untuk lookup

	// Cari user di database
	var userModel domain.UserModel
	result := database.GetDB().First(&userModel, "id = ?", userID)
	if result.Error == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// Return user (tanpa password)
	return c.Status(fiber.StatusOK).JSON(domain.User{
		ID:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	})
}

// Logout handles user logout (untuk Fiber)
// @Summary      User logout
// @Description  Logout user dan hapus authentication cookie. Aksi logout dicatat dalam audit log untuk keamanan.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Logout berhasil: message: 'Logged out successfully'"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Router       /api/v1/auth/logout [post]
func Logout(c *fiber.Ctx) error {
	// Ambil info user dari locals untuk audit logging
	userIDVal := c.Locals("userID")
	usernameVal := c.Locals("username")

	if userIDVal != nil && usernameVal != nil {
		userID := userIDVal.(string)
		username := usernameVal.(string)

		// Ambil alamat IP dan user agent untuk audit log
		ipAddress := getClientIP(c)
		userAgent := c.Get("User-Agent")

		// Log aksi logout
		audit.LogAction(userID, username, audit.ActionLogout, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusSuccess, nil)
	}

	// Hapus cookie aman
	cookie.DeleteSecureCookie(c, cookie.GetAuthTokenCookieName())

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// getClientIP extracts client IP from request (untuk Fiber)
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For bisa mengandung multiple IPs dipisahkan koma
		for i, char := range xff {
			if char == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	xri := c.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to IP() method dari Fiber
	ip := c.IP()
	if ip == "" || ip == "::1" {
		return "127.0.0.1"
	}
	return ip
}

