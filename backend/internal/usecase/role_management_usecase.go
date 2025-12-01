package usecase

import (
	"errors"
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleManagementUseCase interface untuk role management operations
type RoleManagementUseCase interface {
	CreateRole(name, description string, level int) (*domain.RoleModel, error)
	GetRoleByID(id string) (*domain.RoleModel, error)
	GetRoleByName(name string) (*domain.RoleModel, error)
	GetAllRoles() ([]domain.RoleModel, error)
	GetRolesByLevel(level int) ([]domain.RoleModel, error)
	UpdateRole(id, name, description string, level int) (*domain.RoleModel, error)
	DeleteRole(id string) error
	
	// Permission management
	AssignPermissionToRole(roleID, permissionID string) error
	RevokePermissionFromRole(roleID, permissionID string) error
	GetRolePermissions(roleID string) ([]domain.PermissionModel, error)
	RoleHasPermission(roleID, permissionName string) (bool, error)
}

// PermissionManagementUseCase interface untuk permission management operations
type PermissionManagementUseCase interface {
	CreatePermission(name, description, resource, action string, scope domain.PermissionScope) (*domain.PermissionModel, error)
	GetPermissionByID(id string) (*domain.PermissionModel, error)
	GetPermissionByName(name string) (*domain.PermissionModel, error)
	GetAllPermissions() ([]domain.PermissionModel, error)
	GetPermissionsByResource(resource string) ([]domain.PermissionModel, error)
	GetPermissionsByScope(scope domain.PermissionScope) ([]domain.PermissionModel, error)
	UpdatePermission(id, name, description string) (*domain.PermissionModel, error)
	DeletePermission(id string) error
}

type roleManagementUseCase struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

type permissionManagementUseCase struct {
	permissionRepo repository.PermissionRepository
}

// NewRoleManagementUseCaseWithDB creates a new role management use case with injected DB (for testing)
func NewRoleManagementUseCaseWithDB(db *gorm.DB) RoleManagementUseCase {
	return &roleManagementUseCase{
		roleRepo:       repository.NewRoleRepositoryWithDB(db),
		permissionRepo: repository.NewPermissionRepositoryWithDB(db),
	}
}

// NewRoleManagementUseCase creates a new role management use case with default DB (backward compatibility)
func NewRoleManagementUseCase() RoleManagementUseCase {
	return NewRoleManagementUseCaseWithDB(database.GetDB())
}

// NewPermissionManagementUseCase creates a new permission management use case
func NewPermissionManagementUseCase() PermissionManagementUseCase {
	return &permissionManagementUseCase{
		permissionRepo: repository.NewPermissionRepository(),
	}
}

// Role Management Methods
func (uc *roleManagementUseCase) CreateRole(name, description string, level int) (*domain.RoleModel, error) {
	zapLog := logger.GetLogger()

	// Validate name uniqueness
	existing, _ := uc.roleRepo.GetByName(name)
	if existing != nil {
		return nil, errors.New("role name already exists")
	}

	role := &domain.RoleModel{
		ID:          uuid.GenerateUUID(),
		Name:        name,
		Description: description,
		Level:       level,
		IsSystem:    false, // User-created roles are not system roles
	}

	if err := uc.roleRepo.Create(role); err != nil {
		zapLog.Error("Failed to create role", zap.Error(err))
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (uc *roleManagementUseCase) GetRoleByID(id string) (*domain.RoleModel, error) {
	return uc.roleRepo.GetByID(id)
}

func (uc *roleManagementUseCase) GetRoleByName(name string) (*domain.RoleModel, error) {
	return uc.roleRepo.GetByName(name)
}

func (uc *roleManagementUseCase) GetAllRoles() ([]domain.RoleModel, error) {
	return uc.roleRepo.GetAll()
}

func (uc *roleManagementUseCase) GetRolesByLevel(level int) ([]domain.RoleModel, error) {
	return uc.roleRepo.GetByLevel(level)
}

func (uc *roleManagementUseCase) UpdateRole(id, name, description string, level int) (*domain.RoleModel, error) {
	role, err := uc.roleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// System roles cannot be modified
	if role.IsSystem {
		return nil, errors.New("system roles cannot be modified")
	}

	// Validate name uniqueness (if changed)
	if name != "" && name != role.Name {
		existing, _ := uc.roleRepo.GetByName(name)
		if existing != nil && existing.ID != id {
			return nil, errors.New("role name already exists")
		}
		role.Name = name
	}

	if description != "" {
		role.Description = description
	}
	role.Level = level

	if err := uc.roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

func (uc *roleManagementUseCase) DeleteRole(id string) error {
	return uc.roleRepo.Delete(id)
}

func (uc *roleManagementUseCase) AssignPermissionToRole(roleID, permissionID string) error {
	// Validate role exists
	_, err := uc.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Validate permission exists
	_, err = uc.permissionRepo.GetByID(permissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return uc.roleRepo.AssignPermission(roleID, permissionID)
}

func (uc *roleManagementUseCase) RevokePermissionFromRole(roleID, permissionID string) error {
	return uc.roleRepo.RevokePermission(roleID, permissionID)
}

func (uc *roleManagementUseCase) GetRolePermissions(roleID string) ([]domain.PermissionModel, error) {
	return uc.roleRepo.GetPermissions(roleID)
}

func (uc *roleManagementUseCase) RoleHasPermission(roleID, permissionName string) (bool, error) {
	return uc.roleRepo.HasPermission(roleID, permissionName)
}

// Permission Management Methods
func (uc *permissionManagementUseCase) CreatePermission(name, description, resource, action string, scope domain.PermissionScope) (*domain.PermissionModel, error) {
	zapLog := logger.GetLogger()

	// Validate name uniqueness
	existing, _ := uc.permissionRepo.GetByName(name)
	if existing != nil {
		return nil, errors.New("permission name already exists")
	}

	permission := &domain.PermissionModel{
		ID:          uuid.GenerateUUID(),
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		Scope:       scope,
	}

	if err := uc.permissionRepo.Create(permission); err != nil {
		zapLog.Error("Failed to create permission", zap.Error(err))
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (uc *permissionManagementUseCase) GetPermissionByID(id string) (*domain.PermissionModel, error) {
	return uc.permissionRepo.GetByID(id)
}

func (uc *permissionManagementUseCase) GetPermissionByName(name string) (*domain.PermissionModel, error) {
	return uc.permissionRepo.GetByName(name)
}

func (uc *permissionManagementUseCase) GetAllPermissions() ([]domain.PermissionModel, error) {
	return uc.permissionRepo.GetAll()
}

func (uc *permissionManagementUseCase) GetPermissionsByResource(resource string) ([]domain.PermissionModel, error) {
	return uc.permissionRepo.GetByResource(resource)
}

func (uc *permissionManagementUseCase) GetPermissionsByScope(scope domain.PermissionScope) ([]domain.PermissionModel, error) {
	return uc.permissionRepo.GetByScope(scope)
}

func (uc *permissionManagementUseCase) UpdatePermission(id, name, description string) (*domain.PermissionModel, error) {
	permission, err := uc.permissionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Validate name uniqueness (if changed)
	if name != "" && name != permission.Name {
		existing, _ := uc.permissionRepo.GetByName(name)
		if existing != nil && existing.ID != id {
			return nil, errors.New("permission name already exists")
		}
		permission.Name = name
	}

	if description != "" {
		permission.Description = description
	}

	if err := uc.permissionRepo.Update(permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

func (uc *permissionManagementUseCase) DeletePermission(id string) error {
	return uc.permissionRepo.Delete(id)
}

