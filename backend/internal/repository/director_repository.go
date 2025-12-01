package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type DirectorRepository interface {
	Create(director *domain.DirectorModel) error
	GetByCompanyID(companyID string) ([]domain.DirectorModel, error)
	DeleteByCompanyID(companyID string) error
	Delete(id string) error
}

type directorRepository struct {
	db *gorm.DB
}

// NewDirectorRepositoryWithDB creates a new director repository with injected DB (for testing)
func NewDirectorRepositoryWithDB(db *gorm.DB) DirectorRepository {
	return &directorRepository{
		db: db,
	}
}

// NewDirectorRepository creates a new director repository with default DB (backward compatibility)
func NewDirectorRepository() DirectorRepository {
	return NewDirectorRepositoryWithDB(database.GetDB())
}

func (r *directorRepository) Create(director *domain.DirectorModel) error {
	return r.db.Create(director).Error
}

func (r *directorRepository) GetByCompanyID(companyID string) ([]domain.DirectorModel, error) {
	var directors []domain.DirectorModel
	err := r.db.Where("company_id = ?", companyID).Find(&directors).Error
	return directors, err
}

func (r *directorRepository) DeleteByCompanyID(companyID string) error {
	return r.db.Where("company_id = ?", companyID).Delete(&domain.DirectorModel{}).Error
}

func (r *directorRepository) Delete(id string) error {
	return r.db.Delete(&domain.DirectorModel{}, "id = ?", id).Error
}

