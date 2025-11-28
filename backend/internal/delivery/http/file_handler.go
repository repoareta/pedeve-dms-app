package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/storage"
	"github.com/gofiber/fiber/v2"
	storageClient "cloud.google.com/go/storage"
	"go.uber.org/zap"
)

// ServeFile serves a file from GCP Storage or local storage as a proxy
// This allows frontend to access files without requiring public access to the bucket
// @Summary      Serve File
// @Description  Serve file dari storage (GCP Storage atau local) sebagai proxy. Endpoint ini public untuk memungkinkan frontend mengakses file.
// @Tags         Files
// @Accept       json
// @Produce      image/png,image/jpeg,image/jpg,application/octet-stream
// @Param        path  path      string  true  "File path (e.g., logos/filename.png)"
// @Success      200   {file}    file    "File content"
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /api/v1/files/{path} [get]
func ServeFile(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()

	// Get file path from URL parameter (wildcard route captures everything after /files/)
	// For route /api/v1/files/*, use c.Params("*") to get the wildcard value
	filePath := c.Params("*")
	if filePath == "" {
		// Fallback: try regular param if wildcard doesn't work
		filePath = c.Params("path")
	}
	
	if filePath == "" {
		zapLog.Warn("File path is empty in request",
			zap.String("url", c.OriginalURL()),
			zap.String("path", c.Path()),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File path tidak ditemukan dalam URL",
		})
	}

	// Decode URL-encoded path (handles %20 for spaces, %2F for slashes, etc.)
	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		zapLog.Warn("Failed to decode URL path, using original",
			zap.String("original_path", filePath),
			zap.Error(err),
		)
		// Continue with original path if decoding fails
		decodedPath = filePath
	}

	// Log for debugging
	zapLog.Debug("Serving file request",
		zap.String("original_path", filePath),
		zap.String("decoded_path", decodedPath),
		zap.String("url", c.OriginalURL()),
	)

	// Sanitize file path (prevent directory traversal)
	// Note: filepath.Clean will normalize the path but won't remove ".." if it's part of a valid path
	// So we check for ".." before cleaning
	if strings.Contains(decodedPath, "..") {
		zapLog.Warn("Directory traversal attempt detected",
			zap.String("path", decodedPath),
			zap.String("ip", c.IP()),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File path tidak valid: directory traversal tidak diizinkan",
		})
	}

	// Clean the path (normalize separators, remove redundant elements)
	decodedPath = filepath.Clean(decodedPath)

	// Get storage manager
	storageManager, err := storage.GetStorageManager()
	if err != nil {
		zapLog.Error("Failed to initialize storage manager",
			zap.String("file_path", decodedPath),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal menginisialisasi storage manager",
		})
	}

	// Extract bucket path and filename from filePath
	// filePath format: "logos/filename.png" or "documents/file.pdf"
	parts := strings.SplitN(decodedPath, "/", 2)
	if len(parts) != 2 {
		zapLog.Warn("Invalid file path format",
			zap.String("file_path", decodedPath),
			zap.String("expected_format", "bucketPath/filename"),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": fmt.Sprintf("File path harus dalam format: bucketPath/filename. Diterima: %s", decodedPath),
		})
	}
	bucketPath := parts[0]
	filename := parts[1]

	// Validate bucket path and filename are not empty
	if bucketPath == "" || filename == "" {
		zapLog.Warn("Empty bucket path or filename",
			zap.String("bucket_path", bucketPath),
			zap.String("filename", filename),
			zap.String("decoded_path", decodedPath),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Bucket path dan filename tidak boleh kosong",
		})
	}

	zapLog.Info("Processing file serve request",
		zap.String("bucket_path", bucketPath),
		zap.String("filename", filename),
		zap.String("full_path", decodedPath),
	)

	// Check if using GCP Storage
	// Use type assertion with interface check
	type GCPStorageGetter interface {
		GetBucketName() string
		GetClient() *storageClient.Client
		GetContext() context.Context
	}
	
	if gcpStorage, ok := storageManager.(GCPStorageGetter); ok {
		// Serve from GCP Storage
		return serveFromGCPStorage(c, gcpStorage, bucketPath, filename, decodedPath)
	}

	// Serve from local storage
	return serveFromLocalStorage(c, storageManager, bucketPath, filename, decodedPath)
}

// serveFromGCPStorage serves file from GCP Storage
func serveFromGCPStorage(c *fiber.Ctx, gcpStorage interface {
	GetBucketName() string
	GetClient() *storageClient.Client
	GetContext() context.Context
}, bucketPath, filename, fullPath string) error {
	zapLog := logger.GetLogger()

	// Get GCP Storage client and bucket name from GCPStorageManager
	ctx := gcpStorage.GetContext()
	bucketName := gcpStorage.GetBucketName()
	client := gcpStorage.GetClient()

	if bucketName == "" {
		zapLog.Error("Bucket name not found in GCP Storage manager",
			zap.String("file_path", fullPath),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Bucket name tidak ditemukan di konfigurasi storage",
		})
	}

	// Construct object path in bucket
	// Format: bucketPath/filename (e.g., "logos/1764336646_man photo.jpg")
	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)

	zapLog.Info("Attempting to serve file from GCP Storage",
		zap.String("bucket", bucketName),
		zap.String("object_path", objectPath),
		zap.String("bucket_path", bucketPath),
		zap.String("filename", filename),
	)

	// Get object from bucket
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectPath)
	
	// Check if object exists first (optional, but provides better error message)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storageClient.ErrObjectNotExist {
			zapLog.Warn("File not found in GCP Storage",
				zap.String("bucket", bucketName),
				zap.String("object_path", objectPath),
				zap.String("requested_path", fullPath),
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": fmt.Sprintf("File tidak ditemukan di storage: %s/%s", bucketPath, filename),
			})
		}
		zapLog.Error("Failed to check file existence in GCP Storage",
			zap.String("bucket", bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mengecek file di storage",
		})
	}

	// Open object for reading
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storageClient.ErrObjectNotExist {
			zapLog.Warn("File not found when opening reader",
				zap.String("bucket", bucketName),
				zap.String("object_path", objectPath),
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": fmt.Sprintf("File tidak ditemukan: %s/%s", bucketPath, filename),
			})
		}
		zapLog.Error("Failed to open file reader from GCP Storage",
			zap.String("bucket", bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal membaca file dari storage",
		})
	}
	defer reader.Close()

	// Get content type from object attributes
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		zapLog.Warn("Failed to get object attributes, using default content type",
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
	}

	contentType := "application/octet-stream"
	if attrs != nil && attrs.ContentType != "" {
		contentType = attrs.ContentType
	}

	zapLog.Info("Serving file from GCP Storage",
		zap.String("object_path", objectPath),
		zap.String("content_type", contentType),
		zap.Int64("size", attrs.Size),
	)

	// Set headers
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=3600")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Stream file content to response
	bytesWritten, err := io.Copy(c.Response().BodyWriter(), reader)
	if err != nil {
		zapLog.Error("Failed to stream file content",
			zap.String("object_path", objectPath),
			zap.Int64("bytes_written", bytesWritten),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mengirim file ke client",
		})
	}

	zapLog.Info("File served successfully",
		zap.String("object_path", objectPath),
		zap.Int64("bytes_written", bytesWritten),
	)

	return nil
}

// serveFromLocalStorage serves file from local filesystem
func serveFromLocalStorage(c *fiber.Ctx, storageManager storage.StorageManager, bucketPath, filename, fullPath string) error {
	zapLog := logger.GetLogger()

	zapLog.Info("Attempting to serve file from local storage",
		zap.String("bucket_path", bucketPath),
		zap.String("filename", filename),
		zap.String("full_path", fullPath),
	)

	// Check if file exists first
	exists, err := storageManager.FileExists(bucketPath, filename)
	if err != nil {
		zapLog.Error("Failed to check file existence in local storage",
			zap.String("bucket_path", bucketPath),
			zap.String("filename", filename),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mengecek keberadaan file",
		})
	}

	if !exists {
		zapLog.Warn("File not found in local storage",
			zap.String("bucket_path", bucketPath),
			zap.String("filename", filename),
			zap.String("requested_path", fullPath),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": fmt.Sprintf("File tidak ditemukan: %s/%s", bucketPath, filename),
		})
	}

	// Get LocalStorageManager to access basePath
	type LocalStorageGetter interface {
		GetBasePath() string
	}
	
	localStorage, ok := storageManager.(LocalStorageGetter)
	if !ok {
		// Fallback: try to read file using GetFileURL and read from filesystem
		// This is less efficient but works as fallback
		fileURL, err := storageManager.GetFileURL(bucketPath, filename)
		if err != nil {
			zapLog.Error("Failed to get file URL from local storage",
				zap.String("bucket_path", bucketPath),
				zap.String("filename", filename),
				zap.Error(err),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "Gagal mendapatkan URL file",
			})
		}
		
		// Redirect to static file server (fallback)
		// Note: This requires static file server to be configured correctly
		zapLog.Info("Redirecting to static file server (fallback)",
			zap.String("file_url", fileURL),
		)
		return c.Redirect(fileURL, fiber.StatusTemporaryRedirect)
	}

	// Read file directly from filesystem and serve
	basePath := localStorage.GetBasePath()
	filePath := fmt.Sprintf("%s/%s/%s", basePath, bucketPath, filename)
	
	zapLog.Info("Reading file from local filesystem",
		zap.String("file_path", filePath),
	)

	// Read file content
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			zapLog.Warn("File not found in filesystem",
				zap.String("file_path", filePath),
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": fmt.Sprintf("File tidak ditemukan: %s/%s", bucketPath, filename),
			})
		}
		zapLog.Error("Failed to read file from filesystem",
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal membaca file",
		})
	}

	// Detect content type
	contentType := http.DetectContentType(fileData)
	if contentType == "application/octet-stream" {
		// Try to detect from file extension
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		}
	}

	zapLog.Info("Serving file from local storage",
		zap.String("file_path", filePath),
		zap.String("content_type", contentType),
		zap.Int("size", len(fileData)),
	)

	// Set headers
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=3600")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Send file content
	return c.Send(fileData)
}

