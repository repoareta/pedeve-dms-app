# Testing Reports Feature

## Overview
File test untuk fitur Reports sudah disiapkan dan siap digunakan ketika backend Reports diimplementasikan.

## File Test yang Tersedia

### 1. Repository Test
**File**: `backend/internal/repository/report_repository_test.go`

Test cases yang disiapkan:
- `TestReportRepository_Create` - Test membuat report baru
- `TestReportRepository_GetByID` - Test mengambil report by ID
- `TestReportRepository_GetAll` - Test mengambil semua reports
- `TestReportRepository_Update` - Test update report
- `TestReportRepository_Delete` - Test delete report

### 2. UseCase Test
**File**: `backend/internal/usecase/report_usecase_test.go`

Test cases yang disiapkan:
- `TestReportUseCase_CreateReport` - Test business logic untuk create report
- `TestReportUseCase_GetReport` - Test business logic untuk get report
- `TestReportUseCase_GetAllReports` - Test business logic untuk get all reports
- `TestReportUseCase_UpdateReport` - Test business logic untuk update report
- `TestReportUseCase_DeleteReport` - Test business logic untuk delete report

## Cara Menggunakan

### Step 1: Buat Model Report
Tambahkan `ReportModel` di `backend/internal/domain/models.go`:

```go
type ReportModel struct {
    ID            string    `gorm:"primaryKey" json:"id"`
    Period        string    `gorm:"not null;index" json:"period"` // Format: YYYY-MM
    CompanyID     string    `gorm:"not null;index" json:"company_id"`
    InputterID    string    `gorm:"not null;index" json:"inputter_id"`
    Revenue       int64     `gorm:"not null" json:"revenue"`
    Opex          int64     `gorm:"not null" json:"opex"`
    NPAT          int64     `gorm:"not null" json:"npat"`
    Dividend      int64     `gorm:"not null" json:"dividend"`
    FinancialRatio *float64  `gorm:"type:decimal(10,2)" json:"financial_ratio"`
    Attachment    *string   `gorm:"type:text" json:"attachment"`
    Remark        *string   `gorm:"type:text" json:"remark"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    
    // Relationships
    Company  *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
    Inputter *UserModel   `gorm:"foreignKey:InputterID" json:"inputter,omitempty"`
}

func (ReportModel) TableName() string {
    return "reports"
}
```

### Step 2: Update Test Helper
Tambahkan `&domain.ReportModel{}` di `backend/test/helpers/database.go` pada fungsi `SetupTestDB` dan `SetupTestDBPostgres`:

```go
err = db.AutoMigrate(
    // ... existing models ...
    &domain.ReportModel{},
)
```

### Step 3: Buat Repository
Buat file `backend/internal/repository/report_repository.go` dengan interface dan implementasi.

### Step 4: Buat UseCase
Buat file `backend/internal/usecase/report_usecase.go` dengan business logic.

### Step 5: Uncomment Test Cases
Uncomment semua test cases di:
- `report_repository_test.go`
- `report_usecase_test.go`

### Step 6: Implement Helper Functions
Implement helper functions yang ada di bagian bawah file test (yang masih dalam komentar).

### Step 7: Run Tests
```bash
# Run all tests
make test

# Run only report tests
cd backend
go test ./internal/repository/report_repository_test.go -v
go test ./internal/usecase/report_usecase_test.go -v
```

## Best Practices

1. **Test Coverage**: Pastikan semua CRUD operations ter-cover
2. **Edge Cases**: Test untuk invalid input, missing required fields, non-existent IDs
3. **Relationships**: Test foreign key constraints (company_id, inputter_id)
4. **Validation**: Test business rules (misalnya: revenue harus positif, period format harus YYYY-MM)

## Catatan

- Semua test cases saat ini dalam bentuk komentar karena model Report belum dibuat
- Test helper functions juga perlu diimplementasikan
- Pastikan test mengikuti pattern yang sama dengan test lain (Company, User, dll)

