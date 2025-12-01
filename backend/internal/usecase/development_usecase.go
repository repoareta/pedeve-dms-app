package usecase

import (
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DevelopmentUseCase handles development-related operations (seeding, resetting data)
type DevelopmentUseCase interface {
	ResetSubsidiaryData() error
	RunSubsidiarySeeder() (bool, error) // Returns (alreadyExists, error)
	CheckSeederDataExists() (bool, error)
}

type developmentUseCase struct {
	companyRepo              repository.CompanyRepository
	userRepo                 repository.UserRepository
	roleRepo                 repository.RoleRepository
	userCompanyAssignmentRepo repository.UserCompanyAssignmentRepository
}

// NewDevelopmentUseCaseWithDB creates a new development use case with injected DB (for testing)
func NewDevelopmentUseCaseWithDB(db *gorm.DB) DevelopmentUseCase {
	return &developmentUseCase{
		companyRepo:              repository.NewCompanyRepositoryWithDB(db),
		userRepo:                 repository.NewUserRepositoryWithDB(db),
		roleRepo:                 repository.NewRoleRepositoryWithDB(db),
		userCompanyAssignmentRepo: repository.NewUserCompanyAssignmentRepositoryWithDB(db),
	}
}

// NewDevelopmentUseCase creates a new development use case with default DB (backward compatibility)
func NewDevelopmentUseCase() DevelopmentUseCase {
	return NewDevelopmentUseCaseWithDB(database.GetDB())
}

// ResetSubsidiaryData deletes all subsidiary companies and their related users
// This will:
// 1. Get all companies except the root holding (parent_id IS NULL)
// 2. Get all descendants of each subsidiary
// 3. Delete all user_company_assignments for these companies
// 4. Delete all users assigned to these companies (except superadmin)
// 5. Delete all companies (soft delete: set is_active = false)
func (uc *developmentUseCase) ResetSubsidiaryData() error {
	zapLog := logger.GetLogger()
	db := database.GetDB()

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			zapLog.Error("Panic during reset subsidiary data", zap.Any("panic", r))
		}
	}()

	// 1. Get all companies except root holding
	// IMPORTANT: Exclude holding by BOTH parent_id IS NULL AND code != 'PDV' untuk safety
	// Ini memastikan holding tidak terhapus meskipun ada bug di data
	var allCompanies []domain.CompanyModel
	if err := tx.Where("parent_id IS NOT NULL AND code != ?", "PDV").Find(&allCompanies).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get companies: %w", err)
	}
	
	// Double check: juga exclude holding by code untuk extra safety
	// Filter out any company with code 'PDV' (holding) just in case
	filteredCompanies := make([]domain.CompanyModel, 0)
	for _, comp := range allCompanies {
		if comp.Code != "PDV" {
			filteredCompanies = append(filteredCompanies, comp)
		}
	}
	allCompanies = filteredCompanies

	if len(allCompanies) == 0 {
		tx.Rollback()
		zapLog.Info("No subsidiary companies found to reset")
		return nil
	}

	// Collect all company IDs (including descendants)
	companyIDs := make([]string, 0, len(allCompanies))
	for _, comp := range allCompanies {
		companyIDs = append(companyIDs, comp.ID)
	}

	zapLog.Info("Resetting subsidiary data", zap.Int("company_count", len(companyIDs)))

	// 2. First, collect all user IDs that will be affected BEFORE deleting assignments
	// Get user IDs from junction table assignments
	var assignments []domain.UserCompanyAssignmentModel
	if err := tx.Where("company_id IN ?", companyIDs).Find(&assignments).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get user company assignments: %w", err)
	}
	
	// Collect unique user IDs from assignments
	userIDsFromAssignments := make(map[string]bool)
	for _, assignment := range assignments {
		userIDsFromAssignments[assignment.UserID] = true
	}
	
	// Also get users from UserModel.CompanyID
	var usersFromCompanyID []domain.UserModel
	if err := tx.Where("company_id IN ?", companyIDs).Find(&usersFromCompanyID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get users from company_id: %w", err)
	}
	
	// Combine user IDs from both sources
	allUserIDs := make(map[string]bool)
	for userID := range userIDsFromAssignments {
		allUserIDs[userID] = true
	}
	for _, user := range usersFromCompanyID {
		allUserIDs[user.ID] = true
	}

	// Filter out superadmin users
	userIDsToDelete := make([]string, 0)
	for userID := range allUserIDs {
		// Get user to check if superadmin
		var user domain.UserModel
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			continue // Skip if user not found
		}
		if user.Role != "superadmin" && user.Username != "superadmin" {
			userIDsToDelete = append(userIDsToDelete, userID)
		}
	}

	// 3. Delete all user_company_assignments for these companies (by company_id)
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.UserCompanyAssignmentModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user company assignments: %w", err)
	}
	zapLog.Info("Deleted user company assignments by company_id", zap.Int("company_count", len(companyIDs)))

	// 4. Delete all remaining assignments for users that will be deleted (by user_id)
	// This handles edge cases where user might have assignments in other companies
	if len(userIDsToDelete) > 0 {
		if err := tx.Where("user_id IN ?", userIDsToDelete).Delete(&domain.UserCompanyAssignmentModel{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete user assignments by user_id: %w", err)
		}
		zapLog.Info("Deleted user assignments by user_id", zap.Int("user_count", len(userIDsToDelete)))
	}

	// 5. Delete users (hard delete for development reset)
	if len(userIDsToDelete) > 0 {
		if err := tx.Where("id IN ?", userIDsToDelete).Delete(&domain.UserModel{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete users: %w", err)
		}
		zapLog.Info("Deleted users", zap.Int("user_count", len(userIDsToDelete)))
	}

	// 6. Delete all companies (soft delete: set is_active = false)
	if err := tx.Model(&domain.CompanyModel{}).Where("id IN ?", companyIDs).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete companies: %w", err)
	}
	zapLog.Info("Deleted companies (soft delete)", zap.Int("company_count", len(companyIDs)))

	// 7. CRITICAL: Reset holding company level to 0 dan ensure parent_id is NULL
	// Ini penting untuk memastikan holding level tidak kacau setelah reset
	holding, err := uc.companyRepo.GetByCode("PDV")
	if err == nil && holding != nil {
		// Reset holding level to 0 dan pastikan parent_id is NULL
		if err := tx.Model(&domain.CompanyModel{}).
			Where("code = ?", "PDV").
			Updates(map[string]interface{}{
				"level":     0,
				"parent_id": nil,
				"is_active": true, // Ensure holding is active
			}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reset holding level: %w", err)
		}
		zapLog.Info("Reset holding company level to 0", zap.String("holding_id", holding.ID))
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	zapLog.Info("Successfully reset subsidiary data",
		zap.Int("companies_deleted", len(companyIDs)),
		zap.Int("users_deleted", len(userIDsToDelete)),
	)

	return nil
}

// CheckSeederDataExists checks if seeder data already exists
// It checks for the holding company with code "PDV" and its expected subsidiaries
func (uc *developmentUseCase) CheckSeederDataExists() (bool, error) {
	// Check if holding company exists
	holding, err := uc.companyRepo.GetByCode("PDV")
	if err != nil {
		// If holding doesn't exist, seeder data doesn't exist
		return false, nil
	}

	if holding == nil {
		return false, nil
	}

	// Check if at least one expected subsidiary exists (e.g., "ENU", "PTG", etc.)
	expectedCodes := []string{"ENU", "PTG", "PLB", "PRT", "PSH"}
	for _, code := range expectedCodes {
		company, err := uc.companyRepo.GetByCode(code)
		if err == nil && company != nil && company.IsActive {
			return true, nil
		}
	}

	return false, nil
}

// RunSubsidiarySeeder runs the subsidiary seeder
// Returns (alreadyExists, error)
// If alreadyExists is true, it means seeder data already exists and the operation was cancelled
func (uc *developmentUseCase) RunSubsidiarySeeder() (bool, error) {
	zapLog := logger.GetLogger()

	// Check if seeder data already exists
	exists, err := uc.CheckSeederDataExists()
	if err != nil {
		return false, fmt.Errorf("failed to check seeder data: %w", err)
	}

	if exists {
		zapLog.Warn("Seeder data already exists, skipping seeder execution")
		return true, nil
	}

	// Get admin role
	adminRole, err := uc.roleRepo.GetByName("admin")
	if err != nil {
		return false, fmt.Errorf("admin role not found: %w", err)
	}

	// 1. Create or Update Holding Company
	// CRITICAL: Pastikan holding selalu level 0 dan parent_id NULL
	var holdingID string
	existingHolding, _ := uc.companyRepo.GetByCode("PDV")
	if existingHolding != nil {
		holdingID = existingHolding.ID
		existingHolding.Name = "Pedeve Pertamina"
		existingHolding.ShortName = "Pedeve"
		existingHolding.Description = "Perusahaan Holding Induk Pedeve Pertamina"
		existingHolding.Status = "Aktif"
		existingHolding.ParentID = nil // CRITICAL: Must be NULL for holding
		existingHolding.Level = 0      // CRITICAL: Must be 0 for holding
		existingHolding.IsActive = true
		
		// Log before update untuk debugging
		zapLog.Info("Updating holding company",
			zap.String("holding_id", holdingID),
			zap.Int("old_level", existingHolding.Level),
			zap.Bool("old_parent_is_null", existingHolding.ParentID == nil),
		)
		
		if err := uc.companyRepo.Update(existingHolding); err != nil {
			return false, fmt.Errorf("failed to update holding: %w", err)
		}
		
		// Double check: Verify holding level after update
		updatedHolding, _ := uc.companyRepo.GetByID(holdingID)
		if updatedHolding != nil {
			if updatedHolding.Level != 0 {
				zapLog.Error("Holding level is not 0 after update!",
					zap.String("holding_id", holdingID),
					zap.Int("actual_level", updatedHolding.Level),
					zap.Int("expected_level", 0),
				)
				// Force fix: Update directly via repository
				updatedHolding.Level = 0
				updatedHolding.ParentID = nil
				if err := uc.companyRepo.Update(updatedHolding); err != nil {
					zapLog.Error("Failed to fix holding level", zap.Error(err))
				} else {
					zapLog.Info("Fixed holding level to 0")
				}
			}
		}
	} else {
		holdingID = uuid.GenerateUUID()
		holding := &domain.CompanyModel{
			ID:          holdingID,
			Name:        "Pedeve Pertamina",
			ShortName:   "Pedeve",
			Code:        "PDV",
			Description: "Perusahaan Holding Induk Pedeve Pertamina",
			Status:      "Aktif",
			ParentID:    nil,
			Level:       0,
			IsActive:    true,
		}
		if err := uc.companyRepo.Create(holding); err != nil {
			return false, fmt.Errorf("failed to create holding: %w", err)
		}
	}

	// Create admin user for holding
	adminRoleID := adminRole.ID
	existingHoldingUser, _ := uc.userRepo.GetByUsername("admin.pedeve")
	
	var holdingAdminID string
	if existingHoldingUser != nil {
		// User already exists - update CompanyID and RoleID, then create/update assignment
		holdingAdminID = existingHoldingUser.ID
		existingHoldingUser.CompanyID = &holdingID
		existingHoldingUser.RoleID = &adminRoleID
		existingHoldingUser.Role = "admin"
		if err := uc.userRepo.Update(existingHoldingUser); err != nil {
			zapLog.Warn("Failed to update holding admin user", zap.Error(err))
		}
	} else {
		// Create new user
		holdingAdminID = uuid.GenerateUUID()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		holdingAdmin := &domain.UserModel{
			ID:        holdingAdminID,
			Username:  "admin.pedeve",
			Email:     "admin.pedeve@pedeve.com",
			Password:  string(hashedPassword),
			Role:      "admin",
			RoleID:    &adminRoleID,
			CompanyID: &holdingID,
			IsActive:  true,
		}
		if err := uc.userRepo.Create(holdingAdmin); err != nil {
			zapLog.Warn("Failed to create holding admin user", zap.Error(err))
		}
	}
	
	// Create or update entry in junction table for user-company assignment
	if holdingAdminID != "" {
		existingAssignment, _ := uc.userCompanyAssignmentRepo.GetByUserAndCompany(holdingAdminID, holdingID)
		if existingAssignment != nil {
			// Assignment exists - update it
			existingAssignment.RoleID = &adminRoleID
			existingAssignment.IsActive = true
			if err := uc.userCompanyAssignmentRepo.Update(existingAssignment); err != nil {
				zapLog.Warn("Failed to update assignment for holding admin", zap.Error(err))
			}
		} else {
			// Create new assignment
			assignment := &domain.UserCompanyAssignmentModel{
				ID:        uuid.GenerateUUID(),
				UserID:    holdingAdminID,
				CompanyID: holdingID,
				RoleID:    &adminRoleID,
				IsActive:  true,
			}
			if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
				zapLog.Warn("Failed to create assignment for holding admin", zap.Error(err))
			}
		}
	}

	// 2. Create Level 1 Companies
	level1Companies := []struct {
		name        string
		code        string
		description string
		username    string
		email       string
	}{
		{"PT Energi Nusantara", "ENU", "Perusahaan energi dan migas", "admin.enu", "admin.enu@pedeve.com"},
		{"PT Pertamina Gas", "PTG", "Perusahaan gas dan LNG", "admin.ptg", "admin.ptg@pedeve.com"},
		{"PT Pertamina Lubricants", "PLB", "Perusahaan pelumas", "admin.plb", "admin.plb@pedeve.com"},
		{"PT Pertamina Retail", "PRT", "Perusahaan retail dan SPBU", "admin.prt", "admin.prt@pedeve.com"},
		{"PT Pertamina Shipping", "PSH", "Perusahaan shipping dan logistik", "admin.psh", "admin.psh@pedeve.com"},
	}

	level1IDs := make([]string, len(level1Companies))
	for i, comp := range level1Companies {
		existing, _ := uc.companyRepo.GetByCode(comp.code)
		if existing != nil {
			level1IDs[i] = existing.ID
			existing.ParentID = &holdingID
			existing.Level = 1
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := uc.companyRepo.Update(existing); err != nil {
				zapLog.Warn("Failed to update company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		} else {
			companyID := uuid.GenerateUUID()
			level1IDs[i] = companyID
			company := &domain.CompanyModel{
				ID:          companyID,
				Name:        comp.name,
				ShortName:   comp.name,
				Code:        comp.code,
				Description: comp.description,
				Status:      "Aktif",
				ParentID:    &holdingID,
				Level:       1,
				IsActive:    true,
			}
			if err := uc.companyRepo.Create(company); err != nil {
				zapLog.Warn("Failed to create company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		}

		// Create admin user
		companyIDToUse := level1IDs[i]
		adminRoleID := adminRole.ID
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := uc.userRepo.Update(existingUser); err != nil {
				zapLog.Warn("Failed to update user", zap.String("username", comp.username), zap.Error(err))
			}
		} else {
			// Create new user
			userID = uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			user := &domain.UserModel{
				ID:        userID,
				Username:  comp.username,
				Email:     comp.email,
				Password:  string(hashedPassword),
				Role:      "admin",
				RoleID:    &adminRoleID,
				CompanyID: &companyIDToUse,
				IsActive:  true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
				continue // Skip assignment if user creation failed
			}
		}
		
		// Create or update entry in junction table for user-company assignment
		if userID != "" {
			existingAssignment, _ := uc.userCompanyAssignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
			if existingAssignment != nil {
				// Assignment exists - update it
				existingAssignment.RoleID = &adminRoleID
				existingAssignment.IsActive = true
				if err := uc.userCompanyAssignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to update assignment", zap.String("username", comp.username), zap.Error(err))
				}
			} else {
				// Create new assignment
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        uuid.GenerateUUID(),
					UserID:    userID,
					CompanyID: companyIDToUse,
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create assignment", zap.String("username", comp.username), zap.Error(err))
				}
			}
		}
	}

	// 3. Create Level 2 Companies
	level2Companies := []struct {
		name        string
		code        string
		description string
		parentIndex  int
		username    string
		email       string
	}{
		{"PT ENU Exploration", "ENU-EXP", "Eksplorasi minyak dan gas", 0, "admin.enu.exp", "admin.enu.exp@pedeve.com"},
		{"PT ENU Production", "ENU-PRO", "Produksi minyak dan gas", 0, "admin.enu.pro", "admin.enu.pro@pedeve.com"},
		{"PT PTG Distribution", "PTG-DIST", "Distribusi gas", 1, "admin.ptg.dist", "admin.ptg.dist@pedeve.com"},
	}

	level2IDs := make([]string, len(level2Companies))
	for i, comp := range level2Companies {
		parentID := level1IDs[comp.parentIndex]
		existing, _ := uc.companyRepo.GetByCode(comp.code)
		if existing != nil {
			level2IDs[i] = existing.ID
			existing.ParentID = &parentID
			existing.Level = 2
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := uc.companyRepo.Update(existing); err != nil {
				zapLog.Warn("Failed to update company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		} else {
			companyID := uuid.GenerateUUID()
			level2IDs[i] = companyID
			company := &domain.CompanyModel{
				ID:          companyID,
				Name:        comp.name,
				ShortName:   comp.name,
				Code:        comp.code,
				Description: comp.description,
				Status:      "Aktif",
				ParentID:    &parentID,
				Level:       2,
				IsActive:    true,
			}
			if err := uc.companyRepo.Create(company); err != nil {
				zapLog.Warn("Failed to create company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		}

		// Create admin user
		companyIDToUse := level2IDs[i]
		adminRoleID := adminRole.ID
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := uc.userRepo.Update(existingUser); err != nil {
				zapLog.Warn("Failed to update user", zap.String("username", comp.username), zap.Error(err))
			}
		} else {
			// Create new user
			userID = uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			user := &domain.UserModel{
				ID:        userID,
				Username:  comp.username,
				Email:     comp.email,
				Password:  string(hashedPassword),
				Role:      "admin",
				RoleID:    &adminRoleID,
				CompanyID: &companyIDToUse,
				IsActive:  true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
				continue // Skip assignment if user creation failed
			}
		}
		
		// Create or update entry in junction table for user-company assignment
		if userID != "" {
			existingAssignment, _ := uc.userCompanyAssignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
			if existingAssignment != nil {
				// Assignment exists - update it
				existingAssignment.RoleID = &adminRoleID
				existingAssignment.IsActive = true
				if err := uc.userCompanyAssignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to update assignment", zap.String("username", comp.username), zap.Error(err))
				}
			} else {
				// Create new assignment
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        uuid.GenerateUUID(),
					UserID:    userID,
					CompanyID: companyIDToUse,
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create assignment", zap.String("username", comp.username), zap.Error(err))
				}
			}
		}
	}

	// 4. Create Level 3 Companies
	level3Companies := []struct {
		name        string
		code        string
		description string
		parentIndex  int
		username    string
		email       string
	}{
		{"PT ENU-EXP Drilling", "ENU-EXP-DRL", "Layanan pengeboran", 0, "admin.enu.exp.drl", "admin.enu.exp.drl@pedeve.com"},
		{"PT ENU-PRO Refinery", "ENU-PRO-REF", "Kilang minyak", 1, "admin.enu.pro.ref", "admin.enu.pro.ref@pedeve.com"},
	}

	for _, comp := range level3Companies {
		parentID := level2IDs[comp.parentIndex]
		var companyID string
		
		existing, _ := uc.companyRepo.GetByCode(comp.code)
		if existing != nil {
			companyID = existing.ID
			existing.ParentID = &parentID
			existing.Level = 3
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := uc.companyRepo.Update(existing); err != nil {
				zapLog.Warn("Failed to update company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		} else {
			companyID = uuid.GenerateUUID()
			company := &domain.CompanyModel{
				ID:          companyID,
				Name:        comp.name,
				ShortName:   comp.name,
				Code:        comp.code,
				Description: comp.description,
				Status:      "Aktif",
				ParentID:    &parentID,
				Level:       3,
				IsActive:    true,
			}
			if err := uc.companyRepo.Create(company); err != nil {
				zapLog.Warn("Failed to create company", zap.String("code", comp.code), zap.Error(err))
				continue
			}
		}

		// Create admin user
		companyIDToUse := companyID
		adminRoleID := adminRole.ID
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := uc.userRepo.Update(existingUser); err != nil {
				zapLog.Warn("Failed to update user", zap.String("username", comp.username), zap.Error(err))
			}
		} else {
			// Create new user
			userID = uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			user := &domain.UserModel{
				ID:        userID,
				Username:  comp.username,
				Email:     comp.email,
				Password:  string(hashedPassword),
				Role:      "admin",
				RoleID:    &adminRoleID,
				CompanyID: &companyIDToUse,
				IsActive:  true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
				continue // Skip assignment if user creation failed
			}
		}
		
		// Create or update entry in junction table for user-company assignment
		if userID != "" {
			existingAssignment, _ := uc.userCompanyAssignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
			if existingAssignment != nil {
				// Assignment exists - update it
				existingAssignment.RoleID = &adminRoleID
				existingAssignment.IsActive = true
				if err := uc.userCompanyAssignmentRepo.Update(existingAssignment); err != nil {
					zapLog.Warn("Failed to update assignment", zap.String("username", comp.username), zap.Error(err))
				}
			} else {
				// Create new assignment
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        uuid.GenerateUUID(),
					UserID:    userID,
					CompanyID: companyIDToUse,
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create assignment", zap.String("username", comp.username), zap.Error(err))
				}
			}
		}
	}

	zapLog.Info("Subsidiary seeder completed successfully")
	return false, nil
}


