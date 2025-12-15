package usecase

import (
	"fmt"
	"sync"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
)

// Cache entry untuk unread count
type unreadCountCacheEntry struct {
	count     int64
	expiresAt time.Time
}

// In-memory cache untuk unread count dengan TTL 5 detik
var (
	unreadCountCache    = make(map[string]*unreadCountCacheEntry)
	unreadCountCacheMu  sync.RWMutex
	unreadCountCacheTTL = 5 * time.Second // Cache selama 5 detik untuk mengurangi query
)

// NotificationUseCase interface untuk notification operations
type NotificationUseCase interface {
	CreateNotification(userID, notificationType, title, message, resourceType string, resourceID *string) (*domain.NotificationModel, error)
	GetUserNotifications(userID string, unreadOnly bool, limit int) ([]domain.NotificationModel, error)
	GetUserNotificationsWithFilters(userID string, unreadOnly *bool, daysUntilExpiry *int, page, pageSize int) ([]domain.NotificationModel, int64, int, error)
	GetNotificationsWithRBAC(userID, roleName string, companyID *string, unreadOnly *bool, daysUntilExpiry *int, page, pageSize int) ([]domain.NotificationModel, int64, int, error)
	MarkAsRead(notificationID, userID string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
	GetUnreadCountWithRBAC(userID, roleName string, companyID *string) (int64, error)
	DeleteAll(userID string) error
	DeleteAllWithRBAC(userID, roleName string, companyID *string) error
	CheckExpiringDocuments(thresholdDays int) error
}

type notificationUseCase struct {
	notifRepo   repository.NotificationRepository
	docRepo     repository.DocumentRepository
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase() NotificationUseCase {
	return &notificationUseCase{
		notifRepo:   repository.NewNotificationRepository(),
		docRepo:     repository.NewDocumentRepository(),
		userRepo:    repository.NewUserRepository(),
		companyRepo: repository.NewCompanyRepository(),
	}
}

func (uc *notificationUseCase) CreateNotification(userID, notificationType, title, message, resourceType string, resourceID *string) (*domain.NotificationModel, error) {
	notification := &domain.NotificationModel{
		ID:           uuid.GenerateUUID(),
		UserID:       userID,
		Type:         notificationType,
		Title:        title,
		Message:      message,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IsRead:       false,
		CreatedAt:    time.Now(),
	}

	if err := uc.notifRepo.Create(notification); err != nil {
		return nil, err
	}

	// Invalidate cache untuk user ini setelah notification baru dibuat
	invalidateUnreadCountCache(userID)

	return notification, nil
}

func (uc *notificationUseCase) GetUserNotifications(userID string, unreadOnly bool, limit int) ([]domain.NotificationModel, error) {
	return uc.notifRepo.GetByUserID(userID, unreadOnly, limit)
}

func (uc *notificationUseCase) GetUserNotificationsWithFilters(userID string, unreadOnly *bool, daysUntilExpiry *int, page, pageSize int) ([]domain.NotificationModel, int64, int, error) {
	offset := (page - 1) * pageSize

	notifications, total, err := uc.notifRepo.GetByUserIDWithFilters(userID, unreadOnly, daysUntilExpiry, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return notifications, total, totalPages, nil
}

// GetNotificationsWithRBAC mendapatkan notifikasi dengan RBAC (superadmin lihat semua, admin lihat company+descendants, user lihat sendiri)
func (uc *notificationUseCase) GetNotificationsWithRBAC(userID, roleName string, companyID *string, unreadOnly *bool, daysUntilExpiry *int, page, pageSize int) ([]domain.NotificationModel, int64, int, error) {
	offset := (page - 1) * pageSize

	// Superadmin/Administrator melihat semua notifikasi
	if utils.IsSuperAdminLike(roleName) {
		notifications, total, err := uc.notifRepo.GetAllWithFilters(unreadOnly, daysUntilExpiry, pageSize, offset)
		if err != nil {
			return nil, 0, 0, err
		}
		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		return notifications, total, totalPages, nil
	}

	// Admin melihat notifikasi dari company mereka + descendants
	if roleName == "admin" && companyID != nil {
		// Get all descendants
		descendants, err := uc.companyRepo.GetDescendants(*companyID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to get company descendants: %w", err)
		}

		// Collect all company IDs (own company + descendants)
		companyIDs := []string{*companyID}
		for _, desc := range descendants {
			companyIDs = append(companyIDs, desc.ID)
		}

		// Get all users from these companies
		userIDs := []string{}
		for _, compID := range companyIDs {
			users, err := uc.userRepo.GetByCompanyID(compID)
			if err == nil {
				for _, user := range users {
					userIDs = append(userIDs, user.ID)
				}
			}
		}

		if len(userIDs) == 0 {
			return []domain.NotificationModel{}, 0, 0, nil
		}

		notifications, total, err := uc.notifRepo.GetByUserIDsWithFilters(userIDs, unreadOnly, daysUntilExpiry, pageSize, offset)
		if err != nil {
			return nil, 0, 0, err
		}
		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		return notifications, total, totalPages, nil
	}

	// Regular users hanya melihat notifikasi mereka sendiri
	return uc.GetUserNotificationsWithFilters(userID, unreadOnly, daysUntilExpiry, page, pageSize)
}

// GetUnreadCountWithRBAC mendapatkan unread count dengan RBAC (dengan caching)
func (uc *notificationUseCase) GetUnreadCountWithRBAC(userID, roleName string, companyID *string) (int64, error) {
	// Generate cache key berdasarkan role dan context
	cacheKey := fmt.Sprintf("%s:%s", roleName, userID)
	if companyID != nil {
		cacheKey = fmt.Sprintf("%s:%s:%s", roleName, userID, *companyID)
	}

	// Check cache first
	unreadCountCacheMu.RLock()
	if entry, exists := unreadCountCache[cacheKey]; exists {
		if time.Now().Before(entry.expiresAt) {
			unreadCountCacheMu.RUnlock()
			return entry.count, nil
		}
		// Cache expired, remove it
		delete(unreadCountCache, cacheKey)
	}
	unreadCountCacheMu.RUnlock()

	// Superadmin/Administrator melihat semua unread count
	if utils.IsSuperAdminLike(roleName) {
		// Count total unread dengan optimasi query
		var total int64
		// Gunakan index is_read untuk query yang lebih cepat
		err := database.GetDB().Model(&domain.NotificationModel{}).
			Where("is_read = ?", false).
			Count(&total).Error

		// Cache hasil
		if err == nil {
			unreadCountCacheMu.Lock()
			unreadCountCache[cacheKey] = &unreadCountCacheEntry{
				count:     total,
				expiresAt: time.Now().Add(unreadCountCacheTTL),
			}
			unreadCountCacheMu.Unlock()
		}

		return total, err
	}

	// Admin melihat unread count dari company mereka + descendants
	if roleName == "admin" && companyID != nil {
		// Get all descendants
		descendants, err := uc.companyRepo.GetDescendants(*companyID)
		if err != nil {
			return 0, fmt.Errorf("failed to get company descendants: %w", err)
		}

		// Collect all company IDs
		companyIDs := []string{*companyID}
		for _, desc := range descendants {
			companyIDs = append(companyIDs, desc.ID)
		}

		// Get all users from these companies
		userIDs := []string{}
		for _, compID := range companyIDs {
			users, err := uc.userRepo.GetByCompanyID(compID)
			if err == nil {
				for _, user := range users {
					userIDs = append(userIDs, user.ID)
				}
			}
		}

		if len(userIDs) == 0 {
			return 0, nil
		}

		count, err := uc.notifRepo.GetUnreadCountByUserIDs(userIDs)

		// Cache hasil
		if err == nil {
			unreadCountCacheMu.Lock()
			unreadCountCache[cacheKey] = &unreadCountCacheEntry{
				count:     count,
				expiresAt: time.Now().Add(unreadCountCacheTTL),
			}
			unreadCountCacheMu.Unlock()
		}

		return count, err
	}

	// Regular users hanya melihat unread count mereka sendiri
	return uc.GetUnreadCount(userID)
}

func (uc *notificationUseCase) MarkAsRead(notificationID, userID string) error {
	// Verify notification belongs to user
	notification, err := uc.notifRepo.GetByID(notificationID)
	if err != nil {
		return fmt.Errorf("notification not found")
	}

	if notification.UserID != userID {
		return fmt.Errorf("forbidden: notification does not belong to user")
	}

	err = uc.notifRepo.MarkAsRead(notificationID, userID)

	// Invalidate cache untuk user ini setelah mark as read
	if err == nil {
		invalidateUnreadCountCache(userID)
	}

	return err
}

func (uc *notificationUseCase) MarkAllAsRead(userID string) error {
	err := uc.notifRepo.MarkAllAsRead(userID)

	// Invalidate cache untuk user ini setelah mark all as read
	if err == nil {
		invalidateUnreadCountCache(userID)
	}

	return err
}

// invalidateUnreadCountCache menghapus cache unread count untuk user tertentu
func invalidateUnreadCountCache(userID string) {
	unreadCountCacheMu.Lock()
	defer unreadCountCacheMu.Unlock()

	// Hapus semua cache entry yang terkait dengan userID
	keysToDelete := []string{}
	for key := range unreadCountCache {
		if key == fmt.Sprintf("user:%s", userID) {
			keysToDelete = append(keysToDelete, key)
		} else if len(key) > len(userID) && key[len(key)-len(userID):] == userID {
			keysToDelete = append(keysToDelete, key)
		} else if len(key) > len(userID)+1 && key[:len(userID)+1] == userID+":" {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(unreadCountCache, key)
	}
}

func (uc *notificationUseCase) GetUnreadCount(userID string) (int64, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("user:%s", userID)
	unreadCountCacheMu.RLock()
	if entry, exists := unreadCountCache[cacheKey]; exists {
		if time.Now().Before(entry.expiresAt) {
			unreadCountCacheMu.RUnlock()
			return entry.count, nil
		}
		// Cache expired, remove it
		delete(unreadCountCache, cacheKey)
	}
	unreadCountCacheMu.RUnlock()

	// Get from database
	count, err := uc.notifRepo.GetUnreadCount(userID)

	// Cache hasil
	if err == nil {
		unreadCountCacheMu.Lock()
		unreadCountCache[cacheKey] = &unreadCountCacheEntry{
			count:     count,
			expiresAt: time.Now().Add(unreadCountCacheTTL),
		}
		unreadCountCacheMu.Unlock()
	}

	return count, err
}

func (uc *notificationUseCase) DeleteAll(userID string) error {
	return uc.notifRepo.DeleteAllByUserID(userID)
}

// DeleteAllWithRBAC menghapus notifikasi dengan RBAC (superadmin hapus semua, admin hapus company+descendants, user hapus sendiri)
func (uc *notificationUseCase) DeleteAllWithRBAC(userID, roleName string, companyID *string) error {
	// Superadmin/Administrator menghapus semua notifikasi
	if utils.IsSuperAdminLike(roleName) {
		return uc.notifRepo.DeleteAll()
	}

	// Admin menghapus notifikasi dari company mereka + descendants
	if roleName == "admin" && companyID != nil {
		// Get all descendants
		descendants, err := uc.companyRepo.GetDescendants(*companyID)
		if err != nil {
			return fmt.Errorf("failed to get company descendants: %w", err)
		}

		// Collect all company IDs (own company + descendants)
		companyIDs := []string{*companyID}
		for _, desc := range descendants {
			companyIDs = append(companyIDs, desc.ID)
		}

		// Get all users from these companies
		userIDs := []string{}
		for _, compID := range companyIDs {
			users, err := uc.userRepo.GetByCompanyID(compID)
			if err == nil {
				for _, user := range users {
					userIDs = append(userIDs, user.ID)
				}
			}
		}

		if len(userIDs) == 0 {
			return nil // No users to delete notifications for
		}

		return uc.notifRepo.DeleteByUserIDs(userIDs)
	}

	// Regular users hanya menghapus notifikasi mereka sendiri
	return uc.notifRepo.DeleteAllByUserID(userID)
}

// CheckExpiringDocuments adalah helper function untuk check dan create notifications untuk expiring documents
// Ini akan dipanggil oleh scheduler/cron job
func (uc *notificationUseCase) CheckExpiringDocuments(thresholdDays int) error {
	zapLog := logger.GetLogger()

	thresholdDate := time.Now().AddDate(0, 0, thresholdDays)

	// Query documents yang akan expired dalam threshold
	// Note: Ini perlu diimplementasikan di document_repository.go
	// Untuk sekarang, kita akan query langsung menggunakan database.GetDB()
	var documents []domain.DocumentModel
	err := database.GetDB().Where("expiry_date IS NOT NULL AND expiry_date <= ? AND expiry_date > ? AND expiry_notified = ?",
		thresholdDate, time.Now(), false).Find(&documents).Error
	if err != nil {
		zapLog.Error("Failed to query expiring documents", zap.Error(err))
		return err
	}

	for _, doc := range documents {
		daysUntilExpiry := int(time.Until(*doc.ExpiryDate).Hours() / 24)

		// Create notification untuk uploader
		title := fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name)
		message := fmt.Sprintf("Dokumen '%s' akan expired dalam %d hari. Silakan perbarui atau perpanjang dokumen tersebut.", doc.Name, daysUntilExpiry)

		_, err := uc.CreateNotification(
			doc.UploaderID,
			"document_expiry",
			title,
			message,
			"document",
			&doc.ID,
		)
		if err != nil {
			zapLog.Error("Failed to create notification", zap.Error(err), zap.String("document_id", doc.ID))
			continue
		}

		// Mark document as notified
		err = database.GetDB().Model(&doc).Update("expiry_notified", true).Error
		if err != nil {
			zapLog.Error("Failed to mark document as notified", zap.Error(err), zap.String("document_id", doc.ID))
		}
	}

	return nil
}
