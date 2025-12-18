package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// DocumentTypeRepository interface untuk document type operations
type DocumentTypeRepository interface {
	Create(docType *domain.DocumentTypeModel) error
	GetByID(id string) (*domain.DocumentTypeModel, error)
	GetByName(name string) (*domain.DocumentTypeModel, error)
	GetAll(includeInactive bool) ([]domain.DocumentTypeModel, error)
	GetActive() ([]domain.DocumentTypeModel, error)
	Update(docType *domain.DocumentTypeModel) error
	SoftDelete(id string) error // Soft delete: set is_active = false
	CountUsage(id string) (int64, error) // Count documents using this type
	IncrementUsage(id string) error // Increment usage count
	DecrementUsage(id string) error // Decrement usage count
}

type documentTypeRepository struct {
	db *gorm.DB
}

// NewDocumentTypeRepositoryWithDB creates a new document type repository with injected DB (for testing)
func NewDocumentTypeRepositoryWithDB(db *gorm.DB) DocumentTypeRepository {
	return &documentTypeRepository{
		db: db,
	}
}

// NewDocumentTypeRepository creates a new document type repository with default DB
func NewDocumentTypeRepository() DocumentTypeRepository {
	return NewDocumentTypeRepositoryWithDB(database.GetDB())
}

func (r *documentTypeRepository) Create(docType *domain.DocumentTypeModel) error {
	return r.db.Create(docType).Error
}

func (r *documentTypeRepository) GetByID(id string) (*domain.DocumentTypeModel, error) {
	var docType domain.DocumentTypeModel
	err := r.db.Where("id = ?", id).First(&docType).Error
	if err != nil {
		return nil, err
	}
	return &docType, nil
}

func (r *documentTypeRepository) GetByName(name string) (*domain.DocumentTypeModel, error) {
	var docType domain.DocumentTypeModel
	// Case-insensitive search: use LOWER() for both sides
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&docType).Error
	if err != nil {
		return nil, err
	}
	return &docType, nil
}

func (r *documentTypeRepository) GetAll(includeInactive bool) ([]domain.DocumentTypeModel, error) {
	var docTypes []domain.DocumentTypeModel
	tx := r.db.Order("name ASC")
	if !includeInactive {
		tx = tx.Where("is_active = ?", true)
	}
	err := tx.Find(&docTypes).Error
	return docTypes, err
}

func (r *documentTypeRepository) GetActive() ([]domain.DocumentTypeModel, error) {
	var docTypes []domain.DocumentTypeModel
	err := r.db.Where("is_active = ?", true).Order("name ASC").Find(&docTypes).Error
	return docTypes, err
}

func (r *documentTypeRepository) Update(docType *domain.DocumentTypeModel) error {
	return r.db.Save(docType).Error
}

func (r *documentTypeRepository) SoftDelete(id string) error {
	return r.db.Model(&domain.DocumentTypeModel{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *documentTypeRepository) CountUsage(id string) (int64, error) {
	var count int64
	// Get document type name first
	docType, err := r.GetByID(id)
	if err != nil {
		return 0, err
	}
	
	// Count documents that have this document type name in metadata.doc_type
	// Using JSON path query for PostgreSQL: metadata->>'doc_type' extracts text value
	err = r.db.Model(&domain.DocumentModel{}).
		Where("metadata->>'doc_type' = ?", docType.Name).
		Count(&count).Error
	return count, err
}

func (r *documentTypeRepository) IncrementUsage(id string) error {
	return r.db.Model(&domain.DocumentTypeModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

func (r *documentTypeRepository) DecrementUsage(id string) error {
	return r.db.Model(&domain.DocumentTypeModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("GREATEST(usage_count - 1, 0)")).Error
}

