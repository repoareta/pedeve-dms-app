package http

import (
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
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

	// Superadmin can create company at any level
	// Admin can only create sub-company under their company
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
	audit.LogAction(userID, username, audit.ActionCreateUser, audit.ResourceCompany, company.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

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
	audit.LogAction(userID, username, audit.ActionCreateUser, audit.ResourceCompany, company.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

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

	// Check access
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
		hasAccess, err := h.companyUseCase.ValidateCompanyAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to update this company",
			})
		}
	}

	company, err := h.companyUseCase.UpdateCompanyFull(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	// Get full company data with relationships
	fullCompany, _ := h.companyUseCase.GetCompanyByID(company.ID)

	// Audit log
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionUpdateUser, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

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

	// Superadmin can access any company
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

	// Superadmin sees all companies
	if roleName == "superadmin" {
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

		// Combine user's company with all descendants
		// Note: descendants already includes direct children, so we just combine them
		companies = append([]domain.CompanyModel{*userCompany}, descendants...)
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

	// Superadmin can access any company users
	if roleName != "superadmin" && userCompanyID != nil {
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
	audit.LogAction(userID, username, audit.ActionUpdateUser, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

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

		// User tidak boleh menghapus perusahaan mereka sendiri
		if id == userCompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You cannot delete your own company",
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
	audit.LogAction(userID, username, audit.ActionDeleteUser, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Company deleted successfully",
	})
}

