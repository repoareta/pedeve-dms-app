package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// TwoFactorAuth represents 2FA settings for a user
type TwoFactorAuth struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	UserID       string    `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret       string    `gorm:"not null" json:"-"` // TOTP secret
	Enabled      bool      `gorm:"default:false" json:"enabled"`
	BackupCodes  string    `gorm:"type:text" json:"-"` // JSON array of backup codes
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for TwoFactorAuth
func (TwoFactorAuth) TableName() string {
	return "two_factor_auths"
}

// Generate2FASecret generates a new TOTP secret for a user
// @Summary      Generate 2FA secret
// @Description  Generate a new TOTP secret and QR code for 2FA setup
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/2fa/generate [post]
func Generate2FASecret(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID := r.Context().Value(contextKeyUserID).(string)
	username := r.Context().Value(contextKeyUsername).(string)

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "DMS App",
		AccountName: username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate 2FA secret",
		})
		return
	}

	// Save or update 2FA record
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ?", userID).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		twoFA = TwoFactorAuth{
			ID:      GenerateUUID(),
			UserID:  userID,
			Secret:  key.Secret(),
			Enabled: false,
		}
		DB.Create(&twoFA)
	} else {
		twoFA.Secret = key.Secret()
		twoFA.Enabled = false
		DB.Save(&twoFA)
	}

	// Generate QR code image
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate QR code",
		})
		return
	}
	png.Encode(&buf, img)

	// Return secret and QR code
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"secret":    key.Secret(),
		"qr_code":   base64.StdEncoding.EncodeToString(buf.Bytes()),
		"url":       key.URL(),
		"message":   "Scan QR code with authenticator app to enable 2FA",
	})
}

// Verify2FA verifies a TOTP code and enables 2FA
// @Summary      Verify and enable 2FA
// @Description  Verify TOTP code and enable 2FA for user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code  body      map[string]string  true  "TOTP code"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /api/v1/auth/2fa/verify [post]
func Verify2FA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKeyUserID).(string)

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get user's 2FA secret
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ?", userID).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "not_found",
			Message: "2FA not set up. Generate secret first",
		})
		return
	}

	// Verify TOTP code
	valid := totp.Validate(req.Code, twoFA.Secret)
	if !valid {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_code",
			Message: "Invalid verification code",
		})
		return
	}

	// Enable 2FA and generate backup codes
	backupCodes := generateBackupCodes()
	twoFA.Enabled = true
	twoFA.BackupCodes = backupCodes
	DB.Save(&twoFA)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message":     "2FA enabled successfully",
		"backup_codes": backupCodes,
	})
}

// Verify2FALogin verifies 2FA code during login
func Verify2FALogin(userID, code string) (bool, error) {
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		return false, fmt.Errorf("2FA not enabled")
	}

	// Try TOTP code first
	valid := totp.Validate(code, twoFA.Secret)
	if valid {
		return true, nil
	}

	// Try backup codes
	if verifyBackupCode(code, twoFA.BackupCodes) {
		return true, nil
	}

	return false, fmt.Errorf("invalid code")
}

// generateBackupCodes generates backup codes for 2FA
func generateBackupCodes() string {
	codes := make([]string, 10)
	for i := range codes {
		codes[i] = GenerateUUID()[:8]
	}
	jsonData, _ := json.Marshal(codes)
	return string(jsonData)
}

// verifyBackupCode verifies if a code is a valid backup code
func verifyBackupCode(code, backupCodesJSON string) bool {
	var codes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &codes); err != nil {
		return false
	}

		for i, backupCode := range codes {
			if backupCode == code {
				// Remove used backup code
				codes = append(codes[:i], codes[i+1:]...)
				// TODO: Update database with new backup codes
				// DB.Model(&TwoFactorAuth{}).Where("user_id = ?", userID).Update("backup_codes", json.Marshal(codes))
				return true
			}
		}
	return false
}

