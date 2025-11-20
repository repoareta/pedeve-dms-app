package main

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// SeedSuperAdmin creates a superadmin user if it doesn't exist
func SeedSuperAdmin() {
	// Check if superadmin already exists
	var existingUser UserModel
	result := DB.Where("username = ? OR role = ?", "superadmin", "superadmin").First(&existingUser)
	if result.Error == nil {
		log.Println("Superadmin user already exists")
		return
	}

	// Hash password for superadmin
	hashedPassword, err := HashPassword("Pedeve123")
	if err != nil {
		log.Printf("Failed to hash superadmin password: %v", err)
		return
	}

	// Create superadmin user
	now := time.Now()
	superAdmin := &UserModel{
		ID:        uuid.New().String(),
		Username:  "superadmin",
		Email:     "superadmin@pertamina.com",
		Password:  hashedPassword,
		Role:      "superadmin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save to database
	if err := DB.Create(superAdmin).Error; err != nil {
		log.Printf("Failed to create superadmin user: %v", err)
		return
	}

	log.Println("Superadmin user created successfully")
	log.Println("Username: superadmin")
	log.Println("Password: Pedeve123")
}

