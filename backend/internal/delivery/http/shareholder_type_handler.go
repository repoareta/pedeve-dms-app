package http

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
)

// ShareholderTypeHandler handles shareholder type-related HTTP requests
type ShareholderTypeHandler struct {
	shareholderTypeUseCase usecase.ShareholderTypeUseCase
}

// NewShareholderTypeHandler creates a new shareholder type handler
func NewShareholderTypeHandler(shareholderTypeUseCase usecase.ShareholderTypeUseCase) *ShareholderTypeHandler {
	return &ShareholderTypeHandler{
		shareholderTypeUseCase: shareholderTypeUseCase,
	}
}

// GetAllShareholderTypes handles getting all shareholder types
// @Summary      Ambil Semua Shareholder Types
// @Description  Mengambil daftar semua jenis pemegang saham yang aktif. User reguler hanya melihat jenis pemegang saham aktif, superadmin/administrator bisa melihat semua termasuk yang tidak aktif.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        include_inactive  query     bool    false  "Include inactive shareholder types (hanya superadmin/administrator)"
// @Success      200               {array}   domain.ShareholderTypeModel  "Daftar jenis pemegang saham berhasil diambil"
// @Failure      401               {object}  domain.ErrorResponse      "Unauthorized"
// @Failure      500               {object}  domain.ErrorResponse      "Internal server error"
// @Router       /api/v1/shareholder-types [get]
func (h *ShareholderTypeHandler) GetAllShareholderTypes(c *fiber.Ctx) error {
	roleVal := c.Locals("roleName")
	if roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	includeInactive := c.QueryBool("include_inactive", false)
	// Only superadmin/administrator can see inactive types
	if includeInactive && !utils.IsSuperAdminLike(roleName) {
		includeInactive = false
	}

	shareholderTypes, err := h.shareholderTypeUseCase.GetAllShareholderTypes(includeInactive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.JSON(shareholderTypes)
}

// CreateShareholderType handles creating a new shareholder type
// @Summary      Buat Shareholder Type Baru
// @Description  Membuat jenis pemegang saham baru. Hanya superadmin dan administrator yang dapat membuat jenis pemegang saham baru. Jika jenis pemegang saham dengan nama yang sama sudah ada (tidak aktif), akan diaktifkan kembali.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      object  true  "Shareholder type data"  example({"name": "Badan Hukum"})
// @Success      201      {object}  domain.ShareholderTypeModel  "Jenis pemegang saham berhasil dibuat"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request atau jenis pemegang saham sudah ada"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/shareholder-types [post]
func (h *ShareholderTypeHandler) CreateShareholderType(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	roleVal := c.Locals("roleName")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	// Only superadmin and administrator can create shareholder types
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat membuat jenis pemegang saham",
		})
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	shareholderType, err := h.shareholderTypeUseCase.CreateShareholderType(payload.Name, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionCreate, audit.ResourceCompany, shareholderType.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "create_shareholder_type",
		"name":      payload.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(shareholderType)
}

// UpdateShareholderType handles updating a shareholder type
// @Summary      Update Shareholder Type
// @Description  Mengupdate jenis pemegang saham. Hanya superadmin dan administrator yang dapat mengupdate jenis pemegang saham.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Shareholder Type ID"
// @Param        payload  body      object  true  "Shareholder type data"  example({"name": "Updated Name", "is_active": true})
// @Success      200      {object}  domain.ShareholderTypeModel  "Jenis pemegang saham berhasil diupdate"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/shareholder-types/{id} [put]
func (h *ShareholderTypeHandler) UpdateShareholderType(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDVal := c.Locals("userID")
	roleVal := c.Locals("roleName")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	// Only superadmin and administrator can update shareholder types
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat mengupdate jenis pemegang saham",
		})
	}

	var payload struct {
		Name     *string `json:"name"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	shareholderType, err := h.shareholderTypeUseCase.UpdateShareholderType(id, payload.Name, payload.IsActive)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionUpdate, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "update_shareholder_type",
		"name":      payload.Name,
		"is_active": payload.IsActive,
	})

	return c.JSON(shareholderType)
}

// DeleteShareholderType handles deleting a shareholder type
// @Summary      Hapus Shareholder Type
// @Description  Menghapus jenis pemegang saham. Hanya superadmin yang dapat menghapus secara permanen. Administrator hanya dapat soft delete (menonaktifkan). Jika jenis pemegang saham masih digunakan, akan dilakukan soft delete.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Shareholder Type ID"
// @Success      200  {object}  map[string]string  "Jenis pemegang saham berhasil dihapus"
// @Failure      400  {object}  domain.ErrorResponse  "Invalid request atau jenis pemegang saham masih digunakan"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/shareholder-types/{id} [delete]
func (h *ShareholderTypeHandler) DeleteShareholderType(c *fiber.Ctx) error {
	id := c.Params("id")
	userIDVal := c.Locals("userID")
	roleVal := c.Locals("roleName")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := fmt.Sprintf("%v", roleVal)

	// Only superadmin and administrator can delete shareholder types
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat menghapus jenis pemegang saham",
		})
	}

	if err := h.shareholderTypeUseCase.DeleteShareholderType(id, roleName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionDelete, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "delete_shareholder_type",
	})

	return c.JSON(fiber.Map{
		"message": "Jenis pemegang saham berhasil dihapus",
	})
}

