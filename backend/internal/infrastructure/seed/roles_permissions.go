package seed

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SeedRolesAndPermissions seeds default roles and permissions
func SeedRolesAndPermissions() {
	zapLog := logger.GetLogger()
	db := database.GetDB()

	// Check if roles already exist
	var existingRole domain.RoleModel
	if err := db.Where("name = ?", "superadmin").First(&existingRole).Error; err == nil {
		// Past seeding sudah jalan; pastikan role baru (administrator) ada
		zapLog.Info("Roles already seeded, ensuring administrator role exists...")
		ensureAdministratorRole(db, zapLog)
		return
	}

	zapLog.Info("Seeding roles and permissions...")

	// Create Permissions
	permissions := []domain.PermissionModel{
		// Global permissions (superadmin only)
		{ID: uuid.GenerateUUID(), Name: "global:*", Description: "All global permissions", Resource: "global", Action: "*", Scope: domain.ScopeGlobal},

		// Company management permissions
		{ID: uuid.GenerateUUID(), Name: "company:view", Description: "View company information", Resource: "company", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "company:create", Description: "Create new company", Resource: "company", Action: "create", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "company:update", Description: "Update company information", Resource: "company", Action: "update", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "company:delete", Description: "Delete company", Resource: "company", Action: "delete", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "company:manage", Description: "Manage company (all company operations)", Resource: "company", Action: "manage", Scope: domain.ScopeCompany},

		// User management permissions
		{ID: uuid.GenerateUUID(), Name: "user:view", Description: "View user information", Resource: "user", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "user:create", Description: "Create new user", Resource: "user", Action: "create", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "user:update", Description: "Update user information", Resource: "user", Action: "update", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "user:delete", Description: "Delete user", Resource: "user", Action: "delete", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "user:manage", Description: "Manage users (all user operations)", Resource: "user", Action: "manage", Scope: domain.ScopeCompany},

		// Document management permissions
		{ID: uuid.GenerateUUID(), Name: "document:view", Description: "View documents", Resource: "document", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "document:create", Description: "Create new document", Resource: "document", Action: "create", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "document:update", Description: "Update document", Resource: "document", Action: "update", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "document:delete", Description: "Delete document", Resource: "document", Action: "delete", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "document:manage", Description: "Manage documents (all document operations)", Resource: "document", Action: "manage", Scope: domain.ScopeCompany},

		// Dashboard permissions
		{ID: uuid.GenerateUUID(), Name: "dashboard:view", Description: "View dashboard", Resource: "dashboard", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "dashboard:view_all", Description: "View all companies dashboard (superadmin/admin only)", Resource: "dashboard", Action: "view_all", Scope: domain.ScopeGlobal},

		// Report permissions
		{ID: uuid.GenerateUUID(), Name: "report:view", Description: "View reports", Resource: "report", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "report:generate", Description: "Generate reports", Resource: "report", Action: "generate", Scope: domain.ScopeCompany},

		// Audit log permissions
		{ID: uuid.GenerateUUID(), Name: "audit:view", Description: "View audit logs", Resource: "audit", Action: "view", Scope: domain.ScopeCompany},
		{ID: uuid.GenerateUUID(), Name: "audit:view_all", Description: "View all audit logs (superadmin only)", Resource: "audit", Action: "view_all", Scope: domain.ScopeGlobal},
	}

	// Create permissions
	for _, perm := range permissions {
		if err := db.Create(&perm).Error; err != nil {
			zapLog.Warn("Failed to create permission", zap.String("name", perm.Name), zap.Error(err))
		}
	}

	// Create Roles
	roles := []domain.RoleModel{
		{
			ID:          uuid.GenerateUUID(),
			Name:        "superadmin",
			Description: "Super Administrator - Full system access",
			Level:       0,
			IsSystem:    true,
		},
		{
			ID:          uuid.GenerateUUID(),
			Name:        "administrator",
			Description: "Administrator - Full access (kecuali fitur development)",
			Level:       0,
			IsSystem:    true,
		},
		{
			ID:          uuid.GenerateUUID(),
			Name:        "admin",
			Description: "Administrator - Company-level admin access",
			Level:       1,
			IsSystem:    true,
		},
		{
			ID:          uuid.GenerateUUID(),
			Name:        "manager",
			Description: "Manager - Department/team management access",
			Level:       2,
			IsSystem:    true,
		},
		{
			ID:          uuid.GenerateUUID(),
			Name:        "staff",
			Description: "Staff - Basic user access",
			Level:       3,
			IsSystem:    true,
		},
	}

	// Create roles and assign permissions
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			zapLog.Warn("Failed to create role", zap.String("name", role.Name), zap.Error(err))
			continue
		}

		// Assign permissions based on role
		var permissionNames []string
		switch role.Name {
		case "superadmin":
			// Superadmin gets all permissions
			permissionNames = []string{"global:*", "company:manage", "user:manage", "document:manage", "dashboard:view_all", "report:generate", "audit:view_all"}
		case "administrator":
			// Administrator sama dengan superadmin, kecuali akses fitur development dibatasi di handler
			permissionNames = []string{"global:*", "company:manage", "user:manage", "document:manage", "dashboard:view_all", "report:generate", "audit:view_all"}
		case "admin":
			// Admin gets company-level management permissions
			permissionNames = []string{"company:manage", "user:manage", "document:manage", "dashboard:view", "report:generate", "audit:view"}
		case "manager":
			// Manager gets view and limited management permissions
			permissionNames = []string{"company:view", "user:view", "document:view", "document:create", "document:update", "dashboard:view", "report:view"}
		case "staff":
			// Staff gets basic view permissions
			permissionNames = []string{"document:view", "dashboard:view", "report:view"}
		}

		// Assign permissions to role
		for _, permName := range permissionNames {
			var perm domain.PermissionModel
			if err := db.Where("name = ?", permName).First(&perm).Error; err == nil {
				// Create role_permission relationship
				rolePerm := domain.RolePermissionModel{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rolePerm).Error; err != nil {
					zapLog.Warn("Failed to assign permission to role",
						zap.String("role", role.Name),
						zap.String("permission", permName),
						zap.Error(err),
					)
				}
			}
		}
	}

	zapLog.Info("Roles and permissions seeded successfully")
}

// ensureAdministratorRole memastikan role administrator ada dan memiliki permission setara superadmin
func ensureAdministratorRole(db *gorm.DB, zapLog *zap.Logger) {
	// Cek apakah role administrator sudah ada
	var adminRole domain.RoleModel
	if err := db.Where("name = ?", "administrator").First(&adminRole).Error; err == nil {
		zapLog.Info("Administrator role already exists, skipping creation")
		return
	}

	// Buat role administrator
	adminRole = domain.RoleModel{
		ID:          uuid.GenerateUUID(),
		Name:        "administrator",
		Description: "Administrator - Full access (kecuali fitur development)",
		Level:       0,
		IsSystem:    true,
	}
	if err := db.Create(&adminRole).Error; err != nil {
		zapLog.Warn("Failed to create administrator role", zap.Error(err))
		return
	}

	// Assign permission sama dengan superadmin
	permissionNames := []string{"global:*", "company:manage", "user:manage", "document:manage", "dashboard:view_all", "report:generate", "audit:view_all"}
	for _, permName := range permissionNames {
		var perm domain.PermissionModel
		if err := db.Where("name = ?", permName).First(&perm).Error; err == nil {
			rolePerm := domain.RolePermissionModel{
				RoleID:       adminRole.ID,
				PermissionID: perm.ID,
			}
			if err := db.Create(&rolePerm).Error; err != nil {
				zapLog.Warn("Failed to assign permission to administrator role", zap.String("permission", permName), zap.Error(err))
			}
		}
	}

	zapLog.Info("Administrator role created/ensured successfully")
}
