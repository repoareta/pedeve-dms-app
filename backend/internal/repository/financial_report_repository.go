package repository

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

// FinancialReportRepository interface untuk financial report operations
type FinancialReportRepository interface {
	Create(report *domain.FinancialReportModel) error
	GetByID(id string) (*domain.FinancialReportModel, error)
	GetByCompanyID(companyID string) ([]domain.FinancialReportModel, error)
	GetRKAPByCompanyIDAndYear(companyID, year string) (*domain.FinancialReportModel, error)
	GetRealisasiByCompanyIDAndPeriod(companyID, period string) (*domain.FinancialReportModel, error)
	GetRealisasiByCompanyIDAndYear(companyID, year string) ([]domain.FinancialReportModel, error)
	GetRealisasiYTD(companyID, year, month string) (*domain.FinancialReportModel, error)
	Update(report *domain.FinancialReportModel) error
	Delete(id string) error
	DeleteAll() error // For reset functionality
	CountRKAPByCompanyIDAndYear(companyID, year string) (int64, error)
	GetRKAPYearsByCompanyID(companyID string) ([]string, error)
}

type financialReportRepository struct {
	db *gorm.DB
}

// NewFinancialReportRepositoryWithDB creates a new financial report repository with injected DB
func NewFinancialReportRepositoryWithDB(db *gorm.DB) FinancialReportRepository {
	return &financialReportRepository{
		db: db,
	}
}

// NewFinancialReportRepository creates a new financial report repository with default DB
func NewFinancialReportRepository() FinancialReportRepository {
	return NewFinancialReportRepositoryWithDB(database.GetDB())
}

func (r *financialReportRepository) Create(report *domain.FinancialReportModel) error {
	return r.db.Create(report).Error
}

func (r *financialReportRepository) GetByID(id string) (*domain.FinancialReportModel, error) {
	var report domain.FinancialReportModel
	err := r.db.Preload("Company").Preload("Inputter").Where("id = ?", id).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *financialReportRepository) GetByCompanyID(companyID string) ([]domain.FinancialReportModel, error) {
	var reports []domain.FinancialReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ?", companyID).
		Order("year DESC, period DESC, created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *financialReportRepository) GetRKAPByCompanyIDAndYear(companyID, year string) (*domain.FinancialReportModel, error) {
	var report domain.FinancialReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ? AND year = ? AND is_rkap = ?", companyID, year, true).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *financialReportRepository) GetRealisasiByCompanyIDAndPeriod(companyID, period string) (*domain.FinancialReportModel, error) {
	var report domain.FinancialReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ? AND period = ? AND is_rkap = ?", companyID, period, false).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *financialReportRepository) GetRealisasiByCompanyIDAndYear(companyID, year string) ([]domain.FinancialReportModel, error) {
	var reports []domain.FinancialReportModel
	err := r.db.Preload("Company").Preload("Inputter").
		Where("company_id = ? AND year = ? AND is_rkap = ?", companyID, year, false).
		Order("period ASC").
		Find(&reports).Error
	return reports, err
}

// GetRealisasiYTD menghitung akumulasi realisasi dari Januari sampai bulan yang dipilih
func (r *financialReportRepository) GetRealisasiYTD(companyID, year, month string) (*domain.FinancialReportModel, error) {
	var reports []domain.FinancialReportModel

	// Ambil semua realisasi dari Januari sampai bulan yang dipilih
	// Format period: "2024-01", "2024-02", dst
	startPeriod := year + "-01"
	endPeriod := year + "-" + month

	err := r.db.Where("company_id = ? AND year = ? AND is_rkap = ? AND period >= ? AND period <= ?",
		companyID, year, false, startPeriod, endPeriod).
		Order("period ASC").
		Find(&reports).Error

	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Akumulasi semua nilai
	ytd := domain.FinancialReportModel{
		CompanyID: companyID,
		Year:      year,
		Period:    endPeriod,
		IsRKAP:    false,
	}

	for _, report := range reports {
		ytd.CurrentAssets += report.CurrentAssets
		ytd.NonCurrentAssets += report.NonCurrentAssets
		ytd.ShortTermLiabilities += report.ShortTermLiabilities
		ytd.LongTermLiabilities += report.LongTermLiabilities
		ytd.Equity += report.Equity
		ytd.Revenue += report.Revenue
		ytd.OperatingExpenses += report.OperatingExpenses
		ytd.OperatingProfit += report.OperatingProfit
		ytd.OtherIncome += report.OtherIncome
		ytd.Tax += report.Tax
		ytd.NetProfit += report.NetProfit
		ytd.OperatingCashflow += report.OperatingCashflow
		ytd.InvestingCashflow += report.InvestingCashflow
		ytd.FinancingCashflow += report.FinancingCashflow
		ytd.EndingBalance += report.EndingBalance
		ytd.EBITDA += report.EBITDA
	}

	// Hitung rata-rata untuk rasio (atau bisa dihitung ulang dari total)
	// Untuk rasio, biasanya dihitung dari total, bukan rata-rata
	if len(reports) > 0 {
		// Hitung rasio dari total YTD
		totalAssets := ytd.CurrentAssets + ytd.NonCurrentAssets
		totalLiabilities := ytd.ShortTermLiabilities + ytd.LongTermLiabilities

		// ROE = Net Profit / Equity * 100
		if ytd.Equity > 0 {
			ytd.ROE = float64(ytd.NetProfit) / float64(ytd.Equity) * 100
		}

		// ROI = Net Profit / Total Assets * 100
		if totalAssets > 0 {
			ytd.ROI = float64(ytd.NetProfit) / float64(totalAssets) * 100
		}

		// Current Ratio = Current Assets / Short Term Liabilities
		if ytd.ShortTermLiabilities > 0 {
			ytd.CurrentRatio = float64(ytd.CurrentAssets) / float64(ytd.ShortTermLiabilities)
		}

		// Cash Ratio = Current Assets / Short Term Liabilities (simplified, bisa disesuaikan)
		if ytd.ShortTermLiabilities > 0 {
			ytd.CashRatio = float64(ytd.CurrentAssets) / float64(ytd.ShortTermLiabilities)
		}

		// EBITDA Margin = EBITDA / Revenue * 100
		if ytd.Revenue > 0 {
			ytd.EBITDAMargin = float64(ytd.EBITDA) / float64(ytd.Revenue) * 100
		}

		// Net Profit Margin = Net Profit / Revenue * 100
		if ytd.Revenue > 0 {
			ytd.NetProfitMargin = float64(ytd.NetProfit) / float64(ytd.Revenue) * 100
		}

		// Operating Profit Margin = Operating Profit / Revenue * 100
		if ytd.Revenue > 0 {
			ytd.OperatingProfitMargin = float64(ytd.OperatingProfit) / float64(ytd.Revenue) * 100
		}

		// Debt to Equity = Total Liabilities / Equity
		if ytd.Equity > 0 {
			ytd.DebtToEquity = float64(totalLiabilities) / float64(ytd.Equity)
		}
	}

	return &ytd, nil
}

func (r *financialReportRepository) Update(report *domain.FinancialReportModel) error {
	return r.db.Save(report).Error
}

func (r *financialReportRepository) Delete(id string) error {
	return r.db.Delete(&domain.FinancialReportModel{}, "id = ?", id).Error
}

// DeleteAll menghapus semua financial reports (untuk reset functionality)
func (r *financialReportRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM financial_reports").Error
}

// CountRKAPByCompanyIDAndYear menghitung jumlah RKAP untuk validasi (harus hanya 1 per tahun)
func (r *financialReportRepository) CountRKAPByCompanyIDAndYear(companyID, year string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.FinancialReportModel{}).
		Where("company_id = ? AND year = ? AND is_rkap = ?", companyID, year, true).
		Count(&count).Error
	return count, err
}

// GetRKAPYearsByCompanyID mendapatkan daftar tahun yang sudah ada RKAP untuk company tertentu
func (r *financialReportRepository) GetRKAPYearsByCompanyID(companyID string) ([]string, error) {
	var years []string
	err := r.db.Model(&domain.FinancialReportModel{}).
		Select("DISTINCT year").
		Where("company_id = ? AND is_rkap = ?", companyID, true).
		Order("year DESC").
		Pluck("year", &years).Error
	return years, err
}
