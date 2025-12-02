package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// setupTestReportHandler creates a test report handler with test database
func setupTestReportHandler(t *testing.T) (*ReportHandler, *gorm.DB) {
	db := helpers.SetupTestDB(t)

	// Create use case with test database
	reportUseCase := usecase.NewReportUseCaseWithDB(db)

	// Create handler
	handler := NewReportHandler(reportUseCase)
	// Ensure company use case uses test DB as well
	handler.companyUseCase = usecase.NewCompanyUseCaseWithDB(db)

	return handler, db
}

// createTestCompanyForHandler creates a test company
func createTestCompanyForHandler(t *testing.T, db *gorm.DB) *domain.CompanyModel {
	company := &domain.CompanyModel{
		ID:       uuid.GenerateUUID(),
		Code:     "TEST" + uuid.GenerateUUID()[:8],
		Name:     "Test Company",
		Level:    0,
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)
	return company
}

// createTestUserForHandler creates a test user
func createTestUserForHandler(t *testing.T, db *gorm.DB) *domain.UserModel {
	user := &domain.UserModel{
		ID:       uuid.GenerateUUID(),
		Username: "testuser" + uuid.GenerateUUID()[:8],
		Email:    "test" + uuid.GenerateUUID()[:8] + "@example.com",
		Password: "hashedpassword",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// createTestReportForHandler creates a test report
func createTestReportForHandler(t *testing.T, db *gorm.DB, companyID string, userID *string) *domain.ReportModel {
	report := &domain.ReportModel{
		ID:             uuid.GenerateUUID(),
		Period:         "2025-06",
		CompanyID:      companyID,
		InputterID:     userID,
		Revenue:        125000000,
		Opex:           78000000,
		NPAT:           27000000,
		Dividend:       5000000,
		FinancialRatio: 1.6,
		Remark:         stringPtr("Test report"),
	}
	err := db.Create(report).Error
	require.NoError(t, err)
	return report
}

func stringPtr(s string) *string {
	return &s
}

// Helper: create in-memory Excel file for upload with headers + rows
func buildUploadExcel(t *testing.T, rows [][]string) []byte {
	t.Helper()

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	headers := []string{"Period (YYYY-MM)", "Company Code", "Revenue", "OPEX", "NPAT", "Dividend", "Financial Ratio (%)", "Remark"}
	if err := f.SetSheetRow(sheet, "A1", &headers); err != nil {
		t.Fatalf("failed to set headers: %v", err)
	}

	for i, row := range rows {
		cell := "A" + strconv.Itoa(i+2)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("failed to set row %d: %v", i+2, err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write excel: %v", err)
	}
	return buf.Bytes()
}

func TestReportHandler_UploadReports(t *testing.T) {
	handler, db := setupTestReportHandler(t)
	defer helpers.CleanupTestDB(t, db)

	company := createTestCompanyForHandler(t, db)
	user := createTestUserForHandler(t, db)

	t.Run("upload valid excel creates reports", func(t *testing.T) {
		excelData := buildUploadExcel(t, [][]string{
			{"2025-08", company.Code, "100", "200", "300", "400", "1.2", "remark-1"},
			{"2025-09", company.Code, "150", "250", "350", "450", "1.5", ""},
		})

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "reports.xlsx")
		require.NoError(t, err)
		_, err = part.Write(excelData)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		app := fiber.New()
		app.Post("/reports/upload", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.UploadReports(c)
		})

		req := httptest.NewRequest("POST", "/reports/upload", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var payload struct {
			Success int                      `json:"success"`
			Failed  int                      `json:"failed"`
			Errors  []map[string]interface{} `json:"errors"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
		assert.Equal(t, 2, payload.Success)
		assert.Equal(t, 0, payload.Failed)
		assert.Len(t, payload.Errors, 0)

		var count int64
		require.NoError(t, db.Model(&domain.ReportModel{}).Count(&count).Error)
		assert.Equal(t, int64(2), count)
	})

	t.Run("upload with invalid row returns errors", func(t *testing.T) {
		excelData := buildUploadExcel(t, [][]string{
			{"2025-10", "UNKNOWN", "100", "200", "300", "400", "1.2", ""},
		})

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "invalid.xlsx")
		require.NoError(t, err)
		_, err = part.Write(excelData)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		app := fiber.New()
		app.Post("/reports/upload", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.UploadReports(c)
		})

		req := httptest.NewRequest("POST", "/reports/upload", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var payload struct {
			Success int                      `json:"success"`
			Failed  int                      `json:"failed"`
			Errors  []map[string]interface{} `json:"errors"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
		assert.Equal(t, 0, payload.Success)
		assert.Equal(t, 1, payload.Failed)
		assert.NotEmpty(t, payload.Errors)
	})
}

// TestReportHandler_ExportReportsExcel tests Excel export functionality
func TestReportHandler_ExportReportsExcel(t *testing.T) {
	handler, db := setupTestReportHandler(t)
	defer helpers.CleanupTestDB(t, db)

	// Setup test data
	company := createTestCompanyForHandler(t, db)
	user := createTestUserForHandler(t, db)
	userID := user.ID
	_ = createTestReportForHandler(t, db, company.ID, &userID)
	report2 := createTestReportForHandler(t, db, company.ID, &userID)
	report2.Period = "2025-07"
	db.Save(report2)

	t.Run("Export all reports as Excel", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/excel", func(c *fiber.Ctx) error {
			// Set context values (simulating JWT middleware)
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsExcel(c)
		})

		req := httptest.NewRequest("GET", "/export/excel", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify response
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

		// Verify file content (should be Excel file)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "Excel file should not be empty")
	})

	t.Run("Export reports with period filter", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/excel", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsExcel(c)
		})

		req := httptest.NewRequest("GET", "/export/excel?period=2025-06", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "Excel file should not be empty")
	})

	t.Run("Export reports with company filter", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/excel", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsExcel(c)
		})

		req := httptest.NewRequest("GET", "/export/excel?company_id="+company.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "Excel file should not be empty")
	})

	t.Run("Export reports with multiple company filter", func(t *testing.T) {
		company2 := createTestCompanyForHandler(t, db)
		_ = createTestReportForHandler(t, db, company2.ID, &userID)

		app := fiber.New()
		app.Get("/export/excel", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsExcel(c)
		})

		req := httptest.NewRequest("GET", "/export/excel?company_id="+company.ID+","+company2.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "Excel file should not be empty")
	})

	t.Run("Export empty reports list", func(t *testing.T) {
		// Create new handler with empty database
		emptyDB := helpers.SetupTestDB(t)
		defer helpers.CleanupTestDB(t, emptyDB)
		emptyUseCase := usecase.NewReportUseCaseWithDB(emptyDB)
		emptyHandler := NewReportHandler(emptyUseCase)

		app := fiber.New()
		app.Get("/export/excel", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return emptyHandler.ExportReportsExcel(c)
		})

		req := httptest.NewRequest("GET", "/export/excel", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Should still return Excel file (even if empty)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "Excel file should not be empty")
	})
}

// TestReportHandler_ExportReportsPDF tests PDF export functionality
func TestReportHandler_ExportReportsPDF(t *testing.T) {
	handler, db := setupTestReportHandler(t)
	defer helpers.CleanupTestDB(t, db)

	// Setup test data
	company := createTestCompanyForHandler(t, db)
	user := createTestUserForHandler(t, db)
	userID := user.ID
	_ = createTestReportForHandler(t, db, company.ID, &userID)
	report2 := createTestReportForHandler(t, db, company.ID, &userID)
	report2.Period = "2025-07"
	db.Save(report2)

	t.Run("Export all reports as PDF", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsPDF(c)
		})

		req := httptest.NewRequest("GET", "/export/pdf", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify response
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/pdf")

		// Verify file content (should be PDF file)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "PDF file should not be empty")

		// PDF files start with %PDF
		assert.True(t, bytes.HasPrefix(body.Bytes(), []byte("%PDF")), "Response should be a valid PDF file")
	})

	t.Run("Export reports with period filter", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsPDF(c)
		})

		req := httptest.NewRequest("GET", "/export/pdf?period=2025-06", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "PDF file should not be empty")
		assert.True(t, bytes.HasPrefix(body.Bytes(), []byte("%PDF")), "Response should be a valid PDF file")
	})

	t.Run("Export reports with company filter", func(t *testing.T) {
		app := fiber.New()
		app.Get("/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsPDF(c)
		})

		req := httptest.NewRequest("GET", "/export/pdf?company_id="+company.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "PDF file should not be empty")
		assert.True(t, bytes.HasPrefix(body.Bytes(), []byte("%PDF")), "Response should be a valid PDF file")
	})

	t.Run("Export reports with multiple company filter", func(t *testing.T) {
		company2 := createTestCompanyForHandler(t, db)
		_ = createTestReportForHandler(t, db, company2.ID, &userID)

		app := fiber.New()
		app.Get("/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsPDF(c)
		})

		req := httptest.NewRequest("GET", "/export/pdf?company_id="+company.ID+","+company2.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "PDF file should not be empty")
		assert.True(t, bytes.HasPrefix(body.Bytes(), []byte("%PDF")), "Response should be a valid PDF file")
	})

	t.Run("Export empty reports list", func(t *testing.T) {
		// Create new handler with empty database
		emptyDB := helpers.SetupTestDB(t)
		defer helpers.CleanupTestDB(t, emptyDB)
		emptyUseCase := usecase.NewReportUseCaseWithDB(emptyDB)
		emptyHandler := NewReportHandler(emptyUseCase)

		app := fiber.New()
		app.Get("/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return emptyHandler.ExportReportsPDF(c)
		})

		req := httptest.NewRequest("GET", "/export/pdf", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Should still return PDF file (even if empty)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := new(bytes.Buffer)
		_, err = body.ReadFrom(resp.Body)
		require.NoError(t, err)
		assert.Greater(t, body.Len(), 0, "PDF file should not be empty")
		assert.True(t, bytes.HasPrefix(body.Bytes(), []byte("%PDF")), "Response should be a valid PDF file")
	})
}

// TestReportHandler_ExportRoutesOrder tests that export routes are registered before parameterized routes
// This test ensures the route ordering bug doesn't regress
func TestReportHandler_ExportRoutesOrder(t *testing.T) {
	// This test verifies that export routes work correctly
	// by testing that they don't conflict with /reports/:id route
	handler, db := setupTestReportHandler(t)
	defer helpers.CleanupTestDB(t, db)

	company := createTestCompanyForHandler(t, db)
	user := createTestUserForHandler(t, db)
	userID := user.ID
	_ = createTestReportForHandler(t, db, company.ID, &userID)

	t.Run("Export Excel route should work (not match /reports/:id)", func(t *testing.T) {
		app := fiber.New()
		// Simulate route registration order: export routes before :id route
		app.Get("/reports/export/excel", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsExcel(c)
		})
		app.Get("/reports/:id", handler.GetReport)

		req := httptest.NewRequest("GET", "/reports/export/excel", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify it's Excel content type, not JSON
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	})

	t.Run("Export PDF route should work (not match /reports/:id)", func(t *testing.T) {
		app := fiber.New()
		// Simulate route registration order: export routes before :id route
		app.Get("/reports/export/pdf", func(c *fiber.Ctx) error {
			c.Locals("userID", user.ID)
			c.Locals("username", user.Username)
			c.Locals("roleName", "superadmin")
			c.Locals("companyID", nil)
			return handler.ExportReportsPDF(c)
		})
		app.Get("/reports/:id", handler.GetReport)

		req := httptest.NewRequest("GET", "/reports/export/pdf", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify it's PDF content type, not JSON
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/pdf")
	})
}
