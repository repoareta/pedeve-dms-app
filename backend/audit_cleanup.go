package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/render"
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
func getRetentionDays(logType string) int {
	var envKey string
	var defaultDays int

	if logType == LogTypeTechnicalError {
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
		log.Printf("Invalid %s value, using default: %d days", envKey, defaultDays)
		return defaultDays
	}

	return retentionDays
}

// CleanupOldAuditLogs menghapus audit logs yang sudah melewati retention period
func CleanupOldAuditLogs() error {
	// Cleanup user actions
	userActionRetention := getRetentionDays(LogTypeUserAction)
	cutoffDate := time.Now().AddDate(0, 0, -userActionRetention)
	
	result := DB.Where("log_type = ? AND created_at < ?", LogTypeUserAction, cutoffDate).Delete(&AuditLog{})
	if result.Error != nil {
		log.Printf("Error cleaning up user action logs: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d user action logs older than %d days", result.RowsAffected, userActionRetention)
	}

	// Cleanup technical errors
	technicalErrorRetention := getRetentionDays(LogTypeTechnicalError)
	cutoffDate = time.Now().AddDate(0, 0, -technicalErrorRetention)
	
	result = DB.Where("log_type = ? AND created_at < ?", LogTypeTechnicalError, cutoffDate).Delete(&AuditLog{})
	if result.Error != nil {
		log.Printf("Error cleaning up technical error logs: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d technical error logs older than %d days", result.RowsAffected, technicalErrorRetention)
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
		time.Sleep(1 * time.Hour)
		log.Println("Starting initial audit log cleanup...")
		if err := CleanupOldAuditLogs(); err != nil {
			log.Printf("Initial audit log cleanup failed: %v", err)
		}
	}()

	// Jalankan cleanup secara berkala
	go func() {
		for range ticker.C {
			log.Println("Running scheduled audit log cleanup...")
			if err := CleanupOldAuditLogs(); err != nil {
				log.Printf("Scheduled audit log cleanup failed: %v", err)
			}
		}
	}()
}

// GetAuditLogStats mengembalikan statistik audit logs
func GetAuditLogStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total records
	var totalCount int64
	DB.Model(&AuditLog{}).Count(&totalCount)
	stats["total_records"] = totalCount

	// Count by log type
	var userActionCount int64
	DB.Model(&AuditLog{}).Where("log_type = ?", LogTypeUserAction).Count(&userActionCount)
	stats["user_action_count"] = userActionCount

	var technicalErrorCount int64
	DB.Model(&AuditLog{}).Where("log_type = ?", LogTypeTechnicalError).Count(&technicalErrorCount)
	stats["technical_error_count"] = technicalErrorCount

	// Oldest record
	var oldestLog AuditLog
	DB.Order("created_at ASC").First(&oldestLog)
	if oldestLog.ID != "" {
		stats["oldest_record_date"] = oldestLog.CreatedAt
	}

	// Newest record
	var newestLog AuditLog
	DB.Order("created_at DESC").First(&newestLog)
	if newestLog.ID != "" {
		stats["newest_record_date"] = newestLog.CreatedAt
	}

	// Estimated size (rough calculation: average 500 bytes per record)
	estimatedSizeMB := float64(totalCount) * 0.0005
	stats["estimated_size_mb"] = estimatedSizeMB

	return stats, nil
}

// GetAuditLogStatsHandler menangani request GET untuk statistik audit logs
// @Summary      Get audit log statistics
// @Description  Get statistics about audit logs (total records, counts by type, estimated size, etc.)
// @Tags         Audit
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/audit-logs/stats [get]
func GetAuditLogStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := GetAuditLogStats()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get audit log statistics",
		})
		return
	}

	// Tambahkan informasi retention policy
	userActionRetention := getRetentionDays(LogTypeUserAction)
	technicalErrorRetention := getRetentionDays(LogTypeTechnicalError)
	
	stats["retention_policy"] = map[string]interface{}{
		"user_action_days":    userActionRetention,
		"technical_error_days": technicalErrorRetention,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, stats)
}

