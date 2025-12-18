package repository

import (
	"encoding/json"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"gorm.io/gorm"
)

// Permanent resources - data ini disimpan permanent (tidak ada retention/deletion)
var PermanentResources = []string{
	"report",           // Report Management
	"financial_report", // Financial Report (RKAP & Realisasi)
	"document",         // Document Management
	"company",          // Subsidiary
	"user",             // User Management
	"notification",     // Notification Management
}

// IsPermanentResource mengecek apakah resource termasuk permanent (tidak akan dihapus)
func IsPermanentResource(resource string) bool {
	for _, r := range PermanentResources {
		if r == resource {
			return true
		}
	}
	return false
}

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
// Jika resource termasuk permanent (report, document, company, user), akan disimpan di user_activity_logs
// Jika tidak, akan disimpan di audit_logs dengan retention policy
func (al *AuditLogger) Log(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) error {
	detailsJSON := ""
	if details != nil {
		jsonData, err := json.Marshal(details)
		if err == nil {
			detailsJSON = string(jsonData)
		}
	}

	// Cek apakah resource termasuk permanent (tidak akan dihapus)
	isPermanent := IsPermanentResource(resource)

	// Tentukan tipe log berdasarkan aksi (hanya untuk audit_logs, bukan user_activity_logs)
	logType := "user_action"
	if action == "system_error" || action == "database_error" || action == "validation_error" || action == "panic" {
		logType = "technical_error"
	}

	// Jika permanent resource, simpan di user_activity_logs (tidak ada retention)
	if isPermanent {
		activityLog := domain.UserActivityLog{
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
			CreatedAt:  time.Now(),
		}
		return al.db.Create(&activityLog).Error
	}

	// Jika bukan permanent, simpan di audit_logs (dengan retention policy)
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

// GetUserActivityLogs mengambil user activity logs (permanent) dengan filter dan pagination
func GetUserActivityLogs(userID, action, resource, resourceID, status string, limit, offset int) ([]domain.UserActivityLog, int64, error) {
	var logs []domain.UserActivityLog
	var total int64

	query := database.GetDB().Model(&domain.UserActivityLog{})

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

	// Filter berdasarkan resource ID (jika diberikan)
	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}

	// Filter berdasarkan status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Ambil total count
	query.Count(&total)

	// Ambil logs dengan pagination
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}
