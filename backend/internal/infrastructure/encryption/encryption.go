package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/logger"
	"github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/secrets"
)

var (
	encryptionKey []byte
	initialized   bool
)

// InitEncryption menginisialisasi encryption dengan key dari secret manager
// Support: HashiCorp Vault (jika dikonfigurasi) atau Environment Variable
// Key harus 32 bytes (256 bits) untuk AES-256
func InitEncryption() error {
	zapLog := logger.GetLogger()
	
	// Get encryption key dari secret manager (Vault atau Env)
	keyStr, err := secrets.GetEncryptionKeyWithFallback()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}

	key := []byte(keyStr)
	
	// Validasi panjang key (harus 32 bytes untuk AES-256)
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be exactly 32 bytes (256 bits), got %d bytes", len(key))
	}

	encryptionKey = key
	initialized = true
	
	zapLog.Info("Encryption initialized successfully")
	return nil
}

// Encrypt mengenkripsi plaintext menggunakan AES-256-GCM
// Return: string ter-encrypt yang di-encode base64
func Encrypt(plaintext string) (string, error) {
	if !initialized {
		if err := InitEncryption(); err != nil {
			return "", fmt.Errorf("encryption not initialized: %w", err)
		}
	}

	if plaintext == "" {
		return "", nil
	}

	// Buat AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Buat GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce (12 bytes untuk GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode ke base64 untuk storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

// Decrypt mendekripsi ciphertext yang di-encrypt dengan Encrypt
// Input: string ter-encrypt yang di-encode base64
// Return: plaintext string
func Decrypt(ciphertext string) (string, error) {
	if !initialized {
		if err := InitEncryption(); err != nil {
			return "", fmt.Errorf("encryption not initialized: %w", err)
		}
	}

	if ciphertext == "" {
		return "", nil
	}

	// Decode dari base64
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// Jika bukan base64, mungkin data lama yang belum di-encrypt (backward compatibility)
		// Return as-is untuk backward compatibility
		return ciphertext, nil
	}

	// Buat AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Buat GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce (12 bytes pertama)
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		// Data terlalu pendek, mungkin data lama yang belum di-encrypt
		return ciphertext, nil
	}

	nonce, ciphertextBytes := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		// Jika decrypt gagal, mungkin data lama yang belum di-encrypt
		// Return as-is untuk backward compatibility
		return ciphertext, nil
	}

	return string(plaintext), nil
}

// IsEncrypted mengecek apakah string sudah di-encrypt atau belum
// Berguna untuk migration atau backward compatibility
func IsEncrypted(data string) bool {
	if data == "" {
		return false
	}

	// Coba decode base64
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}

	// Cek panjang minimum (nonce + minimal ciphertext)
	// AES-256-GCM nonce = 12 bytes, minimal ciphertext = 16 bytes (1 block)
	if len(decoded) < 28 {
		return false
	}

	// Coba decrypt untuk verifikasi
	// Jika berhasil decrypt dan hasilnya berbeda dari input, berarti encrypted
	decrypted, err := Decrypt(data)
	if err != nil {
		return false
	}

	// Jika decrypted != original, berarti encrypted
	// Tapi jika decrypt gagal dan return original, berarti tidak encrypted
	return decrypted != data || (decrypted == data && len(decoded) >= 28)
}

// EncryptIfNotEncrypted mengenkripsi data hanya jika belum di-encrypt
// Berguna untuk migration data existing
func EncryptIfNotEncrypted(data string) (string, error) {
	if IsEncrypted(data) {
		return data, nil
	}
	return Encrypt(data)
}

// GetEncryptionKeyLength mengembalikan panjang key yang diperlukan
func GetEncryptionKeyLength() int {
	return 32 // 32 bytes = 256 bits untuk AES-256
}

// ValidateEncryptionKey memvalidasi apakah key valid
func ValidateEncryptionKey(key string) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be exactly 32 bytes (256 bits), got %d bytes", len(key))
	}
	return nil
}

