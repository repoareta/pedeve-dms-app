package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// ShareholderTypeRepository interface untuk shareholder type operations
type ShareholderTypeRepository interface {
	Create(shareholderType *domain.ShareholderTypeModel) error
	GetByID(id string) (*domain.ShareholderTypeModel, error)
	GetByName(name string) (*domain.ShareholderTypeModel, error)
	GetAll(includeInactive bool) ([]domain.ShareholderTypeModel, error)
	GetActive() ([]domain.ShareholderTypeModel, error)
	Update(shareholderType *domain.ShareholderTypeModel) error
	SoftDelete(id string) error // Soft delete: set is_active = false
	CountUsage(id string) (int64, error) // Count shareholders using this type
	IncrementUsage(id string) error // Increment usage count
	DecrementUsage(id string) error // Decrement usage count
}

type shareholderTypeRepository struct {
	db *gorm.DB
}

// NewShareholderTypeRepositoryWithDB creates a new shareholder type repository with injected DB (for testing)
func NewShareholderTypeRepositoryWithDB(db *gorm.DB) ShareholderTypeRepository {
	return &shareholderTypeRepository{
		db: db,
	}
}

// NewShareholderTypeRepository creates a new shareholder type repository with default DB
func NewShareholderTypeRepository() ShareholderTypeRepository {
	return NewShareholderTypeRepositoryWithDB(database.GetDB())
}

func (r *shareholderTypeRepository) Create(shareholderType *domain.ShareholderTypeModel) error {
	return r.db.Create(shareholderType).Error
}

func (r *shareholderTypeRepository) GetByID(id string) (*domain.ShareholderTypeModel, error) {
	var shareholderType domain.ShareholderTypeModel
	err := r.db.Where("id = ?", id).First(&shareholderType).Error
	if err != nil {
		return nil, err
	}
	return &shareholderType, nil
}

func (r *shareholderTypeRepository) GetByName(name string) (*domain.ShareholderTypeModel, error) {
	var shareholderType domain.ShareholderTypeModel
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&shareholderType).Error
	if err != nil {
		return nil, err
	}
	return &shareholderType, nil
}

func (r *shareholderTypeRepository) GetAll(includeInactive bool) ([]domain.ShareholderTypeModel, error) {
	var shareholderTypes []domain.ShareholderTypeModel
	query := r.db
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("name ASC").Find(&shareholderTypes).Error
	return shareholderTypes, err
}

func (r *shareholderTypeRepository) GetActive() ([]domain.ShareholderTypeModel, error) {
	var shareholderTypes []domain.ShareholderTypeModel
	err := r.db.Where("is_active = ?", true).Order("name ASC").Find(&shareholderTypes).Error
	return shareholderTypes, err
}

func (r *shareholderTypeRepository) Update(shareholderType *domain.ShareholderTypeModel) error {
	return r.db.Save(shareholderType).Error
}

func (r *shareholderTypeRepository) SoftDelete(id string) error {
	return r.db.Model(&domain.ShareholderTypeModel{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *shareholderTypeRepository) CountUsage(id string) (int64, error) {
	// First get the shareholder type name
	shareholderType, err := r.GetByID(id)
	if err != nil {
		return 0, err
	}
	
	var count int64
	// Count shareholders that use this type (check if type string contains the shareholder type name)
	// Type is stored as comma-separated string in ShareholderModel.Type
	err = r.db.Model(&domain.ShareholderModel{}).
		Where("type LIKE ?", "%"+shareholderType.Name+"%"). // Check if type string contains the name
		Count(&count).Error
	return count, err
}

func (r *shareholderTypeRepository) IncrementUsage(id string) error {
	return r.db.Model(&domain.ShareholderTypeModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *shareholderTypeRepository) DecrementUsage(id string) error {
	return r.db.Model(&domain.ShareholderTypeModel{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("GREATEST(usage_count - 1, 0)")).Error
}

