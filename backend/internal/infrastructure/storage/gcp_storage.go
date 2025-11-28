package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// GCPStorageManager menggunakan Google Cloud Storage
type GCPStorageManager struct {
	client    *storage.Client
	bucketName string
	ctx       context.Context
}

// NewGCPStorageManager membuat instance baru GCPStorageManager
func NewGCPStorageManager(bucketName string) (*GCPStorageManager, error) {
	ctx := context.Background()
	
	// Buat client Storage
	// Jika running di GCP (VM), akan otomatis menggunakan default credentials
	// Jika running di local, bisa pakai Application Default Credentials (ADC)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Storage client: %w", err)
	}

	return &GCPStorageManager{
		client:     client,
		bucketName: bucketName,
		ctx:        ctx,
	}, nil
}

// NewGCPStorageManagerWithCredentials membuat instance dengan custom credentials (untuk testing)
func NewGCPStorageManagerWithCredentials(bucketName string, credentialsJSON []byte) (*GCPStorageManager, error) {
	ctx := context.Background()
	
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create Storage client with credentials: %w", err)
	}

	return &GCPStorageManager{
		client:     client,
		bucketName: bucketName,
		ctx:        ctx,
	}, nil
}

// GetBucketName returns the bucket name
func (g *GCPStorageManager) GetBucketName() string {
	return g.bucketName
}

// GetClient returns the storage client (for internal use)
func (g *GCPStorageManager) GetClient() *storage.Client {
	return g.client
}

// GetContext returns the context (for internal use)
func (g *GCPStorageManager) GetContext() context.Context {
	return g.ctx
}

// Close menutup koneksi ke Storage client
func (g *GCPStorageManager) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// UploadFile uploads a file to GCP Cloud Storage
// bucketPath: path dalam bucket (e.g., "logos", "documents")
// filename: nama file (e.g., "logo.png")
// data: file content sebagai byte array
// contentType: MIME type (e.g., "image/png")
// Returns: public URL untuk file yang di-upload
func (g *GCPStorageManager) UploadFile(bucketPath string, filename string, data []byte, contentType string) (string, error) {
	zapLog := logger.GetLogger()

	// Construct object path dalam bucket
	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)

	// Get bucket handle
	bucket := g.client.Bucket(g.bucketName)
	
	// Create object writer
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(g.ctx)
	
	// Set content type
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=3600" // Cache for 1 hour
	
	// Write data
	if _, err := writer.Write(data); err != nil {
		writer.Close()
		zapLog.Error("Failed to write file to GCP Storage",
			zap.String("bucket", g.bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Close writer (this finalizes the upload)
	if err := writer.Close(); err != nil {
		zapLog.Error("Failed to close GCP Storage writer",
			zap.String("bucket", g.bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Make object publicly readable
	if err := obj.ACL().Set(g.ctx, storage.AllUsers, storage.RoleReader); err != nil {
		zapLog.Warn("Failed to set public ACL on object (may need bucket-level IAM instead)",
			zap.String("bucket", g.bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		// Continue anyway, as bucket might use Uniform access control
	}

	// Generate public URL
	// Format: https://storage.googleapis.com/{bucket}/{object_path}
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)

	zapLog.Info("File uploaded successfully to GCP Storage",
		zap.String("bucket", g.bucketName),
		zap.String("object_path", objectPath),
		zap.String("url", publicURL),
		zap.Int("size", len(data)),
	)

	return publicURL, nil
}

// DeleteFile deletes a file from GCP Cloud Storage
func (g *GCPStorageManager) DeleteFile(bucketPath string, filename string) error {
	zapLog := logger.GetLogger()

	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)
	
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)
	
	if err := obj.Delete(g.ctx); err != nil {
		zapLog.Error("Failed to delete file from GCP Storage",
			zap.String("bucket", g.bucketName),
			zap.String("object_path", objectPath),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	zapLog.Info("File deleted successfully from GCP Storage",
		zap.String("bucket", g.bucketName),
		zap.String("object_path", objectPath),
	)

	return nil
}

// GetFileURL returns the public URL for a file
func (g *GCPStorageManager) GetFileURL(bucketPath string, filename string) (string, error) {
	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)
	return publicURL, nil
}

// GetSignedURL returns a signed URL for temporary access (if bucket is private)
// expiresIn: duration until URL expires (e.g., 1 hour)
// Note: For now, returns public URL. If bucket is private, implement proper signed URL
// generation with service account credentials
func (g *GCPStorageManager) GetSignedURL(bucketPath string, filename string, expiresIn time.Duration) (string, error) {
	// Return public URL (bucket should be configured for public read access to uploaded files)
	// Signed URL requires service account private key which is complex to manage
	// If bucket is private, configure bucket IAM to allow public read access to specific objects
	return g.GetFileURL(bucketPath, filename)
}

// FileExists checks if a file exists in GCP Cloud Storage
func (g *GCPStorageManager) FileExists(bucketPath string, filename string) (bool, error) {
	objectPath := fmt.Sprintf("%s/%s", bucketPath, filename)
	
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)
	
	_, err := obj.Attrs(g.ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// LocalStorageManager menggunakan local filesystem (untuk development/fallback)
type LocalStorageManager struct {
	basePath string
}

// NewLocalStorageManager membuat instance baru LocalStorageManager
func NewLocalStorageManager(basePath string) *LocalStorageManager {
	return &LocalStorageManager{
		basePath: basePath,
	}
}

// UploadFile uploads a file to local filesystem
func (l *LocalStorageManager) UploadFile(bucketPath string, filename string, data []byte, contentType string) (string, error) {
	zapLog := logger.GetLogger()

	// Create directory if not exists
	dirPath := fmt.Sprintf("%s/%s", l.basePath, bucketPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		zapLog.Error("Failed to create directory", zap.Error(err))
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	filePath := fmt.Sprintf("%s/%s/%s", l.basePath, bucketPath, filename)
	file, err := os.Create(filePath)
	if err != nil {
		zapLog.Error("Failed to create file", zap.Error(err))
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		zapLog.Error("Failed to write file", zap.Error(err))
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative URL
	url := fmt.Sprintf("/%s/%s", bucketPath, filename)
	return url, nil
}

// DeleteFile deletes a file from local filesystem
func (l *LocalStorageManager) DeleteFile(bucketPath string, filename string) error {
	filePath := fmt.Sprintf("%s/%s/%s", l.basePath, bucketPath, filename)
	return os.Remove(filePath)
}

// GetFileURL returns the relative URL for a file
func (l *LocalStorageManager) GetFileURL(bucketPath string, filename string) (string, error) {
	url := fmt.Sprintf("/%s/%s", bucketPath, filename)
	return url, nil
}

// FileExists checks if a file exists in local filesystem
func (l *LocalStorageManager) FileExists(bucketPath string, filename string) (bool, error) {
	filePath := fmt.Sprintf("%s/%s/%s", l.basePath, bucketPath, filename)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetBasePath returns the base path for local storage (for internal use)
func (l *LocalStorageManager) GetBasePath() string {
	return l.basePath
}

// GetStorageManager mengembalikan StorageManager berdasarkan konfigurasi
// Priority:
// 1. GCP Cloud Storage (jika GCP_STORAGE_ENABLED=true dan GCP_STORAGE_BUCKET set)
// 2. Local filesystem (fallback untuk development)
func GetStorageManager() (StorageManager, error) {
	zapLog := logger.GetLogger()

	// Cek apakah GCP Storage dikonfigurasi
	gcpEnabled := os.Getenv("GCP_STORAGE_ENABLED")
	gcpBucket := os.Getenv("GCP_STORAGE_BUCKET")

	if gcpEnabled == "true" && gcpBucket != "" {
		gcpManager, err := NewGCPStorageManager(gcpBucket)
		if err != nil {
			zapLog.Warn("Failed to initialize GCP Storage, falling back to local storage",
				zap.Error(err),
			)
		} else {
			zapLog.Info("Using GCP Cloud Storage for file storage",
				zap.String("bucket", gcpBucket),
			)
			return gcpManager, nil
		}
	}

	// Fallback ke local filesystem
	basePath := os.Getenv("UPLOAD_BASE_PATH")
	if basePath == "" {
		basePath = "uploads" // Default path
	}

	zapLog.Info("Using local filesystem for file storage",
		zap.String("base_path", basePath),
	)

	return NewLocalStorageManager(basePath), nil
}

