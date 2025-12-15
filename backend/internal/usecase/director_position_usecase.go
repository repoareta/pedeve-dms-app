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

// DirectorPositionUseCase interface untuk director position operations
type DirectorPositionUseCase interface {
	CreateDirectorPosition(name, createdBy string) (*domain.DirectorPositionModel, error)
	GetDirectorPositionByID(id string) (*domain.DirectorPositionModel, error)
	GetDirectorPositionByName(name string) (*domain.DirectorPositionModel, error)
	GetAllDirectorPositions(includeInactive bool) ([]domain.DirectorPositionModel, error)
	GetActiveDirectorPositions() ([]domain.DirectorPositionModel, error)
	UpdateDirectorPosition(id string, name *string, isActive *bool) (*domain.DirectorPositionModel, error)
	DeleteDirectorPosition(id, requesterRole string) error // Soft delete, only if not in use
}

type directorPositionUseCase struct {
	directorPositionRepo repository.DirectorPositionRepository
	directorRepo         repository.DirectorRepository
}

// NewDirectorPositionUseCaseWithDB creates a new director position use case with injected DB (for testing)
func NewDirectorPositionUseCaseWithDB(db *gorm.DB) DirectorPositionUseCase {
	return &directorPositionUseCase{
		directorPositionRepo: repository.NewDirectorPositionRepositoryWithDB(db),
		directorRepo:         repository.NewDirectorRepositoryWithDB(db),
	}
}

// NewDirectorPositionUseCase creates a new director position use case with default DB
func NewDirectorPositionUseCase() DirectorPositionUseCase {
	return NewDirectorPositionUseCaseWithDB(database.GetDB())
}

func (uc *directorPositionUseCase) CreateDirectorPosition(name, createdBy string) (*domain.DirectorPositionModel, error) {
	// Trim and validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama jabatan tidak boleh kosong")
	}

	// Check if already exists (case-insensitive)
	existing, _ := uc.directorPositionRepo.GetByName(name)
	if existing != nil {
		if !existing.IsActive {
			// Reactivate if exists but inactive
			existing.IsActive = true
			if err := uc.directorPositionRepo.Update(existing); err != nil {
				return nil, fmt.Errorf("gagal mengaktifkan kembali jabatan: %w", err)
			}
			return existing, nil
		}
		// Return existing instead of error to prevent duplicate key error
		// This handles race conditions where multiple requests try to create the same position
		return existing, nil
	}

	// Create new
	directorPosition := &domain.DirectorPositionModel{
		ID:         uuid.GenerateUUID(),
		Name:       name,
		IsActive:   true,
		UsageCount: 0,
		CreatedBy:  createdBy,
	}

	if err := uc.directorPositionRepo.Create(directorPosition); err != nil {
		// Check if error is due to duplicate key (case-sensitive database constraint)
		// If so, try to find existing record (case-insensitive) and return it
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			existing, findErr := uc.directorPositionRepo.GetByName(name)
			if findErr == nil && existing != nil {
				// If found, reactivate if inactive
				if !existing.IsActive {
					existing.IsActive = true
					if updateErr := uc.directorPositionRepo.Update(existing); updateErr == nil {
						return existing, nil
					}
				}
				return existing, nil
			}
		}
		return nil, fmt.Errorf("gagal membuat jabatan: %w", err)
	}

	return directorPosition, nil
}

func (uc *directorPositionUseCase) GetDirectorPositionByID(id string) (*domain.DirectorPositionModel, error) {
	return uc.directorPositionRepo.GetByID(id)
}

func (uc *directorPositionUseCase) GetDirectorPositionByName(name string) (*domain.DirectorPositionModel, error) {
	return uc.directorPositionRepo.GetByName(name)
}

func (uc *directorPositionUseCase) GetAllDirectorPositions(includeInactive bool) ([]domain.DirectorPositionModel, error) {
	return uc.directorPositionRepo.GetAll(includeInactive)
}

func (uc *directorPositionUseCase) GetActiveDirectorPositions() ([]domain.DirectorPositionModel, error) {
	return uc.directorPositionRepo.GetActive()
}

func (uc *directorPositionUseCase) UpdateDirectorPosition(id string, name *string, isActive *bool) (*domain.DirectorPositionModel, error) {
	directorPosition, err := uc.directorPositionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("jabatan tidak ditemukan: %w", err)
	}

	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, errors.New("nama jabatan tidak boleh kosong")
		}

		// Check if name already exists (case-insensitive, excluding current)
		existing, _ := uc.directorPositionRepo.GetByName(trimmedName)
		if existing != nil && existing.ID != id {
			return nil, errors.New("jabatan dengan nama tersebut sudah ada")
		}

		directorPosition.Name = trimmedName
	}

	if isActive != nil {
		directorPosition.IsActive = *isActive
	}

	if err := uc.directorPositionRepo.Update(directorPosition); err != nil {
		return nil, fmt.Errorf("gagal mengupdate jabatan: %w", err)
	}

	return directorPosition, nil
}

func (uc *directorPositionUseCase) DeleteDirectorPosition(id, requesterRole string) error {
	directorPosition, err := uc.directorPositionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("jabatan tidak ditemukan: %w", err)
	}

	// Check usage count
	usageCount, err := uc.directorPositionRepo.CountUsage(id)
	if err != nil {
		return fmt.Errorf("gagal menghitung penggunaan: %w", err)
	}

	if usageCount > 0 {
		// Soft delete only (set is_active = false)
		directorPosition.IsActive = false
		if err := uc.directorPositionRepo.Update(directorPosition); err != nil {
			return fmt.Errorf("gagal menonaktifkan jabatan: %w", err)
		}
		return nil
	}

	// Hard delete if not in use and requester is superadmin
	if strings.ToLower(requesterRole) == "superadmin" {
		return uc.directorPositionRepo.SoftDelete(id)
	}

	// For non-superadmin, just soft delete
	directorPosition.IsActive = false
	return uc.directorPositionRepo.Update(directorPosition)
}

