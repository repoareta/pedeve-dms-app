# 📋 Planning CI/CD Deployment ke GCP

## 🎯 Tujuan
Setup CI/CD pipeline untuk deploy aplikasi DMS ke GCP dengan:
- Migrasi dari Vault ke GCP Secret Manager
- Database menggunakan Cloud SQL PostgreSQL
- Storage menggunakan GCP Cloud Storage
- Deployment otomatis ke VM (frontend-dev & backend-dev)

---

## 📊 Fase Implementasi

### **FASE 1: Migrasi Secret Management (Vault → GCP Secret Manager)**

#### 1.1 Backend Code Changes
- [ ] Buat `GCPSecretManager` struct di `backend/internal/infrastructure/secrets/`
- [ ] Implementasi interface `SecretManager` untuk GCP Secret Manager
- [ ] Update `GetSecretManager()` untuk support GCP Secret Manager
- [ ] Priority order: GCP Secret Manager → Environment Variable → Default
- [ ] Update semua referensi Vault ke GCP Secret Manager
- [ ] Update `GetSecretWithFallback()` untuk GCP Secret Manager

#### 1.2 Secret Mapping
Mapping secret dari Vault ke GCP Secret Manager:
- `encryption_key` → `encryption_key`
- `jwt_secret` → `jwt_secret`
- `superadmin_password` → `superadmin_password` (optional)
- `rate_limit_config` → `rate_limit_config` (optional)

#### 1.3 Dependencies
- [ ] Install GCP Secret Manager client library: `cloud.google.com/go/secretmanager`
- [ ] Update `go.mod` dan `go.sum`

---

### **FASE 2: Migrasi Storage (Local → GCP Cloud Storage)**

#### 2.1 Backend Code Changes
- [ ] Buat `GCPStorageManager` di `backend/internal/infrastructure/storage/`
- [ ] Implementasi interface untuk file upload/download
- [ ] Update `upload_handler.go` untuk menggunakan GCP Storage
- [ ] Migrasi file upload logic dari local filesystem ke GCP Storage
- [ ] Update URL generation untuk GCP Storage public URL

#### 2.2 GCP Storage Setup
- [ ] Buat GCP Storage Bucket (e.g., `pedeve-dms-uploads-dev`)
- [ ] Konfigurasi bucket permissions (IAM)
- [ ] Setup CORS untuk frontend access
- [ ] Konfigurasi lifecycle policy (optional)

#### 2.3 Dependencies
- [ ] Install GCP Storage client: `cloud.google.com/go/storage`

---

### **FASE 3: Database Connection (Cloud SQL via Auth Proxy)**

#### 3.1 Backend Code Changes
- [ ] Update `database.go` untuk menggunakan Cloud SQL connection string
- [ ] Konfigurasi connection via Cloud SQL Auth Proxy (localhost:5432)
- [ ] Update `DATABASE_URL` format untuk Cloud SQL
- [ ] Test connection ke Cloud SQL

#### 3.2 Environment Variables
- [ ] `DATABASE_URL` → dari GCP Secret Manager
- [ ] `DB_HOST` → `127.0.0.1` (via Auth Proxy)
- [ ] `DB_PORT` → `5432` (via Auth Proxy)
- [ ] `DB_USER` → dari GCP Secret Manager
- [ ] `DB_PASSWORD` → dari GCP Secret Manager
- [ ] `DB_NAME` → dari GCP Secret Manager

---

### **FASE 4: CI/CD Pipeline (GitHub Actions → GCP)**

#### 4.1 GitHub Actions Workflow
- [ ] Update `.github/workflows/ci-cd.yml`
- [ ] Tambah job untuk deploy ke GCP
- [ ] Setup GCP authentication di GitHub Actions
- [ ] Build & push Docker images ke GCP Artifact Registry (atau tetap GHCR)
- [ ] Deploy ke VM menggunakan SSH atau gcloud CLI

#### 4.2 Deployment Strategy
**Backend Deployment:**
- [ ] Authenticate via Workload Identity Federation
- [ ] Use `gcloud compute ssh` dengan OS Login untuk akses `backend-dev` VM (34.101.49.147)
- [ ] Pull Docker image terbaru dari GHCR
- [ ] Stop container lama
- [ ] Start container baru dengan environment variables dari Secret Manager
- [ ] Health check endpoint: `https://api-pedeve-dev.aretaamany.com/health`

**Frontend Deployment:**
- [ ] Authenticate via Workload Identity Federation
- [ ] Use `gcloud compute ssh` dengan OS Login untuk akses `frontend-dev` VM (34.128.123.1)
- [ ] Pull Docker image terbaru dari GHCR
- [ ] Update Nginx configuration (jika perlu)
- [ ] Reload Nginx
- [ ] Verify: `https://pedeve-dev.aretaamany.com`

#### 4.3 GitHub Secrets & Configuration Required
**Workload Identity Federation (WIF) Setup:**
- [x] Service Account: `github-actions-deployer` ✅
- [x] WIF Pool & Provider: Already configured ✅
- [x] IAM Roles assigned:
  - `roles/compute.instanceAdmin.v1` ✅
  - `roles/compute.osLogin` ✅
  - `roles/iam.serviceAccountUser` ✅
  - `roles/storage.objectAdmin` ✅
  - `roles/cloudsql.client` ✅

**GitHub Secrets Required:**
- [ ] `GCP_PROJECT_ID` - `pedeve-pertamina-dms`
- [ ] `GCP_WORKLOAD_IDENTITY_PROVIDER` - `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- [ ] `GCP_SERVICE_ACCOUNT` - `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`
- [ ] `GCP_BACKEND_VM_IP` - `34.101.49.147`
- [ ] `GCP_FRONTEND_VM_IP` - `34.128.123.1`
- [ ] `GCP_SSH_USER` - `info@aretaamany.com` (via OS Login)

---

### **FASE 5: VM Setup & Configuration**

#### 5.1 Backend VM (`backend-dev`)
- [ ] Install Docker & Docker Compose
- [ ] Install Cloud SQL Auth Proxy
- [ ] Setup systemd service untuk Cloud SQL Auth Proxy
- [ ] Setup systemd service untuk backend container
- [ ] Konfigurasi firewall rules
- [ ] Setup log rotation
- [ ] Setup monitoring/health checks

#### 5.2 Frontend VM (`frontend-dev`)
- [ ] Install Docker
- [ ] Install & konfigurasi Nginx
- [ ] Setup Nginx reverse proxy untuk backend API
- [ ] Setup SSL/TLS (optional, untuk HTTPS)
- [ ] Konfigurasi static file serving
- [ ] Setup log rotation

#### 5.3 Docker Compose untuk Production
- [ ] Buat `docker-compose.prod.yml` untuk production
- [ ] Konfigurasi environment variables dari Secret Manager
- [ ] Setup health checks
- [ ] Setup restart policies
- [ ] Konfigurasi logging

---

### **FASE 6: Environment Configuration**

#### 6.1 Backend Environment Variables
```bash
# Database (dari GCP Secret Manager)
DATABASE_URL=postgres://user:pass@127.0.0.1:5432/db_dev_pedeve?sslmode=require
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=pedeve_user_db
DB_PASSWORD=<from-secret-manager>
DB_NAME=db_dev_pedeve

# GCP Secret Manager
GCP_PROJECT_ID=pedeve-pertamina-dms
GCP_SECRET_MANAGER_ENABLED=true

# GCP Storage
GCP_STORAGE_BUCKET=pedeve-dev-bucket
GCP_STORAGE_ENABLED=true

# JWT & Encryption (dari GCP Secret Manager)
JWT_SECRET=<from-secret-manager>
ENCRYPTION_KEY=<from-secret-manager>

# Application
PORT=8080
ENV=production
CORS_ORIGIN=https://pedeve-dev.aretaamany.com
```

#### 6.2 Frontend Environment Variables
```bash
# API URL
VITE_API_URL=https://api-pedeve-dev.aretaamany.com/api/v1

# Environment
NODE_ENV=production
```

---

## 🔧 File & Script yang Akan Dibuat/Dimodifikasi

### **File Baru:**
1. `backend/internal/infrastructure/secrets/gcp_secret_manager.go`
2. `backend/internal/infrastructure/storage/gcp_storage.go`
3. `backend/internal/infrastructure/storage/storage_interface.go`
4. `.github/workflows/deploy-gcp.yml` (atau update `ci-cd.yml`)
5. `docker-compose.prod.yml`
6. `scripts/deploy-backend.sh`
7. `scripts/deploy-frontend.sh`
8. `scripts/setup-backend-vm.sh`
9. `scripts/setup-frontend-vm.sh`
10. `scripts/health-check.sh`

### **File yang Dimodifikasi:**
1. `backend/internal/infrastructure/secrets/secrets.go`
2. `backend/internal/infrastructure/database/database.go`
3. `backend/internal/delivery/http/upload_handler.go`
4. `backend/go.mod`
5. `.github/workflows/ci-cd.yml`
6. `backend/Dockerfile` (jika perlu install gcloud CLI)
7. `README.md` (update documentation)

---

## 📦 Dependencies yang Diperlukan

### **Backend Go Dependencies:**
```go
cloud.google.com/go/secretmanager v1.11.0
cloud.google.com/go/storage v1.30.0
google.golang.org/api v0.126.0
```

### **GCP Services:**
- ✅ Cloud SQL PostgreSQL (sudah setup)
- ✅ Secret Manager (sudah setup)
- ⚠️ Cloud Storage (perlu dibuat bucket)
- ⚠️ Artifact Registry (optional, untuk Docker images)

---

## 🔐 Security & Permissions

### **IAM Roles Required:**
- **Backend VM Service Account:**
  - `roles/secretmanager.secretAccessor`
  - `roles/storage.objectAdmin` (untuk upload files)
  - `roles/cloudsql.client` (sudah ada)

- **GitHub Actions Service Account:**
  - `roles/compute.instanceAdmin.v1` (untuk deploy)
  - `roles/iam.serviceAccountUser`
  - `roles/artifactregistry.writer` (jika pakai Artifact Registry)

### **Firewall Rules:**
- ✅ HTTP (80) - Frontend
- ✅ HTTPS (443) - Frontend (optional)
- ✅ Backend API port (8080) - hanya dari Frontend VM
- ✅ SSH (22) - hanya dari authorized IPs

---

## 🚀 Deployment Flow

```
1. Developer push ke branch (main/development)
   ↓
2. GitHub Actions triggered
   ↓
3. Build & Test
   - Lint frontend & backend
   - Run tests
   - Security scan (Trivy)
   ↓
4. Build Docker Images
   - Build frontend image
   - Build backend image
   ↓
5. Push to Registry
   - Push ke GHCR atau GCP Artifact Registry
   ↓
6. Deploy to GCP
   - SSH ke backend VM
   - Pull latest image
   - Restart backend container
   - SSH ke frontend VM
   - Pull latest image
   - Restart frontend container
   ↓
7. Health Check
   - Verify backend health
   - Verify frontend accessibility
   ↓
8. Notification (optional)
   - Slack/Discord notification
   - Email notification
```

---

## ✅ Checklist Pre-Implementation

### **Informasi yang Diperlukan dari User:**
- [x] GCP Project ID: `pedeve-pertamina-dms` ✅
- [x] Backend VM IP address: `34.101.49.147` ✅
- [x] Frontend VM IP address: `34.128.123.1` ✅
- [x] SSH username untuk VM: `info@aretaamany.com` ✅
- [x] SSH keys: Provided (Google-managed SSH keys) ✅
- [x] Domain name untuk frontend: `pedeve-dev.aretaamany.com` ✅
- [x] Domain name untuk backend API: `api-pedeve-dev.aretaamany.com` ✅
- [x] GCP Storage bucket name: `pedeve-dev-bucket` ✅
- [x] Service Account: `github-actions-deployer` (WIF enabled) ✅

### **GCP Resources yang Perlu Dibuat:**
- [x] GCP Storage Bucket: `pedeve-dev-bucket` (sudah dibuat) ✅
  - Location: `asia-southeast2` (Jakarta)
  - Storage Class: Standard
  - Public Access: Not public (Private)
- [x] Service Account: `github-actions-deployer` (sudah ada dengan WIF) ✅
- [x] IAM bindings untuk Service Accounts (sudah dikonfigurasi) ✅
- [x] Workload Identity Federation (WIF) Pool & Provider (sudah dikonfigurasi) ✅
- [ ] Artifact Registry (optional, bisa tetap pakai GHCR)

---

## 📝 Notes

1. **Cloud SQL Auth Proxy**: Sudah terinstall di backend VM, jadi backend hanya perlu connect ke `127.0.0.1:5432`
2. **Secret Manager**: Semua secret sudah ada di GCP Secret Manager, tinggal implementasi code untuk read
3. **Storage**: Perlu buat bucket baru untuk file uploads
4. **Deployment**: Bisa pakai SSH + Docker atau gcloud CLI
5. **Rollback Strategy**: Simpan previous Docker image tag untuk rollback jika perlu

---

## 🎯 Success Criteria

- [ ] Backend bisa read secrets dari GCP Secret Manager
- [ ] Backend bisa upload files ke GCP Storage
- [ ] Backend bisa connect ke Cloud SQL via Auth Proxy
- [ ] CI/CD pipeline berhasil deploy ke VM
- [ ] Frontend bisa akses backend API
- [ ] Health checks berjalan dengan baik
- [ ] Logs accessible untuk debugging

---

## ⚠️ Risks & Mitigation

1. **Risk**: Secret Manager access issues
   - **Mitigation**: Test connection sebelum deploy, setup proper IAM

2. **Risk**: Storage bucket permissions
   - **Mitigation**: Test upload/download, setup CORS dengan benar

3. **Risk**: Database connection timeout
   - **Mitigation**: Setup connection pooling, retry logic

4. **Risk**: Deployment failure
   - **Mitigation**: Implement rollback mechanism, health checks

---

## 📞 Next Steps

Setelah planning ini disetujui:
1. Implementasi Fase 1 (Secret Manager migration)
2. Implementasi Fase 2 (Storage migration)
3. Setup CI/CD pipeline
4. Test deployment
5. Production deployment

