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

// LogTechnicalError mencatat error teknis ke audit log
func LogTechnicalError(err error, r *http.Request, details map[string]interface{}) {
	if err == nil {
		return
	}

	// Ambil stack trace
	stackTrace := string(debug.Stack())

	// Buat detail error
	errorDetails := map[string]interface{}{
		"error":       err.Error(),
		"stack_trace": stackTrace,
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.RawQuery,
	}

	// Gabungkan dengan detail tambahan jika diberikan
	for k, v := range details {
		errorDetails[k] = v
	}

	// Ambil info user jika tersedia (mungkin nil untuk error yang tidak terautentikasi)
	userID := ""
	username := "system"
	if userIDCtx := r.Context().Value(contextKeyUserID); userIDCtx != nil {
		userID = userIDCtx.(string)
	}
	if usernameCtx := r.Context().Value(contextKeyUsername); usernameCtx != nil {
		username = usernameCtx.(string)
	}

	// Ambil alamat IP dan user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.UserAgent()

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

		// Log secara asinkron (non-blocking)
		_ = DB.Create(&auditLog).Error
	}()
}

// ErrorHandlerMiddleware mencatat error teknis ke audit log
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tangkap response
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Panggil handler berikutnya
		next.ServeHTTP(ww, r)

		// Log error (kode status 4xx dan 5xx)
		if ww.statusCode >= 400 {
			// Ekstrak pesan error dari response body jika memungkinkan
			details := map[string]interface{}{
				"status_code": ww.statusCode,
			}

			var errMsg string
			if ww.statusCode >= 500 {
				errMsg = "Server error: " + fmt.Sprint(ww.statusCode)
			} else {
				errMsg = "Client error: " + fmt.Sprint(ww.statusCode)
			}

			// Log error teknis
			LogTechnicalError(fmt.Errorf("%s", errMsg), r, details)
		}
	})
}

// responseWriter membungkus http.ResponseWriter untuk menangkap kode status
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoverMiddleware mencatat panic ke audit log
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic ke audit log
				panicErr := fmt.Errorf("panic: %v", err)
				details := map[string]interface{}{
					"type": "panic",
				}
				LogTechnicalError(panicErr, r, details)

				// Kembalikan response error
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

