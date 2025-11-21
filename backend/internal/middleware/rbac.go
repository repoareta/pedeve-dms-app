package middleware

import (
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Permission represents a permission in the system
type Permission string

const (
	// User permissions
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserDelete Permission = "user:delete"

	// Document permissions
	PermissionDocumentRead   Permission = "document:read"
	PermissionDocumentWrite  Permission = "document:write"
	PermissionDocumentDelete Permission = "document:delete"

	// Admin permissions
	PermissionAdminRead   Permission = "admin:read"
	PermissionAdminWrite  Permission = "admin:write"
	PermissionAdminDelete Permission = "admin:delete"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]Permission{
	"user": {
		PermissionUserRead,
		PermissionDocumentRead,
		PermissionDocumentWrite,
	},
	"admin": {
		PermissionUserRead,
		PermissionUserWrite,
		PermissionDocumentRead,
		PermissionDocumentWrite,
		PermissionDocumentDelete,
		PermissionAdminRead,
	},
	"superadmin": {
		// Superadmin has all permissions
		PermissionUserRead,
		PermissionUserWrite,
		PermissionUserDelete,
		PermissionDocumentRead,
		PermissionDocumentWrite,
		PermissionDocumentDelete,
		PermissionAdminRead,
		PermissionAdminWrite,
		PermissionAdminDelete,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role string, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	// Superadmin has all permissions
	if role == "superadmin" {
		return true
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// RequirePermission middleware checks if user has required permission (untuk Fiber)
func RequirePermission(permission Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from locals
		userIDVal := c.Locals("userID")
		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User context not found",
			})
		}
		userID := userIDVal.(string)

		// Get user from database to check role
		var userModel domain.UserModel
		result := database.GetDB().First(&userModel, "id = ?", userID)
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not found",
			})
		}

		// Check permission
		if !HasPermission(userModel.Role, permission) {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to access this resource",
			})
		}

		// User has permission, continue
		return c.Next()
	}
}

// RequireRole middleware checks if user has required role (untuk Fiber)
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from locals
		userIDVal := c.Locals("userID")
		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User context not found",
			})
		}
		userID := userIDVal.(string)

		// Get user from database to check role
		var userModel domain.UserModel
		result := database.GetDB().First(&userModel, "id = ?", userID)
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not found",
			})
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if userModel.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have the required role to access this resource",
			})
		}

		// User has required role, continue
		return c.Next()
	}
}

