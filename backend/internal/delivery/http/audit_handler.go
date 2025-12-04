package http

import (
	"strconv"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// GetAuditLogsHandler menangani request GET untuk audit logs (untuk Fiber)
// @Summary      Ambil Audit Logs
// @Description  Mengambil audit logs dengan pagination dan filter. Data ini memiliki retention policy: 90 hari untuk user actions, 30 hari untuk technical errors. User reguler hanya bisa melihat audit logs mereka sendiri, sedangkan admin/superadmin bisa melihat semua audit logs. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only). Catatan: Data penting (report, document, company, user) disimpan di endpoint /user-activity-logs dengan permanent storage.
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Nomor halaman (default: 1)"
// @Param        pageSize  query     int     false  "Jumlah item per halaman (default: 10, maksimal: 100)"
// @Param        action    query     string  false  "Filter berdasarkan action (contoh: login, logout, create_document)"
// @Param        resource  query     string  false  "Filter berdasarkan resource (contoh: auth, document, user, company, report)"
// @Param        status    query     string  false  "Filter berdasarkan status (success, failure, error)"
// @Param        logType   query     string  false  "Filter berdasarkan tipe log (user_action atau technical_error)"
// @Success      200       {object}  map[string]interface{}  "Audit logs berhasil diambil. Response berisi data (array audit logs), total, page, pageSize, dan totalPages"
// @Failure      401       {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      404       {object}  domain.ErrorResponse  "User tidak ditemukan di database"
// @Failure      500       {object}  domain.ErrorResponse  "Gagal mengambil audit logs"
// @Router       /api/v1/audit-logs [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Authorization: User reguler hanya melihat logs sendiri, admin/superadmin melihat semua logs
// @note         4. Retention Policy: User actions (90 hari), Technical errors (30 hari) - data akan dihapus otomatis setelah periode tersebut
// @note         5. Permanent Storage: Untuk data penting (report, document, company, user), gunakan endpoint /user-activity-logs
// @note         6. Pagination: Default page=1, pageSize=10, maksimal pageSize=100
// @note         7. Filtering: Filter dapat dikombinasikan untuk hasil yang lebih spesifik
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
	
	// CRITICAL: Admin holding hanya bisa lihat user activity logs (bukan technical logs)
	// Filter out technical_error logs untuk admin (superadmin bisa lihat semua)
	if currentUser.Role == "admin" && logType == audit.LogTypeTechnicalError {
		// Admin tidak boleh lihat technical logs, return empty
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":       []interface{}{},
			"total":      0,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": 0,
		})
	}
	
	// Jika admin tidak specify logType, default ke user_action (bukan technical_error)
	if currentUser.Role == "admin" && logType == "" {
		// Admin default hanya lihat user_action logs
		logType = audit.LogTypeUserAction
	}
	
	// CRITICAL: Admin holding hanya bisa lihat user activity logs (bukan technical logs)
	// Filter out technical_error logs untuk admin (superadmin bisa lihat semua)
	if currentUser.Role == "admin" && logType == "technical_error" {
		// Admin tidak boleh lihat technical logs, return empty
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":       []interface{}{},
			"total":      0,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": 0,
		})
	}
	
	// Jika admin tidak specify logType, default ke user_action (bukan technical_error)
	if currentUser.Role == "admin" && logType == "" {
		// Admin default hanya lihat user_action logs
		logType = audit.LogTypeUserAction
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
// @Summary      Ambil Statistik Audit Logs
// @Description  Mengambil statistik tentang audit logs termasuk total records, jumlah berdasarkan tipe, estimasi ukuran database, dan retention policy. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Statistik audit logs berhasil diambil. Response berisi total_records, user_action_count, technical_error_count, estimated_size_mb, retention_policy (user_action_days, technical_error_days), dll"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      500  {object}  domain.ErrorResponse  "Gagal mengambil statistik audit logs"
// @Router       /api/v1/audit-logs/stats [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Retention Policy: User action logs disimpan selama 90 hari, technical error logs disimpan selama 30 hari
// @note         4. Auto Cleanup: Logs yang expired akan dihapus otomatis oleh background job
// @note         5. Statistics: Statistik dihitung secara real-time dari database
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

// GetUserActivityLogsHandler menangani request GET untuk user activity logs (permanent)
// @Summary      Ambil User Activity Logs
// @Description  Mengambil user activity logs (permanent) untuk resource penting: report, document, company, user. Data ini tidak akan dihapus (permanent storage tanpa retention policy) untuk keperluan compliance dan legal. User reguler hanya bisa melihat logs mereka sendiri, sedangkan admin/superadmin bisa melihat semua logs. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Nomor halaman (default: 1)"
// @Param        pageSize  query     int     false  "Jumlah item per halaman (default: 10, maksimal: 100)"
// @Param        action    query     string  false  "Filter berdasarkan action (contoh: create_report, update_document, create_company, create_user)"
// @Param        resource  query     string  false  "Filter berdasarkan resource (report, document, company, user)"
// @Param        status    query     string  false  "Filter berdasarkan status (success, failure, error)"
// @Success      200       {object}  map[string]interface{}  "User activity logs berhasil diambil. Response berisi data (array user activity logs), total, page, pageSize, dan totalPages"
// @Failure      401       {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      404       {object}  domain.ErrorResponse  "User tidak ditemukan di database"
// @Failure      500       {object}  domain.ErrorResponse  "Gagal mengambil user activity logs"
// @Router       /api/v1/user-activity-logs [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Authorization: User reguler hanya melihat logs sendiri, admin/superadmin melihat semua logs
// @note         4. Permanent Storage: Data ini disimpan permanen tanpa retention policy untuk compliance
// @note         5. Resources: Hanya menampilkan logs untuk resource penting: report, document, company, user
// @note         6. Pagination: Default page=1, pageSize=10, maksimal pageSize=100
// @note         7. Filtering: Filter dapat dikombinasikan untuk hasil yang lebih spesifik
func GetUserActivityLogsHandler(c *fiber.Ctx) error {
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

	// User reguler hanya bisa lihat logs mereka sendiri.
	// Admin/superadmin/administrator bisa lihat semua logs, namun administrator tidak boleh melihat log milik superadmin.
	filterUserID := currentUserID
	isAdminLike := currentUser.Role == "admin" || currentUser.Role == "superadmin" || currentUser.Role == "administrator"
	isAdministrator := currentUser.Role == "administrator"
	if isAdminLike {
		filterUserID = ""
	}

	// Hitung offset
	offset := (page - 1) * pageSize

	// Ambil user activity logs
	logs, total, err := repository.GetUserActivityLogs(filterUserID, action, resource, status, pageSize, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user activity logs",
		})
	}

	// Administrator: sembunyikan log milik superadmin
	if isAdministrator {
		filtered := make([]domain.UserActivityLog, 0, len(logs))
		for _, l := range logs {
			if l.Username != "" && l.Username != "superadmin" {
				filtered = append(filtered, l)
			}
		}
		logs = filtered
		if total > int64(len(filtered)) {
			total = int64(len(filtered))
		}
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
