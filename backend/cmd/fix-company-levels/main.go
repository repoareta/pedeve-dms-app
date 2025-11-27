package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
)

func main() {
	// Set DATABASE_URL if not set
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable")
	}

	fmt.Println("🔧 Fixing Company Levels")
	fmt.Println()

	// Init logger
	logger.InitLogger()
	defer logger.Sync()

	// Init database
	database.InitDB()
	db := database.GetDB()

	fmt.Println("✅ Connected to PostgreSQL")
	fmt.Println()

	// Fix levels using recursive update
	fmt.Println("🔄 Updating company levels...")
	
	// First, set all companies with parent_id = NULL to level 0 (holding)
	result := db.Exec(`
		UPDATE companies 
		SET level = 0 
		WHERE parent_id IS NULL AND is_active = true
	`)
	if result.Error != nil {
		log.Fatalf("❌ Failed to update root companies: %v", result.Error)
	}
	fmt.Printf("   ✅ Set root companies (parent_id = NULL) to level 0 (%d rows)\n", result.RowsAffected)

	// Then, recursively update levels based on parent's level
	// We'll do this in a loop until no more updates are needed
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		result := db.Exec(`
			UPDATE companies c
			SET level = p.level + 1
			FROM companies p
			WHERE c.parent_id = p.id 
			  AND c.level != p.level + 1
			  AND c.is_active = true
			  AND p.is_active = true
		`)
		if result.Error != nil {
			log.Fatalf("❌ Failed to update company levels: %v", result.Error)
		}

		if result.RowsAffected == 0 {
			fmt.Printf("   ✅ No more updates needed (iteration %d)\n", i+1)
			break
		}
		fmt.Printf("   📝 Updated %d companies (iteration %d)\n", result.RowsAffected, i+1)
	}

	// Verify results
	fmt.Println()
	fmt.Println("📊 Verification:")
	var companies []domain.CompanyModel
	err := db.Where("is_active = ?", true).Order("level, name").Find(&companies).Error
	if err != nil {
		log.Fatalf("❌ Failed to query companies: %v", err)
	}

	fmt.Println("   Level | Name                | Code                | Parent ID")
	fmt.Println("   ------|---------------------|---------------------|----------")
	for _, company := range companies {
		parentStr := "NULL"
		if company.ParentID != nil {
			parentStr = (*company.ParentID)[:8] + "..."
		}

		fmt.Printf("   %-5d | %-19s | %-19s | %s\n", company.Level, company.Name, company.Code, parentStr)
	}

	fmt.Println()
	fmt.Printf("🎉 Company levels fixed successfully! Total: %d companies\n", len(companies))
}

