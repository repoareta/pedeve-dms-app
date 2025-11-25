package main

import (
	"fmt"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./cmd/test-password/main.go <email_or_username> <password>")
		fmt.Println("Example: go run ./cmd/test-password/main.go dhani@pertamina.com Pedeve123")
		os.Exit(1)
	}

	emailOrUsername := os.Args[1]
	testPassword := os.Args[2]

	// Initialize logger
	logger.InitLogger()
	zapLog := logger.GetLogger()
	defer zapLog.Sync()

	// Initialize database
	database.InitDB()

	// Find user (case-insensitive search)
	var user domain.UserModel
	result := database.GetDB().Where("LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)", emailOrUsername, emailOrUsername).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Printf("❌ User tidak ditemukan: %s\n", emailOrUsername)
			fmt.Println()
			fmt.Println("🔍 Mencari user dengan pattern serupa...")
			var similarUsers []domain.UserModel
			// Use LOWER for case-insensitive search (works for both PostgreSQL and SQLite)
			database.GetDB().Where("LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", "%"+emailOrUsername+"%", "%"+emailOrUsername+"%").Find(&similarUsers)
			if len(similarUsers) > 0 {
				fmt.Println("   User yang mirip ditemukan:")
				for _, u := range similarUsers {
					fmt.Printf("   - Username: %s, Email: %s, Role: %s, Active: %v\n", u.Username, u.Email, u.Role, u.IsActive)
				}
			} else {
				fmt.Println("   Tidak ada user yang mirip ditemukan.")
			}
			fmt.Println()
			fmt.Println("💡 Gunakan script list-users.sh untuk melihat semua users:")
			fmt.Println("   ./backend/scripts/list-users.sh")
			os.Exit(1)
		}
		zapLog.Fatal("Database error", zap.Error(result.Error))
	}

	fmt.Printf("✅ User ditemukan:\n")
	fmt.Printf("   ID: %s\n", user.ID)
	fmt.Printf("   Username: %s\n", user.Username)
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Role: %s\n", user.Role)
	fmt.Printf("   Is Active: %v\n", user.IsActive)
	fmt.Printf("   Role ID: %v\n", user.RoleID)
	fmt.Printf("   Company ID: %v\n", user.CompanyID)
	fmt.Printf("   Password Hash Length: %d\n", len(user.Password))
	fmt.Printf("   Password Hash Preview: %s...\n", user.Password[:min(20, len(user.Password))])
	fmt.Println()

	// Check if password is hashed (bcrypt hashes start with $2a$ or $2b$)
	if len(user.Password) < 10 || (user.Password[:4] != "$2a$" && user.Password[:4] != "$2b$") {
		fmt.Printf("⚠️  WARNING: Password tidak ter-hash dengan benar!\n")
		fmt.Printf("   Password seharusnya di-hash dengan bcrypt (dimulai dengan $2a$ atau $2b$)\n")
		fmt.Printf("   Panjang password saat ini: %d karakter\n", len(user.Password))
		fmt.Println()
		fmt.Printf("💡 Solusi: Gunakan fitur 'Reset Password' dari superadmin untuk memperbaiki password\n")
		os.Exit(1)
	}

	// Test password
	fmt.Printf("🔐 Testing password...\n")
	passwordValid := password.CheckPasswordHash(testPassword, user.Password)
	if passwordValid {
		fmt.Printf("✅ Password VALID - Login seharusnya berhasil!\n")
	} else {
		fmt.Printf("❌ Password TIDAK VALID - Password yang diberikan tidak cocok dengan hash di database\n")
		fmt.Printf("   Password yang di-test: %s\n", testPassword)
		fmt.Println()
		fmt.Printf("💡 Solusi:\n")
		fmt.Printf("   1. Pastikan password yang digunakan benar\n")
		fmt.Printf("   2. Gunakan fitur 'Reset Password' dari superadmin untuk reset password\n")
		os.Exit(1)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

