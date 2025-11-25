package main

import (
	"fmt"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/seed"
	"go.uber.org/zap"
)

func main() {
	zapLog := logger.GetLogger()

	// Initialize database (InitDB doesn't return error, it panics on failure)
	database.InitDB()

	// Update superadmin password from Vault
	if err := seed.UpdateSuperadminPasswordFromVault(); err != nil {
		zapLog.Error("Failed to update superadmin password", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Superadmin password updated successfully from Vault")
}

