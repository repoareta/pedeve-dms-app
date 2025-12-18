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

// DocumentTypeHandler handles document type-related HTTP requests
type DocumentTypeHandler struct {
	docTypeUseCase usecase.DocumentTypeUseCase
}

// NewDocumentTypeHandler creates a new document type handler
func NewDocumentTypeHandler(docTypeUseCase usecase.DocumentTypeUseCase) *DocumentTypeHandler {
	return &DocumentTypeHandler{
		docTypeUseCase: docTypeUseCase,
	}
}

// GetAllDocumentTypes handles getting all document types
// @Summary      Ambil Semua Document Types
// @Description  Mengambil daftar semua jenis dokumen yang aktif. User reguler hanya melihat jenis dokumen aktif, superadmin/administrator bisa melihat semua termasuk yang tidak aktif.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        include_inactive  query     bool    false  "Include inactive document types (hanya superadmin/administrator)"
// @Success      200               {array}   domain.DocumentTypeModel  "Daftar jenis dokumen berhasil diambil"
// @Failure      401               {object}  domain.ErrorResponse      "Unauthorized"
// @Failure      500               {object}  domain.ErrorResponse      "Internal server error"
// @Router       /api/v1/document-types [get]
func (h *DocumentTypeHandler) GetAllDocumentTypes(c *fiber.Ctx) error {
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

	docTypes, err := h.docTypeUseCase.GetAllDocumentTypes(includeInactive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.JSON(docTypes)
}

// CreateDocumentType handles creating a new document type
// @Summary      Buat Document Type Baru
// @Description  Membuat jenis dokumen baru. Hanya superadmin dan administrator yang dapat membuat jenis dokumen baru. Jika jenis dokumen dengan nama yang sama sudah ada (tidak aktif), akan diaktifkan kembali.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      object  true  "Document type data"  example({"name": "RUPS"})
// @Success      201      {object}  domain.DocumentTypeModel  "Jenis dokumen berhasil dibuat"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request atau jenis dokumen sudah ada"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Router       /api/v1/document-types [post]
func (h *DocumentTypeHandler) CreateDocumentType(c *fiber.Ctx) error {
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

	// Only superadmin and administrator can create document types
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat membuat jenis dokumen",
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

	docType, err := h.docTypeUseCase.CreateDocumentType(payload.Name, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionCreateDoc, audit.ResourceDocument, docType.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "create_document_type",
		"name":      payload.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(docType)
}

// UpdateDocumentType handles updating a document type
// @Summary      Update Document Type
// @Description  Mengupdate jenis dokumen (nama atau status aktif/tidak aktif). Hanya superadmin dan administrator yang dapat mengupdate jenis dokumen.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Document Type ID"
// @Param        payload  body      object  true  "Document type data"  example({"name": "Updated Name", "is_active": true})
// @Success      200      {object}  domain.DocumentTypeModel  "Jenis dokumen berhasil diupdate"
// @Failure      400      {object}  domain.ErrorResponse       "Invalid request"
// @Failure      401      {object}  domain.ErrorResponse       "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse       "Forbidden (hanya superadmin/administrator)"
// @Failure      404      {object}  domain.ErrorResponse       "Document type tidak ditemukan"
// @Router       /api/v1/document-types/{id} [put]
func (h *DocumentTypeHandler) UpdateDocumentType(c *fiber.Ctx) error {
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

	// Only superadmin and administrator can update document types
	if !utils.IsSuperAdminLike(roleName) {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Hanya superadmin dan administrator yang dapat mengupdate jenis dokumen",
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

	// At least one field must be provided
	if payload.Name == nil && payload.IsActive == nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Minimal satu field (name atau is_active) harus diisi",
		})
	}

	docType, err := h.docTypeUseCase.UpdateDocumentType(id, payload.Name, payload.IsActive)
	if err != nil {
		status := fiber.StatusBadRequest
		if strings.Contains(err.Error(), "tidak ditemukan") {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionUpdateDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "update_document_type",
		"name":      payload.Name,
		"is_active": payload.IsActive,
	})

	return c.JSON(docType)
}

// DeleteDocumentType handles deleting a document type (soft delete)
// @Summary      Hapus Document Type
// @Description  Menghapus jenis dokumen (soft delete: set is_active = false). Hanya superadmin dan administrator yang dapat menghapus. Jika jenis dokumen sudah digunakan oleh dokumen, akan di-soft delete (tidak aktif) agar dokumen yang sudah ada tidak error.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Document Type ID"
// @Success      200  {object}  map[string]string  "Jenis dokumen berhasil dihapus"
// @Failure      400  {object}  domain.ErrorResponse  "Invalid request"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden (hanya superadmin/administrator)"
// @Failure      404  {object}  domain.ErrorResponse  "Document type tidak ditemukan"
// @Router       /api/v1/document-types/{id} [delete]
func (h *DocumentTypeHandler) DeleteDocumentType(c *fiber.Ctx) error {
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

	err := h.docTypeUseCase.DeleteDocumentType(id, roleName)
	if err != nil {
		status := fiber.StatusBadRequest
		if strings.Contains(err.Error(), "tidak ditemukan") {
			status = fiber.StatusNotFound
		} else if strings.Contains(err.Error(), "hanya superadmin") {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionDeleteDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "delete_document_type",
	})

	return c.JSON(fiber.Map{
		"message": "Jenis dokumen berhasil dihapus (soft delete)",
	})
}

