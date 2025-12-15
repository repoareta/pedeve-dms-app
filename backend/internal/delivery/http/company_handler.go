package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
)

// CompanyHandler handles company-related HTTP requests
type CompanyHandler struct {
	companyUseCase usecase.CompanyUseCase
}

// NewCompanyHandler creates a new company handler
func NewCompanyHandler(companyUseCase usecase.CompanyUseCase) *CompanyHandler {
	return &CompanyHandler{
		companyUseCase: companyUseCase,
	}
}

// CreateCompany handles company creation
// @Summary      Buat Company Baru
// @Description  Membuat company baru dalam hierarchy. Superadmin bisa membuat company di level manapun. Admin company hanya bisa membuat sub-company di bawah company mereka.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company  body      object  true  "Company data (name, code, description, parent_id optional)"
// @Success      201      {object}  domain.CompanyModel
// @Failure      400      {object}  domain.ErrorResponse
// @Failure      401      {object}  domain.ErrorResponse
// @Failure      403      {object}  domain.ErrorResponse
// @Router       /api/v1/companies [post]
func (h *CompanyHandler) CreateCompany(c *fiber.Ctx) error {
	var req struct {
		Name        string  `json:"name" validate:"required"`
		Code        string  `json:"code" validate:"required"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
	}

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

	// Superadmin/administrator can create company at any level
	// Admin can only create sub-company under their company
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
		if req.ParentID != nil && *req.ParentID != userCompanyID {
			// Check if parent is descendant of user's company
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, *req.ParentID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only create sub-company under your company or its descendants",
				})
			}
		} else if req.ParentID == nil {
			// Non-superadmin: set parent to their company if not specified
			req.ParentID = &userCompanyID
		}
	}

	company, err := h.companyUseCase.CreateCompany(req.Name, req.Code, req.Description, req.ParentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	audit.LogAction(userID, username, audit.ActionCreateCompany, audit.ResourceCompany, company.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusCreated).JSON(company)
}

// CreateCompanyFull handles company creation with full data
// @Summary      Buat Company Baru (Full Data)
// @Description  Membuat company baru dengan data lengkap termasuk shareholders, business fields, dan directors
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        company  body      object  true  "Company full data"
// @Success      201      {object}  domain.CompanyModel
// @Failure      400      {object}  domain.ErrorResponse
// @Failure      401      {object}  domain.ErrorResponse
// @Failure      403      {object}  domain.ErrorResponse
// @Router       /api/v1/companies/full [post]
func (h *CompanyHandler) CreateCompanyFull(c *fiber.Ctx) error {
	var req domain.CompanyCreateRequest

	// Log request body untuk debugging
	bodyBytes := c.Body()
	zapLog := logger.GetLogger()
	zapLog.Info("CreateCompanyFull request body", zap.String("body", string(bodyBytes)))

	if err := c.BodyParser(&req); err != nil {
		zapLog.Error("Failed to parse request body", zap.Error(err), zap.String("body", string(bodyBytes)))
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

	// Superadmin can create company at any level
	// Admin can only create sub-company under their company
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
		if req.ParentID != nil && *req.ParentID != userCompanyID {
			hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, *req.ParentID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only create sub-company under your company or its descendants",
				})
			}
		} else if req.ParentID == nil {
			// Non-superadmin: set parent to their company if not specified
			req.ParentID = &userCompanyID
		}
	}

	company, err := h.companyUseCase.CreateCompanyFull(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	// Get full company data with relationships
	fullCompany, _ := h.companyUseCase.GetCompanyByID(company.ID)

	// Audit log
	audit.LogAction(userID, username, audit.ActionCreateCompany, audit.ResourceCompany, company.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusCreated).JSON(fullCompany)
}

// UpdateCompanyFull handles company update with full data
// @Summary      Update Company (Full Data)
// @Description  Mengupdate company dengan data lengkap termasuk shareholders, business fields, dan directors
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Company ID"
// @Param        company  body      object  true  "Company full data to update"
// @Success      200      {object}  domain.CompanyModel
// @Failure      400      {object}  domain.ErrorResponse
// @Failure      403      {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id}/full [put]
func (h *CompanyHandler) UpdateCompanyFull(c *fiber.Ctx) error {
	id := c.Params("id")
	var req domain.CompanyUpdateRequest

	// Log request body untuk debugging
	bodyBytes := c.Body()
	zapLog := logger.GetLogger()
	zapLog.Info("UpdateCompanyFull request body", zap.String("body", string(bodyBytes)))

	if err := c.BodyParser(&req); err != nil {
		zapLog.Error("Failed to parse request body", zap.Error(err), zap.String("body", string(bodyBytes)))
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// CRITICAL: Log access check untuk debugging
	zapLog.Info("UpdateCompanyFull access check",
		zap.String("company_id", id),
		zap.String("role", roleName),
		zap.Any("user_company_id", companyID),
	)

	// Check access
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

		// CRITICAL: For admin, check holding company access
		// Admin holding boleh edit holding mereka sendiri, tapi tidak boleh edit holding lain
		targetCompany, err := h.companyUseCase.GetCompanyByID(id)
		if err == nil && targetCompany != nil && targetCompany.Code == "PDV" {
			// Allow admin holding to edit their own holding
			if userCompanyID != targetCompany.ID {
				// Admin holding lain tidak bisa edit holding
				zapLog.Warn("Admin attempted to update holding company (not their own), blocking",
					zap.String("user_company_id", userCompanyID),
					zap.String("target_company_id", id),
					zap.String("target_company_code", targetCompany.Code),
				)
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "Admin tidak dapat mengupdate holding company lain. Hanya admin holding yang di-assign di holding tersebut yang dapat mengupdate holding mereka sendiri.",
				})
			}
			// Admin holding bisa edit holding mereka sendiri - continue to access check
			zapLog.Info("Admin holding updating their own holding",
				zap.String("user_company_id", userCompanyID),
				zap.String("target_company_id", id),
			)
		}

		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			zapLog.Warn("Access denied for company update",
				zap.String("user_company_id", userCompanyID),
				zap.String("target_company_id", id),
				zap.Error(err),
			)
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to update this company",
			})
		}
	}

	// Get old company data before update for audit log
	oldCompany, _ := h.companyUseCase.GetCompanyByID(id)

	company, err := h.companyUseCase.UpdateCompanyFull(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	// Get full company data with relationships
	fullCompany, _ := h.companyUseCase.GetCompanyByID(company.ID)

	// Prepare audit details with changes
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)

	// Compare old and new values to create changes map
	changes := make(map[string]map[string]interface{})
	if oldCompany != nil {
		if oldCompany.Name != company.Name {
			changes["name"] = map[string]interface{}{"old": oldCompany.Name, "new": company.Name}
		}
		if oldCompany.ShortName != company.ShortName {
			changes["short_name"] = map[string]interface{}{"old": oldCompany.ShortName, "new": company.ShortName}
		}
		if oldCompany.Description != company.Description {
			changes["description"] = map[string]interface{}{"old": oldCompany.Description, "new": company.Description}
		}
		if oldCompany.NPWP != company.NPWP {
			changes["npwp"] = map[string]interface{}{"old": oldCompany.NPWP, "new": company.NPWP}
		}
		if oldCompany.NIB != company.NIB {
			changes["nib"] = map[string]interface{}{"old": oldCompany.NIB, "new": company.NIB}
		}
		if oldCompany.Status != company.Status {
			changes["status"] = map[string]interface{}{"old": oldCompany.Status, "new": company.Status}
		}
		if oldCompany.Logo != company.Logo {
			changes["logo"] = map[string]interface{}{"old": oldCompany.Logo, "new": company.Logo}
		}
		if oldCompany.Phone != company.Phone {
			changes["phone"] = map[string]interface{}{"old": oldCompany.Phone, "new": company.Phone}
		}
		if oldCompany.Fax != company.Fax {
			changes["fax"] = map[string]interface{}{"old": oldCompany.Fax, "new": company.Fax}
		}
		if oldCompany.Email != company.Email {
			changes["email"] = map[string]interface{}{"old": oldCompany.Email, "new": company.Email}
		}
		if oldCompany.Website != company.Website {
			changes["website"] = map[string]interface{}{"old": oldCompany.Website, "new": company.Website}
		}
		if oldCompany.Address != company.Address {
			changes["address"] = map[string]interface{}{"old": oldCompany.Address, "new": company.Address}
		}
		if oldCompany.OperationalAddress != company.OperationalAddress {
			changes["operational_address"] = map[string]interface{}{"old": oldCompany.OperationalAddress, "new": company.OperationalAddress}
		}
		if oldCompany.Code != company.Code {
			changes["code"] = map[string]interface{}{"old": oldCompany.Code, "new": company.Code}
		}

		// Handle parent_id comparison (handle nil pointers)
		oldParentID := ""
		if oldCompany.ParentID != nil {
			oldParentID = *oldCompany.ParentID
		}
		newParentID := ""
		if company.ParentID != nil {
			newParentID = *company.ParentID
		}
		if oldParentID != newParentID {
			changes["parent_id"] = map[string]interface{}{"old": oldParentID, "new": newParentID}
		}

		// Handle authorized_capital comparison
		oldAuthCapital := int64(0)
		if oldCompany.AuthorizedCapital != nil {
			oldAuthCapital = *oldCompany.AuthorizedCapital
		}
		newAuthCapital := int64(0)
		if company.AuthorizedCapital != nil {
			newAuthCapital = *company.AuthorizedCapital
		}
		if oldAuthCapital != newAuthCapital {
			changes["authorized_capital"] = map[string]interface{}{"old": oldAuthCapital, "new": newAuthCapital}
		}

		// Handle paid_up_capital comparison
		oldPaidCapital := int64(0)
		if oldCompany.PaidUpCapital != nil {
			oldPaidCapital = *oldCompany.PaidUpCapital
		}
		newPaidCapital := int64(0)
		if company.PaidUpCapital != nil {
			newPaidCapital = *company.PaidUpCapital
		}
		if oldPaidCapital != newPaidCapital {
			changes["paid_up_capital"] = map[string]interface{}{"old": oldPaidCapital, "new": newPaidCapital}
		}

		// Handle currency comparison
		if oldCompany.Currency != company.Currency {
			changes["currency"] = map[string]interface{}{"old": oldCompany.Currency, "new": company.Currency}
		}

		// Handle main_parent_company comparison (handle nil pointers)
		oldMainParent := ""
		if oldCompany.MainParentCompanyID != nil {
			oldMainParent = *oldCompany.MainParentCompanyID
		}
		newMainParent := ""
		if company.MainParentCompanyID != nil {
			newMainParent = *company.MainParentCompanyID
		}
		if oldMainParent != newMainParent {
			changes["main_parent_company"] = map[string]interface{}{"old": oldMainParent, "new": newMainParent}
		}

		// Compare shareholders - track additions, deletions, and modifications
		oldShareholders := oldCompany.Shareholders
		newShareholders := fullCompany.Shareholders

		// If count changed, record added/removed shareholders with details
		if len(oldShareholders) < len(newShareholders) {
			// Shareholders were added
			for i := len(oldShareholders); i < len(newShareholders); i++ {
				newSh := newShareholders[i]
				shareholderInfo := map[string]interface{}{
					"action": "added",
					"name":   newSh.Name,
				}
				if newSh.Type != "" {
					shareholderInfo["type"] = newSh.Type
				}
				if newSh.IdentityNumber != "" {
					shareholderInfo["identity_number"] = newSh.IdentityNumber
				}
				if newSh.ShareholderCompanyID != nil {
					shareholderInfo["shareholder_company_id"] = *newSh.ShareholderCompanyID
				}
				changes[fmt.Sprintf("shareholder_added_%d", i-len(oldShareholders))] = shareholderInfo
			}
		} else if len(oldShareholders) > len(newShareholders) {
			// Shareholders were removed
			for i := len(newShareholders); i < len(oldShareholders); i++ {
				oldSh := oldShareholders[i]
				shareholderInfo := map[string]interface{}{
					"action": "removed",
					"name":   oldSh.Name,
				}
				if oldSh.Type != "" {
					shareholderInfo["type"] = oldSh.Type
				}
				if oldSh.IdentityNumber != "" {
					shareholderInfo["identity_number"] = oldSh.IdentityNumber
				}
				if oldSh.ShareholderCompanyID != nil {
					shareholderInfo["shareholder_company_id"] = *oldSh.ShareholderCompanyID
				}
				changes[fmt.Sprintf("shareholder_removed_%d", i-len(newShareholders))] = shareholderInfo
			}
		}

		// Check for individual shareholder changes (for shareholders that exist in both)
		maxLen := len(oldShareholders)
		if len(newShareholders) < maxLen {
			maxLen = len(newShareholders)
		}
		for i := 0; i < maxLen; i++ {
			oldSh := oldShareholders[i]
			newSh := newShareholders[i]
			if oldSh.Name != newSh.Name {
				changes[fmt.Sprintf("shareholder_%d_name", i)] = map[string]interface{}{"old": oldSh.Name, "new": newSh.Name}
			}
			if oldSh.Type != newSh.Type {
				changes[fmt.Sprintf("shareholder_%d_type", i)] = map[string]interface{}{"old": oldSh.Type, "new": newSh.Type}
			}
			// Handle shareholder_company_id comparison
			oldShCompanyID := ""
			if oldSh.ShareholderCompanyID != nil {
				oldShCompanyID = *oldSh.ShareholderCompanyID
			}
			newShCompanyID := ""
			if newSh.ShareholderCompanyID != nil {
				newShCompanyID = *newSh.ShareholderCompanyID
			}
			if oldShCompanyID != newShCompanyID {
				changes[fmt.Sprintf("shareholder_%d_company_id", i)] = map[string]interface{}{"old": oldShCompanyID, "new": newShCompanyID}
			}
			if oldSh.IdentityNumber != newSh.IdentityNumber {
				changes[fmt.Sprintf("shareholder_%d_identity_number", i)] = map[string]interface{}{"old": oldSh.IdentityNumber, "new": newSh.IdentityNumber}
			}
			if oldSh.OwnershipPercent != newSh.OwnershipPercent {
				changes[fmt.Sprintf("shareholder_%d_ownership_percent", i)] = map[string]interface{}{"old": oldSh.OwnershipPercent, "new": newSh.OwnershipPercent}
			}
			if oldSh.ShareSheetCount != nil && newSh.ShareSheetCount != nil && *oldSh.ShareSheetCount != *newSh.ShareSheetCount {
				changes[fmt.Sprintf("shareholder_%d_share_sheet_count", i)] = map[string]interface{}{"old": *oldSh.ShareSheetCount, "new": *newSh.ShareSheetCount}
			}
			if oldSh.ShareValuePerSheet != nil && newSh.ShareValuePerSheet != nil && *oldSh.ShareValuePerSheet != *newSh.ShareValuePerSheet {
				changes[fmt.Sprintf("shareholder_%d_share_value_per_sheet", i)] = map[string]interface{}{"old": *oldSh.ShareValuePerSheet, "new": *newSh.ShareValuePerSheet}
			}
		}

		// Compare directors - track additions, deletions, and modifications
		oldDirectors := oldCompany.Directors
		newDirectors := fullCompany.Directors

		// If count changed, record added/removed directors with details
		if len(oldDirectors) < len(newDirectors) {
			// Directors were added
			for i := len(oldDirectors); i < len(newDirectors); i++ {
				newDir := newDirectors[i]
				changes[fmt.Sprintf("director_added_%d", i-len(oldDirectors))] = map[string]interface{}{
					"action":    "added",
					"position":  newDir.Position,
					"full_name": newDir.FullName,
					"ktp":       newDir.KTP,
				}
			}
		} else if len(oldDirectors) > len(newDirectors) {
			// Directors were removed
			for i := len(newDirectors); i < len(oldDirectors); i++ {
				oldDir := oldDirectors[i]
				changes[fmt.Sprintf("director_removed_%d", i-len(newDirectors))] = map[string]interface{}{
					"action":    "removed",
					"position":  oldDir.Position,
					"full_name": oldDir.FullName,
					"ktp":       oldDir.KTP,
				}
			}
		}

		// Check for individual director changes (for directors that exist in both)
		directorMaxLen := len(oldDirectors)
		if len(newDirectors) < directorMaxLen {
			directorMaxLen = len(newDirectors)
		}
		for i := 0; i < directorMaxLen; i++ {
			oldDir := oldDirectors[i]
			newDir := newDirectors[i]
			if oldDir.FullName != newDir.FullName {
				changes[fmt.Sprintf("director_%d_full_name", i)] = map[string]interface{}{"old": oldDir.FullName, "new": newDir.FullName}
			}
			if oldDir.Position != newDir.Position {
				changes[fmt.Sprintf("director_%d_position", i)] = map[string]interface{}{"old": oldDir.Position, "new": newDir.Position}
			}
			if oldDir.KTP != newDir.KTP {
				changes[fmt.Sprintf("director_%d_ktp", i)] = map[string]interface{}{"old": oldDir.KTP, "new": newDir.KTP}
			}
			if oldDir.NPWP != newDir.NPWP {
				changes[fmt.Sprintf("director_%d_npwp", i)] = map[string]interface{}{"old": oldDir.NPWP, "new": newDir.NPWP}
			}
			if oldDir.DomicileAddress != newDir.DomicileAddress {
				changes[fmt.Sprintf("director_%d_domicile_address", i)] = map[string]interface{}{"old": oldDir.DomicileAddress, "new": newDir.DomicileAddress}
			}
		}

		// Compare business fields (main business)
		// Get main business from BusinessFields (is_main = true) or first one
		var oldMainBusiness *domain.BusinessFieldModel
		if len(oldCompany.BusinessFields) > 0 {
			for _, bf := range oldCompany.BusinessFields {
				if bf.IsMain {
					oldMainBusiness = &bf
					break
				}
			}
			if oldMainBusiness == nil {
				oldMainBusiness = &oldCompany.BusinessFields[0]
			}
		}

		var newMainBusiness *domain.BusinessFieldModel
		if len(fullCompany.BusinessFields) > 0 {
			for _, bf := range fullCompany.BusinessFields {
				if bf.IsMain {
					newMainBusiness = &bf
					break
				}
			}
			if newMainBusiness == nil {
				newMainBusiness = &fullCompany.BusinessFields[0]
			}
		}
		if oldMainBusiness != nil && newMainBusiness != nil {
			if oldMainBusiness.IndustrySector != newMainBusiness.IndustrySector {
				changes["business_industry_sector"] = map[string]interface{}{"old": oldMainBusiness.IndustrySector, "new": newMainBusiness.IndustrySector}
			}
			if oldMainBusiness.KBLI != newMainBusiness.KBLI {
				changes["business_kbli"] = map[string]interface{}{"old": oldMainBusiness.KBLI, "new": newMainBusiness.KBLI}
			}
			if oldMainBusiness.MainBusinessActivity != newMainBusiness.MainBusinessActivity {
				changes["business_main_activity"] = map[string]interface{}{"old": oldMainBusiness.MainBusinessActivity, "new": newMainBusiness.MainBusinessActivity}
			}
			if oldMainBusiness.AdditionalActivities != newMainBusiness.AdditionalActivities {
				changes["business_additional_activities"] = map[string]interface{}{"old": oldMainBusiness.AdditionalActivities, "new": newMainBusiness.AdditionalActivities}
			}
		}
	}

	// Audit log with changes details
	auditDetails := map[string]interface{}{}
	if len(changes) > 0 {
		auditDetails["changes"] = changes
	}

	audit.LogAction(userID, username, audit.ActionUpdateCompany, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, auditDetails)

	return c.Status(fiber.StatusOK).JSON(fullCompany)
}

// GetCompany handles getting company by ID
// @Summary      Ambil Company by ID
// @Description  Mengambil informasi company berdasarkan ID. User hanya bisa mengakses company mereka atau descendants.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company ID"
// @Success      200  {object}  domain.CompanyModel
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id} [get]
func (h *CompanyHandler) GetCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Superadmin/administrator can access any company
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
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this company",
			})
		}
	}

	company, err := h.companyUseCase.GetCompanyByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Company not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(company)
}

// GetAllCompanies handles getting all companies
// @Summary      Ambil Semua Companies
// @Description  Mengambil daftar semua companies. Superadmin melihat semua. User lain hanya melihat company mereka dan descendants.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.CompanyModel
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/companies [get]
func (h *CompanyHandler) GetAllCompanies(c *fiber.Ctx) error {
	roleName := c.Locals("roleName").(string)
	companyID := c.Locals("companyID")

	var companies []domain.CompanyModel
	var err error

	// Superadmin/administrator sees all companies
	if utils.IsSuperAdminLike(roleName) {
		companies, err = h.companyUseCase.GetAllCompanies()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get companies: " + err.Error(),
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

		// Handle *string type
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
				Error:   "internal_error",
				Message: "Failed to get user company: " + err.Error(),
			})
		}

		// Only include user's company if it's active
		if !userCompany.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company is not active",
			})
		}

		// Get all descendants (includes direct children and all nested descendants)
		descendants, err := h.companyUseCase.GetCompanyDescendants(userCompanyID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get company descendants: " + err.Error(),
			})
		}

		// CRITICAL: Remove duplicates by ID to prevent duplicate entries in response
		// This is a safety measure in case GetDescendants returns duplicates
		companyMap := make(map[string]domain.CompanyModel)
		companyMap[userCompany.ID] = *userCompany
		for _, desc := range descendants {
			// Skip if already exists (prevent duplicate)
			if _, exists := companyMap[desc.ID]; !exists {
				companyMap[desc.ID] = desc
			}
		}

		// Convert map back to slice
		companies = make([]domain.CompanyModel, 0, len(companyMap))
		for _, comp := range companyMap {
			companies = append(companies, comp)
		}
	}

	return c.Status(fiber.StatusOK).JSON(companies)
}

// GetCompanyUsers handles getting users assigned to a company
// @Summary      Ambil Users di Company
// @Description  Mengambil daftar users yang di-assign ke company tertentu (menggunakan junction table, support multiple assignments).
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company ID"
// @Success      200  {array}   domain.UserModel
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id}/users [get]
func (h *CompanyHandler) GetCompanyUsers(c *fiber.Ctx) error {
	companyID := c.Params("id")
	userCompanyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Superadmin/administrator can access any company users
	if !utils.IsSuperAdminLike(roleName) && userCompanyID != nil {
		var currentUserCompanyID string
		if companyIDPtr, ok := userCompanyID.(*string); ok && companyIDPtr != nil {
			currentUserCompanyID = *companyIDPtr
		} else if companyIDStr, ok := userCompanyID.(string); ok {
			currentUserCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(currentUserCompanyID, companyID)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this company's users",
			})
		}
	}

	// Get users from junction table (supports multiple company assignments)
	userUseCase := usecase.NewUserManagementUseCase()
	users, err := userUseCase.GetUsersByCompany(companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get company users: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// GetCompanyAncestors handles getting company ancestors
// @Summary      Ambil Ancestors Company
// @Description  Mengambil daftar ancestors (parent companies) dari company tertentu.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company ID"
// @Success      200  {array}   domain.CompanyModel
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id}/ancestors [get]
func (h *CompanyHandler) GetCompanyAncestors(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Company ID is required",
		})
	}

	// Get user info for access validation
	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User role not found",
		})
	}
	roleName := roleNameVal.(string)
	companyID := c.Locals("companyID")

	// Check if company exists
	_, err := h.companyUseCase.GetCompanyByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Company not found",
		})
	}

	// Access control: Superadmin/administrator can see all companies
	if !utils.IsSuperAdminLike(roleName) {
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User company not found",
			})
		}

		// Handle *string type
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

		// Check if user has access to this company or its descendants
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			// Also check if the target company is an ancestor of user's company
			// (user can view ancestors of companies that are ancestors of their company)
			isDescendant, err := h.companyUseCase.ValidateCompanyAccess(id, userCompanyID)
			if err != nil || !isDescendant {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You don't have permission to view ancestors of this company",
				})
			}
		}
	}

	// Get ancestors
	ancestors, err := h.companyUseCase.GetCompanyAncestors(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get company ancestors: " + err.Error(),
		})
	}

	return c.JSON(ancestors)
}

// GetCompanyChildren handles getting company children
// @Summary      Ambil Children Company
// @Description  Mengambil daftar children (sub-companies) dari company tertentu.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company ID"
// @Success      200  {array}   domain.CompanyModel
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id}/children [get]
func (h *CompanyHandler) GetCompanyChildren(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Check access
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
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this company",
			})
		}
	}

	children, err := h.companyUseCase.GetCompanyChildren(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get company children",
		})
	}

	return c.Status(fiber.StatusOK).JSON(children)
}

// UpdateCompany handles company update
// @Summary      Update Company
// @Description  Mengupdate informasi company. Hanya bisa update company sendiri atau descendants.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Company ID"
// @Param        company  body      object  true  "Company data to update"
// @Success      200      {object}  domain.CompanyModel
// @Failure      400      {object}  domain.ErrorResponse
// @Failure      403      {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id} [put]
func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Check access
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
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to update this company",
			})
		}
	}

	company, err := h.companyUseCase.UpdateCompany(id, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionUpdateCompany, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(company)
}

// DeleteCompany handles company deletion (soft delete)
// @Summary      Hapus Company
// @Description  Menghapus company (soft delete: set is_active=false). Hanya bisa delete company sendiri atau descendants.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company ID"
// @Success      200  {object}  map[string]string
// @Failure      403  {object}  domain.ErrorResponse
// @Router       /api/v1/companies/{id} [delete]
func (h *CompanyHandler) DeleteCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Check access
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

		// CRITICAL: Admin tidak boleh menghapus perusahaan mereka sendiri (termasuk holding)
		// Superadmin bisa delete, tapi admin tidak bisa
		if id == userCompanyID {
			zapLog := logger.GetLogger()
			zapLog.Warn("Admin attempted to delete their own company, blocking",
				zap.String("user_company_id", userCompanyID),
				zap.String("role", roleName),
			)
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin tidak dapat menghapus perusahaan mereka sendiri. Hanya superadmin yang dapat menghapus perusahaan.",
			})
		}

		// CRITICAL: Admin tidak boleh menghapus holding company (bahkan jika bukan perusahaan mereka sendiri)
		targetCompany, err := h.companyUseCase.GetCompanyByID(id)
		if err == nil && targetCompany != nil && targetCompany.Code == "PDV" {
			zapLog := logger.GetLogger()
			zapLog.Warn("Admin attempted to delete holding company, blocking",
				zap.String("user_company_id", userCompanyID),
				zap.String("target_company_id", id),
				zap.String("role", roleName),
			)
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin tidak dapat menghapus holding company. Hanya superadmin yang dapat menghapus holding.",
			})
		}

		// Cek apakah target company adalah descendant (boleh dihapus)
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to delete this company",
			})
		}
	}

	if err := h.companyUseCase.DeleteCompany(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionDeleteCompany, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Company deleted successfully",
	})
}
