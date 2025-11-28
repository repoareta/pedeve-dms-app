package http

import (
	"context"
	"fmt"
	"io"
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

	// Get file path from URL parameter
	filePath := c.Params("path")
	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File path tidak ditemukan",
		})
	}

	// Sanitize file path (prevent directory traversal)
	filePath = filepath.Clean(filePath)
	if strings.Contains(filePath, "..") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File path tidak valid",
		})
	}

	// Get storage manager
	storageManager, err := storage.GetStorageManager()
	if err != nil {
		zapLog.Error("Failed to initialize storage manager", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal menginisialisasi storage",
		})
	}

	// Extract bucket path and filename from filePath
	// filePath format: "logos/filename.png" or "documents/file.pdf"
	parts := strings.SplitN(filePath, "/", 2)
	if len(parts) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File path harus dalam format: bucketPath/filename",
		})
	}
	bucketPath := parts[0]
	filename := parts[1]

	// Check if using GCP Storage
	// Use type assertion with interface check
	type GCPStorageGetter interface {
		GetBucketName() string
		GetClient() *storageClient.Client
		GetContext() context.Context
	}
	
	if gcpStorage, ok := storageManager.(GCPStorageGetter); ok {
		// Serve from GCP Storage
		return serveFromGCPStorage(c, gcpStorage, bucketPath, filename)
	}

	// Serve from local storage
	return serveFromLocalStorage(c, storageManager, bucketPath, filename)
}

// serveFromGCPStorage serves file from GCP Storage
func serveFromGCPStorage(c *fiber.Ctx, gcpStorage interface {
	GetBucketName() string
	GetClient() *storageClient.Client
	GetContext() context.Context
}, bucketPath, filename string) error {
	zapLog := logger.GetLogger()

	// Get GCP Storage client and bucket name from GCPStorageManager
	ctx := gcpStorage.GetContext()
	bucketName := gcpStorage.GetBucketName()
	client := gcpStorage.GetClient()

	if bucketName == "" {
		zapLog.Error("Bucket name not found in GCP Storage manager")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Bucket name tidak ditemukan",
		})
	}

	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)

	// Get object from bucket
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectPath)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storageClient.ErrObjectNotExist {
			zapLog.Warn("File not found in GCP Storage",
				zap.String("object_path", objectPath),
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "File tidak ditemukan",
			})
		}
		zapLog.Error("Failed to read file from GCP Storage",
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal membaca file",
		})
	}
	defer reader.Close()

	// Get content type
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		zapLog.Warn("Failed to get object attributes, using default content type",
			zap.Error(err),
		)
	}

	contentType := "application/octet-stream"
	if attrs != nil && attrs.ContentType != "" {
		contentType = attrs.ContentType
	}

	// Set headers
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=3600")

	// Stream file content to response
	_, err = io.Copy(c.Response().BodyWriter(), reader)
	if err != nil {
		zapLog.Error("Failed to stream file content", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mengirim file",
		})
	}

	return nil
}

// serveFromLocalStorage serves file from local filesystem
func serveFromLocalStorage(c *fiber.Ctx, storageManager storage.StorageManager, bucketPath, filename string) error {
	zapLog := logger.GetLogger()

	// For local storage, we can use the existing static file server
	// But for consistency, we'll read and serve the file
	fileURL, err := storageManager.GetFileURL(bucketPath, filename)
	if err != nil {
		zapLog.Error("Failed to get file URL", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mendapatkan file URL",
		})
	}

	// Check if file exists
	exists, err := storageManager.FileExists(bucketPath, filename)
	if err != nil {
		zapLog.Error("Failed to check file existence", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "Gagal mengecek file",
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "File tidak ditemukan",
		})
	}

	// For local storage, redirect to static file server
	// Or read and serve the file directly
	// Since we're using Fiber's static file server for local files,
	// we can just redirect to the static path
	return c.Redirect(fileURL, fiber.StatusTemporaryRedirect)
}

