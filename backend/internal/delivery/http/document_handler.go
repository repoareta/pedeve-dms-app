package http

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// getDocumentsHandler returns list of documents (untuk Fiber)
// @Summary      Get all documents
// @Description  Returns a list of all documents (requires authentication)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.Document
// @Failure      401  {object}  domain.ErrorResponse
// @Router       /api/v1/documents [get]
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
// @Summary      Get document by ID
// @Description  Returns a single document by ID (requires authentication)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  domain.Document
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/documents/{id} [get]
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
// @Summary      Create document
// @Description  Creates a new document (requires authentication)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        document  body      domain.Document  true  "Document object"
// @Success      201       {object}  domain.Document
// @Failure      400       {object}  domain.ErrorResponse
// @Failure      401       {object}  domain.ErrorResponse
// @Router       /api/v1/documents [post]
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
// @Summary      Update document
// @Description  Updates an existing document (requires authentication)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      string    true  "Document ID"
// @Param        document  body      domain.Document  true  "Document object"
// @Success      200       {object}  domain.Document
// @Failure      401       {object}  domain.ErrorResponse
// @Failure      404       {object}  domain.ErrorResponse
// @Router       /api/v1/documents/{id} [put]
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
// @Summary      Delete document
// @Description  Deletes a document by ID (requires authentication)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/documents/{id} [delete]
func DeleteDocumentHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"message": "Document " + id + " deleted successfully",
	})
}

