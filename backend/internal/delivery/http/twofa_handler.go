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
// @Summary      Generate 2FA secret
// @Description  Generate a new TOTP secret and QR code for 2FA setup
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/auth/2fa/generate [post]
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
// @Summary      Verify and enable 2FA
// @Description  Verify TOTP code and enable 2FA for user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code  body      map[string]string  true  "TOTP code"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      401   {object}  domain.ErrorResponse
// @Router       /api/v1/auth/2fa/verify [post]
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
// @Summary      Get 2FA status
// @Description  Get current user's 2FA status
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/auth/2fa/status [get]
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
// @Summary      Disable 2FA
// @Description  Disable 2FA for current user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/auth/2fa/disable [post]
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

