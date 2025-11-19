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

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}


// InitDB initializes database connection
func InitDB() {
	var err error
	var dialector gorm.Dialector

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")

	// Use SQLite for development if DATABASE_URL not set
	if dbURL == "" {
		log.Println("Using SQLite database (development)")
		dialector = sqlite.Open("dms.db")
	} else {
		log.Println("Using PostgreSQL database")
		dialector = postgres.Open(dbURL)
	}

	// Open database connection
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate schema
	err = DB.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected and migrated successfully")
}

// User model for database
type UserModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

