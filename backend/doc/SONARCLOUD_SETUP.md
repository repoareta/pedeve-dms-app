# SonarCloud Integration Setup

Dokumentasi untuk setup integrasi SonarCloud dengan aplikasi.

## Konfigurasi Secrets

Aplikasi ini menggunakan **Secret Manager** (Vault untuk development, GCP Secret Manager untuk production) untuk menyimpan secrets. SonarCloud configuration akan otomatis diambil dari Secret Manager.

### Priority Order (dari yang paling tinggi):
1. **Secret Manager** (Vault atau GCP Secret Manager)
2. **Environment Variables** (fallback)
3. **Default values** (untuk URL saja)

## Setup untuk Development (Vault)

### 1. Store Secrets ke Vault

Gunakan script yang sudah ada atau manual:

#### Option A: Menggunakan Script (Recommended)
```bash
# Set environment variables terlebih dahulu
export SONARCLOUD_URL=https://sonarcloud.io
export SONARCLOUD_TOKEN=your-sonarcloud-token-here
export SONARCLOUD_PROJECT_KEY=repoareta_pedeve-dms-app

# Store ke Vault
cd backend
./scripts/store-all-secrets.sh
```

Script akan otomatis menambahkan SonarCloud secrets ke Vault.

#### Option B: Manual via Vault CLI
```bash
# Set Vault environment
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=dev-root-token-12345

# Store SonarCloud secrets
vault kv put secret/dms-app \
    SONARCLOUD_URL=https://sonarcloud.io \
    SONARCLOUD_TOKEN=your-sonarcloud-token-here \
    SONARCLOUD_PROJECT_KEY=repoareta_pedeve-dms-app
```

### 2. Set Vault Environment Variables untuk Backend

Backend perlu tahu cara mengakses Vault:

```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=dev-root-token-12345
export VAULT_SECRET_PATH=secret/dms-app
```

### 3. Verifikasi Secrets di Vault

```bash
vault kv get secret/dms-app | grep SONARCLOUD
```

## Setup untuk Production (GCP Secret Manager)

### 1. Set GCP Environment Variables

```bash
export GCP_SECRET_MANAGER_ENABLED=true
export GCP_PROJECT_ID=your-gcp-project-id
```

### 2. Store Secrets ke GCP Secret Manager

```bash
# Install gcloud CLI jika belum
gcloud secrets create SONARCLOUD_URL --data-file=- <<< "https://sonarcloud.io"
gcloud secrets create SONARCLOUD_TOKEN --data-file=- <<< "your-sonarcloud-token-here"
gcloud secrets create SONARCLOUD_PROJECT_KEY --data-file=- <<< "repoareta_pedeve-dms-app"
```

## Fallback: Environment Variables (untuk testing)

Jika Secret Manager tidak tersedia, backend akan fallback ke environment variables:

```bash
export SONARCLOUD_URL=https://sonarcloud.io
export SONARCLOUD_TOKEN=your-sonarcloud-token-here
export SONARCLOUD_PROJECT_KEY=repoareta_pedeve-dms-app
```

**Catatan**: Fallback ini hanya untuk development/testing. Untuk production, gunakan Secret Manager.

**⚠️ SECURITY WARNING**: Jangan pernah commit token real ke version control. Gunakan placeholder seperti `your-sonarcloud-token-here` di dokumentasi.

### Cara Mendapatkan Token

1. Login ke [SonarCloud](https://sonarcloud.io)
2. Klik pada profil user (pojok kanan atas)
3. Pilih **My Account** → **Security**
4. Klik **Generate Token**
5. Beri nama token (contoh: `pedeve-dms-app`)
6. Copy token yang dihasilkan (hanya muncul sekali!)

### Project Key

Project key biasanya ada di:
- File `sonar-project.properties` (jika ada)
- Konfigurasi CI/CD (GitHub Actions, dll)
- Dashboard SonarCloud → Project Settings

Untuk project ini: `repoareta_pedeve-dms-app`

## API Endpoints

### Get Issues
```
GET /api/v1/sonarqube/issues
```

Query Parameters (optional):
- `severities`: Comma-separated (BLOCKER, CRITICAL, MAJOR, MINOR, INFO)
- `types`: Comma-separated (BUG, VULNERABILITY, CODE_SMELL)
- `statuses`: Comma-separated (OPEN, CONFIRMED, REOPENED, RESOLVED)

### Export Issues
```
GET /api/v1/sonarqube/issues/export
```

Mengembalikan file JSON untuk download.

## Access Control

Hanya **superadmin** dan **admin** yang dapat mengakses endpoint SonarCloud.

## Default Filters (untuk VAPT)

Default filter yang digunakan:
- **Severity**: BLOCKER, CRITICAL, MAJOR
- **Type**: BUG, VULNERABILITY
- **Status**: OPEN, CONFIRMED, REOPENED

Filter ini dapat diubah melalui UI di Settings → SonarQube Monitor.

## Frontend Usage

1. Buka Settings page
2. Pilih menu **SonarQube Monitor** (hanya visible untuk superadmin/admin)
3. Klik **Refresh** untuk fetch issues dari SonarCloud
4. Gunakan filter untuk memfilter issues
5. Klik **Export JSON** untuk download issues dalam format JSON

## Troubleshooting

### Error 500: "Failed to initialize SonarCloud client"

**Kemungkinan penyebab:**
1. Secrets tidak ditemukan di Secret Manager (Vault/GCP)
2. Environment variables tidak di-set (jika tidak pakai Secret Manager)
3. Token tidak valid
4. Project key salah
5. Vault tidak terhubung atau token tidak valid

**Solusi:**

#### 1. Cek apakah secrets ada di Vault
```bash
# Set Vault environment
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=dev-root-token-12345

# Cek secrets
vault kv get secret/dms-app | grep SONARCLOUD
```

Jika tidak ada, simpan secrets:
```bash
export SONARCLOUD_URL=https://sonarcloud.io
export SONARCLOUD_TOKEN=your-sonarcloud-token-here
export SONARCLOUD_PROJECT_KEY=repoareta_pedeve-dms-app
cd backend
./scripts/store-all-secrets.sh
```

#### 2. Cek Vault connection di backend
Pastikan backend memiliki environment variables untuk Vault:
```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=dev-root-token-12345
export VAULT_SECRET_PATH=secret/dms-app
```

**Restart backend server** setelah set environment variables.

#### 3. Fallback ke Environment Variables (untuk testing)
Jika Vault tidak tersedia, backend akan fallback ke environment variables:
```bash
export SONARCLOUD_URL=https://sonarcloud.io
export SONARCLOUD_TOKEN=your-sonarcloud-token-here
export SONARCLOUD_PROJECT_KEY=repoareta_pedeve-dms-app
```

#### 4. Cek log backend untuk detail error
Backend akan log error detail ke console/log file. Cari pesan:
- `SONARCLOUD_TOKEN is required but not found`
- `Failed to get SONARCLOUD_TOKEN from secret manager`
- `Failed to create SonarCloud client`

### Error 500: "SonarCloud API returned status XXX"

**Penyebab**: 
- Token tidak valid atau expired
- Project key salah
- Network issue

**Solusi**:
1. Verifikasi token masih valid di SonarCloud dashboard
2. Cek project key di SonarCloud → Project Settings
3. Test koneksi dengan script:
   ```bash
   ./scripts/test-sonarqube.sh
   ```

## Notes

- Data tidak disimpan di database, hanya fetch on-demand
- Refresh dilakukan secara manual (button trigger)
- Export menghasilkan file JSON dengan timestamp
- **PENTING**: 
  - Untuk development: Simpan secrets di Vault dan set `VAULT_ADDR`, `VAULT_TOKEN`, `VAULT_SECRET_PATH`
  - Untuk production: Simpan secrets di GCP Secret Manager dan set `GCP_SECRET_MANAGER_ENABLED=true`, `GCP_PROJECT_ID`
  - Fallback: Jika Secret Manager tidak tersedia, backend akan menggunakan environment variables

