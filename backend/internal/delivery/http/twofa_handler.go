package http

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/audit"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Generate2FASecret menghasilkan secret TOTP baru untuk user
// @Summary      Generate Secret 2FA
// @Description  Menghasilkan secret TOTP baru dan QR code untuk setup 2FA. Secret ini digunakan untuk generate kode 2FA di authenticator app (Google Authenticator, Authy, dll). Endpoint ini memerlukan authentication dan CSRF token.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Secret 2FA berhasil di-generate. Response berisi secret, qr_code (base64 image), url (untuk manual entry), dan message"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403  {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Failure      500  {object}  domain.ErrorResponse  "Gagal generate secret 2FA"
// @Router       /api/v1/auth/2fa/generate [post]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan POST method
// @note         3. TOTP Algorithm: Menggunakan TOTP (Time-based One-Time Password) dengan algoritma SHA1
// @note         4. QR Code: QR code dalam format base64 image untuk scan di authenticator app
// @note         5. Secret Storage: Secret disimpan di database dengan encryption (tidak dalam plain text)
func Generate2FASecret(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()
	
	// Ambil user dari locals
	userIDVal := c.Locals("userID")
	usernameVal := c.Locals("username")

	if userIDVal == nil || usernameVal == nil {
		zapLog.Warn("User context not found in request", zap.String("endpoint", "generate_2fa"))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
	}

	userID, ok := userIDVal.(string)
	if !ok {
		zapLog.Error("Invalid userID type in context")
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
	}

	username, ok := usernameVal.(string)
	if !ok {
		zapLog.Error("Invalid username type in context")
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
	}

	result, err := usecase.Generate2FASecretUseCase(userID, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// Verify2FA memverifikasi kode TOTP dan mengaktifkan 2FA
// @Summary      Verifikasi dan Aktifkan 2FA
// @Description  Memverifikasi kode TOTP dari authenticator app dan mengaktifkan 2FA untuk user. Setelah verifikasi berhasil, 2FA akan diaktifkan dan backup codes akan di-generate. Endpoint ini memerlukan authentication dan CSRF token.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code  body      map[string]string  true  "Kode TOTP 6 digit dari authenticator app"
// @Success      200   {object}  map[string]interface{}  "2FA berhasil diaktifkan. Response berisi message, backup_codes (array string), dan enabled: true"
// @Failure      400   {object}  domain.ErrorResponse  "Request body tidak valid atau secret 2FA tidak ditemukan"
// @Failure      401   {object}  domain.ErrorResponse  "Token tidak valid, user tidak terautentikasi, atau kode verifikasi tidak valid"
// @Failure      403   {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Router       /api/v1/auth/2fa/verify [post]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan POST method
// @note         3. TOTP Verification: Kode TOTP divalidasi dengan secret yang di-generate sebelumnya
// @note         4. Backup Codes: Setelah verifikasi berhasil, 10 backup codes akan di-generate untuk recovery
// @note         5. Audit Logging: Aksi enable 2FA dicatat dalam audit log dengan status success
func Verify2FA(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()
	
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		zapLog.Warn("User context not found in Verify2FA",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
	}
	userID := userIDVal.(string)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	result, err := usecase.Verify2FAUseCase(userID, req.Code)
	if err != nil {
		if err.Error() == "invalid verification code" {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "invalid_code",
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
	}

	// Log aksi enable 2FA
	username := ""
	if usernameVal := c.Locals("username"); usernameVal != nil {
		username = usernameVal.(string)
	}
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")
	audit.LogAction(userID, username, audit.ActionEnable2FA, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(result)
}

// Get2FAStatus mengembalikan status 2FA untuk user saat ini
// @Summary      Cek Status 2FA
// @Description  Mengambil status 2FA untuk user yang sedang terautentikasi. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Status 2FA berhasil diambil. Response berisi enabled: true/false"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      500  {object}  domain.ErrorResponse  "Gagal mengambil status 2FA"
// @Router       /api/v1/auth/2fa/status [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Response: Response berisi enabled: true jika 2FA aktif, false jika tidak aktif
// @note         4. User Context: User ID diambil dari JWT claims, tidak dari request body
func Get2FAStatus(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()
	
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		zapLog.Warn("User context not found in request", zap.String("endpoint", "get_2fa_status"))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
	}

	userID, ok := userIDVal.(string)
	if !ok {
		zapLog.Error("Invalid userID type in context")
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
	}

	result, err := usecase.Get2FAStatusUseCase(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// Disable2FA menonaktifkan 2FA untuk user saat ini
// @Summary      Nonaktifkan 2FA
// @Description  Menonaktifkan 2FA untuk user yang sedang terautentikasi. Setelah 2FA dinonaktifkan, user tidak perlu memasukkan kode 2FA saat login. Endpoint ini memerlukan authentication dan CSRF token.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "2FA berhasil dinonaktifkan. Response berisi message: '2FA has been disabled successfully'"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403  {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Failure      404  {object}  domain.ErrorResponse  "2FA tidak aktif untuk user ini (2fa_not_found)"
// @Failure      500  {object}  domain.ErrorResponse  "Gagal menonaktifkan 2FA"
// @Router       /api/v1/auth/2fa/disable [post]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan POST method
// @note         3. Data Deletion: Secret 2FA dan backup codes akan dihapus dari database setelah 2FA dinonaktifkan
// @note         4. Audit Logging: Aksi disable 2FA dicatat dalam audit log dengan status success
// @note         5. Security: Setelah 2FA dinonaktifkan, user hanya perlu password untuk login
func Disable2FA(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()
	
	userIDVal := c.Locals("userID")
	usernameVal := c.Locals("username")

	if userIDVal == nil || usernameVal == nil {
		zapLog.Warn("User context not found in request", zap.String("endpoint", "disable_2fa"))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
	}

	userID, ok := userIDVal.(string)
	if !ok {
		zapLog.Error("Invalid userID type in context")
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
	}

	username, ok := usernameVal.(string)
	if !ok {
		zapLog.Error("Invalid username type in context")
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
	}

	// Ambil alamat IP dan user agent untuk audit log
	ipAddress := getClientIP(c)
	userAgent := c.Get("User-Agent")

	err := usecase.Disable2FAUseCase(userID)
	if err != nil {
		if err.Error() == "2FA is not enabled for this user" {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "2fa_not_found",
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	// Log aksi
	audit.LogAction(userID, username, audit.ActionDisable2FA, audit.ResourceAuth, "", ipAddress, userAgent, audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "2FA has been disabled successfully",
	})
}

