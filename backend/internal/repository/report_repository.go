package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// ReportRepository interface untuk report operations
type ReportRepository interface {
	Create(report *domain.ReportModel) error
	GetByID(id string) (*domain.ReportModel, error)
	GetAll() ([]domain.ReportModel, error)
	GetByCompanyID(companyID string) ([]domain.ReportModel, error)
	GetByCompanyIDAndPeriod(companyID, period string) (*domain.ReportModel, error)
	GetByCompanyIDs(companyIDs []string) ([]domain.ReportModel, error) // For admin to get reports from their company and children
	Update(report *domain.ReportModel) error
	Delete(id string) error
	DeleteAll() error // For reset functionality
	Count() (int64, error)
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepositoryWithDB creates a new report repository with injected DB (for testing)
func NewReportRepositoryWithDB(db *gorm.DB) ReportRepository {
	return &reportRepository{
		db: db,
	}
}

// NewReportRepository creates a new report repository with default DB
func NewReportRepository() ReportRepository {
	return NewReportRepositoryWithDB(database.GetDB())
}

func (r *reportRepository) Create(report *domain.ReportModel) error {
	return r.db.Create(report).Error
}

func (r *reportRepository) GetByID(id string) (*domain.ReportModel, error) {
	var report domain.ReportModel
	err := r.db.Preload("Company").Preload("Inputter").Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) GetAll() ([]domain.ReportModel, error) {
	var reports []domain.ReportModel
	err := r.db.Preload("Company").Preload("Inputter").Order("period DESC, created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *reportRepository) GetByCompanyID(companyID string) ([]domain.ReportModel, error) {
	var reports []domain.ReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ?", companyID).
		Order("period DESC, created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *reportRepository) GetByCompanyIDAndPeriod(companyID, period string) (*domain.ReportModel, error) {
	var report domain.ReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ? AND period = ?", companyID, period).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) GetByCompanyIDs(companyIDs []string) ([]domain.ReportModel, error) {
	var reports []domain.ReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id IN ?", companyIDs).
		Order("period DESC, created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *reportRepository) Update(report *domain.ReportModel) error {
	return r.db.Save(report).Error
}

func (r *reportRepository) Delete(id string) error {
	return r.db.Delete(&domain.ReportModel{}, "id = ?", id).Error
}

func (r *reportRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM reports").Error
}

func (r *reportRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.ReportModel{}).Count(&count).Error
	return count, err
}

