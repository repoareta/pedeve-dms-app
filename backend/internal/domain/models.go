package domain

import "time"

// User merepresentasikan user dalam sistem (domain model)
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Password  string    `json:"-"` // Jangan sertakan password di JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	Role      string    `gorm:"default:'user'" json:"role"`
	Password  string    `gorm:"not null" json:"-"`
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

