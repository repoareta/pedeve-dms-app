package main

import (
	"fmt"
	"os"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// DATABASE_URL must be set via environment variable for security
	// Never hardcode database credentials in source code
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Fprintf(os.Stderr, "❌ DATABASE_URL environment variable is required. Please set it before running this command.\n")
		os.Exit(1)
	}

	fmt.Println("📁 Generating Document Folders for Existing Companies")
	fmt.Println()

	// Init logger
	logger.InitLogger()
	defer logger.Sync()
	zapLog := logger.GetLogger()

	// Init database
	database.InitDB()

	// Reduce GORM logging for seeder
	db := database.GetDB()
	if db != nil {
		db.Logger = db.Logger.LogMode(gormLogger.Silent)
	}

	fmt.Println("✅ Database initialized")
	fmt.Println()

	// Initialize repositories
	companyRepo := repository.NewCompanyRepository()
	documentRepo := repository.NewDocumentRepository()

	fmt.Println("✅ Connected to database")
	fmt.Println()

	// Get all active companies
	allCompanies, err := companyRepo.GetAll(false)
	if err != nil {
		fmt.Printf("❌ Failed to get companies: %v\n", err)
		zapLog.Error("Failed to get companies", zap.Error(err))
		return
	}

	// Filter only active companies
	var activeCompanies []domain.CompanyModel
	for _, comp := range allCompanies {
		if comp.IsActive {
			activeCompanies = append(activeCompanies, comp)
		}
	}

	if len(activeCompanies) == 0 {
		fmt.Println("❌ No active companies found.")
		return
	}

	fmt.Printf("📊 Found %d active companies\n", len(activeCompanies))
	fmt.Println()

	createdCount := 0
	skippedCount := 0
	errorCount := 0

	// Check existing folders by company_id
	var existingFolders []domain.DocumentFolderModel
	if err := db.Where("company_id IS NOT NULL").Find(&existingFolders).Error; err == nil {
		// Create a map of company IDs that already have folders
		companyWithFolders := make(map[string]bool)
		for _, folder := range existingFolders {
			if folder.CompanyID != nil {
				companyWithFolders[*folder.CompanyID] = true
			}
		}

		fmt.Println("🔄 Creating folders for companies...")
		fmt.Println()

		for _, company := range activeCompanies {
			// Skip if folder already exists for this company
			if companyWithFolders[company.ID] {
				fmt.Printf("   ⏭️  Skipping %s (folder already exists)\n", company.Name)
				skippedCount++
				continue
			}

			// Check if folder with same name already exists (for backward compatibility)
			var existingFolder domain.DocumentFolderModel
			if err := db.Where("name = ? AND company_id IS NULL", company.Name).First(&existingFolder).Error; err == nil {
				// Folder exists but without company_id, update it
				existingFolder.CompanyID = &company.ID
				if err := db.Save(&existingFolder).Error; err != nil {
					fmt.Printf("   ❌ Failed to update folder for %s: %v\n", company.Name, err)
					zapLog.Error("Failed to update folder for company",
						zap.String("company_id", company.ID),
						zap.String("company_name", company.Name),
						zap.Error(err))
					errorCount++
					continue
				}
				fmt.Printf("   ✅ Updated existing folder for %s (added company_id)\n", company.Name)
				createdCount++
				continue
			}

			// Create new folder for company
			folder := &domain.DocumentFolderModel{
				ID:        uuid.GenerateUUID(),
				Name:      company.Name,
				CompanyID: &company.ID,
				ParentID:  nil, // Folder root untuk perusahaan
				CreatedBy: "",  // System-created
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := documentRepo.CreateFolder(folder); err != nil {
				fmt.Printf("   ❌ Failed to create folder for %s: %v\n", company.Name, err)
				zapLog.Error("Failed to create folder for company",
					zap.String("company_id", company.ID),
					zap.String("company_name", company.Name),
					zap.Error(err))
				errorCount++
				continue
			}

			fmt.Printf("   ✅ Created folder for %s\n", company.Name)
			createdCount++
		}
	} else {
		fmt.Printf("❌ Failed to check existing folders: %v\n", err)
		zapLog.Error("Failed to check existing folders", zap.Error(err))
		return
	}

	fmt.Println()
	fmt.Println("📊 Summary:")
	fmt.Printf("   ✅ Created/Updated: %d folders\n", createdCount)
	fmt.Printf("   ⏭️  Skipped: %d companies (folders already exist)\n", skippedCount)
	if errorCount > 0 {
		fmt.Printf("   ❌ Errors: %d companies\n", errorCount)
	}
	fmt.Println()
	fmt.Println("✨ Done!")
}
