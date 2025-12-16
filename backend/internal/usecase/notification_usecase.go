package usecase

import (
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
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int64, error)
	GetUnreadCountWithRBAC(userID, roleName string, companyID *string) (int64, error)
	DeleteAll(userID string) error
	DeleteAllWithRBAC(userID, roleName string, companyID *string) error
	CheckExpiringDocuments(thresholdDays int) (notificationsCreated int, documentsFound int, err error)
	CheckExpiringDirectorTerms(thresholdDays int) (notificationsCreated int, directorsFound int, err error)
}

type notificationUseCase struct {
	notifRepo   repository.NotificationRepository
	docRepo     repository.DocumentRepository
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	directorRepo repository.DirectorRepository
	db          *gorm.DB // For direct queries in CheckExpiringDocuments and CheckExpiringDirectorTerms
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

	// Superadmin/Administrator melihat semua unread count
	if utils.IsSuperAdminLike(roleName) {
		// Count total unread dengan optimasi query
		var total int64
		var err error
		// Gunakan index is_read untuk query yang lebih cepat
		// Use repository for consistency, but if we need direct DB access, we can add it
		// For now, use the repository method
		total, err = uc.notifRepo.GetUnreadCount(userID)

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
// Breakdown per folder: dokumen akan di-group berdasarkan folder untuk notifikasi yang lebih terorganisir
// Ini akan dipanggil oleh scheduler/cron job
func (uc *notificationUseCase) CheckExpiringDocuments(thresholdDays int) (notificationsCreated int, documentsFound int, err error) {
	zapLog := logger.GetLogger()

	thresholdDate := time.Now().AddDate(0, 0, thresholdDays)

	// Query documents yang akan expired dalam threshold atau sudah expired, dengan join folder untuk breakdown per folder
	// Include documents yang sudah expired (untuk reminder) dan yang akan expired dalam threshold
	var documents []domain.DocumentModel
	db := uc.db
	if db == nil {
		db = database.GetDB() // Fallback to default DB if not injected
	}
	err = db.
		Preload("Folder").
		Where("expiry_date IS NOT NULL AND expiry_date <= ? AND expiry_notified = ?",
			thresholdDate, false).
		Find(&documents).Error
	if err != nil {
		zapLog.Error("Failed to query expiring documents", zap.Error(err))
		return 0, 0, err
	}

	documentsFound = len(documents)
	notificationsCreated = 0

	// Group documents by folder and uploader for better organization
	folderGroups := make(map[string]map[string][]domain.DocumentModel) // [folderID][uploaderID][]documents

	for _, doc := range documents {
		folderKey := "No Folder"
		if doc.FolderID != nil && doc.Folder != nil {
			folderKey = doc.Folder.Name
		} else if doc.FolderID != nil {
			// Load folder if not loaded
			folder, err := uc.docRepo.GetFolderByID(*doc.FolderID)
			if err == nil && folder != nil {
				folderKey = folder.Name
			}
		}

		if folderGroups[folderKey] == nil {
			folderGroups[folderKey] = make(map[string][]domain.DocumentModel)
		}
		folderGroups[folderKey][doc.UploaderID] = append(folderGroups[folderKey][doc.UploaderID], doc)
	}

	// Create notifications per folder and uploader
	for folderName, uploaderGroups := range folderGroups {
		for uploaderID, docs := range uploaderGroups {
			// Jika ada multiple dokumen di folder yang sama, buat satu notifikasi dengan list
			if len(docs) > 1 {
				docNames := make([]string, len(docs))
				for i, doc := range docs {
					docNames[i] = doc.Name
				}
				title := fmt.Sprintf("%d Dokumen di Folder '%s' Akan Expired", len(docs), folderName)
				message := fmt.Sprintf("Ada %d dokumen di folder '%s' yang akan expired dalam %d hari: %s. Silakan perbarui atau perpanjang dokumen-dokumen tersebut.",
					len(docs), folderName, thresholdDays, strings.Join(docNames, ", "))

				// Buat notifikasi untuk dokumen pertama (resource_id = ID dokumen pertama)
				_, err := uc.CreateNotification(
					uploaderID,
					"document_expiry",
					title,
					message,
					"document",
					&docs[0].ID,
				)
				if err != nil {
					zapLog.Error("Failed to create notification for folder group", zap.Error(err), zap.String("folder", folderName))
				} else {
					notificationsCreated++
				}

				// Mark all documents as notified
				for _, doc := range docs {
					err = db.Model(&doc).Update("expiry_notified", true).Error
					if err != nil {
						zapLog.Error("Failed to mark document as notified", zap.Error(err), zap.String("document_id", doc.ID))
					}
				}
			} else {
				// Single document notification
				doc := docs[0]
				daysUntilExpiry := int(time.Until(*doc.ExpiryDate).Hours() / 24)

				title := fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name)
				var message string
				if daysUntilExpiry < 0 {
					// Sudah expired
					daysAgo := -daysUntilExpiry
					if daysAgo == 0 {
						message = fmt.Sprintf("Dokumen '%s' di folder '%s' sudah expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.",
							doc.Name, folderName)
					} else {
						message = fmt.Sprintf("Dokumen '%s' di folder '%s' sudah expired %d hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.",
							doc.Name, folderName, daysAgo)
					}
				} else if daysUntilExpiry == 0 {
					message = fmt.Sprintf("Dokumen '%s' di folder '%s' akan expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.",
						doc.Name, folderName)
				} else {
					message = fmt.Sprintf("Dokumen '%s' di folder '%s' akan expired dalam %d hari. Silakan perbarui atau perpanjang dokumen tersebut.",
						doc.Name, folderName, daysUntilExpiry)
				}

				_, err := uc.CreateNotification(
					uploaderID,
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
				}

				// Mark document as notified
				err = db.Model(&doc).Update("expiry_notified", true).Error
				if err != nil {
					zapLog.Error("Failed to mark document as notified", zap.Error(err), zap.String("document_id", doc.ID))
				}
			}
		}
	}

	return notificationsCreated, documentsFound, nil
}

// CheckExpiringDirectorTerms adalah helper function untuk check dan create notifications untuk masa jabatan pengurus yang akan berakhir
// Hanya akan check directors yang memiliki EndDate (tidak null)
// Ini akan dipanggil oleh scheduler/cron job
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
		for _, user := range users {
			title := fmt.Sprintf("Masa Jabatan '%s' Akan Berakhir", director.FullName)
			var message string
			if daysUntilExpiry < 0 {
				// Sudah expired
				daysAgo := -daysUntilExpiry
				if daysAgo == 0 {
					message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s sudah berakhir hari ini. Silakan perpanjang atau ganti pengurus tersebut.",
						director.FullName, director.Position, companyName)
				} else {
					message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s sudah berakhir %d hari yang lalu. Silakan perpanjang atau ganti pengurus tersebut.",
						director.FullName, director.Position, companyName, daysAgo)
				}
			} else if daysUntilExpiry == 0 {
				message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s akan berakhir hari ini. Silakan perpanjang atau ganti pengurus tersebut.",
					director.FullName, director.Position, companyName)
			} else {
				message = fmt.Sprintf("Masa jabatan %s sebagai %s di %s akan berakhir dalam %d hari. Silakan perpanjang atau ganti pengurus tersebut.",
					director.FullName, director.Position, companyName, daysUntilExpiry)
			}

			// Use director ID as resource ID (no resource_type for now, or use "director" if we add it)
			directorID := director.ID
			_, err := uc.CreateNotification(
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
