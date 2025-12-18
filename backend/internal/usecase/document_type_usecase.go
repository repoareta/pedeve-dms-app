package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"gorm.io/gorm"
)

// DocumentTypeUseCase interface untuk document type operations
type DocumentTypeUseCase interface {
	CreateDocumentType(name, createdBy string) (*domain.DocumentTypeModel, error)
	GetDocumentTypeByID(id string) (*domain.DocumentTypeModel, error)
	GetDocumentTypeByName(name string) (*domain.DocumentTypeModel, error)
	GetAllDocumentTypes(includeInactive bool) ([]domain.DocumentTypeModel, error)
	GetActiveDocumentTypes() ([]domain.DocumentTypeModel, error)
	UpdateDocumentType(id string, name *string, isActive *bool) (*domain.DocumentTypeModel, error)
	DeleteDocumentType(id, requesterRole string) error // Soft delete, only if not in use
}

type documentTypeUseCase struct {
	docTypeRepo repository.DocumentTypeRepository
	docRepo     repository.DocumentRepository
}

// NewDocumentTypeUseCaseWithDB creates a new document type use case with injected DB (for testing)
func NewDocumentTypeUseCaseWithDB(db *gorm.DB) DocumentTypeUseCase {
	return &documentTypeUseCase{
		docTypeRepo: repository.NewDocumentTypeRepositoryWithDB(db),
		docRepo:     repository.NewDocumentRepositoryWithDB(db),
	}
}

// NewDocumentTypeUseCase creates a new document type use case with default DB
func NewDocumentTypeUseCase() DocumentTypeUseCase {
	return NewDocumentTypeUseCaseWithDB(database.GetDB())
}

func (uc *documentTypeUseCase) CreateDocumentType(name, createdBy string) (*domain.DocumentTypeModel, error) {
	// Trim and validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama jenis dokumen tidak boleh kosong")
	}

	// Check if already exists (case-insensitive)
	existing, err := uc.docTypeRepo.GetByName(name)
	if err == nil && existing != nil {
		// If exists but inactive, reactivate it
		if !existing.IsActive {
			existing.IsActive = true
			existing.CreatedBy = createdBy
			if err := uc.docTypeRepo.Update(existing); err != nil {
				return nil, fmt.Errorf("gagal mengaktifkan jenis dokumen: %w", err)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("jenis dokumen '%s' sudah ada", name)
	}

	// Create new document type
	docType := &domain.DocumentTypeModel{
		ID:        uuid.GenerateUUID(),
		Name:      name,
		IsActive:  true,
		UsageCount: 0,
		CreatedBy: createdBy,
	}

	if err := uc.docTypeRepo.Create(docType); err != nil {
		return nil, fmt.Errorf("gagal membuat jenis dokumen: %w", err)
	}

	return docType, nil
}

func (uc *documentTypeUseCase) GetDocumentTypeByID(id string) (*domain.DocumentTypeModel, error) {
	return uc.docTypeRepo.GetByID(id)
}

func (uc *documentTypeUseCase) GetDocumentTypeByName(name string) (*domain.DocumentTypeModel, error) {
	return uc.docTypeRepo.GetByName(name)
}

func (uc *documentTypeUseCase) GetAllDocumentTypes(includeInactive bool) ([]domain.DocumentTypeModel, error) {
	return uc.docTypeRepo.GetAll(includeInactive)
}

func (uc *documentTypeUseCase) GetActiveDocumentTypes() ([]domain.DocumentTypeModel, error) {
	return uc.docTypeRepo.GetActive()
}

func (uc *documentTypeUseCase) UpdateDocumentType(id string, name *string, isActive *bool) (*domain.DocumentTypeModel, error) {
	// Get existing document type
	docType, err := uc.docTypeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("jenis dokumen tidak ditemukan: %w", err)
	}

	// Update name if provided
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, errors.New("nama jenis dokumen tidak boleh kosong")
		}

		// Check if name already exists (case-insensitive, excluding current)
		existing, err := uc.docTypeRepo.GetByName(trimmedName)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("jenis dokumen '%s' sudah ada", trimmedName)
		}

		docType.Name = trimmedName
	}

	// Update is_active if provided
	if isActive != nil {
		docType.IsActive = *isActive
	}

	// Save changes
	if err := uc.docTypeRepo.Update(docType); err != nil {
		return nil, fmt.Errorf("gagal memperbarui jenis dokumen: %w", err)
	}

	return docType, nil
}

func (uc *documentTypeUseCase) DeleteDocumentType(id, requesterRole string) error {
	// Only superadmin and administrator can delete
	roleLower := strings.ToLower(requesterRole)
	if roleLower != "superadmin" && roleLower != "administrator" {
		return errors.New("hanya superadmin dan administrator yang dapat menghapus jenis dokumen")
	}

	// Check if document type exists
	docType, err := uc.docTypeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("jenis dokumen tidak ditemukan: %w", err)
	}

	// Check usage count
	usageCount, err := uc.docTypeRepo.CountUsage(id)
	if err != nil {
		return fmt.Errorf("gagal mengecek penggunaan jenis dokumen: %w", err)
	}

	// If in use, prevent deletion but allow soft delete (set is_active = false)
	// This way, existing documents won't break, but new documents can't use this type
	if usageCount > 0 {
		// Soft delete: set is_active = false
		docType.IsActive = false
		if err := uc.docTypeRepo.Update(docType); err != nil {
			return fmt.Errorf("gagal menonaktifkan jenis dokumen: %w", err)
		}
		return nil // Successfully soft deleted
	}

	// If not in use, we can soft delete (safer than hard delete)
	docType.IsActive = false
	if err := uc.docTypeRepo.Update(docType); err != nil {
		return fmt.Errorf("gagal menghapus jenis dokumen: %w", err)
	}

	return nil
}

