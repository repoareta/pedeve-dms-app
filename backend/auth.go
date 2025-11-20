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

	// Check if user already exists
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

	// Create user
	now := time.Now()
	userModel := &UserModel{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to database
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

	// Return response
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

	// Get client IP and User Agent for audit logging
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Find user in database (can login with username or email)
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

	// Check password
	if !CheckPasswordHash(req.Password, userModel.Password) {
		// Log failed login attempt
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

	// TODO: Check if 2FA is enabled and require 2FA code if enabled
	// For now, we skip 2FA check but it's ready to be integrated

	// Log successful login
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

	// Return response
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
	// Get user from context (set by JWT middleware)
	userID := r.Context().Value(contextKeyUserID).(string)
	_ = r.Context().Value(contextKeyUsername).(string) // Available but not needed for lookup

	// Find user in database
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

	// Return user (without password)
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
