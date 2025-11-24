package database

import (
	"os"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"go.uber.org/zap"
)

var DB *gorm.DB

// InitDB menginisialisasi koneksi database
func InitDB() {
	zapLog := logger.GetLogger()
	var err error
	var dialector gorm.Dialector

	// Ambil URL database dari environment
	dbURL := os.Getenv("DATABASE_URL")

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
	err = DB.AutoMigrate(&domain.UserModel{}, &domain.TwoFactorAuth{}, &domain.AuditLog{})
	if err != nil {
		zapLog.Fatal("Failed to migrate database", zap.Error(err))
	}

	zapLog.Info("Database connected and migrated successfully")
}

// GetDB mengembalikan instance database
func GetDB() *gorm.DB {
	return DB
}

