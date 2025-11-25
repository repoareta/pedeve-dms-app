# HashiCorp Vault Setup untuk Development

Quick start guide untuk menggunakan HashiCorp Vault di development environment.

## 🚀 Quick Start (2 menit dengan script)

### Option A: Menggunakan Script (Paling Mudah)

```bash
# Run script setup Vault dev server
./scripts/vault-dev.sh

# Script akan:
# - Check Vault installation
# - Start Vault dev server
# - Enable KV secrets engine
# - Generate encryption key
# - Store key di Vault
# - Print environment variables untuk aplikasi

# Copy environment variables dari output, lalu:
make dev
```

### Option B: Manual Setup (5 menit)

### 1. Install Vault

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

### 2. Start Vault Dev Server

```bash
# Terminal 1: Start Vault
vault server -dev

# Output akan menampilkan:
# Unseal Key: <key>
# Root Token: <token>
# Vault Address: http://127.0.0.1:8200
```

**⚠️ Catatan**: Dev mode menyimpan data di memory, data akan hilang saat Vault di-stop.

**🌐 Web UI**: Vault memiliki web interface yang bisa diakses di **http://127.0.0.1:8200/ui**
- Login dengan Root Token dari output di atas
- Bisa manage secrets, policies, authentication, dll secara visual

### 3. Setup Secret

```bash
# Terminal 2: Setup environment
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="<root-token-dari-terminal-1>"

# Enable KV secrets engine
vault secrets enable -version=2 kv

# Generate encryption key (32 bytes)
ENCRYPTION_KEY=$(openssl rand -base64 32 | head -c 32)
echo "Generated key: $ENCRYPTION_KEY"

# Store di Vault
vault kv put secret/dms-app encryption_key="$ENCRYPTION_KEY"

# Verify
vault kv get secret/dms-app
```

### 4. Configure Application

```bash
# Set environment variables untuk aplikasi
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN="<root-token>"
export VAULT_SECRET_PATH="secret/data/dms-app"  # Optional

# Run aplikasi
make dev
```

Aplikasi akan otomatis menggunakan Vault untuk mendapatkan encryption key!

## 📋 Environment Variables

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `VAULT_ADDR` | Yes (untuk Vault) | Vault server address | - |
| `VAULT_TOKEN` | Yes (untuk Vault) | Vault authentication token | - |
| `VAULT_SECRET_PATH` | No | Path ke secret di Vault | `secret/data/dms-app` |
| `ENCRYPTION_KEY` | No (fallback) | Encryption key langsung | - |

## 🔄 Fallback Strategy

Aplikasi akan mencoba mendapatkan encryption key dengan urutan:

1. **HashiCorp Vault** (jika `VAULT_ADDR` dan `VAULT_TOKEN` set) ✅ **Fully Implemented**
2. **Environment Variable** (`ENCRYPTION_KEY`)
3. **Default Key** (development only, dengan warning)

**Status**: Vault integration sudah **fully implemented** dan siap digunakan!

## 🛠️ Vault Commands Cheat Sheet

```bash
# Check status
vault status

# List secrets
vault kv list secret/

# Get secret
vault kv get secret/dms-app

# Update secret
vault kv put secret/dms-app encryption_key="new-key-here"

# Delete secret
vault kv delete secret/dms-app
```

## 🚨 Troubleshooting

### Vault tidak bisa connect

```bash
# Cek apakah Vault running
vault status

# Cek VAULT_ADDR
echo $VAULT_ADDR

# Cek VAULT_TOKEN
echo $VAULT_TOKEN
```

### Secret tidak ditemukan

```bash
# Cek apakah secret path benar
vault kv get secret/dms-app

# Cek apakah KV engine enabled
vault secrets list
```

### Aplikasi masih pakai default key

- Pastikan `VAULT_ADDR` dan `VAULT_TOKEN` sudah di-set
- Cek log aplikasi untuk melihat secret manager yang digunakan
- Restart aplikasi setelah set environment variables

## ✅ Implementasi Lengkap

Vault integration sudah **fully implemented** dan siap digunakan!

- ✅ Vault client library terinstall
- ✅ VaultSecretManager fully implemented
- ✅ Support KV v1 dan KV v2 format
- ✅ Error handling dan logging lengkap
- ✅ Auto-detection Vault vs Env Var

## 📝 Next Steps

### Untuk Production

1. Setup Vault cluster (bukan dev mode)
2. Enable authentication (AppRole, Kubernetes, dll)
3. Store encryption key di Vault
4. Configure application dengan Vault credentials

## 🌐 Vault Web UI

Vault menyediakan web interface untuk manage secrets secara visual!

### Akses Web UI

1. **Start Vault dev server**:
   ```bash
   vault server -dev
   ```

2. **Buka browser**: http://127.0.0.1:8200/ui

3. **Login**: Gunakan Root Token dari output Vault server

### Fitur Web UI

- ✅ **Secrets Management**: View, create, edit, delete secrets
- ✅ **Policies**: Manage access policies
- ✅ **Authentication**: Configure auth methods
- ✅ **Audit Logs**: View audit logs
- ✅ **Health Status**: Monitor Vault health
- ✅ **Key-Value Secrets**: Manage KV secrets dengan UI

### Contoh: Manage Encryption Key via Web UI

1. Buka http://127.0.0.1:8200/ui
2. Login dengan Root Token
3. Navigate ke: **Secrets** → **kv** → **dms-app**
4. Klik **Create secret** atau **Edit**
5. Add key: `encryption_key` dengan value: `your-32-byte-key-here!!`
6. Klik **Save**

Aplikasi akan otomatis membaca dari Vault!

## ✅ Benefits

- ✅ Production-ready dari development
- ✅ Key rotation tanpa restart aplikasi
- ✅ Audit log untuk akses secret
- ✅ Centralized secret management
- ✅ Fallback ke env var jika Vault tidak tersedia
- ✅ **Web UI untuk visual management**

