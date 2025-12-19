package usecase

import (
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
)

// NotificationSettingsUseCase interface untuk notification settings operations
type NotificationSettingsUseCase interface {
	GetSettings(userID string) (*domain.NotificationSettingsModel, error)
	UpdateSettings(userID string, emailEnabled *bool, inAppEnabled *bool, expiryThresholdDays *int) (*domain.NotificationSettingsModel, error)
}

type notificationSettingsUseCase struct {
	settingsRepo repository.NotificationSettingsRepository
}

// NewNotificationSettingsUseCase creates a new notification settings use case
func NewNotificationSettingsUseCase() NotificationSettingsUseCase {
	return &notificationSettingsUseCase{
		settingsRepo: repository.NewNotificationSettingsRepository(),
	}
}

func (uc *notificationSettingsUseCase) GetSettings(userID string) (*domain.NotificationSettingsModel, error) {
	// Get or create settings (dengan default values jika belum ada)
	return uc.settingsRepo.GetOrCreate(userID)
}

func (uc *notificationSettingsUseCase) UpdateSettings(userID string, emailEnabled *bool, inAppEnabled *bool, expiryThresholdDays *int) (*domain.NotificationSettingsModel, error) {
	// Get or create settings
	settings, err := uc.settingsRepo.GetOrCreate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Update fields jika provided
	if emailEnabled != nil {
		settings.EmailEnabled = *emailEnabled
	}
	if inAppEnabled != nil {
		settings.InAppEnabled = *inAppEnabled
	}
	if expiryThresholdDays != nil {
		// Validate: threshold harus antara 1-365 hari
		if *expiryThresholdDays < 1 || *expiryThresholdDays > 365 {
			return nil, fmt.Errorf("expiry_threshold_days must be between 1 and 365 days")
		}
		settings.ExpiryThresholdDays = *expiryThresholdDays
	}

	// Update settings (menggunakan Updates dengan where clause untuk menghindari duplicate key)
	if err := uc.settingsRepo.Update(settings); err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	// Reload settings untuk memastikan data terbaru
	return uc.settingsRepo.GetByUserID(userID)
}
