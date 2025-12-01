package http

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	reportUseCase usecase.ReportUseCase
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportUseCase usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{
		reportUseCase: reportUseCase,
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

	// Validate access: non-superadmin can only create reports for their own company
	if roleName != "superadmin" && companyID != nil {
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

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "view_report", audit.ResourceReport, report.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"company_id": report.CompanyID,
		"period":     report.Period,
	})

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
	c.Set("X-Total-Count", string(rune(total)))
	c.Set("X-Page", string(rune(page)))
	c.Set("X-Page-Size", string(rune(pageSize)))

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "list_reports", audit.ResourceReport, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"company_filter": companyIDFilter,
		"period_filter": periodFilter,
	})

	return c.JSON(map[string]interface{}{
		"data":       paginatedReports,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
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
	f.DeleteSheet("Sheet1")

	// Set headers
	headers := []string{"Period", "Company", "Revenue", "Opex", "NPAT", "Dividend", "Financial Ratio (%)", "Inputter", "Remark"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// Add data
	for i, report := range reports {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), report.Period)
		companyName := "Unknown"
		if report.Company != nil {
			companyName = report.Company.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), companyName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), report.Revenue)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), report.Opex)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), report.NPAT)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), report.Dividend)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("%.2f", report.FinancialRatio))
		inputterName := "N/A"
		if report.Inputter != nil {
			inputterName = report.Inputter.Username
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), inputterName)
		remark := "N/A"
		if report.Remark != nil {
			remark = *report.Remark
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), remark)
	}

	// Auto-size columns
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
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
		"count":         len(reports),
		"company_filter": companyIDFilter,
		"period_filter": periodFilter,
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
		"count":         len(reports),
		"company_filter": companyIDFilter,
		"period_filter": periodFilter,
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

