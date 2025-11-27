package jwt

import (
	"errors"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/secrets"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getJWTSecret())

// getJWTSecret mengambil secret JWT dari Vault atau environment variable
func getJWTSecret() string {
	secret, err := secrets.GetSecretWithFallback("jwt_secret", "JWT_SECRET", "your-secret-key-change-in-production-min-32-chars")
	if err != nil {
		// Fallback to default (development only)
		return "your-secret-key-change-in-production-min-32-chars"
	}
	return secret
}

// Claims merepresentasikan JWT claims dengan company hierarchy support
type Claims struct {
	UserID         string  `json:"user_id"`
	Username       string  `json:"username"`
	RoleID         *string `json:"role_id,omitempty"`         // Role ID (nullable untuk backward compatibility)
	RoleName       string  `json:"role_name"`                 // Role name (e.g., "superadmin", "admin", "manager", "staff")
	CompanyID      *string `json:"company_id,omitempty"`      // Company ID (NULL untuk superadmin)
	CompanyLevel   int     `json:"company_level"`             // Company level dalam hierarchy (0=root, 1=holding, etc)
	HierarchyScope string  `json:"hierarchy_scope"`           // "global", "company", "sub_company"
	Permissions    []string `json:"permissions,omitempty"`    // List of permission names
	jwt.RegisteredClaims
}

// GenerateJWT menghasilkan token JWT untuk user
func GenerateJWT(userID, username string, roleID *string, roleName string, companyID *string, companyLevel int, hierarchyScope string, permissions []string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expired dalam 24 jam

	claims := &Claims{
		UserID:         userID,
		Username:       username,
		RoleID:         roleID,
		RoleName:       roleName,
		CompanyID:      companyID,
		CompanyLevel:   companyLevel,
		HierarchyScope: hierarchyScope,
		Permissions:    permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dms-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT memvalidasi token JWT dan mengembalikan claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

