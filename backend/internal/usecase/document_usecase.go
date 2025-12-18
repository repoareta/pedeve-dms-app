package usecase

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/storage"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DocumentUseCase interface {
	ListFolders(companyID *string) ([]domain.DocumentFolderModel, error)
	CreateFolder(name string, companyID *string, parentID *string, createdBy string) (*domain.DocumentFolderModel, error)
	GetFolderByID(id string) (*domain.DocumentFolderModel, error)
	UpdateFolderName(id string, name string, requesterCompanyID *string, roleName string) (*domain.DocumentFolderModel, error)
	DeleteFolder(id string, requesterCompanyID *string, roleName string) error
	ListDocuments(folderID *string) ([]domain.DocumentModel, error)
	ListDocumentsPaginated(params ListDocumentsParams) ([]domain.DocumentModel, int64, error)
	GetDocumentSummary(companyID *string) ([]domain.DocumentFolderStat, int64, error)
	GetDocumentByID(id string) (*domain.DocumentModel, error)
	UploadDocument(input UploadDocumentInput) (*domain.DocumentModel, error)
	UpdateDocument(id string, input UpdateDocumentInput) (*domain.DocumentModel, error)
	DeleteDocument(id string) error
}

type UploadDocumentInput struct {
	FolderID    *string
	DirectorID  *string // ID direktur/individu yang terkait dengan dokumen (opsional)
	Title       string
	FileName    string
	ContentType string
	Data        []byte
	Size        int64
	Status      string
	UploaderID  string
	Metadata    map[string]interface{}
}

type UpdateDocumentInput struct {
	FolderID        *string
	DirectorID      *string // ID direktur/individu yang terkait dengan dokumen (opsional)
	Title           *string
	Status          *string
	Metadata        map[string]interface{}
	FileName        *string
	FileContentType *string
	FileData        []byte
	FileSize        *int64
}

type ListDocumentsParams struct {
	FolderID   *string
	CompanyID  *string // Filter berdasarkan company_id dari folder
	DirectorID *string // Filter berdasarkan director_id (dokumen individu)
	Search     string
	SortBy     string
	SortDir    string
	Page       int
	PageSize   int
	OwnerID    *string // Uploader ID (optional filter)
	TypeFilter string
}

type documentUseCase struct {
	docRepo     repository.DocumentRepository
	companyRepo repository.CompanyRepository
}

func NewDocumentUseCase() DocumentUseCase {
	return &documentUseCase{
		docRepo:     repository.NewDocumentRepository(),
		companyRepo: repository.NewCompanyRepository(),
	}
}

func NewDocumentUseCaseWithRepo(repo repository.DocumentRepository) DocumentUseCase {
	return &documentUseCase{
		docRepo:     repo,
		companyRepo: repository.NewCompanyRepository(), // Use default for backward compatibility
	}
}

// NewDocumentUseCaseWithDB creates a new document use case with injected DB (for testing)
func NewDocumentUseCaseWithDB(db *gorm.DB) DocumentUseCase {
	return &documentUseCase{
		docRepo:     repository.NewDocumentRepositoryWithDB(db),
		companyRepo: repository.NewCompanyRepositoryWithDB(db),
	}
}

func (uc *documentUseCase) ListFolders(companyID *string) ([]domain.DocumentFolderModel, error) {
	return uc.docRepo.ListFolders(companyID)
}

func (uc *documentUseCase) CreateFolder(name string, companyID *string, parentID *string, createdBy string) (*domain.DocumentFolderModel, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name required")
	}
	if createdBy == "" {
		return nil, fmt.Errorf("creator required")
	}
	folder := &domain.DocumentFolderModel{
		ID:        uuid.GenerateUUID(),
		Name:      name,
		CompanyID: companyID,
		ParentID:  parentID,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.docRepo.CreateFolder(folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (uc *documentUseCase) GetFolderByID(id string) (*domain.DocumentFolderModel, error) {
	return uc.docRepo.GetFolderByID(id)
}

func (uc *documentUseCase) UpdateFolderName(id, name string, requesterCompanyID *string, roleName string) (*domain.DocumentFolderModel, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name required")
	}

	folder, err := uc.docRepo.GetFolderByID(id)
	if err != nil {
		return nil, err
	}

	// Only company owner or superadmin/administrator can rename
	roleLower := strings.ToLower(roleName)
	isSuperAdmin := roleLower == "superadmin" || roleLower == "administrator"

	if !isSuperAdmin {
		// Check if requester's company matches folder's company
		if requesterCompanyID == nil || folder.CompanyID == nil || *requesterCompanyID != *folder.CompanyID {
			return nil, fmt.Errorf("forbidden")
		}
	}

	if err := uc.docRepo.UpdateFolderName(id, name); err != nil {
		return nil, err
	}
	folder.Name = name
	return folder, nil
}

// deleteFolderRecursive menghapus folder beserta semua child folders dan dokumen di dalamnya secara rekursif
func (uc *documentUseCase) deleteFolderRecursive(folderID string) error {
	// Get child folders first
	childFolders, err := uc.docRepo.GetChildFolders(folderID)
	if err != nil {
		return fmt.Errorf("failed to get child folders: %w", err)
	}

	// Recursively delete all child folders first
	for _, child := range childFolders {
		if err := uc.deleteFolderRecursive(child.ID); err != nil {
			return fmt.Errorf("failed to delete child folder %s: %w", child.ID, err)
		}
	}

	// Delete all documents inside this folder
	if err := uc.docRepo.DeleteDocumentsByFolder(folderID); err != nil {
		return fmt.Errorf("failed to delete documents in folder: %w", err)
	}

	// Finally, delete the folder itself
	if err := uc.docRepo.DeleteFolder(folderID); err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	return nil
}

func (uc *documentUseCase) DeleteFolder(id string, requesterCompanyID *string, roleName string) error {
	folder, err := uc.docRepo.GetFolderByID(id)
	if err != nil {
		return err
	}

	// HANYA superadmin/administrator yang boleh menghapus folder
	roleLower := strings.ToLower(roleName)
	isSuperAdmin := roleLower == "superadmin" || roleLower == "administrator"

	if !isSuperAdmin {
		return fmt.Errorf("forbidden: hanya superadmin dan administrator yang dapat menghapus folder")
	}

	// Simpan companyID dan nama folder sebelum dihapus (untuk auto-generate folder baru)
	var companyIDToRegenerate *string
	var companyName string
	if folder.CompanyID != nil {
		companyIDToRegenerate = folder.CompanyID
		// Ambil nama perusahaan dari company repository
		company, err := uc.companyRepo.GetByID(*folder.CompanyID)
		if err == nil && company != nil {
			companyName = company.Name
		} else {
			// Fallback ke nama folder jika tidak bisa ambil company
			companyName = folder.Name
		}
	}

	// Delete folder recursively (including all child folders and documents)
	if err := uc.deleteFolderRecursive(id); err != nil {
		return err
	}

	// Setelah folder dihapus, auto-generate folder baru dengan nama perusahaan yang sama (kosong)
	if companyIDToRegenerate != nil && companyName != "" {
		newFolder := &domain.DocumentFolderModel{
			ID:        uuid.GenerateUUID(),
			Name:      companyName,
			CompanyID: companyIDToRegenerate,
			ParentID:  nil, // Folder root untuk perusahaan
			CreatedBy: "",  // System-created, tidak ada user creator
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := uc.docRepo.CreateFolder(newFolder); err != nil {
			// Log error tapi jangan return error karena folder sudah terhapus
			// Folder baru bisa dibuat manual nanti jika perlu
			fmt.Printf("Warning: failed to auto-generate folder for company %s: %v\n", *companyIDToRegenerate, err)
		}
	}

	return nil
}

func (uc *documentUseCase) ListDocuments(folderID *string) ([]domain.DocumentModel, error) {
	return uc.docRepo.ListDocuments(folderID)
}

func (uc *documentUseCase) ListDocumentsPaginated(params ListDocumentsParams) ([]domain.DocumentModel, int64, error) {
	q := repository.ListDocumentsQuery{
		FolderID:   params.FolderID,
		CompanyID:  params.CompanyID,
		DirectorID: params.DirectorID,
		Search:     params.Search,
		SortBy:     params.SortBy,
		SortDir:    params.SortDir,
		Page:       params.Page,
		PageSize:   params.PageSize,
		UploaderID: params.OwnerID,
		Type:       params.TypeFilter,
	}
	return uc.docRepo.ListDocumentsPaginated(q)
}

func (uc *documentUseCase) GetDocumentSummary(companyID *string) ([]domain.DocumentFolderStat, int64, error) {
	stats, err := uc.docRepo.GetFolderStats(companyID)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.docRepo.GetTotalSize(companyID)
	if err != nil {
		return nil, 0, err
	}
	return stats, total, nil
}

func (uc *documentUseCase) GetDocumentByID(id string) (*domain.DocumentModel, error) {
	return uc.docRepo.GetDocumentByID(id)
}

// normalizeReferenceForComparison normalizes reference for uniqueness comparison
// Removes spaces, converts to uppercase for case-insensitive comparison
func normalizeReferenceForComparison(ref string) string {
	normalized := strings.TrimSpace(ref)
	normalized = strings.ToUpper(normalized)
	normalized = strings.ReplaceAll(normalized, " ", "")
	return normalized
}

// checkReferenceExists checks if a reference number already exists in other documents
func (uc *documentUseCase) checkReferenceExists(reference string, excludeDocumentID string) (bool, error) {
	if reference == "" {
		return false, nil // Empty reference is considered not existing
	}

	// Get all documents
	allDocs, err := uc.docRepo.ListDocuments(nil)
	if err != nil {
		return false, fmt.Errorf("failed to list documents: %w", err)
	}

	// Normalize input reference for comparison
	normalizedInput := normalizeReferenceForComparison(reference)

	// Check if any document has the same reference (excluding current document if specified)
	for _, doc := range allDocs {
		// Skip current document if in edit mode
		if excludeDocumentID != "" && doc.ID == excludeDocumentID {
			continue
		}

		// Parse metadata
		if doc.Metadata == nil {
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(doc.Metadata, &metadata); err != nil {
			continue
		}

		// Get reference from metadata
		existingRef, ok := metadata["reference"].(string)
		if !ok || existingRef == "" {
			continue
		}

		// Normalize existing reference for comparison
		normalizedExisting := normalizeReferenceForComparison(existingRef)

		// Check if they match (case-insensitive, space-insensitive)
		if normalizedInput == normalizedExisting {
			return true, nil // Reference already exists
		}
	}

	return false, nil // Reference is unique
}

func (uc *documentUseCase) UploadDocument(input UploadDocumentInput) (*domain.DocumentModel, error) {
	if input.FileName == "" || len(input.Data) == 0 {
		return nil, fmt.Errorf("file invalid")
	}

	// Validate reference uniqueness if provided
	if input.Metadata != nil {
		if reference, ok := input.Metadata["reference"].(string); ok && reference != "" {
			exists, err := uc.checkReferenceExists(reference, "")
			if err != nil {
				return nil, fmt.Errorf("failed to validate reference: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("nomor referensi \"%s\" sudah digunakan di dokumen lain", reference)
			}
		}
	}

	storageManager, err := storage.GetStorageManager()
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	ext := filepath.Ext(input.FileName)
	newFileName := fmt.Sprintf("%s%s", uuid.GenerateUUID(), ext)
	fileURL, err := storageManager.UploadFile("documents", newFileName, input.Data, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Normalize file URL to use backend proxy endpoint
	// This ensures consistent URL format for both GCP Storage and Local Storage
	// Format: /api/v1/files/documents/filename.pdf
	if strings.HasPrefix(fileURL, "https://storage.googleapis.com/") {
		// Extract path from GCP URL: https://storage.googleapis.com/bucket/documents/filename.pdf
		// Convert to: /api/v1/files/documents/filename.pdf
		parts := strings.SplitN(fileURL, "/", 5) // Split after bucket name
		if len(parts) >= 5 {
			// parts[4] contains "documents/filename.pdf"
			fileURL = fmt.Sprintf("/api/v1/files/%s", parts[4])
		}
	} else if strings.HasPrefix(fileURL, "/") && !strings.HasPrefix(fileURL, "/api/v1/files/") {
		// If using Local Storage, URL format is: /documents/filename.pdf
		// Convert to: /api/v1/files/documents/filename.pdf
		// Remove leading slash if present
		pathWithoutSlash := strings.TrimPrefix(fileURL, "/")
		fileURL = fmt.Sprintf("/api/v1/files/%s", pathWithoutSlash)
	}

	var metadataJSON []byte
	if input.Metadata != nil {
		metadataJSON, _ = json.Marshal(input.Metadata)
	}

	title := input.Title
	if title == "" {
		title = input.FileName
	}

	doc := &domain.DocumentModel{
		ID:         uuid.GenerateUUID(),
		FolderID:   input.FolderID,
		DirectorID: input.DirectorID, // Relasi dengan DirectorModel
		Name:       title,
		FileName:   input.FileName,
		FilePath:   fileURL,
		MimeType:   input.ContentType,
		Size:       input.Size,
		Status:     input.Status,
		Metadata:   datatypes.JSON(metadataJSON),
		UploaderID: input.UploaderID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := uc.docRepo.CreateDocument(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (uc *documentUseCase) UpdateDocument(id string, input UpdateDocumentInput) (*domain.DocumentModel, error) {
	// Get existing document
	doc, err := uc.docRepo.GetDocumentByID(id)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Validate reference uniqueness if provided in metadata update
	if input.Metadata != nil {
		if reference, ok := input.Metadata["reference"].(string); ok && reference != "" {
			exists, err := uc.checkReferenceExists(reference, id)
			if err != nil {
				return nil, fmt.Errorf("failed to validate reference: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("nomor referensi \"%s\" sudah digunakan di dokumen lain", reference)
			}
		}
	}

	// Update fields if provided
	if input.Title != nil {
		doc.Name = *input.Title
	}
	if input.FolderID != nil {
		doc.FolderID = input.FolderID
	}
	if input.DirectorID != nil {
		doc.DirectorID = input.DirectorID
	}
	if input.Status != nil {
		doc.Status = *input.Status
	}
	if input.Metadata != nil {
		metadataJSON, _ := json.Marshal(input.Metadata)
		doc.Metadata = datatypes.JSON(metadataJSON)
	}

	// Optional: update file
	if len(input.FileData) > 0 && input.FileName != nil {
		storageManager, err := storage.GetStorageManager()
		if err != nil {
			return nil, fmt.Errorf("failed to init storage: %w", err)
		}

		ext := filepath.Ext(*input.FileName)
		newFileName := fmt.Sprintf("%s%s", uuid.GenerateUUID(), ext)
		contentType := ""
		if input.FileContentType != nil {
			contentType = *input.FileContentType
		}

		fileURL, err := storageManager.UploadFile("documents", newFileName, input.FileData, contentType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file: %w", err)
		}

		if strings.HasPrefix(fileURL, "https://storage.googleapis.com/") {
			parts := strings.SplitN(fileURL, "/", 5)
			if len(parts) >= 5 {
				fileURL = fmt.Sprintf("/api/v1/files/%s", parts[4])
			}
		} else if strings.HasPrefix(fileURL, "/") && !strings.HasPrefix(fileURL, "/api/v1/files/") {
			pathWithoutSlash := strings.TrimPrefix(fileURL, "/")
			fileURL = fmt.Sprintf("/api/v1/files/%s", pathWithoutSlash)
		}

		doc.FileName = *input.FileName
		doc.FilePath = fileURL
		doc.MimeType = contentType
		if input.FileSize != nil {
			doc.Size = *input.FileSize
		}
	}

	doc.UpdatedAt = time.Now()

	if err := uc.docRepo.UpdateDocument(doc); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return doc, nil
}

func (uc *documentUseCase) DeleteDocument(id string) error {
	return uc.docRepo.DeleteDocument(id)
}
