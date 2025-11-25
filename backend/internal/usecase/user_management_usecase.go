package usecase

import (
	"errors"
	"fmt"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	passwordPkg "github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"github.com/Fajarriswandi/dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// UserManagementUseCase interface untuk user management operations
type UserManagementUseCase interface {
	CreateUser(username, email, password string, companyID, roleID *string) (*domain.UserModel, error)
	GetUserByID(id string) (*domain.UserModel, error)
	GetUsersByCompany(companyID string) ([]domain.UserModel, error)
	GetUsersByRole(roleID string) ([]domain.UserModel, error)
	GetAllUsers() ([]domain.UserModel, error)
	UpdateUser(id, username, email string, companyID, roleID *string) (*domain.UserModel, error)
	UpdateUserPassword(id, newPassword string) error
	AssignUserToCompany(userID, companyID string) error
	AssignUserToRole(userID, roleID string) error
	DeactivateUser(id string) error
	ActivateUser(id string) error
	ToggleUserStatus(id string) (*domain.UserModel, error)
	DeleteUser(id string) error
	ValidateUserAccess(userCompanyID, targetUserID string) (bool, error)
	ResetUserPassword(userID, newPassword string) error
}

type userManagementUseCase struct {
	userRepo    repository.UserRepository
	companyRepo repository.CompanyRepository
	roleRepo    repository.RoleRepository
}

// NewUserManagementUseCase creates a new user management use case
func NewUserManagementUseCase() UserManagementUseCase {
	return &userManagementUseCase{
		userRepo:    repository.NewUserRepository(),
		companyRepo: repository.NewCompanyRepository(),
		roleRepo:    repository.NewRoleRepository(),
	}
}

func (uc *userManagementUseCase) CreateUser(username, email, password string, companyID, roleID *string) (*domain.UserModel, error) {
	zapLog := logger.GetLogger()

	// Validate username uniqueness
	existing, _ := uc.userRepo.GetByUsername(username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Validate email uniqueness
	existing, _ = uc.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// Validate company if provided
	if companyID != nil {
		_, err := uc.companyRepo.GetByID(*companyID)
		if err != nil {
			return nil, fmt.Errorf("company not found: %w", err)
		}
	}

	// Validate role if provided
	if roleID != nil {
		_, err := uc.roleRepo.GetByID(*roleID)
		if err != nil {
			return nil, fmt.Errorf("role not found: %w", err)
		}
	}

	// Hash password
	hashedPassword, err := passwordPkg.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Get role name for legacy field
	roleName := "user"
	if roleID != nil {
		role, err := uc.roleRepo.GetByID(*roleID)
		if err == nil {
			roleName = role.Name
		}
	}

	user := &domain.UserModel{
		ID:        uuid.GenerateUUID(),
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      roleName, // Legacy field
		RoleID:    roleID,
		CompanyID: companyID,
		IsActive:  true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		zapLog.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *userManagementUseCase) GetUserByID(id string) (*domain.UserModel, error) {
	return uc.userRepo.GetByID(id)
}

func (uc *userManagementUseCase) GetUsersByCompany(companyID string) ([]domain.UserModel, error) {
	return uc.userRepo.GetByCompanyID(companyID)
}

func (uc *userManagementUseCase) GetUsersByRole(roleID string) ([]domain.UserModel, error) {
	return uc.userRepo.GetByRoleID(roleID)
}

func (uc *userManagementUseCase) GetAllUsers() ([]domain.UserModel, error) {
	// This should be restricted based on user's company hierarchy
	// For now, return all users (will be filtered by middleware/RLS)
	return uc.userRepo.GetAll()
}

func (uc *userManagementUseCase) UpdateUser(id, username, email string, companyID, roleID *string) (*domain.UserModel, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate username uniqueness (if changed)
	if username != "" && username != user.Username {
		existing, _ := uc.userRepo.GetByUsername(username)
		if existing != nil && existing.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = username
	}

	// Validate email uniqueness (if changed)
	if email != "" && email != user.Email {
		existing, _ := uc.userRepo.GetByEmail(email)
		if existing != nil && existing.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = email
	}

	// Validate company if provided
	if companyID != nil {
		_, err := uc.companyRepo.GetByID(*companyID)
		if err != nil {
			return nil, fmt.Errorf("company not found: %w", err)
		}
		user.CompanyID = companyID
	}

	// Validate role if provided
	if roleID != nil {
		role, err := uc.roleRepo.GetByID(*roleID)
		if err != nil {
			return nil, fmt.Errorf("role not found: %w", err)
		}
		user.RoleID = roleID
		user.Role = role.Name // Update legacy field
	}

	if err := uc.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (uc *userManagementUseCase) UpdateUserPassword(id, newPassword string) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	hashedPassword, err := passwordPkg.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword
	return uc.userRepo.Update(user)
}

func (uc *userManagementUseCase) AssignUserToCompany(userID, companyID string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	_, err = uc.companyRepo.GetByID(companyID)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	user.CompanyID = &companyID
	return uc.userRepo.Update(user)
}

func (uc *userManagementUseCase) AssignUserToRole(userID, roleID string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	role, err := uc.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	user.RoleID = &roleID
	user.Role = role.Name // Update legacy field
	return uc.userRepo.Update(user)
}

func (uc *userManagementUseCase) DeactivateUser(id string) error {
	return uc.userRepo.Deactivate(id)
}

func (uc *userManagementUseCase) ActivateUser(id string) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	user.IsActive = true
	return uc.userRepo.Update(user)
}

func (uc *userManagementUseCase) ToggleUserStatus(id string) (*domain.UserModel, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	user.IsActive = !user.IsActive
	if err := uc.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *userManagementUseCase) DeleteUser(id string) error {
	return uc.userRepo.Delete(id)
}

func (uc *userManagementUseCase) ValidateUserAccess(userCompanyID, targetUserID string) (bool, error) {
	// Get target user
	targetUser, err := uc.userRepo.GetByID(targetUserID)
	if err != nil {
		return false, fmt.Errorf("target user not found: %w", err)
	}

	// If target user has no company, only superadmin can access
	if targetUser.CompanyID == nil {
		return false, nil // Only superadmin can access users without company
	}

	// If user's company is the same as target user's company, allow
	if userCompanyID == *targetUser.CompanyID {
		return true, nil
	}

	// Check if target user's company is a descendant of user's company
	return uc.companyRepo.IsDescendantOf(*targetUser.CompanyID, userCompanyID)
}

// ResetUserPassword resets a user's password (only for superadmin)
func (uc *userManagementUseCase) ResetUserPassword(userID, newPassword string) error {
	zapLog := logger.GetLogger()

	// Get user
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate password strength (min 8 characters)
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Hash new password
	hashedPassword, err := passwordPkg.HashPassword(newPassword)
	if err != nil {
		zapLog.Error("Failed to hash password for reset", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	if err := uc.userRepo.Update(user); err != nil {
		zapLog.Error("Failed to update password", zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}

	zapLog.Info("User password reset successfully", zap.String("user_id", userID), zap.String("username", user.Username))
	return nil
}

