package repository

import (
	"fmt"
	"strings"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	ListFolders(ownerID *string) ([]domain.DocumentFolderModel, error)
	CreateFolder(folder *domain.DocumentFolderModel) error
	GetFolderByID(id string) (*domain.DocumentFolderModel, error)
	UpdateFolderName(id, name string) error
	DeleteFolder(id string) error

	ListDocuments(folderID *string) ([]domain.DocumentModel, error)
	ListDocumentsPaginated(q ListDocumentsQuery) ([]domain.DocumentModel, int64, error)
	GetDocumentByID(id string) (*domain.DocumentModel, error)
	CreateDocument(doc *domain.DocumentModel) error
	UpdateDocument(doc *domain.DocumentModel) error
	DeleteDocument(id string) error
	DeleteDocumentsByFolder(folderID string) error

	GetFolderStats(ownerID *string) ([]domain.DocumentFolderStat, error)
	GetTotalSize(ownerID *string) (int64, error)
}

type ListDocumentsQuery struct {
	FolderID   *string
	Search     string
	SortBy     string
	SortDir    string
	Page       int
	PageSize   int
	UploaderID *string
	Type       string
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository() DocumentRepository {
	return &documentRepository{db: database.GetDB()}
}

func NewDocumentRepositoryWithDB(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) ListFolders(ownerID *string) ([]domain.DocumentFolderModel, error) {
	var folders []domain.DocumentFolderModel
	tx := r.db.Order("created_at DESC")
	if ownerID != nil {
		tx = tx.Where("created_by = ?", *ownerID)
	}
	err := tx.Find(&folders).Error
	return folders, err
}

func (r *documentRepository) CreateFolder(folder *domain.DocumentFolderModel) error {
	return r.db.Create(folder).Error
}

func (r *documentRepository) GetFolderByID(id string) (*domain.DocumentFolderModel, error) {
	var folder domain.DocumentFolderModel
	if err := r.db.Where("id = ?", id).First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *documentRepository) UpdateFolderName(id, name string) error {
	return r.db.Model(&domain.DocumentFolderModel{}).
		Where("id = ?", id).
		Update("name", name).Error
}

func (r *documentRepository) DeleteFolder(id string) error {
	return r.db.Delete(&domain.DocumentFolderModel{}, "id = ?", id).Error
}

func (r *documentRepository) ListDocuments(folderID *string) ([]domain.DocumentModel, error) {
	var docs []domain.DocumentModel
	tx := r.db.Order("created_at DESC")
	if folderID != nil {
		tx = tx.Where("folder_id = ?", *folderID)
	}
	err := tx.Find(&docs).Error
	return docs, err
}

func (r *documentRepository) ListDocumentsPaginated(q ListDocumentsQuery) ([]domain.DocumentModel, int64, error) {
	var docs []domain.DocumentModel
	tx := r.db.Model(&domain.DocumentModel{})

	if q.FolderID != nil {
		tx = tx.Where("folder_id = ?", *q.FolderID)
	}
	if q.UploaderID != nil {
		tx = tx.Where("uploader_id = ?", *q.UploaderID)
	}
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		tx = tx.Where("LOWER(name) LIKE ? OR LOWER(file_name) LIKE ? OR LOWER(mime_type) LIKE ?", like, like, like)
	}
	if q.Type != "" {
		tx = tx.Where("LOWER(mime_type) LIKE ?", "%"+strings.ToLower(q.Type)+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	switch strings.ToLower(q.SortBy) {
	case "name":
		sortField = "name"
	case "size":
		sortField = "size"
	case "created_at", "updated_at":
		sortField = q.SortBy
	}
	dir := "DESC"
	if strings.ToLower(q.SortDir) == "asc" {
		dir = "ASC"
	}

	page := q.Page
	pageSize := q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	if err := tx.Order(fmt.Sprintf("%s %s", sortField, dir)).
		Limit(pageSize).
		Offset(offset).
		Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *documentRepository) GetDocumentByID(id string) (*domain.DocumentModel, error) {
	var doc domain.DocumentModel
	if err := r.db.Where("id = ?", id).First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) CreateDocument(doc *domain.DocumentModel) error {
	return r.db.Create(doc).Error
}

func (r *documentRepository) UpdateDocument(doc *domain.DocumentModel) error {
	return r.db.Model(&domain.DocumentModel{}).Where("id = ?", doc.ID).Updates(doc).Error
}

func (r *documentRepository) DeleteDocument(id string) error {
	return r.db.Delete(&domain.DocumentModel{}, "id = ?", id).Error
}

func (r *documentRepository) DeleteDocumentsByFolder(folderID string) error {
	return r.db.Delete(&domain.DocumentModel{}, "folder_id = ?", folderID).Error
}

func (r *documentRepository) GetFolderStats(ownerID *string) ([]domain.DocumentFolderStat, error) {
	var stats []domain.DocumentFolderStat
	tx := r.db.Model(&domain.DocumentModel{}).
		Select("folder_id, COUNT(*) as file_count, COALESCE(SUM(size),0) as total_size")

	if ownerID != nil {
		tx = tx.Where("uploader_id = ?", *ownerID)
	}

	err := tx.Group("folder_id").Scan(&stats).Error
	return stats, err
}

func (r *documentRepository) GetTotalSize(ownerID *string) (int64, error) {
	var total int64
	tx := r.db.Model(&domain.DocumentModel{}).Select("COALESCE(SUM(size),0)")
	if ownerID != nil {
		tx = tx.Where("uploader_id = ?", *ownerID)
	}
	err := tx.Scan(&total).Error
	return total, err
}
