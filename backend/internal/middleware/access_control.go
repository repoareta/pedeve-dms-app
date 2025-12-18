package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"github.com/repoareta/pedeve-dms-app/backend/internal/utils"
	"go.uber.org/zap"
)

// RequireCompanyAccess middleware memastikan user hanya bisa mengakses company mereka atau descendants
// Ini adalah Row-Level Security (RLS) untuk company hierarchy
func RequireCompanyAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		zapLog := logger.GetLogger()

		// Get user info from JWT claims
		userIDVal := c.Locals("userID")
		companyIDVal := c.Locals("companyID")
		roleNameVal := c.Locals("roleName")

		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User context not found",
			})
		}

		roleName := ""
		if roleNameVal != nil {
			roleName = roleNameVal.(string)
		}

		// Superadmin/administrator bisa akses semua company
		if utils.IsSuperAdminLike(roleName) {
			return c.Next()
		}

		// Get target company ID from request (bisa dari param, query, atau body)
		targetCompanyID := c.Params("company_id")
		if targetCompanyID == "" {
			targetCompanyID = c.Query("company_id")
		}

		// Jika tidak ada target company ID, skip check (untuk endpoints yang tidak spesifik company)
		if targetCompanyID == "" {
			return c.Next()
		}

		// Get user's company ID
		if companyIDVal == nil {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "User is not associated with any company",
			})
		}

		userCompanyID := companyIDVal.(string)

		// Jika target company sama dengan user's company, allow
		if targetCompanyID == userCompanyID {
			return c.Next()
		}

		// Check if target company is a descendant of user's company
		companyRepo := repository.NewCompanyRepository()
		isDescendant, err := companyRepo.IsDescendantOf(targetCompanyID, userCompanyID)
		if err != nil {
			zapLog.Error("Failed to check company hierarchy", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to verify company access",
			})
		}

		if !isDescendant {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have access to this company",
			})
		}

		// User has access, continue
		return c.Next()
	}
}

// RequirePermissionFromJWT middleware checks permission dari JWT claims (tidak perlu query database)
func RequirePermissionFromJWT(permissionName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get permissions from JWT claims
		permissionsVal := c.Locals("permissions")
		roleNameVal := c.Locals("roleName")

		if permissionsVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User context not found",
			})
		}

		permissions, ok := permissionsVal.([]string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid permissions in token",
			})
		}

		// Superadmin has all permissions (check for "*" permission)
		roleName := ""
		if roleNameVal != nil {
			roleName = roleNameVal.(string)
		}

		if utils.IsSuperAdminLike(roleName) {
			return c.Next()
		}

		// Check if user has the required permission
		hasPermission := false
		for _, perm := range permissions {
			if perm == "*" || perm == permissionName {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to access this resource",
			})
		}

		// User has permission, continue
		return c.Next()
	}
}

// RequireRoleFromJWT middleware checks role dari JWT claims (tidak perlu query database)
func RequireRoleFromJWT(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get role from JWT claims
		roleNameVal := c.Locals("roleName")

		if roleNameVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error:   "unauthorized",
				Message: "User context not found",
			})
		}

		roleName := roleNameVal.(string)

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range roles {
			if roleName == role {
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
