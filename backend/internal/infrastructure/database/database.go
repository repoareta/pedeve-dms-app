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
	// Catatan: AutoMigrate akan otomatis sync schema (tambah kolom, index, dll) saat aplikasi start
	// Ini berarti setiap kali deploy ke GCP, schema akan otomatis ter-update sesuai model terbaru
	// TAPI: Data di local TIDAK ikut ter-copy ke production, hanya schema/structure yang di-sync
	err = DB.AutoMigrate(
		&domain.UserModel{},
		&domain.TwoFactorAuth{},
		&domain.AuditLog{},
		&domain.UserActivityLog{}, // Permanent audit log untuk data penting (report, document, company, user)
		&domain.CompanyModel{},
		&domain.RoleModel{},
		&domain.PermissionModel{},
		&domain.RolePermissionModel{},
		&domain.ShareholderModel{},
		&domain.BusinessFieldModel{},
		&domain.DirectorModel{},
		&domain.UserCompanyAssignmentModel{},
		&domain.ReportModel{},          // Report Management
		&domain.FinancialReportModel{}, // Financial Report (RKAP & Realisasi)
		&domain.DocumentFolderModel{},
		&domain.DocumentModel{},
		&domain.DocumentTypeModel{}, // Document Types Management
		&domain.ShareholderTypeModel{},
		&domain.DirectorPositionModel{},     // Shareholder Types Management
		&domain.NotificationModel{},         // Notifications
		&domain.NotificationSettingsModel{}, // Notification Settings
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

	// Migration: Add shareholder_company_id field to shareholders table if not exists
	if dbURL != "" {
		// PostgreSQL
		if err := DB.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 
					FROM information_schema.columns 
					WHERE table_name = 'shareholders' 
					AND column_name = 'shareholder_company_id'
				) THEN
					ALTER TABLE shareholders 
					ADD COLUMN shareholder_company_id VARCHAR(255);
					CREATE INDEX IF NOT EXISTS idx_shareholders_shareholder_company_id ON shareholders(shareholder_company_id);
				END IF;
			END $$;
		`).Error; err != nil {
			zapLog.Warn("Failed to add shareholder_company_id column with PostgreSQL syntax (may already exist)", zap.Error(err))
		} else {
			zapLog.Info("Shareholder company_id column migration completed (PostgreSQL)")
		}
	} else {
		// SQLite - check if column exists first
		var count int64
		if err := DB.Raw(`
			SELECT COUNT(*) FROM pragma_table_info('shareholders') WHERE name = 'shareholder_company_id'
		`).Scan(&count).Error; err == nil && count == 0 {
			if err := DB.Exec("ALTER TABLE shareholders ADD COLUMN shareholder_company_id VARCHAR(255)").Error; err != nil {
				zapLog.Warn("Failed to add shareholder_company_id column to SQLite (may already exist)", zap.Error(err))
			} else {
				zapLog.Info("Shareholder company_id column migration completed (SQLite)")
			}
		} else {
			zapLog.Info("Shareholder company_id column already exists or check failed")
		}
	}

	// Migration: Update ownership_percent column to support 10 decimal places
	if dbURL != "" {
		// PostgreSQL - change to numeric(20,10) for 10 decimal places
		if err := DB.Exec(`
			DO $$
			BEGIN
				ALTER TABLE shareholders 
				ALTER COLUMN ownership_percent TYPE NUMERIC(20,10);
			EXCEPTION
				WHEN OTHERS THEN
					-- Ignore if column doesn't exist or type is already correct
					NULL;
			END $$;
		`).Error; err != nil {
			zapLog.Warn("Failed to update ownership_percent column type (may already be correct)", zap.Error(err))
		} else {
			zapLog.Info("Ownership percent column type migration completed (PostgreSQL)")
		}
	}

	// Migration: Add currency field to companies table if not exists
	// This ensures backward compatibility with existing databases
	// Try PostgreSQL syntax first (DO block)
	if dbURL != "" {
		// PostgreSQL
		if err := DB.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 
					FROM information_schema.columns 
					WHERE table_name = 'companies' 
					AND column_name = 'currency'
				) THEN
					ALTER TABLE companies 
					ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'IDR';
				END IF;
			END $$;
		`).Error; err != nil {
			zapLog.Warn("Failed to add currency column with PostgreSQL syntax (may already exist)", zap.Error(err))
		} else {
			zapLog.Info("Currency column migration completed (PostgreSQL)")
		}
	} else {
		// SQLite - check if column exists first
		var count int64
		if err := DB.Raw(`
			SELECT COUNT(*) FROM pragma_table_info('companies') WHERE name = 'currency'
		`).Scan(&count).Error; err == nil && count == 0 {
			if err := DB.Exec("ALTER TABLE companies ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'IDR'").Error; err != nil {
				zapLog.Warn("Failed to add currency column to SQLite (may already exist)", zap.Error(err))
			} else {
				zapLog.Info("Currency column migration completed (SQLite)")
			}
		} else {
			zapLog.Info("Currency column already exists or check failed")
		}
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

	// Notification indexes untuk optimasi query unread count
	// Composite index untuk query: WHERE user_id = ? AND is_read = ?
	// Index ini sangat penting untuk optimasi GetUnreadCount yang dipanggil setiap 30 detik
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_id_is_read ON notifications(user_id, is_read)").Error; err != nil {
		zapLog.Warn("Failed to create composite index idx_notifications_user_id_is_read", zap.Error(err))
	} else {
		zapLog.Info("Composite index created for notifications", zap.String("index", "idx_notifications_user_id_is_read"))
	}

	// Index untuk query superadmin: WHERE is_read = ? (untuk count semua unread)
	// Coba partial index dulu (PostgreSQL), jika gagal fallback ke regular index (SQLite/PostgreSQL)
	// Partial index lebih efisien karena hanya index row dengan is_read = false
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read) WHERE is_read = false").Error; err != nil {
		// Partial index tidak didukung (SQLite) atau gagal, gunakan regular index
		zapLog.Debug("Partial index not supported or failed, using regular index", zap.Error(err))
		if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_is_read_fallback ON notifications(is_read)").Error; err != nil {
			zapLog.Warn("Failed to create index idx_notifications_is_read_fallback", zap.Error(err))
		} else {
			zapLog.Info("Regular index created for notifications is_read", zap.String("index", "idx_notifications_is_read_fallback"))
		}
	} else {
		zapLog.Info("Partial index created for notifications is_read (PostgreSQL)", zap.String("index", "idx_notifications_is_read"))
	}

	zapLog.Info("Database connected and migrated successfully")
}

// GetDB mengembalikan instance database
func GetDB() *gorm.DB {
	return DB
}
