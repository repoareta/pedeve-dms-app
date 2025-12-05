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
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
)

type DocumentHandler struct {
	docUseCase usecase.DocumentUseCase
}

func NewDocumentHandler(docUseCase usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{docUseCase: docUseCase}
}

// ListFolders handles getting all folders
// @Summary      Ambil Semua Folders
// @Description  Mengambil daftar semua folder. Superadmin/administrator melihat semua folder, user reguler hanya melihat folder milik mereka sendiri.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.DocumentFolderModel  "Daftar folder berhasil diambil"
// @Failure      401  {object}  domain.ErrorResponse         "Unauthorized"
// @Failure      500  {object}  domain.ErrorResponse         "Internal server error"
// @Router       /api/v1/documents/folders [get]
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
	if !utils.IsSuperAdminLike(roleName) {
		ownerFilter = &userIDStr
	}

	folders, err := h.docUseCase.ListFolders(ownerFilter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}
	if audit.ShouldLogView() {
		username, _ := c.Locals("username").(string)
		audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"operation": "list_folders",
		})
	}
	return c.JSON(folders)
}

// CreateFolder handles creating a new folder
// @Summary      Buat Folder Baru
// @Description  Membuat folder baru. Folder dapat memiliki parent folder (opsional) untuk membuat struktur hierarki.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      object  true  "Folder data"  example({"name": "New Folder", "parent_id": "optional-parent-id"})
// @Success      201      {object}  domain.DocumentFolderModel  "Folder berhasil dibuat"
// @Failure      400      {object}  domain.ErrorResponse         "Invalid request"
// @Failure      401      {object}  domain.ErrorResponse         "Unauthorized"
// @Router       /api/v1/documents/folders [post]
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

// UpdateFolder handles updating folder name
// @Summary      Update Nama Folder
// @Description  Mengubah nama folder. Hanya owner folder atau superadmin/administrator yang dapat mengubah nama folder.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Folder ID"
// @Param        payload  body      object  true  "Folder data"  example({"name": "Updated Folder Name"})
// @Success      200      {object}  domain.DocumentFolderModel  "Folder berhasil diupdate"
// @Failure      400      {object}  domain.ErrorResponse         "Invalid request"
// @Failure      401      {object}  domain.ErrorResponse         "Unauthorized"
// @Failure      403      {object}  domain.ErrorResponse         "Forbidden"
// @Router       /api/v1/documents/folders/{id} [put]
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

// DeleteFolder handles deleting a folder and all its documents
// @Summary      Hapus Folder
// @Description  Menghapus folder beserta semua dokumen di dalamnya. Hanya owner folder atau superadmin/administrator yang dapat menghapus folder.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Folder ID"
// @Success      200  {object}  map[string]string  "Folder berhasil dihapus"
// @Failure      400  {object}  domain.ErrorResponse  "Invalid request"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden"
// @Router       /api/v1/documents/folders/{id} [delete]
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

// ListDocuments handles getting all documents with optional pagination and filters
// @Summary      Ambil Semua Documents
// @Description  Mengambil daftar dokumen dengan opsi pagination, search, sort, dan filter. Superadmin/administrator melihat semua dokumen, user reguler hanya melihat dokumen milik mereka sendiri.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        folder_id  query     string  false  "Filter by folder ID"
// @Param        page       query     int     false  "Page number (default: 0, no pagination)"
// @Param        page_size  query     int     false  "Page size (default: 0, no pagination, max: 100)"
// @Param        search     query     string  false  "Search by title or filename"
// @Param        sort_by    query     string  false  "Sort field (created_at, title, size)"
// @Param        sort_dir   query     string  false  "Sort direction (asc, desc)"
// @Param        type       query     string  false  "Filter by file type (pdf, docx, xlsx, etc)"
// @Success      200        {array}   domain.DocumentModel  "Daftar dokumen (tanpa pagination) atau {object} dengan data, total, page, page_size (dengan pagination). Metadata field menggunakan format JSON."
// @Failure      401        {object}  domain.ErrorResponse   "Unauthorized"
// @Failure      403        {object}  domain.ErrorResponse   "Forbidden"
// @Failure      500        {object}  domain.ErrorResponse   "Internal server error"
// @Router       /api/v1/documents [get]
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
	if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
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
	if !utils.IsSuperAdminLike(roleName) {
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

		if audit.ShouldLogView() {
			username, _ := c.Locals("username").(string)
			audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
				"operation": "list_documents_paginated",
				"folder_id": folderID,
				"page":      page,
				"page_size": pageSize,
			})
		}
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

	// RBAC: non-superadmin/administrator hanya melihat dokumen milik sendiri
	if !utils.IsSuperAdminLike(roleName) {
		filtered := make([]domain.DocumentModel, 0, len(docs))
		for _, d := range docs {
			if d.UploaderID == userIDStr {
				filtered = append(filtered, d)
			}
		}
		docs = filtered
	}

	if audit.ShouldLogView() {
		username, _ := c.Locals("username").(string)
		audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, "", getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
			"operation": "list_documents",
			"folder_id": folderID,
		})
	}
	return c.JSON(docs)
}

// DocumentSummary returns folder statistics and total storage
// @Summary      Ambil Ringkasan Documents
// @Description  Mengembalikan statistik folder (file_count, total_size) dan total storage. Data bersifat global untuk semua dokumen.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Ringkasan dokumen berhasil diambil. Response berisi folder_stats (array) dan total_size (number)"
// @Failure      401  {object}  domain.ErrorResponse     "Unauthorized"
// @Failure      500  {object}  domain.ErrorResponse     "Internal server error"
// @Router       /api/v1/documents/summary [get]
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

// GetDocument handles getting a document by ID
// @Summary      Ambil Document by ID
// @Description  Mengambil detail dokumen berdasarkan ID. Superadmin/administrator dapat melihat semua dokumen, user reguler hanya dapat melihat dokumen milik mereka sendiri.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Document ID"
// @Success      200  {object}  domain.DocumentModel  "Detail dokumen berhasil diambil"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden"
// @Failure      404  {object}  domain.ErrorResponse  "Document tidak ditemukan"
// @Router       /api/v1/documents/{id} [get]
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

	if !utils.IsSuperAdminLike(roleName) && doc.UploaderID != userIDStr {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Anda tidak memiliki akses ke dokumen ini",
		})
	}

	if audit.ShouldLogView() {
		username, _ := c.Locals("username").(string)
		audit.LogAction(userIDStr, username, audit.ActionViewDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)
	}
	return c.JSON(doc)
}

// UploadDocument handles uploading a new document
// @Summary      Upload Document Baru
// @Description  Mengupload dokumen baru dengan file, title, status, dan metadata. File wajib diupload dalam format multipart/form-data.
// @Tags         Documents
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file       formData  file    true   "File dokumen yang akan diupload"
// @Param        folder_id  formData  string  false  "ID folder (opsional)"
// @Param        title      formData  string  false  "Judul dokumen (opsional)"
// @Param        status     formData  string  false  "Status dokumen (default: active)"
// @Param        metadata   formData  string  false  "Metadata dalam format JSON string (opsional)"
// @Success      201        {object}  domain.DocumentModel  "Dokumen berhasil diupload"
// @Failure      400        {object}  domain.ErrorResponse   "Invalid request atau file tidak valid"
// @Failure      401        {object}  domain.ErrorResponse   "Unauthorized"
// @Failure      403        {object}  domain.ErrorResponse   "Forbidden (tidak memiliki akses ke folder)"
// @Failure      404        {object}  domain.ErrorResponse   "Folder tidak ditemukan"
// @Router       /api/v1/documents/upload [post]
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
	if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
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

// UpdateDocument handles updating document metadata and optionally replacing the file
// @Summary      Update Document
// @Description  Mengupdate metadata dokumen (folder_id, title, status, metadata) dan opsional mengganti file. Mendukung JSON payload (metadata only) atau multipart/form-data (dengan file). Hanya owner dokumen atau superadmin/administrator yang dapat mengupdate.
// @Tags         Documents
// @Accept       json,multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true   "Document ID"
// @Param        payload    body      object  false  "JSON payload untuk update metadata only"  example({"folder_id": "folder-id", "title": "Updated Title", "status": "active", "metadata": {}})
// @Param        file       formData  file    false  "File baru (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        folder_id  formData  string  false  "ID folder (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        title      formData  string  false  "Judul dokumen (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        status     formData  string  false  "Status dokumen (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        metadata   formData  string  false  "Metadata dalam format JSON string (opsional, hanya jika Content-Type: multipart/form-data)"
// @Success      200        {object}  domain.DocumentModel  "Dokumen berhasil diupdate"
// @Failure      400        {object}  domain.ErrorResponse   "Invalid request"
// @Failure      401        {object}  domain.ErrorResponse   "Unauthorized"
// @Failure      403        {object}  domain.ErrorResponse   "Forbidden"
// @Failure      404        {object}  domain.ErrorResponse   "Document atau folder tidak ditemukan"
// @Router       /api/v1/documents/{id} [put]
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
	if !utils.IsSuperAdminLike(roleName) && existingDoc.UploaderID != userIDStr {
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
		if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
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

	if payload.FolderID != nil && !utils.IsSuperAdminLike(roleName) {
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

// DeleteDocument handles deleting a document
// @Summary      Hapus Document
// @Description  Menghapus dokumen. Hanya owner dokumen atau superadmin/administrator yang dapat menghapus dokumen.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Document ID"
// @Success      200  {object}  map[string]string  "Dokumen berhasil dihapus"
// @Failure      401  {object}  domain.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  domain.ErrorResponse  "Forbidden"
// @Failure      404  {object}  domain.ErrorResponse  "Document tidak ditemukan"
// @Failure      500  {object}  domain.ErrorResponse  "Internal server error"
// @Router       /api/v1/documents/{id} [delete]
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
	if !utils.IsSuperAdminLike(roleName) && existingDoc.UploaderID != userIDStr {
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
