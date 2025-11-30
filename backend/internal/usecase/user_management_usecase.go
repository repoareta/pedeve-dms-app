package usecase

import (
	"errors"
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	passwordPkg "github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/password"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// UserManagementUseCase interface untuk user management operations
type UserManagementUseCase interface {
	CreateUser(username, email, password string, companyID, roleID *string) (*domain.UserModel, error)
	GetUserByID(id string) (*domain.UserModel, error)
	GetUsersByCompany(companyID string) ([]domain.UserModel, error)
	GetUsersByCompanyHierarchy(companyID string) ([]domain.UserModel, error) // Get users from company and all descendants (RBAC)
	GetUsersByRole(roleID string) ([]domain.UserModel, error)
	GetAllUsers() ([]domain.UserModel, error)
	UpdateUser(id, username, email string, companyID, roleID *string) (*domain.UserModel, error)
	UpdateUserPassword(id, newPassword string) error
	AssignUserToCompany(userID, companyID string) error
	AssignUserToRole(userID, roleID string) error
	AssignUserToRoleInCompany(userID, companyID, roleID string) error // Assign role in specific company via junction table
	UnassignUserFromCompany(userID, companyID string) error           // Remove user from company via junction table
	DeactivateUser(id string) error
	ActivateUser(id string) error
	ToggleUserStatus(id string) (*domain.UserModel, error)
	DeleteUser(id string) error
	ValidateUserAccess(userCompanyID, targetUserID string) (bool, error)
	ResetUserPassword(userID, newPassword string) error
	GetUserCompanies(userID string) ([]domain.UserCompanyResponse, error) // Get all companies assigned to user via junction table with role info
}

type userManagementUseCase struct {
	userRepo              repository.UserRepository
	companyRepo           repository.CompanyRepository
	roleRepo              repository.RoleRepository
	assignmentRepo        repository.UserCompanyAssignmentRepository
}

// NewUserManagementUseCase creates a new user management use case
func NewUserManagementUseCase() UserManagementUseCase {
	return &userManagementUseCase{
		userRepo:       repository.NewUserRepository(),
		companyRepo:    repository.NewCompanyRepository(),
		roleRepo:       repository.NewRoleRepository(),
		assignmentRepo: repository.NewUserCompanyAssignmentRepository(),
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
	// If no role provided, leave empty (standby mode) - DO NOT default to "user" or "superadmin"
	var roleName string
	if roleID != nil {
		role, err := uc.roleRepo.GetByID(*roleID)
		if err == nil {
			roleName = role.Name
		} else {
			// If role ID provided but not found, use empty string (standby)
			roleName = ""
		}
	} else {
		// No role ID provided - user is in standby mode
		roleName = ""
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

	// If companyID and roleID are provided, automatically create entry in junction table
	// This ensures user appears in "My Company" and can be managed via "Assign Role"
	if companyID != nil && roleID != nil {
		assignment := &domain.UserCompanyAssignmentModel{
			ID:        uuid.GenerateUUID(),
			UserID:    user.ID,
			CompanyID: *companyID,
			RoleID:    roleID,
			IsActive:  true,
		}
		if err := uc.assignmentRepo.Create(assignment); err != nil {
			// Log error but don't fail - user is already created
			// Junction table entry can be created later via "Assign Role"
			zapLog.Warn("Failed to create junction table entry for new user", 
				zap.String("user_id", user.ID),
				zap.String("company_id", *companyID),
				zap.Error(err))
		}
	} else if companyID != nil {
		// If only companyID provided (no role), create assignment without role (standby)
		assignment := &domain.UserCompanyAssignmentModel{
			ID:        uuid.GenerateUUID(),
			UserID:    user.ID,
			CompanyID: *companyID,
			RoleID:    nil, // No role assigned yet
			IsActive:  true,
		}
		if err := uc.assignmentRepo.Create(assignment); err != nil {
			zapLog.Warn("Failed to create junction table entry for new user (standby)", 
				zap.String("user_id", user.ID),
				zap.String("company_id", *companyID),
				zap.Error(err))
		}
	}

	return user, nil
}

func (uc *userManagementUseCase) GetUserByID(id string) (*domain.UserModel, error) {
	return uc.userRepo.GetByID(id)
}

func (uc *userManagementUseCase) GetUsersByCompany(companyID string) ([]domain.UserModel, error) {
	// Get users from junction table (supports multiple company assignments)
	assignments, err := uc.assignmentRepo.GetByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	// Get unique user IDs
	userIDs := make(map[string]bool)
	for _, assignment := range assignments {
		if assignment.IsActive {
			userIDs[assignment.UserID] = true
		}
	}

	// Get users by IDs
	users := []domain.UserModel{}
	for userID := range userIDs {
		user, err := uc.userRepo.GetByID(userID)
		if err != nil {
			continue // Skip if user not found
		}
		
		// Get role from assignment for this company
		for _, assignment := range assignments {
			if assignment.UserID == userID && assignment.CompanyID == companyID && assignment.RoleID != nil {
				// Add role info from assignment
				user.RoleID = assignment.RoleID
				// Get role name
				if role, err := uc.roleRepo.GetByID(*assignment.RoleID); err == nil {
					user.Role = role.Name
				}
				break
			}
		}
		
		users = append(users, *user)
	}

	return users, nil
}

// GetUsersByCompanyHierarchy gets all users from a company and all its descendants (RBAC)
// This is used for User Management to show only users that the current user has access to
func (uc *userManagementUseCase) GetUsersByCompanyHierarchy(companyID string) ([]domain.UserModel, error) {
	// Get company descendants (includes direct children and all nested descendants)
	companyUseCase := NewCompanyUseCase()
	descendants, err := companyUseCase.GetCompanyDescendants(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company descendants: %w", err)
	}
	
	// Include the company itself
	allCompanyIDs := []string{companyID}
	for _, desc := range descendants {
		allCompanyIDs = append(allCompanyIDs, desc.ID)
	}
	
	// Get all users from junction table for these companies
	allUserIDs := make(map[string]bool)
	userRoleMap := make(map[string]map[string]*string) // userID -> companyID -> roleID
	
	for _, compID := range allCompanyIDs {
		assignments, err := uc.assignmentRepo.GetByCompanyID(compID)
		if err != nil {
			continue // Skip if error getting assignments
		}
		
		for _, assignment := range assignments {
			// Include both active and inactive assignments
			// This ensures users who were unassigned still appear in the list
			allUserIDs[assignment.UserID] = true
			if userRoleMap[assignment.UserID] == nil {
				userRoleMap[assignment.UserID] = make(map[string]*string)
			}
			// Only set role if assignment is active, otherwise keep nil to indicate unassigned
			if assignment.IsActive {
				userRoleMap[assignment.UserID][compID] = assignment.RoleID
			}
		}
	}
	
	// Also get users from UserModel.CompanyID as fallback (backward compatibility)
	// This ensures users created before junction table implementation are still visible
	for _, compID := range allCompanyIDs {
		usersFromCompanyID, err := uc.userRepo.GetByCompanyID(compID)
		if err == nil {
			for _, user := range usersFromCompanyID {
				// Only add if not already in junction table
				if !allUserIDs[user.ID] {
					allUserIDs[user.ID] = true
					// Use role from UserModel if available
					if user.RoleID != nil {
						if userRoleMap[user.ID] == nil {
							userRoleMap[user.ID] = make(map[string]*string)
						}
						userRoleMap[user.ID][compID] = user.RoleID
					}
				}
			}
		}
	}
	
	// Get users by IDs
	users := []domain.UserModel{}
	for userID := range allUserIDs {
		user, err := uc.userRepo.GetByID(userID)
		if err != nil {
			continue // Skip if user not found
		}
		
		// Skip superadmin users for security
		if user.Role == "superadmin" {
			continue
		}
		
		// Get role from assignment for the primary company (user's company)
		// If user is assigned to multiple companies, use the role from the primary company
		if roleMap, ok := userRoleMap[userID]; ok {
			if roleID, ok := roleMap[companyID]; ok && roleID != nil {
				user.RoleID = roleID
				// Get role name
				if role, err := uc.roleRepo.GetByID(*roleID); err == nil {
					user.Role = role.Name
				}
			} else {
				// If not found in primary company, use first available role
				for compID, roleID := range roleMap {
					if roleID != nil {
						user.RoleID = roleID
						if role, err := uc.roleRepo.GetByID(*roleID); err == nil {
							user.Role = role.Name
						}
						break
					}
					_ = compID // Suppress unused variable warning
				}
				// If no active role found, check if user was unassigned
				// User will appear without role or with empty role
				if user.RoleID == nil && user.Role == "" {
					// User was unassigned - keep role empty
					user.Role = ""
				}
			}
		} else if user.RoleID != nil {
			// Fallback: use role from UserModel if no junction table entry
			if role, err := uc.roleRepo.GetByID(*user.RoleID); err == nil {
				user.Role = role.Name
			}
		} else {
			// No role found anywhere - user is in standby/unassigned state
			user.Role = ""
		}
		
		users = append(users, *user)
	}
	
	return users, nil
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

	// Handle company assignment/unassignment
	if companyID != nil {
		if *companyID == "" {
			// Unassign from company (empty string means unassign)
			user.CompanyID = nil
		} else {
			// Assign to company
			_, err := uc.companyRepo.GetByID(*companyID)
			if err != nil {
				return nil, fmt.Errorf("company not found: %w", err)
			}
			user.CompanyID = companyID
		}
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

	// Update junction table if companyID and/or roleID changed
	zapLog := logger.GetLogger()
	if companyID != nil && *companyID != "" {
		// Assign to company - update/create junction table entry
		// Check if assignment already exists for this user-company pair
		existingAssignment, err := uc.assignmentRepo.GetByUserAndCompany(id, *companyID)
		if err == nil && existingAssignment != nil {
			// Assignment exists - update role if roleID provided
			if roleID != nil {
				existingAssignment.RoleID = roleID
				existingAssignment.IsActive = true
				if err := uc.assignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to update junction table assignment", 
						zap.String("user_id", id),
						zap.String("company_id", *companyID),
						zap.Error(err))
				}
			} else {
				// No role provided - just activate assignment
				existingAssignment.IsActive = true
				if err := uc.assignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to activate junction table assignment", 
						zap.String("user_id", id),
						zap.String("company_id", *companyID),
						zap.Error(err))
				}
			}
		} else {
			// Assignment doesn't exist - create new one
			assignment := &domain.UserCompanyAssignmentModel{
				ID:        uuid.GenerateUUID(),
				UserID:    id,
				CompanyID: *companyID,
				RoleID:    roleID,
				IsActive:  true,
			}
			if err := uc.assignmentRepo.Create(assignment); err != nil {
				zapLog.Warn("Failed to create junction table assignment", 
					zap.String("user_id", id),
					zap.String("company_id", *companyID),
					zap.Error(err))
			}
		}
	} else if companyID != nil && *companyID == "" {
		// Unassign from company - deactivate all assignments
		assignments, err := uc.assignmentRepo.GetByUserID(id)
		if err == nil {
			for i := range assignments {
				if assignments[i].IsActive {
					assignments[i].IsActive = false
					if err := uc.assignmentRepo.Update(&assignments[i]); err != nil {
						zapLog.Warn("Failed to deactivate assignment", 
							zap.String("user_id", id),
							zap.String("assignment_id", assignments[i].ID),
							zap.Error(err))
					}
				}
			}
		}
	} else if roleID != nil {
		// Only roleID changed, but no companyID - update all active assignments for this user
		// This is a bit unusual, but we'll update the primary company assignment if exists
		if user.CompanyID != nil {
			existingAssignment, err := uc.assignmentRepo.GetByUserAndCompany(id, *user.CompanyID)
			if err == nil && existingAssignment != nil {
				existingAssignment.RoleID = roleID
				existingAssignment.IsActive = true
				if err := uc.assignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to update junction table assignment role", 
						zap.String("user_id", id),
						zap.String("company_id", *user.CompanyID),
						zap.Error(err))
				}
			}
		}
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

	// Check if assignment already exists
	existingAssignment, err := uc.assignmentRepo.GetByUserAndCompany(userID, companyID)
	if err == nil && existingAssignment != nil {
		// Assignment already exists, just activate it
		if !existingAssignment.IsActive {
			existingAssignment.IsActive = true
			return uc.assignmentRepo.Update(existingAssignment)
		}
		return nil // Already assigned
	}

	// Create new assignment in junction table (supports multiple company assignments)
	assignment := &domain.UserCompanyAssignmentModel{
		ID:        uuid.GenerateUUID(),
		UserID:    userID,
		CompanyID: companyID,
		RoleID:    nil, // Role can be assigned separately via AssignUserToRoleInCompany
		IsActive:  true,
	}

	if err := uc.assignmentRepo.Create(assignment); err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	// Also update UserModel.CompanyID for backward compatibility (set as primary company if null)
	if user.CompanyID == nil {
		user.CompanyID = &companyID
		return uc.userRepo.Update(user)
	}

	return nil
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

// AssignUserToRoleInCompany assigns a role to a user in a specific company via junction table
// This allows the same user to have different roles in different companies
func (uc *userManagementUseCase) AssignUserToRoleInCompany(userID, companyID, roleID string) error {
	// Validate user exists
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate company exists
	_, err = uc.companyRepo.GetByID(companyID)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Validate role exists
	_, err = uc.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Get role name for legacy field sync
	role, err := uc.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Get or create assignment
	assignment, err := uc.assignmentRepo.GetByUserAndCompany(userID, companyID)
	if err != nil {
		// Assignment doesn't exist, create it
		assignment = &domain.UserCompanyAssignmentModel{
			ID:        uuid.GenerateUUID(),
			UserID:    userID,
			CompanyID: companyID,
			RoleID:    &roleID,
			IsActive:  true,
		}
		if err := uc.assignmentRepo.Create(assignment); err != nil {
			return fmt.Errorf("failed to create assignment: %w", err)
		}
	} else {
		// Assignment exists, update role
		assignment.RoleID = &roleID
		assignment.IsActive = true
		if err := uc.assignmentRepo.Update(assignment); err != nil {
			return fmt.Errorf("failed to update assignment: %w", err)
		}
	}

	// Also update UserModel.Role and UserModel.RoleID for backward compatibility
	// This ensures User Management and other parts that read from users table see the correct role
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	user.RoleID = &roleID
	user.Role = role.Name // Update legacy field for backward compatibility
	if err := uc.userRepo.Update(user); err != nil {
		// Log error but don't fail - junction table is the source of truth
		zapLog := logger.GetLogger()
		zapLog.Warn("Failed to sync role to users table", zap.Error(err))
	}

	return nil
}

// UnassignUserFromCompany removes a user from a company via junction table
// This allows user to be removed from one company while keeping assignments to other companies
func (uc *userManagementUseCase) UnassignUserFromCompany(userID, companyID string) error {
	// Validate user exists
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Validate company exists
	_, err = uc.companyRepo.GetByID(companyID)
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Remove assignment from junction table
	if err := uc.assignmentRepo.DeleteByUserAndCompany(userID, companyID); err != nil {
		return fmt.Errorf("failed to remove assignment: %w", err)
	}

	// Also check if this was the primary company (UserModel.CompanyID)
	user, _ := uc.userRepo.GetByID(userID)
	if user.CompanyID != nil && *user.CompanyID == companyID {
		// This was the primary company, set to null
		user.CompanyID = nil
		return uc.userRepo.Update(user)
	}

	return nil
}

// GetUserCompanies gets all companies assigned to a user via junction table
// Returns companies with their role information for that user
func (uc *userManagementUseCase) GetUserCompanies(userID string) ([]domain.UserCompanyResponse, error) {
	// Get all assignments for this user
	assignments, err := uc.assignmentRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	zapLog := logger.GetLogger()
	
	// Get companies from assignments with role info
	companies := make([]domain.UserCompanyResponse, 0, len(assignments))
	for _, assignment := range assignments {
		if !assignment.IsActive {
			continue // Skip inactive assignments
		}
		
		company, err := uc.companyRepo.GetByID(assignment.CompanyID)
		if err != nil {
			// Log but continue - assignment might reference deleted company
			zapLog.Warn("Company not found for assignment", zap.String("company_id", assignment.CompanyID), zap.Error(err))
			continue
		}
		
		// Get role info
		var roleID *string
		roleName := ""
		roleLevel := 999 // Default high level
		
		if assignment.RoleID != nil {
			roleID = assignment.RoleID
			role, err := uc.roleRepo.GetByID(*assignment.RoleID)
			if err == nil {
				roleName = role.Name
				roleLevel = role.Level
			} else {
				zapLog.Warn("Role not found for assignment", zap.String("role_id", *assignment.RoleID), zap.Error(err))
			}
		}
		
		companies = append(companies, domain.UserCompanyResponse{
			Company:    *company,
			RoleID:     roleID,
			Role:       roleName,
			RoleLevel:  roleLevel,
		})
	}

	return companies, nil
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
	zapLog := logger.GetLogger()
	
	// First, delete all user-company assignments from junction table
	// This prevents foreign key constraint violation
	if err := uc.assignmentRepo.DeleteByUserID(id); err != nil {
		zapLog.Warn("Failed to delete user company assignments", 
			zap.String("user_id", id),
			zap.Error(err))
		// Continue even if assignment deletion fails - user might not have any assignments
	}
	
	// Then delete the user
	if err := uc.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
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

