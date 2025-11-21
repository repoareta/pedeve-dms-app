package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/render"
)

const (
	LogTypeUserAction    = "user_action"
	LogTypeTechnicalError = "technical_error"
)

// LogTechnicalError logs a technical error to audit log
func LogTechnicalError(err error, r *http.Request, details map[string]interface{}) {
	if err == nil {
		return
	}

	// Get stack trace
	stackTrace := string(debug.Stack())

	// Build error details
	errorDetails := map[string]interface{}{
		"error":       err.Error(),
		"stack_trace": stackTrace,
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.RawQuery,
	}

	// Merge with additional details if provided
	for k, v := range details {
		errorDetails[k] = v
	}

	// Get user info if available (might be nil for unauthenticated errors)
	userID := ""
	username := "system"
	if userIDCtx := r.Context().Value(contextKeyUserID); userIDCtx != nil {
		userID = userIDCtx.(string)
	}
	if usernameCtx := r.Context().Value(contextKeyUsername); usernameCtx != nil {
		username = usernameCtx.(string)
	}

	// Get IP address and user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.UserAgent()

	// Log to audit log asynchronously
	LogActionAsync(userID, username, "system_error", "system", "", ipAddress, userAgent, "error", errorDetails)
}

// LogActionAsync is a helper that logs action asynchronously with log type
func LogActionAsync(userID, username, action, resource, resourceID, ipAddress, userAgent, status string, details map[string]interface{}) {
	go func() {
		// Determine log type based on action
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

		auditLog := AuditLog{
			ID:         GenerateUUID(),
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

		// Log asynchronously (non-blocking)
		_ = DB.Create(&auditLog).Error
	}()
}

// ErrorHandlerMiddleware logs technical errors to audit log
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the response
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(ww, r)

		// Log errors (4xx and 5xx status codes)
		if ww.statusCode >= 400 {
			// Extract error message from response body if possible
			details := map[string]interface{}{
				"status_code": ww.statusCode,
			}

			var errMsg string
			if ww.statusCode >= 500 {
				errMsg = "Server error: " + fmt.Sprint(ww.statusCode)
			} else {
				errMsg = "Client error: " + fmt.Sprint(ww.statusCode)
			}

			// Log technical error
			LogTechnicalError(fmt.Errorf("%s", errMsg), r, details)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoverMiddleware logs panics to audit log
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic to audit log
				panicErr := fmt.Errorf("panic: %v", err)
				details := map[string]interface{}{
					"type": "panic",
				}
				LogTechnicalError(panicErr, r, details)

				// Return error response
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{
					Error:   "internal_server_error",
					Message: "An unexpected error occurred",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

