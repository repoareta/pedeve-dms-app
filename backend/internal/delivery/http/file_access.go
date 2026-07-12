package http

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
)

var allowedFileBuckets = map[string]bool{
	"logos":     true,
	"documents": true,
}

// authorizeFileAccess memastikan user hanya bisa mengakses file yang diizinkan berdasarkan RBAC.
func authorizeFileAccess(c *fiber.Ctx, bucketPath, filename string) error {
	zapLog := logger.GetLogger()

	if !allowedFileBuckets[bucketPath] {
		zapLog.Warn("Unauthorized file bucket access attempt",
			zap.String("bucket", bucketPath),
			zap.String("filename", filename),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})
	}

	roleNameVal := c.Locals("roleName")
	if roleNameVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	roleName := strings.ToLower(fmt.Sprintf("%v", roleNameVal))
	if utils.IsSuperAdminLike(roleName) || roleName == "admin" {
		return nil
	}

	restricted, allowedCompanyIDs := buildDocumentsCompanyScope(c, roleName)
	if !restricted {
		return nil
	}

	allowedSet := buildCompanyIDSet(allowedCompanyIDs)

	switch bucketPath {
	case "logos":
		return authorizeLogoAccess(c, filename, allowedSet)
	case "documents":
		return authorizeDocumentAccess(c, filename, allowedSet)
	default:
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})
	}
}

func authorizeLogoAccess(c *fiber.Ctx, filename string, allowedSet map[string]bool) error {
	zapLog := logger.GetLogger()
	db := database.GetDB()

	var company domain.CompanyModel
	err := db.Where("logo LIKE ?", "%"+filename+"%").First(&company).Error
	if err != nil {
		zapLog.Warn("Logo file not associated with any company",
			zap.String("filename", filename),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "File not found",
		})
	}

	if !allowedSet[company.ID] {
		zapLog.Warn("Unauthorized logo access attempt",
			zap.String("filename", filename),
			zap.String("company_id", company.ID),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})
	}

	return nil
}

func authorizeDocumentAccess(c *fiber.Ctx, filename string, allowedSet map[string]bool) error {
	zapLog := logger.GetLogger()

	storagePath := fmt.Sprintf("documents/%s", filename)
	apiPath := fmt.Sprintf("/api/v1/files/%s", storagePath)

	var doc domain.DocumentModel
	db := database.GetDB()
	err := db.Where(
		"file_path = ? OR file_path = ? OR file_path LIKE ? OR file_name = ?",
		storagePath, apiPath, "%"+filename, filename,
	).First(&doc).Error
	if err != nil {
		zapLog.Warn("Document file not found in database",
			zap.String("filename", filename),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "File not found",
		})
	}

	companyID, err := resolveDocumentCompanyID(&doc)
	if err != nil || companyID == "" {
		zapLog.Warn("Document has no company association",
			zap.String("document_id", doc.ID),
			zap.String("filename", filename),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})
	}

	if !allowedSet[companyID] {
		zapLog.Warn("Unauthorized document file access attempt",
			zap.String("filename", filename),
			zap.String("document_id", doc.ID),
			zap.String("company_id", companyID),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})
	}

	return nil
}

func resolveDocumentCompanyID(doc *domain.DocumentModel) (string, error) {
	docRepo := repository.NewDocumentRepository()
	if doc.FolderID != nil && *doc.FolderID != "" {
		folder, err := docRepo.GetFolderByID(*doc.FolderID)
		if err != nil {
			return "", err
		}
		if folder.CompanyID != nil && *folder.CompanyID != "" {
			return *folder.CompanyID, nil
		}
	}

	if doc.DirectorID != nil && *doc.DirectorID != "" {
		var director domain.DirectorModel
		if err := database.GetDB().Where("id = ?", *doc.DirectorID).First(&director).Error; err != nil {
			return "", err
		}
		if director.CompanyID != "" {
			return director.CompanyID, nil
		}
	}

	return "", fmt.Errorf("document has no company association")
}
