package repository

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type ShareholderRepository interface {
	Create(shareholder *domain.ShareholderModel) error
	GetByCompanyID(companyID string) ([]domain.ShareholderModel, error)
	DeleteByCompanyID(companyID string) error
	Delete(id string) error
}

type shareholderRepository struct {
	db *gorm.DB
}

func NewShareholderRepository() ShareholderRepository {
	return &shareholderRepository{
		db: database.GetDB(),
	}
}

func (r *shareholderRepository) Create(shareholder *domain.ShareholderModel) error {
	return r.db.Create(shareholder).Error
}

func (r *shareholderRepository) GetByCompanyID(companyID string) ([]domain.ShareholderModel, error) {
	var shareholders []domain.ShareholderModel
	err := r.db.Where("company_id = ?", companyID).Find(&shareholders).Error
	return shareholders, err
}

func (r *shareholderRepository) DeleteByCompanyID(companyID string) error {
	return r.db.Where("company_id = ?", companyID).Delete(&domain.ShareholderModel{}).Error
}

func (r *shareholderRepository) Delete(id string) error {
	return r.db.Delete(&domain.ShareholderModel{}, "id = ?", id).Error
}

