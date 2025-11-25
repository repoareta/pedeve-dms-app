# Secret Management

Secret management dengan support HashiCorp Vault dan Environment Variable fallback.

## Setup Options

### Option 1: HashiCorp Vault (Recommended untuk Production-ready Development)

#### Install Vault

```bash
# macOS
brew install vault

# Linux
wget https://releases.hashicorp.com/vault/1.15.0/vault_1.15.0_linux_amd64.zip
unzip vault_1.15.0_linux_amd64.zip
sudo mv vault /usr/local/bin/

# Verify
vault version
```

#### Start Vault Dev Server

```bash
# Start Vault dalam dev mode (untuk development/testing)
vault server -dev

# Output akan menampilkan:
# - Root Token: <token>
# - Unseal Key: <key>
# - Vault Address: http://127.0.0.1:8200
```

#### Setup Secret di Vault

```bash
# Set environment variables
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="<root-token-dari-output-vault-server>"

# Enable KV secrets engine (jika belum)
vault secrets enable -version=2 kv

# Store encryption key
vault kv put secret/dms-app encryption_key="your-32-byte-encryption-key-here!!"

# Verify
vault kv get secret/dms-app
```

#### Configure Application

```bash
# Set environment variables untuk aplikasi
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="<root-token>"
export VAULT_SECRET_PATH="secret/data/dms-app"  # Optional, default: secret/data/dms-app

# Run aplikasi
make dev
```

**Note**: Vault integration akan otomatis digunakan jika `VAULT_ADDR` dan `VAULT_TOKEN` di-set.

### Option 2: Environment Variable (Simple, untuk Development)

```bash
# Set encryption key langsung
export ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"

# Run aplikasi
make dev
```

### Option 3: Default Key (Development Only)

Jika tidak ada `VAULT_ADDR` dan `ENCRYPTION_KEY`, aplikasi akan menggunakan default key (tidak aman untuk production!).

## Priority Order

1. **HashiCorp Vault** (jika `VAULT_ADDR` dan `VAULT_TOKEN` set)
2. **Environment Variable** (`ENCRYPTION_KEY`)
3. **Default Key** (development only, dengan warning)

## ✅ Vault Integration (Fully Implemented)

Vault integration sudah **fully implemented** dan siap digunakan!

- ✅ Vault client library terinstall (`github.com/hashicorp/vault/api`)
- ✅ `VaultSecretManager.GetEncryptionKey()` fully implemented
- ✅ Support KV v1 dan KV v2 format
- ✅ Error handling dan logging lengkap
- ✅ Auto-detection Vault vs Env Var

Tidak perlu setup tambahan, langsung bisa digunakan!

## Production Setup

### Vault Production

1. Setup Vault cluster (bukan dev mode)
2. Enable authentication (AppRole, Kubernetes, dll)
3. Store encryption key di Vault
4. Configure application dengan Vault credentials

### Environment Variable Production

1. Generate strong 32-byte key
2. Store di secure key management system
3. Set sebagai environment variable di production server
4. Jangan commit ke git!

## Security Notes

- **Development**: Vault dev mode OK untuk testing
- **Production**: Wajib Vault cluster atau secure key management
- **Key Rotation**: Vault memudahkan key rotation tanpa restart aplikasi
- **Audit**: Vault menyediakan audit log untuk akses secret

