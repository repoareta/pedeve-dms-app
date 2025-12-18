package usecase

import (
	"errors"
	"fmt"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"gorm.io/gorm"
)

// ReportUseCase interface untuk report operations
type ReportUseCase interface {
	CreateReport(data *domain.CreateReportRequest) (*domain.ReportModel, error)
	GetReportByID(id string) (*domain.ReportModel, error)
	GetAllReports(userRole string, userCompanyID *string) ([]domain.ReportModel, error)
	GetReportsByCompanyID(companyID string, userRole string, userCompanyID *string) ([]domain.ReportModel, error)
	UpdateReport(id string, data *domain.UpdateReportRequest) (*domain.ReportModel, error)
	DeleteReport(id string) error
	ValidateReportAccess(userRole string, userCompanyID *string, reportCompanyID string) (bool, error)
}

type reportUseCase struct {
	reportRepo  repository.ReportRepository
	companyRepo repository.CompanyRepository
	userRepo    repository.UserRepository
}

// NewReportUseCaseWithDB creates a new report use case with injected DB (for testing)
func NewReportUseCaseWithDB(db *gorm.DB) ReportUseCase {
	return &reportUseCase{
		reportRepo:  repository.NewReportRepositoryWithDB(db),
		companyRepo: repository.NewCompanyRepositoryWithDB(db),
		userRepo:    repository.NewUserRepositoryWithDB(db),
	}
}

// NewReportUseCase creates a new report use case with default DB
func NewReportUseCase() ReportUseCase {
	return NewReportUseCaseWithDB(database.GetDB())
}

func (uc *reportUseCase) CreateReport(data *domain.CreateReportRequest) (*domain.ReportModel, error) {
	// Validate company exists
	_, err := uc.companyRepo.GetByID(data.CompanyID)
	if err != nil {
		return nil, errors.New("company not found")
	}

	// Validate inputter if provided
	if data.InputterID != nil && *data.InputterID != "" {
		_, err := uc.userRepo.GetByID(*data.InputterID)
		if err != nil {
			return nil, errors.New("inputter user not found")
		}
	}

	// Check if report for this company and period already exists
	existing, _ := uc.reportRepo.GetByCompanyIDAndPeriod(data.CompanyID, data.Period)
	if existing != nil {
		return nil, errors.New("report for this company and period already exists")
	}

	// Validate period format (YYYY-MM)
	if len(data.Period) != 7 || data.Period[4] != '-' {
		return nil, errors.New("invalid period format. must be YYYY-MM")
	}

	// Create report
	report := &domain.ReportModel{
		ID:             uuid.GenerateUUID(),
		Period:         data.Period,
		CompanyID:      data.CompanyID,
		InputterID:     data.InputterID,
		Revenue:        data.Revenue,
		Opex:           data.Opex,
		NPAT:           data.NPAT,
		Dividend:       data.Dividend,
		FinancialRatio: data.FinancialRatio,
		Attachment:     data.Attachment,
		Remark:         data.Remark,
	}

	err = uc.reportRepo.Create(report)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Reload with relationships
	return uc.reportRepo.GetByID(report.ID)
}

func (uc *reportUseCase) GetReportByID(id string) (*domain.ReportModel, error) {
	return uc.reportRepo.GetByID(id)
}

func (uc *reportUseCase) GetAllReports(userRole string, userCompanyID *string) ([]domain.ReportModel, error) {
	// Superadmin can see all reports
	if utils.IsSuperAdminLike(userRole) {
		return uc.reportRepo.GetAll()
	}

	// Admin can see reports from their company and all children companies
	if userRole == "admin" && userCompanyID != nil {
		// Get all descendants (children, grandchildren, etc)
		descendants, err := uc.companyRepo.GetDescendants(*userCompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get company descendants: %w", err)
		}

		// Collect all company IDs (own company + descendants)
		companyIDs := []string{*userCompanyID}
		for _, desc := range descendants {
			companyIDs = append(companyIDs, desc.ID)
		}

		return uc.reportRepo.GetByCompanyIDs(companyIDs)
	}

	// Regular users can only see reports from their own company
	if userCompanyID != nil {
		return uc.reportRepo.GetByCompanyID(*userCompanyID)
	}

	return []domain.ReportModel{}, nil
}

func (uc *reportUseCase) GetReportsByCompanyID(companyID string, userRole string, userCompanyID *string) ([]domain.ReportModel, error) {
	// Validate access
	hasAccess, err := uc.ValidateReportAccess(userRole, userCompanyID, companyID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errors.New("access denied: you don't have permission to view reports for this company")
	}

	return uc.reportRepo.GetByCompanyID(companyID)
}

func (uc *reportUseCase) UpdateReport(id string, data *domain.UpdateReportRequest) (*domain.ReportModel, error) {
	// Get existing report
	report, err := uc.reportRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("report not found")
	}

	// Update fields if provided
	if data.Period != nil {
		// Validate period format
		if len(*data.Period) != 7 || (*data.Period)[4] != '-' {
			return nil, errors.New("invalid period format. must be YYYY-MM")
		}
		// Check if new period conflicts with existing report for same company
		if *data.Period != report.Period {
			existing, _ := uc.reportRepo.GetByCompanyIDAndPeriod(report.CompanyID, *data.Period)
			if existing != nil && existing.ID != id {
				return nil, errors.New("report for this company and period already exists")
			}
		}
		report.Period = *data.Period
	}

	if data.CompanyID != nil {
		// Validate company exists
		_, err := uc.companyRepo.GetByID(*data.CompanyID)
		if err != nil {
			return nil, errors.New("company not found")
		}
		report.CompanyID = *data.CompanyID
	}

	if data.InputterID != nil {
		if *data.InputterID != "" {
			// Validate inputter exists
			_, err := uc.userRepo.GetByID(*data.InputterID)
			if err != nil {
				return nil, errors.New("inputter user not found")
			}
		}
		report.InputterID = data.InputterID
	}

	if data.Revenue != nil {
		report.Revenue = *data.Revenue
	}

	if data.Opex != nil {
		report.Opex = *data.Opex
	}

	if data.NPAT != nil {
		report.NPAT = *data.NPAT
	}

	if data.Dividend != nil {
		report.Dividend = *data.Dividend
	}

	if data.FinancialRatio != nil {
		report.FinancialRatio = *data.FinancialRatio
	}

	if data.Attachment != nil {
		report.Attachment = data.Attachment
	}

	if data.Remark != nil {
		report.Remark = data.Remark
	}

	err = uc.reportRepo.Update(report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return uc.reportRepo.GetByID(id)
}

func (uc *reportUseCase) DeleteReport(id string) error {
	_, err := uc.reportRepo.GetByID(id)
	if err != nil {
		return errors.New("report not found")
	}

	return uc.reportRepo.Delete(id)
}

// ValidateReportAccess validates if user has access to report for a company
// Returns true if:
// - User is superadmin (can access all)
// - User is admin and companyID is their company or one of their descendants
// - User is regular user and companyID is their company
func (uc *reportUseCase) ValidateReportAccess(userRole string, userCompanyID *string, reportCompanyID string) (bool, error) {
	// Superadmin can access all
	if utils.IsSuperAdminLike(userRole) {
		return true, nil
	}

	// If user has no company, they can't access any reports (except superadmin)
	if userCompanyID == nil {
		return false, nil
	}

	// If report is for user's own company, allow access
	if *userCompanyID == reportCompanyID {
		return true, nil
	}

	// Admin can access reports from their company and all descendants
	if userRole == "admin" {
		isDescendant, err := uc.companyRepo.IsDescendantOf(reportCompanyID, *userCompanyID)
		if err != nil {
			return false, fmt.Errorf("failed to check company relationship: %w", err)
		}
		return isDescendant, nil
	}

	// Regular users can only access their own company's reports
	return false, nil
}
