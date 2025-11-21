package repository

import (
	"encoding/json"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"gorm.io/gorm"
)

// GetAuditLogs mengambil audit logs dengan filter dan pagination
func GetAuditLogs(userID, action, resource, status, logType string, limit, offset int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := database.GetDB().Model(&domain.AuditLog{})

	// Filter berdasarkan user ID (jika diberikan)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Filter berdasarkan action
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Filter berdasarkan resource
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// Filter berdasarkan status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter berdasarkan log type
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}

	// Ambil total count
	query.Count(&total)

	// Ambil logs dengan pagination
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

// CreateAuditLog membuat audit log baru
func CreateAuditLog(log *domain.AuditLog) error {
	return database.GetDB().Create(log).Error
}

// AuditLogger adalah struct untuk audit logger
type AuditLogger struct {
	db *gorm.DB
}

// NewAuditLogger membuat instance baru AuditLogger
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
	logType := "user_action"
	if action == "system_error" || action == "database_error" || action == "validation_error" || action == "panic" {
		logType = "technical_error"
	}

	auditLog := domain.AuditLog{
		ID:         uuid.GenerateUUID(),
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

