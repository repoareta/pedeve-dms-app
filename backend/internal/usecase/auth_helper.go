package usecase

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// GetUserAuthInfo mendapatkan informasi lengkap user untuk JWT generation
// Returns: roleID, roleName, companyID, companyLevel, hierarchyScope, permissions
func GetUserAuthInfo(userID string) (*string, string, *string, int, string, []string, error) {
	zapLog := logger.GetLogger()
	
	// Get user with relationships
	userRepo := repository.NewUserRepository()
	user, role, company, err := userRepo.GetUserWithRoleAndCompany(userID)
	if err != nil {
		return nil, "", nil, 0, "", nil, err
	}
	
	// Determine role info
	var roleID *string
	roleName := "user" // Default role name
	if role != nil {
		roleID = &role.ID
		roleName = role.Name
	} else if user.Role != "" {
		// Fallback ke legacy role field
		roleName = user.Role
	}
	
	// Determine company info
	var companyID *string
	companyLevel := 0
	hierarchyScope := "global"
	
	if company != nil {
		companyID = &company.ID
		companyLevel = company.Level
		
		// Determine hierarchy scope based on company level
		if companyLevel == 0 {
			hierarchyScope = "global" // Root/Superadmin
		} else if companyLevel == 1 {
			hierarchyScope = "company" // Holding company
		} else {
			hierarchyScope = "sub_company" // Subsidiary or deeper
		}
	} else if user.CompanyID == nil {
		// Superadmin (no company)
		hierarchyScope = "global"
	}
	
	// Get permissions from role
	permissions := []string{}
	if roleID != nil {
		roleRepo := repository.NewRoleRepository()
		permissionModels, err := roleRepo.GetPermissions(*roleID)
		if err == nil {
			for _, perm := range permissionModels {
				permissions = append(permissions, perm.Name)
			}
		} else {
			zapLog.Warn("Failed to get permissions for role", zap.String("role_id", *roleID), zap.Error(err))
		}
	}
	
	// Add default permissions based on role name (backward compatibility)
	if len(permissions) == 0 {
		switch roleName {
		case "superadmin":
			permissions = []string{"*"} // All permissions
		case "admin":
			permissions = []string{"view_dashboard", "manage_users", "manage_documents", "view_reports"}
		case "manager":
			permissions = []string{"view_dashboard", "view_documents", "view_reports"}
		case "staff":
			permissions = []string{"view_dashboard", "view_documents"}
		}
	}
	
	return roleID, roleName, companyID, companyLevel, hierarchyScope, permissions, nil
}

