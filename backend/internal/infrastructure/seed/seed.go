package seed

import (
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"go.uber.org/zap"
)

// SeedSuperAdmin creates a superadmin user if it doesn't exist
func SeedSuperAdmin() {
	zapLog := logger.GetLogger()
	
	// Check if superadmin already exists
	var existingUser domain.UserModel
	result := database.GetDB().Where("username = ? OR role = ?", "superadmin", "superadmin").First(&existingUser)
	if result.Error == nil {
		zapLog.Info("Superadmin user already exists")
		return
	}

	// Hash password for superadmin
	hashedPassword, err := password.HashPassword("Pedeve123")
	if err != nil {
		zapLog.Error("Failed to hash superadmin password", zap.Error(err))
		return
	}

	// Create superadmin user
	now := time.Now()
	superAdmin := &domain.UserModel{
		ID:        uuid.GenerateUUID(),
		Username:  "superadmin",
		Email:     "superadmin@pertamina.com",
		Password:  hashedPassword,
		Role:      "superadmin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to database
	if err := database.GetDB().Create(superAdmin).Error; err != nil {
		zapLog.Error("Failed to create superadmin user", zap.Error(err))
		return
	}

	zapLog.Info("Superadmin user created successfully",
		zap.String("username", "superadmin"),
		zap.String("password", "Pedeve123"),
	)
}

