package http

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
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
