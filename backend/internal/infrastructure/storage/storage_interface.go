package storage

// StorageManager interface untuk file storage management
// Support multiple backends: Local filesystem, GCP Cloud Storage, dll
type StorageManager interface {
	// UploadFile uploads a file and returns the public URL
	UploadFile(bucketPath string, filename string, data []byte, contentType string) (string, error)
	
	// DeleteFile deletes a file from storage
	DeleteFile(bucketPath string, filename string) error
	
	// GetFileURL returns the public URL for a file
	GetFileURL(bucketPath string, filename string) (string, error)
	
	// FileExists checks if a file exists in storage
	FileExists(bucketPath string, filename string) (bool, error)
}

