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
	companyIDVal := c.Locals("companyID")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	var companyFilter *string
	if !utils.IsSuperAdminLike(roleName) {
		// User reguler hanya melihat folder perusahaan mereka
		if companyIDVal != nil {
			var userCompanyID string
			if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
				userCompanyID = *companyIDPtr
				companyFilter = &userCompanyID
			} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
				companyFilter = &companyIDStr
			}
		}
		// Jika user tidak punya company assignment, tidak akan melihat folder apa-apa
	}

	folders, err := h.docUseCase.ListFolders(companyFilter)
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
	roleVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	var userCompanyID *string
	if companyIDVal != nil {
		if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
			userCompanyID = &companyIDStr
		}
	}

	// Non-superadmin harus punya company assignment
	if !utils.IsSuperAdminLike(roleName) && userCompanyID == nil {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "User harus terasosiasi dengan perusahaan untuk membuat folder",
		})
	}

	// Superadmin/administrator bisa membuat folder tanpa company (opsional)
	// User reguler otomatis menggunakan company mereka
	folder, err := h.docUseCase.CreateFolder(payload.Name, userCompanyID, payload.ParentID, userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}
	username, _ := c.Locals("username").(string)
	audit.LogAction(userIDStr, username, audit.ActionCreateDoc, audit.ResourceDocument, folder.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"operation":  "create_folder",
		"name":       payload.Name,
		"parent_id":  payload.ParentID,
		"company_id": userCompanyID,
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
	companyIDVal := c.Locals("companyID")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	var userCompanyID *string
	if companyIDVal != nil {
		if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
			userCompanyID = &companyIDStr
		}
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

	folder, err := h.docUseCase.UpdateFolderName(id, payload.Name, userCompanyID, roleName)
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
	companyIDVal := c.Locals("companyID")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	userIDStr := fmt.Sprintf("%v", userIDVal)
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	// Superadmin/administrator dapat menghapus semua folder tanpa check access
	isSuperAdmin := utils.IsSuperAdminLike(roleName)
	var userCompanyID *string

	if !isSuperAdmin {
		// Non-superadmin perlu check access berdasarkan company_id
		if companyIDVal != nil {
			if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
				userCompanyID = companyIDPtr
			} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
				userCompanyID = &companyIDStr
			}
		}
	}
	// Superadmin bisa pass nil untuk userCompanyID dan tetap bisa delete

	if err := h.docUseCase.DeleteFolder(id, userCompanyID, roleName); err != nil {
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
// @Param        folder_id   query     string  false  "Filter by folder ID"
// @Param        director_id query     string  false  "Filter by director ID (untuk dokumen individu)"
// @Param        page        query     int     false  "Page number (default: 0, no pagination)"
// @Param        page_size   query     int     false  "Page size (default: 0, no pagination, max: 100)"
// @Param        search      query     string  false  "Search by title or filename"
// @Param        sort_by     query     string  false  "Sort field (created_at, title, size)"
// @Param        sort_dir    query     string  false  "Sort direction (asc, desc)"
// @Param        type        query     string  false  "Filter by file type (pdf, docx, xlsx, etc)"
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

	directorID := c.Query("director_id")
	var directorPtr *string
	if directorID != "" {
		directorPtr = &directorID
	}

	// Jika filter folder diberikan, pastikan akses folder untuk non-superadmin
	companyIDValForList := c.Locals("companyID")
	var userCompanyID *string
	if companyIDValForList != nil {
		if companyIDPtr, ok := companyIDValForList.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyIDValForList.(string); ok && companyIDStr != "" {
			userCompanyID = &companyIDStr
		}
	}

	if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
		folder, err := h.docUseCase.GetFolderByID(folderID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		// Check jika folder milik company user
		if userCompanyID == nil || folder.CompanyID == nil || *userCompanyID != *folder.CompanyID {
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

	// Filter berdasarkan company_id untuk non-superadmin
	var companyFilter *string
	if !utils.IsSuperAdminLike(roleName) {
		companyFilter = userCompanyID
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
			CompanyID:  companyFilter,
			DirectorID: directorPtr,
			Search:     search,
			SortBy:     sortBy,
			SortDir:    sortDir,
			Page:       page,
			PageSize:   pageSize,
			OwnerID:    nil, // Tidak filter berdasarkan uploader, hanya company
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

	// RBAC: non-superadmin/administrator hanya melihat dokumen di folder company mereka
	if !utils.IsSuperAdminLike(roleName) && userCompanyID != nil {
		// Filter: hanya dokumen di folder yang memiliki company_id sesuai
		filtered := make([]domain.DocumentModel, 0, len(docs))
		for _, d := range docs {
			if d.Folder != nil && d.Folder.CompanyID != nil && *d.Folder.CompanyID == *userCompanyID {
				filtered = append(filtered, d)
			} else if d.FolderID != nil {
				// Jika folder tidak di-load, cek folder secara terpisah
				folder, err := h.docUseCase.GetFolderByID(*d.FolderID)
				if err == nil && folder.CompanyID != nil && *folder.CompanyID == *userCompanyID {
					filtered = append(filtered, d)
				}
			}
		}
		docs = filtered
	} else if !utils.IsSuperAdminLike(roleName) {
		// User tanpa company tidak melihat dokumen apapun
		docs = []domain.DocumentModel{}
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
	companyIDVal := c.Locals("companyID")
	if userIDVal == nil || roleVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found",
		})
	}
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	var companyFilter *string
	if !utils.IsSuperAdminLike(roleName) {
		// User reguler hanya melihat statistik perusahaan mereka
		if companyIDVal != nil {
			if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
				companyFilter = companyIDPtr
			} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
				companyFilter = &companyIDStr
			}
		}
	}
	// Superadmin/administrator melihat semua (companyFilter = nil)

	stats, total, err := h.docUseCase.GetDocumentSummary(companyFilter)
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

	// Check access berdasarkan company_id folder
	companyIDVal := c.Locals("companyID")
	var userCompanyID *string
	if companyIDVal != nil {
		if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
			userCompanyID = &companyIDStr
		}
	}

	if !utils.IsSuperAdminLike(roleName) {
		// Non-superadmin harus akses folder melalui company_id
		if doc.FolderID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke dokumen ini",
			})
		}
		folder, err := h.docUseCase.GetFolderByID(*doc.FolderID)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Folder dokumen tidak ditemukan",
			})
		}
		if userCompanyID == nil || folder.CompanyID == nil || *userCompanyID != *folder.CompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke dokumen ini",
			})
		}
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
// @Param        director_id formData string  false  "ID direktur/individu yang terkait (opsional)"
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

	// Get director_id early untuk validasi file (jika dokumen untuk individu)
	directorIDStr := c.FormValue("director_id")
	var directorIDPtr *string
	if directorIDStr != "" {
		directorIDPtr = &directorIDStr
	}

	// Validasi tipe file HANYA untuk dokumen individu (jika director_id ada)
	if directorIDPtr != nil {
		allowedExts := []string{".docx", ".xlsx", ".xls", ".pptx", ".ppt", ".pdf", ".jpg", ".jpeg", ".png"}
		allowedMimeTypes := []string{
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",       // .xlsx
			"application/vnd.ms-excel", // .xls
			"application/vnd.openxmlformats-officedocument.presentationml.presentation", // .pptx
			"application/vnd.ms-powerpoint",                                             // .ppt
			"application/pdf",                                                           // .pdf
			"image/jpeg",                                                                // .jpg, .jpeg
			"image/png",                                                                 // .png
		}

		fileName := strings.ToLower(fileHeader.Filename)
		ext := ""
		if idx := strings.LastIndex(fileName, "."); idx >= 0 {
			ext = fileName[idx:]
		}

		mimeType := fileHeader.Header.Get("Content-Type")

		// Validasi extension
		isValidExt := false
		for _, allowedExt := range allowedExts {
			if ext == allowedExt {
				isValidExt = true
				break
			}
		}

		// Validasi MIME type (lebih reliable)
		isValidMimeType := false
		for _, allowedMime := range allowedMimeTypes {
			if mimeType == allowedMime {
				isValidMimeType = true
				break
			}
		}

		// Perlu validasi extension ATAU MIME type (untuk kompatibilitas)
		if !isValidExt && !isValidMimeType {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
				Error:   "invalid_file_format",
				Message: "Format file tidak diizinkan. Hanya DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF, dan gambar (JPG/JPEG/PNG) yang diperbolehkan",
			})
		}
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

	// Jika folder ditentukan, pastikan akses folder untuk non-superadmin
	roleVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	roleName := strings.ToLower(fmt.Sprintf("%v", roleVal))

	var userCompanyID *string
	if companyIDVal != nil {
		if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
			userCompanyID = companyIDPtr
		} else if companyIDStr, ok := companyIDVal.(string); ok && companyIDStr != "" {
			userCompanyID = &companyIDStr
		}
	}

	if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
		folder, err := h.docUseCase.GetFolderByID(*folderPtr)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		// Check jika folder milik company user
		if userCompanyID == nil || folder.CompanyID == nil || *userCompanyID != *folder.CompanyID {
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
		DirectorID:  directorIDPtr,
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
// @Param        payload    body      object  false  "JSON payload untuk update metadata only"  example({"folder_id": "folder-id", "director_id": "director-id", "title": "Updated Title", "status": "active", "metadata": {}})
// @Param        file       formData  file    false  "File baru (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        folder_id   formData  string  false  "ID folder (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        director_id formData  string  false  "ID direktur/individu yang terkait (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        title       formData  string  false  "Judul dokumen (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        status      formData  string  false  "Status dokumen (opsional, hanya jika Content-Type: multipart/form-data)"
// @Param        metadata    formData  string  false  "Metadata dalam format JSON string (opsional, hanya jika Content-Type: multipart/form-data)"
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
	// Check access berdasarkan company_id folder
	companyIDValForUpdate := c.Locals("companyID")
	var userCompanyIDForUpdate *string
	if companyIDValForUpdate != nil {
		if companyIDPtr, ok := companyIDValForUpdate.(*string); ok && companyIDPtr != nil {
			userCompanyIDForUpdate = companyIDPtr
		} else if companyIDStr, ok := companyIDValForUpdate.(string); ok && companyIDStr != "" {
			userCompanyIDForUpdate = &companyIDStr
		}
	}

	if !utils.IsSuperAdminLike(roleName) {
		// Non-superadmin harus akses folder melalui company_id
		if existingDoc.FolderID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses mengubah dokumen ini",
			})
		}
		folder, err := h.docUseCase.GetFolderByID(*existingDoc.FolderID)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Folder dokumen tidak ditemukan",
			})
		}
		if userCompanyIDForUpdate == nil || folder.CompanyID == nil || *userCompanyIDForUpdate != *folder.CompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses mengubah dokumen ini",
			})
		}
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
		var directorIDPtr *string
		if v := c.FormValue("director_id"); v != "" {
			directorIDPtr = &v
		}
		var userCompanyIDForMultipart *string
		if companyIDValForUpdate != nil {
			if companyIDPtr, ok := companyIDValForUpdate.(*string); ok && companyIDPtr != nil {
				userCompanyIDForMultipart = companyIDPtr
			} else if companyIDStr, ok := companyIDValForUpdate.(string); ok && companyIDStr != "" {
				userCompanyIDForMultipart = &companyIDStr
			}
		}

		if folderPtr != nil && !utils.IsSuperAdminLike(roleName) {
			folder, err := h.docUseCase.GetFolderByID(*folderPtr)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
					Error:   "not_found",
					Message: "Folder tidak ditemukan",
				})
			}
			// Check jika folder milik company user
			if userCompanyIDForMultipart == nil || folder.CompanyID == nil || *userCompanyIDForMultipart != *folder.CompanyID {
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
			DirectorID:      directorIDPtr,
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
		FolderID   *string                `json:"folder_id"`
		DirectorID *string                `json:"director_id"`
		Title      *string                `json:"title"`
		Status     *string                `json:"status"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Payload tidak valid",
		})
	}

	if payload.FolderID != nil && !utils.IsSuperAdminLike(roleName) {
		// Get userCompanyID untuk JSON payload
		var userCompanyIDForJSON *string
		if companyIDValForUpdate != nil {
			if companyIDPtr, ok := companyIDValForUpdate.(*string); ok && companyIDPtr != nil {
				userCompanyIDForJSON = companyIDPtr
			} else if companyIDStr, ok := companyIDValForUpdate.(string); ok && companyIDStr != "" {
				userCompanyIDForJSON = &companyIDStr
			}
		}

		folder, err := h.docUseCase.GetFolderByID(*payload.FolderID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
				Error:   "not_found",
				Message: "Folder tidak ditemukan",
			})
		}
		// Check jika folder milik company user
		if userCompanyIDForJSON == nil || folder.CompanyID == nil || *userCompanyIDForJSON != *folder.CompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses ke folder ini",
			})
		}
	}

	doc, err := h.docUseCase.UpdateDocument(id, usecase.UpdateDocumentInput{
		FolderID:   payload.FolderID,
		DirectorID: payload.DirectorID,
		Title:      payload.Title,
		Status:     payload.Status,
		Metadata:   payload.Metadata,
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

	// Get document first
	existingDoc, err := h.docUseCase.GetDocumentByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "Document tidak ditemukan",
		})
	}

	// Superadmin/administrator dapat menghapus semua dokumen tanpa check access
	isSuperAdmin := utils.IsSuperAdminLike(roleName)
	var folder *domain.DocumentFolderModel // Variable untuk menyimpan folder jika diperlukan untuk logging

	if !isSuperAdmin {
		// Non-superadmin harus check access berdasarkan company_id folder
		companyIDValForDelete := c.Locals("companyID")
		var userCompanyIDForDelete *string
		if companyIDValForDelete != nil {
			if companyIDPtr, ok := companyIDValForDelete.(*string); ok && companyIDPtr != nil {
				userCompanyIDForDelete = companyIDPtr
			} else if companyIDStr, ok := companyIDValForDelete.(string); ok && companyIDStr != "" {
				userCompanyIDForDelete = &companyIDStr
			}
		}

		// Non-superadmin harus akses folder melalui company_id
		if existingDoc.FolderID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses menghapus dokumen ini",
			})
		}
		var err error
		folder, err = h.docUseCase.GetFolderByID(*existingDoc.FolderID)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Folder dokumen tidak ditemukan",
			})
		}
		if userCompanyIDForDelete == nil || folder.CompanyID == nil || *userCompanyIDForDelete != *folder.CompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Anda tidak memiliki akses menghapus dokumen ini",
			})
		}
	} else if existingDoc.FolderID != nil {
		// Untuk superadmin, ambil folder hanya untuk logging (jika ada)
		folder, _ = h.docUseCase.GetFolderByID(*existingDoc.FolderID)
	}

	// Delete document (superadmin dapat menghapus semua dokumen)
	if err := h.docUseCase.DeleteDocument(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}
	username, _ := c.Locals("username").(string)

	// Prepare log details dengan informasi lengkap
	logDetails := map[string]interface{}{
		"operation":     "delete_document",
		"file_name":     existingDoc.FileName,
		"document_name": existingDoc.Name,
	}

	// Tambahkan informasi folder jika ada
	if existingDoc.FolderID != nil {
		logDetails["folder_id"] = *existingDoc.FolderID
		if folder != nil && folder.CompanyID != nil {
			logDetails["company_id"] = *folder.CompanyID
		}
	}

	// Tambahkan informasi director jika ini adalah dokumen attachment
	if existingDoc.DirectorID != nil {
		logDetails["director_id"] = *existingDoc.DirectorID
		logDetails["document_type"] = "director_attachment"
	}

	audit.LogAction(userIDStr, username, audit.ActionDeleteDoc, audit.ResourceDocument, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, logDetails)
	return c.JSON(fiber.Map{
		"message": "Document deleted",
	})
}
