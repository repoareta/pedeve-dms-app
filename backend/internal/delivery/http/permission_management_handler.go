package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
)

// PermissionManagementHandler handles permission management HTTP requests
type PermissionManagementHandler struct {
	permissionUseCase usecase.PermissionManagementUseCase
}

// NewPermissionManagementHandler creates a new permission management handler
func NewPermissionManagementHandler(permissionUseCase usecase.PermissionManagementUseCase) *PermissionManagementHandler {
	return &PermissionManagementHandler{
		permissionUseCase: permissionUseCase,
	}
}

// CreatePermission handles permission creation
// @Summary      Buat Permission Baru
// @Description  Membuat permission baru. Hanya superadmin yang bisa membuat permission.
// @Tags         Permission Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        permission  body      object  true  "Permission data"
// @Success      201         {object}  domain.PermissionModel
// @Failure      400         {object}  domain.ErrorResponse
// @Failure      403         {object}  domain.ErrorResponse
// @Router       /api/v1/permissions [post]
func (h *PermissionManagementHandler) CreatePermission(c *fiber.Ctx) error {
	// Hanya superadmin/administrator yang boleh membuat permission baru
	roleName := c.Locals("roleName")
	if roleName == nil || !utils.IsSuperAdminLike(roleName.(string)) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin/administrator can create permissions",
		})
	}

	var req struct {
		Name        string                 `json:"name" validate:"required"`
		Description string                 `json:"description"`
		Resource    string                 `json:"resource" validate:"required"`
		Action      string                 `json:"action" validate:"required"`
		Scope       domain.PermissionScope `json:"scope" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	permission, err := h.permissionUseCase.CreatePermission(req.Name, req.Description, req.Resource, req.Action, req.Scope)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionCreate, audit.ResourcePermission, permission.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusCreated).JSON(permission)
}

// GetPermission handles getting permission by ID
// @Summary      Ambil Permission by ID
// @Description  Mengambil informasi permission berdasarkan ID.
// @Tags         Permission Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  domain.PermissionModel
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/permissions/{id} [get]
func (h *PermissionManagementHandler) GetPermission(c *fiber.Ctx) error {
	id := c.Params("id")
	permission, err := h.permissionUseCase.GetPermissionByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Permission not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(permission)
}

// GetAllPermissions handles getting all permissions
// @Summary      Ambil Semua Permissions
// @Description  Mengambil daftar semua permissions. Bisa difilter berdasarkan resource atau scope.
// @Tags         Permission Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        resource  query     string  false  "Filter by resource"
// @Param        scope     query     string  false  "Filter by scope (global, company, sub_company)"
// @Success      200       {array}   domain.PermissionModel
// @Router       /api/v1/permissions [get]
func (h *PermissionManagementHandler) GetAllPermissions(c *fiber.Ctx) error {
	roleName := c.Locals("roleName")
	if roleName == nil || !utils.IsSuperAdminLike(roleName.(string)) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin/administrator can view permissions",
		})
	}

	resource := c.Query("resource")
	scope := c.Query("scope")

	var permissions []domain.PermissionModel
	var err error

	if resource != "" {
		permissions, err = h.permissionUseCase.GetPermissionsByResource(resource)
	} else if scope != "" {
		permissions, err = h.permissionUseCase.GetPermissionsByScope(domain.PermissionScope(scope))
	} else {
		permissions, err = h.permissionUseCase.GetAllPermissions()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get permissions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(permissions)
}

// UpdatePermission handles permission update
// @Summary      Update Permission
// @Description  Mengupdate informasi permission.
// @Tags         Permission Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id          path      string  true  "Permission ID"
// @Param        permission  body      object  true  "Permission data to update"
// @Success      200         {object}  domain.PermissionModel
// @Failure      400         {object}  domain.ErrorResponse
// @Router       /api/v1/permissions/{id} [put]
func (h *PermissionManagementHandler) UpdatePermission(c *fiber.Ctx) error {
	// Hanya superadmin/administrator yang boleh mengubah permission
	roleName := c.Locals("roleName")
	if roleName == nil || !utils.IsSuperAdminLike(roleName.(string)) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin/administrator can update permissions",
		})
	}

	id := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	permission, err := h.permissionUseCase.UpdatePermission(id, req.Name, req.Description)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionUpdate, audit.ResourcePermission, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(permission)
}

// DeletePermission handles permission deletion
// @Summary      Hapus Permission
// @Description  Menghapus permission.
// @Tags         Permission Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Permission ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Router       /api/v1/permissions/{id} [delete]
func (h *PermissionManagementHandler) DeletePermission(c *fiber.Ctx) error {
	// Hanya superadmin/administrator yang boleh menghapus permission
	roleName := c.Locals("roleName")
	if roleName == nil || !utils.IsSuperAdminLike(roleName.(string)) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin/administrator can delete permissions",
		})
	}

	id := c.Params("id")
	if err := h.permissionUseCase.DeletePermission(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionDelete, audit.ResourcePermission, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Permission deleted successfully",
	})
}
