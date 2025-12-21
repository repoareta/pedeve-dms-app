package usecase

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	MarkAsReadWithRBAC(notificationID, userID, roleName string, companyID *string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
	GetUnreadCountWithRBAC(userID, roleName string, companyID *string) (int64, error)
	DeleteAll(userID string) error
	DeleteAllWithRBAC(userID, roleName string, companyID *string) error
	CheckExpiringDocuments(thresholdDays int) (notificationsCreated int, documentsFound int, err error)
	CheckExpiringDirectorTerms(thresholdDays int) (notificationsCreated int, directorsFound int, err error)
}

type notificationUseCase struct {
	notifRepo    repository.NotificationRepository
	docRepo      repository.DocumentRepository
	userRepo     repository.UserRepository
	companyRepo  repository.CompanyRepository
	directorRepo repository.DirectorRepository
	db           *gorm.DB // For direct queries in CheckExpiringDocuments and CheckExpiringDirectorTerms
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase() NotificationUseCase {
	return NewNotificationUseCaseWithDB(database.GetDB())
}

// NewNotificationUseCaseWithDB creates a new notification use case with injected DB (for testing)
func NewNotificationUseCaseWithDB(db *gorm.DB) NotificationUseCase {
	return &notificationUseCase{
		notifRepo:    repository.NewNotificationRepositoryWithDB(db),
		docRepo:      repository.NewDocumentRepositoryWithDB(db),
		userRepo:     repository.NewUserRepositoryWithDB(db),
		companyRepo:  repository.NewCompanyRepositoryWithDB(db),
		directorRepo: repository.NewDirectorRepositoryWithDB(db),
		db:           db,
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

	// Superadmin/Administrator melihat semua unread count dari semua notifications di sistem
	if utils.IsSuperAdminLike(roleName) {
		// Count total unread dari semua notifications tanpa filter userID
		total, err := uc.notifRepo.GetAllUnreadCount()

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

// MarkAsReadWithRBAC marks notification as read with RBAC support
// - Superadmin/Administrator: can mark any notification as read
// - Admin: can mark notifications from their company + descendants as read
// - Regular users: can only mark their own notifications as read
func (uc *notificationUseCase) MarkAsReadWithRBAC(notificationID, userID, roleName string, companyID *string) error {
	// Get notification first
	notification, err := uc.notifRepo.GetByID(notificationID)
	if err != nil {
		return fmt.Errorf("notification not found")
	}

	// Superadmin/Administrator can mark any notification as read
	if utils.IsSuperAdminLike(roleName) {
		err = uc.notifRepo.MarkAsReadByID(notificationID)
		if err == nil {
			// Invalidate semua cache karena superadmin melihat semua unread count
			// Ini diperlukan karena unread count global berubah
			invalidateAllUnreadCountCache()
			// Juga invalidate cache untuk owner notification (untuk regular users/admins)
			invalidateUnreadCountCache(notification.UserID)
		}
		return err
	}

	// Admin can mark notifications from their company + descendants
	if roleName == "admin" && companyID != nil {
		// Get notification owner's company
		owner, err := uc.userRepo.GetByID(notification.UserID)
		if err != nil {
			return fmt.Errorf("notification owner not found")
		}

		if owner.CompanyID == nil {
			return fmt.Errorf("forbidden: notification owner has no company")
		}

		// Check if owner's company is same or descendant of admin's company
		descendants, err := uc.companyRepo.GetDescendants(*companyID)
		if err != nil {
			return fmt.Errorf("failed to get company descendants: %w", err)
		}

		// Check if owner's company is same as admin's company
		if *owner.CompanyID == *companyID {
			err = uc.notifRepo.MarkAsReadByID(notificationID)
			if err == nil {
				invalidateUnreadCountCache(notification.UserID)
			}
			return err
		}

		// Check if owner's company is descendant of admin's company
		for _, desc := range descendants {
			if desc.ID == *owner.CompanyID {
				err = uc.notifRepo.MarkAsReadByID(notificationID)
				if err == nil {
					invalidateUnreadCountCache(notification.UserID)
				}
				return err
			}
		}

		return fmt.Errorf("forbidden: notification does not belong to your company or its subsidiaries")
	}

	// Regular users can only mark their own notifications as read
	if notification.UserID != userID {
		return fmt.Errorf("forbidden: notification does not belong to user")
	}

	err = uc.notifRepo.MarkAsRead(notificationID, userID)
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

// invalidateAllUnreadCountCache menghapus semua cache unread count (untuk superadmin/administrator)
func invalidateAllUnreadCountCache() {
	unreadCountCacheMu.Lock()
	defer unreadCountCacheMu.Unlock()

	// Hapus semua cache entry yang terkait dengan superadmin/administrator
	keysToDelete := []string{}
	for key := range unreadCountCache {
		// Cache key untuk superadmin/administrator berbentuk "superadmin:userID" atau "administrator:userID"
		// atau "superadmin:userID:companyID" atau "administrator:userID:companyID"
		if len(key) >= 10 && (key[:10] == "superadmin" || key[:13] == "administrator") {
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
// Breakdown per folder: dokumen akan di-group berdasarkan folder untuk notifikasi yang lebih terorganisir
// Ini akan dipanggil oleh scheduler/cron job
// thresholdDays: jumlah hari sebelum expired untuk membuat notifikasi pertama kali (default: 14 hari)
// Dokumen yang kurang dari thresholdDays tapi belum ada notifikasinya akan langsung dibuat notifikasinya
func (uc *notificationUseCase) CheckExpiringDocuments(thresholdDays int) (notificationsCreated int, documentsFound int, err error) {
	zapLog := logger.GetLogger()

	// Gunakan start of day untuk perbandingan tanggal yang konsisten
	now := time.Now()
	// Truncate ke start of day untuk memastikan perbandingan hanya berdasarkan tanggal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	thresholdDate := todayStart.AddDate(0, 0, thresholdDays)

	// Query documents yang akan expired dalam threshold atau sudah expired (termasuk hari ini)
	// PENTING: Ambil dokumen yang sudah expired juga (untuk reminder sampai ditindaklanjuti)
	// Cek apakah sudah ada notifikasi yang belum dibaca untuk dokumen tersebut
	// Jika belum ada notifikasi unread, buat notifikasi baru
	var documents []domain.DocumentModel
	db := uc.db
	if db == nil {
		db = database.GetDB() // Fallback to default DB if not injected
	}

	// Query dokumen yang expired (termasuk hari ini) atau akan expired dalam threshold
	// Gunakan end of day untuk thresholdDate agar semua dokumen yang expired pada hari threshold juga terambil
	// Ini memastikan dokumen yang expired hari ini tetap terambil meskipun ada perbedaan timezone
	thresholdDateEnd := thresholdDate.AddDate(0, 0, 1).Add(-time.Nanosecond) // End of threshold day (23:59:59.999999999)

	// Log query parameters untuk debugging
	zapLog.Info("Querying expiring documents",
		zap.Int("threshold_days", thresholdDays),
		zap.Time("today_start", todayStart),
		zap.Time("threshold_date", thresholdDate),
		zap.Time("threshold_date_end", thresholdDateEnd),
		zap.String("threshold_date_end_formatted", thresholdDateEnd.Format("2006-01-02 15:04:05")),
	)

	// PENTING: expiry_date disimpan di metadata, bukan di kolom expiry_date
	// Query semua dokumen yang memiliki metadata (karena expiry_date ada di metadata)
	// PENTING: Query ini TIDAK memfilter berdasarkan company_id - mencakup SEMUA perusahaan
	// Notifikasi akan dibuat untuk semua dokumen expired dari semua perusahaan dan folder
	// Note: Query kompatibel dengan SQLite (untuk testing) dan PostgreSQL (untuk production)
	var totalDocsWithMetadata int64
	query := db.Model(&domain.DocumentModel{}).Where("metadata IS NOT NULL")
	// Untuk PostgreSQL, tambahkan filter untuk exclude empty JSON
	if db.Dialector.Name() == "postgres" {
		query = query.Where("metadata != '{}'::jsonb")
	} else {
		// Untuk SQLite, gunakan filter yang kompatibel
		query = query.Where("metadata != '{}' AND metadata != 'null' AND metadata != ''")
	}
	query.Count(&totalDocsWithMetadata)

	zapLog.Info("Total documents with metadata",
		zap.Int64("with_metadata", totalDocsWithMetadata),
	)

	var allDocs []domain.DocumentModel
	queryDocs := db.Preload("Folder").Where("metadata IS NOT NULL")
	// Untuk PostgreSQL, tambahkan filter untuk exclude empty JSON
	if db.Dialector.Name() == "postgres" {
		queryDocs = queryDocs.Where("metadata != '{}'::jsonb")
	} else {
		// Untuk SQLite, gunakan filter yang kompatibel
		queryDocs = queryDocs.Where("metadata != '{}' AND metadata != 'null' AND metadata != ''")
	}
	err = queryDocs.Find(&allDocs).Error
	if err != nil {
		zapLog.Error("Failed to query documents", zap.Error(err))
		return 0, 0, err
	}

	// Log untuk memastikan query mencakup semua perusahaan
	zapLog.Info("Querying documents from ALL companies and folders (no company filter)",
		zap.Int("total_documents_queried", len(allDocs)),
	)

	// Filter dokumen yang expired atau akan expired
	// Baca expiry_date dari metadata (field expiry_date di table tidak digunakan lagi)
	documents = []domain.DocumentModel{}
	for _, doc := range allDocs {
		var expiryDate *time.Time

		// Baca expiry_date dari metadata
		if doc.Metadata != nil {
			// Jika kolom expiry_date NULL, coba baca dari metadata
			var metadata map[string]interface{}
			if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
				// Cek berbagai kemungkinan key untuk expiry_date di metadata
				if expiredDateStr, ok := metadata["expired_date"].(string); ok && expiredDateStr != "" {
					// Parse expired_date dari metadata (format: "2025-12-20" atau "2025-12-20T00:00:00Z")
					if parsedDate, err := time.Parse("2006-01-02", expiredDateStr); err == nil {
						expiryDate = &parsedDate
					} else if parsedDate, err := time.Parse(time.RFC3339, expiredDateStr); err == nil {
						expiryDate = &parsedDate
					} else if parsedDate, err := time.Parse("2006-01-02T15:04:05Z07:00", expiredDateStr); err == nil {
						expiryDate = &parsedDate
					}
				} else if expiryDateStr, ok := metadata["expiry_date"].(string); ok && expiryDateStr != "" {
					// Coba key "expiry_date" juga
					if parsedDate, err := time.Parse("2006-01-02", expiryDateStr); err == nil {
						expiryDate = &parsedDate
					} else if parsedDate, err := time.Parse(time.RFC3339, expiryDateStr); err == nil {
						expiryDate = &parsedDate
					}
				}
			}
		}

		// Jika ada expiry_date (dari metadata), cek apakah expired atau akan expired
		if expiryDate != nil {
			// Truncate ke start of day untuk perbandingan yang konsisten
			expiryDateStart := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())
			if expiryDateStart.Before(thresholdDate) || expiryDateStart.Equal(thresholdDate) || expiryDateStart.Before(thresholdDateEnd) {
				documents = append(documents, doc)
			}
		}
	}

	zapLog.Info("Filtered documents with expiry_date (from metadata)",
		zap.Int("total_docs_checked", len(allDocs)),
		zap.Int("docs_with_valid_expiry", len(documents)),
	)

	// Log hasil query dengan detail
	zapLog.Info("Query expiring documents completed",
		zap.Int("threshold_days", thresholdDays),
		zap.Time("threshold_date", thresholdDate),
		zap.Time("threshold_date_end", thresholdDateEnd),
		zap.Time("today_start", todayStart),
		zap.Int("documents_found", len(documents)),
	)

	// Log detail dokumen yang ditemukan untuk debugging
	if len(documents) > 0 {
		for i, doc := range documents {
			if i < 10 { // Log 10 pertama untuk debugging
				expiryStr := "NULL"
				daysUntil := 0
				// Baca expiry_date dari metadata
				var expiryDate *time.Time
				if len(doc.Metadata) > 0 {
					metadata := make(map[string]interface{})
					if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
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
				if expiryDate != nil {
					expiryStr = expiryDate.Format("2006-01-02 15:04:05")
					docExpiryDate := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())
					daysUntil = int(docExpiryDate.Sub(todayStart).Hours() / 24)
				}
				zapLog.Info("Found expiring document",
					zap.String("document_id", doc.ID),
					zap.String("document_name", doc.Name),
					zap.String("expiry_date", expiryStr),
					zap.Int("days_until_expiry", daysUntil),
					zap.String("uploader_id", doc.UploaderID),
				)
			}
		}
		if len(documents) > 10 {
			zapLog.Info("... and more documents", zap.Int("total", len(documents)))
		}
	} else {
		zapLog.Warn("No expiring documents found",
			zap.Time("threshold_date_end", thresholdDateEnd),
			zap.String("threshold_date_end_str", thresholdDateEnd.Format("2006-01-02 15:04:05")),
		)
		// Coba query alternatif untuk debugging - cek dokumen dengan metadata
		var sampleDocs []domain.DocumentModel
		querySample := db.Where("metadata IS NOT NULL")
		// Untuk PostgreSQL, tambahkan filter untuk exclude empty JSON
		if db.Dialector.Name() == "postgres" {
			querySample = querySample.Where("metadata != '{}'::jsonb")
		} else {
			// Untuk SQLite, gunakan filter yang kompatibel
			querySample = querySample.Where("metadata != '{}' AND metadata != 'null' AND metadata != ''")
		}
		querySample.Limit(5).Find(&sampleDocs)
		if len(sampleDocs) > 0 {
			zapLog.Info("Sample documents with metadata (first 5):")
			for _, doc := range sampleDocs {
				expiryStr := "NULL"
				// Baca expiry_date dari metadata
				if len(doc.Metadata) > 0 {
					metadata := make(map[string]interface{})
					if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
						if expiredDateStr, ok := metadata["expired_date"].(string); ok && expiredDateStr != "" {
							expiryStr = expiredDateStr
						} else if expiryDateStr, ok := metadata["expiry_date"].(string); ok && expiryDateStr != "" {
							expiryStr = expiryDateStr
						}
					}
				}
				zapLog.Info("Sample document",
					zap.String("id", doc.ID),
					zap.String("name", doc.Name),
					zap.String("expiry_date_from_metadata", expiryStr),
				)
			}
		}
	}

	documentsFound = len(documents)
	notificationsCreated = 0

	// PENTING: Notifikasi dibuat untuk SEMUA dokumen expired dari SEMUA perusahaan dan folder
	// Tidak ada filter company_id - semua dokumen dari semua perusahaan akan mendapat notifikasi
	// PENTING: Buat notifikasi per dokumen (bukan per folder) agar count lebih akurat
	for _, doc := range documents {
		// Baca expiry_date dari metadata
		var expiryDate *time.Time
		if len(doc.Metadata) > 0 {
			metadata := make(map[string]interface{})
			if err := json.Unmarshal(doc.Metadata, &metadata); err == nil {
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

		// Jika tidak ada expiry_date, skip dokumen ini
		if expiryDate == nil {
			continue
		}

		// Hitung hari dengan perbandingan tanggal yang konsisten
		docExpiryDate := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())
		daysUntilExpiry := int(docExpiryDate.Sub(todayStart).Hours() / 24)
		isExpired := docExpiryDate.Before(todayStart) || docExpiryDate.Equal(todayStart)

		// Ambil nama folder untuk message
		folderName := "No Folder"
		if doc.FolderID != nil && doc.Folder != nil {
			folderName = doc.Folder.Name
		} else if doc.FolderID != nil {
			// Load folder if not loaded
			folder, err := uc.docRepo.GetFolderByID(*doc.FolderID)
			if err == nil && folder != nil {
				folderName = folder.Name
			}
		}

		// Cek apakah sudah ada notifikasi unread untuk dokumen ini dengan status yang sama (expired/akan expired)
		// PENTING: Notifikasi dibuat sekali per dokumen per status (expired vs akan expired)
		// Jika dokumen sudah expired dan sudah ada notifikasi "Sudah Expired", tidak perlu buat lagi
		// Tapi jika dokumen akan expired dan belum ada notifikasi "Akan Expired", tetap buat notifikasi
		hasUnreadNotif := false
		existingNotifs, err := uc.notifRepo.GetByUserID(doc.UploaderID, true, 100)
		if err == nil {
			for _, notif := range existingNotifs {
				if notif.ResourceType == "document" && notif.ResourceID != nil && *notif.ResourceID == doc.ID {
					// Cek apakah status notifikasi sama (expired vs akan expired)
					// Jika dokumen akan expired (daysUntilExpiry > 0) dan sudah ada notifikasi "Akan Expired", skip
					// Jika dokumen sudah expired (daysUntilExpiry < 0) dan sudah ada notifikasi "Sudah Expired", skip
					// Tapi jika dokumen akan expired dan hanya ada notifikasi "Sudah Expired", tetap buat notifikasi "Akan Expired"
					if daysUntilExpiry >= 0 {
						// Dokumen akan expired - cek apakah sudah ada notifikasi "Akan Expired"
						if strings.Contains(notif.Title, "Akan Expired") {
							hasUnreadNotif = true
							break
						}
					} else {
						// Dokumen sudah expired - cek apakah sudah ada notifikasi "Sudah Expired"
						if strings.Contains(notif.Title, "Sudah Expired") {
							hasUnreadNotif = true
							break
						}
					}
				}
			}
		}
		// Jika sudah ada notifikasi unread dengan status yang sama, skip
		// Push notification akan muncul berulang di frontend untuk notifikasi yang belum ditindak lanjuti
		if hasUnreadNotif {
			continue
		}

		// Buat notifikasi untuk setiap dokumen
		var title, message string
		if daysUntilExpiry < 0 {
			// Sudah expired
			daysAgo := -daysUntilExpiry
			title = fmt.Sprintf("Dokumen '%s' Sudah Expired", doc.Name)
			if daysAgo == 0 {
				message = fmt.Sprintf("Dokumen '%s' di folder '%s' sudah expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.",
					doc.Name, folderName)
			} else {
				message = fmt.Sprintf("Dokumen '%s' di folder '%s' sudah expired %d hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.",
					doc.Name, folderName, daysAgo)
			}
		} else if daysUntilExpiry == 0 {
			title = fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name)
			message = fmt.Sprintf("Dokumen '%s' di folder '%s' akan expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.",
				doc.Name, folderName)
		} else {
			title = fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name)
			message = fmt.Sprintf("Dokumen '%s' di folder '%s' akan expired dalam %d hari. Silakan perbarui atau perpanjang dokumen tersebut.",
				doc.Name, folderName, daysUntilExpiry)
		}

		_, err = uc.CreateNotification(
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
		} else {
			notificationsCreated++
			zapLog.Info("Created notification for document",
				zap.String("document_id", doc.ID),
				zap.String("document_name", doc.Name),
				zap.Int("days_until_expiry", daysUntilExpiry),
				zap.Bool("is_expired", isExpired),
			)
		}
	}

	return notificationsCreated, documentsFound, nil
}

// CheckExpiringDirectorTerms adalah helper function untuk check dan create notifications untuk masa jabatan pengurus yang akan berakhir
// Hanya akan check directors yang memiliki EndDate (tidak null)
// Ini akan dipanggil oleh scheduler/cron job
// thresholdDays: jumlah hari sebelum expired untuk membuat notifikasi pertama kali (default: 14 hari)
// Masa jabatan yang kurang dari thresholdDays tapi belum ada notifikasinya akan langsung dibuat notifikasinya
func (uc *notificationUseCase) CheckExpiringDirectorTerms(thresholdDays int) (notificationsCreated int, directorsFound int, err error) {
	zapLog := logger.GetLogger()

	thresholdDate := time.Now().AddDate(0, 0, thresholdDays)

	// Query directors yang akan expired dalam threshold atau sudah expired (hanya yang memiliki EndDate)
	var directors []domain.DirectorModel
	db := uc.db
	if db == nil {
		db = database.GetDB() // Fallback to default DB if not injected
	}
	err = db.
		Where("end_date IS NOT NULL AND end_date <= ?", thresholdDate).
		Find(&directors).Error
	if err != nil {
		zapLog.Error("Failed to query expiring directors", zap.Error(err))
		return 0, 0, err
	}

	directorsFound = len(directors)
	notificationsCreated = 0

	// Get company admins and superadmins who should be notified
	for _, director := range directors {
		daysUntilExpiry := int(time.Until(*director.EndDate).Hours() / 24)

		// Get company name
		company, err := uc.companyRepo.GetByID(director.CompanyID)
		companyName := director.CompanyID // fallback to ID
		if err == nil && company != nil {
			companyName = company.Name
		}

		// Get users who have access to this company (admins and superadmins)
		// For simplicity, we'll notify all users associated with the company
		users, err := uc.userRepo.GetByCompanyID(director.CompanyID)
		if err != nil {
			zapLog.Warn("Failed to get users for company", zap.Error(err), zap.String("company_id", director.CompanyID))
			continue
		}

		// Create notification for each user in the company
		// PENTING: Cek apakah sudah ada notifikasi unread untuk director ini
		// Jika sudah ada, skip (notifikasi akan tetap muncul sampai ditindaklanjuti)
		directorID := director.ID
		for _, user := range users {
			// Cek apakah sudah ada notifikasi unread untuk director ini
			hasUnreadNotification := false
			existingNotifs, err := uc.notifRepo.GetByUserID(user.ID, true, 100)
			if err == nil {
				for _, notif := range existingNotifs {
					if notif.ResourceType == "director" && notif.ResourceID != nil && *notif.ResourceID == directorID {
						hasUnreadNotification = true
						break
					}
				}
			}

			// Jika sudah ada notifikasi unread, skip (notifikasi akan tetap muncul sampai ditindaklanjuti)
			if hasUnreadNotification {
				continue
			}

			var title, message string
			if daysUntilExpiry < 0 {
				// Sudah expired
				daysAgo := -daysUntilExpiry
				title = fmt.Sprintf("Masa Jabatan '%s' Sudah Berakhir", director.FullName)
				if daysAgo == 0 {
					message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s sudah berakhir hari ini. Silakan perpanjang atau ganti pengurus tersebut.",
						director.FullName, director.Position, companyName)
				} else {
					message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s sudah berakhir %d hari yang lalu. Silakan perpanjang atau ganti pengurus tersebut.",
						director.FullName, director.Position, companyName, daysAgo)
				}
			} else if daysUntilExpiry == 0 {
				title = fmt.Sprintf("Masa Jabatan '%s' Akan Berakhir", director.FullName)
				message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s akan berakhir hari ini. Silakan perpanjang atau ganti pengurus tersebut.",
					director.FullName, director.Position, companyName)
			} else {
				title = fmt.Sprintf("Masa Jabatan '%s' Akan Berakhir", director.FullName)
				message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s akan berakhir dalam %d hari. Silakan perpanjang atau ganti pengurus tersebut.",
					director.FullName, director.Position, companyName, daysUntilExpiry)
			}

			// Use director ID as resource ID
			_, err = uc.CreateNotification(
				user.ID,
				"director_term_expiry",
				title,
				message,
				"director",
				&directorID,
			)
			if err != nil {
				zapLog.Error("Failed to create notification for director term expiry", zap.Error(err),
					zap.String("director_id", director.ID),
					zap.String("user_id", user.ID))
			} else {
				notificationsCreated++
			}
		}
	}

	return notificationsCreated, directorsFound, nil
}
