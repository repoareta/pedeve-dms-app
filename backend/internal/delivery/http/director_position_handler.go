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

// DirectorPositionHandler handles director position-related HTTP requests
type DirectorPositionHandler struct {
	directorPositionUseCase usecase.DirectorPositionUseCase
}

// NewDirectorPositionHandler creates a new director position handler
func NewDirectorPositionHandler(directorPositionUseCase usecase.DirectorPositionUseCase) *DirectorPositionHandler {
	return &DirectorPositionHandler{
		directorPositionUseCase: directorPositionUseCase,
	}
}

// GetAllDirectorPositions handles getting all director positions
// @Summary      Ambil Semua Director Positions
// @Description  Mengambil daftar semua jabatan pengurus yang aktif. User reguler hanya melihat jabatan aktif, superadmin/administrator bisa melihat semua termasuk yang tidak aktif.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        include_inactive  query     bool    false  "Include inactive director positions (hanya superadmin/administrator)"
// @Success      200               {array}   domain.DirectorPositionModel  "Daftar jabatan pengurus berhasil diambil"
// @Failure      401               {object}  domain.ErrorResponse      "Unauthorized"
// @Failure      500               {object}  domain.ErrorResponse      "Internal server error"
// @Router       /api/v1/director-positions [get]
func (h *DirectorPositionHandler) GetAllDirectorPositions(c *fiber.Ctx) error {
	roleVal := c.Locals("roleName")
	if roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	includeInactive := c.QueryBool("include_inactive", false)
	// Only superadmin/administrator can see inactive positions
	if includeInactive && !utils.IsSuperAdminLike(roleName) {
		includeInactive = false
	}

	directorPositions, err := h.directorPositionUseCase.GetAllDirectorPositions(includeInactive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.JSON(directorPositions)
}

// CreateDirectorPosition handles creating a new director position
// @Summary      Buat Director Position Baru
// @Description  Membuat jabatan pengurus baru. Hanya superadmin dan administrator yang dapat membuat jabatan baru. Jika jabatan dengan nama yang sama sudah ada (tidak aktif), akan diaktifkan kembali.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      object  true  "Director position data"  example({"name": "Direktur Utama"})
// @Success      201      {object}  domain.DirectorPositionModel  "Jabatan pengurus berhasil dibuat"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request atau jabatan sudah ada"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/director-positions [post]
func (h *DirectorPositionHandler) CreateDirectorPosition(c *fiber.Ctx) error {
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

	// Only superadmin and administrator can create director positions
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat membuat jabatan pengurus",
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

	directorPosition, err := h.directorPositionUseCase.CreateDirectorPosition(payload.Name, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionCreate, audit.ResourceCompany, directorPosition.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "create_director_position",
		"name":      payload.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(directorPosition)
}

// UpdateDirectorPosition handles updating a director position
// @Summary      Update Director Position
// @Description  Mengupdate jabatan pengurus. Hanya superadmin dan administrator yang dapat mengupdate jabatan.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Director Position ID"
// @Param        payload  body      object  true  "Director position data"  example({"name": "Updated Name", "is_active": true})
// @Success      200      {object}  domain.DirectorPositionModel  "Jabatan pengurus berhasil diupdate"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/director-positions/{id} [put]
func (h *DirectorPositionHandler) UpdateDirectorPosition(c *fiber.Ctx) error {
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

	// Only superadmin and administrator can update director positions
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat mengupdate jabatan pengurus",
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

	directorPosition, err := h.directorPositionUseCase.UpdateDirectorPosition(id, payload.Name, payload.IsActive)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionUpdate, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "update_director_position",
		"name":      payload.Name,
		"is_active": payload.IsActive,
	})

	return c.JSON(directorPosition)
}

// DeleteDirectorPosition handles deleting a director position
// @Summary      Hapus Director Position
// @Description  Menghapus jabatan pengurus. Hanya superadmin yang dapat menghapus secara permanen. Administrator hanya dapat soft delete (menonaktifkan). Jika jabatan masih digunakan, akan dilakukan soft delete.
// @Tags         Company Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Director Position ID"
// @Success      200  {object}  map[string]string  "Jabatan pengurus berhasil dihapus"
// @Failure      400  {object}  domain.ErrorResponse  "Invalid request atau jabatan masih digunakan"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/director-positions/{id} [delete]
func (h *DirectorPositionHandler) DeleteDirectorPosition(c *fiber.Ctx) error {
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

	// Only superadmin and administrator can delete director positions
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat menghapus jabatan pengurus",
		})
	}

	if err := h.directorPositionUseCase.DeleteDirectorPosition(id, roleName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionDelete, audit.ResourceCompany, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "delete_director_position",
	})

	return c.JSON(fiber.Map{
		"message": "Jabatan pengurus berhasil dihapus",
	})
}

