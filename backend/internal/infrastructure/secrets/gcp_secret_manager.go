package secrets

import (
	"context"
	"fmt"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// GCPSecretManager menggunakan Google Cloud Secret Manager
type GCPSecretManager struct {
	projectID string
	client    *secretmanager.Client
	ctx       context.Context
}

// NewGCPSecretManager membuat instance baru GCPSecretManager
func NewGCPSecretManager(projectID string) (*GCPSecretManager, error) {
	ctx := context.Background()
	
	// Buat client Secret Manager
	// Jika running di GCP (VM), akan otomatis menggunakan default credentials
	// Jika running di local, bisa pakai Application Default Credentials (ADC)
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}

	return &GCPSecretManager{
		projectID: projectID,
		client:    client,
		ctx:       ctx,
	}, nil
}

// NewGCPSecretManagerWithCredentials membuat instance dengan custom credentials (untuk testing)
func NewGCPSecretManagerWithCredentials(projectID string, credentialsJSON []byte) (*GCPSecretManager, error) {
	ctx := context.Background()
	
	client, err := secretmanager.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client with credentials: %w", err)
	}

	return &GCPSecretManager{
		projectID: projectID,
		client:    client,
		ctx:       ctx,
	}, nil
}

// Close menutup koneksi ke Secret Manager client
func (g *GCPSecretManager) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// GetSecret membaca secret dari GCP Secret Manager
// Format secret name: "secret-name" (tanpa project path)
// Akan otomatis di-construct menjadi: projects/{project}/secrets/{name}/versions/latest
func (g *GCPSecretManager) GetSecret(secretName string) (string, error) {
	zapLog := logger.GetLogger()

	// Construct full secret name
	// Format: projects/{project_id}/secrets/{secret_name}/versions/latest
	secretPath := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", g.projectID, secretName)

	// Access secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := g.client.AccessSecretVersion(g.ctx, req)
	if err != nil {
		zapLog.Error("Failed to access secret from GCP Secret Manager",
			zap.String("secret_name", secretName),
			zap.String("secret_path", secretPath),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to access secret '%s': %w", secretName, err)
	}

	secretValue := string(result.Payload.Data)
	if secretValue == "" {
		return "", fmt.Errorf("secret '%s' is empty", secretName)
	}

	zapLog.Debug("Successfully retrieved secret from GCP Secret Manager",
		zap.String("secret_name", secretName),
		zap.Int("value_length", len(secretValue)),
	)

	return secretValue, nil
}

// GetEncryptionKey mendapatkan encryption key dari GCP Secret Manager
// Mencari secret dengan nama "encryption_key"
func (g *GCPSecretManager) GetEncryptionKey() (string, error) {
	zapLog := logger.GetLogger()

	// Coba ambil dari secret "encryption_key"
	key, err := g.GetSecret("encryption_key")
	if err != nil {
		zapLog.Error("Failed to get encryption_key from GCP Secret Manager", zap.Error(err))
		return "", fmt.Errorf("failed to get encryption_key: %w", err)
	}

	if key == "" {
		return "", fmt.Errorf("encryption_key is empty")
	}

	zapLog.Info("Successfully retrieved encryption key from GCP Secret Manager",
		zap.Int("key_length", len(key)),
	)

	return key, nil
}

// GetSecretWithVersion membaca secret dengan versi tertentu
// Jika versionID adalah "latest" atau kosong, akan menggunakan versi terbaru
func (g *GCPSecretManager) GetSecretWithVersion(secretName string, versionID string) (string, error) {
	if versionID == "" || versionID == "latest" {
		return g.GetSecret(secretName)
	}

	zapLog := logger.GetLogger()
	secretPath := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", g.projectID, secretName, versionID)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := g.client.AccessSecretVersion(g.ctx, req)
	if err != nil {
		zapLog.Error("Failed to access secret version from GCP Secret Manager",
			zap.String("secret_name", secretName),
			zap.String("version", versionID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to access secret '%s' version '%s': %w", secretName, versionID, err)
	}

	return string(result.Payload.Data), nil
}

