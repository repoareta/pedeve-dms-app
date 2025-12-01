package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// RoleRepository interface untuk role operations
type RoleRepository interface {
	Create(role *domain.RoleModel) error
	GetByID(id string) (*domain.RoleModel, error)
	GetByName(name string) (*domain.RoleModel, error)
	GetAll() ([]domain.RoleModel, error)
	GetByLevel(level int) ([]domain.RoleModel, error)
	Update(role *domain.RoleModel) error
	Delete(id string) error
	
	// Permission management
	AssignPermission(roleID, permissionID string) error
	RevokePermission(roleID, permissionID string) error
	GetPermissions(roleID string) ([]domain.PermissionModel, error)
	HasPermission(roleID, permissionName string) (bool, error)
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepositoryWithDB creates a new role repository with injected DB (for testing)
func NewRoleRepositoryWithDB(db *gorm.DB) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

// NewRoleRepository creates a new role repository with default DB (backward compatibility)
func NewRoleRepository() RoleRepository {
	return NewRoleRepositoryWithDB(database.GetDB())
}

func (r *roleRepository) Create(role *domain.RoleModel) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id string) (*domain.RoleModel, error) {
	var role domain.RoleModel
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*domain.RoleModel, error) {
	var role domain.RoleModel
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetAll() ([]domain.RoleModel, error) {
	var roles []domain.RoleModel
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByLevel(level int) ([]domain.RoleModel, error) {
	var roles []domain.RoleModel
	err := r.db.Where("level = ?", level).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *domain.RoleModel) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id string) error {
	// Check if role is system role
	var role domain.RoleModel
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		return err
	}
	
	if role.IsSystem {
		return gorm.ErrInvalidData // System role cannot be deleted
	}
	
	return r.db.Delete(&domain.RoleModel{}, "id = ?", id).Error
}

func (r *roleRepository) AssignPermission(roleID, permissionID string) error {
	// Check if already assigned
	var existing domain.RolePermissionModel
	err := r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&existing).Error
	if err == nil {
		// Already assigned
		return nil
	}
	
	// Assign permission
	return r.db.Create(&domain.RolePermissionModel{
		RoleID:       roleID,
		PermissionID: permissionID,
	}).Error
}

func (r *roleRepository) RevokePermission(roleID, permissionID string) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&domain.RolePermissionModel{}).Error
}

func (r *roleRepository) GetPermissions(roleID string) ([]domain.PermissionModel, error) {
	var permissions []domain.PermissionModel
	err := r.db.Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) HasPermission(roleID, permissionName string) (bool, error) {
	var count int64
	err := r.db.Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.name = ?", roleID, permissionName).
		Count(&count).Error
	return count > 0, err
}

