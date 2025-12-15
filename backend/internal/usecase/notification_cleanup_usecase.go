package usecase

import (
	"os"
	"strconv"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// Konstanta untuk retention policy notifikasi
const (
	// Default retention period (dalam hari)
	// Notifikasi: 30 hari (1 bulan) - cukup untuk kebutuhan internal company
	DefaultNotificationRetentionDays = 30
)

// GetNotificationRetentionDays mengambil periode retention dari environment variable
func GetNotificationRetentionDays() int {
	envKey := "NOTIFICATION_RETENTION_DAYS"
	retentionStr := os.Getenv(envKey)
	if retentionStr == "" {
		return DefaultNotificationRetentionDays
	}

	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil || retentionDays < 0 {
		logger.GetLogger().Warn("Invalid notification retention days value, using default",
			zap.String("env_key", envKey),
			zap.Int("default_days", DefaultNotificationRetentionDays),
			zap.Error(err),
		)
		return DefaultNotificationRetentionDays
	}

	return retentionDays
}

// CleanupOldNotifications menghapus notifikasi yang sudah melewati retention period
func CleanupOldNotifications() error {
	zapLog := logger.GetLogger()
	notifRepo := repository.NewNotificationRepository()

	retentionDays := GetNotificationRetentionDays()
	
	err := notifRepo.DeleteOldNotifications(retentionDays)
	if err != nil {
		zapLog.Error("Error cleaning up old notifications", zap.Error(err))
		return err
	}

	zapLog.Info("Notification cleanup completed",
		zap.Int("retention_days", retentionDays),
	)

	return nil
}

// StartNotificationCleanup memulai background cleanup job untuk notifikasi
func StartNotificationCleanup() {
	zapLog := logger.GetLogger()
	
	// Jalankan cleanup setiap 24 jam
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Jalankan cleanup pertama kali setelah 1 jam
	go func() {
		time.Sleep(1 * time.Hour)
		zapLog.Info("Starting initial notification cleanup")
		if err := CleanupOldNotifications(); err != nil {
			zapLog.Error("Initial notification cleanup failed", zap.Error(err))
		}
	}()

	// Jalankan cleanup secara berkala
	go func() {
		for range ticker.C {
			zapLog.Info("Running scheduled notification cleanup")
			if err := CleanupOldNotifications(); err != nil {
				zapLog.Error("Scheduled notification cleanup failed", zap.Error(err))
			}
		}
	}()
}

