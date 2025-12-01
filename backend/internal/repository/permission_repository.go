package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// PermissionRepository interface untuk permission operations
type PermissionRepository interface {
	Create(permission *domain.PermissionModel) error
	GetByID(id string) (*domain.PermissionModel, error)
	GetByName(name string) (*domain.PermissionModel, error)
	GetAll() ([]domain.PermissionModel, error)
	GetByResource(resource string) ([]domain.PermissionModel, error)
	GetByScope(scope domain.PermissionScope) ([]domain.PermissionModel, error)
	Update(permission *domain.PermissionModel) error
	Delete(id string) error
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepositoryWithDB creates a new permission repository with injected DB (for testing)
func NewPermissionRepositoryWithDB(db *gorm.DB) PermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

// NewPermissionRepository creates a new permission repository with default DB (backward compatibility)
func NewPermissionRepository() PermissionRepository {
	return NewPermissionRepositoryWithDB(database.GetDB())
}

func (r *permissionRepository) Create(permission *domain.PermissionModel) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) GetByID(id string) (*domain.PermissionModel, error) {
	var permission domain.PermissionModel
	err := r.db.Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByName(name string) (*domain.PermissionModel, error) {
	var permission domain.PermissionModel
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetAll() ([]domain.PermissionModel, error) {
	var permissions []domain.PermissionModel
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByResource(resource string) ([]domain.PermissionModel, error) {
	var permissions []domain.PermissionModel
	err := r.db.Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByScope(scope domain.PermissionScope) ([]domain.PermissionModel, error) {
	var permissions []domain.PermissionModel
	err := r.db.Where("scope = ?", scope).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) Update(permission *domain.PermissionModel) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepository) Delete(id string) error {
	return r.db.Delete(&domain.PermissionModel{}, "id = ?", id).Error
}

