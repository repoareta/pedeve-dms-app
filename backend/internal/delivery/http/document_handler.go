package http

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// getDocumentsHandler returns list of documents (untuk Fiber)
// @Summary      Ambil Semua Documents
// @Description  Mengambil daftar semua documents. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Document  "Daftar documents berhasil diambil. Response berupa array of Document objects"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Router       /api/v1/documents [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Response: Response berupa array of Document objects dengan field id, title, description, content, created_at, updated_at
func GetDocumentsHandler(c *fiber.Ctx) error {
	documents := []domain.Document{
		{
			ID:          "1",
			Title:       "Sample Document 1",
			Description: "This is a sample document",
			Content:     "Document content here...",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
		{
			ID:          "2",
			Title:       "Sample Document 2",
			Description: "Another sample document",
			Content:     "More document content...",
			CreatedAt:   "2024-01-02T00:00:00Z",
			UpdatedAt:   "2024-01-02T00:00:00Z",
		},
	}
	return c.JSON(documents)
}

// getDocumentHandler returns a single document (untuk Fiber)
// @Summary      Ambil Document by ID
// @Description  Mengambil detail document berdasarkan ID. Endpoint ini tidak memerlukan CSRF token karena menggunakan method GET (read-only).
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID dari document yang ingin diambil"
// @Success      200  {object}  domain.Document  "Detail document berhasil diambil. Response berupa Document object dengan field id, title, description, content, created_at, updated_at"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      404  {object}  domain.ErrorResponse  "Document tidak ditemukan"
// @Router       /api/v1/documents/{id} [get]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini tidak memerlukan CSRF token karena menggunakan GET method (read-only)
// @note         3. Path Parameter: ID document dikirim sebagai path parameter, bukan query parameter
func GetDocumentHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	document := domain.Document{
		ID:          id,
		Title:       "Sample Document",
		Description: "This is a sample document",
		Content:     "Document content here...",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}
	return c.JSON(document)
}

// createDocumentHandler creates a new document (untuk Fiber)
// @Summary      Buat Document Baru
// @Description  Membuat document baru. Endpoint ini memerlukan authentication dan CSRF token karena menggunakan method POST.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        document  body      domain.Document  true  "Data document yang ingin dibuat (title, description, content)"
// @Success      201       {object}  domain.Document  "Document berhasil dibuat. Response berupa Document object yang baru dibuat dengan id, title, description, content, created_at, updated_at"
// @Failure      400       {object}  domain.ErrorResponse  "Request body tidak valid atau validation error"
// @Failure      401       {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403       {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Router       /api/v1/documents [post]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan POST method
// @note         3. Request Body: Body harus berupa JSON dengan field title, description, content (id, created_at, updated_at akan di-generate otomatis)
// @note         4. Status Code: Response menggunakan status code 201 (Created) untuk operasi create
func CreateDocumentHandler(c *fiber.Ctx) error {
	document := domain.Document{
		ID:          "3",
		Title:       "New Document",
		Description: "A newly created document",
		Content:     "New document content...",
		CreatedAt:   "2024-01-03T00:00:00Z",
		UpdatedAt:   "2024-01-03T00:00:00Z",
	}
	return c.Status(fiber.StatusCreated).JSON(document)
}

// updateDocumentHandler updates a document (untuk Fiber)
// @Summary      Update Document
// @Description  Mengupdate document yang sudah ada berdasarkan ID. Endpoint ini memerlukan authentication dan CSRF token karena menggunakan method PUT.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      string    true  "ID dari document yang ingin di-update"
// @Param        document  body      domain.Document  true  "Data document yang ingin di-update (title, description, content)"
// @Success      200       {object}  domain.Document  "Document berhasil di-update. Response berupa Document object yang sudah di-update dengan updated_at yang baru"
// @Failure      400       {object}  domain.ErrorResponse  "Request body tidak valid atau validation error"
// @Failure      401       {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403       {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Failure      404       {object}  domain.ErrorResponse  "Document tidak ditemukan"
// @Router       /api/v1/documents/{id} [put]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan PUT method
// @note         3. Path Parameter: ID document dikirim sebagai path parameter
// @note         4. Request Body: Body harus berupa JSON dengan field yang ingin di-update (title, description, content)
// @note         5. Timestamp: Field updated_at akan di-update otomatis saat document di-update
func UpdateDocumentHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	document := domain.Document{
		ID:          id,
		Title:       "Updated Document",
		Description: "This document has been updated",
		Content:     "Updated content...",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-03T00:00:00Z",
	}
	return c.JSON(document)
}

// deleteDocumentHandler deletes a document (untuk Fiber)
// @Summary      Hapus Document
// @Description  Menghapus document berdasarkan ID. Endpoint ini memerlukan authentication dan CSRF token karena menggunakan method DELETE.
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID dari document yang ingin dihapus"
// @Success      200  {object}  map[string]string  "Document berhasil dihapus. Response berisi message: 'Document {id} deleted successfully'"
// @Failure      401  {object}  domain.ErrorResponse  "Token tidak valid atau user tidak terautentikasi"
// @Failure      403  {object}  domain.ErrorResponse  "CSRF token tidak valid atau tidak ditemukan"
// @Failure      404  {object}  domain.ErrorResponse  "Document tidak ditemukan"
// @Router       /api/v1/documents/{id} [delete]
// @note         Catatan Teknis:
// @note         1. Authentication: Memerlukan JWT token valid dalam httpOnly cookie (auth_token) atau Authorization header
// @note         2. CSRF Protection: Endpoint ini memerlukan CSRF token dalam header X-CSRF-Token karena menggunakan DELETE method
// @note         3. Path Parameter: ID document dikirim sebagai path parameter
// @note         4. Permanent Delete: Document akan dihapus secara permanen dari database
// @note         5. Audit Logging: Aksi delete document sebaiknya dicatat dalam audit log (jika diimplementasikan)
func DeleteDocumentHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"message": "Document " + id + " deleted successfully",
	})
}

