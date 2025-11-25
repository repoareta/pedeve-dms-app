package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UploadFileResponse response untuk upload file
type UploadFileResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// UploadLogo handles logo upload
// @Summary      Upload Logo Perusahaan
// @Description  Upload logo perusahaan dengan validasi format (PNG, JPG, JPEG) dan ukuran maksimal 5MB
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "File logo (PNG, JPG, JPEG, max 5MB)"
// @Success      200   {object}  UploadFileResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      413   {object}  ErrorResponse
// @Router       /api/v1/upload/logo [post]
func UploadLogo(c *fiber.Ctx) error {
	zapLog := logger.GetLogger()

	// Ambil file dari form
	file, err := c.FormFile("file")
	if err != nil {
		zapLog.Warn("Failed to get file from form", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "File tidak ditemukan dalam request",
		})
	}

	// Validasi ukuran file (max 5MB)
	const maxSize = 5 * 1024 * 1024 // 5MB
	if file.Size > maxSize {
		zapLog.Warn("File too large",
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size),
			zap.Int64("max_size", maxSize),
		)
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error":   "file_too_large",
			"message": "Ukuran file melebihi 5MB",
		})
	}

	// Validasi format file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".png", ".jpg", ".jpeg"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		zapLog.Warn("Invalid file format",
			zap.String("filename", file.Filename),
			zap.String("ext", ext),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_file_format",
			"message": "Format file tidak diizinkan. Hanya PNG, JPG, dan JPEG yang diperbolehkan",
		})
	}

	// Buka file untuk validasi MIME type
	src, err := file.Open()
	if err != nil {
		zapLog.Error("Failed to open uploaded file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Gagal membaca file",
		})
	}
	defer src.Close()

	// Baca header file untuk validasi MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		zapLog.Error("Failed to read file header", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Gagal membaca file",
		})
	}

	// Reset file pointer
	src.Seek(0, 0)

	// Validasi MIME type
	mimeType := http.DetectContentType(buffer)
	allowedMimeTypes := []string{
		"image/png",
		"image/jpeg",
		"image/jpg",
	}
	isValidMime := false
	for _, allowedMime := range allowedMimeTypes {
		if strings.HasPrefix(mimeType, allowedMime) || mimeType == allowedMime {
			isValidMime = true
			break
		}
	}

	if !isValidMime {
		zapLog.Warn("Invalid MIME type",
			zap.String("filename", file.Filename),
			zap.String("mime_type", mimeType),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_file_format",
			"message": "Format file tidak valid. Hanya gambar PNG, JPG, dan JPEG yang diperbolehkan",
		})
	}

	// Buat direktori uploads jika belum ada
	uploadDir := "uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		zapLog.Error("Failed to create upload directory", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Gagal membuat direktori upload",
		})
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Simpan file
	dst, err := os.Create(filepath)
	if err != nil {
		zapLog.Error("Failed to create file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Gagal menyimpan file",
		})
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		zapLog.Error("Failed to save file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Gagal menyimpan file",
		})
	}

	// Generate URL untuk file (relative path)
	fileURL := fmt.Sprintf("/uploads/logos/%s", filename)

	zapLog.Info("Logo uploaded successfully",
		zap.String("filename", filename),
		zap.Int64("size", file.Size),
		zap.String("url", fileURL),
	)

	return c.JSON(UploadFileResponse{
		URL:      fileURL,
		Filename: filename,
		Size:     file.Size,
	})
}

