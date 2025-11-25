package domain

import "time"

// User merepresentasikan user dalam sistem (domain model)
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`      // Legacy field, akan deprecated
	Password  string    `json:"-"`         // Jangan sertakan password di JSON
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
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Role      string    `gorm:"default:'user'" json:"role"` // Legacy field, akan deprecated
	Password  string    `gorm:"not null" json:"-"`
	CompanyID *string   `gorm:"index" json:"company_id"`     // NULL untuk superadmin
	RoleID    *string   `gorm:"index" json:"role_id"`        // Reference ke Role table
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
	UserID     string    `gorm:"index" json:"user_id"`              // Optional untuk system-level errors
	Username   string    `gorm:"index" json:"username"`              // Optional untuk system-level errors
	Action     string    `gorm:"index;not null" json:"action"`       // login, logout, create_document, dll
	Resource   string    `gorm:"index" json:"resource"`              // auth, document, user, dll
	ResourceID string    `gorm:"index" json:"resource_id"`           // ID dari resource yang dioperasikan
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `gorm:"type:text" json:"details"`           // JSON string untuk detail tambahan
	Status     string    `gorm:"index;not null" json:"status"`       // success, failure, error
	LogType    string    `gorm:"index;default:'user_action'" json:"log_type"` // user_action atau technical_error
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName menentukan nama tabel untuk AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
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
	Parent   *Company   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Company  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Users    []User     `json:"users,omitempty" gorm:"foreignKey:CompanyID"`
}

// CompanyModel untuk database (entity)
type CompanyModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;index" json:"name"`
	Code        string    `gorm:"uniqueIndex;not null" json:"code"` // Unique company code
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *string   `gorm:"index" json:"parent_id"`           // NULL untuk root/holding company
	Level       int       `gorm:"not null;default:0;index" json:"level"` // 0=root, 1=holding, 2=subsidiary, etc
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CompanyModel) TableName() string {
	return "companies"
}

// ============================================================================
// ROLE & PERMISSION MODELS
// ============================================================================

// PermissionScope menentukan scope dari permission
type PermissionScope string

const (
	ScopeGlobal    PermissionScope = "global"    // Superadmin only
	ScopeCompany   PermissionScope = "company"   // Company-level access
	ScopeSubCompany PermissionScope = "sub_company" // Sub-company level access
)

// Permission merepresentasikan sebuah permission (domain model)
type Permission struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`        // e.g., "view_dashboard", "manage_users"
	Description string           `json:"description"` // Human-readable description
	Resource    string           `json:"resource"`    // e.g., "dashboard", "users", "documents"
	Action      string           `json:"action"`     // e.g., "view", "create", "update", "delete"
	Scope       PermissionScope  `json:"scope"`      // global, company, sub_company
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// PermissionModel untuk database (entity)
type PermissionModel struct {
	ID          string           `gorm:"primaryKey" json:"id"`
	Name        string           `gorm:"uniqueIndex;not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Resource    string           `gorm:"not null;index" json:"resource"`
	Action      string           `gorm:"not null;index" json:"action"`
	Scope       PermissionScope  `gorm:"not null;default:'company';index" json:"scope"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
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
	RoleID       string `gorm:"primaryKey;index" json:"role_id"`
	PermissionID string `gorm:"primaryKey;index" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermissionModel) TableName() string {
	return "role_permissions"
}

