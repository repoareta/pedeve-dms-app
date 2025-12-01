package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	gormLogger "gorm.io/gorm/logger"
)

func main() {
	// Set DATABASE_URL if not set
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable")
	}

	fmt.Println("🌱 Seeding Reports")
	fmt.Println()

	// Init logger
	logger.InitLogger()
	defer logger.Sync()

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
	reportRepo := repository.NewReportRepository()
	companyRepo := repository.NewCompanyRepository()
	userRepo := repository.NewUserRepository()

	fmt.Println("✅ Connected to database")
	fmt.Println()

	// Get all companies (excluding holding/level 0)
	allCompanies, err := companyRepo.GetAll()
	if err != nil {
		fmt.Printf("❌ Failed to get companies: %v\n", err)
		return
	}

	// Filter to get only subsidiaries (level > 0, excluding holding)
	var subsidiaries []domain.CompanyModel
	for _, comp := range allCompanies {
		if comp.Level > 0 && comp.IsActive {
			subsidiaries = append(subsidiaries, comp)
		}
	}

	if len(subsidiaries) == 0 {
		fmt.Println("❌ No subsidiaries found. Please run seed-companies first.")
		return
	}

	fmt.Printf("📊 Found %d subsidiaries\n", len(subsidiaries))
	fmt.Println()

	// Periods to seed (2025-09, 2025-10, 2025-11, 2025-12)
	periods := []string{"2025-09", "2025-10", "2025-11", "2025-12"}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Get all users for random assignment
	allUsers, err := userRepo.GetAll()
	if err != nil {
		fmt.Printf("⚠️  Failed to get users: %v (will use null inputter)\n", err)
		allUsers = []domain.UserModel{}
	}

	totalCreated := 0
	totalSkipped := 0

	// Create reports for each subsidiary
	for i, company := range subsidiaries {
		fmt.Printf("%d️⃣  Creating reports for %s (Level %d)...\n", i+1, company.Name, company.Level)

		for _, period := range periods {
			// Check if report already exists
			existing, _ := reportRepo.GetByCompanyIDAndPeriod(company.ID, period)
			if existing != nil {
				fmt.Printf("   ⚠️  Report for period %s already exists (skipping)\n", period)
				totalSkipped++
				continue
			}

			// Generate realistic random data
			// Revenue: 50M - 500M
			revenue := int64(rand.Intn(450000000) + 50000000)
			// Opex: 30% - 70% of revenue
			opexPercent := float64(rand.Intn(40)+30) / 100.0
			opex := int64(float64(revenue) * opexPercent)
			// NPAT: Revenue - Opex - Tax (assume 25% tax on profit)
			profit := revenue - opex
			tax := int64(float64(profit) * 0.25)
			if profit < 0 {
				tax = 0
			}
			npat := profit - tax
			// Dividend: 10% - 30% of NPAT (if positive)
			dividend := int64(0)
			if npat > 0 {
				dividendPercent := float64(rand.Intn(20)+10) / 100.0
				dividend = int64(float64(npat) * dividendPercent)
			}
			// Financial Ratio: Revenue / Opex (1.0 - 3.0)
			financialRatio := float64(revenue) / float64(opex)
			if financialRatio > 3.0 {
				financialRatio = 3.0
			}

			// Randomly assign inputter (or null)
			var inputterID *string
			if len(allUsers) > 0 && rand.Float32() > 0.3 { // 70% chance to assign user
				randomUser := allUsers[rand.Intn(len(allUsers))]
				inputterID = &randomUser.ID
			}

			// Create report
			report := &domain.ReportModel{
				ID:             uuid.GenerateUUID(),
				Period:         period,
				CompanyID:      company.ID,
				InputterID:     inputterID,
				Revenue:        revenue,
				Opex:           opex,
				NPAT:           npat,
				Dividend:       dividend,
				FinancialRatio: financialRatio,
				Attachment:     nil, // Null as requested
				Remark:         nil, // Optional, can be null
			}

			if err := reportRepo.Create(report); err != nil {
				fmt.Printf("   ❌ Failed to create report for period %s: %v\n", period, err)
				continue
			}

			fmt.Printf("   ✅ Created report for period %s (Revenue: %d, NPAT: %d)\n", period, revenue, npat)
			totalCreated++
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Printf("   ✅ Total Reports Created: %d\n", totalCreated)
	fmt.Printf("   ⚠️  Total Reports Skipped (already exist): %d\n", totalSkipped)
	fmt.Printf("   📦 Total Subsidiaries: %d\n", len(subsidiaries))
	fmt.Printf("   📅 Periods per Subsidiary: 4 (September, October, November, December 2025)\n")
	fmt.Printf("   📈 Expected Total Reports: %d\n", len(subsidiaries)*4)
	fmt.Println()
	fmt.Println("🎉 Seeding completed successfully!")
}

