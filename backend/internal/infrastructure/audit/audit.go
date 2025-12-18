package audit

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

const (
	LogTypeUserAction     = "user_action"
	LogTypeTechnicalError = "technical_error"
)

// Instance audit logger global
var auditLogger *repository.AuditLogger

// Permanent resources - data ini disimpan permanent (tidak ada retention/deletion)
// Note: Actual check dilakukan di repository untuk avoid import cycle
var PermanentResources = []string{
	ResourceReport,          // Report Management
	ResourceFinancialReport, // Financial Report (RKAP & Realisasi)
	ResourceDocument,        // Document Management
	ResourceCompany,         // Subsidiary
	ResourceUser,            // User Management
}

// InitAuditLogger menginisialisasi audit logger
func InitAuditLogger() {
	zapLog := logger.GetLogger()
	auditLogger = repository.NewAuditLogger(database.GetDB())

	// Auto migrate tabel audit log (regular, dengan retention)
	if err := database.GetDB().AutoMigrate(&domain.AuditLog{}); err != nil {
		zapLog.Error("Error migrating audit log table", zap.Error(err))
		return
	}

	// Auto migrate tabel user activity log (permanent, tanpa retention)
	if err := database.GetDB().AutoMigrate(&domain.UserActivityLog{}); err != nil {
		zapLog.Error("Error migrating user activity log table", zap.Error(err))
		return
	}

	// Buat composite index untuk performa cleanup query (created_at + log_type)
	// Index ini akan mempercepat query cleanup yang sering memfilter berdasarkan waktu dan tipe log
	if database.GetDB().Migrator().HasIndex(&domain.AuditLog{}, "idx_audit_logs_created_at_log_type") {
		// Index sudah ada, skip
	} else {
		if err := database.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_log_type ON audit_logs(created_at, log_type)").Error; err != nil {
			zapLog.Warn("Failed to create composite index for audit logs", zap.Error(err))
			// Tidak fatal, lanjutkan saja
		} else {
			zapLog.Info("Composite index created for audit logs", zap.String("index", "idx_audit_logs_created_at_log_type"))
		}
	}

	// Buat index untuk user activity logs
	if database.GetDB().Migrator().HasIndex(&domain.UserActivityLog{}, "idx_user_activity_logs_created_at") {
		// Index sudah ada, skip
	} else {
		if err := database.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at ON user_activity_logs(created_at)").Error; err != nil {
			zapLog.Warn("Failed to create index for user activity logs", zap.Error(err))
		} else {
			zapLog.Info("Index created for user activity logs", zap.String("index", "idx_user_activity_logs_created_at"))
		}
	}

	zapLog.Info("Audit logger initialized",
		zap.Strings("permanent_resources", PermanentResources),
		zap.String("message", "Permanent resources will be stored in user_activity_logs table without retention policy"),
	)
}

// LogAction adalah fungsi helper untuk mencatat aksi
func LogAction(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	if auditLogger != nil {
		go func() {
			_ = auditLogger.Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status, details)
		}()
	}
}

// Constants untuk action types
const (
	// Authentication actions
	ActionLogin         = "login"
	ActionLogout        = "logout"
	ActionRegister      = "register"
	ActionFailedLogin   = "failed_login"
	ActionPasswordReset = "password_reset"

	// Generic CRUD actions (bisa digunakan untuk semua resource)
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionView   = "view"

	// User Management actions
	ActionCreateUser = "create_user"
	ActionUpdateUser = "update_user"
	ActionDeleteUser = "delete_user"

	// Company/Subsidiary actions
	ActionCreateCompany = "create_company"
	ActionUpdateCompany = "update_company"
	ActionDeleteCompany = "delete_company"

	// Document actions
	ActionCreateDoc = "create_document"
	ActionUpdateDoc = "update_document"
	ActionDeleteDoc = "delete_document"
	ActionViewDoc   = "view_document"

	// File Management actions (untuk modul File Management)
	ActionCreateFile   = "create_file"
	ActionUpdateFile   = "update_file"
	ActionDeleteFile   = "delete_file"
	ActionDownloadFile = "download_file"
	ActionViewFile     = "view_file"
	ActionUploadFile   = "upload_file"

	// Report Management actions (untuk modul Report Management)
	ActionGenerateReport = "generate_report"
	ActionViewReport     = "view_report"
	ActionExportReport   = "export_report"
	ActionDeleteReport   = "delete_report"

	// 2FA actions
	ActionEnable2FA  = "enable_2fa"
	ActionDisable2FA = "disable_2fa"

	// Notification actions
	ActionMarkNotificationRead     = "mark_notification_read"
	ActionMarkAllNotificationsRead = "mark_all_notifications_read"
)

// Constants untuk resource types
const (
	ResourceUser            = "user"
	ResourceDocument        = "document"
	ResourceAuth            = "auth"
	ResourceCompany         = "company"
	ResourceRole            = "role"
	ResourcePermission      = "permission"
	ResourceFile            = "file"             // Untuk modul File Management
	ResourceReport          = "report"           // Untuk modul Report Management
	ResourceFinancialReport = "financial_report" // Untuk modul Financial Report (RKAP & Realisasi)
	ResourceNotification    = "notification"     // Untuk modul Notification
)

// Constants untuk status
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)
