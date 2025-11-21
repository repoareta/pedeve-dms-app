package database

import (
	"os"

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

