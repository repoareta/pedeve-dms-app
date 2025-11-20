package main

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index;not null" json:"user_id"`
	Username   string    `gorm:"index" json:"username"`
	Action     string    `gorm:"index;not null" json:"action"` // e.g., "login", "create_document", "delete_user"
	Resource   string    `gorm:"index" json:"resource"`         // e.g., "user", "document"
	ResourceID string    `gorm:"index" json:"resource_id"`      // ID of the affected resource
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `gorm:"type:text" json:"details"` // JSON string with additional details
	Status     string    `gorm:"index" json:"status"`      // "success", "failure", "error"
	CreatedAt  time.Time `json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogger handles audit logging
type AuditLogger struct {
	db *gorm.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *gorm.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// Log creates an audit log entry
func (al *AuditLogger) Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) error {
	detailsJSON := ""
	if details != nil {
		jsonData, err := json.Marshal(details)
		if err == nil {
			detailsJSON = string(jsonData)
		}
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
		CreatedAt:  time.Now(),
	}

	return al.db.Create(&auditLog).Error
}

// Global audit logger instance
var auditLogger *AuditLogger

// InitAuditLogger initializes the audit logger
func InitAuditLogger() {
	auditLogger = NewAuditLogger(DB)
	// Auto migrate audit log table
	DB.AutoMigrate(&AuditLog{})
}

// LogAction is a helper function to log actions
func LogAction(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	if auditLogger != nil {
		go func() {
			_ = auditLogger.Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status, details)
		}()
	}
}

// GetAuditLogs retrieves audit logs with filters
func GetAuditLogs(userID, action, resource, status string, limit, offset int) ([]AuditLog, int64, error) {
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

	// Get total count
	query.Count(&total)

	// Get logs with pagination
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

// Common audit action types
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

// Common resource types
const (
	ResourceUser     = "user"
	ResourceDocument = "document"
	ResourceAuth     = "auth"
)

// Common status values
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)

