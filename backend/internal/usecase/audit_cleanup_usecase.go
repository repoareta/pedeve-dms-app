package usecase

import (
	"os"
	"strconv"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/audit"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// Konstanta untuk retention policy
const (
	// Default retention period (dalam hari)
	// User actions: 90 hari (3 bulan)
	// Technical errors: 30 hari (1 bulan)
	DefaultUserActionRetentionDays    = 90
	DefaultTechnicalErrorRetentionDays = 30
)

// getRetentionDays mengambil periode retention dari environment variable
func GetRetentionDays(logType string) int {
	var envKey string
	var defaultDays int

	if logType == audit.LogTypeTechnicalError {
		envKey = "AUDIT_LOG_TECHNICAL_ERROR_RETENTION_DAYS"
		defaultDays = DefaultTechnicalErrorRetentionDays
	} else {
		envKey = "AUDIT_LOG_USER_ACTION_RETENTION_DAYS"
		defaultDays = DefaultUserActionRetentionDays
	}

	retentionStr := os.Getenv(envKey)
	if retentionStr == "" {
		return defaultDays
	}

	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil || retentionDays < 0 {
		logger.GetLogger().Warn("Invalid retention days value, using default",
			zap.String("env_key", envKey),
			zap.Int("default_days", defaultDays),
			zap.Error(err),
		)
		return defaultDays
	}

	return retentionDays
}

// CleanupOldAuditLogs menghapus audit logs yang sudah melewati retention period
func CleanupOldAuditLogs() error {
	zapLog := logger.GetLogger()
	
	// Cleanup user actions
	userActionRetention := GetRetentionDays(audit.LogTypeUserAction)
	cutoffDate := time.Now().AddDate(0, 0, -userActionRetention)
	
	result := database.GetDB().Where("log_type = ? AND created_at < ?", audit.LogTypeUserAction, cutoffDate).Delete(&domain.AuditLog{})
	if result.Error != nil {
		zapLog.Error("Error cleaning up user action logs", zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected > 0 {
		zapLog.Info("Cleaned up user action logs",
			zap.Int64("count", result.RowsAffected),
			zap.Int("retention_days", userActionRetention),
		)
	}

	// Cleanup technical errors
	technicalErrorRetention := GetRetentionDays(audit.LogTypeTechnicalError)
	cutoffDate = time.Now().AddDate(0, 0, -technicalErrorRetention)
	
	result = database.GetDB().Where("log_type = ? AND created_at < ?", audit.LogTypeTechnicalError, cutoffDate).Delete(&domain.AuditLog{})
	if result.Error != nil {
		zapLog.Error("Error cleaning up technical error logs", zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected > 0 {
		zapLog.Info("Cleaned up technical error logs",
			zap.Int64("count", result.RowsAffected),
			zap.Int("retention_days", technicalErrorRetention),
		)
	}

	return nil
}

// StartAuditLogCleanup memulai background cleanup job untuk audit logs
func StartAuditLogCleanup() {
	// Jalankan cleanup setiap 24 jam
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Jalankan cleanup pertama kali setelah 1 jam
	go func() {
		zapLog := logger.GetLogger()
		time.Sleep(1 * time.Hour)
		zapLog.Info("Starting initial audit log cleanup")
		if err := CleanupOldAuditLogs(); err != nil {
			zapLog.Error("Initial audit log cleanup failed", zap.Error(err))
		}
	}()

	// Jalankan cleanup secara berkala
	go func() {
		zapLog := logger.GetLogger()
		for range ticker.C {
			zapLog.Info("Running scheduled audit log cleanup")
			if err := CleanupOldAuditLogs(); err != nil {
				zapLog.Error("Scheduled audit log cleanup failed", zap.Error(err))
			}
		}
	}()
}

// GetAuditLogStats mengembalikan statistik audit logs
func GetAuditLogStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total records
	var totalCount int64
	database.GetDB().Model(&domain.AuditLog{}).Count(&totalCount)
	stats["total_records"] = totalCount

	// Count by log type
	var userActionCount int64
	database.GetDB().Model(&domain.AuditLog{}).Where("log_type = ?", audit.LogTypeUserAction).Count(&userActionCount)
	stats["user_action_count"] = userActionCount

	var technicalErrorCount int64
	database.GetDB().Model(&domain.AuditLog{}).Where("log_type = ?", audit.LogTypeTechnicalError).Count(&technicalErrorCount)
	stats["technical_error_count"] = technicalErrorCount

	// Oldest record
	var oldestLog domain.AuditLog
	database.GetDB().Order("created_at ASC").First(&oldestLog)
	if oldestLog.ID != "" {
		stats["oldest_record_date"] = oldestLog.CreatedAt
	}

	// Newest record
	var newestLog domain.AuditLog
	database.GetDB().Order("created_at DESC").First(&newestLog)
	if newestLog.ID != "" {
		stats["newest_record_date"] = newestLog.CreatedAt
	}

	// Estimated size (rough calculation: average 500 bytes per record)
	estimatedSizeMB := float64(totalCount) * 0.0005
	stats["estimated_size_mb"] = estimatedSizeMB

	return stats, nil
}

