package usecase

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/encryption"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Verify2FALogin memverifikasi kode 2FA saat login
func Verify2FALogin(userID, code string) (bool, error) {
	var twoFA domain.TwoFactorAuth
	result := database.GetDB().Where("user_id = ? AND enabled = ?", userID, true).First(&twoFA)
	if result.Error != nil {
		return false, fmt.Errorf("2FA not enabled")
	}

	// Decrypt secret (handle backward compatibility untuk data lama)
	secret, err := encryption.Decrypt(twoFA.Secret)
	if err != nil {
		// Jika decrypt gagal, gunakan as-is (backward compatibility)
		secret = twoFA.Secret
	}

	// Coba kode TOTP terlebih dahulu
	valid := totp.Validate(code, secret)
	if valid {
		return true, nil
	}

	// Decrypt backup codes (handle backward compatibility)
	backupCodes, err := encryption.Decrypt(twoFA.BackupCodes)
	if err != nil {
		// Jika decrypt gagal, gunakan as-is (backward compatibility)
		backupCodes = twoFA.BackupCodes
	}

	// Coba backup codes
	if verifyBackupCode(code, backupCodes) {
		return true, nil
	}

	return false, fmt.Errorf("invalid code")
}

// Generate2FASecretUseCase menghasilkan secret TOTP baru untuk user
func Generate2FASecretUseCase(userID, username string) (map[string]interface{}, error) {
	zapLog := logger.GetLogger()
	zapLog.Info("Generating 2FA secret", zap.String("user_id", userID), zap.String("username", username))

	// Generate kunci TOTP
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Pedeve Apps",
		AccountName: username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		zapLog.Error("Error generating TOTP key", zap.Error(err))
		return nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	zapLog.Debug("TOTP key generated successfully", zap.String("user_id", userID))

	// Encrypt secret sebelum disimpan
	encryptedSecret, err := encryption.Encrypt(key.Secret())
	if err != nil {
		zapLog.Error("Error encrypting 2FA secret", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt 2FA secret: %w", err)
	}

	// Simpan atau update record 2FA
	var twoFA domain.TwoFactorAuth
	result := database.GetDB().Where("user_id = ?", userID).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		twoFA = domain.TwoFactorAuth{
			ID:      uuid.GenerateUUID(),
			UserID:  userID,
			Secret:  encryptedSecret, // Simpan encrypted secret
			Enabled: false,
		}
		if err := database.GetDB().Create(&twoFA).Error; err != nil {
			zapLog.Error("Error creating 2FA record", zap.Error(err))
			return nil, fmt.Errorf("failed to save 2FA secret: %w", err)
		}
		zapLog.Debug("2FA record created successfully", zap.String("user_id", userID))
	} else if result.Error != nil {
		zapLog.Error("Error querying 2FA record", zap.Error(result.Error))
		return nil, fmt.Errorf("failed to query 2FA record: %w", result.Error)
	} else {
		twoFA.Secret = encryptedSecret // Update dengan encrypted secret
		twoFA.Enabled = false
		if err := database.GetDB().Save(&twoFA).Error; err != nil {
			zapLog.Error("Error updating 2FA record", zap.Error(err))
			return nil, fmt.Errorf("failed to update 2FA secret: %w", err)
		}
		zapLog.Debug("2FA record updated successfully", zap.String("user_id", userID))
	}

	// Generate gambar QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		zapLog.Error("Error generating QR code image", zap.Error(err))
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	if err := png.Encode(&buf, img); err != nil {
		zapLog.Error("Error encoding PNG", zap.Error(err))
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	zapLog.Debug("QR code generated successfully", zap.Int("size_bytes", buf.Len()))

	// Kembalikan secret dan QR code
	return map[string]interface{}{
		"secret":  key.Secret(),
		"qr_code": base64.StdEncoding.EncodeToString(buf.Bytes()),
		"url":     key.URL(),
		"message": "Scan QR code with authenticator app to enable 2FA",
	}, nil
}

// Verify2FAUseCase memverifikasi kode TOTP dan mengaktifkan 2FA
func Verify2FAUseCase(userID, code string) (map[string]interface{}, error) {
	// Ambil secret 2FA user
	var twoFA domain.TwoFactorAuth
	result := database.GetDB().Where("user_id = ?", userID).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("2FA not set up. Generate secret first")
	}

	// Decrypt secret (handle backward compatibility untuk data lama)
	secret, err := encryption.Decrypt(twoFA.Secret)
	if err != nil {
		// Jika decrypt gagal, gunakan as-is (backward compatibility)
		secret = twoFA.Secret
	}

	// Verifikasi kode TOTP
	valid := totp.Validate(code, secret)
	if !valid {
		return nil, fmt.Errorf("invalid verification code")
	}

	// Aktifkan 2FA dan generate backup codes
	backupCodesJSON := generateBackupCodes()
	
	// Encrypt backup codes sebelum disimpan
	encryptedBackupCodes, err := encryption.Encrypt(backupCodesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt backup codes: %w", err)
	}
	
	twoFA.Enabled = true
	twoFA.BackupCodes = encryptedBackupCodes // Simpan encrypted backup codes
	if err := database.GetDB().Save(&twoFA).Error; err != nil {
		return nil, fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Parse backup codes untuk dikembalikan sebagai array
	var backupCodesArray []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &backupCodesArray); err != nil {
		backupCodesArray = []string{}
	}

	return map[string]interface{}{
		"message":      "2FA enabled successfully",
		"backup_codes": backupCodesArray,
	}, nil
}

// Get2FAStatusUseCase mengembalikan status 2FA untuk user
func Get2FAStatusUseCase(userID string) (map[string]interface{}, error) {
	var twoFA domain.TwoFactorAuth
	result := database.GetDB().Where("user_id = ?", userID).First(&twoFA)

	if result.Error == gorm.ErrRecordNotFound {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	if result.Error != nil {
		logger.GetLogger().Error("Error querying 2FA status", zap.Error(result.Error))
		return nil, fmt.Errorf("failed to get 2FA status: %w", result.Error)
	}

	return map[string]interface{}{
		"enabled": twoFA.Enabled,
	}, nil
}

// Disable2FAUseCase menonaktifkan 2FA untuk user
func Disable2FAUseCase(userID string) error {
	// Cari record 2FA yang ada
	var twoFA domain.TwoFactorAuth
	result := database.GetDB().Where("user_id = ?", userID).First(&twoFA)

	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("2FA is not enabled for this user")
	}

	if result.Error != nil {
		logger.GetLogger().Error("Error querying 2FA", zap.Error(result.Error))
		return fmt.Errorf("failed to disable 2FA: %w", result.Error)
	}

	// Nonaktifkan 2FA
	twoFA.Enabled = false
	if err := database.GetDB().Save(&twoFA).Error; err != nil {
		logger.GetLogger().Error("Error disabling 2FA", zap.Error(err))
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// generateBackupCodes menghasilkan backup codes untuk 2FA
func generateBackupCodes() string {
	codes := make([]string, 10)
	for i := range codes {
		codes[i] = uuid.GenerateUUID()[:8]
	}
	jsonData, _ := json.Marshal(codes)
	return string(jsonData)
}

// verifyBackupCode memverifikasi apakah kode adalah backup code yang valid
// backupCodesJSON sudah di-decrypt sebelumnya (dari Verify2FALogin)
func verifyBackupCode(code, backupCodesJSON string) bool {
	if backupCodesJSON == "" {
		return false
	}

	var codes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &codes); err != nil {
		return false
	}

	for i, backupCode := range codes {
		if backupCode == code {
			// Hapus backup code yang sudah digunakan
			codes = append(codes[:i], codes[i+1:]...)
			// TODO: Update database dengan backup codes baru (perlu encrypt lagi)
			return true
		}
	}
	return false
}
