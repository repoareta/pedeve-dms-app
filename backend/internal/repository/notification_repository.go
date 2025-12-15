package repository

import (
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// NotificationRepository interface untuk notification operations
type NotificationRepository interface {
	Create(notification *domain.NotificationModel) error
	GetByID(id string) (*domain.NotificationModel, error)
	GetByUserID(userID string, unreadOnly bool, limit int) ([]domain.NotificationModel, error)
	GetByUserIDWithFilters(userID string, unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
	DeleteOldNotifications(daysOld int) error
	DeleteAllByUserID(userID string) error
	DeleteAll() error
	DeleteByUserIDs(userIDs []string) error
	GetAllWithFilters(unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error)
	GetByUserIDsWithFilters(userIDs []string, unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error)
	GetUnreadCountByUserIDs(userIDs []string) (int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{
		db: database.GetDB(),
	}
}

func (r *notificationRepository) Create(notification *domain.NotificationModel) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id string) (*domain.NotificationModel, error) {
	var notification domain.NotificationModel
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) GetByUserID(userID string, unreadOnly bool, limit int) ([]domain.NotificationModel, error) {
	var notifications []domain.NotificationModel
	query := r.db.Where("user_id = ?", userID)
	
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	
	// Load documents manually untuk notifications dengan resource_type = 'document'
	// Hanya load jika ada notifications
	if len(notifications) > 0 {
		docRepo := NewDocumentRepository()
		for i := range notifications {
			if notifications[i].ResourceType == "document" && notifications[i].ResourceID != nil && *notifications[i].ResourceID != "" {
				doc, err := docRepo.GetDocumentByID(*notifications[i].ResourceID)
				if err == nil && doc != nil {
					notifications[i].Document = doc
				}
			}
		}
	}
	
	return notifications, nil
}

func (r *notificationRepository) GetByUserIDWithFilters(userID string, unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error) {
	var notifications []domain.NotificationModel
	var total int64
	
	query := r.db.Model(&domain.NotificationModel{}).Where("user_id = ?", userID)
	
	// Filter by read status
	// Jika unreadOnly = true, hanya ambil yang belum dibaca
	// Jika unreadOnly = false atau nil, ambil semua (tidak filter)
	if unreadOnly != nil && *unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	// Filter by expiry date (join dengan documents)
	if daysUntilExpiry != nil {
		thresholdDate := time.Now().AddDate(0, 0, *daysUntilExpiry)
		query = query.
			Joins("LEFT JOIN documents ON notifications.resource_id::text = documents.id::text AND notifications.resource_type = 'document'").
			Where("documents.expiry_date IS NOT NULL AND documents.expiry_date <= ? AND documents.expiry_date > ?", thresholdDate, time.Now())
	}
	
	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get notifications
	err := query.
		Order("notifications.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Load documents manually untuk notifications dengan resource_type = 'document'
	// Hanya load jika ada notifications
	if len(notifications) > 0 {
		docRepo := NewDocumentRepository()
		for i := range notifications {
			if notifications[i].ResourceType == "document" && notifications[i].ResourceID != nil && *notifications[i].ResourceID != "" {
				doc, err := docRepo.GetDocumentByID(*notifications[i].ResourceID)
				if err == nil && doc != nil {
					notifications[i].Document = doc
				}
			}
		}
	}
	
	return notifications, total, nil
}

func (r *notificationRepository) MarkAsRead(id, userID string) error {
	now := time.Now()
	return r.db.Model(&domain.NotificationModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	now := time.Now()
	return r.db.Model(&domain.NotificationModel{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

func (r *notificationRepository) GetUnreadCount(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.NotificationModel{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) DeleteOldNotifications(daysOld int) error {
	thresholdDate := time.Now().AddDate(0, 0, -daysOld)
	return r.db.Where("created_at < ?", thresholdDate).Delete(&domain.NotificationModel{}).Error
}

func (r *notificationRepository) DeleteAllByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.NotificationModel{}).Error
}

// DeleteAll menghapus semua notifikasi (untuk superadmin)
func (r *notificationRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM notifications").Error
}

// DeleteByUserIDs menghapus notifikasi berdasarkan user IDs (untuk admin)
func (r *notificationRepository) DeleteByUserIDs(userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	return r.db.Where("user_id IN ?", userIDs).Delete(&domain.NotificationModel{}).Error
}

// GetAllWithFilters untuk superadmin - melihat semua notifikasi
func (r *notificationRepository) GetAllWithFilters(unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error) {
	var notifications []domain.NotificationModel
	var total int64
	
	query := r.db.Model(&domain.NotificationModel{})
	
	// Filter by read status
	// Jika unreadOnly = true, hanya ambil yang belum dibaca
	// Jika unreadOnly = false atau nil, ambil semua (tidak filter)
	if unreadOnly != nil && *unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	// Filter by expiry date (join dengan documents)
	if daysUntilExpiry != nil {
		thresholdDate := time.Now().AddDate(0, 0, *daysUntilExpiry)
		query = query.
			Joins("LEFT JOIN documents ON notifications.resource_id::text = documents.id::text AND notifications.resource_type = 'document'").
			Where("documents.expiry_date IS NOT NULL AND documents.expiry_date <= ? AND documents.expiry_date > ?", thresholdDate, time.Now())
	}
	
	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get notifications
	err := query.
		Order("notifications.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Load documents manually
	if len(notifications) > 0 {
		docRepo := NewDocumentRepository()
		for i := range notifications {
			if notifications[i].ResourceType == "document" && notifications[i].ResourceID != nil && *notifications[i].ResourceID != "" {
				doc, err := docRepo.GetDocumentByID(*notifications[i].ResourceID)
				if err == nil && doc != nil {
					notifications[i].Document = doc
				}
			}
		}
	}
	
	return notifications, total, nil
}

// GetByUserIDsWithFilters untuk admin - melihat notifikasi dari user IDs tertentu (company + descendants)
func (r *notificationRepository) GetByUserIDsWithFilters(userIDs []string, unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error) {
	var notifications []domain.NotificationModel
	var total int64
	
	if len(userIDs) == 0 {
		return []domain.NotificationModel{}, 0, nil
	}
	
	query := r.db.Model(&domain.NotificationModel{}).Where("user_id IN ?", userIDs)
	
	// Filter by read status
	// Jika unreadOnly = true, hanya ambil yang belum dibaca
	// Jika unreadOnly = false atau nil, ambil semua (tidak filter)
	if unreadOnly != nil && *unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	
	// Filter by expiry date (join dengan documents)
	if daysUntilExpiry != nil {
		thresholdDate := time.Now().AddDate(0, 0, *daysUntilExpiry)
		query = query.
			Joins("LEFT JOIN documents ON notifications.resource_id::text = documents.id::text AND notifications.resource_type = 'document'").
			Where("documents.expiry_date IS NOT NULL AND documents.expiry_date <= ? AND documents.expiry_date > ?", thresholdDate, time.Now())
	}
	
	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get notifications
	err := query.
		Order("notifications.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Load documents manually
	if len(notifications) > 0 {
		docRepo := NewDocumentRepository()
		for i := range notifications {
			if notifications[i].ResourceType == "document" && notifications[i].ResourceID != nil && *notifications[i].ResourceID != "" {
				doc, err := docRepo.GetDocumentByID(*notifications[i].ResourceID)
				if err == nil && doc != nil {
					notifications[i].Document = doc
				}
			}
		}
	}
	
	return notifications, total, nil
}

// GetUnreadCountByUserIDs untuk admin - menghitung unread dari user IDs tertentu
func (r *notificationRepository) GetUnreadCountByUserIDs(userIDs []string) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	
	var count int64
	err := r.db.Model(&domain.NotificationModel{}).
		Where("user_id IN ? AND is_read = ?", userIDs, false).
		Count(&count).Error
	return count, err
}

