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
)

type DocumentUseCase interface {
	ListFolders(ownerID *string) ([]domain.DocumentFolderModel, error)
	CreateFolder(name string, parentID *string, createdBy string) (*domain.DocumentFolderModel, error)
	GetFolderByID(id string) (*domain.DocumentFolderModel, error)
	UpdateFolderName(id, name, requesterID, roleName string) (*domain.DocumentFolderModel, error)
	DeleteFolder(id, requesterID, roleName string) error
	ListDocuments(folderID *string) ([]domain.DocumentModel, error)
	ListDocumentsPaginated(params ListDocumentsParams) ([]domain.DocumentModel, int64, error)
	GetDocumentSummary(ownerID *string) ([]domain.DocumentFolderStat, int64, error)
	GetDocumentByID(id string) (*domain.DocumentModel, error)
	UploadDocument(input UploadDocumentInput) (*domain.DocumentModel, error)
	UpdateDocument(id string, input UpdateDocumentInput) (*domain.DocumentModel, error)
	DeleteDocument(id string) error
}

type UploadDocumentInput struct {
	FolderID    *string
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
	Search     string
	SortBy     string
	SortDir    string
	Page       int
	PageSize   int
	OwnerID    *string
	TypeFilter string
}

type documentUseCase struct {
	docRepo repository.DocumentRepository
}

func NewDocumentUseCase() DocumentUseCase {
	return &documentUseCase{
		docRepo: repository.NewDocumentRepository(),
	}
}

func NewDocumentUseCaseWithRepo(repo repository.DocumentRepository) DocumentUseCase {
	return &documentUseCase{
		docRepo: repo,
	}
}

func (uc *documentUseCase) ListFolders(ownerID *string) ([]domain.DocumentFolderModel, error) {
	return uc.docRepo.ListFolders(ownerID)
}

func (uc *documentUseCase) CreateFolder(name string, parentID *string, createdBy string) (*domain.DocumentFolderModel, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name required")
	}
	if createdBy == "" {
		return nil, fmt.Errorf("creator required")
	}
	folder := &domain.DocumentFolderModel{
		ID:        uuid.GenerateUUID(),
		Name:      name,
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

func (uc *documentUseCase) UpdateFolderName(id, name, requesterID, roleName string) (*domain.DocumentFolderModel, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name required")
	}

	folder, err := uc.docRepo.GetFolderByID(id)
	if err != nil {
		return nil, err
	}

	// Only owner or superadmin/administrator can rename
	roleLower := strings.ToLower(roleName)
	if roleLower != "superadmin" && roleLower != "administrator" && folder.CreatedBy != requesterID {
		return nil, fmt.Errorf("forbidden")
	}

	if err := uc.docRepo.UpdateFolderName(id, name); err != nil {
		return nil, err
	}
	folder.Name = name
	return folder, nil
}

func (uc *documentUseCase) DeleteFolder(id, requesterID, roleName string) error {
	folder, err := uc.docRepo.GetFolderByID(id)
	if err != nil {
		return err
	}

	roleLower := strings.ToLower(roleName)
	if roleLower != "superadmin" && roleLower != "administrator" && folder.CreatedBy != requesterID {
		return fmt.Errorf("forbidden")
	}

	// Delete documents inside folder
	if err := uc.docRepo.DeleteDocumentsByFolder(id); err != nil {
		return err
	}

	// Delete folder
	return uc.docRepo.DeleteFolder(id)
}

func (uc *documentUseCase) ListDocuments(folderID *string) ([]domain.DocumentModel, error) {
	return uc.docRepo.ListDocuments(folderID)
}

func (uc *documentUseCase) ListDocumentsPaginated(params ListDocumentsParams) ([]domain.DocumentModel, int64, error) {
	q := repository.ListDocumentsQuery{
		FolderID:   params.FolderID,
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

func (uc *documentUseCase) GetDocumentSummary(ownerID *string) ([]domain.DocumentFolderStat, int64, error) {
	stats, err := uc.docRepo.GetFolderStats(ownerID)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.docRepo.GetTotalSize(ownerID)
	if err != nil {
		return nil, 0, err
	}
	return stats, total, nil
}

func (uc *documentUseCase) GetDocumentByID(id string) (*domain.DocumentModel, error) {
	return uc.docRepo.GetDocumentByID(id)
}

func (uc *documentUseCase) UploadDocument(input UploadDocumentInput) (*domain.DocumentModel, error) {
	if input.FileName == "" || len(input.Data) == 0 {
		return nil, fmt.Errorf("file invalid")
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

	// Update fields if provided
	if input.Title != nil {
		doc.Name = *input.Title
	}
	if input.FolderID != nil {
		doc.FolderID = input.FolderID
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
