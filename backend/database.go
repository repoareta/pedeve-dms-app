package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// GenerateUUID menghasilkan string UUID baru
func GenerateUUID() string {
	return uuid.New().String()
}


// InitDB menginisialisasi koneksi database
func InitDB() {
	var err error
	var dialector gorm.Dialector

	// Ambil URL database dari environment
	dbURL := os.Getenv("DATABASE_URL")

	// Gunakan SQLite untuk development jika DATABASE_URL tidak diset
	if dbURL == "" {
		log.Println("Using SQLite database (development)")
		dialector = sqlite.Open("dms.db")
	} else {
		log.Println("Using PostgreSQL database")
		dialector = postgres.Open(dbURL)
	}

	// Buka koneksi database
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate schema
	err = DB.AutoMigrate(&UserModel{}, &TwoFactorAuth{}, &AuditLog{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

// Model User untuk database
type UserModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Role      string    `gorm:"default:'user'" json:"role"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName menentukan nama tabel untuk UserModel
func (UserModel) TableName() string {
	return "users"
}

