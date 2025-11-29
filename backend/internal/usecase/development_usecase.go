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

// NewDevelopmentUseCase creates a new development use case
func NewDevelopmentUseCase() DevelopmentUseCase {
	return &developmentUseCase{
		companyRepo:              repository.NewCompanyRepository(),
		userRepo:                 repository.NewUserRepository(),
		roleRepo:                 repository.NewRoleRepository(),
		userCompanyAssignmentRepo: repository.NewUserCompanyAssignmentRepository(),
	}
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

	// 1. Get all companies except root holding (parent_id IS NULL)
	var allCompanies []domain.CompanyModel
	if err := tx.Where("parent_id IS NOT NULL").Find(&allCompanies).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get companies: %w", err)
	}

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

	// 2. Delete all user_company_assignments for these companies
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.UserCompanyAssignmentModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user company assignments: %w", err)
	}
	zapLog.Info("Deleted user company assignments", zap.Int("company_count", len(companyIDs)))

	// 3. Get all users assigned to these companies (from junction table or CompanyID)
	var usersToDelete []domain.UserModel
	if err := tx.Where("company_id IN ?", companyIDs).Find(&usersToDelete).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get users: %w", err)
	}

	// Filter out superadmin users
	userIDsToDelete := make([]string, 0)
	for _, user := range usersToDelete {
		if user.Role != "superadmin" && user.Username != "superadmin" {
			userIDsToDelete = append(userIDsToDelete, user.ID)
		}
	}

	// 4. Delete users (hard delete for development reset)
	if len(userIDsToDelete) > 0 {
		if err := tx.Where("id IN ?", userIDsToDelete).Delete(&domain.UserModel{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete users: %w", err)
		}
		zapLog.Info("Deleted users", zap.Int("user_count", len(userIDsToDelete)))
	}

	// 5. Delete all companies (soft delete: set is_active = false)
	if err := tx.Model(&domain.CompanyModel{}).Where("id IN ?", companyIDs).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete companies: %w", err)
	}
	zapLog.Info("Deleted companies (soft delete)", zap.Int("company_count", len(companyIDs)))

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
	var holdingID string
	existingHolding, _ := uc.companyRepo.GetByCode("PDV")
	if existingHolding != nil {
		holdingID = existingHolding.ID
		existingHolding.Name = "Pedeve Pertamina"
		existingHolding.ShortName = "Pedeve"
		existingHolding.Description = "Perusahaan Holding Induk Pedeve Pertamina"
		existingHolding.Status = "Aktif"
		existingHolding.ParentID = nil
		existingHolding.Level = 0
		existingHolding.IsActive = true
		if err := uc.companyRepo.Update(existingHolding); err != nil {
			return false, fmt.Errorf("failed to update holding: %w", err)
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
	existingHoldingUser, _ := uc.userRepo.GetByUsername("admin.pedeve")
	if existingHoldingUser == nil {
		holdingAdminID := uuid.GenerateUUID()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminRoleID := adminRole.ID
		holdingAdmin := &domain.UserModel{
			ID:       holdingAdminID,
			Username: "admin.pedeve",
			Email:    "admin.pedeve@pedeve.com",
			Password: string(hashedPassword),
			Role:     "admin",
			RoleID:   &adminRoleID,
			IsActive: true,
		}
		if err := uc.userRepo.Create(holdingAdmin); err != nil {
			zapLog.Warn("Failed to create holding admin user", zap.Error(err))
		} else {
			// Create assignment in junction table
			assignmentID := uuid.GenerateUUID()
			assignment := &domain.UserCompanyAssignmentModel{
				ID:        assignmentID,
				UserID:    holdingAdminID,
				CompanyID: holdingID,
				RoleID:    &adminRoleID,
				IsActive:  true,
			}
			if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
				zapLog.Warn("Failed to create user company assignment", zap.Error(err))
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
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		if existingUser == nil {
			userID := uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			adminRoleID := adminRole.ID
			user := &domain.UserModel{
				ID:       userID,
				Username: comp.username,
				Email:    comp.email,
				Password: string(hashedPassword),
				Role:     "admin",
				RoleID:   &adminRoleID,
				IsActive: true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
			} else {
				// Create assignment in junction table
				assignmentID := uuid.GenerateUUID()
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        assignmentID,
					UserID:    userID,
					CompanyID: level1IDs[i],
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create user company assignment", zap.Error(err))
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
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		if existingUser == nil {
			userID := uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			adminRoleID := adminRole.ID
			user := &domain.UserModel{
				ID:       userID,
				Username: comp.username,
				Email:    comp.email,
				Password: string(hashedPassword),
				Role:     "admin",
				RoleID:   &adminRoleID,
				IsActive: true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
			} else {
				// Create assignment in junction table
				assignmentID := uuid.GenerateUUID()
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        assignmentID,
					UserID:    userID,
					CompanyID: level2IDs[i],
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create user company assignment", zap.Error(err))
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
		existingUser, _ := uc.userRepo.GetByUsername(comp.username)
		if existingUser == nil {
			userID := uuid.GenerateUUID()
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			adminRoleID := adminRole.ID
			user := &domain.UserModel{
				ID:       userID,
				Username: comp.username,
				Email:    comp.email,
				Password: string(hashedPassword),
				Role:     "admin",
				RoleID:   &adminRoleID,
				IsActive: true,
			}
			if err := uc.userRepo.Create(user); err != nil {
				zapLog.Warn("Failed to create user", zap.String("username", comp.username), zap.Error(err))
			} else {
				// Create assignment in junction table
				assignmentID := uuid.GenerateUUID()
				assignment := &domain.UserCompanyAssignmentModel{
					ID:        assignmentID,
					UserID:    userID,
					CompanyID: companyID,
					RoleID:    &adminRoleID,
					IsActive:  true,
				}
				if err := uc.userCompanyAssignmentRepo.Create(assignment); err != nil {
					zapLog.Warn("Failed to create user company assignment", zap.Error(err))
				}
			}
		}
	}

	zapLog.Info("Subsidiary seeder completed successfully")
	return false, nil
}

