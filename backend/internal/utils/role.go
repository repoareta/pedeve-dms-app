package utils

import "strings"

// IsSuperAdminLike returns true untuk superadmin dan administrator (hak istimewa setara).
func IsSuperAdminLike(roleName string) bool {
	switch strings.ToLower(roleName) {
	case "superadmin", "administrator":
		return true
	default:
		return false
	}
}
