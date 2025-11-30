package main

import (
	"fmt"
	"os"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Set DATABASE_URL if not set
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable")
	}

	fmt.Println("🌱 Seeding Companies and Users")
	fmt.Println()

	// Init logger
	logger.InitLogger()
	defer logger.Sync()

	// Init database
	// Note: GORM logs will appear during AutoMigrate, but we'll silence them after
	database.InitDB()
	
	// Reduce GORM logging for seeder (set to Silent to avoid verbose output)
	// This must be done AFTER InitDB() because InitDB() sets the logger
	db := database.GetDB()
	if db != nil {
		db.Logger = db.Logger.LogMode(gormLogger.Silent)
	}
	
	fmt.Println("✅ Database initialized (GORM logging disabled for cleaner output)")
	fmt.Println()

	// Initialize repositories
	companyRepo := repository.NewCompanyRepository()
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	assignmentRepo := repository.NewUserCompanyAssignmentRepository()

	fmt.Println("✅ Connected to database")
	fmt.Println()

	// Get or create admin role
	adminRole, err := roleRepo.GetByName("admin")
	if err != nil {
		fmt.Println("❌ Admin role not found. Please create admin role first.")
		return
	}
	fmt.Printf("✅ Found admin role: %s (ID: %s)\n", adminRole.Name, adminRole.ID)
	fmt.Println()

	// 1. Create or Update Holding Company
	fmt.Println("1️⃣  Creating/Updating Holding Company...")
	var holdingID string
	
	// Check if holding already exists by code
	existingHoldingByCode, _ := companyRepo.GetByCode("PDV")
	if existingHoldingByCode != nil {
		// Update existing holding
		holdingID = existingHoldingByCode.ID
		fmt.Printf("   ⚠️  Holding company with code PDV already exists: %s (updating)\n", existingHoldingByCode.Name)
		existingHoldingByCode.Name = "Pedeve Pertamina"
		existingHoldingByCode.ShortName = "Pedeve"
		existingHoldingByCode.Description = "Perusahaan Holding Induk Pedeve Pertamina"
		existingHoldingByCode.Status = "Aktif"
		existingHoldingByCode.ParentID = nil
		existingHoldingByCode.Level = 0
		existingHoldingByCode.IsActive = true
		if err := companyRepo.Update(existingHoldingByCode); err != nil {
			fmt.Printf("   ❌ Failed to update holding: %v\n", err)
			return
		}
		fmt.Printf("   ✅ Updated: %s (ID: %s)\n", existingHoldingByCode.Name, holdingID)
	} else {
		// Check if any root holding exists
		existingHolding, _ := companyRepo.GetRootHolding()
		if existingHolding != nil {
			fmt.Printf("   ⚠️  Root holding company already exists: %s (code: %s)\n", existingHolding.Name, existingHolding.Code)
			fmt.Println("   Using existing holding as root...")
			holdingID = existingHolding.ID
		} else {
			// Create new holding
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
			if err := companyRepo.Create(holding); err != nil {
				fmt.Printf("   ❌ Failed to create holding: %v\n", err)
				return
			}
			fmt.Printf("   ✅ Created: %s (ID: %s)\n", holding.Name, holdingID)
		}
	}

	// Create admin user for holding (check if exists first)
	adminRoleID := adminRole.ID
	existingHoldingUser, _ := userRepo.GetByUsername("admin.pedeve")
	
	var holdingAdminID string
	if existingHoldingUser != nil {
		// User already exists - update CompanyID and RoleID, then create/update assignment
		holdingAdminID = existingHoldingUser.ID
		existingHoldingUser.CompanyID = &holdingID
		existingHoldingUser.RoleID = &adminRoleID
		existingHoldingUser.Role = "admin"
		if err := userRepo.Update(existingHoldingUser); err != nil {
			fmt.Printf("   ⚠️  Failed to update holding admin user: %v\n", err)
		} else {
			fmt.Printf("   ⚠️  User admin.pedeve already exists (updated company assignment)\n")
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
		if err := userRepo.Create(holdingAdmin); err != nil {
			fmt.Printf("   ⚠️  Failed to create holding admin user: %v\n", err)
		} else {
			fmt.Printf("   ✅ Created user: %s (password: admin123)\n", holdingAdmin.Username)
		}
	}
	
	// Create or update entry in junction table for user-company assignment
	existingAssignment, _ := assignmentRepo.GetByUserAndCompany(holdingAdminID, holdingID)
	if existingAssignment != nil {
		// Assignment exists - update it
		existingAssignment.RoleID = &adminRoleID
		existingAssignment.IsActive = true
		if err := assignmentRepo.Update(existingAssignment); err != nil {
			fmt.Printf("      ⚠️  Failed to update assignment for user admin.pedeve: %v\n", err)
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
		if err := assignmentRepo.Create(assignment); err != nil {
			fmt.Printf("      ⚠️  Failed to create assignment for user admin.pedeve: %v\n", err)
		}
	}
	fmt.Println()

	// 2. Create Level 1 Companies (direct children of holding) - 5 companies
	fmt.Println("2️⃣  Creating Level 1 Companies (5 companies)...")
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
		companyID := uuid.GenerateUUID()
		level1IDs[i] = companyID
		
		// Check if company with same code already exists (even if inactive)
		existing, _ := companyRepo.GetByCode(comp.code)
		if existing != nil {
			// Use existing company ID
			level1IDs[i] = existing.ID
			fmt.Printf("   ⚠️  Company with code %s already exists: %s (reusing)\n", comp.code, existing.Name)
			// Update existing company to be active and set correct parent
			existing.ParentID = &holdingID
			existing.Level = 1
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := companyRepo.Update(existing); err != nil {
				fmt.Printf("   ❌ Failed to update %s: %v\n", comp.name, err)
				continue
			}
			fmt.Printf("   ✅ Updated: %s (ID: %s)\n", comp.name, existing.ID)
		} else {
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
			if err := companyRepo.Create(company); err != nil {
				fmt.Printf("   ❌ Failed to create %s: %v\n", comp.name, err)
				continue
			}
			fmt.Printf("   ✅ Created: %s (ID: %s)\n", comp.name, companyID)
		}

		// Create admin user (check if exists first)
		companyIDToUse := level1IDs[i]
		adminRoleID := adminRole.ID
		existingUser, _ := userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := userRepo.Update(existingUser); err != nil {
				fmt.Printf("      ⚠️  Failed to update user %s: %v\n", comp.username, err)
			} else {
				fmt.Printf("      ⚠️  User %s already exists (updated company assignment)\n", comp.username)
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
			if err := userRepo.Create(user); err != nil {
				fmt.Printf("   ⚠️  Failed to create user %s: %v\n", comp.username, err)
				continue // Skip assignment if user creation failed
			}
			fmt.Printf("      ✅ User: %s (password: admin123)\n", comp.username)
		}
		
		// Create or update entry in junction table for user-company assignment
		existingAssignment, _ := assignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
		if existingAssignment != nil {
			// Assignment exists - update it
			existingAssignment.RoleID = &adminRoleID
			existingAssignment.IsActive = true
			if err := assignmentRepo.Update(existingAssignment); err != nil {
				fmt.Printf("      ⚠️  Failed to update assignment for user %s: %v\n", comp.username, err)
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
			if err := assignmentRepo.Create(assignment); err != nil {
				fmt.Printf("      ⚠️  Failed to create assignment for user %s: %v\n", comp.username, err)
			}
		}
	}
	fmt.Println()

	// 3. Create Level 2 Companies (children of level 1) - 3 companies
	fmt.Println("3️⃣  Creating Level 2 Companies (3 companies)...")
	level2Companies := []struct {
		name        string
		code        string
		description string
		parentIndex  int // Index in level1Companies
		username    string
		email       string
	}{
		{"PT ENU Exploration", "ENU-EXP", "Eksplorasi minyak dan gas", 0, "admin.enu.exp", "admin.enu.exp@pedeve.com"},
		{"PT ENU Production", "ENU-PRO", "Produksi minyak dan gas", 0, "admin.enu.pro", "admin.enu.pro@pedeve.com"},
		{"PT PTG Distribution", "PTG-DIST", "Distribusi gas", 1, "admin.ptg.dist", "admin.ptg.dist@pedeve.com"},
	}

	level2IDs := make([]string, len(level2Companies))
	for i, comp := range level2Companies {
		companyID := uuid.GenerateUUID()
		parentID := level1IDs[comp.parentIndex]
		
		// Check if company with same code already exists
		existing, _ := companyRepo.GetByCode(comp.code)
		if existing != nil {
			level2IDs[i] = existing.ID
			fmt.Printf("   ⚠️  Company with code %s already exists: %s (reusing)\n", comp.code, existing.Name)
			existing.ParentID = &parentID
			existing.Level = 2
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := companyRepo.Update(existing); err != nil {
				fmt.Printf("   ❌ Failed to update %s: %v\n", comp.name, err)
				continue
			}
			fmt.Printf("   ✅ Updated: %s (ID: %s, Parent: %s)\n", comp.name, existing.ID, parentID[:8]+"...")
		} else {
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
			if err := companyRepo.Create(company); err != nil {
				fmt.Printf("   ❌ Failed to create %s: %v\n", comp.name, err)
				continue
			}
			fmt.Printf("   ✅ Created: %s (ID: %s, Parent: %s)\n", comp.name, companyID, parentID[:8]+"...")
		}

		// Create admin user (check if exists first)
		companyIDToUse := level2IDs[i]
		adminRoleID := adminRole.ID
		existingUser, _ := userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := userRepo.Update(existingUser); err != nil {
				fmt.Printf("      ⚠️  Failed to update user %s: %v\n", comp.username, err)
			} else {
				fmt.Printf("      ⚠️  User %s already exists (updated company assignment)\n", comp.username)
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
			if err := userRepo.Create(user); err != nil {
				fmt.Printf("   ⚠️  Failed to create user %s: %v\n", comp.username, err)
				continue // Skip assignment if user creation failed
			}
			fmt.Printf("      ✅ User: %s (password: admin123)\n", comp.username)
		}
		
		// Create or update entry in junction table for user-company assignment
		existingAssignment, _ := assignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
		if existingAssignment != nil {
			// Assignment exists - update it
			existingAssignment.RoleID = &adminRoleID
			existingAssignment.IsActive = true
			if err := assignmentRepo.Update(existingAssignment); err != nil {
				fmt.Printf("      ⚠️  Failed to update assignment for user %s: %v\n", comp.username, err)
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
			if err := assignmentRepo.Create(assignment); err != nil {
				fmt.Printf("      ⚠️  Failed to create assignment for user %s: %v\n", comp.username, err)
			}
		}
	}
	fmt.Println()

	// 4. Create Level 3 Companies (children of level 2) - 2 companies
	fmt.Println("4️⃣  Creating Level 3 Companies (2 companies)...")
	level3Companies := []struct {
		name        string
		code        string
		description string
		parentIndex  int // Index in level2Companies
		username    string
		email       string
	}{
		{"PT ENU-EXP Drilling", "ENU-EXP-DRL", "Layanan pengeboran", 0, "admin.enu.exp.drl", "admin.enu.exp.drl@pedeve.com"},
		{"PT ENU-PRO Refinery", "ENU-PRO-REF", "Kilang minyak", 1, "admin.enu.pro.ref", "admin.enu.pro.ref@pedeve.com"},
	}

	for _, comp := range level3Companies {
		companyID := uuid.GenerateUUID()
		parentID := level2IDs[comp.parentIndex]
		
		// Check if company with same code already exists
		existing, _ := companyRepo.GetByCode(comp.code)
		if existing != nil {
			fmt.Printf("   ⚠️  Company with code %s already exists: %s (reusing)\n", comp.code, existing.Name)
			existing.ParentID = &parentID
			existing.Level = 3
			existing.IsActive = true
			existing.Name = comp.name
			existing.ShortName = comp.name
			existing.Description = comp.description
			existing.Status = "Aktif"
			if err := companyRepo.Update(existing); err != nil {
				fmt.Printf("   ❌ Failed to update %s: %v\n", comp.name, err)
				continue
			}
			companyID = existing.ID
			fmt.Printf("   ✅ Updated: %s (ID: %s, Parent: %s)\n", comp.name, companyID, parentID[:8]+"...")
		} else {
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
			if err := companyRepo.Create(company); err != nil {
				fmt.Printf("   ❌ Failed to create %s: %v\n", comp.name, err)
				continue
			}
			fmt.Printf("   ✅ Created: %s (ID: %s, Parent: %s)\n", comp.name, companyID, parentID[:8]+"...")
		}

		// Create admin user (check if exists first)
		companyIDToUse := companyID
		adminRoleID := adminRole.ID
		existingUser, _ := userRepo.GetByUsername(comp.username)
		
		var userID string
		if existingUser != nil {
			// User already exists - update CompanyID and RoleID, then create/update assignment
			userID = existingUser.ID
			existingUser.CompanyID = &companyIDToUse
			existingUser.RoleID = &adminRoleID
			existingUser.Role = "admin"
			if err := userRepo.Update(existingUser); err != nil {
				fmt.Printf("      ⚠️  Failed to update user %s: %v\n", comp.username, err)
			} else {
				fmt.Printf("      ⚠️  User %s already exists (updated company assignment)\n", comp.username)
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
			if err := userRepo.Create(user); err != nil {
				fmt.Printf("   ⚠️  Failed to create user %s: %v\n", comp.username, err)
				continue // Skip assignment if user creation failed
			}
			fmt.Printf("      ✅ User: %s (password: admin123)\n", comp.username)
		}
		
		// Create or update entry in junction table for user-company assignment
		existingAssignment, _ := assignmentRepo.GetByUserAndCompany(userID, companyIDToUse)
		if existingAssignment != nil {
			// Assignment exists - update it
			existingAssignment.RoleID = &adminRoleID
			existingAssignment.IsActive = true
			if err := assignmentRepo.Update(existingAssignment); err != nil {
				fmt.Printf("      ⚠️  Failed to update assignment for user %s: %v\n", comp.username, err)
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
			if err := assignmentRepo.Create(assignment); err != nil {
				fmt.Printf("      ⚠️  Failed to create assignment for user %s: %v\n", comp.username, err)
			}
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Println("   ✅ 1 Holding Company (Pedeve Pertamina)")
	fmt.Println("   ✅ 5 Level 1 Companies (Anak Perusahaan)")
	fmt.Println("   ✅ 3 Level 2 Companies (Cucu Perusahaan)")
	fmt.Println("   ✅ 2 Level 3 Companies (Cicit Perusahaan)")
	fmt.Println("   ✅ Total: 11 Companies (1 holding + 10 subsidiaries)")
	fmt.Println("   ✅ Total: 11 Admin Users (1 for holding + 10 for subsidiaries)")
	fmt.Println()
	fmt.Println("🔑 Default Password for all users: admin123")
	fmt.Println()
	fmt.Println("📋 Company Hierarchy:")
	fmt.Println("   Pedeve Pertamina (Holding)")
	fmt.Println("   ├── PT Energi Nusantara")
	fmt.Println("   │   ├── PT ENU Exploration")
	fmt.Println("   │   │   └── PT ENU-EXP Drilling")
	fmt.Println("   │   └── PT ENU Production")
	fmt.Println("   │       └── PT ENU-PRO Refinery")
	fmt.Println("   ├── PT Pertamina Gas")
	fmt.Println("   │   └── PT PTG Distribution")
	fmt.Println("   ├── PT Pertamina Lubricants")
	fmt.Println("   ├── PT Pertamina Retail")
	fmt.Println("   └── PT Pertamina Shipping")
	fmt.Println()
	fmt.Println("🎉 Seeding completed successfully!")
}

