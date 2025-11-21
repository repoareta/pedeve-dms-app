package main

import "time"

// User merepresentasikan user dalam sistem
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Password  string    `json:"-"` // Jangan sertakan password di JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest merepresentasikan payload request login
type LoginRequest struct {
	Username string `json:"username" example:"admin"` // Bisa username atau email
	Password string `json:"password" example:"password123"`
	Code     string `json:"code,omitempty" example:"123456"` // Kode 2FA (opsional)
}

// RegisterRequest merepresentasikan payload request registrasi
type RegisterRequest struct {
	Username string `json:"username" example:"admin"`
	Email    string `json:"email" example:"admin@example.com"`
	Password string `json:"password" example:"password123"`
}

// AuthResponse merepresentasikan response autentikasi
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse merepresentasikan response error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

