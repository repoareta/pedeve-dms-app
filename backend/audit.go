package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuditLog merepresentasikan entry audit log
type AuditLog struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index" json:"user_id"`           // Opsional untuk system errors
	Username   string    `gorm:"index" json:"username"`          // Opsional untuk system errors
	Action     string    `gorm:"index;not null" json:"action"`   // contoh: "login", "create_document", "system_error"
	Resource   string    `gorm:"index" json:"resource"`          // contoh: "user", "document", "system"
	ResourceID string    `gorm:"index" json:"resource_id"`       // ID dari resource yang terpengaruh
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `gorm:"type:text" json:"details"`       // String JSON dengan detail tambahan (bisa besar untuk stack trace)
	Status     string    `gorm:"index" json:"status"`            // "success", "failure", "error"
	LogType    string    `gorm:"index;default:'user_action'" json:"log_type"` // "user_action" atau "technical_error"
	CreatedAt  time.Time `gorm:"index" json:"created_at"`        // Index untuk cleanup dan query berdasarkan waktu
}

// TableName menentukan nama tabel untuk AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogger menangani audit logging
type AuditLogger struct {
	db *gorm.DB
}

// NewAuditLogger membuat audit logger baru
func NewAuditLogger(db *gorm.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// Log membuat entry audit log
func (al *AuditLogger) Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) error {
	detailsJSON := ""
	if details != nil {
		jsonData, err := json.Marshal(details)
		if err == nil {
			detailsJSON = string(jsonData)
		}
	}

	// Tentukan tipe log berdasarkan aksi
	logType := LogTypeUserAction
	if action == "system_error" || action == "database_error" || action == "validation_error" || action == "panic" {
		logType = LogTypeTechnicalError
	}

	auditLog := AuditLog{
		ID:         GenerateUUID(),
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    detailsJSON,
		Status:     status,
		LogType:    logType,
		CreatedAt:  time.Now(),
	}

	return al.db.Create(&auditLog).Error
}

// Instance audit logger global
var auditLogger *AuditLogger

// InitAuditLogger menginisialisasi audit logger
func InitAuditLogger() {
	auditLogger = NewAuditLogger(DB)
	// Auto migrate tabel audit log
	if err := DB.AutoMigrate(&AuditLog{}); err != nil {
		log.Printf("Error migrating audit log table: %v", err)
		return
	}

	// Buat composite index untuk performa cleanup query (created_at + log_type)
	// Index ini akan mempercepat query cleanup yang sering memfilter berdasarkan waktu dan tipe log
	if DB.Migrator().HasIndex(&AuditLog{}, "idx_audit_logs_created_at_log_type") {
		// Index sudah ada, skip
	} else {
		if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_log_type ON audit_logs(created_at, log_type)").Error; err != nil {
			log.Printf("Warning: Failed to create composite index for audit logs: %v", err)
			// Tidak fatal, lanjutkan saja
		} else {
			log.Println("Composite index created for audit logs (created_at, log_type)")
		}
	}
}

// LogAction adalah fungsi helper untuk mencatat aksi
func LogAction(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	if auditLogger != nil {
		go func() {
			_ = auditLogger.Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status, details)
		}()
	}
}

// GetAuditLogs mengambil audit logs dengan filter
func GetAuditLogs(userID, action, resource, status, logType string, limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := DB.Model(&AuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}

	// Ambil total count
	query.Count(&total)

	// Ambil logs dengan pagination
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

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
// @Success      200       {object}  map[string]interface{}
// @Failure      401       {object}  ErrorResponse
// @Router       /api/v1/audit-logs [get]
func GetAuditLogsHandler(c *fiber.Ctx) error {
	// Ambil user saat ini dari locals
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
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
	var currentUser UserModel
	if err := DB.First(&currentUser, "id = ?", currentUserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
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
	logs, total, err := GetAuditLogs(filterUserID, action, resource, status, logType, pageSize, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
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

// Tipe aksi audit umum
const (
	ActionLogin        = "login"
	ActionLogout       = "logout"
	ActionRegister     = "register"
	ActionCreateUser   = "create_user"
	ActionUpdateUser   = "update_user"
	ActionDeleteUser   = "delete_user"
	ActionCreateDoc    = "create_document"
	ActionUpdateDoc    = "update_document"
	ActionDeleteDoc    = "delete_document"
	ActionViewDoc      = "view_document"
	ActionEnable2FA    = "enable_2fa"
	ActionDisable2FA   = "disable_2fa"
	ActionFailedLogin  = "failed_login"
	ActionPasswordReset = "password_reset"
)

// Tipe resource umum
const (
	ResourceUser     = "user"
	ResourceDocument = "document"
	ResourceAuth     = "auth"
)

// Nilai status umum
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)
