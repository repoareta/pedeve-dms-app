package main

import (
	"net/http"

	"github.com/go-chi/render"
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

// RequirePermission middleware checks if user has required permission
func RequirePermission(permission Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context
			userID := r.Context().Value(contextKeyUserID).(string)

			// Get user from database to check role
			var userModel UserModel
			result := DB.First(&userModel, "id = ?", userID)
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrorResponse{
					Error:   "unauthorized",
					Message: "User not found",
				})
				return
			}

			// Check permission
			if !HasPermission(userModel.Role, permission) {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, ErrorResponse{
					Error:   "forbidden",
					Message: "You don't have permission to access this resource",
				})
				return
			}

			// User has permission, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context
			userID := r.Context().Value(contextKeyUserID).(string)

			// Get user from database to check role
			var userModel UserModel
			result := DB.First(&userModel, "id = ?", userID)
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrorResponse{
					Error:   "unauthorized",
					Message: "User not found",
				})
				return
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
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, ErrorResponse{
					Error:   "forbidden",
					Message: "You don't have the required role to access this resource",
				})
				return
			}

			// User has required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

