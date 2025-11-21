package http

import (
	"strconv"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/audit"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/repository"
	"github.com/Fajarriswandi/dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// GetAuditLogsHandler menangani request GET untuk audit logs (untuk Fiber)
// @Summary      Get audit logs
// @Description  Get audit logs with pagination and filters
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        pageSize  query     int     false  "Page size (default: 10)"
// @Param        action    query     string  false  "Filter by action"
// @Param        resource  query     string  false  "Filter by resource"
// @Param        status    query     string  false  "Filter by status"
// @Param        logType   query     string  false  "Filter by log type (user_action or technical_error)"
// @Success      200       {object}  map[string]interface{}
// @Failure      401       {object}  domain.ErrorResponse
// @Router       /api/v1/audit-logs [get]
func GetAuditLogsHandler(c *fiber.Ctx) error {
	// Ambil user saat ini dari locals
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}

	currentUserID := userIDVal.(string)

	// Parse parameter query
	page := 1
	pageSize := 10
	action := c.Query("action")
	resource := c.Query("resource")
	status := c.Query("status")
	logType := c.Query("logType") // Filter berdasarkan tipe log: "user_action" atau "technical_error"

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Ambil user dari database untuk cek role
	var currentUser domain.UserModel
	if err := database.GetDB().First(&currentUser, "id = ?", currentUserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
	}

	// User reguler hanya bisa lihat audit logs mereka sendiri
	filterUserID := currentUserID
	if currentUser.Role == "admin" || currentUser.Role == "superadmin" {
		// Admin bisa lihat semua logs, jangan filter berdasarkan userID
		filterUserID = ""
	}

	// Hitung offset
	offset := (page - 1) * pageSize

	// Ambil audit logs
	logs, total, err := repository.GetAuditLogs(filterUserID, action, resource, status, logType, pageSize, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get audit logs",
		})
	}

	// Kembalikan response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":       logs,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAuditLogStatsHandler menangani request GET untuk statistik audit logs (untuk Fiber)
// @Summary      Get audit log statistics
// @Description  Get statistics about audit logs (total records, counts by type, estimated size, etc.)
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /api/v1/audit-logs/stats [get]
func GetAuditLogStatsHandler(c *fiber.Ctx) error {
	stats, err := usecase.GetAuditLogStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get audit log statistics",
		})
	}

	// Tambahkan informasi retention policy
	userActionRetention := usecase.GetRetentionDays(audit.LogTypeUserAction)
	technicalErrorRetention := usecase.GetRetentionDays(audit.LogTypeTechnicalError)
	
	stats["retention_policy"] = fiber.Map{
		"user_action_days":     userActionRetention,
		"technical_error_days": technicalErrorRetention,
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

