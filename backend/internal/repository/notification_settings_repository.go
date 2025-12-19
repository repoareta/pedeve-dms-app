package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"gorm.io/gorm"
)

// NotificationSettingsRepository interface untuk notification settings operations
type NotificationSettingsRepository interface {
	GetByUserID(userID string) (*domain.NotificationSettingsModel, error)
	Create(settings *domain.NotificationSettingsModel) error
	Update(settings *domain.NotificationSettingsModel) error
	GetOrCreate(userID string) (*domain.NotificationSettingsModel, error)
}

type notificationSettingsRepository struct {
	db *gorm.DB
}

// NewNotificationSettingsRepository creates a new notification settings repository
func NewNotificationSettingsRepository() NotificationSettingsRepository {
	return NewNotificationSettingsRepositoryWithDB(database.GetDB())
}

// NewNotificationSettingsRepositoryWithDB creates a new notification settings repository with injected DB (for testing)
func NewNotificationSettingsRepositoryWithDB(db *gorm.DB) NotificationSettingsRepository {
	return &notificationSettingsRepository{db: db}
}

func (r *notificationSettingsRepository) GetByUserID(userID string) (*domain.NotificationSettingsModel, error) {
	var settings domain.NotificationSettingsModel
	// Query dengan First untuk mendapatkan record pertama yang match
	// PENTING: Query langsung tanpa Model() untuk memastikan nilai dari database ter-load dengan benar
	// Default value GORM hanya digunakan saat INSERT, bukan saat SELECT
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		// Return error as-is, termasuk gorm.ErrRecordNotFound
		return nil, err
	}
	return &settings, nil
}

func (r *notificationSettingsRepository) Create(settings *domain.NotificationSettingsModel) error {
	// Create settings - GORM akan menyimpan semua field termasuk false values
	// Default value hanya digunakan jika field tidak diset (zero value)
	return r.db.Create(settings).Error
}

func (r *notificationSettingsRepository) Update(settings *domain.NotificationSettingsModel) error {
	// Gunakan Updates dengan where clause untuk memastikan hanya update, tidak create
	// Ini mencegah duplicate key error jika ada race condition
	// UpdatedAt akan di-update otomatis oleh GORM jika menggunakan Updates
	return r.db.Model(&domain.NotificationSettingsModel{}).
		Where("user_id = ?", settings.UserID).
		Updates(map[string]interface{}{
			"email_enabled":         settings.EmailEnabled,
			"in_app_enabled":        settings.InAppEnabled,
			"expiry_threshold_days": settings.ExpiryThresholdDays,
		}).Error
}

// GetOrCreate mendapatkan settings atau membuat baru jika belum ada (dengan default values)
func (r *notificationSettingsRepository) GetOrCreate(userID string) (*domain.NotificationSettingsModel, error) {
	// Coba get settings yang sudah ada
	settings, err := r.GetByUserID(userID)
	if err == nil {
		// Settings ditemukan, return langsung
		return settings, nil
	}

	// Jika tidak ditemukan, buat baru dengan default values
	if err == gorm.ErrRecordNotFound {
		// Generate ID sebelum create untuk menghindari duplicate key error
		settings = &domain.NotificationSettingsModel{
			ID:                  uuid.GenerateUUID(), // Generate UUID untuk primary key
			UserID:              userID,
			EmailEnabled:        true,
			InAppEnabled:        true,
			ExpiryThresholdDays: 14, // Default: 14 hari
		}
		if err := r.Create(settings); err != nil {
			// Jika create gagal karena duplicate (race condition), coba get lagi
			if existingSettings, getErr := r.GetByUserID(userID); getErr == nil {
				return existingSettings, nil
			}
			return nil, err
		}
		return settings, nil
	}

	// Error selain ErrRecordNotFound
	return nil, err
}
