package usecase

import (
	"errors"
	"fmt"
	"os"
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
	companyRepo       repository.CompanyRepository
	shareholderRepo   repository.ShareholderRepository
	businessFieldRepo repository.BusinessFieldRepository
	directorRepo      repository.DirectorRepository
	documentRepo      repository.DocumentRepository
}

// NewCompanyUseCaseWithDB creates a new company use case with injected DB (for testing)
func NewCompanyUseCaseWithDB(db *gorm.DB) CompanyUseCase {
	return &companyUseCase{
		companyRepo:       repository.NewCompanyRepositoryWithDB(db),
		shareholderRepo:   repository.NewShareholderRepositoryWithDB(db),
		businessFieldRepo: repository.NewBusinessFieldRepositoryWithDB(db),
		directorRepo:      repository.NewDirectorRepositoryWithDB(db),
		documentRepo:      repository.NewDocumentRepositoryWithDB(db),
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

	// Validate: Only one company with parentID == nil can exist
	// This prevents multiple root companies (companies without parent)
	// Note: Level 0 hanya untuk code "PDV", tapi validasi ini berlaku untuk semua parentID == nil
	if parentID == nil {
		existingHolding, err := uc.companyRepo.GetRootHolding()
		if err == nil && existingHolding != nil {
			return nil, errors.New("holding company already exists")
		}
	}

	// Determine level
	// CRITICAL: Level 0 hanya untuk holding company yang sebenarnya (misalnya code = "PDV")
	// Perusahaan tanpa parent_id yang baru dibuat menggunakan level 1 sebagai default (temporary)
	// Level akan di-recalculate dengan benar setelah parent_id di-set nanti
	level := 1 // Default level untuk perusahaan tanpa parent_id (bukan level 0)
	if parentID != nil {
		parent, err := uc.companyRepo.GetByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("parent company not found: %w", err)
		}
		level = parent.Level + 1
	}
	// Note: Level 0 hanya untuk code "PDV", tidak di-set di CreateCompany
	// Level akan di-update dengan benar di UpdateCompanyFull jika code = "PDV"

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

	// Otomatis buat folder untuk perusahaan dengan nama perusahaan
	folder := &domain.DocumentFolderModel{
		ID:        uuid.GenerateUUID(),
		Name:      name, // Nama folder sama dengan nama perusahaan
		CompanyID: &company.ID,
		ParentID:  nil, // Folder root untuk perusahaan
		CreatedBy: "",  // System-created, tidak ada user creator
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.documentRepo.CreateFolder(folder); err != nil {
		zapLog.Warn("Failed to create folder for company",
			zap.String("company_id", company.ID),
			zap.String("company_name", name),
			zap.Error(err))
		// Tidak return error karena company sudah dibuat, folder bisa dibuat manual nanti
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

	// #region agent log
	logFile, _ := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		parentIDValue := "nil"
		if data.ParentID != nil {
			parentIDValue = *data.ParentID
		}
		logEntry := fmt.Sprintf(`{"id":"log_%d","timestamp":%d,"location":"company_usecase.go:163","message":"CreateCompanyFull entry","data":{"code":"%s","parent_id":"%s","name":"%s"},"sessionId":"debug-session","runId":"run1","hypothesisId":"A"}`+"\n", time.Now().UnixNano(), time.Now().UnixMilli(), data.Code, parentIDValue, data.Name)
		_, _ = logFile.WriteString(logEntry)
		_ = logFile.Close()
	}
	// #endregion

	// Validate code uniqueness - check dengan lock untuk prevent race condition
	existing, _ := uc.companyRepo.GetByCode(data.Code)
	if existing != nil {
		zapLog.Warn("Company code already exists", zap.String("code", data.Code), zap.String("existing_id", existing.ID))
		return nil, errors.New("company code already exists")
	}

	// #region agent log
	logFile2, _ := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile2 != nil {
		parentIDValue := "nil"
		if data.ParentID != nil {
			parentIDValue = *data.ParentID
		}
		logEntry := fmt.Sprintf(`{"id":"log_%d","timestamp":%d,"location":"company_usecase.go:174","message":"Before holding validation check","data":{"parent_id":"%s","parent_id_is_nil":%t},"sessionId":"debug-session","runId":"run1","hypothesisId":"A"}`+"\n", time.Now().UnixNano(), time.Now().UnixMilli(), parentIDValue, data.ParentID == nil)
		_, _ = logFile2.WriteString(logEntry)
		_ = logFile2.Close()
	}
	// #endregion

	// REMOVED: Validasi yang memblokir pembuatan perusahaan tanpa parent_id
	// Sebelumnya: validasi ini memblokir pembuatan perusahaan baru jika parent_id == nil dan sudah ada holding company
	// Sekarang: perusahaan bisa dibuat tanpa parent_id, dan parent_id akan di-setup nanti secara terpisah
	// Validasi ini dihapus karena user ingin bisa membuat perusahaan baru tanpa parent_id di awal

	// Determine level
	// CRITICAL: Level 0 hanya untuk holding company yang sebenarnya (misalnya code = "PDV")
	// Perusahaan tanpa parent_id yang baru dibuat menggunakan level 1 sebagai default (temporary)
	// Level akan di-recalculate dengan benar setelah parent_id di-set nanti
	level := 1 // Default level untuk perusahaan tanpa parent_id (bukan level 0)
	if data.ParentID != nil {
		parent, err := uc.companyRepo.GetByID(*data.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent company not found: %w", err)
		}
		level = parent.Level + 1
	}

	// Set default currency to IDR if not provided
	currency := data.Currency
	if currency == "" {
		currency = "IDR"
		zapLog.Info("Company currency not provided in create request, using default IDR",
			zap.String("code", data.Code),
		)
	} else {
		zapLog.Info("Creating company with currency",
			zap.String("code", data.Code),
			zap.String("currency", currency),
		)
	}

	company := &domain.CompanyModel{
		ID:                  uuid.GenerateUUID(),
		Name:                data.Name,
		ShortName:           data.ShortName,
		Code:                data.Code,
		Description:         data.Description,
		NPWP:                data.NPWP,
		NIB:                 data.NIB,
		Status:              data.Status,
		Logo:                data.Logo,
		Phone:               data.Phone,
		Fax:                 data.Fax,
		Email:               data.Email,
		Website:             data.Website,
		Address:             data.Address,
		OperationalAddress:  data.OperationalAddress,
		AuthorizedCapital:   data.AuthorizedCapital,
		PaidUpCapital:       data.PaidUpCapital,
		Currency:            currency,
		ParentID:            data.ParentID,
		MainParentCompanyID: data.MainParentCompany,
		Level:               level,
		IsActive:            data.Status == "Aktif",
	}

	if err := uc.companyRepo.Create(company); err != nil {
		zapLog.Error("Failed to create company", zap.Error(err))
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Otomatis buat folder untuk perusahaan dengan nama perusahaan
	folder := &domain.DocumentFolderModel{
		ID:        uuid.GenerateUUID(),
		Name:      data.Name, // Nama folder sama dengan nama perusahaan
		CompanyID: &company.ID,
		ParentID:  nil, // Folder root untuk perusahaan
		CreatedBy: "",  // System-created, tidak ada user creator
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.documentRepo.CreateFolder(folder); err != nil {
		zapLog.Warn("Failed to create folder for company",
			zap.String("company_id", company.ID),
			zap.String("company_name", data.Name),
			zap.Error(err))
		// Tidak return error karena company sudah dibuat, folder bisa dibuat manual nanti
	}

	// Create shareholders
	for _, sh := range data.Shareholders {
		shareholder := &domain.ShareholderModel{
			ID:                   uuid.GenerateUUID(),
			CompanyID:            company.ID,
			ShareholderCompanyID: sh.ShareholderCompanyID, // ID perusahaan pemegang saham (nullable)
			Type:                 sh.Type,
			Name:                 sh.Name,
			IdentityNumber:       sh.IdentityNumber,
			OwnershipPercent:     sh.OwnershipPercent,
			ShareCount:           sh.ShareCount,
			ShareSheetCount:      sh.ShareSheetCount,
			ShareValuePerSheet:   sh.ShareValuePerSheet,
			IsMainParent:         sh.IsMainParent,
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
		var endDate *time.Time
		if dir.EndDate != nil && !dir.EndDate.Time.IsZero() {
			endDate = &dir.EndDate.Time
		}
		director := &domain.DirectorModel{
			ID:              uuid.GenerateUUID(),
			CompanyID:       company.ID,
			Position:        dir.Position,
			FullName:        dir.FullName,
			KTP:             dir.KTP,
			NPWP:            dir.NPWP,
			StartDate:       startDate,
			EndDate:         endDate,
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
	// CRITICAL: Always update parent_id, even if nil (to remove parent)
	// Frontend must always send parent_id field (null if empty, to allow removing parent)
	// Normalize parent_id: convert empty string to nil
	var newParentID *string
	if data.ParentID != nil {
		if *data.ParentID == "" {
			newParentID = nil
		} else {
			newParentID = data.ParentID
		}
	} else {
		// If ParentID is nil in request, it means user wants to remove parent (set to NULL)
		// Frontend should always send parent_id field (null if empty)
		newParentID = nil
	}

	// Only update parent_id if it's different from current value
	oldParentID := company.ParentID
	parentIDChanged := false
	if (oldParentID == nil && newParentID != nil) ||
		(oldParentID != nil && newParentID == nil) ||
		(oldParentID != nil && newParentID != nil && *oldParentID != *newParentID) {
		parentIDChanged = true
	}

	if parentIDChanged {
		// Validasi: hanya boleh ada satu holding
		if newParentID == nil {
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
		company.ParentID = newParentID

		// Recalculate level
		if company.ParentID == nil {
			// CRITICAL: Level 0 hanya untuk holding company yang sebenarnya (misalnya code = "PDV")
			// Perusahaan tanpa parent_id menggunakan level 1 sebagai default (bukan level 0)
			// Level akan di-recalculate dengan benar setelah parent_id di-set nanti
			// Hanya holding company yang sebenarnya yang boleh level 0
			if company.Code == "PDV" {
				// Holding company yang sebenarnya: level 0
				company.Level = 0
				zapLog.Info("Holding company (PDV) set with level 0",
					zap.String("company_id", id),
					zap.String("company_code", company.Code),
				)
			} else {
				// Perusahaan tanpa parent_id (bukan holding): level 1 sebagai default
				company.Level = 1
				zapLog.Info("Company without parent_id set with level 1 (default, not holding)",
					zap.String("company_id", id),
					zap.String("company_code", company.Code),
				)
			}
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
	} else {
		// Parent ID tidak berubah, tetap gunakan nilai yang ada
		zapLog.Info("Parent ID unchanged",
			zap.String("company_id", id),
			zap.Any("parent_id", company.ParentID),
		)
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
	company.AuthorizedCapital = data.AuthorizedCapital
	company.PaidUpCapital = data.PaidUpCapital
	// Update currency - always update, default to IDR if empty
	// CRITICAL: Always update currency field, even if empty (will default to IDR)
	// This ensures that if frontend sends empty string or doesn't send field, we still update it
	if data.Currency != "" {
		company.Currency = data.Currency
		zapLog.Info("Updating company currency",
			zap.String("company_id", id),
			zap.String("new_currency", data.Currency),
			zap.String("old_currency", company.Currency),
		)
	} else {
		// If currency is empty or not provided, set to default IDR
		oldCurrency := company.Currency
		company.Currency = "IDR"
		zapLog.Info("Company currency not provided or empty, setting to default IDR",
			zap.String("company_id", id),
			zap.String("old_currency", oldCurrency),
			zap.String("new_currency", "IDR"),
		)
	}
	company.MainParentCompanyID = data.MainParentCompany
	company.IsActive = data.Status == "Aktif"

	if err := uc.companyRepo.Update(company); err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	// Jika parent_id berubah, update level semua descendants
	// CRITICAL: Only update descendants if parent_id actually changed AND company is not holding
	if parentIDChanged && company.Code != "PDV" {
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

	// Delete existing related data (except directors - we'll handle them separately to preserve IDs)
	if err := uc.shareholderRepo.DeleteByCompanyID(id); err != nil {
		zapLog.Warn("Failed to delete shareholders", zap.String("company_id", id), zap.Error(err))
	}
	if err := uc.businessFieldRepo.DeleteByCompanyID(id); err != nil {
		zapLog.Warn("Failed to delete business fields", zap.String("company_id", id), zap.Error(err))
	}
	// Directors will be handled separately to preserve existing IDs and maintain document relationships

	// Create new shareholders
	for _, sh := range data.Shareholders {
		shareholder := &domain.ShareholderModel{
			ID:                   uuid.GenerateUUID(),
			CompanyID:            company.ID,
			ShareholderCompanyID: sh.ShareholderCompanyID, // ID perusahaan pemegang saham (nullable)
			Type:                 sh.Type,
			Name:                 sh.Name,
			IdentityNumber:       sh.IdentityNumber,
			OwnershipPercent:     sh.OwnershipPercent,
			ShareCount:           sh.ShareCount,
			ShareSheetCount:      sh.ShareSheetCount,
			ShareValuePerSheet:   sh.ShareValuePerSheet,
			IsMainParent:         sh.IsMainParent,
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

	// Update/create directors (preserve existing IDs to maintain document relationships)
	existingDirectors, err := uc.directorRepo.GetByCompanyID(id)
	if err != nil {
		zapLog.Warn("Failed to get existing directors", zap.String("company_id", id), zap.Error(err))
		existingDirectors = []domain.DirectorModel{}
	}

	// Create map of existing directors by identifier (full_name + ktp) for matching
	existingDirectorMap := make(map[string]*domain.DirectorModel)
	for i := range existingDirectors {
		key := existingDirectors[i].FullName + "|" + existingDirectors[i].KTP
		existingDirectorMap[key] = &existingDirectors[i]
	}

	// Track which directors are still in use
	usedDirectorIDs := make(map[string]bool)

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
		var endDate *time.Time
		if dir.EndDate != nil {
			dateOnly := *dir.EndDate
			if !dateOnly.IsZero() {
				endDate = &dateOnly.Time
			}
		}

		// Match director berdasarkan identifier (full_name + ktp)
		key := dir.FullName + "|" + dir.KTP
		existingDirector, exists := existingDirectorMap[key]

		if exists && existingDirector != nil {
			// Update existing director (preserve ID to maintain document relationships)
			existingDirector.Position = dir.Position
			existingDirector.FullName = dir.FullName
			existingDirector.KTP = dir.KTP
			existingDirector.NPWP = dir.NPWP
			existingDirector.StartDate = startDate
			existingDirector.EndDate = endDate
			existingDirector.DomicileAddress = dir.DomicileAddress
			existingDirector.CompanyID = company.ID

			// Update using GORM
			if err := uc.directorRepo.Update(existingDirector); err != nil {
				zapLog.Error("Failed to update director",
					zap.String("director_id", existingDirector.ID),
					zap.String("full_name", existingDirector.FullName),
					zap.Error(err))
			} else {
				usedDirectorIDs[existingDirector.ID] = true
				zapLog.Info("Updated existing director",
					zap.String("director_id", existingDirector.ID),
					zap.String("full_name", existingDirector.FullName))
			}
		} else {
			// Create new director
			director := &domain.DirectorModel{
				ID:              uuid.GenerateUUID(),
				CompanyID:       company.ID,
				Position:        dir.Position,
				FullName:        dir.FullName,
				KTP:             dir.KTP,
				NPWP:            dir.NPWP,
				StartDate:       startDate,
				EndDate:         endDate,
				DomicileAddress: dir.DomicileAddress,
			}
			if err := uc.directorRepo.Create(director); err != nil {
				zapLog.Error("Failed to create director", zap.Error(err))
			} else {
				usedDirectorIDs[director.ID] = true
				zapLog.Info("Created new director",
					zap.String("director_id", director.ID),
					zap.String("full_name", director.FullName))
			}
		}
	}

	// Delete directors that are no longer in the new data
	for _, existingDirector := range existingDirectors {
		if !usedDirectorIDs[existingDirector.ID] {
			if err := uc.directorRepo.Delete(existingDirector.ID); err != nil {
				zapLog.Warn("Failed to delete removed director",
					zap.String("director_id", existingDirector.ID),
					zap.String("full_name", existingDirector.FullName),
					zap.Error(err))
			} else {
				zapLog.Info("Deleted removed director",
					zap.String("director_id", existingDirector.ID),
					zap.String("full_name", existingDirector.FullName))
			}
		}
	}

	return company, nil
}
