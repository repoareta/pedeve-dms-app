package http

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/audit"
	"github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// UserManagementHandler handles user management HTTP requests
type UserManagementHandler struct {
	userUseCase usecase.UserManagementUseCase
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(userUseCase usecase.UserManagementUseCase) *UserManagementHandler {
	return &UserManagementHandler{
		userUseCase: userUseCase,
	}
}

// CreateUser handles user creation
// @Summary      Buat User Baru
// @Description  Membuat user baru. Admin hanya bisa membuat user di company mereka atau descendants. Superadmin role tidak bisa dibuat dari antarmuka ini.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      object  true  "User data"
// @Success      201   {object}  domain.UserModel
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      403   {object}  domain.ErrorResponse
// @Router       /api/v1/users [post]
func (h *UserManagementHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Username  string  `json:"username" validate:"required"`
		Email     string  `json:"email" validate:"required,email"`
		Password  string  `json:"password" validate:"required,min=8"`
		CompanyID *string `json:"company_id"`
		RoleID    *string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Prevent creating superadmin user
	if req.RoleID != nil {
		// Check if role is superadmin
		roleUseCase := usecase.NewRoleManagementUseCase()
		role, err := roleUseCase.GetRoleByID(*req.RoleID)
		if err == nil && role != nil && role.Name == "superadmin" {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Superadmin role cannot be assigned through this interface. Superadmin is a system account managed separately.",
			})
		}
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Superadmin can create user in any company or without company (standby)
	// Admin can only create user in their company or descendants (RBAC)
	if roleName != "superadmin" && companyID != nil {
		// Get user's company ID safely (handle both *string and string)
		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}
		
		// If company_id is provided, validate access
		if req.CompanyID != nil {
			companyUseCase := usecase.NewCompanyUseCase()
			hasAccess, err := companyUseCase.ValidateCompanyAccess(userCompanyID, *req.CompanyID)
			if err != nil || !hasAccess {
				return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
					Error:   "forbidden",
					Message: "You can only create users in your company or its descendants",
				})
			}
		}
		// Note: Non-superadmin can create user without company (standby mode)
		// They just won't be auto-assigned to any company
	}

	user, err := h.userUseCase.CreateUser(req.Username, req.Email, req.Password, req.CompanyID, req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	audit.LogAction(userID, username, audit.ActionCreateUser, audit.ResourceUser, user.ID, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)
	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUser handles getting user by ID
// @Summary      Ambil User by ID
// @Description  Mengambil informasi user berdasarkan ID.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  domain.UserModel
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id} [get]
func (h *UserManagementHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID := c.Locals("companyID")
	roleName := c.Locals("roleName").(string)

	// Check access
	if roleName != "superadmin" && companyID != nil {
		userCompanyID := companyID.(string)
		hasAccess, err := h.userUseCase.ValidateUserAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this user",
			})
		}
	}

	user, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// GetAllUsers handles getting all users
// @Summary      Ambil Semua Users
// @Description  Mengambil daftar semua users. Filtered berdasarkan company access (RBAC). Superadmin melihat semua users. Non-superadmin hanya melihat users di company mereka dan descendants.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.UserModel
// @Router       /api/v1/users [get]
func (h *UserManagementHandler) GetAllUsers(c *fiber.Ctx) error {
	// Get roleName and companyID from JWT
	roleNameVal := c.Locals("roleName")
	companyIDVal := c.Locals("companyID")
	
	roleName := ""
	if roleNameVal != nil {
		if rn, ok := roleNameVal.(string); ok {
			roleName = rn
		}
	}
	
	// Superadmin sees all users (except other superadmins for security)
	if roleName == "superadmin" {
		users, err := h.userUseCase.GetAllUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get users",
			})
		}
		
		// Filter out superadmin users for security
		filtered := []domain.UserModel{}
		for _, user := range users {
			if user.Role != "superadmin" {
				filtered = append(filtered, user)
			}
		}
		return c.Status(fiber.StatusOK).JSON(filtered)
	}
	
	// Non-superadmin: filter by company hierarchy
	if companyIDVal == nil {
		// User has no company, return empty list
		return c.Status(fiber.StatusOK).JSON([]domain.UserModel{})
	}
	
	// Get user's company ID safely
	var userCompanyID string
	if companyIDPtr, ok := companyIDVal.(*string); ok && companyIDPtr != nil {
		userCompanyID = *companyIDPtr
	} else if companyIDStr, ok := companyIDVal.(string); ok {
		userCompanyID = companyIDStr
	} else {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Invalid company ID format",
		})
	}
	
	// Get users from user's company and all descendants (RBAC)
	users, err := h.userUseCase.GetUsersByCompanyHierarchy(userCompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get users: " + err.Error(),
		})
	}
	
	// Filter out superadmin users for security
	filtered := []domain.UserModel{}
	for _, user := range users {
		if user.Role != "superadmin" {
			filtered = append(filtered, user)
		}
	}
	
	return c.Status(fiber.StatusOK).JSON(filtered)
}

// UpdateUser handles user update
// @Summary      Update User
// @Description  Mengupdate informasi user. Superadmin tidak bisa mengedit dirinya sendiri.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "User ID"
// @Param        user  body      object  true  "User data to update"
// @Success      200   {object}  domain.UserModel
// @Failure      400   {object}  domain.ErrorResponse
// @Failure      403   {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserManagementHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	currentUserID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	// Prevent superadmin from editing themselves
	if roleName == "superadmin" && id == currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin cannot edit their own account. Please use Vault or system administrator for account changes.",
		})
	}

	// Get target user to check if they are superadmin
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Prevent editing superadmin user (even by other superadmins)
	if targetUser.Role == "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin account cannot be edited through this interface. Please use Vault or system administrator.",
		})
	}

	var req struct {
		Username  string  `json:"username"`
		Email     string  `json:"email"`
		CompanyID *string `json:"company_id"` // nil = no change, empty string "" = unassign, non-empty = assign
		RoleID    *string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Safe type assertion for companyID (can be *string or nil)
	var userCompanyID string
	companyIDVal := c.Locals("companyID")
	if companyIDVal != nil {
		switch v := companyIDVal.(type) {
		case string:
			userCompanyID = v
		case *string:
			if v != nil {
				userCompanyID = *v
			}
		}
	}

	// Check access
	if roleName != "superadmin" && userCompanyID != "" {
		hasAccess, err := h.userUseCase.ValidateUserAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to update this user",
			})
		}
	}

	user, err := h.userUseCase.UpdateUser(id, req.Username, req.Email, req.CompanyID, req.RoleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionUpdateUser, audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(user)
}

// DeleteUser handles user deletion
// @Summary      Hapus User
// @Description  Menghapus user. Superadmin tidak bisa menghapus dirinya sendiri atau user superadmin lainnya.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id} [delete]
func (h *UserManagementHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	currentUserID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	// Prevent superadmin from deleting themselves
	if roleName == "superadmin" && id == currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin cannot delete their own account.",
		})
	}

	// Get target user to check if they are superadmin
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Prevent deleting superadmin user
	if targetUser.Role == "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin account cannot be deleted through this interface.",
		})
	}

	companyID := c.Locals("companyID")

	// Check access
	if roleName != "superadmin" && companyID != nil {
		userCompanyID := companyID.(string)
		hasAccess, err := h.userUseCase.ValidateUserAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to delete this user",
			})
		}
	}

	if err := h.userUseCase.DeleteUser(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, audit.ActionDeleteUser, audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, nil)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// ToggleUserStatus handles toggling user active status
// @Summary      Toggle Status User
// @Description  Mengaktifkan atau menonaktifkan user. Superadmin tidak bisa menonaktifkan dirinya sendiri atau user superadmin lainnya.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  domain.UserModel
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id}/toggle-status [patch]
func (h *UserManagementHandler) ToggleUserStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	currentUserID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	// Prevent superadmin from deactivating themselves
	if roleName == "superadmin" && id == currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin cannot deactivate their own account.",
		})
	}

	// Get target user to check if they are superadmin
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Prevent deactivating superadmin user (if trying to deactivate active superadmin)
	if targetUser.Role == "superadmin" && targetUser.IsActive {
		// Allow activating superadmin if inactive, but not deactivating if active
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin account cannot be deactivated through this interface.",
		})
	}

	companyID := c.Locals("companyID")

	// Check access
	if roleName != "superadmin" && companyID != nil {
		userCompanyID := companyID.(string)
		hasAccess, err := h.userUseCase.ValidateUserAccess(userCompanyID, id)
		if err != nil || !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to modify this user",
			})
		}
	}

	user, err := h.userUseCase.ToggleUserStatus(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "toggle_failed",
			Message: err.Error(),
		})
	}

	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	var action string
	if user.IsActive {
		action = "activate_user"
	} else {
		action = "deactivate_user"
	}
	audit.LogAction(userID, username, action, audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"is_active": user.IsActive,
	})

	return c.Status(fiber.StatusOK).JSON(user)
}

// ResetUserPassword handles password reset for users (superadmin only)
// @Summary      Reset Password User
// @Description  Reset password untuk user. Hanya superadmin yang bisa melakukan reset password. Superadmin tidak bisa reset password dirinya sendiri.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Param        request  body      object  true  "New password"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      403  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id}/reset-password [post]
func (h *UserManagementHandler) ResetUserPassword(c *fiber.Ctx) error {
	id := c.Params("id")
	currentUserID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	// Only superadmin can reset passwords
	if roleName != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Only superadmin can reset user passwords",
		})
	}

	// Prevent superadmin from resetting their own password
	if id == currentUserID {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin cannot reset their own password through this interface",
		})
	}

	var req struct {
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate password
	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "validation_error",
			Message: "Password must be at least 8 characters long",
		})
	}

	// Get target user to verify they exist
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Reset password
	if err := h.userUseCase.ResetUserPassword(id, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
	}

	// Log action
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "reset_user_password", audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"target_username": targetUser.Username,
		"target_email":     targetUser.Email,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password reset successfully",
		"user_id": id,
	})
}

// AssignUserToCompany handles assigning user to a company
// @Summary      Assign User ke Company
// @Description  Mengassign user ke company tertentu. Superadmin bisa assign ke semua company, admin bisa assign ke company mereka sendiri.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "User ID"
// @Param        request    body      object  true  "Company ID and Role ID"
// @Success      200        {object}  domain.UserModel
// @Failure      400        {object}  domain.ErrorResponse
// @Failure      403        {object}  domain.ErrorResponse
// @Failure      404        {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id}/assign-company [post]
func (h *UserManagementHandler) AssignUserToCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	roleName := c.Locals("roleName").(string)
	companyID := c.Locals("companyID")

	var req struct {
		CompanyID string  `json:"company_id" validate:"required"`
		RoleID    *string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Superadmin can assign to any company
	// Admin can only assign to their own company
	if roleName != "superadmin" {
		// Check if user has company
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Only superadmin and admin can assign users to companies",
			})
		}

		// Get user's company ID
		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// Admin can only assign to their own company
		if req.CompanyID != userCompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin can only assign users to their own company",
			})
		}

		// Only admin role can assign (not manager or staff)
		if roleName != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Only superadmin and admin can assign users to companies",
			})
		}
	}

	// Get target user to verify they exist
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Prevent assigning superadmin user
	if targetUser.Role == "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin account cannot be assigned to a company",
		})
	}

	// Assign user to company (now supports multiple company assignments via junction table)
	if err := h.userUseCase.AssignUserToCompany(id, req.CompanyID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "assign_failed",
			Message: err.Error(),
		})
	}

	// If role_id is provided, assign role in this specific company via junction table
	// This allows same user to have different roles in different companies
	if req.RoleID != nil {
		if err := h.userUseCase.AssignUserToRoleInCompany(id, req.CompanyID, *req.RoleID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
				Error:   "assign_role_failed",
				Message: err.Error(),
			})
		}
	}

	// Get updated user
	updatedUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get updated user",
		})
	}

	// Log action
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "assign_user_to_company", audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"target_username": targetUser.Username,
		"company_id":       req.CompanyID,
		"role_id":         req.RoleID,
	})

	return c.Status(fiber.StatusOK).JSON(updatedUser)
}

// UnassignUserFromCompany handles removing user from a company
// @Summary      Unassign User dari Company
// @Description  Menghapus assignment user dari company tertentu. User bisa tetap di-assign ke company lain.
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "User ID"
// @Param        request    body      object  true  "Company ID"
// @Success      200        {object}  domain.UserModel
// @Failure      400        {object}  domain.ErrorResponse
// @Failure      403        {object}  domain.ErrorResponse
// @Failure      404        {object}  domain.ErrorResponse
// @Router       /api/v1/users/{id}/unassign-company [post]
func (h *UserManagementHandler) UnassignUserFromCompany(c *fiber.Ctx) error {
	id := c.Params("id")
	roleName := c.Locals("roleName").(string)
	companyID := c.Locals("companyID")

	var req struct {
		CompanyID string `json:"company_id" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Superadmin can unassign from any company
	// Admin can only unassign from their own company
	if roleName != "superadmin" {
		// Check if user has company
		if companyID == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Only superadmin and admin can unassign users from companies",
			})
		}

		// Get user's company ID
		var userCompanyID string
		if companyIDPtr, ok := companyID.(*string); ok && companyIDPtr != nil {
			userCompanyID = *companyIDPtr
		} else if companyIDStr, ok := companyID.(string); ok {
			userCompanyID = companyIDStr
		} else {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid company ID format",
			})
		}

		// Admin can only unassign from their own company
		if req.CompanyID != userCompanyID {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin can only unassign users from their own company",
			})
		}

		// Only admin role can unassign (not manager or staff)
		if roleName != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "Only superadmin and admin can unassign users from companies",
			})
		}
	}

	// Get target user to verify they exist
	targetUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(domain.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
	}

	// Prevent unassigning superadmin user
	if targetUser.Role == "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Superadmin account cannot be unassigned from company",
		})
	}

	// Unassign user from company via junction table
	if err := h.userUseCase.UnassignUserFromCompany(id, req.CompanyID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{
			Error:   "unassign_failed",
			Message: err.Error(),
		})
	}

	// Get updated user
	updatedUser, err := h.userUseCase.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get updated user",
		})
	}

	// Log action
	userID := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	audit.LogAction(userID, username, "unassign_user_from_company", audit.ResourceUser, id, getClientIP(c), c.Get("User-Agent"), audit.StatusSuccess, map[string]interface{}{
		"target_username": targetUser.Username,
		"company_id":       req.CompanyID,
	})

	return c.Status(fiber.StatusOK).JSON(updatedUser)
}

// GetMyCompanies handles getting all companies assigned to the current user
// @Summary      Ambil Companies User Saat Ini
// @Description  Mengambil daftar semua companies yang di-assign ke user yang sedang login (dari junction table, support multiple assignments).
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.UserCompanyResponse
// @Failure      401  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /api/v1/users/me/companies [get]
func (h *UserManagementHandler) GetMyCompanies(c *fiber.Ctx) error {
	// Get user ID from JWT token (stored in locals by middleware)
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in token",
		})
	}

	// Get all companies assigned to this user
	companies, err := h.userUseCase.GetUserCompanies(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user companies: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(companies)
}

