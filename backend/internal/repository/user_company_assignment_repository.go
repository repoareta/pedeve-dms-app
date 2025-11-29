package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// UserCompanyAssignmentRepository interface untuk user-company assignment operations
type UserCompanyAssignmentRepository interface {
	Create(assignment *domain.UserCompanyAssignmentModel) error
	GetByID(id string) (*domain.UserCompanyAssignmentModel, error)
	GetByUserID(userID string) ([]domain.UserCompanyAssignmentModel, error)
	GetByCompanyID(companyID string) ([]domain.UserCompanyAssignmentModel, error)
	GetByUserAndCompany(userID, companyID string) (*domain.UserCompanyAssignmentModel, error)
	GetAll() ([]domain.UserCompanyAssignmentModel, error)
	Update(assignment *domain.UserCompanyAssignmentModel) error
	Delete(id string) error
	DeleteByUserAndCompany(userID, companyID string) error
}

type userCompanyAssignmentRepository struct {
	db *gorm.DB
}

// NewUserCompanyAssignmentRepository creates a new user-company assignment repository
func NewUserCompanyAssignmentRepository() UserCompanyAssignmentRepository {
	return &userCompanyAssignmentRepository{
		db: database.GetDB(),
	}
}

func (r *userCompanyAssignmentRepository) Create(assignment *domain.UserCompanyAssignmentModel) error {
	return r.db.Create(assignment).Error
}

func (r *userCompanyAssignmentRepository) GetByID(id string) (*domain.UserCompanyAssignmentModel, error) {
	var assignment domain.UserCompanyAssignmentModel
	err := r.db.Where("id = ?", id).First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *userCompanyAssignmentRepository) GetByUserID(userID string) ([]domain.UserCompanyAssignmentModel, error) {
	var assignments []domain.UserCompanyAssignmentModel
	err := r.db.Where("user_id = ?", userID).Find(&assignments).Error
	return assignments, err
}

func (r *userCompanyAssignmentRepository) GetByCompanyID(companyID string) ([]domain.UserCompanyAssignmentModel, error) {
	var assignments []domain.UserCompanyAssignmentModel
	err := r.db.Where("company_id = ?", companyID).Find(&assignments).Error
	return assignments, err
}

func (r *userCompanyAssignmentRepository) GetByUserAndCompany(userID, companyID string) (*domain.UserCompanyAssignmentModel, error) {
	var assignment domain.UserCompanyAssignmentModel
	err := r.db.Where("user_id = ? AND company_id = ?", userID, companyID).First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *userCompanyAssignmentRepository) GetAll() ([]domain.UserCompanyAssignmentModel, error) {
	var assignments []domain.UserCompanyAssignmentModel
	err := r.db.Find(&assignments).Error
	return assignments, err
}

func (r *userCompanyAssignmentRepository) Update(assignment *domain.UserCompanyAssignmentModel) error {
	return r.db.Save(assignment).Error
}

func (r *userCompanyAssignmentRepository) Delete(id string) error {
	return r.db.Delete(&domain.UserCompanyAssignmentModel{}, "id = ?", id).Error
}

func (r *userCompanyAssignmentRepository) DeleteByUserAndCompany(userID, companyID string) error {
	return r.db.Where("user_id = ? AND company_id = ?", userID, companyID).Delete(&domain.UserCompanyAssignmentModel{}).Error
}

