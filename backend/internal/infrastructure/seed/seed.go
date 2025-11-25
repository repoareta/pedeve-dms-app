package seed

import (
	"fmt"
	"os"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/secrets"
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
		
		// Auto-sync password dari Vault jika enabled (untuk production)
		if shouldSyncSuperadminPassword() {
			if err := syncSuperadminPasswordFromVault(&existingUser); err != nil {
				zapLog.Warn("Failed to sync superadmin password from Vault", zap.Error(err))
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

	// Get superadmin password from Vault or environment variable or use default
	superadminPassword := getSuperadminPassword()
	
	// Hash password for superadmin
	hashedPassword, err := password.HashPassword(superadminPassword)
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
		zap.String("role_id", superadminRole.ID),
		zap.Bool("password_from_vault", superadminPassword != "Pedeve123"),
	)
}

// getSuperadminPassword retrieves superadmin password from Vault, env var, or default
func getSuperadminPassword() string {
	zapLog := logger.GetLogger()
	
	// Priority 1: Vault
	secretManager := secrets.GetSecretManager()
	password, err := secretManager.GetSecret("superadmin_password")
	if err == nil && password != "" {
		zapLog.Info("Superadmin password loaded from Vault")
		return password
	}
	zapLog.Debug("Superadmin password not found in Vault, trying fallback", zap.Error(err))
	
	// Priority 2: Environment variable
	if envPassword := os.Getenv("SUPERADMIN_PASSWORD"); envPassword != "" {
		zapLog.Info("Superadmin password loaded from environment variable")
		return envPassword
	}
	
	// Priority 3: Default (hardcoded fallback)
	zapLog.Warn("Using default superadmin password. Set SUPERADMIN_PASSWORD env var or store in Vault for production!")
	return "Pedeve123"
}

// shouldSyncSuperadminPassword checks if auto-sync password from Vault is enabled
func shouldSyncSuperadminPassword() bool {
	// Enable auto-sync jika env var SUPERADMIN_AUTO_SYNC_PASSWORD=true
	return os.Getenv("SUPERADMIN_AUTO_SYNC_PASSWORD") == "true"
}

// syncSuperadminPasswordFromVault updates superadmin password from Vault if different
func syncSuperadminPasswordFromVault(user *domain.UserModel) error {
	zapLog := logger.GetLogger()
	
	// Get password from Vault
	vaultPassword := getSuperadminPassword()
	if vaultPassword == "" {
		return fmt.Errorf("superadmin password not found in Vault or env")
	}
	
	// Check if password is different (verify current password)
	if password.CheckPasswordHash(vaultPassword, user.Password) {
		// Password sama, tidak perlu update
		zapLog.Debug("Superadmin password already matches Vault, skipping update")
		return nil
	}
	
	// Password berbeda, update dengan password dari Vault
	hashedPassword, err := password.HashPassword(vaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	
	user.Password = hashedPassword
	if err := database.GetDB().Save(user).Error; err != nil {
		return fmt.Errorf("failed to update superadmin password: %w", err)
	}
	
	zapLog.Info("Superadmin password synced from Vault successfully",
		zap.String("username", user.Username),
	)
	return nil
}

// UpdateSuperadminPasswordFromVault updates superadmin password from Vault (public function untuk dipanggil dari luar)
func UpdateSuperadminPasswordFromVault() error {
	zapLog := logger.GetLogger()
	
	var existingUser domain.UserModel
	result := database.GetDB().Where("username = ? OR role = ?", "superadmin", "superadmin").First(&existingUser)
	if result.Error != nil {
		return fmt.Errorf("superadmin user not found: %w", result.Error)
	}
	
	if err := syncSuperadminPasswordFromVault(&existingUser); err != nil {
		zapLog.Error("Failed to update superadmin password from Vault", zap.Error(err))
		return err
	}
	
	return nil
}

