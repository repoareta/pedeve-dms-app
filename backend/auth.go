package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Register handles user registration
// @Summary      Register new user
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest  true  "User registration data"
// @Success      201   {object}  AuthResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /api/v1/auth/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate and sanitize input
	if err := ValidateRegisterInput(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Cek apakah user sudah ada
	var existingUser UserModel
	result := DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser)
	if result.Error == nil {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, ErrorResponse{
			Error:   "user_exists",
			Message: "Username or email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to hash password",
		})
		return
	}

	// Buat user baru
	now := time.Now()
	userModel := &UserModel{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Simpan ke database
	if err := DB.Create(userModel).Error; err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create user",
		})
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(userModel.ID, userModel.Username)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	// Set cookie aman dengan token JWT (httpOnly cookie untuk keamanan yang lebih baik)
	SetSecureCookie(w, authTokenCookie, token)

	// Kembalikan response
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, AuthResponse{
		Token: token,
		User: User{
			ID:        userModel.ID,
			Username:  userModel.Username,
			Email:     userModel.Email,
			Role:      userModel.Role,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		},
	})
}

// Login handles user login
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Login credentials"
// @Success      200          {object}  AuthResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /api/v1/auth/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate and sanitize input
	if err := ValidateLoginInput(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Ambil IP client dan User Agent untuk audit logging
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Cari user di database (bisa login dengan username atau email)
	var userModel UserModel
	result := DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&userModel)
	if result.Error == gorm.ErrRecordNotFound {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
		return
	}

	// Cek password
	if !CheckPasswordHash(req.Password, userModel.Password) {
		// Log percobaan login yang gagal
		LogAction(userModel.ID, userModel.Username, ActionFailedLogin, ResourceAuth, "", ipAddress, userAgent, StatusFailure, map[string]interface{}{
			"reason": "invalid_password",
		})

		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
		return
	}

	// Cek apakah 2FA diaktifkan
	var twoFA TwoFactorAuth
	result2FA := DB.Where("user_id = ? AND enabled = ?", userModel.ID, true).First(&twoFA)
	is2FAEnabled := result2FA.Error == nil

	// Jika 2FA diaktifkan, verifikasi kode
	if is2FAEnabled {
		if req.Code == "" {
			// 2FA diperlukan tapi kode tidak diberikan
			render.Status(r, http.StatusOK)
			render.JSON(w, r, map[string]interface{}{
				"requires_2fa": true,
				"message":      "2FA verification required. Please enter your 2FA code.",
			})
			return
		}

		// Verifikasi kode 2FA
		valid, err := Verify2FALogin(userModel.ID, req.Code)
		if err != nil || !valid {
			// Log percobaan 2FA yang gagal
			LogAction(userModel.ID, userModel.Username, ActionFailedLogin, ResourceAuth, "", ipAddress, userAgent, StatusFailure, map[string]interface{}{
				"reason": "invalid_2fa_code",
			})

			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{
				Error:   "invalid_2fa_code",
				Message: "Invalid 2FA code",
			})
			return
		}
	}

	// Log login yang berhasil
	LogAction(userModel.ID, userModel.Username, ActionLogin, ResourceAuth, "", ipAddress, userAgent, StatusSuccess, nil)

	// Generate JWT token
	token, err := GenerateJWT(userModel.ID, userModel.Username)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	// Set cookie aman dengan token JWT (httpOnly cookie untuk keamanan yang lebih baik)
	SetSecureCookie(w, authTokenCookie, token)

	// Kembalikan response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, AuthResponse{
		Token: token,
		User: User{
			ID:        userModel.ID,
			Username:  userModel.Username,
			Email:     userModel.Email,
			Role:      userModel.Role,
			CreatedAt: userModel.CreatedAt,
			UpdatedAt: userModel.UpdatedAt,
		},
	})
}

// GetProfile returns current user profile
// @Summary      Get user profile
// @Description  Get current authenticated user profile
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  User
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/profile [get]
func GetProfile(w http.ResponseWriter, r *http.Request) {
	// Ambil user dari context (diset oleh JWT middleware)
	userID := r.Context().Value(contextKeyUserID).(string)
	_ = r.Context().Value(contextKeyUsername).(string) // Tersedia tapi tidak diperlukan untuk lookup

	// Cari user di database
	var userModel UserModel
	result := DB.First(&userModel, "id = ?", userID)
	if result.Error == gorm.ErrRecordNotFound {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
		return
	}

	// Return user (tanpa password)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, User{
		ID:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		Role:      userModel.Role,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	})
}

// Logout handles user logout
// @Summary      User logout
// @Description  Logout user and clear authentication cookie
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  ErrorResponse
// @Router       /api/v1/auth/logout [post]
func Logout(w http.ResponseWriter, r *http.Request) {
	// Ambil info user dari context untuk audit logging
	userIDValue := r.Context().Value(contextKeyUserID)
	usernameValue := r.Context().Value(contextKeyUsername)

	if userIDValue != nil && usernameValue != nil {
		userID := userIDValue.(string)
		username := usernameValue.(string)

		// Ambil alamat IP dan user agent untuk audit log
		ipAddress := getClientIP(r)
		userAgent := r.UserAgent()

		// Log aksi logout
		LogAction(userID, username, ActionLogout, ResourceAuth, "", ipAddress, userAgent, StatusSuccess, nil)
	}

	// Hapus cookie aman
	DeleteSecureCookie(w, authTokenCookie)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"message": "Logged out successfully",
	})
}
