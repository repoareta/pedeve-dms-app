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

// ShareholderTypeUseCase interface untuk shareholder type operations
type ShareholderTypeUseCase interface {
	CreateShareholderType(name, createdBy string) (*domain.ShareholderTypeModel, error)
	GetShareholderTypeByID(id string) (*domain.ShareholderTypeModel, error)
	GetShareholderTypeByName(name string) (*domain.ShareholderTypeModel, error)
	GetAllShareholderTypes(includeInactive bool) ([]domain.ShareholderTypeModel, error)
	GetActiveShareholderTypes() ([]domain.ShareholderTypeModel, error)
	UpdateShareholderType(id string, name *string, isActive *bool) (*domain.ShareholderTypeModel, error)
	DeleteShareholderType(id, requesterRole string) error // Soft delete, only if not in use
}

type shareholderTypeUseCase struct {
	shareholderTypeRepo repository.ShareholderTypeRepository
	shareholderRepo     repository.ShareholderRepository
}

// NewShareholderTypeUseCaseWithDB creates a new shareholder type use case with injected DB (for testing)
func NewShareholderTypeUseCaseWithDB(db *gorm.DB) ShareholderTypeUseCase {
	return &shareholderTypeUseCase{
		shareholderTypeRepo: repository.NewShareholderTypeRepositoryWithDB(db),
		shareholderRepo:     repository.NewShareholderRepositoryWithDB(db),
	}
}

// NewShareholderTypeUseCase creates a new shareholder type use case with default DB
func NewShareholderTypeUseCase() ShareholderTypeUseCase {
	return NewShareholderTypeUseCaseWithDB(database.GetDB())
}

func (uc *shareholderTypeUseCase) CreateShareholderType(name, createdBy string) (*domain.ShareholderTypeModel, error) {
	// Trim and validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama jenis pemegang saham tidak boleh kosong")
	}

	// Check if already exists (case-insensitive)
	existing, _ := uc.shareholderTypeRepo.GetByName(name)
	if existing != nil {
		if !existing.IsActive {
			// Reactivate if exists but inactive
			existing.IsActive = true
			if err := uc.shareholderTypeRepo.Update(existing); err != nil {
				return nil, fmt.Errorf("gagal mengaktifkan kembali jenis pemegang saham: %w", err)
			}
			return existing, nil
		}
		return nil, errors.New("jenis pemegang saham dengan nama tersebut sudah ada")
	}

	// Create new
	shareholderType := &domain.ShareholderTypeModel{
		ID:        uuid.GenerateUUID(),
		Name:      name,
		IsActive:  true,
		UsageCount: 0,
		CreatedBy: createdBy,
	}

	if err := uc.shareholderTypeRepo.Create(shareholderType); err != nil {
		return nil, fmt.Errorf("gagal membuat jenis pemegang saham: %w", err)
	}

	return shareholderType, nil
}

func (uc *shareholderTypeUseCase) GetShareholderTypeByID(id string) (*domain.ShareholderTypeModel, error) {
	return uc.shareholderTypeRepo.GetByID(id)
}

func (uc *shareholderTypeUseCase) GetShareholderTypeByName(name string) (*domain.ShareholderTypeModel, error) {
	return uc.shareholderTypeRepo.GetByName(name)
}

func (uc *shareholderTypeUseCase) GetAllShareholderTypes(includeInactive bool) ([]domain.ShareholderTypeModel, error) {
	return uc.shareholderTypeRepo.GetAll(includeInactive)
}

func (uc *shareholderTypeUseCase) GetActiveShareholderTypes() ([]domain.ShareholderTypeModel, error) {
	return uc.shareholderTypeRepo.GetActive()
}

func (uc *shareholderTypeUseCase) UpdateShareholderType(id string, name *string, isActive *bool) (*domain.ShareholderTypeModel, error) {
	shareholderType, err := uc.shareholderTypeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("jenis pemegang saham tidak ditemukan: %w", err)
	}

	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, errors.New("nama jenis pemegang saham tidak boleh kosong")
		}

		// Check if name already exists (case-insensitive, excluding current)
		existing, _ := uc.shareholderTypeRepo.GetByName(trimmedName)
		if existing != nil && existing.ID != id {
			return nil, errors.New("jenis pemegang saham dengan nama tersebut sudah ada")
		}

		shareholderType.Name = trimmedName
	}

	if isActive != nil {
		shareholderType.IsActive = *isActive
	}

	if err := uc.shareholderTypeRepo.Update(shareholderType); err != nil {
		return nil, fmt.Errorf("gagal mengupdate jenis pemegang saham: %w", err)
	}

	return shareholderType, nil
}

func (uc *shareholderTypeUseCase) DeleteShareholderType(id, requesterRole string) error {
	shareholderType, err := uc.shareholderTypeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("jenis pemegang saham tidak ditemukan: %w", err)
	}

	// Check usage count
	usageCount, err := uc.shareholderTypeRepo.CountUsage(id)
	if err != nil {
		return fmt.Errorf("gagal menghitung penggunaan: %w", err)
	}

	if usageCount > 0 {
		// Soft delete only (set is_active = false)
		shareholderType.IsActive = false
		if err := uc.shareholderTypeRepo.Update(shareholderType); err != nil {
			return fmt.Errorf("gagal menonaktifkan jenis pemegang saham: %w", err)
		}
		return nil
	}

	// Hard delete if not in use and requester is superadmin
	if strings.ToLower(requesterRole) == "superadmin" {
		return uc.shareholderTypeRepo.SoftDelete(id)
	}

	// For non-superadmin, just soft delete
	shareholderType.IsActive = false
	return uc.shareholderTypeRepo.Update(shareholderType)
}

