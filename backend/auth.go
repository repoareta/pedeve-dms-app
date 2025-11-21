package main

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Login handles user login (untuk Fiber)
// @Summary      User login
// @Description  Authentikasi user dan kembalikan JWT token. Mendukung login dengan username atau email. Jika 2FA aktif, akan memerlukan kode verifikasi tambahan.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Login credentials (username/email dan password, opsional: code untuk 2FA)"
// @Success      200          {object}  AuthResponse  "Login berhasil, token JWT dikembalikan dalam response body dan httpOnly cookie"
// @Success      200          {object}  map[string]interface{}  "2FA diperlukan: requires_2fa: true, kirim code pada request berikutnya"
// @Failure      400          {object}  ErrorResponse  "Request body tidak valid atau validation error"
// @Failure      401          {object}  ErrorResponse  "Kredensial tidak valid atau 2FA code salah"
// @Failure      429          {object}  ErrorResponse  "Terlalu banyak request, rate limit terlampaui"
// @Router       /api/v1/auth/login [post]
// @Security     None
// @note         Catatan Teknis:
// @note         1. Rate Limiting: Endpoint ini memiliki rate limiting khusus (5 requests per minute, burst: 5) untuk mencegah brute force attack
// @note         2. Password Hashing: Menggunakan bcrypt dengan cost factor default Go (10 rounds)
// @note         3. JWT Token: Token disimpan dalam httpOnly cookie (auth_token) untuk keamanan XSS, lifetime 24 jam
// @note         4. 2FA Support: Jika user mengaktifkan 2FA, endpoint ini akan return requires_2fa: true pada request pertama
// @note         5. Audit Logging: Semua percobaan login (berhasil/gagal) dicatat dalam audit log untuk keamanan
// @note         6. CSRF Protection: Tidak berlaku untuk endpoint public ini, hanya untuk authenticated requests
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Login credentials"
// @Success      200          {object}  AuthResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/v1/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate and sanitize input
	if err := ValidateLoginInput(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	// Ambil IP client dan User Agent untuk audit logging
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")

	// Cari user di database (bisa login dengan username atau email)
	var userModel UserModel
	result := DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&userModel)
	if result.Error == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	// Cek password
	if !CheckPasswordHash(req.Password, userModel.Password) {
		// Log percobaan login yang gagal
		LogAction(userModel.ID, userModel.Username, ActionFailedLogin, ResourceAuth, "", ipAddress, userAgent, StatusFailure, map[string]interface{}{
			"reason": "invalid_password",
		})

		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	// Cek apakah 2FA diaktifkan
	var twoFA TwoFactorAuth
	result2FA := DB.Where("user_id = ? AND enabled = ?", userModel.ID, true).First(&twoFA)
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
		valid, err := Verify2FALogin(userModel.ID, req.Code)
		if err != nil || !valid {
			// Log percobaan 2FA yang gagal
			LogAction(userModel.ID, userModel.Username, ActionFailedLogin, ResourceAuth, "", ipAddress, userAgent, StatusFailure, map[string]interface{}{
				"reason": "invalid_2fa_code",
			})

			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Error:   "invalid_2fa_code",
				Message: "Invalid 2FA code",
			})
		}
	}

	// Log login yang berhasil
	LogAction(userModel.ID, userModel.Username, ActionLogin, ResourceAuth, "", ipAddress, userAgent, StatusSuccess, nil)

	// Generate JWT token
	token, err := GenerateJWT(userModel.ID, userModel.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
	}

	// Set cookie aman dengan token JWT (httpOnly cookie untuk keamanan yang lebih baik)
	SetSecureCookie(c, authTokenCookie, token)

	// Kembalikan response
	return c.Status(fiber.StatusOK).JSON(AuthResponse{
		Token: token,
		User: User{
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
// @Success      200  {object}  User  "Profil user berhasil diambil (tanpa informasi password)"
// @Failure      401  {object}  ErrorResponse  "Token tidak valid atau expired, user tidak terautentikasi"
// @Failure      404  {object}  ErrorResponse  "User tidak ditemukan di database"
// @Router       /api/v1/auth/profile [get]
// @Security     BearerAuth
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam cookie httpOnly (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Data Privacy: Password tidak pernah dikembalikan dalam response untuk keamanan
// @note         4. User Context: User ID dan username diambil dari JWT claims, tidak dari request body
func GetProfile(c *fiber.Ctx) error {
	// Ambil user dari locals (diset oleh JWT middleware)
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)
	_ = c.Locals("username") // Tersedia tapi tidak diperlukan untuk lookup

	// Cari user di database
	var userModel UserModel
	result := DB.First(&userModel, "id = ?", userID)
	if result.Error == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// Return user (tanpa password)
	return c.Status(fiber.StatusOK).JSON(User{
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
// @Failure      401  {object}  ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Router       /api/v1/auth/logout [post]
// @Security     BearerAuth
// @note         Catatan Teknis:
// @note         1. Cookie Deletion: HttpOnly cookie (auth_token) dihapus dengan MaxAge: -1 untuk memastikan cookie dihapus di browser
// @note         2. CSRF Protection: Memerlukan CSRF token yang valid dalam header X-CSRF-Token untuk keamanan
// @note         3. Audit Logging: Semua aksi logout dicatat dengan IP address dan user agent untuk audit trail
// @note         4. State Clear: Token dihapus dari cookie, namun tidak ada blacklist token (stateless JWT approach)
// @note         5. Security: Setelah logout, user harus login kembali untuk mengakses protected endpoints
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
		LogAction(userID, username, ActionLogout, ResourceAuth, "", ipAddress, userAgent, StatusSuccess, nil)
	}

	// Hapus cookie aman
	DeleteSecureCookie(c, authTokenCookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
