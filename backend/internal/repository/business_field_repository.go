package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type BusinessFieldRepository interface {
	Create(businessField *domain.BusinessFieldModel) error
	GetByCompanyID(companyID string) (*domain.BusinessFieldModel, error)
	Update(businessField *domain.BusinessFieldModel) error
	DeleteByCompanyID(companyID string) error
}

type businessFieldRepository struct {
	db *gorm.DB
}

// NewBusinessFieldRepositoryWithDB creates a new business field repository with injected DB (for testing)
func NewBusinessFieldRepositoryWithDB(db *gorm.DB) BusinessFieldRepository {
	return &businessFieldRepository{
		db: db,
	}
}

// NewBusinessFieldRepository creates a new business field repository with default DB (backward compatibility)
func NewBusinessFieldRepository() BusinessFieldRepository {
	return NewBusinessFieldRepositoryWithDB(database.GetDB())
}

func (r *businessFieldRepository) Create(businessField *domain.BusinessFieldModel) error {
	return r.db.Create(businessField).Error
}

func (r *businessFieldRepository) GetByCompanyID(companyID string) (*domain.BusinessFieldModel, error) {
	var businessField domain.BusinessFieldModel
	err := r.db.Where("company_id = ? AND is_main = ?", companyID, true).First(&businessField).Error
	if err != nil {
		return nil, err
	}
	return &businessField, nil
}

func (r *businessFieldRepository) Update(businessField *domain.BusinessFieldModel) error {
	return r.db.Save(businessField).Error
}

func (r *businessFieldRepository) DeleteByCompanyID(companyID string) error {
	return r.db.Where("company_id = ?", companyID).Delete(&domain.BusinessFieldModel{}).Error
}

