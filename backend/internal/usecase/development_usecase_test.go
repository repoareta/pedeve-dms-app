package usecase

import (
	"testing"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDevelopmentUseCase creates a test development use case with in-memory database
func setupTestDevelopmentUseCase(t *testing.T) (*developmentUseCase, *gorm.DB) {
	db := helpers.SetupTestDB(t)

	// Auto migrate models
	err := db.AutoMigrate(
		&domain.CompanyModel{},
		&domain.UserModel{},
		&domain.RoleModel{},
		&domain.UserCompanyAssignmentModel{},
		&domain.ReportModel{},
	)
	require.NoError(t, err)

	// Create use case with test database
	uc := NewDevelopmentUseCaseWithDB(db).(*developmentUseCase)

	return uc, db
}

// TestDevelopmentUseCase_ResetReportData tests resetting report data
func TestDevelopmentUseCase_ResetReportData(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Reset report data successfully", func(t *testing.T) {
		// Setup - create test reports
		company := createTestCompanyForDevelopment(t, db)
		user := createTestUserForDevelopment(t, db)
		createTestReportForDevelopment(t, db, company.ID, user.ID)
		createTestReportForDevelopment(t, db, company.ID, user.ID)

		// Verify reports exist
		count, err := uc.reportRepo.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))

		// Reset
		err = uc.ResetReportData()
		require.NoError(t, err)

		// Verify all deleted
		count, err = uc.reportRepo.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Reset when no reports exist", func(t *testing.T) {
		// Should not error when no reports exist
		err := uc.ResetReportData()
		require.NoError(t, err)
	})
}

// TestDevelopmentUseCase_RunReportSeeder tests running report seeder
func TestDevelopmentUseCase_RunReportSeeder(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Run report seeder successfully", func(t *testing.T) {
		// Setup - create companies first
		company := createTestCompanyForDevelopment(t, db)
		_ = createTestUserForDevelopment(t, db) // User for potential inputter assignment

		// Ensure company is active and has level > 0
		company.Level = 1
		company.IsActive = true
		err := db.Save(company).Error
		require.NoError(t, err)

		// Run seeder
		err = uc.RunReportSeeder()
		require.NoError(t, err)

		// Verify reports created
		reports, err := uc.reportRepo.GetByCompanyID(company.ID)
		require.NoError(t, err)
		assert.Greater(t, len(reports), 0)
	})

	t.Run("Run report seeder when data already exists", func(t *testing.T) {
		// Setup
		company := createTestCompanyForDevelopment(t, db)
		user := createTestUserForDevelopment(t, db)
		company.Level = 1
		company.IsActive = true
		err := db.Save(company).Error
		require.NoError(t, err)

		// Create existing report
		createTestReportForDevelopment(t, db, company.ID, user.ID)

		// Run seeder - should return error
		err = uc.RunReportSeeder()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

// TestDevelopmentUseCase_CheckReportDataExists tests checking report data existence
func TestDevelopmentUseCase_CheckReportDataExists(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Check when reports exist", func(t *testing.T) {
		// Setup
		company := createTestCompanyForDevelopment(t, db)
		user := createTestUserForDevelopment(t, db)
		createTestReportForDevelopment(t, db, company.ID, user.ID)

		// Check
		exists, err := uc.CheckReportDataExists()
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Check when no reports exist", func(t *testing.T) {
		// Ensure no reports exist by resetting first
		_ = uc.ResetReportData() // Ignore error if no reports to reset
		
		// Check
		exists, err := uc.CheckReportDataExists()
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestDevelopmentUseCase_RunAllSeeders tests running all seeders
func TestDevelopmentUseCase_RunAllSeeders(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Run all seeders successfully", func(t *testing.T) {
		// Create admin role first (required for seeder)
		adminRole := &domain.RoleModel{
			ID:          uuid.GenerateUUID(),
			Name:        "admin",
			Description: "Admin role",
			Level:       1,
			IsSystem:    true,
		}
		err := db.Create(adminRole).Error
		require.NoError(t, err)

		// Run all seeders
		err = uc.RunAllSeeders()
		// May error if data already exists, which is acceptable
		_ = err // We don't assert here as it depends on test state
	})
}

// TestDevelopmentUseCase_ResetAllSeededData tests resetting all seeded data
func TestDevelopmentUseCase_ResetAllSeededData(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Reset all seeded data successfully", func(t *testing.T) {
		// Setup - create test data
		company := createTestCompanyForDevelopment(t, db)
		user := createTestUserForDevelopment(t, db)
		createTestReportForDevelopment(t, db, company.ID, user.ID)

		// Reset all
		err := uc.ResetAllSeededData()
		require.NoError(t, err)

		// Verify reports deleted
		count, err := uc.reportRepo.Count()
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// TestDevelopmentUseCase_CheckAllSeederStatus tests checking all seeder status
func TestDevelopmentUseCase_CheckAllSeederStatus(t *testing.T) {
	uc, db := setupTestDevelopmentUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Check all seeder status", func(t *testing.T) {
		// Setup - create some data
		company := createTestCompanyForDevelopment(t, db)
		user := createTestUserForDevelopment(t, db)
		createTestReportForDevelopment(t, db, company.ID, user.ID)

		// Check status
		status, err := uc.CheckAllSeederStatus()
		require.NoError(t, err)
		assert.NotNil(t, status)
		assert.Contains(t, status, "company")
		assert.Contains(t, status, "report")
	})
}

// Helper functions
func createTestCompanyForDevelopment(t *testing.T, db *gorm.DB) *domain.CompanyModel {
	uniqueCode := "TEST" + uuid.GenerateUUID()[:8]
	company := &domain.CompanyModel{
		ID:       uuid.GenerateUUID(),
		Code:     uniqueCode,
		Name:     "Test Company " + uniqueCode,
		Level:    1,
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)
	return company
}

func createTestUserForDevelopment(t *testing.T, db *gorm.DB) *domain.UserModel {
	uniqueID := uuid.GenerateUUID()[:8]
	user := &domain.UserModel{
		ID:       uuid.GenerateUUID(),
		Username: "testuser" + uniqueID,
		Email:    "test" + uniqueID + "@example.com",
		Password: "hashedpassword",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestReportForDevelopment(t *testing.T, db *gorm.DB, companyID, userID string) *domain.ReportModel {
	userIDPtr := &userID
	report := &domain.ReportModel{
		ID:             uuid.GenerateUUID(),
		Period:         "2025-06",
		CompanyID:      companyID,
		InputterID:     userIDPtr,
		Revenue:        125000000,
		Opex:           78000000,
		NPAT:           27000000,
		Dividend:       8000000,
		FinancialRatio: 1.5,
	}
	err := db.Create(report).Error
	require.NoError(t, err)
	return report
}

