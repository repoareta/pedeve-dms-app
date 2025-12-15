package main

import (
	"fmt"
	"os"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"golang.org/x/crypto/bcrypt"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Write to stderr IMMEDIATELY at the very start
	fmt.Fprintf(os.Stderr, "STEP: Seeder main() started\n")
	os.Stderr.Sync()

	// Ensure output is flushed immediately
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "PANIC in seeder: %v\n", r)
			os.Stderr.Sync()
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "STEP: Seeder main() completed successfully\n")
		os.Stderr.Sync()
	}()

	// Set DATABASE_URL if not set
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable")
	}

	// Write to BOTH stdout and stderr IMMEDIATELY to ensure it's captured
	fmt.Fprintf(os.Stderr, "STEP: About to print seeding message\n")
	os.Stderr.Sync()
	fmt.Println("🌱 Seeding Companies and Users")
	fmt.Println()
	os.Stdout.Sync() // Force flush output
	fmt.Fprintf(os.Stderr, "STEP: Seeding message printed, about to start logger init\n")
	os.Stderr.Sync()

	// Write to stderr IMMEDIATELY to ensure it's captured
	fmt.Fprintf(os.Stderr, "STEP: Starting seeder - before logger init\n")
	os.Stderr.Sync()

	// Init logger
	fmt.Fprintf(os.Stderr, "STEP: Before logger init\n")
	os.Stderr.Sync()

	// Wrap logger init in recover to catch any panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "PANIC in logger.InitLogger(): %v\n", r)
				os.Stderr.Sync()
				panic(r) // Re-panic to be caught by outer defer
			}
		}()
		logger.InitLogger()
	}()

	defer logger.Sync()
	fmt.Fprintf(os.Stderr, "STEP: After logger init\n")
	os.Stderr.Sync()

	// Init database
	// Note: GORM logs will appear during AutoMigrate, but we'll silence them after
	fmt.Fprintf(os.Stderr, "STEP: Initializing database...\n")
	os.Stderr.Sync()

	// Wrap InitDB in a recover to catch any panics
	fmt.Fprintf(os.Stderr, "STEP: About to call database.InitDB()\n")
	os.Stderr.Sync()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "PANIC in database.InitDB(): %v\n", r)
				os.Stderr.Sync()
				panic(r) // Re-panic to be caught by outer defer
			}
		}()
		database.InitDB()
	}()

	fmt.Fprintf(os.Stderr, "STEP: Database initialized - InitDB() completed\n")
	os.Stderr.Sync()
	fmt.Fprintf(os.Stdout, "STEP: Database initialized - InitDB() completed\n")
	os.Stdout.Sync()

	// Reduce GORM logging for seeder (set to Silent to avoid verbose output)
	// This must be done AFTER InitDB() because InitDB() sets the logger
	fmt.Fprintf(os.Stderr, "STEP: Getting DB instance and setting logger to Silent...\n")
	os.Stderr.Sync()
	db := database.GetDB()
	if db != nil {
		db.Logger = db.Logger.LogMode(gormLogger.Silent)
		fmt.Fprintf(os.Stderr, "STEP: GORM logger set to Silent\n")
		os.Stderr.Sync()
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: database.GetDB() returned nil\n")
		os.Stderr.Sync()
		os.Exit(1)
	}

	fmt.Println("✅ Database initialized (GORM logging disabled for cleaner output)")
	fmt.Println()
	os.Stdout.Sync() // Force flush output
	fmt.Fprintf(os.Stderr, "STEP: After database init message printed\n")
	os.Stderr.Sync()

	// Initialize repositories
	fmt.Fprintf(os.Stderr, "STEP: Initializing repositories...\n")
	os.Stderr.Sync()
	companyRepo := repository.NewCompanyRepository()
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	assignmentRepo := repository.NewUserCompanyAssignmentRepository()

	// Initialize usecase for company operations (to use proper business logic)
	companyUseCase := usecase.NewCompanyUseCase()
	fmt.Fprintf(os.Stderr, "STEP: Repositories initialized\n")
	os.Stderr.Sync()

	fmt.Println("✅ Connected to database")
	fmt.Println()
	os.Stdout.Sync() // Force flush output
	fmt.Fprintf(os.Stderr, "STEP: After connected message\n")
	os.Stderr.Sync()

	// Get or create admin role
	fmt.Fprintf(os.Stderr, "Checking for admin role...\n")
	os.Stderr.Sync()
	adminRole, err := roleRepo.GetByName("admin")
	if err != nil {
		fmt.Println("⚠️  Admin role not found. Attempting to create admin role...")
		os.Stdout.Sync()

		// Try to create admin role if it doesn't exist
		adminRoleID := uuid.GenerateUUID()
		adminRole = &domain.RoleModel{
			ID:          adminRoleID,
			Name:        "admin",
			Description: "Administrator - Company-level admin access",
			Level:       1,
			IsSystem:    true,
		}
		if err := roleRepo.Create(adminRole); err != nil {
			fmt.Printf("   ❌ Failed to create admin role: %v\n", err)
			fmt.Println("   ⚠️  Will continue without creating users (companies will still be created)")
			adminRole = nil
		} else {
			fmt.Printf("   ✅ Created admin role: %s (ID: %s)\n", adminRole.Name, adminRole.ID)
		}
	} else {
		fmt.Printf("✅ Found admin role: %s (ID: %s)\n", adminRole.Name, adminRole.ID)
		os.Stdout.Sync()
	}
	fmt.Println()
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "STEP: After admin role check\n")
	os.Stderr.Sync()

	// Get or create administrator role
	var administratorRole *domain.RoleModel
	administratorRole, err = roleRepo.GetByName("administrator")
	if err != nil {
		fmt.Println("⚠️  Administrator role not found. Please create administrator role first.")
		fmt.Println("   Skipping administrator user creation...")
		administratorRole = nil
	} else {
		fmt.Printf("✅ Found administrator role: %s (ID: %s)\n", administratorRole.Name, administratorRole.ID)
		fmt.Println()
	}

	// 1. Create or Update Holding Company
	fmt.Println("1️⃣  Creating/Updating Holding Company...")
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "STEP: Starting holding company creation...\n")
	os.Stderr.Sync()
	var holdingID string

	// Check if holding already exists by code
	existingHoldingByCode, _ := companyRepo.GetByCode("PDV")
	if existingHoldingByCode != nil {
		// Update existing holding using usecase to ensure proper logic
		holdingID = existingHoldingByCode.ID
		fmt.Printf("   ⚠️  Holding company with code PDV already exists: %s (updating)\n", existingHoldingByCode.Name)

		// If company is inactive, activate it first before updating
		if !existingHoldingByCode.IsActive {
			fmt.Printf("   ⚠️  Holding company is inactive, activating it first...\n")
			os.Stdout.Sync()
			fmt.Fprintf(os.Stderr, "STEP: Activating inactive holding company...\n")
			os.Stderr.Sync()

			// Directly update is_active in database to bypass validation
			db := database.GetDB()
			if err := db.Model(&domain.CompanyModel{}).Where("id = ?", holdingID).Update("is_active", true).Error; err != nil {
				fmt.Printf("   ❌ Failed to activate holding: %v\n", err)
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "ERROR: Failed to activate holding: %v\n", err)
				os.Stderr.Sync()
				return
			}
			fmt.Printf("   ✅ Activated holding company\n")
			os.Stdout.Sync()

			// Reload company to get updated data (not used but kept for potential future use)
			_, _ = companyRepo.GetByID(holdingID)
		}

		// Use UpdateCompanyFull to ensure proper level calculation
		fmt.Fprintf(os.Stderr, "STEP: Updating existing holding...\n")
		os.Stderr.Sync()
		updateData := &domain.CompanyUpdateRequest{
			Name:        "Pedeve Pertamina",
			ShortName:   "Pedeve",
			Description: "Perusahaan Holding Induk Pedeve Pertamina",
			Status:      "Aktif",
			ParentID:    nil, // Holding must have parent_id = nil
		}
		updated, err := companyUseCase.UpdateCompanyFull(holdingID, updateData)
		if err != nil {
			fmt.Printf("   ❌ Failed to update holding: %v\n", err)
			os.Stdout.Sync()
			fmt.Fprintf(os.Stderr, "ERROR: Failed to update holding: %v\n", err)
			os.Stderr.Sync()
			return
		}
		holdingID = updated.ID
		fmt.Printf("   ✅ Updated: %s (ID: %s, Level: %d)\n", updated.Name, holdingID, updated.Level)
	} else {
		// Check if any root holding exists
		fmt.Fprintf(os.Stderr, "STEP: Checking for root holding...\n")
		os.Stderr.Sync()
		existingHolding, _ := companyRepo.GetRootHolding()
		if existingHolding != nil {
			fmt.Printf("   ⚠️  Root holding company already exists: %s (code: %s)\n", existingHolding.Name, existingHolding.Code)
			fmt.Println("   Using existing holding as root...")
			os.Stdout.Sync()
			holdingID = existingHolding.ID
		} else {
			// Create new holding using CreateCompanyFull to ensure proper logic
			fmt.Fprintf(os.Stderr, "STEP: Creating new holding...\n")
			os.Stderr.Sync()
			createData := &domain.CompanyCreateRequest{
				Name:        "Pedeve Pertamina",
				ShortName:   "Pedeve",
				Code:        "PDV",
				Description: "Perusahaan Holding Induk Pedeve Pertamina",
				Status:      "Aktif",
				ParentID:    nil, // Holding has no parent
				Currency:    "IDR",
			}
			holding, err := companyUseCase.CreateCompanyFull(createData)
			if err != nil {
				fmt.Printf("   ❌ Failed to create holding: %v\n", err)
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "ERROR: Failed to create holding: %v\n", err)
				os.Stderr.Sync()
				return
			}
			holdingID = holding.ID
			fmt.Printf("   ✅ Created: %s (ID: %s, Level: %d)\n", holding.Name, holdingID, holding.Level)
			os.Stdout.Sync()
			fmt.Fprintf(os.Stderr, "STEP: Holding created successfully, ID: %s\n", holdingID)
			os.Stderr.Sync()
		}
	}
	fmt.Fprintf(os.Stderr, "STEP: Holding company step completed, holdingID: %s\n", holdingID)
	os.Stderr.Sync()

	// Create admin user for holding (check if exists first)
	// Skip user creation if admin role is not available
	if adminRole == nil {
		fmt.Println("   ⚠️  Skipping user creation for holding (admin role not available)")
	} else {
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
	}
	fmt.Println()

	// 2. Create Level 1 Companies (direct children of holding) - All companies are level 1
	fmt.Println("2️⃣  Creating Level 1 Companies (direct children of holding)...")
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "STEP: Starting level 1 companies creation, count: %d\n", 37)
	os.Stderr.Sync()
	level1Companies := []struct {
		name        string
		code        string
		description string
		username    string
		email       string
	}{
		{"PT Pertamina Hulu Energi Arun", "PHE-ARUN", "Perusahaan hulu energi Arun", "admin.phe.arun", "admin.phe.arun@pedeve.com"},
		{"PT Pertamina Internasional EP", "PIE", "Perusahaan internasional EP", "admin.pie", "admin.pie@pedeve.com"},
		{"PT Pertamina Irak EP", "PIE-IRAK", "Perusahaan Irak EP", "admin.pie.irak", "admin.pie.irak@pedeve.com"},
		{"PT Pertamina Hulu West Ganal", "PHE-WG", "Perusahaan hulu West Ganal", "admin.phe.wg", "admin.phe.wg@pedeve.com"},
		{"PT Pertamina Retail", "PRT", "Perusahaan retail dan SPBU", "admin.prt", "admin.prt@pedeve.com"},
		{"PT Pertamina Power Indonesia", "PPI", "Perusahaan power Indonesia", "admin.ppi", "admin.ppi@pedeve.com"},
		{"PT Pertamina Trans Kontinental", "PTK", "Perusahaan trans kontinental", "admin.ptk", "admin.ptk@pedeve.com"},
		{"PT Pertamina East Natuna", "PEN", "Perusahaan East Natuna", "admin.pen", "admin.pen@pedeve.com"},
		{"PT Pertamina Ep Cepu ADK", "PEP-CEPU-ADK", "Perusahaan EP Cepu ADK", "admin.pep.cepu.adk", "admin.pep.cepu.adk@pedeve.com"},
		{"PT Pertamina Algeria EP", "PIE-ALG", "Perusahaan Algeria EP", "admin.pie.alg", "admin.pie.alg@pedeve.com"},
		{"PT Pertamina Malaysia EP", "PIE-MYS", "Perusahaan Malaysia EP", "admin.pie.mys", "admin.pie.mys@pedeve.com"},
		{"PT Patra Jasa", "PJ", "Perusahaan Patra Jasa", "admin.pj", "admin.pj@pedeve.com"},
		{"PT Pelita Air Service", "PAS", "Perusahaan Pelita Air Service", "admin.pas", "admin.pas@pedeve.com"},
		{"PT Pertamedika Bali Hospital", "PMB", "Perusahaan Pertamedika Bali Hospital", "admin.pmb", "admin.pmb@pedeve.com"},
		{"PT Pertamina Hulu Borneo", "PHE-BORNEO", "Perusahaan hulu Borneo", "admin.phe.borneo", "admin.phe.borneo@pedeve.com"},
		{"PT Pertamina Geothermal Energi Tbk", "PGE", "Perusahaan geothermal energi", "admin.pge", "admin.pge@pedeve.com"},
		{"PT Pertamina Bina Medika", "PBM", "Perusahaan Bina Medika", "admin.pbm", "admin.pbm@pedeve.com"},
		{"PT Pertamina Hulu Attaka", "PHE-ATTAKA", "Perusahaan hulu Attaka", "admin.phe.attaka", "admin.phe.attaka@pedeve.com"},
		{"PT Pertamina Hulu Energi", "PHE", "Perusahaan hulu energi", "admin.phe", "admin.phe@pedeve.com"},
		{"PT Pertamina EP", "PEP", "Perusahaan EP", "admin.pep", "admin.pep@pedeve.com"},
		{"PT Pertamina EP Cepu", "PEP-CEPU", "Perusahaan EP Cepu", "admin.pep.cepu", "admin.pep.cepu@pedeve.com"},
		{"PT Pertamina Drilling Service Indonesia", "PDSI", "Perusahaan drilling service Indonesia", "admin.pdsi", "admin.pdsi@pedeve.com"},
		{"PT Pertamina Hulu Indonesia", "PHI", "Perusahaan hulu Indonesia", "admin.phi", "admin.phi@pedeve.com"},
		{"PT Pertamina Hulu Sanga Sanga", "PHE-SS", "Perusahaan hulu Sanga Sanga", "admin.phe.ss", "admin.phe.ss@pedeve.com"},
		{"PT Pertamina Hulu Kalimantan Timur", "PHE-KALTIM", "Perusahaan hulu Kalimantan Timur", "admin.phe.kaltim", "admin.phe.kaltim@pedeve.com"},
		{"PT Pertamina Hulu Rokan", "PHE-ROKAN", "Perusahaan hulu Rokan", "admin.phe.rokan", "admin.phe.rokan@pedeve.com"},
		{"PT Kilang Pertamina International", "KPI", "Perusahaan kilang internasional", "admin.kpi", "admin.kpi@pedeve.com"},
		{"PT Kilang Pertamina Balikpapan", "KPB", "Perusahaan kilang Balikpapan", "admin.kpb", "admin.kpb@pedeve.com"},
		{"PT Tuban Petrochemical Industries", "TPI", "Perusahaan petrochemical Tuban", "admin.tpi", "admin.tpi@pedeve.com"},
		{"PT Pertamina Lubricants", "PL", "Perusahaan pelumas", "admin.pl", "admin.pl@pedeve.com"},
		{"PT Pertamina Maintenance and Construction", "PMC", "Perusahaan maintenance and construction", "admin.pmc", "admin.pmc@pedeve.com"},
		{"PT Pertamina Gas", "PG", "Perusahaan gas", "admin.pg", "admin.pg@pedeve.com"},
		{"PT Pertamina International Shipping", "PIS", "Perusahaan international shipping", "admin.pis", "admin.pis@pedeve.com"},
		{"PT Mitra Tours and Travel", "MTT", "Perusahaan Mitra Tours and Travel", "admin.mtt", "admin.mtt@pedeve.com"},
		{"PT Pertamina Training and Consulting", "PTC", "Perusahaan training and consulting", "admin.ptc", "admin.ptc@pedeve.com"},
		{"PT Badak Natural Gas Liquefaction", "BNGL", "Perusahaan Badak Natural Gas Liquefaction", "admin.bngl", "admin.bngl@pedeve.com"},
		{"PT Trans Javagas Pipeline", "TJG", "Perusahaan Trans Javagas Pipeline", "admin.tjg", "admin.tjg@pedeve.com"},
	}

	level1IDs := make([]string, len(level1Companies))
	for i, comp := range level1Companies {
		fmt.Fprintf(os.Stderr, "STEP: Processing company %d/%d: %s (%s)\n", i+1, len(level1Companies), comp.name, comp.code)
		os.Stderr.Sync()

		// Check if company with same code already exists (even if inactive)
		existing, _ := companyRepo.GetByCode(comp.code)
		if existing != nil {
			// Use existing company ID
			level1IDs[i] = existing.ID
			fmt.Printf("   ⚠️  Company with code %s already exists: %s (reusing)\n", comp.code, existing.Name)

			// If company is inactive, activate it first before updating
			if !existing.IsActive {
				fmt.Printf("   ⚠️  Company %s is inactive, activating it first...\n", comp.code)
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "STEP: Activating inactive company %s...\n", comp.code)
				os.Stderr.Sync()

				// Directly update is_active in database to bypass validation
				db := database.GetDB()
				if err := db.Model(&domain.CompanyModel{}).Where("id = ?", existing.ID).Update("is_active", true).Error; err != nil {
					fmt.Printf("   ❌ Failed to activate %s: %v\n", comp.name, err)
					os.Stdout.Sync()
					fmt.Fprintf(os.Stderr, "ERROR: Failed to activate company %s: %v\n", comp.code, err)
					os.Stderr.Sync()
					continue
				}
				fmt.Printf("   ✅ Activated company %s\n", comp.code)
				os.Stdout.Sync()

				// Reload company to get updated data
				existing, _ = companyRepo.GetByID(existing.ID)
			}

			// Update existing company using usecase to ensure proper level calculation
			updateData := &domain.CompanyUpdateRequest{
				Name:        comp.name,
				ShortName:   comp.name,
				Description: comp.description,
				Status:      "Aktif",
				ParentID:    &holdingID, // Set parent to holding
			}
			updated, err := companyUseCase.UpdateCompanyFull(existing.ID, updateData)
			if err != nil {
				fmt.Printf("   ❌ Failed to update %s: %v\n", comp.name, err)
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "ERROR: Failed to update company %s: %v\n", comp.code, err)
				os.Stderr.Sync()
				continue
			}
			level1IDs[i] = updated.ID
			fmt.Printf("   ✅ Updated: %s (ID: %s, Level: %d)\n", comp.name, updated.ID, updated.Level)
			os.Stdout.Sync()
			fmt.Fprintf(os.Stderr, "STEP: Company updated: %s\n", comp.code)
			os.Stderr.Sync()
		} else {
			// Create new company using CreateCompanyFull to ensure proper logic
			fmt.Fprintf(os.Stderr, "STEP: Creating new company: %s\n", comp.code)
			os.Stderr.Sync()
			createData := &domain.CompanyCreateRequest{
				Name:        comp.name,
				ShortName:   comp.name,
				Code:        comp.code,
				Description: comp.description,
				Status:      "Aktif",
				ParentID:    &holdingID, // Set parent to holding
				Currency:    "IDR",
			}
			company, err := companyUseCase.CreateCompanyFull(createData)
			if err != nil {
				fmt.Printf("   ❌ Failed to create %s: %v\n", comp.name, err)
				os.Stdout.Sync()
				fmt.Fprintf(os.Stderr, "ERROR: Failed to create company %s: %v\n", comp.code, err)
				os.Stderr.Sync()
				continue
			}
			level1IDs[i] = company.ID
			fmt.Printf("   ✅ Created: %s (ID: %s, Level: %d)\n", comp.name, company.ID, company.Level)
			os.Stdout.Sync()
			fmt.Fprintf(os.Stderr, "STEP: Company created: %s (ID: %s)\n", comp.code, company.ID)
			os.Stderr.Sync()
		}

		// Create admin user (check if exists first)
		// Skip user creation if admin role is not available
		if adminRole == nil {
			fmt.Printf("      ⚠️  Skipping user creation for %s (admin role not available)\n", comp.name)
			continue
		}

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

	// 3. Create Administrator User (Global Role)
	if administratorRole != nil {
		fmt.Println("3️⃣  Creating/Updating Administrator User...")
		existingAdministrator, _ := userRepo.GetByUsername("administrator")
		administratorRoleID := administratorRole.ID

		if existingAdministrator != nil {
			// User already exists - update role and password
			existingAdministrator.RoleID = &administratorRoleID
			existingAdministrator.Role = "administrator"
			existingAdministrator.Email = "administrator@pertamina.com"
			existingAdministrator.CompanyID = nil // Administrator is global role, no company assignment
			// Update password
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Pedeve123!@#"), bcrypt.DefaultCost)
			existingAdministrator.Password = string(hashedPassword)
			if err := userRepo.Update(existingAdministrator); err != nil {
				fmt.Printf("   ⚠️  Failed to update administrator user (ID: %s): %v\n", existingAdministrator.ID, err)
			} else {
				fmt.Printf("   ⚠️  Administrator user already exists (ID: %s, updated)\n", existingAdministrator.ID)
			}
		} else {
			// Create new administrator user
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Pedeve123!@#"), bcrypt.DefaultCost)
			administratorUser := &domain.UserModel{
				ID:        uuid.GenerateUUID(),
				Username:  "administrator",
				Email:     "administrator@pertamina.com",
				Password:  string(hashedPassword),
				Role:      "administrator",
				RoleID:    &administratorRoleID,
				CompanyID: nil, // Administrator is global role, no company assignment
				IsActive:  true,
			}
			if err := userRepo.Create(administratorUser); err != nil {
				fmt.Printf("   ⚠️  Failed to create administrator user: %v\n", err)
			} else {
				fmt.Printf("   ✅ Created administrator user: %s (email: %s, password: Pedeve123!@#)\n", administratorUser.Username, administratorUser.Email)
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Println("   ✅ 1 Holding Company (Pedeve Pertamina - Level 0)")
	fmt.Printf("   ✅ %d Level 1 Companies (Anak Perusahaan langsung dari holding)\n", len(level1Companies))
	fmt.Printf("   ✅ Total: %d Companies (1 holding + %d subsidiaries)\n", len(level1Companies)+1, len(level1Companies))
	fmt.Printf("   ✅ Total: %d Admin Users (1 for holding + %d for subsidiaries)\n", len(level1Companies)+1, len(level1Companies))
	if administratorRole != nil {
		fmt.Println("   ✅ 1 Administrator User (Global Role)")
	}
	fmt.Println()
	fmt.Println("🔑 Default Password for all users: admin123")
	if administratorRole != nil {
		fmt.Println("🔑 Administrator Password: Pedeve123!@#")
	}
	fmt.Println()
	fmt.Println("📋 Company Structure:")
	fmt.Println("   Pedeve Pertamina (Holding - Level 0)")
	for _, comp := range level1Companies {
		fmt.Printf("   ├── %s (Level 1)\n", comp.name)
	}
	fmt.Println()
	fmt.Println("🎉 Seeding completed successfully!")
	os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "STEP: Seeding completed successfully!\n")
	os.Stderr.Sync()
}
