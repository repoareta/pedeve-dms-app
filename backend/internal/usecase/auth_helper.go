package usecase

import (
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
)

// GetUserAuthInfo mendapatkan informasi lengkap user untuk JWT generation
// Returns: roleID, roleName, companyID, companyLevel, hierarchyScope, permissions
// Role yang dikembalikan adalah role TERTINGGI dari semua assignments (untuk dashboard)
func GetUserAuthInfo(userID string) (*string, string, *string, int, string, []string, error) {
	zapLog := logger.GetLogger()

	// Get user with relationships
	userRepo := repository.NewUserRepository()
	user, role, company, err := userRepo.GetUserWithRoleAndCompany(userID)
	if err != nil {
		return nil, "", nil, 0, "", nil, err
	}

	// Get all assignments untuk user ini (dari junction table)
	assignmentRepo := repository.NewUserCompanyAssignmentRepository()
	assignments, err := assignmentRepo.GetByUserID(userID)
	if err == nil && len(assignments) > 0 {
		// Cari role tertinggi dari semua assignments aktif
		roleRepo := repository.NewRoleRepository()
		var highestRole *domain.RoleModel
		highestLevel := 999 // Start dengan level tinggi, semakin kecil semakin tinggi role

		for _, assignment := range assignments {
			if !assignment.IsActive || assignment.RoleID == nil {
				continue
			}

			// Get role detail
			roleDetail, err := roleRepo.GetByID(*assignment.RoleID)
			if err != nil {
				continue
			}

			// Role level: 0=superadmin, 1=admin, 2=manager, 3=staff
			// Semakin kecil level, semakin tinggi role
			if roleDetail.Level < highestLevel {
				highestLevel = roleDetail.Level
				highestRole = roleDetail
			}
		}

		// Jika ditemukan role dari assignments, gunakan yang tertinggi
		if highestRole != nil {
			roleID := &highestRole.ID
			roleName := highestRole.Name

			// Get company dari assignment dengan role tertinggi (atau primary company)
			var companyID *string
			var companyLevel int
			hierarchyScope := "global"

			// Cari company dari assignment dengan role tertinggi.
			// Untuk role global (superadmin/administrator), prioritaskan holding (level 0) jika ada.
			for _, assignment := range assignments {
				if assignment.IsActive && assignment.RoleID != nil {
					roleDetail, err := roleRepo.GetByID(*assignment.RoleID)
					if err == nil && roleDetail.Level == highestLevel {
						// Ambil data company
						companyRepo := repository.NewCompanyRepository()
						if comp, err := companyRepo.GetByID(assignment.CompanyID); err == nil {
							// Jika belum ada company terpilih, atau kita menemukan holding (level 0), set.
							if companyID == nil || comp.Level == 0 {
								companyID = &assignment.CompanyID
								companyLevel = comp.Level
								if companyLevel == 0 {
									hierarchyScope = "global"
								} else if companyLevel == 1 {
									hierarchyScope = "company"
								} else {
									hierarchyScope = "sub_company"
								}
							}
						}
					}
				}
			}

			// Fallback ke company dari UserModel jika tidak ada dari assignment
			if companyID == nil && user.CompanyID != nil {
				companyID = user.CompanyID
				if company != nil {
					companyLevel = company.Level
					if companyLevel == 0 {
						hierarchyScope = "global"
					} else if companyLevel == 1 {
						hierarchyScope = "company"
					} else {
						hierarchyScope = "sub_company"
					}
				}
			}

			// Get permissions from role
			permissions := []string{}
			if roleID != nil {
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
				case "superadmin", "administrator":
					permissions = []string{"*"}
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
	}

	// Fallback: gunakan role dari UserModel (backward compatibility)
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
		case "superadmin", "administrator":
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
