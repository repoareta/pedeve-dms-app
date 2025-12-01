package repository

import (
	"testing"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestReportRepository_Create tests creating a new report
func TestReportRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Create report successfully", func(t *testing.T) {
		// Create test company first
		company := &domain.CompanyModel{
			ID:       uuid.GenerateUUID(),
			Code:     "TEST001",
			Name:     "Test Company",
			Level:    0,
			IsActive: true,
		}
		err := testDB.Create(company).Error
		require.NoError(t, err)

		// Create test user
		user := &domain.UserModel{
			ID:       uuid.GenerateUUID(),
			Username: "testuser",
			Email:    "test@example.com",
			Password: "hashedpassword",
			IsActive: true,
		}
		err = testDB.Create(user).Error
		require.NoError(t, err)

		report := &domain.ReportModel{
			ID:             uuid.GenerateUUID(),
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &user.ID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
			Remark:         stringPtr("Test remark"),
		}

		err = repo.Create(report)
		require.NoError(t, err)
		assert.NotEmpty(t, report.ID)
	})

	t.Run("Create report with missing required fields", func(t *testing.T) {
		// Note: GORM and SQLite may not strictly enforce NOT NULL constraints
		// Validation is typically done at usecase/handler level
		// This test verifies that repository doesn't crash with invalid data
		report := &domain.ReportModel{
			ID: uuid.GenerateUUID(),
			// Missing required fields: Period, CompanyID, Revenue, Opex, NPAT, Dividend, FinancialRatio
		}

		err := repo.Create(report)
		// Repository level doesn't validate - validation happens at usecase level
		// We just verify it doesn't panic
		_ = err // Error may or may not occur depending on DB constraints
	})
}

// TestReportRepository_GetByID tests retrieving a report by ID
func TestReportRepository_GetByID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Get existing report", func(t *testing.T) {
		// Setup test data
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		report := createTestReport(t, testDB, company.ID, user.ID)

		// Test
		result, err := repo.GetByID(report.ID)
		require.NoError(t, err)
		assert.Equal(t, report.ID, result.ID)
		assert.Equal(t, report.Period, result.Period)
	})

	t.Run("Get non-existent report", func(t *testing.T) {
		_, err := repo.GetByID("non-existent-id")
		assert.Error(t, err)
	})
}

// TestReportRepository_GetAll tests retrieving all reports
func TestReportRepository_GetAll(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Get all reports", func(t *testing.T) {
		// Setup test data
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		createTestReport(t, testDB, company.ID, user.ID)
		createTestReport(t, testDB, company.ID, user.ID)

		// Test
		reports, err := repo.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(reports), 2)
	})
}

// TestReportRepository_Update tests updating a report
func TestReportRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Update report successfully", func(t *testing.T) {
		// Setup
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		report := createTestReport(t, testDB, company.ID, user.ID)

		// Update
		report.Revenue = 150000000
		updatedRemark := "Updated remark"
		report.Remark = &updatedRemark
		err := repo.Update(report)
		require.NoError(t, err)

		// Verify
		updated, err := repo.GetByID(report.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(150000000), updated.Revenue)
		assert.NotNil(t, updated.Remark)
		assert.Equal(t, "Updated remark", *updated.Remark)
	})
}

// TestReportRepository_Delete tests deleting a report
func TestReportRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Delete report successfully", func(t *testing.T) {
		// Setup
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		report := createTestReport(t, testDB, company.ID, user.ID)

		// Delete
		err := repo.Delete(report.ID)
		require.NoError(t, err)

		// Verify
		_, err = repo.GetByID(report.ID)
		assert.Error(t, err)
	})
}

// TestReportRepository_GetByCompanyID tests retrieving reports by company ID
func TestReportRepository_GetByCompanyID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Get reports by company ID", func(t *testing.T) {
		// Setup
		company1 := createTestCompany(t, testDB)
		company2 := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)

		// Create reports for company1
		createTestReport(t, testDB, company1.ID, user.ID)
		createTestReport(t, testDB, company1.ID, user.ID)

		// Create report for company2
		createTestReport(t, testDB, company2.ID, user.ID)

		// Test
		reports, err := repo.GetByCompanyID(company1.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(reports))
		for _, report := range reports {
			assert.Equal(t, company1.ID, report.CompanyID)
		}
	})

	t.Run("Get reports for company with no reports", func(t *testing.T) {
		company := createTestCompany(t, testDB)
		reports, err := repo.GetByCompanyID(company.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(reports))
	})
}

// TestReportRepository_GetByCompanyIDs tests retrieving reports by multiple company IDs
func TestReportRepository_GetByCompanyIDs(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Get reports by multiple company IDs", func(t *testing.T) {
		// Setup
		company1 := createTestCompany(t, testDB)
		company2 := createTestCompany(t, testDB)
		company3 := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)

		// Create reports
		createTestReport(t, testDB, company1.ID, user.ID)
		createTestReport(t, testDB, company2.ID, user.ID)
		createTestReport(t, testDB, company3.ID, user.ID)

		// Test
		reports, err := repo.GetByCompanyIDs([]string{company1.ID, company2.ID})
		require.NoError(t, err)
		assert.Equal(t, 2, len(reports))
		for _, report := range reports {
			assert.Contains(t, []string{company1.ID, company2.ID}, report.CompanyID)
		}
	})
}

// TestReportRepository_GetByCompanyIDAndPeriod tests retrieving report by company ID and period
func TestReportRepository_GetByCompanyIDAndPeriod(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Get report by company ID and period", func(t *testing.T) {
		// Setup
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		report := createTestReport(t, testDB, company.ID, user.ID)

		// Test
		result, err := repo.GetByCompanyIDAndPeriod(company.ID, report.Period)
		require.NoError(t, err)
		assert.Equal(t, report.ID, result.ID)
		assert.Equal(t, report.Period, result.Period)
	})

	t.Run("Get non-existent report by company ID and period", func(t *testing.T) {
		company := createTestCompany(t, testDB)
		_, err := repo.GetByCompanyIDAndPeriod(company.ID, "2025-99")
		assert.Error(t, err)
	})
}

// TestReportRepository_Count tests counting reports
func TestReportRepository_Count(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Count reports", func(t *testing.T) {
		// Setup
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		createTestReport(t, testDB, company.ID, user.ID)
		createTestReport(t, testDB, company.ID, user.ID)

		// Test
		count, err := repo.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}

// TestReportRepository_DeleteAll tests deleting all reports
func TestReportRepository_DeleteAll(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer helpers.CleanupTestDB(t, testDB)

	repo := NewReportRepositoryWithDB(testDB)

	t.Run("Delete all reports", func(t *testing.T) {
		// Setup
		company := createTestCompany(t, testDB)
		user := createTestUser(t, testDB)
		createTestReport(t, testDB, company.ID, user.ID)
		createTestReport(t, testDB, company.ID, user.ID)

		// Verify reports exist
		countBefore, _ := repo.Count()
		assert.GreaterOrEqual(t, countBefore, int64(2))

		// Delete all
		err := repo.DeleteAll()
		require.NoError(t, err)

		// Verify all deleted
		countAfter, _ := repo.Count()
		assert.Equal(t, int64(0), countAfter)
	})
}

// Helper functions
func createTestCompany(t *testing.T, db *gorm.DB) *domain.CompanyModel {
	// Use unique code to avoid constraint issues
	uniqueCode := "TEST" + uuid.GenerateUUID()[:8]
	company := &domain.CompanyModel{
		ID:       uuid.GenerateUUID(),
		Code:     uniqueCode,
		Name:     "Test Company " + uniqueCode,
		Level:    0,
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)
	return company
}

func createTestUser(t *testing.T, db *gorm.DB) *domain.UserModel {
	// Use unique username and email to avoid constraint issues
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

func createTestReport(t *testing.T, db *gorm.DB, companyID, userID string) *domain.ReportModel {
	userIDPtr := &userID
	remark := "Test remark"
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
		Remark:         &remark,
	}
	err := db.Create(report).Error
	require.NoError(t, err)
	return report
}

func stringPtr(s string) *string {
	return &s
}

