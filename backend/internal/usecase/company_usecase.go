package usecase

import (
	"errors"
	"fmt"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"github.com/Fajarriswandi/dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// CompanyUseCase interface untuk company operations
type CompanyUseCase interface {
	CreateCompany(name, code, description string, parentID *string) (*domain.CompanyModel, error)
	GetCompanyByID(id string) (*domain.CompanyModel, error)
	GetCompanyByCode(code string) (*domain.CompanyModel, error)
	GetAllCompanies() ([]domain.CompanyModel, error)
	GetCompanyChildren(id string) ([]domain.CompanyModel, error)
	GetCompanyDescendants(id string) ([]domain.CompanyModel, error)
	GetCompanyAncestors(id string) ([]domain.CompanyModel, error)
	UpdateCompany(id, name, description string) (*domain.CompanyModel, error)
	DeleteCompany(id string) error
	ValidateCompanyAccess(userCompanyID, targetCompanyID string) (bool, error)
}

type companyUseCase struct {
	companyRepo repository.CompanyRepository
}

// NewCompanyUseCase creates a new company use case
func NewCompanyUseCase() CompanyUseCase {
	return &companyUseCase{
		companyRepo: repository.NewCompanyRepository(),
	}
}

func (uc *companyUseCase) CreateCompany(name, code, description string, parentID *string) (*domain.CompanyModel, error) {
	zapLog := logger.GetLogger()

	// Validate code uniqueness
	existing, _ := uc.companyRepo.GetByCode(code)
	if existing != nil {
		return nil, errors.New("company code already exists")
	}

	// Determine level
	level := 0
	if parentID != nil {
		parent, err := uc.companyRepo.GetByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("parent company not found: %w", err)
		}
		level = parent.Level + 1
	}

	company := &domain.CompanyModel{
		ID:          uuid.GenerateUUID(),
		Name:        name,
		Code:        code,
		Description: description,
		ParentID:    parentID,
		Level:       level,
		IsActive:    true,
	}

	if err := uc.companyRepo.Create(company); err != nil {
		zapLog.Error("Failed to create company", zap.Error(err))
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	return company, nil
}

func (uc *companyUseCase) GetCompanyByID(id string) (*domain.CompanyModel, error) {
	return uc.companyRepo.GetByID(id)
}

func (uc *companyUseCase) GetCompanyByCode(code string) (*domain.CompanyModel, error) {
	return uc.companyRepo.GetByCode(code)
}

func (uc *companyUseCase) GetAllCompanies() ([]domain.CompanyModel, error) {
	return uc.companyRepo.GetAll()
}

func (uc *companyUseCase) GetCompanyChildren(id string) ([]domain.CompanyModel, error) {
	return uc.companyRepo.GetChildren(id)
}

func (uc *companyUseCase) GetCompanyDescendants(id string) ([]domain.CompanyModel, error) {
	return uc.companyRepo.GetDescendants(id)
}

func (uc *companyUseCase) GetCompanyAncestors(id string) ([]domain.CompanyModel, error) {
	return uc.companyRepo.GetAncestors(id)
}

func (uc *companyUseCase) UpdateCompany(id, name, description string) (*domain.CompanyModel, error) {
	company, err := uc.companyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	company.Name = name
	company.Description = description

	if err := uc.companyRepo.Update(company); err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	return company, nil
}

func (uc *companyUseCase) DeleteCompany(id string) error {
	// Soft delete: set is_active = false
	return uc.companyRepo.Delete(id)
}

func (uc *companyUseCase) ValidateCompanyAccess(userCompanyID, targetCompanyID string) (bool, error) {
	// If user's company is the same as target, allow
	if userCompanyID == targetCompanyID {
		return true, nil
	}

	// Check if target company is a descendant of user's company
	return uc.companyRepo.IsDescendantOf(targetCompanyID, userCompanyID)
}

