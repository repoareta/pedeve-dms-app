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

// StartNotificationScheduler memulai background scheduler untuk check expiring documents dan director terms
// Default threshold: 30 hari (bisa diubah via environment variable NOTIFICATION_EXPIRY_THRESHOLD_DAYS)
func StartNotificationScheduler() {
	zapLog := logger.GetLogger()
	notificationUC := NewNotificationUseCase()

	// Get threshold from environment variable (default: 30 days)
	thresholdDays := 30
	thresholdStr := os.Getenv("NOTIFICATION_EXPIRY_THRESHOLD_DAYS")
	if thresholdStr != "" {
		if parsed, err := strconv.Atoi(thresholdStr); err == nil && parsed > 0 {
			thresholdDays = parsed
		}
	}

	// Jalankan check setiap 24 jam (sekali sehari)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Jalankan check pertama kali setelah 5 menit (memberi waktu untuk server startup)
	go func() {
		time.Sleep(5 * time.Minute)
		zapLog.Info("Starting initial notification expiry check", zap.Int("threshold_days", thresholdDays))
		
		// Check expiring documents
		docNotifs, docFound, err := notificationUC.CheckExpiringDocuments(thresholdDays)
		if err != nil {
			zapLog.Error("Initial expiring documents check failed", zap.Error(err))
		} else {
			zapLog.Info("Initial expiring documents check completed", zap.Int("documents_found", docFound), zap.Int("notifications_created", docNotifs))
		}
		
		// Check expiring director terms
		dirNotifs, dirFound, err := notificationUC.CheckExpiringDirectorTerms(thresholdDays)
		if err != nil {
			zapLog.Error("Initial director term expiry check failed", zap.Error(err))
		} else {
			zapLog.Info("Initial expiring director terms check completed", zap.Int("directors_found", dirFound), zap.Int("notifications_created", dirNotifs))
		}
	}()

	// Jalankan check secara berkala (sekali sehari)
	go func() {
		for range ticker.C {
			zapLog.Info("Running scheduled notification expiry check", zap.Int("threshold_days", thresholdDays))
			
			// Check expiring documents
			docNotifs, docFound, err := notificationUC.CheckExpiringDocuments(thresholdDays)
			if err != nil {
				zapLog.Error("Scheduled expiring documents check failed", zap.Error(err))
			} else {
				zapLog.Info("Scheduled expiring documents check completed", zap.Int("documents_found", docFound), zap.Int("notifications_created", docNotifs))
			}
			
			// Check expiring director terms
			dirNotifs, dirFound, err := notificationUC.CheckExpiringDirectorTerms(thresholdDays)
			if err != nil {
				zapLog.Error("Scheduled expiring director terms check failed", zap.Error(err))
			} else {
				zapLog.Info("Scheduled expiring director terms check completed", zap.Int("directors_found", dirFound), zap.Int("notifications_created", dirNotifs))
			}
		}
	}()
}

