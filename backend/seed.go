package main

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// SeedSuperAdmin membuat user superadmin kalau belum ada
func SeedSuperAdmin() {
	// Cek apakah superadmin sudah ada
	var existingUser UserModel
	result := DB.Where("username = ? OR role = ?", "superadmin", "superadmin").First(&existingUser)
	if result.Error == nil {
		log.Println("Superadmin user already exists")
		return
	}

	// Hash password untuk superadmin
	hashedPassword, err := HashPassword("Pedeve123")
	if err != nil {
		log.Printf("Failed to hash superadmin password: %v", err)
		return
	}

	// Buat user superadmin
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

	// Simpan ke database
	if err := DB.Create(superAdmin).Error; err != nil {
		log.Printf("Failed to create superadmin user: %v", err)
		return
	}

	log.Println("Superadmin user created successfully")
	log.Println("Username: superadmin")
	log.Println("Password: Pedeve123")
}

