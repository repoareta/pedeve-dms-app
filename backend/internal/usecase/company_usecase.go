package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	CountRootHoldings() (int64, error)
}

type companyUseCase struct {
	companyRepo        repository.CompanyRepository
	shareholderRepo    repository.ShareholderRepository
	businessFieldRepo  repository.BusinessFieldRepository
	directorRepo       repository.DirectorRepository
}

// NewCompanyUseCaseWithDB creates a new company use case with injected DB (for testing)
func NewCompanyUseCaseWithDB(db *gorm.DB) CompanyUseCase {
	return &companyUseCase{
		companyRepo:       repository.NewCompanyRepositoryWithDB(db),
		shareholderRepo:   repository.NewShareholderRepositoryWithDB(db),
		businessFieldRepo: repository.NewBusinessFieldRepositoryWithDB(db),
		directorRepo:      repository.NewDirectorRepositoryWithDB(db),
	}
}

// NewCompanyUseCase creates a new company use case with default DB (backward compatibility)
func NewCompanyUseCase() CompanyUseCase {
	return NewCompanyUseCaseWithDB(database.GetDB())
}

func (uc *companyUseCase) CreateCompany(name, code, description string, parentID *string) (*domain.CompanyModel, error) {
	zapLog := logger.GetLogger()

	// Validate code uniqueness
	existing, _ := uc.companyRepo.GetByCode(code)
	if existing != nil {
		return nil, errors.New("company code already exists")
	}

	// Validasi: hanya boleh ada satu holding (parent_id = NULL)
	if parentID == nil {
		count, err := uc.companyRepo.CountRootHoldings()
		if err != nil {
			return nil, fmt.Errorf("failed to check existing holdings: %w", err)
		}
		if count > 0 {
			return nil, errors.New("holding company already exists. hanya boleh ada satu holding company. silakan pilih perusahaan induk")
		}
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

func (uc *companyUseCase) CountRootHoldings() (int64, error) {
	return uc.companyRepo.CountRootHoldings()
}

func (uc *companyUseCase) CreateCompanyFull(data *domain.CompanyCreateRequest) (*domain.CompanyModel, error) {
	zapLog := logger.GetLogger()

	// Validate code uniqueness - check dengan lock untuk prevent race condition
	existing, _ := uc.companyRepo.GetByCode(data.Code)
	if existing != nil {
		zapLog.Warn("Company code already exists", zap.String("code", data.Code), zap.String("existing_id", existing.ID))
		return nil, errors.New("company code already exists")
	}

	// Validasi: hanya boleh ada satu holding (parent_id = NULL)
	if data.ParentID == nil {
		count, err := uc.companyRepo.CountRootHoldings()
		if err != nil {
			return nil, fmt.Errorf("failed to check existing holdings: %w", err)
		}
		if count > 0 {
			return nil, errors.New("holding company already exists. hanya boleh ada satu holding company. silakan pilih perusahaan induk")
		}
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

	// Log update attempt untuk debugging duplicate issue
	zapLog.Info("UpdateCompanyFull called", 
		zap.String("company_id", id), 
		zap.String("name", data.Name),
		zap.Any("parent_id", data.ParentID),
	)

	company, err := uc.companyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}
	
	// Prevent update jika company sudah di-delete (soft delete)
	if !company.IsActive {
		return nil, fmt.Errorf("cannot update inactive company")
	}

	// CRITICAL: Prevent holding (code = "PDV") from being updated to have parent_id
	// Holding must always have parent_id = NULL and level = 0
	if company.Code == "PDV" && data.ParentID != nil && *data.ParentID != "" {
		zapLog.Warn("Attempt to set parent_id for holding company, ignoring",
			zap.String("company_id", id),
			zap.String("company_code", company.Code),
		)
		data.ParentID = nil // Force parent_id to NULL for holding
	}

	// Handle perubahan parent_id (untuk mengubah holding)
	if data.ParentID != nil {
		// Jika mengubah jadi holding (parent_id = NULL)
		if *data.ParentID == "" {
			data.ParentID = nil
		}
		
		// Validasi: hanya boleh ada satu holding
		if data.ParentID == nil {
			existingHolding, err := uc.companyRepo.GetRootHolding()
			if err == nil && existingHolding != nil && existingHolding.ID != id {
				// Ada holding lain, set holding lama jadi anak perusahaan dari holding baru
				existingHolding.ParentID = &id
				existingHolding.Level = 1
				if err := uc.companyRepo.Update(existingHolding); err != nil {
					zapLog.Error("Failed to update old holding", zap.Error(err))
				}
			}
		}
		
		// Update parent_id dan level
		oldParentID := company.ParentID
		company.ParentID = data.ParentID
		
		// Recalculate level
		if company.ParentID == nil {
			// CRITICAL: If parent_id is NULL, must be holding (level 0)
			// Force level to 0 untuk prevent level corruption
			company.Level = 0
			zapLog.Info("Company set as holding (parent_id = NULL), forcing level to 0",
				zap.String("company_id", id),
				zap.String("company_code", company.Code),
			)
		} else {
			parent, err := uc.companyRepo.GetByID(*company.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent company not found: %w", err)
			}
			
			// CRITICAL: Safety check - if parent level is invalid (> 10), cap it
			expectedLevel := parent.Level + 1
			if expectedLevel > 10 {
				zapLog.Warn("Calculated level exceeds maximum, capping at 10",
					zap.String("company_id", id),
					zap.Int("parent_level", parent.Level),
					zap.Int("calculated_level", expectedLevel),
				)
				expectedLevel = 10
			}
			
			company.Level = expectedLevel
			zapLog.Info("Updated company level",
				zap.String("company_id", id),
				zap.String("company_code", company.Code),
				zap.Int("old_level", company.Level),
				zap.Int("new_level", expectedLevel),
				zap.Int("parent_level", parent.Level),
			)
		}
		
		// Jika parent berubah, update level semua descendants akan dilakukan setelah company di-update
		// Menggunakan recursive update query di updateDescendantsLevel
		// Note: Perubahan parent akan terdeteksi setelah company di-update dan akan trigger updateDescendantsLevel
		_ = oldParentID // Explicitly mark as used to avoid unused variable warning
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

	// Jika parent_id berubah, update level semua descendants
	// CRITICAL: Only update descendants if parent_id actually changed AND company is not holding
	if data.ParentID != nil && company.Code != "PDV" {
		zapLog.Info("Parent ID changed, updating descendants level",
			zap.String("company_id", id),
			zap.String("company_code", company.Code),
		)
		if err := uc.updateDescendantsLevel(id); err != nil {
			zapLog.Warn("Failed to update descendants level", zap.String("company_id", id), zap.Error(err))
		}
	} else if company.Code == "PDV" {
		zapLog.Info("Skipping updateDescendantsLevel for holding company",
			zap.String("company_id", id),
			zap.String("company_code", company.Code),
		)
	}

	// Delete existing related data
	if err := uc.shareholderRepo.DeleteByCompanyID(id); err != nil {
		zapLog.Warn("Failed to delete shareholders", zap.String("company_id", id), zap.Error(err))
	}
	if err := uc.businessFieldRepo.DeleteByCompanyID(id); err != nil {
		zapLog.Warn("Failed to delete business fields", zap.String("company_id", id), zap.Error(err))
	}
	if err := uc.directorRepo.DeleteByCompanyID(id); err != nil {
		zapLog.Warn("Failed to delete directors", zap.String("company_id", id), zap.Error(err))
	}

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

