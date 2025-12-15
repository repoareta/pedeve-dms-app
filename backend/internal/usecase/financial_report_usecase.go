package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FinancialReportUseCase interface untuk financial report operations
type FinancialReportUseCase interface {
	CreateFinancialReport(data *domain.CreateFinancialReportRequest, userID, username, ipAddress, userAgent string) (*domain.FinancialReportModel, error)
	UpdateFinancialReport(id string, data *domain.UpdateFinancialReportRequest, userID, username, ipAddress, userAgent string) (*domain.FinancialReportModel, error)
	GetFinancialReportByID(id string) (*domain.FinancialReportModel, error)
	GetFinancialReportsByCompanyID(companyID string) ([]domain.FinancialReportModel, error)
	GetRKAPByCompanyIDAndYear(companyID, year string) (*domain.FinancialReportModel, error)
	GetRealisasiByCompanyIDAndPeriod(companyID, period string) (*domain.FinancialReportModel, error)
	GetComparison(companyID, year, month string) (*domain.FinancialReportComparisonResponse, error)
	GetRKAPYearsByCompanyID(companyID string) ([]string, error)
	DeleteFinancialReport(id string, userID, username, ipAddress, userAgent string) error
	ExportPerformanceExcel(companyID, startPeriod, endPeriod string) ([]byte, error)
}

type financialReportUseCase struct {
	repo repository.FinancialReportRepository
}

// NewFinancialReportUseCaseWithDB creates a new financial report use case with injected DB
func NewFinancialReportUseCaseWithDB(db *gorm.DB) FinancialReportUseCase {
	return &financialReportUseCase{
		repo: repository.NewFinancialReportRepositoryWithDB(db),
	}
}

// NewFinancialReportUseCase creates a new financial report use case with default DB
func NewFinancialReportUseCase() FinancialReportUseCase {
	return NewFinancialReportUseCaseWithDB(database.GetDB())
}

func (uc *financialReportUseCase) CreateFinancialReport(data *domain.CreateFinancialReportRequest, userID, username, ipAddress, userAgent string) (*domain.FinancialReportModel, error) {
	zapLog := logger.GetLogger()

	// Validasi: Ratio fields tidak boleh melebihi 100 (untuk persentase)
	if data.ROE > 100 || data.ROI > 100 || data.CurrentRatio > 100 || data.CashRatio > 100 ||
		data.EBITDAMargin > 100 || data.NetProfitMargin > 100 || data.OperatingProfitMargin > 100 {
		return nil, errors.New("nilai rasio keuangan tidak boleh melebihi 100%")
	}

	// Validasi: RKAP hanya boleh 1x per tahun per perusahaan
	if data.IsRKAP {
		count, err := uc.repo.CountRKAPByCompanyIDAndYear(data.CompanyID, data.Year)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing RKAP: %w", err)
		}
		if count > 0 {
			return nil, errors.New("RKAP untuk tahun ini sudah ada. Hanya boleh ada satu RKAP per tahun per perusahaan")
		}

		// Pastikan period untuk RKAP adalah tahun saja (format: "2024")
		if data.Period != data.Year {
			data.Period = data.Year
		}
	} else {
		// Validasi: Realisasi harus format YYYY-MM
		if len(data.Period) != 7 || data.Period[4] != '-' {
			return nil, errors.New("period untuk realisasi harus format YYYY-MM (contoh: 2024-01)")
		}

		// Pastikan year sesuai dengan period
		if data.Period[:4] != data.Year {
			return nil, errors.New("year harus sesuai dengan period (4 digit pertama dari period)")
		}

		// Cek apakah realisasi untuk period ini sudah ada
		existing, _ := uc.repo.GetRealisasiByCompanyIDAndPeriod(data.CompanyID, data.Period)
		if existing != nil {
			return nil, fmt.Errorf("realisasi untuk periode %s sudah ada", data.Period)
		}
	}

	report := &domain.FinancialReportModel{
		ID:                    uuid.GenerateUUID(),
		CompanyID:             data.CompanyID,
		Year:                  data.Year,
		Period:                data.Period,
		IsRKAP:                data.IsRKAP,
		InputterID:            nil,
		CurrentAssets:         data.CurrentAssets,
		NonCurrentAssets:      data.NonCurrentAssets,
		ShortTermLiabilities:  data.ShortTermLiabilities,
		LongTermLiabilities:   data.LongTermLiabilities,
		Equity:                data.Equity,
		Revenue:               data.Revenue,
		OperatingExpenses:     data.OperatingExpenses,
		OperatingProfit:       data.OperatingProfit,
		OtherIncome:           data.OtherIncome,
		Tax:                   data.Tax,
		NetProfit:             data.NetProfit,
		OperatingCashflow:     data.OperatingCashflow,
		InvestingCashflow:     data.InvestingCashflow,
		FinancingCashflow:     data.FinancingCashflow,
		EndingBalance:         data.EndingBalance,
		ROE:                   data.ROE,
		ROI:                   data.ROI,
		CurrentRatio:          data.CurrentRatio,
		CashRatio:             data.CashRatio,
		EBITDA:                data.EBITDA,
		EBITDAMargin:          data.EBITDAMargin,
		NetProfitMargin:       data.NetProfitMargin,
		OperatingProfitMargin: data.OperatingProfitMargin,
		DebtToEquity:          data.DebtToEquity,
		Remark:                data.Remark,
	}

	if userID != "" {
		report.InputterID = &userID
	}

	if err := uc.repo.Create(report); err != nil {
		zapLog.Error("Failed to create financial report", zap.Error(err))
		return nil, fmt.Errorf("failed to create financial report: %w", err)
	}

	// Audit trail - simpan semua field values untuk create
	reportType := "RKAP"
	if !data.IsRKAP {
		reportType = "Realisasi"
	}

	// Simpan semua field values sebagai changes untuk create
	changes := make(map[string]interface{})
	changes["current_assets"] = map[string]interface{}{"new": report.CurrentAssets}
	changes["non_current_assets"] = map[string]interface{}{"new": report.NonCurrentAssets}
	changes["short_term_liabilities"] = map[string]interface{}{"new": report.ShortTermLiabilities}
	changes["long_term_liabilities"] = map[string]interface{}{"new": report.LongTermLiabilities}
	changes["equity"] = map[string]interface{}{"new": report.Equity}
	changes["revenue"] = map[string]interface{}{"new": report.Revenue}
	changes["operating_expenses"] = map[string]interface{}{"new": report.OperatingExpenses}
	changes["operating_profit"] = map[string]interface{}{"new": report.OperatingProfit}
	changes["other_income"] = map[string]interface{}{"new": report.OtherIncome}
	changes["tax"] = map[string]interface{}{"new": report.Tax}
	changes["net_profit"] = map[string]interface{}{"new": report.NetProfit}
	changes["operating_cashflow"] = map[string]interface{}{"new": report.OperatingCashflow}
	changes["investing_cashflow"] = map[string]interface{}{"new": report.InvestingCashflow}
	changes["financing_cashflow"] = map[string]interface{}{"new": report.FinancingCashflow}
	changes["ending_balance"] = map[string]interface{}{"new": report.EndingBalance}
	changes["roe"] = map[string]interface{}{"new": report.ROE}
	changes["roi"] = map[string]interface{}{"new": report.ROI}
	changes["current_ratio"] = map[string]interface{}{"new": report.CurrentRatio}
	changes["cash_ratio"] = map[string]interface{}{"new": report.CashRatio}
	changes["ebitda"] = map[string]interface{}{"new": report.EBITDA}
	changes["ebitda_margin"] = map[string]interface{}{"new": report.EBITDAMargin}
	changes["net_profit_margin"] = map[string]interface{}{"new": report.NetProfitMargin}
	changes["operating_profit_margin"] = map[string]interface{}{"new": report.OperatingProfitMargin}
	changes["debt_to_equity"] = map[string]interface{}{"new": report.DebtToEquity}
	if report.Remark != nil && *report.Remark != "" {
		changes["remark"] = map[string]interface{}{"new": *report.Remark}
	}

	audit.LogAction(userID, username, audit.ActionCreate, audit.ResourceFinancialReport, report.ID, ipAddress, userAgent, "success", map[string]interface{}{
		"company_id": data.CompanyID,
		"year":       data.Year,
		"period":     data.Period,
		"type":       reportType,
		"changes":    changes, // Simpan semua field values untuk create
	})

	return report, nil
}

func (uc *financialReportUseCase) UpdateFinancialReport(id string, data *domain.UpdateFinancialReportRequest, userID, username, ipAddress, userAgent string) (*domain.FinancialReportModel, error) {
	zapLog := logger.GetLogger()

	report, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("financial report not found: %w", err)
	}

	// Simpan data lama untuk audit trail - simpan SEMUA field
	oldData := map[string]interface{}{
		"year":    report.Year,
		"period":  report.Period,
		"is_rkap": report.IsRKAP,
		// Neraca
		"current_assets":         report.CurrentAssets,
		"non_current_assets":     report.NonCurrentAssets,
		"short_term_liabilities": report.ShortTermLiabilities,
		"long_term_liabilities":  report.LongTermLiabilities,
		"equity":                 report.Equity,
		// Laba Rugi
		"revenue":            report.Revenue,
		"operating_expenses": report.OperatingExpenses,
		"operating_profit":   report.OperatingProfit,
		"other_income":       report.OtherIncome,
		"tax":                report.Tax,
		"net_profit":         report.NetProfit,
		// Cashflow
		"operating_cashflow": report.OperatingCashflow,
		"investing_cashflow": report.InvestingCashflow,
		"financing_cashflow": report.FinancingCashflow,
		"ending_balance":     report.EndingBalance,
		// Rasio
		"roe":                     report.ROE,
		"roi":                     report.ROI,
		"current_ratio":           report.CurrentRatio,
		"cash_ratio":              report.CashRatio,
		"ebitda":                  report.EBITDA,
		"ebitda_margin":           report.EBITDAMargin,
		"net_profit_margin":       report.NetProfitMargin,
		"operating_profit_margin": report.OperatingProfitMargin,
		"debt_to_equity":          report.DebtToEquity,
		"remark":                  report.Remark,
	}

	// Validasi: Ratio fields tidak boleh melebihi 100 (untuk persentase)
	if (data.ROE != nil && *data.ROE > 100) ||
		(data.ROI != nil && *data.ROI > 100) ||
		(data.CurrentRatio != nil && *data.CurrentRatio > 100) ||
		(data.CashRatio != nil && *data.CashRatio > 100) ||
		(data.EBITDAMargin != nil && *data.EBITDAMargin > 100) ||
		(data.NetProfitMargin != nil && *data.NetProfitMargin > 100) ||
		(data.OperatingProfitMargin != nil && *data.OperatingProfitMargin > 100) {
		return nil, errors.New("nilai rasio keuangan tidak boleh melebihi 100%")
	}

	// Update fields
	if data.Year != nil {
		report.Year = *data.Year
	}
	if data.Period != nil {
		report.Period = *data.Period
	}
	if data.IsRKAP != nil {
		report.IsRKAP = *data.IsRKAP
	}

	// Validasi: Jika update menjadi RKAP, cek apakah sudah ada RKAP untuk tahun tersebut
	if data.IsRKAP != nil && *data.IsRKAP {
		year := report.Year
		if data.Year != nil {
			year = *data.Year
		}
		// Jika sudah ada RKAP lain (bukan yang sedang di-update), tolak
		existingRKAP, err := uc.repo.GetRKAPByCompanyIDAndYear(report.CompanyID, year)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing RKAP: %w", err)
		}
		if existingRKAP != nil && existingRKAP.ID != id {
			return nil, errors.New("RKAP untuk tahun ini sudah ada. Hanya boleh ada satu RKAP per tahun per perusahaan")
		}
	}

	// Update Neraca
	if data.CurrentAssets != nil {
		report.CurrentAssets = *data.CurrentAssets
	}
	if data.NonCurrentAssets != nil {
		report.NonCurrentAssets = *data.NonCurrentAssets
	}
	if data.ShortTermLiabilities != nil {
		report.ShortTermLiabilities = *data.ShortTermLiabilities
	}
	if data.LongTermLiabilities != nil {
		report.LongTermLiabilities = *data.LongTermLiabilities
	}
	if data.Equity != nil {
		report.Equity = *data.Equity
	}

	// Update Laba Rugi
	if data.Revenue != nil {
		report.Revenue = *data.Revenue
	}
	if data.OperatingExpenses != nil {
		report.OperatingExpenses = *data.OperatingExpenses
	}
	if data.OperatingProfit != nil {
		report.OperatingProfit = *data.OperatingProfit
	}
	if data.OtherIncome != nil {
		report.OtherIncome = *data.OtherIncome
	}
	if data.Tax != nil {
		report.Tax = *data.Tax
	}
	if data.NetProfit != nil {
		report.NetProfit = *data.NetProfit
	}

	// Update Cashflow
	if data.OperatingCashflow != nil {
		report.OperatingCashflow = *data.OperatingCashflow
	}
	if data.InvestingCashflow != nil {
		report.InvestingCashflow = *data.InvestingCashflow
	}
	if data.FinancingCashflow != nil {
		report.FinancingCashflow = *data.FinancingCashflow
	}
	if data.EndingBalance != nil {
		report.EndingBalance = *data.EndingBalance
	}

	// Update Rasio
	if data.ROE != nil {
		report.ROE = *data.ROE
	}
	if data.ROI != nil {
		report.ROI = *data.ROI
	}
	if data.CurrentRatio != nil {
		report.CurrentRatio = *data.CurrentRatio
	}
	if data.CashRatio != nil {
		report.CashRatio = *data.CashRatio
	}
	if data.EBITDA != nil {
		report.EBITDA = *data.EBITDA
	}
	if data.EBITDAMargin != nil {
		report.EBITDAMargin = *data.EBITDAMargin
	}
	if data.NetProfitMargin != nil {
		report.NetProfitMargin = *data.NetProfitMargin
	}
	if data.OperatingProfitMargin != nil {
		report.OperatingProfitMargin = *data.OperatingProfitMargin
	}
	if data.DebtToEquity != nil {
		report.DebtToEquity = *data.DebtToEquity
	}

	if data.Remark != nil {
		report.Remark = data.Remark
	}

	if err := uc.repo.Update(report); err != nil {
		zapLog.Error("Failed to update financial report", zap.Error(err))
		return nil, fmt.Errorf("failed to update financial report: %w", err)
	}

	// Audit trail dengan perubahan
	reportType := "RKAP"
	if !report.IsRKAP {
		reportType = "Realisasi"
	}
	changes := make(map[string]interface{})
	for key, oldValue := range oldData {
		if newValue := getFieldValue(report, key); newValue != oldValue {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": newValue,
			}
		}
	}

	audit.LogAction(userID, username, audit.ActionUpdate, audit.ResourceFinancialReport, report.ID, ipAddress, userAgent, "success", map[string]interface{}{
		"company_id": report.CompanyID,
		"year":       report.Year,
		"period":     report.Period,
		"type":       reportType,
		"changes":    changes,
	})

	return report, nil
}

func (uc *financialReportUseCase) GetFinancialReportByID(id string) (*domain.FinancialReportModel, error) {
	return uc.repo.GetByID(id)
}

func (uc *financialReportUseCase) GetFinancialReportsByCompanyID(companyID string) ([]domain.FinancialReportModel, error) {
	return uc.repo.GetByCompanyID(companyID)
}

func (uc *financialReportUseCase) GetRKAPByCompanyIDAndYear(companyID, year string) (*domain.FinancialReportModel, error) {
	return uc.repo.GetRKAPByCompanyIDAndYear(companyID, year)
}

func (uc *financialReportUseCase) GetRealisasiByCompanyIDAndPeriod(companyID, period string) (*domain.FinancialReportModel, error) {
	return uc.repo.GetRealisasiByCompanyIDAndPeriod(companyID, period)
}

func (uc *financialReportUseCase) GetRKAPYearsByCompanyID(companyID string) ([]string, error) {
	return uc.repo.GetRKAPYearsByCompanyID(companyID)
}

func (uc *financialReportUseCase) GetComparison(companyID, year, month string) (*domain.FinancialReportComparisonResponse, error) {
	// Ambil RKAP untuk tahun tersebut
	rkap, err := uc.repo.GetRKAPByCompanyIDAndYear(companyID, year)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get RKAP: %w", err)
	}

	// Ambil Realisasi YTD sampai bulan yang dipilih
	realisasiYTD, err := uc.repo.GetRealisasiYTD(companyID, year, month)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get realisasi YTD: %w", err)
	}

	response := &domain.FinancialReportComparisonResponse{
		CompanyID:    companyID,
		Year:         year,
		Month:        month,
		RKAP:         rkap,
		RealisasiYTD: realisasiYTD,
		Comparison:   make(map[string]domain.ComparisonItem),
	}

	// Buat perbandingan untuk setiap field
	if rkap != nil && realisasiYTD != nil {
		// Neraca
		response.Comparison["current_assets"] = createComparisonItem(rkap.CurrentAssets, realisasiYTD.CurrentAssets)
		response.Comparison["non_current_assets"] = createComparisonItem(rkap.NonCurrentAssets, realisasiYTD.NonCurrentAssets)
		response.Comparison["short_term_liabilities"] = createComparisonItem(rkap.ShortTermLiabilities, realisasiYTD.ShortTermLiabilities)
		response.Comparison["long_term_liabilities"] = createComparisonItem(rkap.LongTermLiabilities, realisasiYTD.LongTermLiabilities)
		response.Comparison["equity"] = createComparisonItem(rkap.Equity, realisasiYTD.Equity)

		// Laba Rugi
		response.Comparison["revenue"] = createComparisonItem(rkap.Revenue, realisasiYTD.Revenue)
		response.Comparison["operating_expenses"] = createComparisonItem(rkap.OperatingExpenses, realisasiYTD.OperatingExpenses)
		response.Comparison["operating_profit"] = createComparisonItem(rkap.OperatingProfit, realisasiYTD.OperatingProfit)
		response.Comparison["other_income"] = createComparisonItem(rkap.OtherIncome, realisasiYTD.OtherIncome)
		response.Comparison["tax"] = createComparisonItem(rkap.Tax, realisasiYTD.Tax)
		response.Comparison["net_profit"] = createComparisonItem(rkap.NetProfit, realisasiYTD.NetProfit)

		// Cashflow
		response.Comparison["operating_cashflow"] = createComparisonItem(rkap.OperatingCashflow, realisasiYTD.OperatingCashflow)
		response.Comparison["investing_cashflow"] = createComparisonItem(rkap.InvestingCashflow, realisasiYTD.InvestingCashflow)
		response.Comparison["financing_cashflow"] = createComparisonItem(rkap.FinancingCashflow, realisasiYTD.FinancingCashflow)
		response.Comparison["ending_balance"] = createComparisonItem(rkap.EndingBalance, realisasiYTD.EndingBalance)

		// Rasio
		response.Comparison["roe"] = createComparisonItem(rkap.ROE, realisasiYTD.ROE)
		response.Comparison["roi"] = createComparisonItem(rkap.ROI, realisasiYTD.ROI)
		response.Comparison["current_ratio"] = createComparisonItem(rkap.CurrentRatio, realisasiYTD.CurrentRatio)
		response.Comparison["cash_ratio"] = createComparisonItem(rkap.CashRatio, realisasiYTD.CashRatio)
		response.Comparison["ebitda"] = createComparisonItem(rkap.EBITDA, realisasiYTD.EBITDA)
		response.Comparison["ebitda_margin"] = createComparisonItem(rkap.EBITDAMargin, realisasiYTD.EBITDAMargin)
		response.Comparison["net_profit_margin"] = createComparisonItem(rkap.NetProfitMargin, realisasiYTD.NetProfitMargin)
		response.Comparison["operating_profit_margin"] = createComparisonItem(rkap.OperatingProfitMargin, realisasiYTD.OperatingProfitMargin)
		response.Comparison["debt_to_equity"] = createComparisonItem(rkap.DebtToEquity, realisasiYTD.DebtToEquity)
	}

	return response, nil
}

func (uc *financialReportUseCase) DeleteFinancialReport(id string, userID, username, ipAddress, userAgent string) error {
	zapLog := logger.GetLogger()

	report, err := uc.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("financial report not found: %w", err)
	}

	if err := uc.repo.Delete(id); err != nil {
		zapLog.Error("Failed to delete financial report", zap.Error(err))
		return fmt.Errorf("failed to delete financial report: %w", err)
	}

	// Audit trail
	reportType := "RKAP"
	if !report.IsRKAP {
		reportType = "Realisasi"
	}
	audit.LogAction(userID, username, audit.ActionDelete, audit.ResourceFinancialReport, id, ipAddress, userAgent, "success", map[string]interface{}{
		"company_id": report.CompanyID,
		"year":       report.Year,
		"period":     report.Period,
		"type":       reportType,
	})

	return nil
}

// Helper functions
func createComparisonItem(rkap, realisasi interface{}) domain.ComparisonItem {
	var rkapVal, realisasiVal float64

	switch v := rkap.(type) {
	case int64:
		rkapVal = float64(v)
	case float64:
		rkapVal = v
	default:
		rkapVal = 0
	}

	switch v := realisasi.(type) {
	case int64:
		realisasiVal = float64(v)
	case float64:
		realisasiVal = v
	default:
		realisasiVal = 0
	}

	difference := realisasiVal - rkapVal
	percentage := 0.0
	if rkapVal != 0 {
		percentage = (realisasiVal / rkapVal) * 100
	}

	return domain.ComparisonItem{
		RKAP:         rkap,
		RealisasiYTD: realisasi,
		Difference:   difference,
		Percentage:   percentage,
	}
}

func getFieldValue(report *domain.FinancialReportModel, field string) interface{} {
	switch field {
	case "year":
		return report.Year
	case "period":
		return report.Period
	case "is_rkap":
		return report.IsRKAP
	// Neraca
	case "current_assets":
		return report.CurrentAssets
	case "non_current_assets":
		return report.NonCurrentAssets
	case "short_term_liabilities":
		return report.ShortTermLiabilities
	case "long_term_liabilities":
		return report.LongTermLiabilities
	case "equity":
		return report.Equity
	// Laba Rugi
	case "revenue":
		return report.Revenue
	case "operating_expenses":
		return report.OperatingExpenses
	case "operating_profit":
		return report.OperatingProfit
	case "other_income":
		return report.OtherIncome
	case "tax":
		return report.Tax
	case "net_profit":
		return report.NetProfit
	// Cashflow
	case "operating_cashflow":
		return report.OperatingCashflow
	case "investing_cashflow":
		return report.InvestingCashflow
	case "financing_cashflow":
		return report.FinancingCashflow
	case "ending_balance":
		return report.EndingBalance
	// Rasio
	case "roe":
		return report.ROE
	case "roi":
		return report.ROI
	case "current_ratio":
		return report.CurrentRatio
	case "cash_ratio":
		return report.CashRatio
	case "ebitda":
		return report.EBITDA
	case "ebitda_margin":
		return report.EBITDAMargin
	case "net_profit_margin":
		return report.NetProfitMargin
	case "operating_profit_margin":
		return report.OperatingProfitMargin
	case "debt_to_equity":
		return report.DebtToEquity
	case "remark":
		return report.Remark
	default:
		return nil
	}
}

// ExportPerformanceExcel generates Excel file with 4 sheets (Balance Sheet, Profit & Loss, Cashflow, Ratio)
// Each sheet contains chart and table data with RKAP vs Realisasi comparison
func (uc *financialReportUseCase) ExportPerformanceExcel(companyID, startPeriod, endPeriod string) ([]byte, error) {
	// #region agent log
	logEntryExport := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "ALL",
		"location":     "financial_report_usecase.go:594",
		"message":      "ExportPerformanceExcel called",
		"data": map[string]interface{}{
			"companyID":   companyID,
			"startPeriod": startPeriod,
			"endPeriod":   endPeriod,
		},
		"timestamp": time.Now().UnixMilli(),
	}
	if logData, err := json.Marshal(logEntryExport); err == nil {
		if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
	}
	// #endregion

	// Validate period format (YYYY-MM)
	if len(startPeriod) != 7 || len(endPeriod) != 7 {
		return nil, fmt.Errorf("invalid period format, expected YYYY-MM")
	}

	// Get all financial reports for company
	reports, err := uc.repo.GetByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial reports: %w", err)
	}

	// Extract year and months from periods
	startYear := startPeriod[:4]
	startMonth := startPeriod[5:7]

	// Filter reports by period range
	var filteredReports []domain.FinancialReportModel
	var rkapReport *domain.FinancialReportModel

	for _, report := range reports {
		// Get RKAP for the year
		if report.IsRKAP && report.Year == startYear {
			rkapReport = &report
		}

		// Get realisasi reports in the range
		if !report.IsRKAP {
			reportYear := report.Period[:4]
			reportMonth := report.Period[5:7]
			if reportYear == startYear && reportMonth >= startMonth && reportMonth <= endPeriod[5:7] {
				filteredReports = append(filteredReports, report)
			}
		}
	}

	// Sort by period
	sort.Slice(filteredReports, func(i, j int) bool {
		return filteredReports[i].Period < filteredReports[j].Period
	})

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Delete default Sheet1 (ignore error if Sheet1 doesn't exist)
	_ = f.DeleteSheet("Sheet1")

	// Generate sheets (always create sheets, even if empty)
	// #region agent log
	logEntryGen := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "ALL",
		"location":     "financial_report_usecase.go:665",
		"message":      "About to generate sheets",
		"data": map[string]interface{}{
			"numFilteredReports": len(filteredReports),
			"hasRKAP":            rkapReport != nil,
		},
		"timestamp": time.Now().UnixMilli(),
	}
	if logData, err := json.Marshal(logEntryGen); err == nil {
		if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
	}
	// #endregion

	if err := uc.generateBalanceSheetSheet(f, filteredReports, rkapReport, startPeriod, endPeriod); err != nil {
		return nil, fmt.Errorf("failed to generate balance sheet sheet: %w", err)
	}

	if err := uc.generateProfitLossSheet(f, filteredReports, rkapReport, startPeriod, endPeriod); err != nil {
		return nil, fmt.Errorf("failed to generate profit loss sheet: %w", err)
	}

	if err := uc.generateCashflowSheet(f, filteredReports, rkapReport, startPeriod, endPeriod); err != nil {
		return nil, fmt.Errorf("failed to generate cashflow sheet: %w", err)
	}

	if err := uc.generateRatioSheet(f, filteredReports, rkapReport, startPeriod, endPeriod); err != nil {
		return nil, fmt.Errorf("failed to generate ratio sheet: %w", err)
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper function to get column letter from number (1=A, 2=B, ..., 27=AA, etc.)
func (uc *financialReportUseCase) getColumnLetter(colNum int) string {
	result := ""
	for colNum > 0 {
		colNum--
		result = string(rune('A'+colNum%26)) + result
		colNum /= 26
	}
	return result
}

// Helper function to format Excel reference with proper quoting for sheet names containing spaces
// Excel requires single quotes around sheet names that contain spaces or special characters
// If sheet name contains single quotes, they must be escaped by doubling them
func (uc *financialReportUseCase) formatExcelRef(sheetName, rangeRef string) string {
	// Check if sheet name contains spaces or special characters that require quoting
	needsQuotes := false
	for _, r := range sheetName {
		if r == ' ' || r == '\'' || r == '[' || r == ']' || r == '(' || r == ')' {
			needsQuotes = true
			break
		}
	}
	if needsQuotes {
		// Escape single quotes by doubling them
		escapedSheetName := strings.ReplaceAll(sheetName, "'", "''")
		return fmt.Sprintf("'%s'!%s", escapedSheetName, rangeRef)
	}
	return fmt.Sprintf("%s!%s", sheetName, rangeRef)
}

// Helper function to format currency value (IDR)
func formatCurrencyValue(value int64) string {
	if value == 0 {
		return "Rp 0"
	}

	absValue := value
	if absValue < 0 {
		absValue = -absValue
	}

	sign := ""
	if value < 0 {
		sign = "-"
	}

	// Format: Rp X.XXB for billions, Rp X.XXJt for millions
	if absValue >= 1000000000 {
		return fmt.Sprintf("%sRp %.2fB", sign, float64(absValue)/1000000000.0)
	} else if absValue >= 1000000 {
		return fmt.Sprintf("%sRp %.2fJt", sign, float64(absValue)/1000000.0)
	} else if absValue >= 1000 {
		return fmt.Sprintf("%sRp %.2fRb", sign, float64(absValue)/1000.0)
	}
	return fmt.Sprintf("%sRp %s", sign, formatNumber(int64(absValue)))
}

// Helper function to format number with thousand separators
func formatNumber(value int64) string {
	// Simple formatting without library
	str := fmt.Sprintf("%d", value)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

// Helper function to format ratio value
func formatRatioValue(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// Table item structure
type tableItem struct {
	label    string
	field    string
	isRatio  bool
	getRkap  func(*domain.FinancialReportModel) int64
	getReal  func(domain.FinancialReportModel) int64
	getRkapF func(*domain.FinancialReportModel) float64
	getRealF func(domain.FinancialReportModel) float64
}

// Helper function to write table with multi-level headers (like frontend)
func (uc *financialReportUseCase) writeFinancialTable(
	f *excelize.File,
	sheetName string,
	items []tableItem,
	reports []domain.FinancialReportModel,
	rkap *domain.FinancialReportModel,
	startRow int,
) (dataStartRow int, chartDataStartRow int) {
	// Row for main header (category names with merged cells for RKAP/Realisasi)
	headerRow1 := startRow
	// Row for sub-header (RKAP, Realisasi)
	headerRow2 := startRow + 1
	// Row for data start (monthRow is not needed, months are in column A of each data row)
	dataStartRow = startRow + 3
	chartDataStartRow = dataStartRow

	monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	// Header style for category
	categoryHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E3F2FD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	// Header style for RKAP/Realisasi (light grey background, bold, center)
	subHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F5F5F5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	// Data cell style (left aligned for currency values as per app display)
	dataCellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	// Write "Bulan" in first column, spanning both header rows (merge headerRow1 and headerRow2)
	// IMPORTANT: Set value first, then merge, then apply style
	bulanCellA1 := fmt.Sprintf("A%d", headerRow1)
	bulanCellA2 := fmt.Sprintf("A%d", headerRow2)
	_ = f.SetCellValue(sheetName, bulanCellA1, "Bulan")
	// Set value in second cell before merging (required for Excelize)
	_ = f.SetCellValue(sheetName, bulanCellA2, "")
	_ = f.MergeCell(sheetName, bulanCellA1, bulanCellA2)
	// Style for merged "Bulan" header (light grey background, bold, center)
	bulanHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F5F5F5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	_ = f.SetCellStyle(sheetName, bulanCellA1, bulanCellA2, bulanHeaderStyle)

	// Write category headers and sub-headers
	col := 2
	for _, item := range items {
		// Category header (merge 2 columns)
		// IMPORTANT: Set values first, then merge, then apply style
		startCell, _ := excelize.CoordinatesToCellName(col, headerRow1)
		endCell, _ := excelize.CoordinatesToCellName(col+1, headerRow1)
		_ = f.SetCellValue(sheetName, startCell, item.label)
		// Set value in end cell before merging (required for Excelize merge to work correctly)
		_ = f.SetCellValue(sheetName, endCell, "")
		_ = f.MergeCell(sheetName, startCell, endCell)
		_ = f.SetCellStyle(sheetName, startCell, endCell, categoryHeaderStyle)

		// Sub-headers: RKAP and Realisasi
		rkapCell, _ := excelize.CoordinatesToCellName(col, headerRow2)
		realCell, _ := excelize.CoordinatesToCellName(col+1, headerRow2)
		_ = f.SetCellValue(sheetName, rkapCell, "RKAP")
		_ = f.SetCellValue(sheetName, realCell, "Realisasi")
		_ = f.SetCellStyle(sheetName, rkapCell, rkapCell, subHeaderStyle)
		_ = f.SetCellStyle(sheetName, realCell, realCell, subHeaderStyle)

		col += 2
	}

	// Month row (row 5) - not needed, month names will be in column A for each data row
	// This row can be left empty or used for additional styling

	numMonths := len(reports)
	if numMonths == 0 && rkap != nil {
		numMonths = 3 // Minimum for chart
	}

	// Write data rows - one row per month
	for monthIdx := 0; monthIdx < numMonths; monthIdx++ {
		dataRow := dataStartRow + monthIdx
		var monthName string
		var report *domain.FinancialReportModel
		if monthIdx < len(reports) {
			report = &reports[monthIdx]
			if len(report.Period) >= 7 {
				monthNum := report.Period[5:7]
				var monthIndex int
				_, _ = fmt.Sscanf(monthNum, "%d", &monthIndex)
				if monthIndex > 0 && monthIndex <= 12 {
					monthName = monthNames[monthIndex-1]
				} else {
					monthName = monthNum
				}
			}
		} else {
			monthName = fmt.Sprintf("Month %d", monthIdx+1)
		}

		// Month name in first column (left aligned, bold)
		monthCellStyle, _ := f.NewStyle(&excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 10},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "#000000", Style: 1},
				{Type: "top", Color: "#000000", Style: 1},
				{Type: "right", Color: "#000000", Style: 1},
				{Type: "bottom", Color: "#000000", Style: 1},
			},
		})
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", dataRow), monthName)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", dataRow), fmt.Sprintf("A%d", dataRow), monthCellStyle)

		// Write data for each item - each item has RKAP and Realisasi for this month
		col = 2
		for _, item := range items {
			var rkapVal string
			var realVal string

			if item.isRatio {
				// For ratio fields
				if rkap != nil && item.getRkapF != nil {
					val := item.getRkapF(rkap)
					rkapVal = formatRatioValue(val)
				} else {
					rkapVal = "-"
				}
				if report != nil && item.getRealF != nil {
					val := item.getRealF(*report)
					realVal = formatRatioValue(val)
				} else {
					realVal = "-"
				}
			} else {
				// For currency fields
				if rkap != nil && item.getRkap != nil {
					val := item.getRkap(rkap)
					rkapVal = formatCurrencyValue(val)
				} else {
					rkapVal = "-"
				}
				if report != nil && item.getReal != nil {
					val := item.getReal(*report)
					realVal = formatCurrencyValue(val)
				} else {
					realVal = "-"
				}
			}

			// Write RKAP value (same for all months since it's annual) as string to preserve format
			rkapCell, _ := excelize.CoordinatesToCellName(col, dataRow)
			_ = f.SetCellValue(sheetName, rkapCell, rkapVal)
			_ = f.SetCellStyle(sheetName, rkapCell, rkapCell, dataCellStyle)

			// Write Realisasi value (monthly specific) as string to preserve format
			realCell, _ := excelize.CoordinatesToCellName(col+1, dataRow)
			_ = f.SetCellValue(sheetName, realCell, realVal)
			_ = f.SetCellStyle(sheetName, realCell, realCell, dataCellStyle)

			col += 2
		}

		// Set column widths (only once, on first month)
		if monthIdx == 0 {
			// Set width for month column
			_ = f.SetColWidth(sheetName, "A", "A", 15)
			// Set width for all data columns (RKAP and Realisasi for each category)
			for i := 0; i < len(items)*2; i++ {
				_ = f.SetColWidth(sheetName, uc.getColumnLetter(2+i), uc.getColumnLetter(2+i), 18)
			}
		}
	}

	return dataStartRow, chartDataStartRow
}

// generateBalanceSheetSheet generates Balance Sheet sheet with chart and table
func (uc *financialReportUseCase) generateBalanceSheetSheet(
	f *excelize.File,
	reports []domain.FinancialReportModel,
	rkap *domain.FinancialReportModel,
	startPeriod, endPeriod string,
) error {
	// #region agent log
	logEntryBS := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "B",
		"location":     "financial_report_usecase.go:943",
		"message":      "generateBalanceSheetSheet called",
		"data": map[string]interface{}{
			"numReports": len(reports),
			"hasRKAP":    rkap != nil,
		},
		"timestamp": time.Now().UnixMilli(),
	}
	if logData, err := json.Marshal(logEntryBS); err == nil {
		if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
	}
	// #endregion

	sheetName := "Balance Sheet"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Title
	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("Neraca (Balance Sheet) - Periode %s - %s", startPeriod, endPeriod))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// Define balance sheet items
	items := []tableItem{
		{
			label:   "A. Aset Lancar",
			field:   "current_assets",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.CurrentAssets
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.CurrentAssets
			},
		},
		{
			label:   "B. Aset Tidak Lancar",
			field:   "non_current_assets",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.NonCurrentAssets
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.NonCurrentAssets
			},
		},
		{
			label:   "C. Liabilitas Jangka Pendek",
			field:   "short_term_liabilities",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.ShortTermLiabilities
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.ShortTermLiabilities
			},
		},
		{
			label:   "D. Liabilitas Jangka Panjang",
			field:   "long_term_liabilities",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.LongTermLiabilities
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.LongTermLiabilities
			},
		},
		{
			label:   "E. Ekuitas",
			field:   "equity",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.Equity
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.Equity
			},
		},
	}

	// Write table starting from row 3 (after title at row 1, row 2 is empty)
	// This will write: row 3-4 = headers, row 6+ = data rows
	// IMPORTANT: Table must be written first before chart data to prevent overlap
	_, _ = uc.writeFinancialTable(f, sheetName, items, reports, rkap, 3)

	// Calculate where chart data should be placed (after table, with enough spacing)
	// Count actual number of months for calculation
	numMonths := len(reports)
	if numMonths == 0 && rkap != nil {
		numMonths = 3 // Minimum for chart
	}

	// IMPORTANT: Chart data should start from row 20 (safe position, not too far down)
	// This ensures chart data is accessible and chart can reference it correctly
	chartDataStartRow := 20

	// Chart data for Total Assets, Total Liabilities, Equity
	chartMonthRow := chartDataStartRow
	chartDataRow := chartMonthRow + 1

	// Write month labels for chart (simple, one column per month)
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartMonthRow), "Month")
	col := 2
	monthNamesShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	for i := 0; i < numMonths; i++ {
		var monthName string
		if i < len(reports) && len(reports[i].Period) >= 7 {
			monthNum := reports[i].Period[5:7]
			var monthIndex int
			_, _ = fmt.Sscanf(monthNum, "%d", &monthIndex)
			if monthIndex > 0 && monthIndex <= 12 {
				monthName = monthNamesShort[monthIndex-1]
			} else {
				monthName = monthNum
			}
		} else {
			monthName = fmt.Sprintf("M%d", i+1)
		}
		cell, _ := excelize.CoordinatesToCellName(col, chartMonthRow)
		_ = f.SetCellValue(sheetName, cell, monthName)
		col++
	}

	// Write chart data rows (using numeric values for chart, but hidden far below table)
	// Write Total Assets (RKAP) at chartMonthRow+1, Total Assets (Realisasi) at chartMonthRow+2
	// Write Total Liabilities (RKAP) at chartMonthRow+3, Total Liabilities (Realisasi) at chartMonthRow+4
	// Write Total Assets (RKAP) at chartMonthRow+1, Total Assets (Realisasi) at chartMonthRow+2
	// Write Total Liabilities (RKAP) at chartMonthRow+3, Total Liabilities (Realisasi) at chartMonthRow+4
	// Write Equity (RKAP) at chartMonthRow+5, Equity (Realisasi) at chartMonthRow+6
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Total Assets (RKAP)")
	col = 2
	totalAssetsRKAP := int64(0)
	if rkap != nil {
		totalAssetsRKAP = rkap.CurrentAssets + rkap.NonCurrentAssets
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, totalAssetsRKAP) // Keep as number for chart
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Total Assets (Realisasi)")
	col = 2
	for _, report := range reports {
		totalAssets := report.CurrentAssets + report.NonCurrentAssets
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, totalAssets) // Keep as number for chart
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Total Liabilities (RKAP)")
	col = 2
	totalLiabRKAP := int64(0)
	if rkap != nil {
		totalLiabRKAP = rkap.ShortTermLiabilities + rkap.LongTermLiabilities
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, totalLiabRKAP) // Keep as number for chart
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Total Liabilities (Realisasi)")
	col = 2
	for _, report := range reports {
		totalLiab := report.ShortTermLiabilities + report.LongTermLiabilities
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, totalLiab) // Keep as number for chart
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Equity (RKAP)")
	col = 2
	equityRKAP := int64(0)
	if rkap != nil {
		equityRKAP = rkap.Equity
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, equityRKAP) // Keep as number for chart
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Equity (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.Equity) // Keep as number for chart
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}

	// Create chart
	if numMonths > 0 {
		// #region agent log
		logEntry := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "A",
			"location":     "financial_report_usecase.go:1167",
			"message":      "Balance Sheet chart creation start",
			"data": map[string]interface{}{
				"sheetName":     sheetName,
				"numMonths":     numMonths,
				"chartMonthRow": chartMonthRow,
				"lenItems":      len(items),
				"chartPos":      fmt.Sprintf("%s%d", uc.getColumnLetter(len(items)*2+3), 3),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err := json.Marshal(logEntry); err == nil {
			if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion

		// Calculate the last column: data starts at column 2 (B), loop writes to columns 2, 3, 4, ..., 2+numMonths-1
		// So last column = 2 + numMonths - 1 = 1 + numMonths
		// For 12 months: columns 2-13 (B-M), lastCol = 13
		lastCol := 1 + numMonths
		lastColLetter := uc.getColumnLetter(lastCol)
		categoriesRange := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow, lastColLetter, chartMonthRow))

		// #region agent log
		logEntry2 := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "D",
			"location":     "financial_report_usecase.go:1261",
			"message":      "Balance Sheet chart range calculated",
			"data": map[string]interface{}{
				"endColCalc":      lastColLetter,
				"lastCol":         lastCol,
				"startCol":        2,
				"categoriesRange": categoriesRange,
				"numMonths":       numMonths,
				"chartMonthRow":   chartMonthRow,
				"expectedRange":   fmt.Sprintf("B%d:%s%d", chartMonthRow, lastColLetter, chartMonthRow),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err := json.Marshal(logEntry2); err == nil {
			if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion

		// Chart data rows: Total Assets (RKAP) at chartMonthRow+1, Total Assets (Realisasi) at chartMonthRow+2,
		// Total Liabilities (RKAP) at chartMonthRow+3, Total Liabilities (Realisasi) at chartMonthRow+4,
		// Equity (RKAP) at chartMonthRow+5, Equity (Realisasi) at chartMonthRow+6
		// Use consistent chart position (column C, row 3) for all sheets
		// Build series with correct ranges
		series1Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+1, lastColLetter, chartMonthRow+1))
		series2Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+2, lastColLetter, chartMonthRow+2))
		series3Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+3, lastColLetter, chartMonthRow+3))
		series4Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+4, lastColLetter, chartMonthRow+4))
		series5Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+5, lastColLetter, chartMonthRow+5))
		series6Values := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+6, lastColLetter, chartMonthRow+6))

		// #region agent log - chart ranges
		logEntryRanges := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "E",
			"location":     "financial_report_usecase.go:1295",
			"message":      "Balance Sheet chart series ranges",
			"data": map[string]interface{}{
				"categoriesRange": categoriesRange,
				"series1Values":   series1Values,
				"series2Values":   series2Values,
				"series3Values":   series3Values,
				"series4Values":   series4Values,
				"series5Values":   series5Values,
				"series6Values":   series6Values,
				"chartMonthRow":   chartMonthRow,
				"lastColLetter":   lastColLetter,
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err := json.Marshal(logEntryRanges); err == nil {
			if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion

		err = f.AddChart(sheetName, "C3", &excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+1)),
					Categories: categoriesRange,
					Values:     series1Values,
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+2)),
					Categories: categoriesRange,
					Values:     series2Values,
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+3)),
					Categories: categoriesRange,
					Values:     series3Values,
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+4)),
					Categories: categoriesRange,
					Values:     series4Values,
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+5)),
					Categories: categoriesRange,
					Values:     series5Values,
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+6)),
					Categories: categoriesRange,
					Values:     series6Values,
				},
			},
			Title: []excelize.RichTextRun{
				{Text: "Balance Sheet Overview"},
			},
			Legend: excelize.ChartLegend{
				Position: "right",
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName: false,
				ShowSerName: true,
				ShowVal:     false,
			},
		})
		// #region agent log
		logEntry3 := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "C",
			"location":     "financial_report_usecase.go:1221",
			"message":      "Balance Sheet AddChart result",
			"data": map[string]interface{}{
				"error": err != nil,
				"errorMsg": func() string {
					if err != nil {
						return err.Error()
					} else {
						return ""
					}
				}(),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err2 := json.Marshal(logEntry3); err2 == nil {
			if logFile, err2 := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion
		if err != nil {
			return fmt.Errorf("failed to add chart: %w", err)
		}
	}

	return nil
}

// Similar simplified implementations for other sheets
func (uc *financialReportUseCase) generateProfitLossSheet(
	f *excelize.File,
	reports []domain.FinancialReportModel,
	rkap *domain.FinancialReportModel,
	startPeriod, endPeriod string,
) error {
	// #region agent log
	logEntryPL := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "B",
		"location":     "financial_report_usecase.go:1328",
		"message":      "generateProfitLossSheet called",
		"data": map[string]interface{}{
			"numReports": len(reports),
			"hasRKAP":    rkap != nil,
		},
		"timestamp": time.Now().UnixMilli(),
	}
	if logData, err := json.Marshal(logEntryPL); err == nil {
		if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
	}
	// #endregion

	sheetName := "Profit & Loss"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("Laba Rugi (Profit & Loss) - Periode %s - %s", startPeriod, endPeriod))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// Define profit & loss items
	items := []tableItem{
		{
			label:   "A. Revenue",
			field:   "revenue",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.Revenue
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.Revenue
			},
		},
		{
			label:   "B. Beban Usaha",
			field:   "operating_expenses",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.OperatingExpenses
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.OperatingExpenses
			},
		},
		{
			label:   "C. Laba Usaha",
			field:   "operating_profit",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.OperatingProfit
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.OperatingProfit
			},
		},
		{
			label:   "D. Pendapatan Lain-Lain",
			field:   "other_income",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.OtherIncome
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.OtherIncome
			},
		},
		{
			label:   "E. Tax",
			field:   "tax",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.Tax
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.Tax
			},
		},
		{
			label:   "F. Laba Bersih",
			field:   "net_profit",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.NetProfit
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.NetProfit
			},
		},
	}

	// Write table
	_, _ = uc.writeFinancialTable(f, sheetName, items, reports, rkap, 3)

	numMonths := len(reports)
	if numMonths == 0 && rkap != nil {
		numMonths = 3
	}

	// IMPORTANT: Chart data should start from row 20 (safe position, not too far down)
	// This ensures chart data is accessible and chart can reference it correctly
	chartDataStartRow := 20

	// Chart data for Revenue and Net Profit
	chartMonthRow := chartDataStartRow
	chartDataRow := chartMonthRow + 1

	// Write month labels for chart
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartMonthRow), "Month")
	col := 2
	monthNamesShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	for i := 0; i < numMonths; i++ {
		var monthName string
		if i < len(reports) && len(reports[i].Period) >= 7 {
			monthNum := reports[i].Period[5:7]
			var monthIndex int
			_, _ = fmt.Sscanf(monthNum, "%d", &monthIndex)
			if monthIndex > 0 && monthIndex <= 12 {
				monthName = monthNamesShort[monthIndex-1]
			} else {
				monthName = monthNum
			}
		} else {
			monthName = fmt.Sprintf("M%d", i+1)
		}
		cell, _ := excelize.CoordinatesToCellName(col, chartMonthRow)
		_ = f.SetCellValue(sheetName, cell, monthName)
		col++
	}

	// Chart data rows: Revenue and Net Profit
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Revenue (RKAP)")
	col = 2
	rkapRev := int64(0)
	if rkap != nil {
		rkapRev = rkap.Revenue
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapRev)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Revenue (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.Revenue)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Net Profit (RKAP)")
	col = 2
	rkapNP := int64(0)
	if rkap != nil {
		rkapNP = rkap.NetProfit
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapNP)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Net Profit (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.NetProfit)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}

	// Create chart
	if numMonths > 0 {
		// #region agent log
		logEntryPL := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "A",
			"location":     "financial_report_usecase.go:1498",
			"message":      "Profit Loss chart creation start",
			"data": map[string]interface{}{
				"sheetName":     sheetName,
				"numMonths":     numMonths,
				"chartMonthRow": chartMonthRow,
				"lenItems":      len(items),
				"chartPos":      fmt.Sprintf("%s%d", uc.getColumnLetter(len(items)*2+3), 3),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err := json.Marshal(logEntryPL); err == nil {
			if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion

		// Calculate the last column: data starts at column 2 (B), so last column = 2 + numMonths - 1 = 1 + numMonths
		lastCol := 1 + numMonths
		lastColLetter := uc.getColumnLetter(lastCol)
		categoriesRange := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow, lastColLetter, chartMonthRow))

		// Chart data rows: Revenue (RKAP) at chartMonthRow+1, Revenue (Realisasi) at chartMonthRow+2,
		// Net Profit (RKAP) at chartMonthRow+3, Net Profit (Realisasi) at chartMonthRow+4
		// Use consistent chart position (column C, row 3) for all sheets
		err = f.AddChart(sheetName, "C3", &excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+1)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+1, lastColLetter, chartMonthRow+1)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+2)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+2, lastColLetter, chartMonthRow+2)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+3)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+3, lastColLetter, chartMonthRow+3)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+4)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+4, lastColLetter, chartMonthRow+4)),
				},
			},
			Title: []excelize.RichTextRun{
				{Text: "Profit & Loss Overview"},
			},
			Legend: excelize.ChartLegend{
				Position: "right",
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName: false,
				ShowSerName: true,
				ShowVal:     false,
			},
		})
		// #region agent log
		logEntryPL2 := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "C",
			"location":     "financial_report_usecase.go:1540",
			"message":      "Profit Loss AddChart result",
			"data": map[string]interface{}{
				"error": err != nil,
				"errorMsg": func() string {
					if err != nil {
						return err.Error()
					} else {
						return ""
					}
				}(),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err2 := json.Marshal(logEntryPL2); err2 == nil {
			if logFile, err2 := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion
		if err != nil {
			return fmt.Errorf("failed to add chart: %w", err)
		}
	}

	return nil
}

func (uc *financialReportUseCase) generateCashflowSheet(
	f *excelize.File,
	reports []domain.FinancialReportModel,
	rkap *domain.FinancialReportModel,
	startPeriod, endPeriod string,
) error {
	// #region agent log
	logEntryCF := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": "B",
		"location":     "financial_report_usecase.go:1628",
		"message":      "generateCashflowSheet called (WORKING)",
		"data": map[string]interface{}{
			"numReports": len(reports),
			"hasRKAP":    rkap != nil,
		},
		"timestamp": time.Now().UnixMilli(),
	}
	if logData, err := json.Marshal(logEntryCF); err == nil {
		if logFile, err := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
	}
	// #endregion

	sheetName := "Cashflow"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("Cashflow - Periode %s - %s", startPeriod, endPeriod))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// Define cashflow items
	items := []tableItem{
		{
			label:   "A. Arus kas bersih dari operasi",
			field:   "operating_cashflow",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.OperatingCashflow
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.OperatingCashflow
			},
		},
		{
			label:   "B. Arus kas bersih dari investasi",
			field:   "investing_cashflow",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.InvestingCashflow
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.InvestingCashflow
			},
		},
		{
			label:   "C. Arus kas bersih dari pendanaan",
			field:   "financing_cashflow",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.FinancingCashflow
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.FinancingCashflow
			},
		},
		{
			label:   "D. Saldo Akhir",
			field:   "ending_balance",
			isRatio: false,
			getRkap: func(r *domain.FinancialReportModel) int64 {
				if r == nil {
					return 0
				}
				return r.EndingBalance
			},
			getReal: func(r domain.FinancialReportModel) int64 {
				return r.EndingBalance
			},
		},
	}

	// Write table
	tableDataStartRow, _ := uc.writeFinancialTable(f, sheetName, items, reports, rkap, 3)

	numMonths := len(reports)
	if numMonths == 0 && rkap != nil {
		numMonths = 3
	}

	// IMPORTANT: Calculate exact end of table and place chart data far below
	tableEndRow := tableDataStartRow + numMonths - 1
	chartDataStartRow := tableEndRow + 50

	// Chart data - Net Cashflow and Ending Balance
	chartMonthRow := chartDataStartRow
	chartDataRow := chartMonthRow + 1

	// Write month labels for chart
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartMonthRow), "Month")
	col := 2
	monthNamesShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	for i := 0; i < numMonths; i++ {
		var monthName string
		if i < len(reports) && len(reports[i].Period) >= 7 {
			monthNum := reports[i].Period[5:7]
			var monthIndex int
			_, _ = fmt.Sscanf(monthNum, "%d", &monthIndex)
			if monthIndex > 0 && monthIndex <= 12 {
				monthName = monthNamesShort[monthIndex-1]
			} else {
				monthName = monthNum
			}
		} else {
			monthName = fmt.Sprintf("M%d", i+1)
		}
		cell, _ := excelize.CoordinatesToCellName(col, chartMonthRow)
		_ = f.SetCellValue(sheetName, cell, monthName)
		col++
	}

	// Chart data rows - Net Cashflow and Ending Balance
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Net Cashflow (RKAP)")
	col = 2
	rkapNetCF := int64(0)
	if rkap != nil {
		rkapNetCF = rkap.OperatingCashflow + rkap.InvestingCashflow + rkap.FinancingCashflow
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapNetCF)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Net Cashflow (Realisasi)")
	col = 2
	for _, report := range reports {
		realNetCF := report.OperatingCashflow + report.InvestingCashflow + report.FinancingCashflow
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, realNetCF)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Ending Balance (RKAP)")
	col = 2
	rkapEB := int64(0)
	if rkap != nil {
		rkapEB = rkap.EndingBalance
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapEB)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "Ending Balance (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.EndingBalance)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}

	// Create chart
	if numMonths > 0 {
		// Calculate the last column: data starts at column 2 (B), so last column = 2 + numMonths - 1 = 1 + numMonths
		lastCol := 1 + numMonths
		lastColLetter := uc.getColumnLetter(lastCol)
		categoriesRange := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow, lastColLetter, chartMonthRow))

		err = f.AddChart(sheetName, fmt.Sprintf("%s%d", uc.getColumnLetter(len(items)*2+3), 3), &excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+1)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+1, lastColLetter, chartMonthRow+1)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+2)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+2, lastColLetter, chartMonthRow+2)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+3)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+3, lastColLetter, chartMonthRow+3)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+4)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+4, lastColLetter, chartMonthRow+4)),
				},
			},
			Title: []excelize.RichTextRun{
				{Text: "Cashflow Overview"},
			},
			Legend: excelize.ChartLegend{
				Position: "right",
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName: false,
				ShowSerName: true,
				ShowVal:     false,
			},
		})
		// #region agent log
		logEntryCF2 := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "run1",
			"hypothesisId": "C",
			"location":     "financial_report_usecase.go:1689",
			"message":      "Cashflow AddChart result (WORKING)",
			"data": map[string]interface{}{
				"error": err != nil,
				"errorMsg": func() string {
					if err != nil {
						return err.Error()
					} else {
						return ""
					}
				}(),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if logData, err2 := json.Marshal(logEntryCF2); err2 == nil {
			if logFile, err2 := os.OpenFile("/Users/f/Documents/Projects/pedeve-dms-app/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
				_, _ = logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
		}
		// #endregion
		if err != nil {
			return fmt.Errorf("failed to add chart: %w", err)
		}
	}

	return nil
}

func (uc *financialReportUseCase) generateRatioSheet(
	f *excelize.File,
	reports []domain.FinancialReportModel,
	rkap *domain.FinancialReportModel,
	startPeriod, endPeriod string,
) error {
	sheetName := "Ratio"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	_ = f.SetCellValue(sheetName, "A1", fmt.Sprintf("Rasio Keuangan (%%) - Periode %s - %s", startPeriod, endPeriod))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	_ = f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// Define ratio items
	items := []tableItem{
		{
			label:   "ROE (Return on Equity)",
			field:   "roe",
			isRatio: true,
			getRkapF: func(r *domain.FinancialReportModel) float64 {
				if r == nil {
					return 0
				}
				return r.ROE
			},
			getRealF: func(r domain.FinancialReportModel) float64 {
				return r.ROE
			},
		},
		{
			label:   "ROI (Return on Investment)",
			field:   "roi",
			isRatio: true,
			getRkapF: func(r *domain.FinancialReportModel) float64 {
				if r == nil {
					return 0
				}
				return r.ROI
			},
			getRealF: func(r domain.FinancialReportModel) float64 {
				return r.ROI
			},
		},
		{
			label:   "Rasio Lancar",
			field:   "current_ratio",
			isRatio: true,
			getRkapF: func(r *domain.FinancialReportModel) float64 {
				if r == nil {
					return 0
				}
				return r.CurrentRatio
			},
			getRealF: func(r domain.FinancialReportModel) float64 {
				return r.CurrentRatio
			},
		},
		{
			label:   "Rasio Kas",
			field:   "cash_ratio",
			isRatio: true,
			getRkapF: func(r *domain.FinancialReportModel) float64 {
				if r == nil {
					return 0
				}
				return r.CashRatio
			},
			getRealF: func(r domain.FinancialReportModel) float64 {
				return r.CashRatio
			},
		},
	}

	// Write table
	tableDataStartRow, _ := uc.writeFinancialTable(f, sheetName, items, reports, rkap, 3)

	numMonths := len(reports)
	if numMonths == 0 && rkap != nil {
		numMonths = 3
	}

	// IMPORTANT: Calculate exact end of table and place chart data far below
	tableEndRow := tableDataStartRow + numMonths - 1
	chartDataStartRow := tableEndRow + 50

	// Chart data - ROE and ROI
	chartMonthRow := chartDataStartRow
	chartDataRow := chartMonthRow + 1

	// Write month labels for chart
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartMonthRow), "Month")
	col := 2
	monthNamesShort := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	for i := 0; i < numMonths; i++ {
		var monthName string
		if i < len(reports) && len(reports[i].Period) >= 7 {
			monthNum := reports[i].Period[5:7]
			var monthIndex int
			_, _ = fmt.Sscanf(monthNum, "%d", &monthIndex)
			if monthIndex > 0 && monthIndex <= 12 {
				monthName = monthNamesShort[monthIndex-1]
			} else {
				monthName = monthNum
			}
		} else {
			monthName = fmt.Sprintf("M%d", i+1)
		}
		cell, _ := excelize.CoordinatesToCellName(col, chartMonthRow)
		_ = f.SetCellValue(sheetName, cell, monthName)
		col++
	}

	// Chart data rows: ROE and ROI
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "ROE (RKAP)")
	col = 2
	rkapROE := 0.0
	if rkap != nil {
		rkapROE = rkap.ROE
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapROE)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "ROE (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.ROE)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "ROI (RKAP)")
	col = 2
	rkapROI := 0.0
	if rkap != nil {
		rkapROI = rkap.ROI
	}
	for i := 0; i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, rkapROI)
		col++
	}
	chartDataRow++
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", chartDataRow), "ROI (Realisasi)")
	col = 2
	for _, report := range reports {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, report.ROI)
		col++
	}
	for i := len(reports); i < numMonths; i++ {
		cell, _ := excelize.CoordinatesToCellName(col, chartDataRow)
		_ = f.SetCellValue(sheetName, cell, 0)
		col++
	}

	// Create chart
	if numMonths > 0 {
		// Calculate the last column: data starts at column 2 (B), so last column = 2 + numMonths - 1 = 1 + numMonths
		lastCol := 1 + numMonths
		lastColLetter := uc.getColumnLetter(lastCol)
		categoriesRange := uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow, lastColLetter, chartMonthRow))

		err = f.AddChart(sheetName, fmt.Sprintf("%s%d", uc.getColumnLetter(len(items)*2+3), 3), &excelize.Chart{
			Type: excelize.Line,
			Series: []excelize.ChartSeries{
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+1)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+1, lastColLetter, chartMonthRow+1)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+2)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+2, lastColLetter, chartMonthRow+2)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+3)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+3, lastColLetter, chartMonthRow+3)),
				},
				{
					Name:       uc.formatExcelRef(sheetName, fmt.Sprintf("$A$%d", chartMonthRow+4)),
					Categories: categoriesRange,
					Values:     uc.formatExcelRef(sheetName, fmt.Sprintf("$B$%d:$%s$%d", chartMonthRow+4, lastColLetter, chartMonthRow+4)),
				},
			},
			Title: []excelize.RichTextRun{
				{Text: "Ratio Overview"},
			},
			Legend: excelize.ChartLegend{
				Position: "right",
			},
			PlotArea: excelize.ChartPlotArea{
				ShowCatName: false,
				ShowSerName: true,
				ShowVal:     false,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to add chart: %w", err)
		}
	}

	return nil
}
