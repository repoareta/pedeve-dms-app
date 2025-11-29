package http

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/cookie"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/jwt"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/password"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/validation"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Login handles user login (untuk Fiber)
// @Summary      Login User
// @Description  Autentikasi user dan kembalikan JWT token. Mendukung login dengan username atau email. Jika 2FA aktif, akan memerlukan kode verifikasi tambahan pada request berikutnya.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials  body      domain.LoginRequest  true  "Kredensial login (username/email, password, dan opsional: code untuk 2FA)"
// @Success      200          {object}  domain.AuthResponse  "Login berhasil. Token JWT dikembalikan dalam response body dan disimpan dalam httpOnly cookie (auth_token) untuk keamanan."
// @Success      200          {object}  map[string]interface{}  "2FA diperlukan. Response berisi requires_2fa: true. Kirim kode 2FA pada request login berikutnya dengan field 'code'."
// @Failure      400          {object}  domain.ErrorResponse  "Request body tidak valid atau validation error (username/password tidak memenuhi syarat)"
// @Failure      401          {object}  domain.ErrorResponse  "Kredensial tidak valid atau kode 2FA salah"
// @Failure      429          {object}  domain.ErrorResponse  "Terlalu banyak request, rate limit terlampaui (5 req/min untuk auth endpoints)"
// @Router       /api/v1/auth/login [post]
// @note         Catatan Teknis:
// @note         1. Authentication: JWT token disimpan dalam httpOnly cookie untuk mencegah XSS attacks
// @note         2. 2FA Support: Jika user memiliki 2FA aktif, response pertama akan berisi requires_2fa: true
// @note         3. Rate Limiting: Endpoint ini memiliki rate limiting khusus (5 req/min, burst: 5) untuk mencegah brute force
// @note         4. Audit Logging: Semua percobaan login (berhasil/gagal) dicatat dalam audit log
// @note         5. Password: Password di-hash menggunakan bcrypt sebelum disimpan di database
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
	zapLog := logger.GetLogger()
	var userModel domain.UserModel
	// Support login dengan username atau email (case-insensitive)
	result := database.GetDB().Where("LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)", req.Username, req.Username).First(&userModel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Log untuk debugging
			zapLog.Debug("User not found for login attempt",
				zap.String("input", req.Username),
				zap.String("ip", getClientIP(c)),
			)
		} else {
			// Database error
			zapLog.Error("Database error during login",
				zap.String("input", req.Username),
				zap.String("ip", getClientIP(c)),
				zap.Error(result.Error),
			)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email/username or password",
		})
	}

	zapLog.Debug("User found for login",
		zap.String("user_id", userModel.ID),
		zap.String("username", userModel.Username),
		zap.Bool("is_active", userModel.IsActive),
		zap.String("role", userModel.Role),
	)

	// Cek apakah user aktif
	if !userModel.IsActive {
		zapLog.Warn("Login attempt for inactive user",
			zap.String("username", userModel.Username),
			zap.String("ip", getClientIP(c)),
		)
		audit.LogAction(userModel.ID, userModel.Username, audit.ActionFailedLogin, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusFailure, map[string]interface{}{
			"reason": "user_inactive",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "account_inactive",
			Message: "Your account is inactive. Please contact administrator.",
		})
	}

	// Cek password
	passwordValid := password.CheckPasswordHash(req.Password, userModel.Password)
	if !passwordValid {
		zapLog.Warn("Invalid password for login attempt",
			zap.String("user_id", userModel.ID),
			zap.String("username", userModel.Username),
			zap.String("email", userModel.Email),
			zap.String("input_email_or_username", req.Username),
			zap.String("ip", getClientIP(c)),
			zap.Bool("is_active", userModel.IsActive),
			zap.String("role", userModel.Role),
			zap.String("role_id", func() string {
				if userModel.RoleID != nil {
					return *userModel.RoleID
				}
				return "nil"
			}()),
			zap.String("company_id", func() string {
				if userModel.CompanyID != nil {
					return *userModel.CompanyID
				}
				return "nil"
			}()),
		)
		// Log percobaan login yang gagal
		audit.LogAction(userModel.ID, userModel.Username, audit.ActionFailedLogin, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusFailure, map[string]interface{}{
			"reason": "invalid_password",
		})

		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email/username or password",
		})
	}

	zapLog.Debug("Password verified successfully",
		zap.String("user_id", userModel.ID),
	)

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

	// Get user auth info (role, company, permissions) untuk JWT claims
	roleID, roleName, companyID, companyLevel, hierarchyScope, permissions, err := usecase.GetUserAuthInfo(userModel.ID)
	if err != nil {
		zapLog.Warn("Failed to get user auth info, using fallback",
			zap.String("user_id", userModel.ID),
			zap.Error(err),
		)
		// Fallback jika error (backward compatibility)
		roleName = userModel.Role
		if roleName == "" {
			roleName = "user"
		}
		permissions = []string{}
		// Set default values untuk fallback
		if userModel.RoleID != nil {
			roleID = userModel.RoleID
		}
		if userModel.CompanyID != nil {
			companyID = userModel.CompanyID
			companyLevel = 1 // Default level
			hierarchyScope = "company"
		} else {
			hierarchyScope = "global"
		}
	}

	// Generate JWT token dengan claims lengkap
	token, err := jwt.GenerateJWT(userModel.ID, userModel.Username, roleID, roleName, companyID, companyLevel, hierarchyScope, permissions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
	}

	// Set cookie aman dengan token JWT (httpOnly cookie untuk keamanan yang lebih baik)
	cookie.SetSecureCookie(c, cookie.GetAuthTokenCookieName(), token)

	// Kembalikan response dengan role tertinggi dari GetUserAuthInfo
	return c.Status(fiber.StatusOK).JSON(domain.AuthResponse{
		Token: token,
		User: domain.User{
			ID:        userModel.ID,
			Username:  userModel.Username,
			Email:     userModel.Email,
			Role:      roleName, // Gunakan roleName dari GetUserAuthInfo (role tertinggi)
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		},
	})
}

// GetProfile returns current user profile (untuk Fiber)
// @Summary      Ambil Profil User
// @Description  Mengambil profil user yang sedang terautentikasi. Data user diambil dari JWT token yang tersimpan dalam httpOnly cookie. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.User  "Profil user berhasil diambil. Response berisi id, username, email, role, created_at, dan updated_at (tanpa informasi password)"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau expired, user tidak terautentikasi"
// @Failure      404  {object}  domain.ErrorResponse  "User tidak ditemukan di database"
// @Router       /api/v1/auth/profile [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Data Privacy: Password tidak pernah dikembalikan dalam response untuk keamanan
// @note         4. User Context: User ID dan username diambil dari JWT claims, tidak dari request body
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
	userResponse := domain.User{
		ID:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	}
	// Include company_id if available
	if userModel.CompanyID != nil {
		userResponse.CompanyID = userModel.CompanyID
	}
	return c.Status(fiber.StatusOK).JSON(userResponse)
}

// Logout handles user logout (untuk Fiber)
// @Summary      Logout User
// @Description  Logout user dan hapus authentication cookie (auth_token). Aksi logout dicatat dalam audit log untuk keamanan dan audit trail. Endpoint ini memerlukan CSRF token karena menggunakan method POST.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Logout berhasil. Response berisi message: 'Logged out successfully'. Cookie auth_token akan dihapus."
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403  {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Router       /api/v1/auth/logout [post]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan POST method
// @note         3. Cookie Deletion: Cookie auth_token akan dihapus setelah logout berhasil
// @note         4. Audit Logging: Aksi logout dicatat dalam audit log dengan status success
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

// UpdateProfileEmail handles email update for current user
// @Summary      Update Email User
// @Description  Mengupdate email user yang sedang terautentikasi. User hanya bisa mengupdate email mereka sendiri.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      object  true  "New email"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/auth/profile/email [put]
func UpdateProfileEmail(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	zapLog := logger.GetLogger()

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate email format
	if err := validation.ValidateEmail(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	// Check if email already exists (by another user)
	var existingUser domain.UserModel
	result := database.GetDB().Where("email = ? AND id != ?", req.Email, userID).First(&existingUser)
	if result.Error == nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "email_exists",
			Message: "Email already exists",
		})
	}

	// Get current user
	var userModel domain.UserModel
	if err := database.GetDB().First(&userModel, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// Update email
	userModel.Email = req.Email
	if err := database.GetDB().Save(&userModel).Error; err != nil {
		zapLog.Error("Failed to update email", zap.String("user_id", userID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update email",
		})
	}

	// Log action
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")
	audit.LogAction(userID, username, "update_email", audit.ResourceUser, userID, ipAddress, userAgent, audit.StatusSuccess, map[string]interface{}{
		"new_email": req.Email,
	})

	return c.Status(fiber.StatusOK).JSON(domain.User{
		ID:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	})
}

// ChangePassword handles password change for current user
// @Summary      Change Password User
// @Description  Mengubah password user yang sedang terautentikasi. User harus memberikan password lama untuk verifikasi.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      object  true  "Old password and new password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/auth/profile/password [put]
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	zapLog := logger.GetLogger()

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate new password length
	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "validation_error",
			Message: "New password must be at least 8 characters long",
		})
	}

	// Get current user
	var userModel domain.UserModel
	if err := database.GetDB().First(&userModel, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// Verify old password
	if !password.CheckPasswordHash(req.OldPassword, userModel.Password) {
		// Log failed password change attempt
		ipAddress := getClientIP(c)
		userAgent := c.Get("User-Agent")
		audit.LogAction(userID, username, "change_password_failed", audit.ResourceAuth, userID, ipAddress, userAgent, audit.StatusFailure, map[string]interface{}{
			"reason": "invalid_old_password",
		})

		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "invalid_password",
			Message: "Old password is incorrect",
		})
	}

	// Hash new password
	hashedPassword, err := password.HashPassword(req.NewPassword)
	if err != nil {
		zapLog.Error("Failed to hash new password", zap.String("user_id", userID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "hash_failed",
			Message: "Failed to hash new password",
		})
	}

	// Update password
	userModel.Password = hashedPassword
	if err := database.GetDB().Save(&userModel).Error; err != nil {
		zapLog.Error("Failed to update password", zap.String("user_id", userID), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update password",
		})
	}

	// Log action
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")
	audit.LogAction(userID, username, "change_password", audit.ResourceAuth, userID, ipAddress, userAgent, audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password changed successfully",
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
