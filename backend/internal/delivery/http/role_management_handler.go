package http

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// RoleManagementHandler handles role management HTTP requests
type RoleManagementHandler struct {
	roleUseCase usecase.RoleManagementUseCase
}

// NewRoleManagementHandler creates a new role management handler
func NewRoleManagementHandler(roleUseCase usecase.RoleManagementUseCase) *RoleManagementHandler {
	return &RoleManagementHandler{
		roleUseCase: roleUseCase,
	}
}

// CreateRole handles role creation
// @Summary      Buat Role Baru
// @Description  Membuat role baru. Hanya superadmin yang bisa membuat role.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        role  body      object  true  "Role data"
// @Success      201   {object}  domain.RoleModel
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      403   {object}  domain.ErrorResponse
// @Router       /api/v1/roles [post]
func (h *RoleManagementHandler) CreateRole(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Level       int    `json:"level" validate:"required,min=0,max=10"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	role, err := h.roleUseCase.CreateRole(req.Name, req.Description, req.Level)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionCreate, audit.ResourceRole, role.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusCreated).JSON(role)
}

// GetRole handles getting role by ID
// @Summary      Ambil Role by ID
// @Description  Mengambil informasi role berdasarkan ID.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  domain.RoleModel
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/roles/{id} [get]
func (h *RoleManagementHandler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")
	role, err := h.roleUseCase.GetRoleByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Role not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(role)
}

// GetAllRoles handles getting all roles
// @Summary      Ambil Semua Roles
// @Description  Mengambil daftar semua roles.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.RoleModel
// @Router       /api/v1/roles [get]
func (h *RoleManagementHandler) GetAllRoles(c *fiber.Ctx) error {
	roles, err := h.roleUseCase.GetAllRoles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get roles",
		})
	}

	return c.Status(fiber.StatusOK).JSON(roles)
}

// UpdateRole handles role update
// @Summary      Update Role
// @Description  Mengupdate informasi role. System roles tidak bisa diupdate.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Role ID"
// @Param        role  body      object  true  "Role data to update"
// @Success      200   {object}  domain.RoleModel
// @Failure      400   {object}  domain.ErrorResponse
// @Router       /api/v1/roles/{id} [put]
func (h *RoleManagementHandler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Level       int    `json:"level"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	role, err := h.roleUseCase.UpdateRole(id, req.Name, req.Description, req.Level)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionUpdate, audit.ResourceRole, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(role)
}

// DeleteRole handles role deletion
// @Summary      Hapus Role
// @Description  Menghapus role. System roles tidak bisa dihapus.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleManagementHandler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.roleUseCase.DeleteRole(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionDelete, audit.ResourceRole, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Role deleted successfully",
	})
}

// GetRolePermissions handles getting role permissions
// @Summary      Ambil Permissions Role
// @Description  Mengambil daftar permissions yang dimiliki oleh role.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {array}   domain.PermissionModel
// @Router       /api/v1/roles/{id}/permissions [get]
func (h *RoleManagementHandler) GetRolePermissions(c *fiber.Ctx) error {
	id := c.Params("id")
	permissions, err := h.roleUseCase.GetRolePermissions(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get role permissions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(permissions)
}

// AssignPermissionToRole handles assigning permission to role
// @Summary      Assign Permission ke Role
// @Description  Menambahkan permission ke role.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id            path      string  true  "Role ID"
// @Param        permission_id body      object  true  "Permission ID"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  domain.ErrorResponse
// @Router       /api/v1/roles/{id}/permissions [post]
func (h *RoleManagementHandler) AssignPermissionToRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		PermissionID string `json:"permission_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.roleUseCase.AssignPermissionToRole(id, req.PermissionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "assignment_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "assign_permission", audit.ResourceRole, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"permission_id": req.PermissionID,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Permission assigned successfully",
	})
}

// RevokePermissionFromRole handles revoking permission from role
// @Summary      Revoke Permission dari Role
// @Description  Menghapus permission dari role.
// @Tags         Role Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id            path      string  true  "Role ID"
// @Param        permission_id body      object  true  "Permission ID"
// @Success      200           {object}  map[string]string
// @Failure      400           {object}  domain.ErrorResponse
// @Router       /api/v1/roles/{id}/permissions [delete]
func (h *RoleManagementHandler) RevokePermissionFromRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		PermissionID string `json:"permission_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.roleUseCase.RevokePermissionFromRole(id, req.PermissionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "revoke_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "revoke_permission", audit.ResourceRole, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"permission_id": req.PermissionID,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Permission revoked successfully",
	})
}

