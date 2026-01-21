package usecase

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"gorm.io/gorm"
)

type CapabilityStatus string

const (
	CapabilityPass CapabilityStatus = "pass"
	CapabilityWarn CapabilityStatus = "warn"
	CapabilityFail CapabilityStatus = "fail"
)

type CapabilityCheck struct {
	Module  string                 `json:"module"`
	Status  CapabilityStatus       `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type AdminCapabilityReport struct {
	UserID string `json:"user_id"`

	RoleNameLower    string  `json:"role_name"`
	PrimaryCompanyID *string `json:"primary_company_id,omitempty"`

	AssignedRootCompanyIDs []string `json:"assigned_root_company_ids"`
	ExpectedCompanyIDs     []string `json:"expected_company_ids"`

	GeneratedAt time.Time         `json:"generated_at"`
	Checks      []CapabilityCheck `json:"checks"`
}

type AdminCapabilityUseCase interface {
	CheckAdminCapabilities(userID string) (*AdminCapabilityReport, error)
}

type adminCapabilityUseCase struct {
	db *gorm.DB

	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	companyRepo    repository.CompanyRepository
	assignmentRepo repository.UserCompanyAssignmentRepository
	documentRepo   repository.DocumentRepository
}

func NewAdminCapabilityUseCase() AdminCapabilityUseCase {
	return NewAdminCapabilityUseCaseWithDB(database.GetDB())
}

func NewAdminCapabilityUseCaseWithDB(db *gorm.DB) AdminCapabilityUseCase {
	return &adminCapabilityUseCase{
		db:             db,
		userRepo:       repository.NewUserRepositoryWithDB(db),
		roleRepo:       repository.NewRoleRepositoryWithDB(db),
		companyRepo:    repository.NewCompanyRepositoryWithDB(db),
		assignmentRepo: repository.NewUserCompanyAssignmentRepositoryWithDB(db),
		documentRepo:   repository.NewDocumentRepositoryWithDB(db),
	}
}

func (uc *adminCapabilityUseCase) CheckAdminCapabilities(userID string) (*AdminCapabilityReport, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("userID is required")
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	roleLower := strings.ToLower(strings.TrimSpace(user.Role))
	if user.RoleID != nil && *user.RoleID != "" {
		if r, rErr := uc.roleRepo.GetByID(*user.RoleID); rErr == nil && r != nil && strings.TrimSpace(r.Name) != "" {
			roleLower = strings.ToLower(strings.TrimSpace(r.Name))
		}
	}

	report := &AdminCapabilityReport{
		UserID:                 user.ID,
		RoleNameLower:          roleLower,
		PrimaryCompanyID:       user.CompanyID,
		AssignedRootCompanyIDs: []string{},
		ExpectedCompanyIDs:     []string{},
		GeneratedAt:            time.Now(),
		Checks:                 []CapabilityCheck{},
	}

	if roleLower != "admin" {
		return report, fmt.Errorf("user role bukan admin")
	}

	rootsSet := map[string]bool{}
	if user.CompanyID != nil && *user.CompanyID != "" {
		rootsSet[*user.CompanyID] = true
	}

	assignments, _ := uc.assignmentRepo.GetByUserID(user.ID)
	for _, a := range assignments {
		if !a.IsActive {
			continue
		}
		if strings.TrimSpace(a.CompanyID) == "" {
			continue
		}
		rootsSet[a.CompanyID] = true
	}

	roots := keysSorted(rootsSet)
	report.AssignedRootCompanyIDs = roots

	expectedSet := map[string]bool{}
	for _, root := range roots {
		expectedSet[root] = true
		if desc, dErr := uc.companyRepo.GetDescendants(root); dErr == nil {
			for _, c := range desc {
				if strings.TrimSpace(c.ID) != "" {
					expectedSet[c.ID] = true
				}
			}
		}
	}
	report.ExpectedCompanyIDs = keysSorted(expectedSet)

	// Permission check (best-effort). Permission enforcement di app saat ini campuran: sebagian pakai permission JWT, sebagian pakai roleName.
	report.Checks = append(report.Checks, uc.checkAdminPermissions(user))

	// Company scope check (harus bisa mencakup root + descendants dari multi-company assignment)
	report.Checks = append(report.Checks, uc.checkCompanyScope(report))

	// Documents module check (admin harus dianggap setara administrator: unrestricted listing)
	report.Checks = append(report.Checks, uc.checkDocumentsListing(report))

	// Reports/Financial reports scope check: banyak flow masih tergantung primary company saja.
	report.Checks = append(report.Checks, uc.checkReportsScopeGap(report))

	// User management scope check (data assignment harus ada untuk multi-company).
	report.Checks = append(report.Checks, uc.checkUserAssignmentSanity(report))

	return report, nil
}

func (uc *adminCapabilityUseCase) checkAdminPermissions(user *domain.UserModel) CapabilityCheck {
	// Required permissions sesuai seed/roles_permissions.go untuk role admin.
	required := []string{
		"company:manage",
		"user:manage",
		"document:manage",
		"dashboard:view",
		"report:generate",
		"audit:view",
	}

	if user.RoleID == nil || *user.RoleID == "" {
		return CapabilityCheck{
			Module:  "permissions",
			Status:  CapabilityWarn,
			Message: "role_id tidak terisi; tidak bisa memverifikasi permission dari tabel roles_permissions",
		}
	}

	perms, err := uc.roleRepo.GetPermissions(*user.RoleID)
	if err != nil {
		return CapabilityCheck{
			Module:  "permissions",
			Status:  CapabilityWarn,
			Message: "gagal mengambil permissions untuk role admin",
		}
	}

	have := map[string]bool{}
	for _, p := range perms {
		if strings.TrimSpace(p.Name) != "" {
			have[p.Name] = true
		}
	}

	missing := []string{}
	for _, r := range required {
		if !have[r] {
			missing = append(missing, r)
		}
	}

	if len(missing) > 0 {
		return CapabilityCheck{
			Module:  "permissions",
			Status:  CapabilityWarn,
			Message: "ada permission admin yang belum ter-assign",
			Details: map[string]interface{}{
				"missing": missing,
			},
		}
	}

	return CapabilityCheck{
		Module: "permissions",
		Status: CapabilityPass,
	}
}

func (uc *adminCapabilityUseCase) checkCompanyScope(report *AdminCapabilityReport) CapabilityCheck {
	if len(report.AssignedRootCompanyIDs) == 0 {
		return CapabilityCheck{
			Module:  "companies",
			Status:  CapabilityFail,
			Message: "admin tidak punya primary company dan tidak ada assignment company aktif",
		}
	}
	return CapabilityCheck{
		Module: "companies",
		Status: CapabilityPass,
		Details: map[string]interface{}{
			"root_count":     len(report.AssignedRootCompanyIDs),
			"expected_count": len(report.ExpectedCompanyIDs),
		},
	}
}

func (uc *adminCapabilityUseCase) checkDocumentsListing(report *AdminCapabilityReport) CapabilityCheck {
	// Admin diminta unrestricted seperti administrator di modul Documents.
	// Listing folder dengan companyIDs kosong berarti "tanpa filter" (lihat document_repository.go).
	folders, err := uc.documentRepo.ListFolders(nil)
	if err != nil {
		return CapabilityCheck{
			Module:  "documents",
			Status:  CapabilityFail,
			Message: "gagal list folders (unrestricted)",
		}
	}

	return CapabilityCheck{
		Module: "documents",
		Status: CapabilityPass,
		Details: map[string]interface{}{
			"folder_count": len(folders),
		},
	}
}

func (uc *adminCapabilityUseCase) checkReportsScopeGap(report *AdminCapabilityReport) CapabilityCheck {
	// Secara historis, modul laporan banyak flow masih memakai primary company saja (companyID dari JWT),
	// sehingga jika admin di-assign ke multiple root companies, ada potensi gap.
	if len(report.AssignedRootCompanyIDs) <= 1 {
		return CapabilityCheck{
			Module: "reports",
			Status: CapabilityPass,
			Details: map[string]interface{}{
				"note": "single-root scope",
			},
		}
	}

	primaryOnlySet := map[string]bool{}
	if report.PrimaryCompanyID != nil && *report.PrimaryCompanyID != "" {
		root := *report.PrimaryCompanyID
		primaryOnlySet[root] = true
		if desc, err := uc.companyRepo.GetDescendants(root); err == nil {
			for _, c := range desc {
				if strings.TrimSpace(c.ID) != "" {
					primaryOnlySet[c.ID] = true
				}
			}
		}
	}

	missing := []string{}
	for _, id := range report.ExpectedCompanyIDs {
		if !primaryOnlySet[id] {
			missing = append(missing, id)
		}
	}

	if len(missing) == 0 {
		return CapabilityCheck{
			Module: "reports",
			Status: CapabilityPass,
		}
	}

	return CapabilityCheck{
		Module:  "reports",
		Status:  CapabilityWarn,
		Message: "potensi gap: sebagian flow laporan masih scope ke primary company saja",
		Details: map[string]interface{}{
			"missing_company_count": len(missing),
		},
	}
}

func (uc *adminCapabilityUseCase) checkUserAssignmentSanity(report *AdminCapabilityReport) CapabilityCheck {
	if len(report.AssignedRootCompanyIDs) == 0 {
		return CapabilityCheck{
			Module:  "user_management",
			Status:  CapabilityFail,
			Message: "tidak ada company root untuk scope user management",
		}
	}
	return CapabilityCheck{
		Module: "user_management",
		Status: CapabilityPass,
		Details: map[string]interface{}{
			"root_count": len(report.AssignedRootCompanyIDs),
		},
	}
}

func keysSorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k, v := range set {
		if !v {
			continue
		}
		if strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
