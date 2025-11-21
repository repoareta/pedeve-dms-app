package audit

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

const (
	LogTypeUserAction     = "user_action"
	LogTypeTechnicalError = "technical_error"
)

// Instance audit logger global
var auditLogger *repository.AuditLogger

// InitAuditLogger menginisialisasi audit logger
func InitAuditLogger() {
	zapLog := logger.GetLogger()
	auditLogger = repository.NewAuditLogger(database.GetDB())
	
	// Auto migrate tabel audit log
	if err := database.GetDB().AutoMigrate(&domain.AuditLog{}); err != nil {
		zapLog.Error("Error migrating audit log table", zap.Error(err))
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

// Constants untuk resource types
const (
	ResourceUser     = "user"
	ResourceDocument = "document"
	ResourceAuth     = "auth"
)

// Constants untuk status
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)

