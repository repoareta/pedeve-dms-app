package secrets

import (
	"fmt"
	"os"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// SecretManager interface untuk key management
// Support multiple backends: Vault, Environment Variable, dll
type SecretManager interface {
	GetEncryptionKey() (string, error)
}

// EnvSecretManager menggunakan environment variable
type EnvSecretManager struct{}

func (e *EnvSecretManager) GetEncryptionKey() (string, error) {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY not set")
	}
	return key, nil
}

// VaultSecretManager menggunakan HashiCorp Vault
type VaultSecretManager struct {
	address string
	token   string
	path    string // Path ke secret di Vault (e.g., "secret/data/dms-app")
}

func NewVaultSecretManager(address, token, path string) *VaultSecretManager {
	return &VaultSecretManager{
		address: address,
		token:   token,
		path:    path,
	}
}

func (v *VaultSecretManager) GetEncryptionKey() (string, error) {
	zapLog := logger.GetLogger()

	// Buat Vault client
	config := &api.Config{
		Address: v.address,
	}

	client, err := api.NewClient(config)
	if err != nil {
		zapLog.Error("Failed to create Vault client", zap.Error(err))
		return "", fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set token
	client.SetToken(v.token)

	// Baca secret dari Vault
	// Support KV v2 format: secret/data/path -> data di dalam "data" field
	secret, err := client.Logical().Read(v.path)
	if err != nil {
		zapLog.Error("Failed to read secret from Vault", 
			zap.String("path", v.path),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to read secret from Vault: %w", err)
	}

	if secret == nil {
		zapLog.Error("Secret not found in Vault", zap.String("path", v.path))
		return "", fmt.Errorf("secret not found at path: %s", v.path)
	}

	// Handle KV v2 format (secret/data/path)
	// KV v2 menyimpan data di dalam "data" field
	var data map[string]interface{}
	if secret.Data["data"] != nil {
		// KV v2 format
		data = secret.Data["data"].(map[string]interface{})
		zapLog.Debug("Using KV v2 format", zap.String("path", v.path))
	} else {
		// KV v1 format atau direct data
		data = secret.Data
		zapLog.Debug("Using KV v1 format or direct data", zap.String("path", v.path))
	}

	// Ambil encryption_key dari data
	encryptionKey, ok := data["encryption_key"]
	if !ok {
		zapLog.Error("encryption_key not found in Vault secret", 
			zap.String("path", v.path),
			zap.Any("available_keys", getKeys(data)),
		)
		return "", fmt.Errorf("encryption_key not found in secret at path: %s", v.path)
	}

	// Convert ke string
	keyStr, ok := encryptionKey.(string)
	if !ok {
		zapLog.Error("encryption_key is not a string", 
			zap.String("path", v.path),
			zap.Any("type", fmt.Sprintf("%T", encryptionKey)),
		)
		return "", fmt.Errorf("encryption_key is not a string at path: %s", v.path)
	}

	if keyStr == "" {
		return "", fmt.Errorf("encryption_key is empty at path: %s", v.path)
	}

	zapLog.Info("Successfully retrieved encryption key from Vault", 
		zap.String("path", v.path),
		zap.Int("key_length", len(keyStr)),
	)

	return keyStr, nil
}

// getKeys helper untuk mendapatkan list keys dari map (untuk logging)
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetSecretManager mengembalikan SecretManager berdasarkan konfigurasi
// Priority:
// 1. Vault (jika VAULT_ADDR dan VAULT_TOKEN set)
// 2. Environment Variable (ENCRYPTION_KEY)
// 3. Default key untuk development
func GetSecretManager() SecretManager {
	zapLog := logger.GetLogger()

	// Cek apakah Vault dikonfigurasi
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultPath := os.Getenv("VAULT_SECRET_PATH")

	if vaultAddr != "" && vaultToken != "" {
		if vaultPath == "" {
			vaultPath = "secret/data/dms-app" // Default path
		}
		zapLog.Info("Using HashiCorp Vault for secret management",
			zap.String("address", vaultAddr),
			zap.String("path", vaultPath),
		)
		return NewVaultSecretManager(vaultAddr, vaultToken, vaultPath)
	}

	// Fallback ke environment variable
	zapLog.Info("Using environment variable for secret management")
	return &EnvSecretManager{}
}

// GetEncryptionKeyWithFallback mendapatkan encryption key dengan fallback strategy
// 1. Coba dari SecretManager (Vault atau Env)
// 2. Jika gagal, gunakan default key untuk development
func GetEncryptionKeyWithFallback() (string, error) {
	zapLog := logger.GetLogger()
	manager := GetSecretManager()

	key, err := manager.GetEncryptionKey()
	if err != nil {
		// Fallback ke default key untuk development
		// HARUS tepat 32 bytes (256 bits) untuk AES-256
		zapLog.Warn("Failed to get encryption key from secret manager, using default key (NOT SECURE FOR PRODUCTION!)",
			zap.Error(err),
		)
		return "default-encryption-key-32-chars!", nil // Default 32-byte key (tepat 32 karakter)
	}

	return key, nil
}

