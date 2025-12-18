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

// setupTestReportUseCase creates a test report use case with in-memory database
func setupTestReportUseCase(t *testing.T) (*reportUseCase, *gorm.DB) {
	db := helpers.SetupTestDB(t)
	
	// Auto migrate models
	err := db.AutoMigrate(
		&domain.CompanyModel{},
		&domain.UserModel{},
		&domain.ReportModel{},
	)
	require.NoError(t, err)

	// Create use case with test database
	uc := NewReportUseCaseWithDB(db).(*reportUseCase)
	
	return uc, db
}

// TestReportUseCase_CreateReport tests creating a new report
func TestReportUseCase_CreateReport(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Create report successfully", func(t *testing.T) {
		// Setup test data
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)

		// Create report request
		userID := user.ID
		req := &domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
			Remark:         stringPtr("Test remark"),
		}

		report, err := uc.CreateReport(req)
		require.NoError(t, err)
		assert.NotEmpty(t, report.ID)
		assert.Equal(t, req.Period, report.Period)
		assert.Equal(t, req.Revenue, report.Revenue)
	})

	t.Run("Create report with invalid company ID", func(t *testing.T) {
		// Don't create user, just use a random ID since company validation happens first
		randomUserID := uuid.GenerateUUID()
		req := &domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      "non-existent-company",
			InputterID:     &randomUserID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		}

		_, err := uc.CreateReport(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "company")
	})

	t.Run("Create report with invalid inputter ID", func(t *testing.T) {
		// Create company with unique code to avoid constraint issues
		companyID := uuid.GenerateUUID()
		uniqueCode := "TEST" + uuid.GenerateUUID()[:8]
		company := &domain.CompanyModel{
			ID:       companyID,
			Code:     uniqueCode,
			Name:     "Test Company Invalid User",
			Level:    1,
			IsActive: true,
		}
		err := db.Create(company).Error
		require.NoError(t, err)

		invalidUserID := "non-existent-user"
		req := &domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &invalidUserID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		}

		_, err = uc.CreateReport(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user")
	})

	t.Run("Create report with missing required fields", func(t *testing.T) {
		req := &domain.CreateReportRequest{
			// Missing required fields
		}

		_, err := uc.CreateReport(req)
		assert.Error(t, err)
	})
}

// TestReportUseCase_GetReport tests retrieving a report by ID
func TestReportUseCase_GetReport(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Get existing report", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID
		report, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Test
		result, err := uc.GetReportByID(report.ID)
		require.NoError(t, err)
		assert.Equal(t, report.ID, result.ID)
	})

	t.Run("Get non-existent report", func(t *testing.T) {
		_, err := uc.GetReportByID("non-existent-id")
		assert.Error(t, err)
	})
}

// TestReportUseCase_GetAllReports tests retrieving all reports
func TestReportUseCase_GetAllReports(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Get all reports", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)

		userID := user.ID
		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		_, err = uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-07",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        150000000,
			Opex:           90000000,
			NPAT:           35000000,
			Dividend:       10000000,
			FinancialRatio: 1.6,
		})
		require.NoError(t, err)

		// Test
		reports, err := uc.GetAllReports("superadmin", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(reports), 2)
	})
}

// TestReportUseCase_UpdateReport tests updating a report
func TestReportUseCase_UpdateReport(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Update report successfully", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID
		report, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Update
		updatedRemark := "Updated remark"
		updateReq := &domain.UpdateReportRequest{
			Revenue:        int64Ptr(150000000),
			Opex:            int64Ptr(90000000),
			NPAT:            int64Ptr(35000000),
			Dividend:        int64Ptr(10000000),
			FinancialRatio:  float64Ptr(1.6),
			Remark:          &updatedRemark,
		}

		updated, err := uc.UpdateReport(report.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, int64(150000000), updated.Revenue)
		assert.Equal(t, "Updated remark", *updated.Remark)
	})

	t.Run("Update non-existent report", func(t *testing.T) {
		updateReq := &domain.UpdateReportRequest{
			Revenue: int64Ptr(150000000),
		}

		_, err := uc.UpdateReport("non-existent-id", updateReq)
		assert.Error(t, err)
	})
}

// TestReportUseCase_DeleteReport tests deleting a report
func TestReportUseCase_DeleteReport(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Delete report successfully", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID
		report, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Delete
		err = uc.DeleteReport(report.ID)
		require.NoError(t, err)

		// Verify
		_, err = uc.GetReportByID(report.ID)
		assert.Error(t, err)
	})
}

// TestReportUseCase_GetReportsByCompanyID tests retrieving reports by company ID with RBAC
func TestReportUseCase_GetReportsByCompanyID(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Superadmin can get reports for any company", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Test
		reports, err := uc.GetReportsByCompanyID(company.ID, "superadmin", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(reports), 1)
	})

	t.Run("Admin can get reports for their company", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Test
		reports, err := uc.GetReportsByCompanyID(company.ID, "admin", &company.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(reports), 1)
	})

	t.Run("Admin cannot get reports for other company", func(t *testing.T) {
		// Setup
		company1 := createTestCompanyForReport(t, db)
		company2 := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company1.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Test - admin from company2 trying to access company1 reports
		// Should return error (access denied) for security
		_, err = uc.GetReportsByCompanyID(company1.ID, "admin", &company2.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})
}

// TestReportUseCase_ValidateReportAccess tests RBAC validation for report access
func TestReportUseCase_ValidateReportAccess(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Superadmin can access any report", func(t *testing.T) {
		company := createTestCompanyForReport(t, db)
		hasAccess, err := uc.ValidateReportAccess("superadmin", nil, company.ID)
		require.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("Admin can access their company reports", func(t *testing.T) {
		company := createTestCompanyForReport(t, db)
		hasAccess, err := uc.ValidateReportAccess("admin", &company.ID, company.ID)
		require.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("Admin cannot access other company reports", func(t *testing.T) {
		company1 := createTestCompanyForReport(t, db)
		company2 := createTestCompanyForReport(t, db)
		hasAccess, err := uc.ValidateReportAccess("admin", &company1.ID, company2.ID)
		require.NoError(t, err)
		assert.False(t, hasAccess)
	})
}

// TestReportUseCase_GetAllReports_RBAC tests RBAC for GetAllReports
func TestReportUseCase_GetAllReports_RBAC(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Admin only sees their company reports", func(t *testing.T) {
		// Setup
		company1 := createTestCompanyForReport(t, db)
		company2 := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		// Create reports for both companies
		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company1.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		_, err = uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-07",
			CompanyID:      company2.ID,
			InputterID:     &userID,
			Revenue:        150000000,
			Opex:           90000000,
			NPAT:           35000000,
			Dividend:       10000000,
			FinancialRatio: 1.6,
		})
		require.NoError(t, err)

		// Test - admin from company1 should only see company1 reports
		reports, err := uc.GetAllReports("admin", &company1.ID)
		require.NoError(t, err)
		for _, report := range reports {
			assert.Equal(t, company1.ID, report.CompanyID)
		}
	})

	t.Run("Regular user only sees their company reports", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Test
		reports, err := uc.GetAllReports("staff", &company.ID)
		require.NoError(t, err)
		for _, report := range reports {
			assert.Equal(t, company.ID, report.CompanyID)
		}
	})
}

// TestReportUseCase_CreateReport_DuplicatePeriod tests duplicate period validation
func TestReportUseCase_CreateReport_DuplicatePeriod(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Cannot create duplicate report for same company and period", func(t *testing.T) {
		// Setup
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		// Create first report
		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Try to create duplicate
		_, err = uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company.ID,
			InputterID:     &userID,
			Revenue:        150000000,
			Opex:           90000000,
			NPAT:           35000000,
			Dividend:       10000000,
			FinancialRatio: 1.6,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Can create report for same period but different company", func(t *testing.T) {
		// Setup
		company1 := createTestCompanyForReport(t, db)
		company2 := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		// Create report for company1
		_, err := uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company1.ID,
			InputterID:     &userID,
			Revenue:        125000000,
			Opex:           78000000,
			NPAT:           27000000,
			Dividend:       8000000,
			FinancialRatio: 1.5,
		})
		require.NoError(t, err)

		// Create report for company2 with same period (should succeed)
		_, err = uc.CreateReport(&domain.CreateReportRequest{
			Period:         "2025-06",
			CompanyID:      company2.ID,
			InputterID:     &userID,
			Revenue:        150000000,
			Opex:           90000000,
			NPAT:           35000000,
			Dividend:       10000000,
			FinancialRatio: 1.6,
		})
		assert.NoError(t, err)
	})
}

// TestReportUseCase_CreateReport_InvalidPeriodFormat tests period format validation
func TestReportUseCase_CreateReport_InvalidPeriodFormat(t *testing.T) {
	uc, db := setupTestReportUseCase(t)
	defer helpers.CleanupTestDB(t, db)

	t.Run("Reject invalid period format", func(t *testing.T) {
		company := createTestCompanyForReport(t, db)
		user := createTestUserForReport(t, db)
		userID := user.ID

		invalidPeriods := []string{
			"2025",
			"2025/06",
			"06-2025",
			"2025-6",
			"2025-006",
			"invalid",
		}

		for _, period := range invalidPeriods {
			_, err := uc.CreateReport(&domain.CreateReportRequest{
				Period:         period,
				CompanyID:      company.ID,
				InputterID:     &userID,
				Revenue:        125000000,
				Opex:           78000000,
				NPAT:           27000000,
				Dividend:       8000000,
				FinancialRatio: 1.5,
			})
			assert.Error(t, err, "Should reject period: %s", period)
		}
	})
}

// Helper functions
func createTestCompanyForReport(t *testing.T, db *gorm.DB) *domain.CompanyModel {
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

func createTestUserForReport(t *testing.T, db *gorm.DB) *domain.UserModel {
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

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

