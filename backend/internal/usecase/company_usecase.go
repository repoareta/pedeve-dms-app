package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"github.com/Fajarriswandi/dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// CompanyUseCase interface untuk company operations
type CompanyUseCase interface {
	CreateCompany(name, code, description string, parentID *string) (*domain.CompanyModel, error)
	CreateCompanyFull(data *domain.CompanyCreateRequest) (*domain.CompanyModel, error)
	GetCompanyByID(id string) (*domain.CompanyModel, error)
	GetCompanyByCode(code string) (*domain.CompanyModel, error)
	GetAllCompanies() ([]domain.CompanyModel, error)
	GetCompanyChildren(id string) ([]domain.CompanyModel, error)
	GetCompanyDescendants(id string) ([]domain.CompanyModel, error)
	GetCompanyAncestors(id string) ([]domain.CompanyModel, error)
	UpdateCompany(id, name, description string) (*domain.CompanyModel, error)
	UpdateCompanyFull(id string, data *domain.CompanyUpdateRequest) (*domain.CompanyModel, error)
	DeleteCompany(id string) error
	ValidateCompanyAccess(userCompanyID, targetCompanyID string) (bool, error)
}

type companyUseCase struct {
	companyRepo        repository.CompanyRepository
	shareholderRepo    repository.ShareholderRepository
	businessFieldRepo  repository.BusinessFieldRepository
	directorRepo       repository.DirectorRepository
}

// NewCompanyUseCase creates a new company use case
func NewCompanyUseCase() CompanyUseCase {
	return &companyUseCase{
		companyRepo:       repository.NewCompanyRepository(),
		shareholderRepo:   repository.NewShareholderRepository(),
		businessFieldRepo: repository.NewBusinessFieldRepository(),
		directorRepo:      repository.NewDirectorRepository(),
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

func (uc *companyUseCase) CreateCompanyFull(data *domain.CompanyCreateRequest) (*domain.CompanyModel, error) {
	zapLog := logger.GetLogger()

	// Validate code uniqueness
	existing, _ := uc.companyRepo.GetByCode(data.Code)
	if existing != nil {
		return nil, errors.New("company code already exists")
	}

	// Determine level
	level := 0
	if data.ParentID != nil {
		parent, err := uc.companyRepo.GetByID(*data.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent company not found: %w", err)
		}
		level = parent.Level + 1
	}

	company := &domain.CompanyModel{
		ID:                uuid.GenerateUUID(),
		Name:              data.Name,
		ShortName:         data.ShortName,
		Code:              data.Code,
		Description:       data.Description,
		NPWP:              data.NPWP,
		NIB:               data.NIB,
		Status:            data.Status,
		Logo:              data.Logo,
		Phone:             data.Phone,
		Fax:               data.Fax,
		Email:             data.Email,
		Website:           data.Website,
		Address:           data.Address,
		OperationalAddress: data.OperationalAddress,
		ParentID:          data.ParentID,
		MainParentCompanyID: data.MainParentCompany,
		Level:             level,
		IsActive:          data.Status == "Aktif",
	}

	if err := uc.companyRepo.Create(company); err != nil {
		zapLog.Error("Failed to create company", zap.Error(err))
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Create shareholders
	for _, sh := range data.Shareholders {
		shareholder := &domain.ShareholderModel{
			ID:              uuid.GenerateUUID(),
			CompanyID:       company.ID,
			Type:            sh.Type,
			Name:            sh.Name,
			IdentityNumber:  sh.IdentityNumber,
			OwnershipPercent: sh.OwnershipPercent,
			ShareCount:      sh.ShareCount,
			IsMainParent:    sh.IsMainParent,
		}
		if err := uc.shareholderRepo.Create(shareholder); err != nil {
			zapLog.Error("Failed to create shareholder", zap.Error(err))
		}
	}

	// Create main business field
	if data.MainBusiness != nil {
		var startOpDate *time.Time
		if data.MainBusiness.StartOperationDate != nil && !data.MainBusiness.StartOperationDate.Time.IsZero() {
			startOpDate = &data.MainBusiness.StartOperationDate.Time
		}
		businessField := &domain.BusinessFieldModel{
			ID:                   uuid.GenerateUUID(),
			CompanyID:            company.ID,
			IndustrySector:       data.MainBusiness.IndustrySector,
			KBLI:                 data.MainBusiness.KBLI,
			MainBusinessActivity: data.MainBusiness.MainBusinessActivity,
			AdditionalActivities: data.MainBusiness.AdditionalActivities,
			StartOperationDate:   startOpDate,
			IsMain:               true,
		}
		if err := uc.businessFieldRepo.Create(businessField); err != nil {
			zapLog.Error("Failed to create business field", zap.Error(err))
		}
	}

	// Create directors
	for _, dir := range data.Directors {
		var startDate *time.Time
		if dir.StartDate != nil && !dir.StartDate.Time.IsZero() {
			startDate = &dir.StartDate.Time
		}
		director := &domain.DirectorModel{
			ID:              uuid.GenerateUUID(),
			CompanyID:       company.ID,
			Position:        dir.Position,
			FullName:        dir.FullName,
			KTP:             dir.KTP,
			NPWP:            dir.NPWP,
			StartDate:       startDate,
			DomicileAddress: dir.DomicileAddress,
		}
		if err := uc.directorRepo.Create(director); err != nil {
			zapLog.Error("Failed to create director", zap.Error(err))
		}
	}

	return company, nil
}

func (uc *companyUseCase) UpdateCompanyFull(id string, data *domain.CompanyUpdateRequest) (*domain.CompanyModel, error) {
	zapLog := logger.GetLogger()

	company, err := uc.companyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	// Update company fields
	company.Name = data.Name
	company.ShortName = data.ShortName
	company.Description = data.Description
	company.NPWP = data.NPWP
	company.NIB = data.NIB
	company.Status = data.Status
	company.Logo = data.Logo
	company.Phone = data.Phone
	company.Fax = data.Fax
	company.Email = data.Email
	company.Website = data.Website
	company.Address = data.Address
	company.OperationalAddress = data.OperationalAddress
	company.MainParentCompanyID = data.MainParentCompany
	company.IsActive = data.Status == "Aktif"

	if err := uc.companyRepo.Update(company); err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	// Delete existing related data
	uc.shareholderRepo.DeleteByCompanyID(id)
	uc.businessFieldRepo.DeleteByCompanyID(id)
	uc.directorRepo.DeleteByCompanyID(id)

	// Create new shareholders
	for _, sh := range data.Shareholders {
		shareholder := &domain.ShareholderModel{
			ID:              uuid.GenerateUUID(),
			CompanyID:       company.ID,
			Type:            sh.Type,
			Name:            sh.Name,
			IdentityNumber:  sh.IdentityNumber,
			OwnershipPercent: sh.OwnershipPercent,
			ShareCount:      sh.ShareCount,
			IsMainParent:    sh.IsMainParent,
		}
		if err := uc.shareholderRepo.Create(shareholder); err != nil {
			zapLog.Error("Failed to create shareholder", zap.Error(err))
		}
	}

	// Create/update main business field
	if data.MainBusiness != nil {
		var startOpDate *time.Time
		if data.MainBusiness.StartOperationDate != nil {
			// DateOnly embedded time.Time, jadi bisa langsung diakses sebagai time.Time
			dateOnly := *data.MainBusiness.StartOperationDate
			if !dateOnly.IsZero() {
				// DateOnly adalah struct dengan embedded time.Time, akses field Time
				startOpDate = &dateOnly.Time
			}
		}
		businessField := &domain.BusinessFieldModel{
			ID:                   uuid.GenerateUUID(),
			CompanyID:            company.ID,
			IndustrySector:       data.MainBusiness.IndustrySector,
			KBLI:                 data.MainBusiness.KBLI,
			MainBusinessActivity: data.MainBusiness.MainBusinessActivity,
			AdditionalActivities: data.MainBusiness.AdditionalActivities,
			StartOperationDate:   startOpDate,
			IsMain:               true,
		}
		if err := uc.businessFieldRepo.Create(businessField); err != nil {
			zapLog.Error("Failed to create business field", zap.Error(err))
		}
	}

	// Create new directors
	for _, dir := range data.Directors {
		var startDate *time.Time
		if dir.StartDate != nil {
			// DateOnly embedded time.Time, jadi bisa langsung diakses sebagai time.Time
			dateOnly := *dir.StartDate
			if !dateOnly.IsZero() {
				// DateOnly adalah struct dengan embedded time.Time, akses field Time
				startDate = &dateOnly.Time
			}
		}
		director := &domain.DirectorModel{
			ID:              uuid.GenerateUUID(),
			CompanyID:       company.ID,
			Position:        dir.Position,
			FullName:        dir.FullName,
			KTP:             dir.KTP,
			NPWP:            dir.NPWP,
			StartDate:       startDate,
			DomicileAddress: dir.DomicileAddress,
		}
		if err := uc.directorRepo.Create(director); err != nil {
			zapLog.Error("Failed to create director", zap.Error(err))
		}
	}

	return company, nil
}

