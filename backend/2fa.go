package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// TwoFactorAuth merepresentasikan pengaturan 2FA untuk user
type TwoFactorAuth struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret      string    `gorm:"not null" json:"-"` // Secret TOTP
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	BackupCodes string    `gorm:"type:text" json:"-"` // Array JSON dari backup codes
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName menentukan nama tabel untuk TwoFactorAuth
func (TwoFactorAuth) TableName() string {
	return "two_factor_auths"
}

// Generate2FASecret menghasilkan secret TOTP baru untuk user
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
	// Ambil user dari context
	userIDValue := r.Context().Value(contextKeyUserID)
	usernameValue := r.Context().Value(contextKeyUsername)

	if userIDValue == nil || usernameValue == nil {
		log.Printf("Error: user context not found in request")
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		log.Printf("Error: invalid userID type in context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
		return
	}

	username, ok := usernameValue.(string)
	if !ok {
		log.Printf("Error: invalid username type in context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
		return
	}

	log.Printf("Generating 2FA secret for user: %s (ID: %s)", username, userID)

	// Generate kunci TOTP
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Pedeve Apps",
		AccountName: username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Printf("Error generating TOTP key: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate 2FA secret",
		})
		return
	}

	log.Printf("TOTP key generated successfully, secret: %s", key.Secret())

	// Simpan atau update record 2FA
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ?", userID).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		twoFA = TwoFactorAuth{
			ID:      GenerateUUID(),
			UserID:  userID,
			Secret:  key.Secret(),
			Enabled: false,
		}
		if err := DB.Create(&twoFA).Error; err != nil {
			log.Printf("Error creating 2FA record: %v", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to save 2FA secret",
			})
			return
		}
		log.Printf("2FA record created successfully")
	} else if result.Error != nil {
		log.Printf("Error querying 2FA record: %v", result.Error)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to query 2FA record",
		})
		return
	} else {
		twoFA.Secret = key.Secret()
		twoFA.Enabled = false
		if err := DB.Save(&twoFA).Error; err != nil {
			log.Printf("Error updating 2FA record: %v", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to update 2FA secret",
			})
			return
		}
		log.Printf("2FA record updated successfully")
	}

	// Generate gambar QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		log.Printf("Error generating QR code image: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate QR code",
		})
		return
	}

	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error encoding PNG: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to encode QR code",
		})
		return
	}

	log.Printf("QR code generated successfully, size: %d bytes", buf.Len())

	// Kembalikan secret dan QR code
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"secret":  key.Secret(),
		"qr_code": base64.StdEncoding.EncodeToString(buf.Bytes()),
		"url":     key.URL(),
		"message": "Scan QR code with authenticator app to enable 2FA",
	})
}

// Verify2FA memverifikasi kode TOTP dan mengaktifkan 2FA
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

	// Ambil secret 2FA user
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

	// Verifikasi kode TOTP
	valid := totp.Validate(req.Code, twoFA.Secret)
	if !valid {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_code",
			Message: "Invalid verification code",
		})
		return
	}

	// Aktifkan 2FA dan generate backup codes
	backupCodesJSON := generateBackupCodes()
	twoFA.Enabled = true
	twoFA.BackupCodes = backupCodesJSON
	DB.Save(&twoFA)

	// Parse backup codes untuk dikembalikan sebagai array
	var backupCodesArray []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &backupCodesArray); err != nil {
		backupCodesArray = []string{}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message":      "2FA enabled successfully",
		"backup_codes": backupCodesArray,
	})
}

// Verify2FALogin memverifikasi kode 2FA saat login
func Verify2FALogin(userID, code string) (bool, error) {
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ? AND enabled = ?", userID, true).First(&twoFA)
	if result.Error == gorm.ErrRecordNotFound {
		return false, fmt.Errorf("2FA not enabled")
	}

	// Coba kode TOTP terlebih dahulu
	valid := totp.Validate(code, twoFA.Secret)
	if valid {
		return true, nil
	}

	// Coba backup codes
	if verifyBackupCode(code, twoFA.BackupCodes) {
		return true, nil
	}

	return false, fmt.Errorf("invalid code")
}

// generateBackupCodes menghasilkan backup codes untuk 2FA
func generateBackupCodes() string {
	codes := make([]string, 10)
	for i := range codes {
		codes[i] = GenerateUUID()[:8]
	}
	jsonData, _ := json.Marshal(codes)
	return string(jsonData)
}

// verifyBackupCode memverifikasi apakah kode adalah backup code yang valid
func verifyBackupCode(code, backupCodesJSON string) bool {
	var codes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &codes); err != nil {
		return false
	}

	for i, backupCode := range codes {
		if backupCode == code {
			// Hapus backup code yang sudah digunakan
			codes = append(codes[:i], codes[i+1:]...)
			// TODO: Update database dengan backup codes baru
			// DB.Model(&TwoFactorAuth{}).Where("user_id = ?", userID).Update("backup_codes", json.Marshal(codes))
			return true
		}
	}
	return false
}

// Get2FAStatus mengembalikan status 2FA untuk user saat ini
// @Summary      Get 2FA status
// @Description  Get current user's 2FA status
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/2fa/status [get]
func Get2FAStatus(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value(contextKeyUserID)
	if userIDValue == nil {
		log.Printf("Error: user context not found in request")
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		log.Printf("Error: invalid userID type in context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
		return
	}

	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ?", userID).First(&twoFA)

	if result.Error == gorm.ErrRecordNotFound {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"enabled": false,
		})
		return
	}

	if result.Error != nil {
		log.Printf("Error querying 2FA status: %v", result.Error)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get 2FA status",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"enabled": twoFA.Enabled,
	})
}

// Disable2FA menonaktifkan 2FA untuk user saat ini
// @Summary      Disable 2FA
// @Description  Disable 2FA for current user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/v1/auth/2fa/disable [post]
func Disable2FA(w http.ResponseWriter, r *http.Request) {
	userIDValue := r.Context().Value(contextKeyUserID)
	usernameValue := r.Context().Value(contextKeyUsername)

	if userIDValue == nil || usernameValue == nil {
		log.Printf("Error: user context not found in request")
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "unauthorized",
			Message: "User context not found. Please ensure you are authenticated.",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		log.Printf("Error: invalid userID type in context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
		return
	}

	username, ok := usernameValue.(string)
	if !ok {
		log.Printf("Error: invalid username type in context")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user context",
		})
		return
	}

	// Ambil alamat IP dan user agent untuk audit log
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.UserAgent()

	// Cari record 2FA yang ada
	var twoFA TwoFactorAuth
	result := DB.Where("user_id = ?", userID).First(&twoFA)

	if result.Error == gorm.ErrRecordNotFound {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, ErrorResponse{
			Error:   "2fa_not_found",
			Message: "2FA is not enabled for this user",
		})
		return
	}

	if result.Error != nil {
		log.Printf("Error querying 2FA: %v", result.Error)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to disable 2FA",
		})
		return
	}

	// Nonaktifkan 2FA
	twoFA.Enabled = false
	if err := DB.Save(&twoFA).Error; err != nil {
		log.Printf("Error disabling 2FA: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to disable 2FA",
		})
		return
	}

	// Log aksi
	LogAction(userID, username, "disable_2fa", "auth", "", ipAddress, userAgent, "success", nil)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": "2FA has been disabled successfully",
	})
}
