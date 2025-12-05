package http

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	reportUseCase  usecase.ReportUseCase
	companyUseCase usecase.CompanyUseCase
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportUseCase usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{
		reportUseCase:  reportUseCase,
		companyUseCase: usecase.NewCompanyUseCase(),
	}
}

// CreateReport handles report creation
// @Summary      Buat Report Baru
// @Description  Membuat report bulanan baru untuk perusahaan. Setiap perusahaan hanya bisa membuat report untuk perusahaan mereka sendiri.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        report  body      domain.CreateReportRequest  true  "Report data"
// @Success      201     {object}  domain.ReportModel
// @Failure      400     {object}  domain.ErrorResponse
// @Failure      401     {object}  domain.ErrorResponse
// @Failure      403     {object}  domain.ErrorResponse
// @Router       /api/v1/reports [post]
func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	var req domain.CreateReportRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get user info from JWT
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Validate access: non-superadmin/administrator can only create reports for their own company
	if !utils.IsSuperAdminLike(roleName) && companyID != nil {
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

		// Check if user has access to create report for this company
		hasAccess, err := h.reportUseCase.ValidateReportAccess(roleName, &userCompanyID, req.CompanyID)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to create reports for this company",
			})
		}
	}

	report, err := h.reportUseCase.CreateReport(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	audit.LogAction(userID, username, "create_report", audit.ResourceReport, report.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"company_id": report.CompanyID,
		"period":     report.Period,
	})

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetReport handles getting a single report by ID
// @Summary      Get Report by ID
// @Description  Mendapatkan detail report berdasarkan ID. User hanya bisa melihat report dari perusahaan mereka (atau anak perusahaan untuk admin).
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Report ID"
// @Success      200  {object}  domain.ReportModel
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/reports/{id} [get]
func (h *ReportHandler) GetReport(c *fiber.Ctx) error {
	id := c.Params("id")

	report, err := h.reportUseCase.GetReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Report not found",
		})
	}

	// Get user info for access validation
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Validate access
	hasAccess, err := h.reportUseCase.ValidateReportAccess(roleName, userCompanyID, report.CompanyID)
	if err != nil || !hasAccess {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to view this report",
		})
	}

	// Audit log (opsional untuk view)
	if audit.ShouldLogView() {
		userID := c.Locals("userID").(string)
		username := c.Locals("username").(string)
		audit.LogAction(userID, username, "view_report", audit.ResourceReport, report.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"company_id": report.CompanyID,
			"period":     report.Period,
		})
	}

	return c.JSON(report)
}

// GetAllReports handles getting all reports with RBAC
// @Summary      Get All Reports
// @Description  Mendapatkan semua reports. Superadmin melihat semua, admin melihat reports dari perusahaan mereka dan anak perusahaan, user biasa hanya melihat reports dari perusahaan mereka.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company_id  query     string  false  "Filter by company ID"
// @Param        period      query     string  false  "Filter by period (YYYY-MM)"
// @Param        page        query     int     false  "Page number (default: 1)"
// @Param        page_size   query     int     false  "Page size (default: 10)"
// @Success      200         {object}  map[string]interface{}
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      401         {object}  domain.ErrorResponse
// @Router       /api/v1/reports [get]
func (h *ReportHandler) GetAllReports(c *fiber.Ctx) error {
	// Get user info
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Get all reports based on RBAC
	reports, err := h.reportUseCase.GetAllReports(roleName, userCompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	// Apply filters
	filteredReports := reports
	companyIDFilter := c.Query("company_id")
	periodFilter := c.Query("period")

	if companyIDFilter != "" {
		var filtered []domain.ReportModel
		// Support multiple company IDs (comma-separated)
		companyIDs := strings.Split(companyIDFilter, ",")
		companyIDMap := make(map[string]bool)
		for _, id := range companyIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				companyIDMap[id] = true
			}
		}

		for _, r := range filteredReports {
			if companyIDMap[r.CompanyID] {
				filtered = append(filtered, r)
			}
		}
		filteredReports = filtered
	}

	if periodFilter != "" {
		var filtered []domain.ReportModel
		for _, r := range filteredReports {
			if r.Period == periodFilter {
				filtered = append(filtered, r)
			}
		}
		filteredReports = filtered
	}

	// Pagination
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total := len(filteredReports)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var paginatedReports []domain.ReportModel
	if start < total {
		paginatedReports = filteredReports[start:end]
	}

	// Set pagination headers
	c.Set("X-Total-Count", strconv.Itoa(total))
	c.Set("X-Page", strconv.Itoa(page))
	c.Set("X-Page-Size", strconv.Itoa(pageSize))

	// Audit log (opsional untuk view)
	if audit.ShouldLogView() {
		userID := c.Locals("userID").(string)
		username := c.Locals("username").(string)
		audit.LogAction(userID, username, "list_reports", audit.ResourceReport, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"total":          total,
			"page":           page,
			"page_size":      pageSize,
			"company_filter": companyIDFilter,
			"period_filter":  periodFilter,
		})
	}

	return c.JSON(map[string]interface{}{
		"data":        paginatedReports,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// GetReportsByCompany handles getting reports for a specific company
// @Summary      Get Reports by Company ID
// @Description  Mendapatkan semua reports untuk perusahaan tertentu. User hanya bisa melihat reports dari perusahaan mereka (atau anak perusahaan untuk admin).
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company_id  path      string  true  "Company ID"
// @Success      200         {array}   domain.ReportModel
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      401         {object}  domain.ErrorResponse
// @Failure      403         {object}  domain.ErrorResponse
// @Router       /api/v1/reports/company/{company_id} [get]
func (h *ReportHandler) GetReportsByCompany(c *fiber.Ctx) error {
	companyID := c.Params("company_id")

	// Get user info
	userCompanyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	var userCompanyIDPtr *string
	if userCompanyID != nil {
		if companyIDPtr, ok := userCompanyID.(*string); ok && companyIDPtr != nil {
			userCompanyIDPtr = companyIDPtr
		} else if companyIDStr, ok := userCompanyID.(string); ok {
			userCompanyIDPtr = &companyIDStr
		}
	}

	reports, err := h.reportUseCase.GetReportsByCompanyID(companyID, roleName, userCompanyIDPtr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "access_denied",
			Message: err.Error(),
		})
	}

	return c.JSON(reports)
}

// UpdateReport handles report update
// @Summary      Update Report
// @Description  Mengupdate report. User hanya bisa mengupdate report dari perusahaan mereka.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                  true  "Report ID"
// @Param        report  body      domain.UpdateReportRequest  true  "Report update data"
// @Success      200     {object}  domain.ReportModel
// @Failure      400     {object}  domain.ErrorResponse
// @Failure      401     {object}  domain.ErrorResponse
// @Failure      403     {object}  domain.ErrorResponse
// @Failure      404     {object}  domain.ErrorResponse
// @Router       /api/v1/reports/{id} [put]
func (h *ReportHandler) UpdateReport(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get existing report to check access
	existingReport, err := h.reportUseCase.GetReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Report not found",
		})
	}

	// Get user info for access validation
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Validate access
	hasAccess, err := h.reportUseCase.ValidateReportAccess(roleName, userCompanyID, existingReport.CompanyID)
	if err != nil || !hasAccess {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to update this report",
		})
	}

	var req domain.UpdateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// If updating company_id, validate access to new company
	if req.CompanyID != nil && *req.CompanyID != existingReport.CompanyID {
		hasAccess, err := h.reportUseCase.ValidateReportAccess(roleName, userCompanyID, *req.CompanyID)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to move report to this company",
			})
		}
	}

	report, err := h.reportUseCase.UpdateReport(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	audit.LogAction(userID, username, "update_report", audit.ResourceReport, report.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"company_id": report.CompanyID,
		"period":     report.Period,
	})

	return c.JSON(report)
}

// DeleteReport handles report deletion
// @Summary      Delete Report
// @Description  Menghapus report. User hanya bisa menghapus report dari perusahaan mereka.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Report ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/reports/{id} [delete]
func (h *ReportHandler) DeleteReport(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get existing report to check access
	existingReport, err := h.reportUseCase.GetReportByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Report not found",
		})
	}

	// Get user info for access validation
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Validate access
	hasAccess, err := h.reportUseCase.ValidateReportAccess(roleName, userCompanyID, existingReport.CompanyID)
	if err != nil || !hasAccess {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to delete this report",
		})
	}

	err = h.reportUseCase.DeleteReport(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	audit.LogAction(userID, username, "delete_report", audit.ResourceReport, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"company_id": existingReport.CompanyID,
		"period":     existingReport.Period,
	})

	return c.JSON(map[string]string{
		"message": "Report deleted successfully",
	})
}

// ExportReportsExcel handles exporting reports to Excel
// @Summary      Export Reports to Excel
// @Description  Export semua reports yang dapat diakses user ke format Excel
// @Tags         Reports
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        company_id  query     string  false  "Filter by company ID"
// @Param        period      query     string  false  "Filter by period (YYYY-MM)"
// @Success      200         {file}    application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      401         {object}  domain.ErrorResponse
// @Router       /api/v1/reports/export/excel [get]
func (h *ReportHandler) ExportReportsExcel(c *fiber.Ctx) error {
	// Get user info
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Get all reports based on RBAC
	reports, err := h.reportUseCase.GetAllReports(roleName, userCompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	// Apply filters
	companyIDFilter := c.Query("company_id")
	periodFilter := c.Query("period")

	if companyIDFilter != "" {
		var filtered []domain.ReportModel
		// Support multiple company IDs (comma-separated)
		companyIDs := strings.Split(companyIDFilter, ",")
		companyIDMap := make(map[string]bool)
		for _, id := range companyIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				companyIDMap[id] = true
			}
		}

		for _, r := range reports {
			if companyIDMap[r.CompanyID] {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	if periodFilter != "" {
		var filtered []domain.ReportModel
		for _, r := range reports {
			if r.Period == periodFilter {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	sheetName := "Reports"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to create Excel sheet",
		})
	}
	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// Log error but continue (Sheet1 might not exist)
		fmt.Printf("Warning: Failed to delete Sheet1: %v\n", err)
	}

	// Set headers
	headers := []string{"Period", "Company", "Revenue", "Opex", "NPAT", "Dividend", "Financial Ratio (%)", "Inputter", "Remark"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		style, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to create style: %v", err),
			})
		}
		if err := f.SetCellStyle(sheetName, cell, cell, style); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell style: %v", err),
			})
		}
	}

	// Add data
	for i, report := range reports {
		row := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), report.Period); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		companyName := "Unknown"
		if report.Company != nil {
			companyName = report.Company.Name
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), companyName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), report.Revenue); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), report.Opex); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), report.NPAT); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), report.Dividend); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", report.FinancialRatio)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		inputterName := "N/A"
		if report.Inputter != nil {
			inputterName = report.Inputter.Username
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), inputterName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		remark := "N/A"
		if report.Remark != nil {
			remark = *report.Remark
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), remark); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
	}

	// Auto-size columns
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		if err := f.SetColWidth(sheetName, col, col, 15); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "export_failed",
				Message: fmt.Sprintf("Failed to set column width: %v", err),
			})
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to generate Excel file",
		})
	}

	// Audit log
	audit.LogAction(userID, username, "export_reports_excel", audit.ResourceReport, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"count":          len(reports),
		"company_filter": companyIDFilter,
		"period_filter":  periodFilter,
	})

	// Set headers and send file
	filename := fmt.Sprintf("reports_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(buf.Bytes())
}

// ExportReportsPDF handles exporting reports to PDF
// @Summary      Export Reports to PDF
// @Description  Export semua reports yang dapat diakses user ke format PDF
// @Tags         Reports
// @Accept       json
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        company_id  query     string  false  "Filter by company ID"
// @Param        period      query     string  false  "Filter by period (YYYY-MM)"
// @Success      200         {file}    application/pdf
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      401         {object}  domain.ErrorResponse
// @Router       /api/v1/reports/export/pdf [get]
func (h *ReportHandler) ExportReportsPDF(c *fiber.Ctx) error {
	// Get user info
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)

	var userCompanyID *string
	if companyID != nil {
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = &companyIDStr
		}
	}

	// Get all reports based on RBAC
	reports, err := h.reportUseCase.GetAllReports(roleName, userCompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	// Apply filters
	companyIDFilter := c.Query("company_id")
	periodFilter := c.Query("period")

	if companyIDFilter != "" {
		var filtered []domain.ReportModel
		// Support multiple company IDs (comma-separated)
		companyIDs := strings.Split(companyIDFilter, ",")
		companyIDMap := make(map[string]bool)
		for _, id := range companyIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				companyIDMap[id] = true
			}
		}

		for _, r := range reports {
			if companyIDMap[r.CompanyID] {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	if periodFilter != "" {
		var filtered []domain.ReportModel
		for _, r := range reports {
			if r.Period == periodFilter {
				filtered = append(filtered, r)
			}
		}
		reports = filtered
	}

	// Create PDF
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Reports Export")
	pdf.Ln(15)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	headers := []string{"Period", "Company", "Revenue", "Opex", "NPAT", "Dividend", "Financial Ratio"}
	colWidths := []float64{30, 50, 30, 30, 30, 30, 30}

	// Print headers
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 9)
	for _, report := range reports {
		companyName := "Unknown"
		if report.Company != nil {
			companyName = report.Company.Name
			if len(companyName) > 30 {
				companyName = companyName[:27] + "..."
			}
		}

		pdf.CellFormat(colWidths[0], 6, report.Period, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 6, companyName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 6, formatNumber(report.Revenue), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[3], 6, formatNumber(report.Opex), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[4], 6, formatNumber(report.NPAT), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[5], 6, formatNumber(report.Dividend), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[6], 6, fmt.Sprintf("%.2f%%", report.FinancialRatio), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to generate PDF file",
		})
	}

	// Audit log
	audit.LogAction(userID, username, "export_reports_pdf", audit.ResourceReport, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"count":          len(reports),
		"company_filter": companyIDFilter,
		"period_filter":  periodFilter,
	})

	// Set headers and send file
	filename := fmt.Sprintf("reports_%s.pdf", time.Now().Format("20060102_150405"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(buf.Bytes())
}

// Helper function to format numbers
func formatNumber(num int64) string {
	str := strconv.FormatInt(num, 10)
	n := len(str)
	if n <= 3 {
		return str
	}
	result := ""
	for i := n - 1; i >= 0; i-- {
		result = string(str[i]) + result
		if (n-i)%3 == 0 && i > 0 {
			result = "," + result
		}
	}
	return result
}

// DownloadTemplate handles downloading Excel template for report upload
// @Summary      Download Report Template
// @Description  Download template Excel file untuk upload reports. Template berisi kolom-kolom yang diperlukan dengan contoh data.
// @Tags         Reports
// @Accept       json
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Success      200  {file}  application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /api/v1/reports/template [get]
func (h *ReportHandler) DownloadTemplate(c *fiber.Ctx) error {
	// Get user info to determine accessible companies
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	var accessibleCompanies []domain.CompanyModel
	var err error

	// Get accessible companies based on user role
	if roleName == "superadmin" {
		// Superadmin can access all companies
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

	// Set headers
	headers := []string{"Period (YYYY-MM)", "Company Code", "Revenue", "OPEX", "NPAT", "Dividend", "Financial Ratio (%)", "Remark"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
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

	// Add data rows with company codes pre-filled
	// Each company gets one row with example data
	currentMonth := time.Now().Format("2006-01")
	for i, company := range activeCompanies {
		row := i + 2
		// Period (current month as example)
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), currentMonth); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		// Company Code (pre-filled from database)
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), company.Code); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "template_failed",
				Message: fmt.Sprintf("Failed to set cell value: %v", err),
			})
		}
		// Example values for other columns (user can modify)
		exampleValues := []interface{}{1000000, 500000, 300000, 100000, 30.5, "Example remark"}
		columns := []string{"C", "D", "E", "F", "G", "H"}
		for j, value := range exampleValues {
			cell := fmt.Sprintf("%s%d", columns[j], row)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
					Error:   "template_failed",
					Message: fmt.Sprintf("Failed to set cell value: %v", err),
				})
			}
		}
	}

	// Set column widths
	columnWidths := map[string]float64{
		"A": 15, // Period
		"B": 15, // Company Code
		"C": 15, // Revenue
		"D": 15, // OPEX
		"E": 15, // NPAT
		"F": 15, // Dividend
		"G": 20, // Financial Ratio
		"H": 30, // Remark
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
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=report_template.xlsx")

	return c.Send(buf.Bytes())
}

// ValidateExcelFile validates an Excel file before upload
// @Summary      Validate Excel File
// @Description  Validates Excel file format and data before upload. Returns validation errors and parsed data.
// @Tags         Reports
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel file (.xlsx, .xls)"
// @Success      200   {object}  map[string]interface{}  "Response dengan valid, errors, dan data"
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      500   {object}  domain.ErrorResponse
// @Router       /api/v1/reports/validate [post]
func (h *ReportHandler) ValidateExcelFile(c *fiber.Ctx) error {
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

	// Get accessible companies
	var accessibleCompanyCodes map[string]bool
	if roleName == "superadmin" {
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

	// Validate headers (row 0, index 0)
	headers := rows[0]
	expectedHeaders := []string{"Period (YYYY-MM)", "Company Code", "Revenue", "OPEX", "NPAT", "Dividend", "Financial Ratio (%)", "Remark"}
	if len(headers) < len(expectedHeaders) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "Header tidak lengkap. Pastikan semua kolom ada",
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

		// Period (YYYY-MM)
		if len(row) > 0 {
			period := strings.TrimSpace(row[0])
			rowData["period"] = period
			if period == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Period (YYYY-MM)",
					"message": "Period wajib diisi",
				})
			} else {
				// Validate format YYYY-MM
				if _, err := time.Parse("2006-01", period); err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Period (YYYY-MM)",
						"message": "Format period harus YYYY-MM (contoh: 2024-01)",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Period (YYYY-MM)",
				"message": "Period wajib diisi",
			})
		}

		// Company Code
		if len(row) > 1 {
			companyCode := strings.TrimSpace(row[1])
			rowData["company_code"] = companyCode
			if companyCode == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Company Code",
					"message": "Company Code wajib diisi",
				})
			} else {
				// Validate company code exists and is accessible
				if accessibleCompanyCodes != nil && !accessibleCompanyCodes[companyCode] {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Company Code",
						"message": fmt.Sprintf("Company Code '%s' tidak ditemukan atau tidak dapat diakses", companyCode),
					})
				} else {
					// Check if company exists in database
					company, err := h.companyUseCase.GetCompanyByCode(companyCode)
					if err != nil || company == nil || !company.IsActive {
						rowErrors = append(rowErrors, map[string]interface{}{
							"row":     rowNum,
							"column":  "Company Code",
							"message": fmt.Sprintf("Company Code '%s' tidak ditemukan", companyCode),
						})
					}
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Company Code",
				"message": "Company Code wajib diisi",
			})
		}

		// Revenue
		if len(row) > 2 {
			revenueStr := strings.TrimSpace(row[2])
			rowData["revenue"] = revenueStr
			if revenueStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Revenue",
					"message": "Revenue wajib diisi",
				})
			} else {
				if revenue, err := strconv.ParseFloat(revenueStr, 64); err != nil || revenue < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Revenue",
						"message": "Revenue harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Revenue",
				"message": "Revenue wajib diisi",
			})
		}

		// OPEX
		if len(row) > 3 {
			opexStr := strings.TrimSpace(row[3])
			rowData["opex"] = opexStr
			if opexStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "OPEX",
					"message": "OPEX wajib diisi",
				})
			} else {
				if opex, err := strconv.ParseFloat(opexStr, 64); err != nil || opex < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "OPEX",
						"message": "OPEX harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "OPEX",
				"message": "OPEX wajib diisi",
			})
		}

		// NPAT
		if len(row) > 4 {
			npatStr := strings.TrimSpace(row[4])
			rowData["npat"] = npatStr
			if npatStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "NPAT",
					"message": "NPAT wajib diisi",
				})
			} else {
				if _, err := strconv.ParseFloat(npatStr, 64); err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "NPAT",
						"message": "NPAT harus berupa angka",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "NPAT",
				"message": "NPAT wajib diisi",
			})
		}

		// Dividend
		if len(row) > 5 {
			dividendStr := strings.TrimSpace(row[5])
			rowData["dividend"] = dividendStr
			if dividendStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Dividend",
					"message": "Dividend wajib diisi",
				})
			} else {
				if dividend, err := strconv.ParseFloat(dividendStr, 64); err != nil || dividend < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Dividend",
						"message": "Dividend harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Dividend",
				"message": "Dividend wajib diisi",
			})
		}

		// Financial Ratio (optional)
		if len(row) > 6 {
			financialRatioStr := strings.TrimSpace(row[6])
			rowData["financial_ratio"] = financialRatioStr
			if financialRatioStr != "" {
				if _, err := strconv.ParseFloat(financialRatioStr, 64); err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Financial Ratio (%)",
						"message": "Financial Ratio harus berupa angka",
					})
				}
			}
		}

		// Remark (optional)
		if len(row) > 7 {
			rowData["remark"] = strings.TrimSpace(row[7])
		}

		// Add row errors to errors list
		errors = append(errors, rowErrors...)

		// Add row data
		data = append(data, rowData)
	}

	// Return validation result
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
		"data":   data,
	})
}

// UploadReports handles uploading Excel file and creating reports in bulk
// @Summary      Upload Reports from Excel
// @Description  Upload Excel file, validate rows, and create reports in database.
// @Tags         Reports
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel file (.xlsx, .xls)"
// @Success      200   {object}  map[string]interface{}  "Response dengan success, failed, dan errors"
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      500   {object}  domain.ErrorResponse
// @Router       /api/v1/reports/upload [post]
func (h *ReportHandler) UploadReports(c *fiber.Ctx) error {
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

	// Determine accessible company codes (same logic as validation)
	var accessibleCompanyCodes map[string]bool
	if roleName == "superadmin" {
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

	headers := rows[0]
	expectedHeaders := []string{"Period (YYYY-MM)", "Company Code", "Revenue", "OPEX", "NPAT", "Dividend", "Financial Ratio (%)", "Remark"}
	if len(headers) < len(expectedHeaders) {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_excel_file",
			Message: "Header tidak lengkap. Pastikan semua kolom ada",
		})
	}

	// Process rows
	errorsList := []map[string]interface{}{}
	successCount := 0
	failedCount := 0

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		rowNum := rowIndex + 1

		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		rowErrors := []map[string]interface{}{}

		// Period (YYYY-MM)
		var period string
		if len(row) > 0 {
			period = strings.TrimSpace(row[0])
			if period == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Period (YYYY-MM)",
					"message": "Period wajib diisi",
				})
			} else if _, err := time.Parse("2006-01", period); err != nil {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Period (YYYY-MM)",
					"message": "Format period harus YYYY-MM (contoh: 2024-01)",
				})
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Period (YYYY-MM)",
				"message": "Period wajib diisi",
			})
		}

		// Company Code
		var companyCode string
		var companyIDValue string
		if len(row) > 1 {
			companyCode = strings.TrimSpace(row[1])
			if companyCode == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Company Code",
					"message": "Company Code wajib diisi",
				})
			} else {
				if accessibleCompanyCodes != nil && !accessibleCompanyCodes[companyCode] {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Company Code",
						"message": fmt.Sprintf("Company Code '%s' tidak ditemukan atau tidak dapat diakses", companyCode),
					})
				} else {
					company, err := h.companyUseCase.GetCompanyByCode(companyCode)
					if err != nil || company == nil || !company.IsActive {
						rowErrors = append(rowErrors, map[string]interface{}{
							"row":     rowNum,
							"column":  "Company Code",
							"message": fmt.Sprintf("Company Code '%s' tidak ditemukan", companyCode),
						})
					} else {
						companyIDValue = company.ID
					}
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Company Code",
				"message": "Company Code wajib diisi",
			})
		}

		// Revenue
		var revenueValue float64
		if len(row) > 2 {
			revenueStr := strings.TrimSpace(row[2])
			if revenueStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Revenue",
					"message": "Revenue wajib diisi",
				})
			} else {
				var err error
				revenueValue, err = strconv.ParseFloat(revenueStr, 64)
				if err != nil || revenueValue < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Revenue",
						"message": "Revenue harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Revenue",
				"message": "Revenue wajib diisi",
			})
		}

		// OPEX
		var opexValue float64
		if len(row) > 3 {
			opexStr := strings.TrimSpace(row[3])
			if opexStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "OPEX",
					"message": "OPEX wajib diisi",
				})
			} else {
				var err error
				opexValue, err = strconv.ParseFloat(opexStr, 64)
				if err != nil || opexValue < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "OPEX",
						"message": "OPEX harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "OPEX",
				"message": "OPEX wajib diisi",
			})
		}

		// NPAT
		var npatValue float64
		if len(row) > 4 {
			npatStr := strings.TrimSpace(row[4])
			if npatStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "NPAT",
					"message": "NPAT wajib diisi",
				})
			} else {
				var err error
				npatValue, err = strconv.ParseFloat(npatStr, 64)
				if err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "NPAT",
						"message": "NPAT harus berupa angka",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "NPAT",
				"message": "NPAT wajib diisi",
			})
		}

		// Dividend
		var dividendValue float64
		if len(row) > 5 {
			dividendStr := strings.TrimSpace(row[5])
			if dividendStr == "" {
				rowErrors = append(rowErrors, map[string]interface{}{
					"row":     rowNum,
					"column":  "Dividend",
					"message": "Dividend wajib diisi",
				})
			} else {
				var err error
				dividendValue, err = strconv.ParseFloat(dividendStr, 64)
				if err != nil || dividendValue < 0 {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Dividend",
						"message": "Dividend harus berupa angka positif atau nol",
					})
				}
			}
		} else {
			rowErrors = append(rowErrors, map[string]interface{}{
				"row":     rowNum,
				"column":  "Dividend",
				"message": "Dividend wajib diisi",
			})
		}

		// Financial Ratio (optional, default 0)
		var financialRatioValue float64
		if len(row) > 6 {
			financialRatioStr := strings.TrimSpace(row[6])
			if financialRatioStr != "" {
				var err error
				financialRatioValue, err = strconv.ParseFloat(financialRatioStr, 64)
				if err != nil {
					rowErrors = append(rowErrors, map[string]interface{}{
						"row":     rowNum,
						"column":  "Financial Ratio (%)",
						"message": "Financial Ratio harus berupa angka",
					})
				}
			}
		}

		// Remark (optional)
		var remarkValue *string
		if len(row) > 7 {
			remark := strings.TrimSpace(row[7])
			if remark != "" {
				remarkValue = &remark
			}
		}

		// Collect errors and continue
		if len(rowErrors) > 0 {
			errorsList = append(errorsList, rowErrors...)
			failedCount++
			continue
		}

		// Prepare request
		inputterID := userID
		createReq := domain.CreateReportRequest{
			Period:         period,
			CompanyID:      companyIDValue,
			InputterID:     &inputterID,
			Revenue:        int64(revenueValue),
			Opex:           int64(opexValue),
			NPAT:           int64(npatValue),
			Dividend:       int64(dividendValue),
			FinancialRatio: financialRatioValue,
			Remark:         remarkValue,
		}

		report, err := h.reportUseCase.CreateReport(&createReq)
		if err != nil {
			errorsList = append(errorsList, map[string]interface{}{
				"row":     rowNum,
				"column":  "general",
				"message": err.Error(),
			})
			failedCount++
			continue
		}

		// Audit log per success
		audit.LogAction(userID, username, "upload_report", audit.ResourceReport, report.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"company_id": companyIDValue,
			"period":     period,
			"source":     "excel_upload",
		})

		successCount++
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"success": successCount,
		"failed":  failedCount,
		"errors":  errorsList,
	})
}
