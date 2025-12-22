package repository

import (
	"encoding/json"
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
	MarkAsReadByID(id string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
	DeleteOldNotifications(daysOld int) error
	DeleteAllByUserID(userID string) error
	DeleteAll() error
	DeleteByUserIDs(userIDs []string) error
	GetAllWithFilters(unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error)
	GetByUserIDsWithFilters(userIDs []string, unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error)
	GetUnreadCountByUserIDs(userIDs []string) (int64, error)
	GetAllUnreadCount() (int64, error) // Untuk superadmin - menghitung semua unread notifications
}

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository() NotificationRepository {
	return NewNotificationRepositoryWithDB(database.GetDB())
}

// NewNotificationRepositoryWithDB creates a new notification repository with injected DB (for testing)
func NewNotificationRepositoryWithDB(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		db: db,
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
	// Jika unreadOnly = true, hanya ambil yang belum dibaca (is_read = false)
	// Jika unreadOnly = false, hanya ambil yang sudah dibaca (is_read = true)
	// Jika unreadOnly = nil, ambil semua (tidak filter)
	if unreadOnly != nil {
		if *unreadOnly {
			query = query.Where("is_read = ?", false)
		} else {
			query = query.Where("is_read = ?", true)
		}
	}

	// Filter by expiry date (join dengan documents)
	// PENTING: expiry_date sekarang disimpan di metadata, bukan di kolom expiry_date
	// Filter ini tidak bisa menggunakan WHERE langsung karena expiry_date ada di JSON metadata
	// Filter akan dilakukan di application layer setelah load documents
	if daysUntilExpiry != nil {
		// Filter hanya untuk document notifications
		query = query.Where("notifications.resource_type = 'document'")
		// Note: Filter berdasarkan expiry_date dari metadata akan dilakukan di application layer
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

	// Filter berdasarkan days_until_expiry di application layer (karena expiry_date ada di metadata)
	if daysUntilExpiry != nil {
		notifications, total = filterNotificationsByExpiry(notifications, *daysUntilExpiry)
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

// MarkAsReadByID marks notification as read by ID only (untuk RBAC - superadmin/admin)
func (r *notificationRepository) MarkAsReadByID(id string) error {
	now := time.Now()
	result := r.db.Model(&domain.NotificationModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// filterNotificationsByExpiry memfilter notifications berdasarkan days_until_expiry
// Hanya menampilkan document notifications yang akan expired dalam X hari atau sudah expired
func filterNotificationsByExpiry(notifications []domain.NotificationModel, daysUntilExpiry int) ([]domain.NotificationModel, int64) {
	filteredNotifications := []domain.NotificationModel{}
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())

	for _, notif := range notifications {
		// Hanya filter untuk document notifications
		if notif.ResourceType == "document" && notif.Document != nil {
			// Baca expiry_date dari metadata
			var expiryDate *time.Time
			if len(notif.Document.Metadata) > 0 {
				var metadata map[string]interface{}
				if err := json.Unmarshal(notif.Document.Metadata, &metadata); err == nil {
					// Cek berbagai kemungkinan key untuk expiry_date di metadata
					if expiredDateStr, ok := metadata["expired_date"].(string); ok && expiredDateStr != "" {
						if parsedDate, err := time.Parse("2006-01-02", expiredDateStr); err == nil {
							expiryDate = &parsedDate
						} else if parsedDate, err := time.Parse(time.RFC3339, expiredDateStr); err == nil {
							expiryDate = &parsedDate
						}
					} else if expiryDateStr, ok := metadata["expiry_date"].(string); ok && expiryDateStr != "" {
						if parsedDate, err := time.Parse("2006-01-02", expiryDateStr); err == nil {
							expiryDate = &parsedDate
						} else if parsedDate, err := time.Parse(time.RFC3339, expiryDateStr); err == nil {
							expiryDate = &parsedDate
						}
					}
				}
			}

			// Jika ada expiry_date, cek apakah dalam threshold
			if expiryDate != nil {
				docExpiryDate := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())
				daysUntil := int(docExpiryDate.Sub(todayStart).Hours() / 24)

				// Filter: hanya yang akan expired dalam X hari (0 sampai X hari ke depan)
				// daysUntil = 0 berarti expired hari ini, daysUntil > 0 berarti akan expired
				// daysUntil < 0 berarti sudah expired (lebih dari hari ini)
				// Filter "Kurang dari X Hari Expired" berarti: 0 <= daysUntil <= daysUntilExpiry
				if daysUntil >= 0 && daysUntil <= daysUntilExpiry {
					filteredNotifications = append(filteredNotifications, notif)
				}
			}
		} else {
			// Untuk non-document notifications, tidak filter (tampilkan semua)
			filteredNotifications = append(filteredNotifications, notif)
		}
	}

	return filteredNotifications, int64(len(filteredNotifications))
}

// GetAllWithFilters untuk superadmin - melihat semua notifikasi
func (r *notificationRepository) GetAllWithFilters(unreadOnly *bool, daysUntilExpiry *int, limit, offset int) ([]domain.NotificationModel, int64, error) {
	var notifications []domain.NotificationModel
	var total int64

	query := r.db.Model(&domain.NotificationModel{})

	// Filter by read status
	// Jika unreadOnly = true, hanya ambil yang belum dibaca (is_read = false)
	// Jika unreadOnly = false, hanya ambil yang sudah dibaca (is_read = true)
	// Jika unreadOnly = nil, ambil semua (tidak filter)
	if unreadOnly != nil {
		if *unreadOnly {
			query = query.Where("is_read = ?", false)
		} else {
			query = query.Where("is_read = ?", true)
		}
	}

	// Filter by expiry date (join dengan documents)
	// PENTING: expiry_date sekarang disimpan di metadata, bukan di kolom expiry_date
	// Filter ini tidak bisa menggunakan WHERE langsung karena expiry_date ada di JSON metadata
	// Filter akan dilakukan di application layer setelah load documents
	if daysUntilExpiry != nil {
		// Filter hanya untuk document notifications
		query = query.Where("notifications.resource_type = 'document'")
		// Note: Filter berdasarkan expiry_date dari metadata akan dilakukan di application layer
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

	// Filter berdasarkan days_until_expiry di application layer (karena expiry_date ada di metadata)
	if daysUntilExpiry != nil {
		notifications, total = filterNotificationsByExpiry(notifications, *daysUntilExpiry)
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
	// Jika unreadOnly = true, hanya ambil yang belum dibaca (is_read = false)
	// Jika unreadOnly = false, hanya ambil yang sudah dibaca (is_read = true)
	// Jika unreadOnly = nil, ambil semua (tidak filter)
	if unreadOnly != nil {
		if *unreadOnly {
			query = query.Where("is_read = ?", false)
		} else {
			query = query.Where("is_read = ?", true)
		}
	}

	// Filter by expiry date (join dengan documents)
	// PENTING: expiry_date sekarang disimpan di metadata, bukan di kolom expiry_date
	// Filter ini tidak bisa menggunakan WHERE langsung karena expiry_date ada di JSON metadata
	// Filter akan dilakukan di application layer setelah load documents
	if daysUntilExpiry != nil {
		// Filter hanya untuk document notifications
		query = query.Where("notifications.resource_type = 'document'")
		// Note: Filter berdasarkan expiry_date dari metadata akan dilakukan di application layer
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

	// Filter berdasarkan days_until_expiry di application layer (karena expiry_date ada di metadata)
	if daysUntilExpiry != nil {
		notifications, total = filterNotificationsByExpiry(notifications, *daysUntilExpiry)
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

// GetAllUnreadCount untuk superadmin - menghitung semua unread notifications di sistem
func (r *notificationRepository) GetAllUnreadCount() (int64, error) {
	var count int64
	err := r.db.Model(&domain.NotificationModel{}).
		Where("is_read = ?", false).
		Count(&count).Error
	return count, err
}
