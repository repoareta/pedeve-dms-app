package http

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FinancialReportHandler handles financial report-related HTTP requests
type FinancialReportHandler struct {
	financialReportUseCase usecase.FinancialReportUseCase
	companyUseCase         usecase.CompanyUseCase
}

// NewFinancialReportHandler creates a new financial report handler
func NewFinancialReportHandler(financialReportUseCase usecase.FinancialReportUseCase) *FinancialReportHandler {
	return &FinancialReportHandler{
		financialReportUseCase: financialReportUseCase,
		companyUseCase:         usecase.NewCompanyUseCase(),
	}
}

// CreateFinancialReport handles financial report creation
// @Summary      Buat Financial Report Baru (RKAP atau Realisasi)
// @Description  Membuat financial report baru (RKAP tahunan atau Realisasi bulanan). RKAP hanya boleh 1x per tahun per perusahaan.
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        report  body      domain.CreateFinancialReportRequest  true  "Financial Report data"
// @Success      201     {object}  domain.FinancialReportModel
// @Failure      400     {object}  domain.ErrorResponse
// @Failure      401     {object}  domain.ErrorResponse
// @Failure      403     {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports [post]
func (h *FinancialReportHandler) CreateFinancialReport(c *fiber.Ctx) error {
	var req domain.CreateFinancialReportRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	// Get user info from JWT
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Get IP and User Agent
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "")

	// Authorization: Check if user can create report for this company
	if !utils.IsSuperAdminLike(roleName) {
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// User can only create report for their own company or descendants
		if req.CompanyID != userCompanyID {
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, req.CompanyID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only create financial report for your company or its descendants",
				})
			}
		}
	}

	report, err := h.financialReportUseCase.CreateFinancialReport(&req, userID, username, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// UpdateFinancialReport handles financial report update
// @Summary      Update Financial Report
// @Description  Mengupdate financial report yang sudah ada. Validasi RKAP 1x per tahun tetap berlaku.
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                              true  "Financial Report ID"
// @Param        report  body      domain.UpdateFinancialReportRequest  true  "Financial Report data"
// @Success      200     {object}  domain.FinancialReportModel
// @Failure      400     {object}  domain.ErrorResponse
// @Failure      401     {object}  domain.ErrorResponse
// @Failure      403     {object}  domain.ErrorResponse
// @Failure      404     {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/{id} [put]
func (h *FinancialReportHandler) UpdateFinancialReport(c *fiber.Ctx) error {
	id := c.Params("id")
	var req domain.UpdateFinancialReportRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	// Get user info from JWT
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Get IP and User Agent
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "")

	// Get existing report to check authorization
	existingReport, err := h.financialReportUseCase.GetFinancialReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Financial report not found",
		})
	}

	// Authorization: Check if user can update this report
	if !utils.IsSuperAdminLike(roleName) {
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// User can only update report for their own company or descendants
		if existingReport.CompanyID != userCompanyID {
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, existingReport.CompanyID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only update financial report for your company or its descendants",
				})
			}
		}
	}

	report, err := h.financialReportUseCase.UpdateFinancialReport(id, &req, userID, username, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetFinancialReportByID handles getting a financial report by ID
// @Summary      Ambil Financial Report by ID
// @Description  Mengambil financial report berdasarkan ID
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Financial Report ID"
// @Success      200 {object}  domain.FinancialReportModel
// @Failure      401 {object}  domain.ErrorResponse
// @Failure      404 {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/{id} [get]
func (h *FinancialReportHandler) GetFinancialReportByID(c *fiber.Ctx) error {
	id := c.Params("id")

	report, err := h.financialReportUseCase.GetFinancialReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Financial report not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(report)
}

// GetFinancialReportsByCompanyID handles getting all financial reports for a company
// @Summary      Ambil Semua Financial Reports untuk Company
// @Description  Mengambil semua financial reports (RKAP dan Realisasi) untuk perusahaan tertentu
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company_id  path      string  true  "Company ID"
// @Success      200         {array}   domain.FinancialReportModel
// @Failure      401         {object}  domain.ErrorResponse
// @Failure      403         {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/company/{company_id} [get]
func (h *FinancialReportHandler) GetFinancialReportsByCompanyID(c *fiber.Ctx) error {
	companyID := c.Params("company_id")

	reports, err := h.financialReportUseCase.GetFinancialReportsByCompanyID(companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get financial reports: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(reports)
}

// GetComparison handles getting comparison between RKAP and Realisasi YTD
// @Summary      Ambil Perbandingan RKAP vs Realisasi YTD
// @Description  Mengambil perbandingan antara RKAP tahunan dan Realisasi YTD sampai bulan tertentu
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company_id  query     string  true  "Company ID"
// @Param        year        query     string  true  "Year (format: YYYY)"
// @Param        month       query     string  true  "Month (format: MM, 01-12)"
// @Success      200         {object}  domain.FinancialReportComparisonResponse
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      401         {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/compare [get]
func (h *FinancialReportHandler) GetComparison(c *fiber.Ctx) error {
	companyID := c.Query("company_id")
	year := c.Query("year")
	month := c.Query("month")

	if companyID == "" || year == "" || month == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "company_id, year, and month are required",
		})
	}

	comparison, err := h.financialReportUseCase.GetComparison(companyID, year, month)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "comparison_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(comparison)
}

// GetRKAPYearsByCompanyID handles getting list of years that have RKAP for a company
// @Summary      Ambil Daftar Tahun yang Sudah Ada RKAP
// @Description  Mengambil daftar tahun yang sudah ada RKAP untuk perusahaan tertentu
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company_id  path      string  true  "Company ID"
// @Success      200         {array}   string
// @Failure      401         {object}  domain.ErrorResponse
// @Failure      403         {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/rkap-years/{company_id} [get]
func (h *FinancialReportHandler) GetRKAPYearsByCompanyID(c *fiber.Ctx) error {
	companyID := c.Params("company_id")

	// Get user info from JWT
	roleName := c.Locals("roleName").(string)
	companyIDFromJWT := c.Locals("companyID")

	// Authorization: Check if user can access this company's data
	if !utils.IsSuperAdminLike(roleName) {
		if companyIDFromJWT == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		var userCompanyID string
		if companyIDPtr, ok := companyIDFromJWT.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyIDFromJWT.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// User can only access their own company or descendants
		if companyID != userCompanyID {
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, companyID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only access financial report data for your company or its descendants",
				})
			}
		}
	}

	years, err := h.financialReportUseCase.GetRKAPYearsByCompanyID(companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get RKAP years: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(years)
}

// DeleteFinancialReport handles financial report deletion
// @Summary      Hapus Financial Report
// @Description  Menghapus financial report. Hanya bisa dihapus oleh user yang memiliki akses ke company tersebut.
// @Tags         Financial Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Financial Report ID"
// @Success      200 {object}  map[string]string
// @Failure      401 {object}  domain.ErrorResponse
// @Failure      403 {object}  domain.ErrorResponse
// @Failure      404 {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/{id} [delete]
func (h *FinancialReportHandler) DeleteFinancialReport(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get user info from JWT
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Get IP and User Agent
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "")

	// Get existing report to check authorization
	existingReport, err := h.financialReportUseCase.GetFinancialReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Financial report not found",
		})
	}

	// Authorization: Check if user can delete this report
	if !utils.IsSuperAdminLike(roleName) {
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// User can only delete report for their own company or descendants
		if existingReport.CompanyID != userCompanyID {
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, existingReport.CompanyID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only delete financial report for your company or its descendants",
				})
			}
		}
	}

	if err := h.financialReportUseCase.DeleteFinancialReport(id, userID, username, ipAddress, userAgent); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Financial report deleted successfully",
	})
}

// ExportPerformanceExcel handles exporting performance data to Excel
// @Summary      Export Performance Data to Excel
// @Description  Export performance data (Balance Sheet, Profit & Loss, Cashflow, Ratio) dengan chart untuk periode tertentu
// @Tags         Financial Reports
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        company_id    path      string  true   "Company ID"
// @Param        start_period  query     string  true   "Start period (YYYY-MM)"
// @Param        end_period    query     string  true   "End period (YYYY-MM)"
// @Success      200           {file}    application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Failure      400           {object}  domain.ErrorResponse
// @Failure      401           {object}  domain.ErrorResponse
// @Failure      403           {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{company_id}/performance/export/excel [get]
func (h *FinancialReportHandler) ExportPerformanceExcel(c *fiber.Ctx) error {
	companyID := c.Params("company_id")
	startPeriod := c.Query("start_period")
	endPeriod := c.Query("end_period")

	if companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Company ID is required",
		})
	}

	if startPeriod == "" || endPeriod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Start period and end period are required (format: YYYY-MM)",
		})
	}

	// Get user info for authorization check
	roleName := c.Locals("roleName").(string)
	userCompanyID := c.Locals("companyID")

	// Authorization check
	if !utils.IsSuperAdminLike(roleName) {
		if userCompanyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Unauthorized access",
			})
		}

		var userCompanyIDStr string
		if companyIDPtr, ok := userCompanyID.(*string); ok && companyIDPtr != nil {
			userCompanyIDStr = *companyIDPtr
		} else if companyIDStr, ok := userCompanyID.(string); ok {
			userCompanyIDStr = companyIDStr
		}

		if userCompanyIDStr != companyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You can only export performance data for your own company",
			})
		}
	}

	// Get company info for filename
	company, err := h.companyUseCase.GetCompanyByID(companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Company not found",
		})
	}

	// Generate Excel
	excelData, err := h.financialReportUseCase.ExportPerformanceExcel(companyID, startPeriod, endPeriod)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
	}

	// Generate filename
	filename := fmt.Sprintf("Performance_%s_%s_%s.xlsx",
		strings.ReplaceAll(company.Name, " ", "_"),
		startPeriod,
		endPeriod,
	)

	// Set headers and send file
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(excelData)
}

// GenerateBulkUploadTemplate handles downloading Excel template for financial report bulk upload
// @Summary      Download Financial Report Bulk Upload Template
// @Description  Download template Excel file untuk upload financial reports dalam jumlah banyak. Template berisi semua perusahaan yang dapat diakses user dengan kolom-kolom yang diperlukan.
// @Tags         Financial Reports
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        period   query     string  false  "Period untuk template (YYYY-MM), default: current month"
// @Param        is_rkap  query     bool    false  "Apakah template untuk RKAP (true) atau Realisasi (false), default: false"
// @Success      200      {file}    application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Failure      500      {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/bulk-upload/template [get]
func (h *FinancialReportHandler) GenerateBulkUploadTemplate(c *fiber.Ctx) error {
	// Use current year for template
	year := time.Now().Format("2006")
	
	// Get user info to determine accessible companies
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	var accessibleCompanies []domain.CompanyModel
	var err error

	// Get accessible companies based on user role
	if roleName == "superadmin" || roleName == "administrator" {
		// Superadmin/Administrator can access all companies
		accessibleCompanies, err = h.companyUseCase.GetAllCompanies()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: "Failed to get companies",
			})
		}
	} else {
		// Non-superadmin: get their company and all descendants
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// Get user's company
		userCompany, err := h.companyUseCase.GetCompanyByID(userCompanyID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: "Failed to get user company",
			})
		}

		// Get all descendants (including user's company)
		descendants, err := h.companyUseCase.GetCompanyDescendants(userCompanyID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: "Failed to get company descendants",
			})
		}

		// Include user's company itself
		accessibleCompanies = append([]domain.CompanyModel{*userCompany}, descendants...)
	}

	// Filter only active companies
	var activeCompanies []domain.CompanyModel
	for _, company := range accessibleCompanies {
		if company.IsActive {
			activeCompanies = append(activeCompanies, company)
		}
	}

	if len(activeCompanies) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "template_failed",
			Message: "No accessible companies found",
		})
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheetName := "Template"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "template_failed",
			Message: "Failed to create Excel sheet",
		})
	}
	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// Log error but continue (Sheet1 might not exist)
		fmt.Printf("Warning: Failed to delete Sheet1: %v\n", err)
	}

	// Set headers - semua field Financial Report (Is RKAP dihapus karena default false untuk bulk upload realisasi bulanan)
	headers := []string{
		"Kode Perusahaan",
		"Nama Perusahaan",
		"Tahun",
		"Bulan",
		// Neraca
		"Aset Lancar",
		"Aset Tidak Lancar",
		"Liabilitas Jangka Pendek",
		"Liabilitas Jangka Panjang",
		"Ekuitas",
		// Laba Rugi
		"Pendapatan",
		"Beban Usaha",
		"Laba Usaha",
		"Pendapatan Lain-Lain",
		"Pajak",
		"Laba Bersih",
		// Cashflow
		"Arus Kas Operasi",
		"Arus Kas Investasi",
		"Arus Kas Pendanaan",
		"Saldo Akhir",
		// Rasio
		"ROE (%)",
		"ROI (%)",
		"Rasio Lancar (%)",
		"Rasio Kas (%)",
		"EBITDA",
		"EBITDA Margin (%)",
		"Net Profit Margin (%)",
		"Operating Profit Margin (%)",
		"Debt to Equity",
		// Metadata
		"Keterangan",
	}

	// Set header row with styling
	for i, header := range headers {
		colName := columnIndexToName(i)
		cell := fmt.Sprintf("%s1", colName)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		style, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to create style: %v", err),
			})
		}
		if err := f.SetCellStyle(sheetName, cell, cell, style); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to set cell style: %v", err),
			})
		}
	}

	// Add data rows: setiap company dengan 12 bulan (1-12)
	row := 2
	for _, company := range activeCompanies {
		// Generate 12 rows per company (satu per bulan)
		for bulan := 1; bulan <= 12; bulan++ {
			col := 0
			
			// Company Code
			if err := setCellValue(f, sheetName, row, col, company.Code); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
			
			// Company Name
			if err := setCellValue(f, sheetName, row, col, company.Name); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
			
			// Tahun (current year)
			if err := setCellValue(f, sheetName, row, col, year); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
			
			// Bulan (1-12)
			if err := setCellValue(f, sheetName, row, col, bulan); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
		
		// Is RKAP tidak perlu di template (default false untuk bulk upload realisasi bulanan)
		// Neraca fields (leave empty for user to fill)
		for j := 0; j < 5; j++ {
			if err := setCellValue(f, sheetName, row, col, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
		}
		
		// Laba Rugi fields (leave empty for user to fill)
		for j := 0; j < 6; j++ {
			if err := setCellValue(f, sheetName, row, col, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
		}
		
		// Cashflow fields (leave empty for user to fill)
		for j := 0; j < 4; j++ {
			if err := setCellValue(f, sheetName, row, col, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
		}
		
		// Rasio fields (leave empty for user to fill)
		for j := 0; j < 10; j++ {
			if err := setCellValue(f, sheetName, row, col, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			col++
		}
		
			// Remark (optional)
			if err := setCellValue(f, sheetName, row, col, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
			
			row++ // Move to next row
		}
	}

	// Set column widths (Is RKAP dihapus, kolom bergeser)
	columnWidths := map[string]float64{
		"A": 15, // Company Code
		"B": 30, // Company Name
		"C": 10, // Tahun
		"D": 10, // Bulan
		// Neraca
		"E": 18, "F": 20, "G": 22, "H": 21, "I": 15,
		// Laba Rugi
		"J": 15, "K": 20, "L": 18, "M": 15, "N": 10, "O": 15,
		// Cashflow
		"P": 20, "Q": 20, "R": 20, "S": 15,
		// Rasio
		"T": 12, "U": 12, "V": 15, "W": 12, "X": 15,
		"Y": 18, "Z": 20, "AA": 25, "AB": 15,
		// Remark
		"AC": 30,
	}
	for col, width := range columnWidths {
		if err := f.SetColWidth(sheetName, col, col, width); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to set column width: %v", err),
			})
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "template_failed",
			Message: "Failed to write Excel file",
		})
	}

	// Set response headers
	filename := fmt.Sprintf("financial_report_template_%s.xlsx", year)
	
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

// Helper function to convert column index (0-based) to Excel column name (A, B, ..., Z, AA, AB, ...)
func columnIndexToName(colIndex int) string {
	result := ""
	colIndex++ // Convert to 1-based
	for colIndex > 0 {
		colIndex--
		result = string(rune('A'+colIndex%26)) + result
		colIndex /= 26
	}
	return result
}

// Helper function to set cell value safely
func setCellValue(f *excelize.File, sheetName string, row, col int, value interface{}) error {
	colName := columnIndexToName(col)
	cell := fmt.Sprintf("%s%d", colName, row)
	return f.SetCellValue(sheetName, cell, value)
}

// ValidateBulkExcelFile validates an Excel file before bulk upload
// @Summary      Validate Financial Report Bulk Upload Excel File
// @Description  Validates Excel file format and data before bulk upload. Returns validation errors and parsed data.
// @Tags         Financial Reports
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel file (.xlsx, .xls)"
// @Success      200   {object}  map[string]interface{}  "Response dengan valid, errors, dan data"
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      500   {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/bulk-upload/validate [post]
func (h *FinancialReportHandler) ValidateBulkExcelFile(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "File tidak ditemukan dalam request",
		})
	}

	// Validate file extension
	filename := file.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") && !strings.HasSuffix(strings.ToLower(filename), ".xls") {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_file_format",
			Message: "Format file tidak valid. Hanya file Excel (.xlsx, .xls) yang diperbolehkan",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_read_error",
			Message: "Gagal membaca file",
		})
	}
	defer src.Close()

	// Read file into buffer
	fileData := make([]byte, file.Size)
	if _, err := src.Read(fileData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_read_error",
			Message: "Gagal membaca file",
		})
	}

	// Open Excel file
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel tidak valid atau corrupt",
		})
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Get user info for company access validation
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Get accessible company codes
	var accessibleCompanyCodes map[string]bool
	if roleName == "superadmin" || roleName == "administrator" {
		companies, err := h.companyUseCase.GetAllCompanies()
		if err == nil {
			accessibleCompanyCodes = make(map[string]bool)
			for _, company := range companies {
				if company.IsActive {
					accessibleCompanyCodes[company.Code] = true
				}
			}
		}
	} else if companyID != nil {
		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		}

		if userCompanyID != "" {
			userCompany, err := h.companyUseCase.GetCompanyByID(userCompanyID)
			if err == nil {
				descendants, err := h.companyUseCase.GetCompanyDescendants(userCompanyID)
				if err == nil {
					accessibleCompanyCodes = make(map[string]bool)
					accessibleCompanyCodes[userCompany.Code] = true
					for _, company := range descendants {
						if company.IsActive {
							accessibleCompanyCodes[company.Code] = true
						}
					}
				}
			}
		}
	}

	// Get first sheet name
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel tidak memiliki sheet",
		})
	}

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "Gagal membaca data dari Excel",
		})
	}

	if len(rows) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel harus memiliki minimal header dan 1 baris data",
		})
	}

	// Expected headers (sesuai dengan template baru: Kode Perusahaan, Nama Perusahaan, Tahun, Bulan, kemudian data keuangan)
	expectedHeaders := []string{
		"Kode Perusahaan", "Nama Perusahaan", "Tahun", "Bulan",
		// Neraca
		"Aset Lancar", "Aset Tidak Lancar", "Liabilitas Jangka Pendek", "Liabilitas Jangka Panjang", "Ekuitas",
		// Laba Rugi
		"Pendapatan", "Beban Usaha", "Laba Usaha", "Pendapatan Lain-Lain", "Pajak", "Laba Bersih",
		// Cashflow
		"Arus Kas Operasi", "Arus Kas Investasi", "Arus Kas Pendanaan", "Saldo Akhir",
		// Rasio
		"ROE (%)", "ROI (%)", "Rasio Lancar (%)", "Rasio Kas (%)", "EBITDA", "EBITDA Margin (%)",
		"Net Profit Margin (%)", "Operating Profit Margin (%)", "Debt to Equity",
		// Metadata
		"Remark",
	}

	// Validate headers (optional check - just warn if mismatch)
	headers := rows[0]
	if len(headers) < len(expectedHeaders) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: fmt.Sprintf("Header tidak lengkap. Diperlukan minimal %d kolom, ditemukan %d", len(expectedHeaders), len(headers)),
		})
	}

	// Parse and validate data rows
	var errors []map[string]interface{}
	var data []map[string]interface{}

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		rowNum := rowIndex + 1 // Excel row number (1-based)

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		rowData := make(map[string]interface{})
		rowErrors := []map[string]interface{}{}

		colIndex := 0

		// Company Code (required)
		var companyCode string
		var companyIDValue string
		if len(row) > colIndex {
			companyCode = strings.TrimSpace(row[colIndex])
			if companyCode == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Kode Perusahaan",
					"message": "Kode Perusahaan wajib diisi",
				})
			} else {
				if accessibleCompanyCodes != nil && !accessibleCompanyCodes[companyCode] {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Kode Perusahaan",
						"message": fmt.Sprintf("Kode Perusahaan '%s' tidak ditemukan atau tidak dapat diakses", companyCode),
					})
				} else {
					company, err := h.companyUseCase.GetCompanyByCode(companyCode)
					if err != nil || company == nil || !company.IsActive {
						rowErrors = append(rowErrors, map[string]interface{}{
							"row":     rowNum,
							"column":  "Kode Perusahaan",
							"message": fmt.Sprintf("Kode Perusahaan '%s' tidak ditemukan", companyCode),
						})
					} else {
						companyIDValue = company.ID
						rowData["company_id"] = companyIDValue
					}
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Kode Perusahaan",
				"message": "Kode Perusahaan wajib diisi",
			})
		}
		colIndex++

		// Skip Company Name (colIndex 1) - it's just for reference
		colIndex++

		// Helper function untuk convert cell value ke string (handle berbagai tipe)
		cellToString := func(cellValue interface{}) string {
			if cellValue == nil {
				return ""
			}
			switch v := cellValue.(type) {
			case string:
				return v
			case float64:
				// Excel number sebagai float
				if v == float64(int64(v)) {
					return fmt.Sprintf("%.0f", v) // Remove decimal if whole number
				}
				return fmt.Sprintf("%g", v)
			case int:
				return fmt.Sprintf("%d", v)
			case int64:
				return fmt.Sprintf("%d", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}

		// Helper function untuk parse tahun dengan fleksibel (handle berbagai format Excel)
		parseYear := func(value string) (string, error) {
			value = strings.TrimSpace(value)
			if value == "" {
				return "", fmt.Errorf("tahun tidak boleh kosong")
			}

			// Try parse as float first (Excel might return "2025.0" or "2025")
			if yearFloat, err := strconv.ParseFloat(value, 64); err == nil {
				yearInt := int(yearFloat)
				if yearFloat != float64(yearInt) {
					return "", fmt.Errorf("tahun harus berupa bilangan bulat, nilai: '%s'", value)
				}
				yearStr := fmt.Sprintf("%04d", yearInt)
				if yearInt < 1900 || yearInt > 2100 {
					return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
				}
				return yearStr, nil
			}

			// Try parse as int directly
			if yearInt, err := strconv.Atoi(value); err == nil {
				yearStr := fmt.Sprintf("%04d", yearInt)
				if yearInt < 1900 || yearInt > 2100 {
					return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
				}
				return yearStr, nil
			}

			// If it's already a string, validate it's 4 digits
			if len(value) == 4 {
				if yearInt, err := strconv.Atoi(value); err == nil {
					if yearInt < 1900 || yearInt > 2100 {
						return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
					}
					return value, nil
				}
			}

			return "", fmt.Errorf("tahun harus berupa 4 digit (format YYYY), nilai yang diterima: '%s'", value)
		}

		// Tahun (required, format YYYY)
		var year string
		if len(row) > colIndex {
			rawValue := row[colIndex]
			yearStr := cellToString(rawValue) // Convert to string first
			parsedYear, err := parseYear(yearStr)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Tahun",
					"message": err.Error(),
				})
			} else {
				year = parsedYear
				rowData["year"] = year
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Tahun",
				"message": "Tahun wajib diisi dengan format YYYY",
			})
		}
		colIndex++

		// Helper function untuk parse bulan dengan fleksibel (handle berbagai format Excel)
		parseBulan := func(value string) (int, error) {
			value = strings.TrimSpace(value)
			if value == "" {
				return 0, fmt.Errorf("bulan tidak boleh kosong")
			}

			// Try parse as float first (Excel might return "1.0" or "1")
			if bulanFloat, err := strconv.ParseFloat(value, 64); err == nil {
				bulanInt := int(bulanFloat)
				if bulanFloat != float64(bulanInt) {
					return 0, fmt.Errorf("bulan harus berupa bilangan bulat (1-12), nilai: '%s'", value)
				}
				if bulanInt < 1 || bulanInt > 12 {
					return 0, fmt.Errorf("bulan harus antara 1-12, nilai: '%s'", value)
				}
				return bulanInt, nil
			}

			// Try parse as int directly
			if bulanInt, err := strconv.Atoi(value); err == nil {
				if bulanInt < 1 || bulanInt > 12 {
					return 0, fmt.Errorf("bulan harus antara 1-12, nilai: '%s'", value)
				}
				return bulanInt, nil
			}

			return 0, fmt.Errorf("bulan harus berupa angka (1-12), nilai yang diterima: '%s'", value)
		}

		// Bulan (required, 1-12)
		var bulan int
		var period string
		
		if len(row) > colIndex {
			rawValue := row[colIndex]
			bulanStr := cellToString(rawValue) // Convert to string first
			parsedBulan, err := parseBulan(bulanStr)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Bulan",
					"message": err.Error(),
				})
			} else {
				bulan = parsedBulan
				if year != "" {
					period = fmt.Sprintf("%s-%02d", year, bulan)
					rowData["period"] = period
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Bulan",
				"message": "Bulan wajib diisi (1-12)",
			})
		}
		colIndex++

		// Is RKAP tidak ada di template (default false untuk bulk upload realisasi bulanan)
		isRKAP := false
		rowData["is_rkap"] = isRKAP

		// Validate numeric fields - Helper function untuk parse int64
		parseInt64Field := func(value string, fieldName string, required bool, allowNegative bool) (int64, bool) {
			if value == "" {
				if required {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  fieldName,
						"message": fmt.Sprintf("%s wajib diisi", fieldName),
					})
					return 0, false
				}
				return 0, true // Optional field, empty is OK
			}
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s harus berupa angka", fieldName),
				})
				return 0, false
			}
			if !allowNegative && parsed < 0 {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s tidak boleh negatif", fieldName),
				})
				return 0, false
			}
			return int64(parsed), true
		}

		// Validate float64 fields
		// MaxDecimal10_2: 99999999.99 (untuk decimal(10,2))
		const MaxDecimal10_2 = 99999999.99
		parseFloat64Field := func(value string, fieldName string, required bool, allowNegative bool, isPercentage bool) (float64, bool) {
			if value == "" {
				if required {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  fieldName,
						"message": fmt.Sprintf("%s wajib diisi", fieldName),
					})
					return 0, false
				}
				return 0, true // Optional field, empty is OK
			}
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s harus berupa angka", fieldName),
				})
				return 0, false
			}
			if !allowNegative && parsed < 0 {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s tidak boleh negatif", fieldName),
				})
				return 0, false
			}
			// Validasi untuk decimal(10,2) - maksimal 99999999.99
			if parsed > MaxDecimal10_2 {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s: nilai terlalu besar (maksimal %.2f)", fieldName, MaxDecimal10_2),
				})
				return 0, false
			}
			// Validasi untuk persentase tidak boleh > 100%
			if isPercentage && parsed > 100 {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s: nilai tidak boleh melebihi 100%%", fieldName),
				})
				return 0, false
			}
			return parsed, true
		}

		// Neraca fields (all required, but allow 0)
		neracaFields := []string{"Current Assets", "Non Current Assets", "Short Term Liabilities", "Long Term Liabilities", "Equity"}
		for _, fieldName := range neracaFields {
			if len(row) > colIndex {
				value, ok := parseInt64Field(strings.TrimSpace(row[colIndex]), fieldName, true, true)
				if ok {
					rowData[strings.ToLower(strings.ReplaceAll(fieldName, " ", "_"))] = value
				}
			} else {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s wajib diisi", fieldName),
				})
			}
			colIndex++
		}

		// Laba Rugi fields (all required, allow negative for expenses and tax)
		labaRugiFields := []struct {
			name           string
			allowNegative  bool
		}{
			{"Revenue", false},
			{"Operating Expenses", true},
			{"Operating Profit", true},
			{"Other Income", true},
			{"Tax", true},
			{"Net Profit", true},
		}
		for _, field := range labaRugiFields {
			fieldKey := strings.ToLower(strings.ReplaceAll(field.name, " ", "_"))
			if len(row) > colIndex {
				value, ok := parseInt64Field(strings.TrimSpace(row[colIndex]), field.name, true, field.allowNegative)
				if ok {
					rowData[fieldKey] = value
				}
			} else {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  field.name,
					"message": fmt.Sprintf("%s wajib diisi", field.name),
				})
			}
			colIndex++
		}

		// Cashflow fields (all required, allow negative)
		cashflowFields := []string{"Operating Cashflow", "Investing Cashflow", "Financing Cashflow", "Ending Balance"}
		for _, fieldName := range cashflowFields {
			if len(row) > colIndex {
				value, ok := parseInt64Field(strings.TrimSpace(row[colIndex]), fieldName, true, true)
				if ok {
					rowData[strings.ToLower(strings.ReplaceAll(fieldName, " ", "_"))] = value
				}
			} else {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  fieldName,
					"message": fmt.Sprintf("%s wajib diisi", fieldName),
				})
			}
			colIndex++
		}

		// Rasio fields (optional, default 0)
		rasioFields := []struct {
			name          string
			isInt64       bool
		}{
			{"ROE (%)", false},
			{"ROI (%)", false},
			{"Current Ratio", false},
			{"Cash Ratio", false},
			{"EBITDA", true},
			{"EBITDA Margin (%)", false},
			{"Net Profit Margin (%)", false},
			{"Operating Profit Margin (%)", false},
			{"Debt to Equity", false},
		}
		for _, field := range rasioFields {
			fieldKey := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(field.name, " (%)", ""), " ", "_"))
			if len(row) > colIndex {
				valueStr := strings.TrimSpace(row[colIndex])
				if valueStr != "" {
					if field.isInt64 {
						value, ok := parseInt64Field(valueStr, field.name, false, true)
						if ok {
							rowData[fieldKey] = value
						}
					} else {
						// Tentukan apakah field ini persentase
						isPercentage := strings.Contains(field.name, "ROE") || 
							strings.Contains(field.name, "ROI") || 
							strings.Contains(field.name, "Rasio Lancar") || 
							strings.Contains(field.name, "Rasio Kas") ||
							strings.Contains(field.name, "Margin") ||
							strings.Contains(field.name, "(%")
						
						value, ok := parseFloat64Field(valueStr, field.name, false, true, isPercentage)
						if ok {
							rowData[fieldKey] = value
						}
					}
				}
			}
			colIndex++
		}

		// Remark (optional)
		if len(row) > colIndex {
			remark := strings.TrimSpace(row[colIndex])
			if remark != "" {
				rowData["remark"] = remark
			}
		}

		// Add row errors to errors list
		errors = append(errors, rowErrors...)

		// Add row data only if no errors
		if len(rowErrors) == 0 {
			data = append(data, rowData)
		}
	}

	// Return validation result
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
		"data":   data,
	})
}

// UploadBulkFinancialReports handles uploading Excel file and creating financial reports in bulk
// @Summary      Upload Financial Reports from Excel (Bulk)
// @Description  Upload Excel file, validate rows, and create financial reports in database. Supports both RKAP and Realisasi.
// @Tags         Financial Reports
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel file (.xlsx, .xls)"
// @Success      200   {object}  map[string]interface{}  "Response dengan success, failed, dan errors"
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      500   {object}  domain.ErrorResponse
// @Router       /api/v1/financial-reports/bulk-upload [post]
func (h *FinancialReportHandler) UploadBulkFinancialReports(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "File tidak ditemukan dalam request",
		})
	}

	filename := strings.ToLower(file.Filename)
	if !strings.HasSuffix(filename, ".xlsx") && !strings.HasSuffix(filename, ".xls") {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_file_format",
			Message: "Format file tidak valid. Hanya file Excel (.xlsx, .xls) yang diperbolehkan",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_read_error",
			Message: "Gagal membaca file",
		})
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_read_error",
			Message: "Gagal membaca file",
		})
	}

	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel tidak valid atau corrupt",
		})
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	// Get user info
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	roleName := c.Locals("roleName").(string)
	companyID := c.Locals("companyID")

	// Determine accessible company codes
	var accessibleCompanyCodes map[string]bool
	if roleName == "superadmin" || roleName == "administrator" {
		companies, err := h.companyUseCase.GetAllCompanies()
		if err == nil {
			accessibleCompanyCodes = make(map[string]bool)
			for _, company := range companies {
				if company.IsActive {
					accessibleCompanyCodes[company.Code] = true
				}
			}
		}
	} else if companyID != nil {
		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		}

		if userCompanyID != "" {
			userCompany, err := h.companyUseCase.GetCompanyByID(userCompanyID)
			if err == nil {
				descendants, err := h.companyUseCase.GetCompanyDescendants(userCompanyID)
				if err == nil {
					accessibleCompanyCodes = make(map[string]bool)
					accessibleCompanyCodes[userCompany.Code] = true
					for _, company := range descendants {
						if company.IsActive {
							accessibleCompanyCodes[company.Code] = true
						}
					}
				}
			}
		}
	}

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel tidak memiliki sheet",
		})
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "Gagal membaca data dari Excel",
		})
	}

	if len(rows) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "File Excel harus memiliki minimal header dan 1 baris data",
		})
	}

	// Get IP and User Agent for audit log
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent", "")

	// Process rows - reuse validation logic but create reports
	errorsList := []map[string]interface{}{}
	successCount := 0
	failedCount := 0
	createdCount := 0
	updatedCount := 0

	// Helper functions for parsing (same as validation)
	// MaxInt64: 9223372036854775807 (untuk PostgreSQL bigint)
	// MaxDecimal10_2: 99999999.99 (untuk decimal(10,2))
	const MaxDecimal10_2 = 99999999.99
	
	parseInt64Field := func(value string, allowNegative bool) (int64, error) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}
		if !allowNegative && parsed < 0 {
			return 0, fmt.Errorf("nilai tidak boleh negatif")
		}
		// Validasi tidak melebihi MaxInt64
		if parsed > float64(math.MaxInt64) {
			return 0, fmt.Errorf("nilai terlalu besar (maksimal %d)", math.MaxInt64)
		}
		// Validasi tidak kurang dari MinInt64
		if parsed < float64(math.MinInt64) {
			return 0, fmt.Errorf("nilai terlalu kecil (minimal %d)", math.MinInt64)
		}
		return int64(parsed), nil
	}

	parseFloat64Field := func(value string, allowNegative bool, isPercentage bool) (float64, error) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}
		if !allowNegative && parsed < 0 {
			return 0, fmt.Errorf("nilai tidak boleh negatif")
		}
		// Validasi untuk decimal(10,2) - maksimal 99999999.99
		if parsed > MaxDecimal10_2 {
			return 0, fmt.Errorf("nilai terlalu besar (maksimal %.2f)", MaxDecimal10_2)
		}
		// Validasi untuk persentase tidak boleh > 100%
		if isPercentage && parsed > 100 {
			return 0, fmt.Errorf("nilai tidak boleh melebihi 100%%")
		}
		return parsed, nil
	}

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		rowNum := rowIndex + 1

		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		rowErrors := []map[string]interface{}{}
		colIndex := 0

		// Company Code (colIndex 0)
		var companyIDValue string
		if len(row) > colIndex {
			companyCode := strings.TrimSpace(row[colIndex])
			if companyCode == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Kode Perusahaan",
					"message": "Kode Perusahaan wajib diisi",
				})
			} else {
				if accessibleCompanyCodes != nil && !accessibleCompanyCodes[companyCode] {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Kode Perusahaan",
						"message": fmt.Sprintf("Kode Perusahaan '%s' tidak ditemukan atau tidak dapat diakses", companyCode),
					})
				} else {
					company, err := h.companyUseCase.GetCompanyByCode(companyCode)
					if err != nil || company == nil || !company.IsActive {
						rowErrors = append(rowErrors, map[string]interface{}{
							"row":     rowNum,
							"column":  "Kode Perusahaan",
							"message": fmt.Sprintf("Kode Perusahaan '%s' tidak ditemukan", companyCode),
						})
					} else {
						companyIDValue = company.ID
					}
				}
			}
		}
		colIndex++ // Move to Company Name (colIndex 1)
		
		// Skip Company Name (colIndex 1) - it's just for reference, we don't need to validate it
		colIndex++ // Move to Tahun (colIndex 2)

		// Helper function untuk convert cell value ke string (handle berbagai tipe)
		cellToString := func(cellValue interface{}) string {
			if cellValue == nil {
				return ""
			}
			switch v := cellValue.(type) {
			case string:
				return v
			case float64:
				// Excel number sebagai float
				if v == float64(int64(v)) {
					return fmt.Sprintf("%.0f", v) // Remove decimal if whole number
				}
				return fmt.Sprintf("%g", v)
			case int:
				return fmt.Sprintf("%d", v)
			case int64:
				return fmt.Sprintf("%d", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}

		// Helper function untuk parse tahun dengan fleksibel (handle berbagai format Excel)
		parseYear := func(value string) (string, error) {
			value = strings.TrimSpace(value)
			if value == "" {
				return "", fmt.Errorf("tahun tidak boleh kosong")
			}

			// Try parse as float first (Excel might return "2025.0" or "2025")
			if yearFloat, err := strconv.ParseFloat(value, 64); err == nil {
				yearInt := int(yearFloat)
				if yearFloat != float64(yearInt) {
					return "", fmt.Errorf("tahun harus berupa bilangan bulat, nilai: '%s'", value)
				}
				yearStr := fmt.Sprintf("%04d", yearInt)
				if yearInt < 1900 || yearInt > 2100 {
					return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
				}
				return yearStr, nil
			}

			// Try parse as int directly
			if yearInt, err := strconv.Atoi(value); err == nil {
				yearStr := fmt.Sprintf("%04d", yearInt)
				if yearInt < 1900 || yearInt > 2100 {
					return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
				}
				return yearStr, nil
			}

			// If it's already a string, validate it's 4 digits
			if len(value) == 4 {
				if yearInt, err := strconv.Atoi(value); err == nil {
					if yearInt < 1900 || yearInt > 2100 {
						return "", fmt.Errorf("tahun harus antara 1900-2100, nilai: '%s'", value)
					}
					return value, nil
				}
			}

			return "", fmt.Errorf("tahun harus berupa 4 digit (format YYYY), nilai yang diterima: '%s'", value)
		}

		// Tahun (required, format YYYY)
		var year string
		if len(row) > colIndex {
			rawValue := row[colIndex]
			yearStr := cellToString(rawValue) // Convert to string first
			parsedYear, err := parseYear(yearStr)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Tahun",
					"message": err.Error(),
				})
			} else {
				year = parsedYear
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Tahun",
				"message": "Tahun wajib diisi dengan format YYYY",
			})
		}
		colIndex++

		// Helper function untuk parse bulan dengan fleksibel (handle berbagai format Excel)
		parseBulan := func(value string) (int, error) {
			value = strings.TrimSpace(value)
			if value == "" {
				return 0, fmt.Errorf("bulan tidak boleh kosong")
			}

			// Try parse as float first (Excel might return "1.0" or "1")
			if bulanFloat, err := strconv.ParseFloat(value, 64); err == nil {
				bulanInt := int(bulanFloat)
				if bulanFloat != float64(bulanInt) {
					return 0, fmt.Errorf("bulan harus berupa bilangan bulat (1-12), nilai: '%s'", value)
				}
				if bulanInt < 1 || bulanInt > 12 {
					return 0, fmt.Errorf("bulan harus antara 1-12, nilai: '%s'", value)
				}
				return bulanInt, nil
			}

			// Try parse as int directly
			if bulanInt, err := strconv.Atoi(value); err == nil {
				if bulanInt < 1 || bulanInt > 12 {
					return 0, fmt.Errorf("bulan harus antara 1-12, nilai: '%s'", value)
				}
				return bulanInt, nil
			}

			return 0, fmt.Errorf("bulan harus berupa angka (1-12), nilai yang diterima: '%s'", value)
		}

		// Bulan (required, 1-12)
		var bulan int
		var period string
		bulanParsed := false
		
		if len(row) > colIndex {
			rawValue := row[colIndex]
			bulanStr := cellToString(rawValue) // Convert to string first
			parsedBulan, err := parseBulan(bulanStr)
			if err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Bulan",
					"message": err.Error(),
				})
			} else {
				bulan = parsedBulan
				if year != "" {
					period = fmt.Sprintf("%s-%02d", year, bulan)
				}
				bulanParsed = true
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Bulan",
				"message": "Bulan wajib diisi (1-12)",
			})
		}
		colIndex++

		// Is RKAP tidak ada di template (default false untuk bulk upload realisasi bulanan)
		isRKAP := false

		// Validate year and bulan are set
		if !bulanParsed || year == "" {
			// Bulan atau tahun parsing gagal, error sudah ditambahkan
			errorsList = append(errorsList, rowErrors...)
			failedCount++
			continue
		}
		
		// Generate period from year and bulan
		if period == "" {
			period = fmt.Sprintf("%s-%02d", year, bulan)
		}

		// If there are errors so far, skip this row
		if len(rowErrors) > 0 {
			errorsList = append(errorsList, rowErrors...)
			failedCount++
			continue
		}

		// Build CreateFinancialReportRequest
		req := domain.CreateFinancialReportRequest{
			CompanyID: companyIDValue,
			Year:      year,
			Period:    period,
			IsRKAP:    isRKAP,
		}

		// Parse Neraca fields
		for i, fieldName := range []string{"Current Assets", "Non Current Assets", "Short Term Liabilities", "Long Term Liabilities", "Equity"} {
			if len(row) > colIndex {
				value, err := parseInt64Field(strings.TrimSpace(row[colIndex]), true)
				if err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  fieldName,
						"message": fmt.Sprintf("%s: %v", fieldName, err),
					})
				} else {
					switch i {
					case 0:
						req.CurrentAssets = value
					case 1:
						req.NonCurrentAssets = value
					case 2:
						req.ShortTermLiabilities = value
					case 3:
						req.LongTermLiabilities = value
					case 4:
						req.Equity = value
					}
				}
			}
			colIndex++
		}

		// Parse Laba Rugi fields (headers sudah dalam Bahasa Indonesia)
		labaRugiFields := []struct {
			name          string
			allowNegative bool
			target        *int64
		}{
			{"Pendapatan", false, &req.Revenue},
			{"Beban Usaha", true, &req.OperatingExpenses},
			{"Laba Usaha", true, &req.OperatingProfit},
			{"Pendapatan Lain-Lain", true, &req.OtherIncome},
			{"Pajak", true, &req.Tax},
			{"Laba Bersih", true, &req.NetProfit},
		}
		for _, field := range labaRugiFields {
			if len(row) > colIndex {
				value, err := parseInt64Field(strings.TrimSpace(row[colIndex]), field.allowNegative)
				if err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  field.name,
						"message": fmt.Sprintf("%s: %v", field.name, err),
					})
				} else {
					*field.target = value
				}
			}
			colIndex++
		}

		// Parse Cashflow fields (headers sudah dalam Bahasa Indonesia)
		for i, fieldName := range []string{"Arus Kas Operasi", "Arus Kas Investasi", "Arus Kas Pendanaan", "Saldo Akhir"} {
			if len(row) > colIndex {
				value, err := parseInt64Field(strings.TrimSpace(row[colIndex]), true)
				if err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  fieldName,
						"message": fmt.Sprintf("%s: %v", fieldName, err),
					})
				} else {
					switch i {
					case 0:
						req.OperatingCashflow = value
					case 1:
						req.InvestingCashflow = value
					case 2:
						req.FinancingCashflow = value
					case 3:
						req.EndingBalance = value
					}
				}
			}
			colIndex++
		}

		// Parse Rasio fields (headers sudah dalam Bahasa Indonesia)
		rasioFields := []struct {
			name          string
			isInt64       bool
			targetInt64   *int64
			targetFloat64 *float64
		}{
			{"ROE (%)", false, nil, &req.ROE},
			{"ROI (%)", false, nil, &req.ROI},
			{"Rasio Lancar (%)", false, nil, &req.CurrentRatio},
			{"Rasio Kas (%)", false, nil, &req.CashRatio},
			{"EBITDA", true, &req.EBITDA, nil},
			{"EBITDA Margin (%)", false, nil, &req.EBITDAMargin},
			{"Net Profit Margin (%)", false, nil, &req.NetProfitMargin},
			{"Operating Profit Margin (%)", false, nil, &req.OperatingProfitMargin},
			{"Debt to Equity", false, nil, &req.DebtToEquity},
		}
		for _, field := range rasioFields {
			if len(row) > colIndex {
				valueStr := strings.TrimSpace(row[colIndex])
				if valueStr != "" {
					if field.isInt64 {
						value, err := parseInt64Field(valueStr, true)
						if err != nil {
							rowErrors = append(rowErrors, map[string]interface{}{
								"row":     rowNum,
								"column":  field.name,
								"message": fmt.Sprintf("%s: %v", field.name, err),
							})
						} else {
							*field.targetInt64 = value
						}
					} else {
						// Tentukan apakah field ini persentase
						isPercentage := strings.Contains(field.name, "ROE") || 
							strings.Contains(field.name, "ROI") || 
							strings.Contains(field.name, "Rasio Lancar") || 
							strings.Contains(field.name, "Rasio Kas") ||
							strings.Contains(field.name, "Margin") ||
							strings.Contains(field.name, "(%")
						
						value, err := parseFloat64Field(valueStr, true, isPercentage)
						if err != nil {
							rowErrors = append(rowErrors, map[string]interface{}{
								"row":     rowNum,
								"column":  field.name,
								"message": fmt.Sprintf("%s: %v", field.name, err),
							})
						} else {
							*field.targetFloat64 = value
						}
					}
				}
			}
			colIndex++
		}

		// Remark (optional)
		if len(row) > colIndex {
			remark := strings.TrimSpace(row[colIndex])
			if remark != "" {
				req.Remark = &remark
			}
		}

		// If there are parsing errors, skip this row
		if len(rowErrors) > 0 {
			errorsList = append(errorsList, rowErrors...)
			failedCount++
			continue
		}

		// Check if financial report already exists (upsert logic)
		var existingReport *domain.FinancialReportModel
		var err error
		
		if isRKAP {
			// For RKAP, check by company_id, year, and is_rkap
			existingReport, err = h.financialReportUseCase.GetRKAPByCompanyIDAndYear(companyIDValue, year)
			if err != nil && err != gorm.ErrRecordNotFound {
				errorsList = append(errorsList, map[string]interface{}{
					"row":     rowNum,
					"column":  "general",
					"message": fmt.Sprintf("Error checking existing RKAP: %v", err),
				})
				failedCount++
				continue
			}
		} else {
			// For Realisasi, check by company_id, period, and is_rkap
			existingReport, err = h.financialReportUseCase.GetRealisasiByCompanyIDAndPeriod(companyIDValue, period)
			if err != nil && err != gorm.ErrRecordNotFound {
				errorsList = append(errorsList, map[string]interface{}{
					"row":     rowNum,
					"column":  "general",
					"message": fmt.Sprintf("Error checking existing Realisasi: %v", err),
				})
				failedCount++
				continue
			}
		}

		// Upsert: Update if exists, Create if not
		if existingReport != nil {
			// Convert CreateFinancialReportRequest to UpdateFinancialReportRequest
			updateReq := domain.UpdateFinancialReportRequest{
				Year:   &year,
				Period: &period,
				IsRKAP: &isRKAP,
				// Neraca
				CurrentAssets:        &req.CurrentAssets,
				NonCurrentAssets:     &req.NonCurrentAssets,
				ShortTermLiabilities: &req.ShortTermLiabilities,
				LongTermLiabilities:  &req.LongTermLiabilities,
				Equity:               &req.Equity,
				// Laba Rugi
				Revenue:           &req.Revenue,
				OperatingExpenses: &req.OperatingExpenses,
				OperatingProfit:   &req.OperatingProfit,
				OtherIncome:       &req.OtherIncome,
				Tax:               &req.Tax,
				NetProfit:         &req.NetProfit,
				// Cashflow
				OperatingCashflow: &req.OperatingCashflow,
				InvestingCashflow: &req.InvestingCashflow,
				FinancingCashflow: &req.FinancingCashflow,
				EndingBalance:     &req.EndingBalance,
				// Rasio
				ROE:                   &req.ROE,
				ROI:                   &req.ROI,
				CurrentRatio:          &req.CurrentRatio,
				CashRatio:             &req.CashRatio,
				EBITDA:                &req.EBITDA,
				EBITDAMargin:          &req.EBITDAMargin,
				NetProfitMargin:       &req.NetProfitMargin,
				OperatingProfitMargin: &req.OperatingProfitMargin,
				DebtToEquity:          &req.DebtToEquity,
				Remark:                req.Remark,
			}

			// Update existing report
			_, err = h.financialReportUseCase.UpdateFinancialReport(existingReport.ID, &updateReq, userID, username, ipAddress, userAgent)
			if err != nil {
				// Parse error untuk menentukan kolom yang bermasalah
				errMsg := err.Error()
				column := "general"
				
				// Cek apakah error terkait validasi rasio > 100%
				if strings.Contains(errMsg, "rasio keuangan tidak boleh melebihi 100%") {
					// Tentukan kolom mana yang bermasalah
					if updateReq.ROE != nil && *updateReq.ROE > 100 {
						column = "ROE (%)"
					} else if updateReq.ROI != nil && *updateReq.ROI > 100 {
						column = "ROI (%)"
					} else if updateReq.CurrentRatio != nil && *updateReq.CurrentRatio > 100 {
						column = "Rasio Lancar (%)"
					} else if updateReq.CashRatio != nil && *updateReq.CashRatio > 100 {
						column = "Rasio Kas (%)"
					} else if updateReq.EBITDAMargin != nil && *updateReq.EBITDAMargin > 100 {
						column = "EBITDA Margin (%)"
					} else if updateReq.NetProfitMargin != nil && *updateReq.NetProfitMargin > 100 {
						column = "Net Profit Margin (%)"
					} else if updateReq.OperatingProfitMargin != nil && *updateReq.OperatingProfitMargin > 100 {
						column = "Operating Profit Margin (%)"
					}
					errMsg = fmt.Sprintf("%s: nilai tidak boleh melebihi 100%%", column)
				} else if strings.Contains(errMsg, "numeric field overflow") || strings.Contains(errMsg, "SQLSTATE 22003") {
					// Error overflow dari database - cari field mana yang bermasalah
					column = "Data numerik terlalu besar"
					errMsg = "Nilai terlalu besar untuk disimpan. Pastikan: int64 tidak melebihi 9,223,372,036,854,775,807 dan persentase/rasio tidak melebihi 99,999,999.99"
				}
				
				errorsList = append(errorsList, map[string]interface{}{
					"row":     rowNum,
					"column":  column,
					"message": errMsg,
				})
				failedCount++
				continue
			}
			updatedCount++
		} else {
			// Create new report
			_, err = h.financialReportUseCase.CreateFinancialReport(&req, userID, username, ipAddress, userAgent)
			if err != nil {
				// Parse error untuk menentukan kolom yang bermasalah
				errMsg := err.Error()
				column := "general"
				
				// Cek apakah error terkait validasi rasio > 100%
				if strings.Contains(errMsg, "rasio keuangan tidak boleh melebihi 100%") {
					// Tentukan kolom mana yang bermasalah
					if req.ROE > 100 {
						column = "ROE (%)"
					} else if req.ROI > 100 {
						column = "ROI (%)"
					} else if req.CurrentRatio > 100 {
						column = "Rasio Lancar (%)"
					} else if req.CashRatio > 100 {
						column = "Rasio Kas (%)"
					} else if req.EBITDAMargin > 100 {
						column = "EBITDA Margin (%)"
					} else if req.NetProfitMargin > 100 {
						column = "Net Profit Margin (%)"
					} else if req.OperatingProfitMargin > 100 {
						column = "Operating Profit Margin (%)"
					}
					errMsg = fmt.Sprintf("%s: nilai tidak boleh melebihi 100%%", column)
				}
				
				errorsList = append(errorsList, map[string]interface{}{
					"row":     rowNum,
					"column":  column,
					"message": errMsg,
				})
				failedCount++
				continue
			}
			createdCount++
		}

		successCount++
	}

	// Log hasil upload untuk debugging
	zapLog := logger.GetLogger()
	zapLog.Info("Bulk upload completed",
		zap.Int("success", successCount),
		zap.Int("failed", failedCount),
		zap.Int("created", createdCount),
		zap.Int("updated", updatedCount),
		zap.Int("total_errors", len(errorsList)),
	)
	if len(errorsList) > 0 {
		zapLog.Warn("Upload errors found", zap.Any("errors", errorsList))
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"success":      successCount,
		"failed":       failedCount,
		"created":      createdCount,
		"updated":      updatedCount,
		"errors":       errorsList,
		"message":      fmt.Sprintf("Upload selesai: %d berhasil (%d dibuat, %d diupdate), %d gagal", successCount, createdCount, updatedCount, failedCount),
	})
}
