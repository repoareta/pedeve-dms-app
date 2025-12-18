package usecase

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
	"github.com/repoareta/pedeve-dms-app/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DevelopmentUseCase handles development-related operations (seeding, resetting data)
type DevelopmentUseCase interface {
	ResetSubsidiaryData() error
	RunSubsidiarySeeder() (bool, error) // Returns (alreadyExists, error)
	CheckSeederDataExists() (bool, error)
	ResetReportData() error
	RunReportSeeder() error
	CheckReportDataExists() (bool, error)
	ResetAllFinancialReports() error // Reset all financial reports (all companies)
	// Combined operations
	RunAllSeeders() error                           // Run all seeders in order: company -> reports
	ResetAllSeededData() error                      // Reset all seeded data: reports -> company
	CheckAllSeederStatus() (map[string]bool, error) // Check status of all seeders
}

type developmentUseCase struct {
	db                        *gorm.DB
	companyRepo               repository.CompanyRepository
	userRepo                  repository.UserRepository
	roleRepo                  repository.RoleRepository
	userCompanyAssignmentRepo repository.UserCompanyAssignmentRepository
	reportRepo                repository.ReportRepository
	financialReportRepo       repository.FinancialReportRepository
}

// NewDevelopmentUseCaseWithDB creates a new development use case with injected DB (for testing)
func NewDevelopmentUseCaseWithDB(db *gorm.DB) DevelopmentUseCase {
	return &developmentUseCase{
		db:                        db,
		companyRepo:               repository.NewCompanyRepositoryWithDB(db),
		userRepo:                  repository.NewUserRepositoryWithDB(db),
		roleRepo:                  repository.NewRoleRepositoryWithDB(db),
		userCompanyAssignmentRepo: repository.NewUserCompanyAssignmentRepositoryWithDB(db),
		reportRepo:                repository.NewReportRepositoryWithDB(db),
		financialReportRepo:       repository.NewFinancialReportRepositoryWithDB(db),
	}
}

// NewDevelopmentUseCase creates a new development use case with default DB (backward compatibility)
func NewDevelopmentUseCase() DevelopmentUseCase {
	return NewDevelopmentUseCaseWithDB(database.GetDB())
}

// ResetSubsidiaryData deletes all subsidiary companies and their related users
// This will:
// 1. Get all companies except the root holding (parent_id IS NULL)
// 2. Get all descendants of each subsidiary
// 3. Delete all user_company_assignments for these companies
// 4. Delete all users assigned to these companies (except superadmin)
// 5. Delete all companies (soft delete: set is_active = false)
func (uc *developmentUseCase) ResetSubsidiaryData() error {
	zapLog := logger.GetLogger()
	db := uc.db
	if db == nil {
		db = database.GetDB()
	}

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			zapLog.Error("Panic during reset subsidiary data", zap.Any("panic", r))
		}
	}()

	// 1. Get all companies except root holding
	// IMPORTANT: Exclude holding by BOTH parent_id IS NULL AND code != 'PDV' untuk safety
	// Ini memastikan holding tidak terhapus meskipun ada bug di data
	var allCompanies []domain.CompanyModel
	if err := tx.Where("parent_id IS NOT NULL AND code != ?", "PDV").Find(&allCompanies).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get companies: %w", err)
	}

	// Double check: juga exclude holding by code untuk extra safety
	// Filter out any company with code 'PDV' (holding) just in case
	filteredCompanies := make([]domain.CompanyModel, 0)
	for _, comp := range allCompanies {
		if comp.Code != "PDV" {
			filteredCompanies = append(filteredCompanies, comp)
		}
	}
	allCompanies = filteredCompanies

	if len(allCompanies) == 0 {
		tx.Rollback()
		zapLog.Info("No subsidiary companies found to reset")
		return nil
	}

	// Collect all company IDs (including descendants)
	companyIDs := make([]string, 0, len(allCompanies))
	for _, comp := range allCompanies {
		companyIDs = append(companyIDs, comp.ID)
	}

	zapLog.Info("Resetting subsidiary data", zap.Int("company_count", len(companyIDs)))

	// 2. First, collect all user IDs that will be affected BEFORE deleting assignments
	// Get user IDs from junction table assignments
	var assignments []domain.UserCompanyAssignmentModel
	if err := tx.Where("company_id IN ?", companyIDs).Find(&assignments).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get user company assignments: %w", err)
	}

	// Collect unique user IDs from assignments
	userIDsFromAssignments := make(map[string]bool)
	for _, assignment := range assignments {
		userIDsFromAssignments[assignment.UserID] = true
	}

	// Also get users from UserModel.CompanyID
	var usersFromCompanyID []domain.UserModel
	if err := tx.Where("company_id IN ?", companyIDs).Find(&usersFromCompanyID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get users from company_id: %w", err)
	}

	// Combine user IDs from both sources
	allUserIDs := make(map[string]bool)
	for userID := range userIDsFromAssignments {
		allUserIDs[userID] = true
	}
	for _, user := range usersFromCompanyID {
		allUserIDs[user.ID] = true
	}

	// Filter out superadmin users
	userIDsToDelete := make([]string, 0)
	for userID := range allUserIDs {
		// Get user to check if superadmin
		var user domain.UserModel
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			continue // Skip if user not found
		}
		if user.Role != "superadmin" && user.Username != "superadmin" {
			userIDsToDelete = append(userIDsToDelete, userID)
		}
	}

	// 3. Delete all user_company_assignments for these companies (by company_id)
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.UserCompanyAssignmentModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user company assignments: %w", err)
	}
	zapLog.Info("Deleted user company assignments by company_id", zap.Int("company_count", len(companyIDs)))

	// 4. Delete all remaining assignments for users that will be deleted (by user_id)
	// This handles edge cases where user might have assignments in other companies
	if len(userIDsToDelete) > 0 {
		if err := tx.Where("user_id IN ?", userIDsToDelete).Delete(&domain.UserCompanyAssignmentModel{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete user assignments by user_id: %w", err)
		}
		zapLog.Info("Deleted user assignments by user_id", zap.Int("user_count", len(userIDsToDelete)))
	}

	// 5. Delete all related data for companies (shareholders, directors, business_fields)
	// Delete shareholders (both as company owner and as shareholder company reference)
	// Delete shareholders where company_id IN companyIDs (shareholders of these companies)
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.ShareholderModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete shareholders: %w", err)
	}
	// Delete shareholders where shareholder_company_id IN companyIDs (these companies as shareholders in other companies)
	if err := tx.Where("shareholder_company_id IN ?", companyIDs).Delete(&domain.ShareholderModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete shareholders by shareholder_company_id: %w", err)
	}
	zapLog.Info("Deleted shareholders", zap.Int("company_count", len(companyIDs)))

	// Delete directors
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.DirectorModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete directors: %w", err)
	}
	zapLog.Info("Deleted directors", zap.Int("company_count", len(companyIDs)))

	// Delete business_fields
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.BusinessFieldModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete business_fields: %w", err)
	}
	zapLog.Info("Deleted business_fields", zap.Int("company_count", len(companyIDs)))

	// Delete reports (from reports table)
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.ReportModel{}).Error; err != nil {
		// Log warning but don't fail - reports might not exist
		zapLog.Warn("Failed to delete reports (might not exist)", zap.Error(err))
	} else {
		zapLog.Info("Deleted reports", zap.Int("company_count", len(companyIDs)))
	}

	// Delete financial_reports (from financial_reports table) - CRITICAL: must be deleted before companies
	if err := tx.Where("company_id IN ?", companyIDs).Delete(&domain.FinancialReportModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete financial_reports: %w", err)
	}
	zapLog.Info("Deleted financial_reports", zap.Int("company_count", len(companyIDs)))

	// 6. Delete users (hard delete for development reset)
	if len(userIDsToDelete) > 0 {
		if err := tx.Where("id IN ?", userIDsToDelete).Delete(&domain.UserModel{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete users: %w", err)
		}
		zapLog.Info("Deleted users", zap.Int("user_count", len(userIDsToDelete)))
	}

	// 7. Delete all companies (hard delete for development reset)
	if err := tx.Where("id IN ?", companyIDs).Delete(&domain.CompanyModel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete companies: %w", err)
	}
	zapLog.Info("Deleted companies (hard delete)", zap.Int("company_count", len(companyIDs)))

	// 8. CRITICAL: Reset holding company level to 0 dan ensure parent_id is NULL
	// Ini penting untuk memastikan holding level tidak kacau setelah reset
	holding, err := uc.companyRepo.GetByCode("PDV")
	if err == nil && holding != nil {
		// Reset holding level to 0 dan pastikan parent_id is NULL
		if err := tx.Model(&domain.CompanyModel{}).
			Where("code = ?", "PDV").
			Updates(map[string]interface{}{
				"level":     0,
				"parent_id": nil,
				"is_active": true, // Ensure holding is active
			}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reset holding level: %w", err)
		}
		zapLog.Info("Reset holding company level to 0", zap.String("holding_id", holding.ID))
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	zapLog.Info("Successfully reset subsidiary data",
		zap.Int("companies_deleted", len(companyIDs)),
		zap.Int("users_deleted", len(userIDsToDelete)),
	)

	return nil
}

// CheckSeederDataExists checks if seeder data already exists
// It checks for the holding company with code "PDV" and at least one subsidiary
func (uc *developmentUseCase) CheckSeederDataExists() (bool, error) {
	// Check if holding company exists
	holding, err := uc.companyRepo.GetByCode("PDV")
	if err != nil {
		// If holding doesn't exist, seeder data doesn't exist
		return false, nil
	}

	if holding == nil {
		return false, nil
	}

	// Check if at least one subsidiary exists (any company with parent_id pointing to holding or any company)
	var count int64
	if err := uc.db.Model(&domain.CompanyModel{}).
		Where("parent_id IS NOT NULL AND is_active = ?", true).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to count subsidiaries: %w", err)
	}

	return count > 0, nil
}

// RunSubsidiarySeeder runs the subsidiary seeder using the latest seeder command
// Returns (alreadyExists, error)
// If alreadyExists is true, it means seeder data already exists and the operation was cancelled
func (uc *developmentUseCase) RunSubsidiarySeeder() (bool, error) {
	zapLog := logger.GetLogger()

	// Check if seeder data already exists
	exists, err := uc.CheckSeederDataExists()
	if err != nil {
		return false, fmt.Errorf("failed to check seeder data: %w", err)
	}

	if exists {
		zapLog.Warn("Seeder data already exists, skipping seeder execution")
		return true, nil
	}

	// Determine how to run the seeder
	// Strategy 1: Try to use pre-built binary (for production) - CHECK FIRST
	// Strategy 2: Fallback to go run (for development) - requires backendDir
	var cmd *exec.Cmd
	var useBinary bool
	var backendDir string
	var found bool

	// Helper function to check if a file is an executable binary
	checkBinary := func(path string) bool {
		if path == "" {
			return false
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// Check if it's executable
			return info.Mode().Perm()&0111 != 0
		}
		return false
	}

	// Helper function to check if a directory contains cmd/seed-companies
	checkBackendDir := func(dir string) bool {
		if dir == "" {
			return false
		}
		seedPath := filepath.Join(dir, "cmd", "seed-companies")
		if info, err := os.Stat(seedPath); err == nil && info.IsDir() {
			return true
		}
		return false
	}

	// FIRST: Check for pre-built binary (production scenario)
	// Get executable directory to search for binary in same location
	execDir := "/root" // Default fallback
	if execPath, err := os.Executable(); err == nil {
		execDir = filepath.Dir(execPath)
	}

	binaryLocations := []string{
		"/root/seed-companies",                   // Production default location
		"/app/seed-companies",                    // Alternative production location
		filepath.Join(execDir, "seed-companies"), // Same dir as executable
		os.Getenv("SEED_COMPANIES_BINARY"),       // Allow override via env var
	}

	for _, binPath := range binaryLocations {
		if binPath == "" {
			continue
		}
		if checkBinary(binPath) {
			cmd = exec.Command(binPath)
			useBinary = true
			zapLog.Info("Using pre-built seeder binary", zap.String("binary_path", binPath))
			// If binary found, we don't need backendDir
			found = true
			break
		}
	}

	// SECOND: Only search for backendDir if binary not found (development scenario)
	if !useBinary {
		// Strategy 1: Try environment variable BACKEND_DIR (for production/Docker)
		if backendDirEnv := os.Getenv("BACKEND_DIR"); backendDirEnv != "" {
			if checkBackendDir(backendDirEnv) {
				backendDir = backendDirEnv
				found = true
				zapLog.Info("Found backend directory from BACKEND_DIR env", zap.String("dir", backendDir))
			}
		}

		// Strategy 2: Try current working directory
		if !found {
			wd, err := os.Getwd()
			if err == nil {
				// Check if we're in backend directory
				if checkBackendDir(wd) {
					backendDir = wd
					found = true
					zapLog.Info("Found backend directory from current working directory", zap.String("dir", backendDir))
				} else if checkBackendDir(filepath.Join(wd, "backend")) {
					backendDir = filepath.Join(wd, "backend")
					found = true
					zapLog.Info("Found backend directory from current working directory/backend", zap.String("dir", backendDir))
				}
			}
		}

		// Strategy 3: Try from executable path (multiple levels up)
		if !found {
			execPath, err := os.Executable()
			if err == nil {
				// Try different levels up from executable
				// Common locations: /app, /app/backend, /root, /root/backend, etc.
				baseDir := filepath.Dir(execPath)
				for i := 0; i < 5; i++ {
					// Check current level
					if checkBackendDir(baseDir) {
						backendDir = baseDir
						found = true
						zapLog.Info("Found backend directory from executable path", zap.String("dir", backendDir), zap.Int("levels_up", i))
						break
					}
					// Check backend subdirectory
					if checkBackendDir(filepath.Join(baseDir, "backend")) {
						backendDir = filepath.Join(baseDir, "backend")
						found = true
						zapLog.Info("Found backend directory from executable path/backend", zap.String("dir", backendDir), zap.Int("levels_up", i))
						break
					}
					// Go up one level
					parentDir := filepath.Dir(baseDir)
					if parentDir == baseDir {
						break // Reached root
					}
					baseDir = parentDir
				}
			}
		}

		// Strategy 4: Try common production paths
		if !found {
			commonPaths := []string{
				"/app/backend",
				"/app",
				"/root/backend",
				"/root",
				"/usr/local/backend",
				"/opt/backend",
			}
			for _, path := range commonPaths {
				if checkBackendDir(path) {
					backendDir = path
					found = true
					zapLog.Info("Found backend directory from common path", zap.String("dir", backendDir))
					break
				}
			}
		}

		// If binary not found, we need backendDir for go run
		if !found {
			// Log diagnostic information
			wd, _ := os.Getwd()
			execPath, _ := os.Executable()
			zapLog.Error("Failed to find backend directory",
				zap.String("working_dir", wd),
				zap.String("executable_path", execPath),
				zap.String("backend_dir_env", os.Getenv("BACKEND_DIR")),
			)
			return false, fmt.Errorf("failed to find backend directory with cmd/seed-companies. Working dir: %s, Executable: %s", wd, execPath)
		}

		seedPath := filepath.Join(backendDir, "cmd", "seed-companies")
		zapLog.Info("Running company seeder with go run", zap.String("backend_dir", backendDir), zap.String("seed_path", seedPath))

		// Check if go is available
		if _, err := exec.LookPath("go"); err != nil {
			return false, fmt.Errorf("go compiler not found and seeder binary not found. Please build seeder binary or install Go")
		}
		cmd = exec.Command("go", "run", "./cmd/seed-companies")
		cmd.Dir = backendDir
		zapLog.Info("Using go run to execute seeder", zap.String("backend_dir", backendDir))
	} else {
		zapLog.Info("Running company seeder with binary")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Preserve DATABASE_URL environment variable - CRITICAL for same database connection
	// Helper function to mask sensitive parts of DB URL for logging
	maskDBURL := func(url string) string {
		if len(url) <= 20 {
			return "***"
		}
		return url[:10] + "***" + url[len(url)-10:]
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try to get from secrets (same way database.InitDB does)
		// This ensures seeder uses the same database connection
		zapLog.Warn("DATABASE_URL not in environment, seeder may use different database")
	} else {
		zapLog.Info("Using DATABASE_URL for seeder command", zap.String("db_url", maskDBURL(dbURL)))
	}

	// Always set DATABASE_URL in command environment to ensure same database
	cmdEnv := os.Environ()
	if dbURL != "" {
		// Override DATABASE_URL in command environment
		cmdEnv = append(cmdEnv, "DATABASE_URL="+dbURL)
	}
	cmd.Env = cmdEnv

	// Capture command output for debugging
	// Use separate writers to ensure we capture all output
	var stdout, stderr bytes.Buffer
	// Write to both original streams AND buffers to ensure output is visible
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	// Set unbuffered output to ensure immediate flushing
	cmd.Env = append(cmd.Env, "GODEBUG=gctrace=0") // Disable GC tracing

	// Run command and capture exit code
	runErr := cmd.Run()
	var exitCode int
	if runErr != nil {
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		zapLog.Error("Seeder command failed",
			zap.Error(runErr),
			zap.Int("exit_code", exitCode),
			zap.String("stdout", stdout.String()),
			zap.String("stderr", stderr.String()),
		)
		return false, fmt.Errorf("failed to run company seeder (exit code %d): %w (stdout: %s, stderr: %s)", exitCode, runErr, stdout.String(), stderr.String())
	}

	// Log seeder output for debugging (even if exit code is 0, check if companies were created)
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Check if seeder output indicates success or failure
	if len(stdoutStr) > 0 {
		zapLog.Info("Seeder stdout", zap.String("output", stdoutStr))
		// Check for common error messages in output (seeder exits early with return, not error)
		// But now seeder will continue even if admin role not found, so check for actual success indicators
		if strings.Contains(stdoutStr, "❌ Failed to create holding") ||
			strings.Contains(stdoutStr, "❌ Failed to create") ||
			(!strings.Contains(stdoutStr, "Creating/Updating Holding Company") && !strings.Contains(stdoutStr, "Created:") && !strings.Contains(stdoutStr, "Updated:")) {
			// Only error if we don't see any company creation/update messages
			if !strings.Contains(stdoutStr, "Created:") && !strings.Contains(stdoutStr, "Updated:") {
				zapLog.Error("Seeder output indicates failure or early exit", zap.String("output", stdoutStr))
				return false, fmt.Errorf("seeder exited early or failed. output: %s", stdoutStr)
			}
		}
	}
	if len(stderrStr) > 0 {
		zapLog.Warn("Seeder stderr", zap.String("output", stderrStr))
	} else {
		zapLog.Warn("Seeder stderr is EMPTY - this suggests stderr was not captured or seeder exited before writing to stderr")
	}

	// Verify that subsidiaries were created by checking the database
	// This ensures data is committed and visible before proceeding
	db := uc.db
	if db == nil {
		db = database.GetDB()
	}

	// Refresh connection to ensure we see the latest data
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			zapLog.Warn("Failed to ping database connection", zap.Error(err))
		}
	}

	// Verify that companies were created - CRITICAL check
	// Add small delay to ensure data is committed (especially if seeder uses different transaction)
	time.Sleep(1 * time.Second)

	// Check output first to see if seeder actually ran (variables already declared above)
	// stdoutStr and stderrStr are already declared at line 403-404

	// Check if seeder output contains actual seeder messages (not just GORM logs)
	// Also check stderr for STEP messages
	hasSeederOutput := strings.Contains(stdoutStr, "Creating/Updating Holding Company") ||
		strings.Contains(stdoutStr, "Created:") ||
		strings.Contains(stdoutStr, "Updated:") ||
		strings.Contains(stdoutStr, "Found admin role") ||
		strings.Contains(stdoutStr, "Admin role not found") ||
		strings.Contains(stdoutStr, "Database initialized") ||
		strings.Contains(stdoutStr, "Connected to database") ||
		strings.Contains(stderrStr, "STEP:") ||
		strings.Contains(stderrStr, "Initializing database") ||
		strings.Contains(stderrStr, "Repositories initialized")

	zapLog.Info("Checking seeder output",
		zap.Bool("has_seeder_output", hasSeederOutput),
		zap.Int("stdout_length", len(stdoutStr)),
		zap.Int("stderr_length", len(stderrStr)),
		zap.String("stdout_preview", truncateString(stdoutStr, 500)),
		zap.String("stderr_preview", truncateString(stderrStr, 500)),
	)

	var totalCount int64
	if err := db.Model(&domain.CompanyModel{}).Where("is_active = ?", true).Count(&totalCount).Error; err != nil {
		zapLog.Error("Failed to verify companies after seeder", zap.Error(err))
		return false, fmt.Errorf("failed to verify companies after seeder: %w", err)
	}

	if totalCount == 0 {
		if !hasSeederOutput {
			zapLog.Error("No companies found and no seeder output detected - seeder may have exited early or output not captured",
				zap.String("stdout_preview", truncateString(stdoutStr, 1000)),
				zap.String("stderr_preview", truncateString(stderrStr, 1000)),
			)
			return false, fmt.Errorf("seeder command completed but no companies were created and no seeder output detected. this suggests seeder exited before creating companies or output was not captured. stdout (first 1000 chars): %s", truncateString(stdoutStr, 1000))
		}

		zapLog.Error("No companies found after seeder execution - seeder failed to create companies",
			zap.String("stdout_preview", truncateString(stdoutStr, 1000)),
			zap.String("stderr_preview", truncateString(stderrStr, 1000)),
		)
		return false, fmt.Errorf("seeder command completed but no companies were created. stdout (first 1000 chars): %s", truncateString(stdoutStr, 1000))
	}

	var subsidiariesCount int64
	if err := db.Model(&domain.CompanyModel{}).
		Where("parent_id IS NOT NULL AND is_active = ? AND code != ?", true, "PDV").
		Count(&subsidiariesCount).Error; err != nil {
		zapLog.Warn("Failed to count subsidiaries after seeder", zap.Error(err))
	} else {
		zapLog.Info("Verified companies after seeder",
			zap.Int64("total_companies", totalCount),
			zap.Int64("subsidiaries", subsidiariesCount),
		)
		if subsidiariesCount == 0 {
			// Try fallback query
			if err := db.Model(&domain.CompanyModel{}).
				Where("level > ? AND is_active = ?", 0, true).
				Count(&subsidiariesCount).Error; err == nil && subsidiariesCount > 0 {
				zapLog.Info("Found subsidiaries using fallback verification", zap.Int64("count", subsidiariesCount))
			} else {
				zapLog.Warn("No subsidiaries found after seeder execution",
					zap.Int64("total_companies", totalCount),
					zap.String("stdout", stdout.String()),
				)
			}
		}
	}

	zapLog.Info("Subsidiary seeder completed successfully")
	return false, nil
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ResetReportData deletes all reports
func (uc *developmentUseCase) ResetReportData() error {
	zapLog := logger.GetLogger()

	count, err := uc.reportRepo.Count()
	if err != nil {
		return fmt.Errorf("failed to count reports: %w", err)
	}

	if count == 0 {
		zapLog.Info("No reports found to reset")
		return nil
	}

	err = uc.reportRepo.DeleteAll()
	if err != nil {
		return fmt.Errorf("failed to delete reports: %w", err)
	}

	zapLog.Info("Successfully reset report data", zap.Int64("reports_deleted", count))
	return nil
}

// ResetAllFinancialReports deletes all financial reports from all companies
func (uc *developmentUseCase) ResetAllFinancialReports() error {
	zapLog := logger.GetLogger()

	// Count financial reports before deletion
	var count int64
	err := uc.db.Model(&domain.FinancialReportModel{}).Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to count financial reports: %w", err)
	}

	if count == 0 {
		zapLog.Info("No financial reports found to reset")
		return nil
	}

	// Delete all financial reports
	err = uc.financialReportRepo.DeleteAll()
	if err != nil {
		return fmt.Errorf("failed to delete financial reports: %w", err)
	}

	zapLog.Info("Successfully reset all financial reports", zap.Int64("financial_reports_deleted", count))
	return nil
}

// RunReportSeeder runs the report seeder
func (uc *developmentUseCase) RunReportSeeder() error {
	zapLog := logger.GetLogger()

	// Check if reports already exist
	exists, err := uc.CheckReportDataExists()
	if err != nil {
		return fmt.Errorf("failed to check report data: %w", err)
	}

	if exists {
		zapLog.Warn("Report data already exists, skipping seeder execution")
		return fmt.Errorf("report data already exists")
	}

	// Get all subsidiaries directly from database with specific filter
	// Use direct DB query to ensure we get the latest data after seeder execution
	db := uc.db
	if db == nil {
		db = database.GetDB()
	}

	var subsidiaries []domain.CompanyModel
	// Query for active subsidiaries (companies with parent_id IS NOT NULL, excluding holding)
	err = db.Where("parent_id IS NOT NULL AND is_active = ? AND code != ?", true, "PDV").
		Find(&subsidiaries).Error
	if err != nil {
		return fmt.Errorf("failed to get subsidiaries: %w", err)
	}

	if len(subsidiaries) == 0 {
		// Try to get any active company with level > 0 as fallback
		err = db.Where("level > ? AND is_active = ?", 0, true).
			Find(&subsidiaries).Error
		if err != nil {
			return fmt.Errorf("failed to get subsidiaries: %w", err)
		}
		if len(subsidiaries) == 0 {
			// Debug: Check what companies actually exist
			var allCompanies []domain.CompanyModel
			db.Where("is_active = ?", true).Find(&allCompanies)
			zapLog.Error("No subsidiaries found",
				zap.Int("total_active_companies", len(allCompanies)),
				zap.Any("company_codes", func() []string {
					codes := make([]string, 0, len(allCompanies))
					for _, c := range allCompanies {
						codes = append(codes, c.Code)
					}
					return codes
				}()),
			)
			return fmt.Errorf("no subsidiaries found. please run seed-companies first. found %d active companies total", len(allCompanies))
		}
		zapLog.Info("Found subsidiaries using fallback query (by level)", zap.Int("count", len(subsidiaries)))
	} else {
		zapLog.Info("Found subsidiaries", zap.Int("count", len(subsidiaries)))
	}

	// Get all users for random assignment
	allUsers, err := uc.userRepo.GetAll()
	if err != nil {
		zapLog.Warn("Failed to get users, will use null inputter", zap.Error(err))
		allUsers = []domain.UserModel{}
	}

	// Periods to seed
	periods := []string{"2025-09", "2025-10", "2025-11", "2025-12"}

	// Note: As of Go 1.20, rand.Seed is deprecated and not needed
	// The global random number generator is automatically seeded

	totalCreated := 0

	// Create reports for each subsidiary
	for _, company := range subsidiaries {
		for _, period := range periods {
			// Check if report already exists
			existing, _ := uc.reportRepo.GetByCompanyIDAndPeriod(company.ID, period)
			if existing != nil {
				continue
			}

			// Generate realistic random data
			revenue := int64(rand.Intn(450000000) + 50000000)
			opexPercent := float64(rand.Intn(40)+30) / 100.0
			opex := int64(float64(revenue) * opexPercent)
			profit := revenue - opex
			tax := int64(float64(profit) * 0.25)
			if profit < 0 {
				tax = 0
			}
			npat := profit - tax
			dividend := int64(0)
			if npat > 0 {
				dividendPercent := float64(rand.Intn(20)+10) / 100.0
				dividend = int64(float64(npat) * dividendPercent)
			}
			financialRatio := float64(revenue) / float64(opex)
			if financialRatio > 3.0 {
				financialRatio = 3.0
			}

			// Randomly assign inputter (or null)
			var inputterID *string
			if len(allUsers) > 0 && rand.Float32() > 0.3 {
				randomUser := allUsers[rand.Intn(len(allUsers))]
				inputterID = &randomUser.ID
			}

			// Create report
			report := &domain.ReportModel{
				ID:             uuid.GenerateUUID(),
				Period:         period,
				CompanyID:      company.ID,
				InputterID:     inputterID,
				Revenue:        revenue,
				Opex:           opex,
				NPAT:           npat,
				Dividend:       dividend,
				FinancialRatio: financialRatio,
				Attachment:     nil,
				Remark:         nil,
			}

			if err := uc.reportRepo.Create(report); err != nil {
				zapLog.Warn("Failed to create report", zap.String("company", company.ID), zap.String("period", period), zap.Error(err))
				continue
			}

			totalCreated++
		}
	}

	zapLog.Info("Report seeder completed successfully", zap.Int("reports_created", totalCreated))
	return nil
}

// CheckReportDataExists checks if report data already exists
func (uc *developmentUseCase) CheckReportDataExists() (bool, error) {
	count, err := uc.reportRepo.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RunAllSeeders runs all seeders in the correct order
// Order: 0. Roles & Permissions (if needed), 1. Company seeder, 2. Report seeder
func (uc *developmentUseCase) RunAllSeeders() error {
	zapLog := logger.GetLogger()

	// Step 0: Ensure roles exist (required for company seeder)
	zapLog.Info("Step 0: Ensuring roles and permissions exist...")
	checkDB := uc.db
	if checkDB == nil {
		checkDB = database.GetDB()
	}

	// Check if admin role exists
	var adminRole domain.RoleModel
	if err := checkDB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		zapLog.Warn("Admin role not found, this may cause company seeder to fail. Please ensure roles are seeded first.")
	} else {
		zapLog.Info("Admin role found, proceeding with company seeder")
	}

	// Step 1: Run company seeder first
	zapLog.Info("Step 1: Running company seeder...")
	alreadyExists, err := uc.RunSubsidiarySeeder()
	if err != nil {
		return fmt.Errorf("failed to run company seeder: %w", err)
	}
	if alreadyExists {
		zapLog.Info("Company seeder data already exists, skipping")
	} else {
		zapLog.Info("Company seeder completed successfully")
	}

	// Refresh database connection to ensure we see the latest data
	refreshDB := uc.db
	if refreshDB == nil {
		refreshDB = database.GetDB()
	}
	if sqlDB, sqlErr := refreshDB.DB(); sqlErr == nil {
		if err := sqlDB.Ping(); err != nil {
			zapLog.Warn("Failed to ping database connection", zap.Error(err))
		} else {
			zapLog.Info("Database connection refreshed before report seeder")
		}
	}

	// Step 2: Run report seeder (depends on companies)
	zapLog.Info("Step 2: Running report seeder...")
	err = uc.RunReportSeeder()
	if err != nil {
		if err.Error() == "report data already exists" {
			zapLog.Info("Report seeder data already exists, skipping")
		} else {
			return fmt.Errorf("failed to run report seeder: %w", err)
		}
	} else {
		zapLog.Info("Report seeder completed successfully")
	}

	zapLog.Info("All seeders completed successfully")
	return nil
}

// ResetAllSeededData resets all seeded data in reverse order
// Order: 1. Report data, 2. Company data
func (uc *developmentUseCase) ResetAllSeededData() error {
	zapLog := logger.GetLogger()

	// Step 1: Reset report data first (depends on companies)
	zapLog.Info("Step 1: Resetting report data...")
	err := uc.ResetReportData()
	if err != nil {
		return fmt.Errorf("failed to reset report data: %w", err)
	}
	zapLog.Info("Report data reset completed")

	// Step 2: Reset company data
	zapLog.Info("Step 2: Resetting company data...")
	err = uc.ResetSubsidiaryData()
	if err != nil {
		return fmt.Errorf("failed to reset company data: %w", err)
	}
	zapLog.Info("Company data reset completed")

	zapLog.Info("All seeded data reset completed successfully")
	return nil
}

// CheckAllSeederStatus checks the status of all seeders
func (uc *developmentUseCase) CheckAllSeederStatus() (map[string]bool, error) {
	status := make(map[string]bool)

	// Check company seeder status
	companyExists, err := uc.CheckSeederDataExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check company seeder status: %w", err)
	}
	status["company"] = companyExists

	// Check report seeder status
	reportExists, err := uc.CheckReportDataExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check report seeder status: %w", err)
	}
	status["report"] = reportExists

	return status, nil
}
