package http

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
)

type DevelopmentHandler struct {
	devUseCase      usecase.DevelopmentUseCase
	notificationUC  usecase.NotificationUseCase
	logger          *zap.Logger
}

func NewDevelopmentHandler(devUseCase usecase.DevelopmentUseCase) *DevelopmentHandler {
	return &DevelopmentHandler{
		devUseCase:     devUseCase,
		notificationUC: usecase.NewNotificationUseCase(),
		logger:         logger.GetLogger(),
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

// ResetReportData handles resetting all report data
// @Summary      Reset Data Reports
// @Description  Menghapus semua data reports (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/reset-reports [post]
func (h *DevelopmentHandler) ResetReportData(c *fiber.Ctx) error {
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
		h.logger.Warn("Non-superadmin attempted to reset report data",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Reset report data
	if err := h.devUseCase.ResetReportData(); err != nil {
		h.logger.Error("Failed to reset report data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "reset_report_data", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data reports berhasil di-reset",
		"success": true,
	})
}

// RunReportSeeder handles running the report seeder
// @Summary      Jalankan Seeder Data Reports
// @Description  Menjalankan seeder untuk membuat sample data reports (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      409  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/run-report-seeder [post]
func (h *DevelopmentHandler) RunReportSeeder(c *fiber.Ctx) error {
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
		h.logger.Warn("Non-superadmin attempted to run report seeder",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Run seeder
	err := h.devUseCase.RunReportSeeder()
	if err != nil {
		if err.Error() == "report data already exists" {
			// Audit log
			userID := c.Locals("userID").(string)
			username := c.Locals("username").(string)
			audit.LogAction(userID, username, "run_report_seeder", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusFailure, map[string]interface{}{
				"reason": "Data seeder sudah ada",
			})

			return c.Status(fiber.StatusConflict).JSON(domain.ErrorResponse{
				Error:   "seeder_already_exists",
				Message: "Data reports sudah tersedia. Proses dibatalkan untuk mencegah duplikasi data.",
			})
		}

		h.logger.Error("Failed to run report seeder", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "seeder_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "run_report_seeder", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Seeder data reports berhasil dijalankan",
		"success": true,
	})
}

// CheckReportDataExists checks if report data already exists
// @Summary      Cek Status Seeder Data Reports
// @Description  Mengecek apakah data seeder reports sudah ada (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-report-status [get]
func (h *DevelopmentHandler) CheckReportDataExists(c *fiber.Ctx) error {
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

	// Check report data
	exists, err := h.devUseCase.CheckReportDataExists()
	if err != nil {
		h.logger.Error("Failed to check report data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"exists": exists,
		"message": func() string {
			if exists {
				return "Data reports sudah tersedia"
			}
			return "Data reports belum tersedia"
		}(),
	})
}

// RunAllSeeders handles running all seeders in order
// @Summary      Jalankan Semua Seeder Data
// @Description  Menjalankan semua seeder secara berurutan: Company -> Reports. Memastikan relasi data terjaga.
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/run-all-seeders [post]
func (h *DevelopmentHandler) RunAllSeeders(c *fiber.Ctx) error {
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
		h.logger.Warn("Non-superadmin attempted to run all seeders",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Run all seeders
	err := h.devUseCase.RunAllSeeders()
	if err != nil {
		h.logger.Error("Failed to run all seeders", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "seeder_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "run_all_seeders", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Semua seeder berhasil dijalankan",
		"success": true,
		"details": map[string]string{
			"company": "Seeder company dijalankan (atau sudah ada)",
			"report":  "Seeder report dijalankan (atau sudah ada)",
		},
	})
}

// ResetAllSeededData handles resetting all seeded data
// @Summary      Reset Semua Data Seeder
// @Description  Mereset semua data yang sudah di-seed secara berurutan: Reports -> Company. Memastikan relasi data dihapus dengan benar.
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/reset-all-seeded-data [post]
func (h *DevelopmentHandler) ResetAllSeededData(c *fiber.Ctx) error {
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
		h.logger.Warn("Non-superadmin attempted to reset all seeded data",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Reset all seeded data
	err := h.devUseCase.ResetAllSeededData()
	if err != nil {
		h.logger.Error("Failed to reset all seeded data", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "reset_all_seeded_data", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Semua data seeder berhasil di-reset",
		"success": true,
		"details": map[string]string{
			"report":  "Data reports dihapus",
			"company": "Data companies dihapus",
		},
	})
}

// ResetAllFinancialReports handles resetting all financial reports from all companies
// @Summary      Reset Semua Data Laporan Keuangan
// @Description  Menghapus semua data laporan keuangan (Financial Reports) dari semua perusahaan (hanya superadmin)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/reset-all-financial-reports [post]
func (h *DevelopmentHandler) ResetAllFinancialReports(c *fiber.Ctx) error {
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
		h.logger.Warn("Non-superadmin attempted to reset all financial reports",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin yang dapat mengakses fitur ini",
		})
	}

	// Reset all financial reports
	err := h.devUseCase.ResetAllFinancialReports()
	if err != nil {
		h.logger.Error("Failed to reset all financial reports", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "reset_all_financial_reports", "development", "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Semua data laporan keuangan berhasil di-reset",
		"success": true,
	})
}

// CheckAllSeederStatus checks the status of all seeders
// @Summary      Cek Status Semua Seeder
// @Description  Mengecek status semua seeder data (company, reports, dll)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-all-seeder-status [get]
func (h *DevelopmentHandler) CheckAllSeederStatus(c *fiber.Ctx) error {
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

	// Check all seeder status
	status, err := h.devUseCase.CheckAllSeederStatus()
	if err != nil {
		h.logger.Error("Failed to check all seeder status", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": status,
		"message": func() string {
			allExists := status["company"] && status["report"]
			if allExists {
				return "Semua data seeder sudah tersedia"
			}
			return "Beberapa data seeder belum tersedia"
		}(),
	})
}

// CreateTestNotification creates a test notification for the current user
// @Summary      Create Test Notification
// @Description  Membuat notifikasi test untuk user yang sedang login (untuk testing fitur notifikasi)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      object  true  "Notification data"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  domain.ErrorResponse
// @Failure      500      {object}  domain.ErrorResponse
// @Router       /development/create-test-notification [post]
func (h *DevelopmentHandler) CreateTestNotification(c *fiber.Ctx) error {
	// Get user ID from context
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)

	var req struct {
		Title        string  `json:"title"`
		Message      string  `json:"message"`
		Type         string  `json:"type"`
		ResourceType string  `json:"resource_type"`
		ResourceID   *string `json:"resource_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Set defaults
	if req.Type == "" {
		req.Type = "info"
	}
	if req.Title == "" {
		req.Title = "Test Notification"
	}
	if req.Message == "" {
		req.Message = "Ini adalah notifikasi test untuk memastikan fitur notifikasi berfungsi dengan baik."
	}
	if req.ResourceType == "" {
		req.ResourceType = "system"
	}

	// Create notification
	notification, err := h.notificationUC.CreateNotification(
		userID,
		req.Type,
		req.Title,
		req.Message,
		req.ResourceType,
		req.ResourceID,
	)
	if err != nil {
		h.logger.Error("Failed to create test notification", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "create_failed",
			Message: "Failed to create notification",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Test notification created successfully",
		"notification": notification,
	})
}

// CreateTestNotifications creates multiple test notifications for the current user
// @Summary      Create Test Notifications
// @Description  Membuat beberapa notifikasi test untuk user yang sedang login (untuk testing fitur notifikasi)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/create-test-notifications [post]
func (h *DevelopmentHandler) CreateTestNotifications(c *fiber.Ctx) error {
	// Get user ID from context
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}
	userID := userIDVal.(string)

	// Create multiple test notifications
	testNotifications := []struct {
		Type         string
		Title        string
		Message      string
		ResourceType string
	}{
		{
			Type:         "info",
			Title:        "Selamat Datang!",
			Message:      "Ini adalah notifikasi test pertama. Fitur notifikasi sudah berfungsi dengan baik!",
			ResourceType: "system",
		},
		{
			Type:         "warning",
			Title:        "Peringatan: Dokumen Akan Expired",
			Message:      "Dokumen 'Surat Izin Usaha' akan expired dalam 7 hari. Silakan perbarui dokumen tersebut.",
			ResourceType: "document",
		},
		{
			Type:         "success",
			Title:        "Dokumen Berhasil Diupload",
			Message:      "Dokumen 'Laporan Keuangan Q1 2024' berhasil diupload dan sudah tersedia di sistem.",
			ResourceType: "document",
		},
		{
			Type:         "info",
			Title:        "Laporan Baru Tersedia",
			Message:      "Laporan bulanan untuk periode Desember 2024 sudah tersedia. Silakan review dan approve.",
			ResourceType: "report",
		},
		{
			Type:         "warning",
			Title:        "Perhatian: Perubahan Data",
			Message:      "Data perusahaan 'PT ABC' telah diubah oleh administrator. Silakan review perubahan tersebut.",
			ResourceType: "company",
		},
	}

	var created []*domain.NotificationModel
	var errors []string

	for _, testNotif := range testNotifications {
		notification, err := h.notificationUC.CreateNotification(
			userID,
			testNotif.Type,
			testNotif.Title,
			testNotif.Message,
			testNotif.ResourceType,
			nil,
		)
		if err != nil {
			h.logger.Error("Failed to create test notification", zap.Error(err), zap.String("title", testNotif.Title))
			errors = append(errors, testNotif.Title)
		} else {
			created = append(created, notification)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Test notifications created",
		"created":  len(created),
		"failed":   len(errors),
		"errors":   errors,
		"notifications": created,
	})
}

// CheckExpiringDocuments godoc
// @Summary      Check Expiring Documents
// @Description  Trigger manual check untuk dokumen yang akan expired dan create notifications (untuk testing, superadmin dan administrator)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        threshold_days  body      int     false  "Threshold days untuk check (default: 30)"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-expiring-documents [post]
func (h *DevelopmentHandler) CheckExpiringDocuments(c *fiber.Ctx) error {
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

	if !utils.IsSuperAdminLike(roleName) {
		h.logger.Warn("Non-superadmin/administrator attempted to check expiring documents",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda harus login sebagai superadmin atau administrator!",
		})
	}

	var req struct {
		ThresholdDays int `json:"threshold_days"`
	}

	if err := c.BodyParser(&req); err != nil {
		// Default threshold jika tidak ada request body
		req.ThresholdDays = 30
	}

	if req.ThresholdDays <= 0 {
		req.ThresholdDays = 30
	}

	// Check expiring documents
	notificationsCreated, documentsFound, err := h.notificationUC.CheckExpiringDocuments(req.ThresholdDays)
	if err != nil {
		h.logger.Error("Failed to check expiring documents", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: fmt.Sprintf("Failed to check expiring documents: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":             "Expiring documents check completed",
		"threshold_days":      req.ThresholdDays,
		"documents_found":     documentsFound,
		"notifications_created": notificationsCreated,
	})
}

// CheckExpiringDirectorTerms godoc
// @Summary      Check Expiring Director Terms
// @Description  Trigger manual check untuk masa jabatan pengurus yang akan berakhir dan create notifications (untuk testing, superadmin dan administrator)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        threshold_days  body      int     false  "Threshold days untuk check (default: 30)"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-expiring-director-terms [post]
func (h *DevelopmentHandler) CheckExpiringDirectorTerms(c *fiber.Ctx) error {
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

	if !utils.IsSuperAdminLike(roleName) {
		h.logger.Warn("Non-superadmin/administrator attempted to check expiring director terms",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda harus login sebagai superadmin atau administrator!",
		})
	}

	var req struct {
		ThresholdDays int `json:"threshold_days"`
	}

	if err := c.BodyParser(&req); err != nil {
		// Default threshold jika tidak ada request body
		req.ThresholdDays = 30
	}

	if req.ThresholdDays <= 0 {
		req.ThresholdDays = 30
	}

	// Check expiring director terms
	notificationsCreated, directorsFound, err := h.notificationUC.CheckExpiringDirectorTerms(req.ThresholdDays)
	if err != nil {
		h.logger.Error("Failed to check expiring director terms", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: fmt.Sprintf("Failed to check expiring director terms: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":             "Expiring director terms check completed",
		"threshold_days":      req.ThresholdDays,
		"directors_found":     directorsFound,
		"notifications_created": notificationsCreated,
	})
}

// CheckAllExpiringNotifications godoc
// @Summary      Check All Expiring Notifications
// @Description  Trigger manual check untuk semua expiring notifications (documents dan director terms) sekaligus (untuk testing, superadmin dan administrator)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        threshold_days  body      int     false  "Threshold days untuk check (default: 30)"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/check-all-expiring-notifications [post]
func (h *DevelopmentHandler) CheckAllExpiringNotifications(c *fiber.Ctx) error {
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

	if !utils.IsSuperAdminLike(roleName) {
		h.logger.Warn("Non-superadmin/administrator attempted to check all expiring notifications",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda harus login sebagai superadmin atau administrator!",
		})
	}

	var req struct {
		ThresholdDays int `json:"threshold_days"`
	}

	if err := c.BodyParser(&req); err != nil {
		// Default threshold jika tidak ada request body
		req.ThresholdDays = 30
	}

	if req.ThresholdDays <= 0 {
		req.ThresholdDays = 30
	}

	// Check expiring documents
	docNotifications, documentsFound, docErr := h.notificationUC.CheckExpiringDocuments(req.ThresholdDays)
	if docErr != nil {
		h.logger.Error("Failed to check expiring documents", zap.Error(docErr))
	}

	// Check expiring director terms
	dirNotifications, directorsFound, dirErr := h.notificationUC.CheckExpiringDirectorTerms(req.ThresholdDays)
	if dirErr != nil {
		h.logger.Error("Failed to check expiring director terms", zap.Error(dirErr))
	}

	// Return error jika salah satu atau keduanya gagal
	if docErr != nil || dirErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "check_failed",
			Message: fmt.Sprintf("Some checks failed. Documents: %v, Director Terms: %v", docErr, dirErr),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":        "All expiring notifications check completed",
		"threshold_days": req.ThresholdDays,
		"documents": fiber.Map{
			"found":     documentsFound,
			"notifications_created": docNotifications,
		},
		"directors": fiber.Map{
			"found":     directorsFound,
			"notifications_created": dirNotifications,
		},
		"total_notifications_created": docNotifications + dirNotifications,
	})
}

// CreateNotificationForDocument godoc
// @Summary      Create Notification for Document
// @Description  Create notification langsung untuk dokumen tertentu berdasarkan document ID (untuk testing, superadmin dan administrator)
// @Tags         Development
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        document_id  body      string  true  "Document ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /development/create-notification-for-document [post]
func (h *DevelopmentHandler) CreateNotificationForDocument(c *fiber.Ctx) error {
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

	if !utils.IsSuperAdminLike(roleName) {
		h.logger.Warn("Non-superadmin/administrator attempted to create notification for document",
			zap.String("roleName", roleName),
			zap.String("username", c.Locals("username").(string)),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda harus login sebagai superadmin atau administrator!",
		})
	}

	var req struct {
		DocumentID string `json:"document_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body. document_id is required",
		})
	}

	if req.DocumentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "document_id is required",
		})
	}

	// Get document
	docRepo := repository.NewDocumentRepository()
	doc, err := docRepo.GetDocumentByID(req.DocumentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Document not found",
		})
	}

	// Log document info untuk debugging - dengan detail lengkap
	h.logger.Info("Creating notification for document",
		zap.String("document_id", doc.ID),
		zap.String("document_name", doc.Name),
		zap.String("uploader_id", doc.UploaderID),
		zap.Any("expiry_date", doc.ExpiryDate),
		zap.Bool("expiry_date_is_nil", doc.ExpiryDate == nil),
		zap.Any("metadata", doc.Metadata),
	)
	
	// Cek expiry_date dari kolom expiry_date atau dari metadata.expired_date
	var expiryDateToUse *time.Time = doc.ExpiryDate
	var expiryDateFromDB *time.Time // Declare di scope yang lebih luas untuk logging
	
	// Jika expiry_date NULL, coba ambil dari metadata
	if expiryDateToUse == nil && len(doc.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
			if expiredDateStr, ok := metadata["expired_date"].(string); ok && expiredDateStr != "" {
				// Parse expired_date dari metadata (format: "2025-12-11T02:08:09.373Z")
				if parsedDate, err := time.Parse(time.RFC3339, expiredDateStr); err == nil {
					expiryDateToUse = &parsedDate
					h.logger.Info("Found expiry date in metadata",
						zap.String("document_id", doc.ID),
						zap.String("expired_date_from_metadata", expiredDateStr),
						zap.Time("parsed_date", parsedDate),
					)
				} else {
					h.logger.Warn("Failed to parse expired_date from metadata",
						zap.String("document_id", doc.ID),
						zap.String("expired_date_str", expiredDateStr),
						zap.Error(err),
					)
				}
			}
		}
	}
	
	// Debug: Cek langsung dari database apakah expiry_date ada (fallback)
	if expiryDateToUse == nil {
		errCheck := database.GetDB().Model(&domain.DocumentModel{}).
			Select("expiry_date").
			Where("id = ?", doc.ID).
			Scan(&expiryDateFromDB).Error
		
		if errCheck == nil && expiryDateFromDB != nil {
			h.logger.Info("Found expiry date in database column",
				zap.String("document_id", doc.ID),
				zap.Time("expiry_date_from_db", *expiryDateFromDB),
			)
			expiryDateToUse = expiryDateFromDB
		}
	}

	// Calculate days until expiry (jika ada)
	var daysUntilExpiry int
	var title, message string
	
	if expiryDateToUse != nil {
		daysUntilExpiry = int(time.Until(*expiryDateToUse).Hours() / 24)
		title = fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name)
		if daysUntilExpiry < 0 {
			message = fmt.Sprintf("Dokumen '%s' sudah expired sejak %d hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.", doc.Name, -daysUntilExpiry)
		} else if daysUntilExpiry == 0 {
			message = fmt.Sprintf("Dokumen '%s' akan expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.", doc.Name)
		} else {
			message = fmt.Sprintf("Dokumen '%s' akan expired dalam %d hari. Silakan perbarui atau perpanjang dokumen tersebut.", doc.Name, daysUntilExpiry)
		}
		h.logger.Info("Document has expiry date",
			zap.String("document_id", doc.ID),
			zap.Time("expiry_date", *expiryDateToUse),
			zap.Int("days_until_expiry", daysUntilExpiry),
		)
	} else {
		// Untuk testing: create notification meskipun tidak ada expiry_date
		title = fmt.Sprintf("Notifikasi untuk Dokumen '%s'", doc.Name)
		message = fmt.Sprintf("Ini adalah notifikasi test untuk dokumen '%s'. Dokumen ini tidak memiliki tanggal berakhir.", doc.Name)
		daysUntilExpiry = 0
		h.logger.Warn("Document does not have expiry_date",
			zap.String("document_id", doc.ID),
			zap.Any("expiry_date_from_db", expiryDateFromDB),
		)
	}
	
	// Untuk testing: buat notifikasi untuk user yang sedang login (bukan uploader)
	// Ini memungkinkan superadmin untuk test notifikasi mereka sendiri
	userIDVal := c.Locals("userID")
	var targetUserID string
	if userIDVal != nil {
		targetUserID = userIDVal.(string)
		h.logger.Info("Creating notification for logged-in user (testing)",
			zap.String("logged_in_user_id", targetUserID),
			zap.String("document_uploader_id", doc.UploaderID),
		)
	} else {
		// Fallback ke uploader jika tidak ada user context
		targetUserID = doc.UploaderID
		h.logger.Warn("No user context, using document uploader",
			zap.String("document_uploader_id", doc.UploaderID),
		)
	}
	
	notification, err := h.notificationUC.CreateNotification(
		targetUserID,
		"document_expiry",
		title,
		message,
		"document",
		&doc.ID,
	)
	if err != nil {
		h.logger.Error("Failed to create notification for document", zap.Error(err), zap.String("document_id", req.DocumentID))
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "create_failed",
			Message: "Failed to create notification",
		})
	}

	// Mark document as notified (optional, untuk prevent duplicate)
	err = database.GetDB().Model(&doc).Update("expiry_notified", true).Error
	if err != nil {
		h.logger.Warn("Failed to mark document as notified", zap.Error(err), zap.String("document_id", req.DocumentID))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Notification created successfully for document",
		"notification": notification,
		"document": fiber.Map{
			"id":           doc.ID,
			"name":         doc.Name,
			"expiry_date":  doc.ExpiryDate,
			"days_until_expiry": daysUntilExpiry,
		},
	})
}
