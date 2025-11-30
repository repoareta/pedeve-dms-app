package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"go.uber.org/zap"
)

type DevelopmentHandler struct {
	devUseCase usecase.DevelopmentUseCase
	logger     *zap.Logger
}

func NewDevelopmentHandler(devUseCase usecase.DevelopmentUseCase) *DevelopmentHandler {
	return &DevelopmentHandler{
		devUseCase: devUseCase,
		logger:     logger.GetLogger(),
	}
}

// ResetSubsidiaryData handles resetting all subsidiary data
// @Summary      Reset Data Subsidiary
// @Description  Menghapus semua data subsidiary dan user yang terkait (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/reset-subsidiary [post]
func (h *DevelopmentHandler) ResetSubsidiaryData(c *fiber.Ctx) error {
	// Check if user is superadmin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		h.logger.Warn("RoleName not found in context", zap.String("path", c.Path()))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("Invalid roleName type in context", zap.String("path", c.Path()), zap.Any("roleName", roleNameVal))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	if roleName != "superadmin" {
		h.logger.Warn("Non-superadmin attempted to reset subsidiary data",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Reset subsidiary data
	if err := h.devUseCase.ResetSubsidiaryData(); err != nil {
		h.logger.Error("Failed to reset subsidiary data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "reset_subsidiary_data", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data subsidiary berhasil di-reset",
		"success": true,
	})
}

// RunSubsidiarySeeder handles running the subsidiary seeder
// @Summary      Jalankan Seeder Data Subsidiary
// @Description  Menjalankan seeder untuk membuat sample data subsidiary (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      409  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/run-subsidiary-seeder [post]
func (h *DevelopmentHandler) RunSubsidiarySeeder(c *fiber.Ctx) error {
	// Check if user is superadmin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("Invalid roleName type in context", zap.String("path", c.Path()), zap.Any("roleName", roleNameVal))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	if roleName != "superadmin" {
		h.logger.Warn("Non-superadmin attempted to run seeder",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Run seeder
	alreadyExists, err := h.devUseCase.RunSubsidiarySeeder()
	if err != nil {
		h.logger.Error("Failed to run subsidiary seeder", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "seeder_failed",
			Message: err.Error(),
		})
	}

	if alreadyExists {
		// Audit log
		userID := c.Locals("userID").(string)
		username := c.Locals("username").(string)
		audit.LogAction(userID, username, "run_subsidiary_seeder", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusFailure, map[string]interface{}{
			"reason": "Data seeder sudah ada",
		})

		return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{
			Error:   "seeder_already_exists",
			Message: "Data seeder sudah tersedia. Proses dibatalkan untuk mencegah duplikasi data.",
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "run_subsidiary_seeder", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Seeder data subsidiary berhasil dijalankan",
		"success": true,
	})
}

// CheckSeederDataExists checks if seeder data already exists
// @Summary      Cek Status Seeder Data
// @Description  Mengecek apakah data seeder sudah ada (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-seeder-status [get]
func (h *DevelopmentHandler) CheckSeederDataExists(c *fiber.Ctx) error {
	// Check if user is superadmin
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	roleName, ok := roleNameVal.(string)
	if !ok {
		h.logger.Warn("Invalid roleName type in context", zap.String("path", c.Path()), zap.Any("roleName", roleNameVal))
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	if roleName != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Check seeder data
	exists, err := h.devUseCase.CheckSeederDataExists()
	if err != nil {
		h.logger.Error("Failed to check seeder data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"exists": exists,
		"message": func() string {
			if exists {
				return "Data seeder sudah tersedia"
			}
			return "Data seeder belum tersedia"
		}(),
	})
}
