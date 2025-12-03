package middleware

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/repoareta/pedeve-dms-app/backend/internal/domain"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/database"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/uuid"
)

const (
	LogTypeUserAction     = "user_action"
	LogTypeTechnicalError = "technical_error"
)

// isSuspiciousPath mengecek apakah path adalah request mencurigakan (bot/scanner)
// Request ini biasanya tidak perlu di-log sebagai technical error untuk mengurangi noise
func isSuspiciousPath(path string) bool {
	suspiciousPatterns := []string{
		"/boaform/",       // Router/device admin panels
		"/wp-admin/",      // WordPress admin
		"/wp-login.php",   // WordPress login
		"/.env",           // Environment file access
		"/.git/",          // Git directory access
		"/phpmyadmin/",    // phpMyAdmin
		"/admin.php",      // Generic admin panels
		"/administrator/", // Joomla admin
		"/manager/",       // Tomcat manager
		"/solr/",          // Apache Solr
		"/actuator/",      // Spring Boot actuator
		"/.well-known/",   // Well-known paths (some are legit, but many bots scan this)
		"/phpinfo",        // PHP info
		"/config.php",     // Config files
		"/web.config",     // IIS config
	}

	for _, pattern := range suspiciousPatterns {
		if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
			return true
		}
	}

	return false
}

// LogTechnicalError mencatat error teknis ke audit log (untuk Fiber)
func LogTechnicalError(err error, c *fiber.Ctx, details map[string]interface{}) {
	if err == nil {
		return
	}

	// Jika path mencurigakan, abaikan seluruh logging (apapun status code-nya)
	if isSuspiciousPath(c.Path()) {
		return
	}

	// Ambil stack trace
	stackTrace := string(debug.Stack())

	// Buat detail error
	errorDetails := map[string]interface{}{
		"error":       err.Error(),
		"stack_trace": stackTrace,
		"method":      c.Method(),
		"path":        c.Path(),
		"query":       c.Queries(),
	}

	// Gabungkan dengan detail tambahan jika diberikan
	for k, v := range details {
		errorDetails[k] = v
	}

	// Ambil info user jika tersedia (mungkin nil untuk error yang tidak terautentikasi)
	userID := ""
	username := "system"
	if userIDVal := c.Locals("userID"); userIDVal != nil {
		userID = userIDVal.(string)
	}
	if usernameVal := c.Locals("username"); usernameVal != nil {
		username = usernameVal.(string)
	}

	// Ambil alamat IP dan user agent
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := c.Get("User-Agent")

	// Log ke audit log secara asinkron
	LogActionAsync(userID, username, "system_error", "system", "", ipAddress, userAgent, "error", errorDetails)
}

// LogActionAsync adalah helper yang mencatat aksi secara asinkron dengan tipe log
func LogActionAsync(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	go func() {
		// Tentukan tipe log berdasarkan aksi
		logType := LogTypeUserAction
		if action == "system_error" || action == "database_error" || action == "validation_error" {
			logType = LogTypeTechnicalError
		}

		detailsJSON := ""
		if details != nil {
			jsonData, err := json.Marshal(details)
			if err == nil {
				detailsJSON = string(jsonData)
			}
		}

		auditLog := domain.AuditLog{
			ID:         uuid.GenerateUUID(),
			UserID:     userID,
			Username:   username,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Details:    detailsJSON,
			Status:     status,
			LogType:    logType,
			CreatedAt:  time.Now(),
		}

		// Log secara asinkron (non-blocking)
		_ = database.GetDB().Create(&auditLog).Error
	}()
}

// ErrorHandlerMiddleware mencatat error teknis ke audit log (untuk Fiber)
func ErrorHandlerMiddleware(c *fiber.Ctx) error {
	err := c.Next()

	// Log error setelah handler selesai (jika ada error atau status code >= 400)
	statusCode := c.Response().StatusCode()
	if statusCode >= 400 {
		// Skip error logging untuk path mencurigakan (bot/scanner) agar audit log tidak penuh noise
		if isSuspiciousPath(c.Path()) {
			return err
		}

		// Ekstrak pesan error dari response body jika memungkinkan
		details := map[string]interface{}{
			"status_code": statusCode,
		}

		var errMsg string
		if statusCode >= 500 {
			errMsg = "Server error: " + fmt.Sprint(statusCode)
		} else {
			errMsg = "Client error: " + fmt.Sprint(statusCode)
		}

		// Log error teknis
		LogTechnicalError(fmt.Errorf("%s", errMsg), c, details)
	}

	return err
}

// RecoverMiddleware mencatat panic ke audit log (untuk Fiber)
func RecoverMiddleware(c *fiber.Ctx) error {
	defer func() {
		if err := recover(); err != nil {
			// Log panic ke audit log
			panicErr := fmt.Errorf("panic: %v", err)
			details := map[string]interface{}{
				"type": "panic",
			}
			LogTechnicalError(panicErr, c, details)

			// Kembalikan response error
			_ = c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{
				Error:   "internal_server_error",
				Message: "An unexpected error occurred",
			})
		}
	}()

	return c.Next()
}

// LogAction adalah fungsi helper untuk mencatat aksi (wrapper untuk repository)
func LogAction(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	LogActionAsync(userID, username, action, resource, resourceID, ipAddress, userAgent, status, details)
}
