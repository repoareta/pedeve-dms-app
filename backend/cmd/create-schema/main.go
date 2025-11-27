package main

import (
	"log"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
)

func main() {
	// Set DATABASE_URL if not set
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable")
	}

	// Init logger
	logger.InitLogger()
	defer logger.Sync()

	log.Println("Creating database schema...")
	
	// Init database (will auto-migrate)
	database.InitDB()
	
	log.Println("✅ Database schema created successfully!")
}

