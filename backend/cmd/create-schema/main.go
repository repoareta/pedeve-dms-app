package main

import (
	"log"
	"os"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
)

func main() {
	// DATABASE_URL must be set via environment variable for security
	// Never hardcode database credentials in source code
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("❌ DATABASE_URL environment variable is required. Please set it before running this command.")
	}

	// Init logger
	logger.InitLogger()
	defer logger.Sync()

	log.Println("Creating database schema...")

	// Init database (will auto-migrate)
	database.InitDB()

	log.Println("✅ Database schema created successfully!")
}
