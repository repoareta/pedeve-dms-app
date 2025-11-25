# Encryption Utility

Utility untuk encrypt/decrypt data sensitif menggunakan AES-256-GCM.

## Setup

### Option 1: HashiCorp Vault (Recommended untuk Production-ready)

```bash
# Start Vault dev server
vault server -dev

# Set environment variables
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="<root-token>"
export VAULT_SECRET_PATH="secret/data/dms-app"  # Optional

# Store encryption key di Vault
vault kv put secret/dms-app encryption_key="your-32-byte-key-here!!"
```

Lihat `../secrets/README.md` untuk setup lengkap.

### Option 2: Environment Variable

```bash
# Production: Generate key 32 bytes (256 bits)
export ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"

# Development: Default key akan digunakan (tidak aman untuk production!)
```

**⚠️ PENTING**: Key harus tepat 32 bytes (256 bits) untuk AES-256.

### Priority Order

1. **HashiCorp Vault** (jika `VAULT_ADDR` dan `VAULT_TOKEN` set)
2. **Environment Variable** (`ENCRYPTION_KEY`)
3. **Default Key** (development only, dengan warning)

### 2. Initialize di main.go

Encryption sudah di-initialize otomatis di `cmd/api/main.go`.

## Usage

### Encrypt Data

```go
import "github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/encryption"

// Encrypt plaintext
encrypted, err := encryption.Encrypt("sensitive-data")
if err != nil {
    return err
}
// Simpan encrypted ke database
```

### Decrypt Data

```go
// Ambil encrypted dari database
encrypted := db.GetEncryptedData()

// Decrypt
plaintext, err := encryption.Decrypt(encrypted)
if err != nil {
    return err
}
// Gunakan plaintext
```

### Backward Compatibility

Utility ini otomatis handle backward compatibility untuk data existing yang belum di-encrypt:

```go
// Jika decrypt gagal (data belum di-encrypt), return as-is
decrypted, err := encryption.Decrypt(data)
// Jika err atau decrypted == data, berarti data belum di-encrypt
```

### Helper Functions

```go
// Cek apakah data sudah di-encrypt
if encryption.IsEncrypted(data) {
    // Data sudah encrypted
}

// Encrypt hanya jika belum di-encrypt (untuk migration)
encrypted, err := encryption.EncryptIfNotEncrypted(data)
```

## Implementation Details

- **Algorithm**: AES-256-GCM
- **Nonce**: Random 12 bytes per encryption
- **Encoding**: Base64 untuk storage
- **Key Management**: Environment variable `ENCRYPTION_KEY`

## Security Notes

1. **Production**: Wajib set `ENCRYPTION_KEY` environment variable dengan key yang kuat (32 bytes)
2. **Key Rotation**: Jika perlu rotate key, perlu migrate semua data existing
3. **Backup**: Simpan encryption key di secure key management system (bukan di code!)

## Current Usage

- ✅ 2FA Secret (`two_factor_auths.secret`)
- ✅ 2FA Backup Codes (`two_factor_auths.backup_codes`)

## Future Usage

Untuk modul utama yang memerlukan encryption, cukup tambahkan:

```go
// Before save
encrypted, err := encryption.Encrypt(sensitiveData)
if err != nil {
    return err
}
model.SensitiveField = encrypted

// After read
decrypted, err := encryption.Decrypt(model.SensitiveField)
if err != nil {
    // Handle error atau backward compatibility
}
```

