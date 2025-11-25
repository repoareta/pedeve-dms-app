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
		// Update email jika belum ada atau kosong (untuk backward compatibility)
		if existingUser.Email == "" || existingUser.Email == "superadmin@example.com" {
			existingUser.Email = "superadmin@pertamina.com"
			if err := database.GetDB().Save(&existingUser).Error; err != nil {
				zapLog.Warn("Failed to update superadmin email", zap.Error(err))
			} else {
				zapLog.Info("Superadmin email updated", zap.String("email", existingUser.Email))
			}
		}
		zapLog.Info("Superadmin user already exists")
		return
	}

	// Get superadmin role
	var superadminRole domain.RoleModel
	if err := database.GetDB().Where("name = ?", "superadmin").First(&superadminRole).Error; err != nil {
		zapLog.Warn("Superadmin role not found, user will be created without role_id", zap.Error(err))
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
		Role:      "superadmin", // Legacy field
		RoleID:    &superadminRole.ID, // New RBAC field
		CompanyID: nil, // Superadmin tidak punya company (global access)
		IsActive:  true,
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
		zap.String("role_id", superadminRole.ID),
	)
}

