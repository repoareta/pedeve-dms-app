package main

import (
	"log"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/database"
)

// Seed database with default admin user
func main() {
	// Initialize database
	database.InitDB()

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "admin123"
	}

	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}

	// Check if admin already exists
	var existingUser database.UserModel
	result := database.DB.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		log.Printf("Admin user '%s' already exists", username)
		return
	}

	// Hash password
	hashedPassword, err := database.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	admin := &database.UserModel{
		ID:       database.GenerateUUID(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	if err := database.DB.Create(admin).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	log.Printf("Admin user created successfully!")
	log.Printf("Username: %s", username)
	log.Printf("Password: %s", password)
	log.Printf("Email: %s", email)
	log.Println("\n⚠️  Please change the default password after first login!")
}

