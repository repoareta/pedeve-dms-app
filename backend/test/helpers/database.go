package helpers

import (
	"os"
	"testing"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
// This is fast and doesn't require external database setup
func SetupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for fast tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silence logs during tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&domain.CompanyModel{},
		&domain.UserModel{},
		&domain.RoleModel{},
		&domain.PermissionModel{},
		&domain.RolePermissionModel{},
		&domain.ShareholderModel{},
		&domain.BusinessFieldModel{},
		&domain.DirectorModel{},
		&domain.UserCompanyAssignmentModel{},
		&domain.AuditLog{},
		&domain.UserActivityLog{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// SetupTestDBPostgres creates a PostgreSQL test database (for integration tests)
// Requires TEST_DATABASE_URL environment variable
func SetupTestDBPostgres(t *testing.T) *gorm.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL test")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silence logs during tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test PostgreSQL database: %v", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&domain.CompanyModel{},
		&domain.UserModel{},
		&domain.RoleModel{},
		&domain.PermissionModel{},
		&domain.RolePermissionModel{},
		&domain.ShareholderModel{},
		&domain.BusinessFieldModel{},
		&domain.DirectorModel{},
		&domain.UserCompanyAssignmentModel{},
		&domain.AuditLog{},
		&domain.UserActivityLog{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test PostgreSQL database: %v", err)
	}

	return db
}

// CleanupTestDB drops all tables and closes the database connection
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

