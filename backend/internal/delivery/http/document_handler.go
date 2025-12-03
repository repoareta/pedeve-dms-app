package http

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
)

type DocumentHandler struct {
	docUseCase usecase.DocumentUseCase
}

func NewDocumentHandler(docUseCase usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{docUseCase: docUseCase}
}

// List folders
func (h *DocumentHandler) ListFolders(c *fiber.Ctx) error {
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

	var ownerFilter *string
	if roleName != "superadmin" {
		ownerFilter = &userIDStr
	}

	folders, err := h.docUseCase.ListFolders(ownerFilter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}
	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "list_folders",
	})
	return c.JSON(folders)
}

// Create folder
func (h *DocumentHandler) CreateFolder(c *fiber.Ctx) error {
	var payload struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)

	folder, err := h.docUseCase.CreateFolder(payload.Name, payload.ParentID, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}
	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionCreateDoc, audit.ResourceDocument, folder.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "create_folder",
		"name":      payload.Name,
		"parent_id": payload.ParentID,
	})
	return c.Status(fiber.StatusCreated).JSON(folder)
}

// Update folder name
func (h *DocumentHandler) UpdateFolder(c *fiber.Ctx) error {
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

	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	folder, err := h.docUseCase.UpdateFolderName(id, payload.Name, userIDStr, roleName)
	if err != nil {
		status := fiber.StatusBadRequest
		if strings.Contains(err.Error(), "forbidden") {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionUpdateDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "rename_folder",
		"name":      payload.Name,
	})

	return c.JSON(folder)
}

// Delete folder and its documents
func (h *DocumentHandler) DeleteFolder(c *fiber.Ctx) error {
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

	if err := h.docUseCase.DeleteFolder(id, userIDStr, roleName); err != nil {
		status := fiber.StatusBadRequest
		if strings.Contains(err.Error(), "forbidden") {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionDeleteDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "delete_folder",
	})

	return c.JSON(fiber.Map{
		"message": "Folder dan seluruh file di dalamnya telah dihapus",
	})
}

// List documents
func (h *DocumentHandler) ListDocuments(c *fiber.Ctx) error {
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

	folderID := c.Query("folder_id")
	var folderPtr *string
	if folderID != "" {
		folderPtr = &folderID
	}

	// Jika filter folder diberikan, pastikan kepemilikan folder untuk non-superadmin
	if folderPtr != nil && roleName != "superadmin" {
		folder, err := h.docUseCase.GetFolderByID(folderID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		if folder.CreatedBy != userIDStr {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke folder ini",
			})
		}
	}

	page := c.QueryInt("page", 0)
	pageSize := c.QueryInt("page_size", 0)
	search := strings.TrimSpace(c.Query("search"))
	sortBy := strings.TrimSpace(c.Query("sort_by"))
	sortDir := strings.TrimSpace(c.Query("sort_dir"))
	typeFilter := strings.TrimSpace(c.Query("type"))

	usePaginated := page > 0 || pageSize > 0 || search != "" || sortBy != "" || sortDir != "" || typeFilter != ""

	ownerFilter := (*string)(nil)
	if roleName != "superadmin" {
		ownerFilter = &userIDStr
	}

	if usePaginated {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}

		docs, total, err := h.docUseCase.ListDocumentsPaginated(usecase.ListDocumentsParams{
			FolderID:   folderPtr,
			Search:     search,
			SortBy:     sortBy,
			SortDir:    sortDir,
			Page:       page,
			PageSize:   pageSize,
			OwnerID:    ownerFilter,
			TypeFilter: typeFilter,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
		}

		username, _ := c.Locals("username").(string)
		audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"operation": "list_documents_paginated",
			"folder_id": folderID,
			"page":      page,
			"page_size": pageSize,
		})
		return c.JSON(fiber.Map{
			"data":      docs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}

	docs, err := h.docUseCase.ListDocuments(folderPtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	// RBAC: non-superadmin hanya melihat dokumen milik sendiri
	if roleName != "superadmin" {
		filtered := make([]domain.DocumentModel, 0, len(docs))
		for _, d := range docs {
			if d.UploaderID == userIDStr {
				filtered = append(filtered, d)
			}
		}
		docs = filtered
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "list_documents",
		"folder_id": folderID,
	})
	return c.JSON(docs)
}

// DocumentSummary mengembalikan statistik folder (file_count, total_size) dan total storage
func (h *DocumentHandler) DocumentSummary(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	roleVal := c.Locals("roleName")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	// Storage bersifat global: tampilkan penggunaan seluruh dokumen
	// (tidak difilter per user), sesuai kebutuhan dashboard storage.
	ownerFilter := (*string)(nil)

	stats, total, err := h.docUseCase.GetDocumentSummary(ownerFilter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"folder_stats": stats,
		"total_size":   total,
	})
}

// Get document by ID
func (h *DocumentHandler) GetDocument(c *fiber.Ctx) error {
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

	doc, err := h.docUseCase.GetDocumentByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Document tidak ditemukan",
		})
	}

	if roleName != "superadmin" && doc.UploaderID != userIDStr {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda tidak memiliki akses ke dokumen ini",
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)
	return c.JSON(doc)
}

// Upload document (multipart)
func (h *DocumentHandler) UploadDocument(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User tidak ditemukan",
		})
	}
	uploaderID := fmt.Sprintf("%v", userID)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "File wajib diupload",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_error",
			Message: "Gagal membaca file",
		})
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "file_error",
			Message: "Gagal membaca file",
		})
	}

	folderID := c.FormValue("folder_id")
	var folderPtr *string
	if folderID != "" {
		folderPtr = &folderID
	}

	// Jika folder ditentukan, pastikan kepemilikan folder untuk non-superadmin
	roleVal := c.Locals("roleName")
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))
	if folderPtr != nil && roleName != "superadmin" {
		folder, err := h.docUseCase.GetFolderByID(*folderPtr)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		if folder.CreatedBy != uploaderID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke folder ini",
			})
		}
	}

	title := c.FormValue("title")
	status := c.FormValue("status")
	if status == "" {
		status = "active"
	}
	metadata := c.FormValue("metadata")
	metaMap := map[string]interface{}{}
	if metadata != "" {
		_ = json.Unmarshal([]byte(metadata), &metaMap)
	}

	doc, err := h.docUseCase.UploadDocument(usecase.UploadDocumentInput{
		FolderID:    folderPtr,
		Title:       title,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Data:        data,
		Size:        fileHeader.Size,
		Status:      status,
		UploaderID:  uploaderID,
		Metadata:    metaMap,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(uploaderID, username, audit.ActionCreateDoc, audit.ResourceDocument, doc.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "upload_document",
		"folder_id": folderID,
		"title":     title,
		"status":    status,
	})

	return c.Status(fiber.StatusCreated).JSON(doc)
}

// Update document metadata
func (h *DocumentHandler) UpdateDocument(c *fiber.Ctx) error {
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

	// Ambil dokumen untuk cek kepemilikan
	existingDoc, err := h.docUseCase.GetDocumentByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Document tidak ditemukan",
		})
	}
	if roleName != "superadmin" && existingDoc.UploaderID != userIDStr {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda tidak memiliki akses mengubah dokumen ini",
		})
	}

	contentType := c.Get("Content-Type")

	// Jika multipart, izinkan update file sekaligus metadata
	if contentType != "" && strings.HasPrefix(strings.ToLower(contentType), "multipart/") {
		// Parse optional file
		fileHeader, _ := c.FormFile("file") // optional
		var data []byte
		var fname *string
		var ftype *string
		var fsize *int64

		if fileHeader != nil {
			file, err := fileHeader.Open()
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
					Error:   "file_error",
					Message: "Gagal membaca file",
				})
			}
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
					Error:   "file_error",
					Message: "Gagal membaca file",
				})
			}
			name := fileHeader.Filename
			fname = &name
			ct := fileHeader.Header.Get("Content-Type")
			ftype = &ct
			size := fileHeader.Size
			fsize = &size
		}

		var folderPtr *string
		if v := c.FormValue("folder_id"); v != "" {
			folderPtr = &v
		}
		if folderPtr != nil && roleName != "superadmin" {
			folder, err := h.docUseCase.GetFolderByID(*folderPtr)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
					Error:   "not_found",
					Message: "Folder tidak ditemukan",
				})
			}
			if folder.CreatedBy != userIDStr {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "Anda tidak memiliki akses ke folder ini",
				})
			}
		}
		var titlePtr *string
		if v := c.FormValue("title"); v != "" {
			titlePtr = &v
		}
		var statusPtr *string
		if v := c.FormValue("status"); v != "" {
			statusPtr = &v
		}
		metaMap := map[string]interface{}{}
		if metadata := c.FormValue("metadata"); metadata != "" {
			_ = json.Unmarshal([]byte(metadata), &metaMap)
		}

		doc, err := h.docUseCase.UpdateDocument(id, usecase.UpdateDocumentInput{
			FolderID:        folderPtr,
			Title:           titlePtr,
			Status:          statusPtr,
			Metadata:        metaMap,
			FileName:        fname,
			FileContentType: ftype,
			FileData:        data,
			FileSize:        fsize,
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
				Error:   "update_failed",
				Message: err.Error(),
			})
		}
		username, _ := c.Locals("username").(string)
		audit.LogAction(userIDStr, username, audit.ActionUpdateDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"operation":     "update_document",
			"folder_id":     folderPtr,
			"title":         titlePtr,
			"status":        statusPtr,
			"file_replaced": fileHeader != nil,
		})
		return c.JSON(doc)
	}

	// JSON payload (metadata only)
	var payload struct {
		FolderID *string                `json:"folder_id"`
		Title    *string                `json:"title"`
		Status   *string                `json:"status"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	if payload.FolderID != nil && roleName != "superadmin" {
		folder, err := h.docUseCase.GetFolderByID(*payload.FolderID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		if folder.CreatedBy != userIDStr {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke folder ini",
			})
		}
	}

	doc, err := h.docUseCase.UpdateDocument(id, usecase.UpdateDocumentInput{
		FolderID: payload.FolderID,
		Title:    payload.Title,
		Status:   payload.Status,
		Metadata: payload.Metadata,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionUpdateDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "update_document",
		"folder_id": payload.FolderID,
		"title":     payload.Title,
		"status":    payload.Status,
	})

	return c.JSON(doc)
}

// Delete document
func (h *DocumentHandler) DeleteDocument(c *fiber.Ctx) error {
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

	// Pastikan hanya pemilik atau superadmin
	existingDoc, err := h.docUseCase.GetDocumentByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Document tidak ditemukan",
		})
	}
	if roleName != "superadmin" && existingDoc.UploaderID != userIDStr {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda tidak memiliki akses menghapus dokumen ini",
		})
	}

	if err := h.docUseCase.DeleteDocument(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}
	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionDeleteDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation": "delete_document",
	})
	return c.JSON(fiber.Map{
		"message": "Document deleted",
	})
}
