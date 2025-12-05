package domain

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/datatypes"
)

// User merepresentasikan user dalam sistem (domain model)
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`       // Legacy field, akan deprecated
	Password  string    `json:"-"`          // Jangan sertakan password di JSON
	CompanyID *string   `json:"company_id"` // NULL untuk superadmin, required untuk user lain
	RoleID    *string   `json:"role_id"`    // Reference ke Role table
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships (populated on query)
	Company    *Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	RoleDetail *Role    `json:"role_detail,omitempty" gorm:"foreignKey:RoleID"`
}

// LoginRequest merepresentasikan payload request login
type LoginRequest struct {
	Username string `json:"username" example:"admin" validate:"required,min=3"` // Bisa username atau email
	Password string `json:"password" example:"password123" validate:"required,min=8"`
	Code     string `json:"code,omitempty" example:"123456" validate:"omitempty,len=6,numeric"` // Kode 2FA (opsional)
}

// RegisterRequest merepresentasikan payload request registrasi (untuk dokumentasi saja, endpoint sudah dihapus)
type RegisterRequest struct {
	Username string `json:"username" example:"admin" validate:"required,min=3,max=50,alphanum_underscore"`
	Email    string `json:"email" example:"admin@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=8,max=128,password_strength"`
}

// AuthResponse merepresentasikan response autentikasi
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse merepresentasikan response error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// UserModel untuk database (entity)
type UserModel struct {
	ID       string `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	// Role is a legacy field. We intentionally do NOT set a default here so that
	// new users can be created in "standby" mode without any role.
	Role      string    `json:"role"` // Legacy field, akan deprecated (can be empty for standby users)
	Password  string    `gorm:"not null" json:"-"`
	CompanyID *string   `gorm:"index" json:"company_id"` // NULL untuk superadmin
	RoleID    *string   `gorm:"index" json:"role_id"`    // Reference ke Role table
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName menentukan nama tabel untuk UserModel
func (UserModel) TableName() string {
	return "users"
}

// TwoFactorAuth merepresentasikan pengaturan 2FA untuk user
type TwoFactorAuth struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret      string    `gorm:"not null" json:"-"` // Secret TOTP
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	BackupCodes string    `gorm:"type:text" json:"-"` // Array JSON dari backup codes
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName menentukan nama tabel untuk TwoFactorAuth
func (TwoFactorAuth) TableName() string {
	return "two_factor_auths"
}

// AuditLog merepresentasikan audit log entry
type AuditLog struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index" json:"user_id"`         // Optional untuk system-level errors
	Username   string    `gorm:"index" json:"username"`        // Optional untuk system-level errors
	Action     string    `gorm:"index;not null" json:"action"` // login, logout, create_document, dll
	Resource   string    `gorm:"index" json:"resource"`        // auth, document, user, dll
	ResourceID string    `gorm:"index" json:"resource_id"`     // ID dari resource yang dioperasikan
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `gorm:"type:text" json:"details"`                    // JSON string untuk detail tambahan
	Status     string    `gorm:"index;not null" json:"status"`                // success, failure, error
	LogType    string    `gorm:"index;default:'user_action'" json:"log_type"` // user_action atau technical_error
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName menentukan nama tabel untuk AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// UserActivityLog merepresentasikan permanent audit log untuk data penting
// Data ini tidak akan dihapus (permanent storage) untuk compliance dan legal purposes
// Resource: report, document, company, user
type UserActivityLog struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"index" json:"user_id"`
	Username   string    `gorm:"index" json:"username"`
	Action     string    `gorm:"index;not null" json:"action"`   // create_report, update_document, dll
	Resource   string    `gorm:"index;not null" json:"resource"` // report, document, company, user
	ResourceID string    `gorm:"index" json:"resource_id"`       // ID dari resource yang dioperasikan
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `gorm:"type:text" json:"details"`     // JSON string untuk detail tambahan
	Status     string    `gorm:"index;not null" json:"status"` // success, failure, error
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName menentukan nama tabel untuk UserActivityLog
func (UserActivityLog) TableName() string {
	return "user_activity_logs"
}

// Document merepresentasikan sebuah document (domain model)
type Document struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// DocumentFolderModel merepresentasikan folder penyimpanan dokumen
type DocumentFolderModel struct {
	ID        string                `gorm:"primaryKey" json:"id"`
	Name      string                `gorm:"uniqueIndex;not null" json:"name"`
	ParentID  *string               `gorm:"index" json:"parent_id"`
	CreatedBy string                `gorm:"index;not null;default:''" json:"created_by"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Children  []DocumentFolderModel `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (DocumentFolderModel) TableName() string {
	return "document_folders"
}

// DocumentModel merepresentasikan file dokumen yang diupload
// @Description Document model dengan metadata dalam format JSON
type DocumentModel struct {
	ID         string         `gorm:"primaryKey" json:"id"`
	FolderID   *string        `gorm:"index" json:"folder_id"`
	Name       string         `gorm:"not null" json:"name"`      // Judul dokumen
	FileName   string         `gorm:"not null" json:"file_name"` // Nama file asli
	FilePath   string         `gorm:"not null" json:"file_path"` // URL/path hasil upload
	MimeType   string         `gorm:"not null" json:"mime_type"`
	Size       int64          `gorm:"not null" json:"size"` // Size in bytes
	Status     string         `gorm:"default:'active'" json:"status"`
	Metadata   datatypes.JSON `json:"metadata" swaggertype:"object"` // Metadata tambahan (opsional, format JSON)
	UploaderID string         `gorm:"index" json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`

	Folder *DocumentFolderModel `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
}

func (DocumentModel) TableName() string {
	return "documents"
}

// DocumentFolderStat menyimpan agregasi dokumen per folder
type DocumentFolderStat struct {
	FolderID  *string `json:"folder_id"`
	FileCount int64   `json:"file_count"`
	TotalSize int64   `json:"total_size"`
}

// DocumentTypeModel merepresentasikan jenis dokumen yang bisa digunakan
type DocumentTypeModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"` // Nama jenis dokumen (unique)
	IsActive    bool      `gorm:"default:true;index" json:"is_active"` // Soft delete: false jika dihapus
	UsageCount  int64     `gorm:"default:0" json:"usage_count"` // Jumlah dokumen yang menggunakan jenis ini
	CreatedBy   string    `gorm:"index;not null" json:"created_by"` // User yang membuat
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DocumentTypeModel) TableName() string {
	return "document_types"
}

// ============================================================================
// COMPANY HIERARCHY MODELS
// ============================================================================

// Company merepresentasikan perusahaan dalam hierarchy (domain model)
type Company struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`        // Unique company code
	Description string    `json:"description"` // Optional description
	ParentID    *string   `json:"parent_id"`   // NULL untuk root/holding company
	Level       int       `json:"level"`       // 0=root, 1=holding, 2=subsidiary, 3=sub-subsidiary, etc
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships (populated on query)
	Parent   *Company  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Company `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Users    []User    `json:"users,omitempty" gorm:"foreignKey:CompanyID"`
}

// CompanyModel untuk database (entity)
type CompanyModel struct {
	ID                  string    `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"not null;index" json:"name"`
	ShortName           string    `gorm:"index" json:"short_name"`          // Nama singkat
	Code                string    `gorm:"uniqueIndex;not null" json:"code"` // Unique company code
	Description         string    `gorm:"type:text" json:"description"`
	NPWP                string    `gorm:"index" json:"npwp"`                     // Nomor Pokok Wajib Pajak
	NIB                 string    `gorm:"index" json:"nib"`                      // Nomor Induk Berusaha
	Status              string    `gorm:"default:'Aktif'" json:"status"`         // Status perusahaan
	Logo                string    `json:"logo"`                                  // Path/URL logo
	Phone               string    `json:"phone"`                                 // Telepon
	Fax                 string    `json:"fax"`                                   // Fax
	Email               string    `json:"email"`                                 // Email
	Website             string    `json:"website"`                               // Website
	Address             string    `gorm:"type:text" json:"address"`              // Alamat perusahaan
	OperationalAddress  string    `gorm:"type:text" json:"operational_address"`  // Alamat operasional
	ParentID            *string   `gorm:"index" json:"parent_id"`                // NULL untuk root/holding company
	MainParentCompanyID *string   `gorm:"index" json:"main_parent_company"`      // ID perusahaan induk utama
	Level               int       `gorm:"not null;default:0;index" json:"level"` // 0=root, 1=holding, 2=subsidiary, etc
	IsActive            bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Relationships
	Shareholders   []ShareholderModel   `gorm:"foreignKey:CompanyID" json:"shareholders,omitempty"`
	BusinessFields []BusinessFieldModel `gorm:"foreignKey:CompanyID" json:"business_fields,omitempty"`
	Directors      []DirectorModel      `gorm:"foreignKey:CompanyID" json:"directors,omitempty"`
}

func (CompanyModel) TableName() string {
	return "companies"
}

// ShareholderModel merepresentasikan pemegang saham perusahaan
type ShareholderModel struct {
	ID               string    `gorm:"primaryKey" json:"id"`
	CompanyID        string    `gorm:"index;not null" json:"company_id"`
	Type             string    `gorm:"not null" json:"type"`                // Jenis: Badan Hukum, Individu, dll
	Name             string    `gorm:"not null" json:"name"`                // Nama pemegang saham
	IdentityNumber   string    `json:"identity_number"`                     // KTP/NPWP
	OwnershipPercent float64   `gorm:"not null" json:"ownership_percent"`   // Persentase kepemilikan
	ShareCount       int64     `json:"share_count"`                         // Jumlah saham
	IsMainParent     bool      `gorm:"default:false" json:"is_main_parent"` // Apakah perusahaan induk utama
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (ShareholderModel) TableName() string {
	return "shareholders"
}

// BusinessFieldModel merepresentasikan bidang usaha perusahaan
type BusinessFieldModel struct {
	ID                   string     `gorm:"primaryKey" json:"id"`
	CompanyID            string     `gorm:"index;not null" json:"company_id"`
	IndustrySector       string     `gorm:"not null" json:"industry_sector"`         // Sektor industri
	KBLI                 string     `json:"kbli"`                                    // Klasifikasi Baku Lapangan Usaha Indonesia
	MainBusinessActivity string     `gorm:"type:text" json:"main_business_activity"` // Uraian kegiatan usaha utama
	AdditionalActivities string     `gorm:"type:text" json:"additional_activities"`  // Kegiatan usaha tambahan
	StartOperationDate   *time.Time `json:"start_operation_date"`                    // Tanggal mulai beroperasi
	IsMain               bool       `gorm:"default:true" json:"is_main"`             // Apakah bidang usaha utama
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (BusinessFieldModel) TableName() string {
	return "business_fields"
}

// DirectorModel merepresentasikan pengurus/dewan direksi perusahaan
type DirectorModel struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	CompanyID       string     `gorm:"index;not null" json:"company_id"`
	Position        string     `gorm:"not null" json:"position"`          // Jabatan: Direktur Utama, Komisaris, dll
	FullName        string     `gorm:"not null" json:"full_name"`         // Nama lengkap
	KTP             string     `json:"ktp"`                               // Nomor KTP
	NPWP            string     `json:"npwp"`                              // Nomor NPWP
	StartDate       *time.Time `json:"start_date"`                        // Tanggal awal jabatan (nullable)
	DomicileAddress string     `gorm:"type:text" json:"domicile_address"` // Alamat domisili
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (DirectorModel) TableName() string {
	return "directors"
}

// UserCompanyAssignmentModel untuk junction table - support multiple company assignments per user
type UserCompanyAssignmentModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	CompanyID string    `gorm:"index;not null" json:"company_id"`
	RoleID    *string   `gorm:"index" json:"role_id"`                // Role di company ini (bisa berbeda per company)
	IsActive  bool      `gorm:"default:true;index" json:"is_active"` // Status assignment (bisa dinonaktifkan tanpa hapus)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User    *UserModel    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Company *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Role    *RoleModel    `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (UserCompanyAssignmentModel) TableName() string {
	return "user_company_assignments"
}

// ReportModel merepresentasikan laporan bulanan perusahaan
type ReportModel struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	Period         string    `gorm:"index;not null" json:"period"` // Format: YYYY-MM (e.g., "2025-09")
	CompanyID      string    `gorm:"index;not null" json:"company_id"`
	InputterID     *string   `gorm:"index" json:"inputter_id"` // User yang menginput (bisa null)
	Revenue        int64     `gorm:"not null" json:"revenue"`
	Opex           int64     `gorm:"not null" json:"opex"`
	NPAT           int64     `gorm:"not null" json:"npat"` // Net Profit After Tax
	Dividend       int64     `gorm:"not null" json:"dividend"`
	FinancialRatio float64   `gorm:"not null" json:"financial_ratio"` // Mandatory
	Attachment     *string   `gorm:"type:text" json:"attachment"`     // Optional, bisa null
	Remark         *string   `gorm:"type:text" json:"remark"`         // Optional, bisa null
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Company  *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Inputter *UserModel    `gorm:"foreignKey:InputterID" json:"inputter,omitempty"`
}

func (ReportModel) TableName() string {
	return "reports"
}

// CreateReportRequest untuk request body create report
type CreateReportRequest struct {
	Period         string  `json:"period" validate:"required,regexp=^\\d{4}-\\d{2}$"` // Format: YYYY-MM
	CompanyID      string  `json:"company_id" validate:"required"`
	InputterID     *string `json:"inputter_id"` // Optional
	Revenue        int64   `json:"revenue" validate:"required"`
	Opex           int64   `json:"opex" validate:"required"`
	NPAT           int64   `json:"npat" validate:"required"`
	Dividend       int64   `json:"dividend" validate:"required"`
	FinancialRatio float64 `json:"financial_ratio" validate:"required"`
	Attachment     *string `json:"attachment"` // Optional
	Remark         *string `json:"remark"`     // Optional
}

// UpdateReportRequest untuk request body update report
type UpdateReportRequest struct {
	Period         *string  `json:"period" validate:"omitempty,regexp=^\\d{4}-\\d{2}$"`
	CompanyID      *string  `json:"company_id"`
	InputterID     *string  `json:"inputter_id"`
	Revenue        *int64   `json:"revenue"`
	Opex           *int64   `json:"opex"`
	NPAT           *int64   `json:"npat"`
	Dividend       *int64   `json:"dividend"`
	FinancialRatio *float64 `json:"financial_ratio"`
	Attachment     *string  `json:"attachment"`
	Remark         *string  `json:"remark"`
}

// UserCompanyResponse untuk response GetMyCompanies (company dengan role info)
type UserCompanyResponse struct {
	Company   CompanyModel `json:"company"`
	RoleID    *string      `json:"role_id"`
	Role      string       `json:"role"`       // Role name
	RoleLevel int          `json:"role_level"` // Role level untuk sorting (0=superadmin, 1=admin, 2=manager, 3=staff)
}

// ShareholderRequest untuk request body (tanpa CreatedAt/UpdatedAt)
type ShareholderRequest struct {
	Type             string  `json:"type"`
	Name             string  `json:"name"`
	IdentityNumber   string  `json:"identity_number"`
	OwnershipPercent float64 `json:"ownership_percent"`
	ShareCount       int64   `json:"share_count"`
	IsMainParent     bool    `json:"is_main_parent"`
}

// DateOnly untuk parsing tanggal format YYYY-MM-DD
type DateOnly struct {
	time.Time
}

// UnmarshalJSON untuk parsing tanggal format YYYY-MM-DD
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// MarshalJSON untuk serialize tanggal ke format YYYY-MM-DD
func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format("2006-01-02"))
}

// BusinessFieldRequest untuk request body (tanpa CreatedAt/UpdatedAt)
type BusinessFieldRequest struct {
	IndustrySector       string    `json:"industry_sector"`
	KBLI                 string    `json:"kbli"`
	MainBusinessActivity string    `json:"main_business_activity"`
	AdditionalActivities string    `json:"additional_activities"`
	StartOperationDate   *DateOnly `json:"start_operation_date"`
}

// DirectorRequest untuk request body (tanpa CreatedAt/UpdatedAt)
type DirectorRequest struct {
	Position        string    `json:"position"`
	FullName        string    `json:"full_name"`
	KTP             string    `json:"ktp"`
	NPWP            string    `json:"npwp"`
	StartDate       *DateOnly `json:"start_date"`
	DomicileAddress string    `json:"domicile_address"`
}

// CompanyCreateRequest untuk create company dengan data lengkap
type CompanyCreateRequest struct {
	Name               string                `json:"name"`
	ShortName          string                `json:"short_name"`
	Code               string                `json:"code"`
	Description        string                `json:"description"`
	NPWP               string                `json:"npwp"`
	NIB                string                `json:"nib"`
	Status             string                `json:"status"`
	Logo               string                `json:"logo"`
	Phone              string                `json:"phone"`
	Fax                string                `json:"fax"`
	Email              string                `json:"email"`
	Website            string                `json:"website"`
	Address            string                `json:"address"`
	OperationalAddress string                `json:"operational_address"`
	ParentID           *string               `json:"parent_id"`
	MainParentCompany  *string               `json:"main_parent_company"`
	Shareholders       []ShareholderRequest  `json:"shareholders"`
	MainBusiness       *BusinessFieldRequest `json:"main_business"`
	Directors          []DirectorRequest     `json:"directors"`
}

// CompanyUpdateRequest untuk update company dengan data lengkap
type CompanyUpdateRequest struct {
	Name               string                `json:"name"`
	ShortName          string                `json:"short_name"`
	Description        string                `json:"description"`
	NPWP               string                `json:"npwp"`
	NIB                string                `json:"nib"`
	Status             string                `json:"status"`
	Logo               string                `json:"logo"`
	Phone              string                `json:"phone"`
	Fax                string                `json:"fax"`
	Email              string                `json:"email"`
	Website            string                `json:"website"`
	Address            string                `json:"address"`
	OperationalAddress string                `json:"operational_address"`
	ParentID           *string               `json:"parent_id"` // Untuk mengubah parent company
	MainParentCompany  *string               `json:"main_parent_company"`
	Shareholders       []ShareholderRequest  `json:"shareholders"`
	MainBusiness       *BusinessFieldRequest `json:"main_business"`
	Directors          []DirectorRequest     `json:"directors"`
}

// ============================================================================
// ROLE & PERMISSION MODELS
// ============================================================================

// PermissionScope menentukan scope dari permission
type PermissionScope string

const (
	ScopeGlobal     PermissionScope = "global"      // Superadmin only
	ScopeCompany    PermissionScope = "company"     // Company-level access
	ScopeSubCompany PermissionScope = "sub_company" // Sub-company level access
)

// Permission merepresentasikan sebuah permission (domain model)
type Permission struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`        // e.g., "view_dashboard", "manage_users"
	Description string          `json:"description"` // Human-readable description
	Resource    string          `json:"resource"`    // e.g., "dashboard", "users", "documents"
	Action      string          `json:"action"`      // e.g., "view", "create", "update", "delete"
	Scope       PermissionScope `json:"scope"`       // global, company, sub_company
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PermissionModel untuk database (entity)
type PermissionModel struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"uniqueIndex;not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Resource    string          `gorm:"not null;index" json:"resource"`
	Action      string          `gorm:"not null;index" json:"action"`
	Scope       PermissionScope `gorm:"not null;default:'company';index" json:"scope"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (PermissionModel) TableName() string {
	return "permissions"
}

// Role merepresentasikan role dalam sistem (domain model)
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`        // e.g., "superadmin", "admin", "manager", "staff"
	Description string    `json:"description"` // Human-readable description
	Level       int       `json:"level"`       // 0=superadmin, 1=admin, 2=manager, 3=staff
	IsSystem    bool      `json:"is_system"`   // System role tidak bisa dihapus
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
	Users       []User       `json:"users,omitempty" gorm:"foreignKey:RoleID"`
}

// RoleModel untuk database (entity)
type RoleModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Level       int       `gorm:"not null;default:3;index" json:"level"` // 0=superadmin, 1=admin, 2=manager, 3=staff
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"`  // System role tidak bisa dihapus
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RoleModel) TableName() string {
	return "roles"
}

// RolePermissionModel untuk many-to-many relationship
type RolePermissionModel struct {
	RoleID       string    `gorm:"primaryKey;index" json:"role_id"`
	PermissionID string    `gorm:"primaryKey;index" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermissionModel) TableName() string {
	return "role_permissions"
}
