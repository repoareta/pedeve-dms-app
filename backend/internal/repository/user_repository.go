package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// UserRepository interface untuk user operations
type UserRepository interface {
	Create(user *domain.UserModel) error
	GetByID(id string) (*domain.UserModel, error)
	GetByUsername(username string) (*domain.UserModel, error)
	GetByEmail(email string) (*domain.UserModel, error)
	GetByUsernameOrEmail(usernameOrEmail string) (*domain.UserModel, error)
	GetByCompanyID(companyID string) ([]domain.UserModel, error)
	GetByRoleID(roleID string) ([]domain.UserModel, error)
	GetAll() ([]domain.UserModel, error)
	Update(user *domain.UserModel) error
	Delete(id string) error
	Deactivate(id string) error
	
	// Get user with relationships
	GetUserWithRoleAndCompany(userID string) (*domain.UserModel, *domain.RoleModel, *domain.CompanyModel, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepositoryWithDB creates a new user repository with injected DB (for testing)
func NewUserRepositoryWithDB(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// NewUserRepository creates a new user repository with default DB (backward compatibility)
func NewUserRepository() UserRepository {
	return NewUserRepositoryWithDB(database.GetDB())
}

func (r *userRepository) Create(user *domain.UserModel) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*domain.UserModel, error) {
	var user domain.UserModel
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*domain.UserModel, error) {
	var user domain.UserModel
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.UserModel, error) {
	var user domain.UserModel
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsernameOrEmail(usernameOrEmail string) (*domain.UserModel, error) {
	var user domain.UserModel
	err := r.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByCompanyID(companyID string) ([]domain.UserModel, error) {
	var users []domain.UserModel
	err := r.db.Where("company_id = ?", companyID).Find(&users).Error
	return users, err
}

func (r *userRepository) GetByRoleID(roleID string) ([]domain.UserModel, error) {
	var users []domain.UserModel
	err := r.db.Where("role_id = ?", roleID).Find(&users).Error
	return users, err
}

func (r *userRepository) GetAll() ([]domain.UserModel, error) {
	var users []domain.UserModel
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *domain.UserModel) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&domain.UserModel{}, "id = ?", id).Error
}

func (r *userRepository) Deactivate(id string) error {
	return r.db.Model(&domain.UserModel{}).Where("id = ?", id).Update("is_active", false).Error
}

// GetUserWithRoleAndCompany gets user with role and company relationships
func (r *userRepository) GetUserWithRoleAndCompany(userID string) (*domain.UserModel, *domain.RoleModel, *domain.CompanyModel, error) {
	var user domain.UserModel
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, nil, nil, err
	}
	
	var role *domain.RoleModel
	if user.RoleID != nil {
		var roleModel domain.RoleModel
		if err := r.db.Where("id = ?", *user.RoleID).First(&roleModel).Error; err == nil {
			role = &roleModel
		}
	}
	
	var company *domain.CompanyModel
	if user.CompanyID != nil {
		var companyModel domain.CompanyModel
		if err := r.db.Where("id = ?", *user.CompanyID).First(&companyModel).Error; err == nil {
			company = &companyModel
		}
	}
	
	return &user, role, company, nil
}

