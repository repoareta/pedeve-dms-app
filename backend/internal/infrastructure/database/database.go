package database

import (
	"os"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/secrets"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// getDatabaseURL mendapatkan database URL dari GCP Secret Manager, Vault, atau environment variable
func getDatabaseURL() string {
	// Priority 1: Try to get from Secret Manager (GCP Secret Manager atau Vault)
	dbURL, err := secrets.GetSecretWithFallback("database_url", "DATABASE_URL", "")
	if err == nil && dbURL != "" {
		return dbURL
	}

	// Priority 2: Fallback to environment variable
	return os.Getenv("DATABASE_URL")
}

// InitDB menginisialisasi koneksi database
func InitDB() {
	zapLog := logger.GetLogger()
	var err error
	var dialector gorm.Dialector

	// Ambil URL database dari Vault atau environment variable
	dbURL := getDatabaseURL()

	// Gunakan SQLite untuk development jika DATABASE_URL tidak diset
	if dbURL == "" {
		zapLog.Info("Using SQLite database (development)")
		dialector = sqlite.Open("dms.db")
	} else {
		zapLog.Info("Using PostgreSQL database")
		dialector = postgres.Open(dbURL)
	}

	// Buka koneksi database
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		zapLog.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Konfigurasi connection pooling (hanya untuk PostgreSQL)
	// SQLite tidak memerlukan connection pooling karena file-based
	if dbURL != "" {
		sqlDB, err := DB.DB()
		if err != nil {
			zapLog.Fatal("Failed to get underlying sql.DB", zap.Error(err))
		}

		// SetMaxOpenConns: Maksimal koneksi yang bisa dibuka secara bersamaan
		// Default: unlimited, kita set 25 untuk production stability
		sqlDB.SetMaxOpenConns(25)

		// SetMaxIdleConns: Maksimal koneksi idle yang dipertahankan di pool
		// Default: 2, kita set 5 untuk mengurangi overhead pembuatan koneksi baru
		sqlDB.SetMaxIdleConns(5)

		// SetConnMaxLifetime: Maksimal waktu hidup koneksi sebelum ditutup
		// Default: unlimited, kita set 5 menit untuk mencegah koneksi stale
		sqlDB.SetConnMaxLifetime(5 * time.Minute)

		// SetConnMaxIdleTime: Maksimal waktu koneksi idle sebelum ditutup
		// Default: unlimited, kita set 10 menit untuk cleanup otomatis
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)

		zapLog.Info("Connection pool configured",
			zap.Int("max_open_conns", 25),
			zap.Int("max_idle_conns", 5),
			zap.Duration("conn_max_lifetime", 5*time.Minute),
			zap.Duration("conn_max_idle_time", 10*time.Minute),
		)
	}

	// Auto migrate schema
	err = DB.AutoMigrate(
		&domain.UserModel{},
		&domain.TwoFactorAuth{},
		&domain.AuditLog{},
		&domain.CompanyModel{},
		&domain.RoleModel{},
		&domain.PermissionModel{},
		&domain.RolePermissionModel{},
		&domain.ShareholderModel{},
		&domain.BusinessFieldModel{},
		&domain.DirectorModel{},
		&domain.UserCompanyAssignmentModel{},
	)
	if err != nil {
		zapLog.Fatal("Failed to migrate database", zap.Error(err))
	}

	// Ensure 'role' column on users table has no default value.
	// This is important so that new users created without an explicit role
	// don't accidentally get a default like 'user' or 'superadmin'.
	if err := DB.Exec("ALTER TABLE users ALTER COLUMN role DROP DEFAULT").Error; err != nil {
		zapLog.Warn("Failed to drop default for users.role (this may be expected on SQLite or if already dropped)", zap.Error(err))
	}

	// Create indexes untuk performance
	// Company hierarchy indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_companies_parent_id ON companies(parent_id)").Error; err != nil {
		zapLog.Warn("Failed to create index idx_companies_parent_id", zap.Error(err))
	}
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_companies_level ON companies(level)").Error; err != nil {
		zapLog.Warn("Failed to create index idx_companies_level", zap.Error(err))
	}

	// User company relationship indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id)").Error; err != nil {
		zapLog.Warn("Failed to create index idx_users_company_id", zap.Error(err))
	}
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id)").Error; err != nil {
		zapLog.Warn("Failed to create index idx_users_role_id", zap.Error(err))
	}

	zapLog.Info("Database connected and migrated successfully")
}

// GetDB mengembalikan instance database
func GetDB() *gorm.DB {
	return DB
}
