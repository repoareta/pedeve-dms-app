package secrets

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// SecretManager interface untuk key management
// Support multiple backends: Vault, Environment Variable, dll
type SecretManager interface {
	GetEncryptionKey() (string, error)
	GetSecret(key string) (string, error) // Generic method untuk get secret by key
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

func (e *EnvSecretManager) GetSecret(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s not set", key)
	}
	return value, nil
}

// convertToKVv2Path mengkonversi path untuk KV v2 API
// "secret/dms-app" -> "secret/data/dms-app"
func convertToKVv2Path(path string) string {
	if !strings.Contains(path, "/data/") && strings.HasPrefix(path, "secret/") {
		return strings.Replace(path, "secret/", "secret/data/", 1)
	}
	return path
}

// VaultSecretManager menggunakan HashiCorp Vault
type VaultSecretManager struct {
	address string
	token   string
	path    string // Path ke secret di Vault (e.g., "secret/dms-app" atau "secret/data/dms-app")
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
	// Support KV v2 format: untuk path "secret/dms-app", Vault API menggunakan "secret/data/dms-app"
	vaultPath := convertToKVv2Path(v.path)
	secret, err := client.Logical().Read(vaultPath)
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

// GetAllSecrets reads all secrets from Vault and returns all key-value pairs
func (v *VaultSecretManager) GetAllSecrets(path string) (map[string]interface{}, error) {
	zapLog := logger.GetLogger()

	// Buat Vault client
	config := &api.Config{
		Address: v.address,
	}

	client, err := api.NewClient(config)
	if err != nil {
		zapLog.Error("Failed to create Vault client", zap.Error(err))
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set token
	client.SetToken(v.token)

	// Baca secret dari Vault
	vaultPath := convertToKVv2Path(path)
	secret, err := client.Logical().Read(vaultPath)
	if err != nil {
		zapLog.Error("Failed to read secret from Vault",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read secret from Vault: %w", err)
	}

	if secret == nil {
		zapLog.Error("Secret not found in Vault", zap.String("path", path))
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// Handle KV v2 format (secret/data/path)
	var data map[string]interface{}
	if secret.Data["data"] != nil {
		// KV v2 format
		data = secret.Data["data"].(map[string]interface{})
		zapLog.Debug("Using KV v2 format", zap.String("path", path))
	} else {
		// KV v1 format atau direct data
		data = secret.Data
		zapLog.Debug("Using KV v1 format or direct data", zap.String("path", path))
	}

	return data, nil
}

// GetSecret reads a specific secret key from Vault
func (v *VaultSecretManager) GetSecret(key string) (string, error) {
	zapLog := logger.GetLogger()

	// Get all secrets from Vault
	data, err := v.GetAllSecrets(v.path)
	if err != nil {
		return "", err
	}

	// Extract specific key
	value, ok := data[key]
	if !ok {
		zapLog.Error("Secret key not found in Vault",
			zap.String("key", key),
			zap.String("path", v.path),
			zap.Any("available_keys", getKeys(data)),
		)
		return "", fmt.Errorf("secret key '%s' not found at path: %s", key, v.path)
	}

	// Convert to string
	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("secret key '%s' is not a string", key)
	}

	if valueStr == "" {
		return "", fmt.Errorf("secret key '%s' is empty", key)
	}

	zapLog.Debug("Successfully retrieved secret from Vault",
		zap.String("key", key),
		zap.String("path", v.path),
	)

	return valueStr, nil
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
			vaultPath = "secret/dms-app" // Default path for KV v2
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

// GetSecretWithFallback mendapatkan secret dengan fallback strategy
// 1. Coba dari SecretManager (Vault atau Env)
// 2. Jika gagal, gunakan environment variable
// 3. Jika masih gagal, return error atau default value
func GetSecretWithFallback(key string, envKey string, defaultValue string) (string, error) {
	zapLog := logger.GetLogger()
	manager := GetSecretManager()

	// Try to get from secret manager
	value, err := manager.GetSecret(key)
	if err == nil && value != "" {
		return value, nil
	}

	// Fallback to environment variable
	if envKey != "" {
		value = os.Getenv(envKey)
		if value != "" {
			zapLog.Info("Using environment variable for secret",
				zap.String("key", key),
				zap.String("env_key", envKey),
			)
			return value, nil
		}
	}

	// Fallback to default value if provided
	if defaultValue != "" {
		zapLog.Warn("Using default value for secret (NOT SECURE FOR PRODUCTION!)",
			zap.String("key", key),
		)
		return defaultValue, nil
	}

	return "", fmt.Errorf("secret '%s' not found and no fallback available", key)
}

