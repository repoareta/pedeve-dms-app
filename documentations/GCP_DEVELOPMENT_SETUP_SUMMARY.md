# 📋 Catatan Setup Server Development GCP

Dokumentasi lengkap tentang semua konfigurasi, akses, dan resources yang sudah di-setup di Server Development Google Cloud Platform.

---

## 🌐 Informasi Project GCP

| Parameter | Value |
|-----------|-------|
| **Project ID** | `pedeve-pertamina-dms` |
| **Project Number** | `1076379007862` |
| **Region** | `asia-southeast2` (Jakarta) |
| **Zone** | `asia-southeast2-a` |

---

## 🖥️ Virtual Machines (Compute Engine)

### Backend VM
| Parameter | Value |
|-----------|-------|
| **Name** | `backend-dev` |
| **External IP** | `34.101.49.147` |
| **Region/Zone** | `asia-southeast2-a` |
| **Machine Type** | `e2-medium` (2 vCPU, 4GB RAM) |
| **SSH User** | `info@aretaamany.com` (via OS Login) |
| **Domain** | `api-pedeve-dev.aretaamany.com` |
| **Purpose** | Run Docker container backend + Cloud SQL Proxy |
| **Network Tags** | `backend-api-server`, `https-server` |

### Frontend VM
| Parameter | Value |
|-----------|-------|
| **Name** | `frontend-dev` |
| **External IP** | `34.128.123.1` |
| **Region/Zone** | `asia-southeast2-a` |
| **Machine Type** | `e2-micro` (1 vCPU, 2GB RAM) |
| **SSH User** | `info@aretaamany.com` (via OS Login) |
| **Domain** | `pedeve-dev.aretaamany.com` |
| **Purpose** | Serve static files via Nginx |
| **Network Tags** | `https-server` |

---

## 🗄️ Cloud SQL PostgreSQL

| Parameter | Value |
|-----------|-------|
| **Instance Name** | `postgres-dev` |
| **Engine** | PostgreSQL 16 |
| **Region** | `asia-southeast2` (Jakarta) |
| **Machine Type** | Custom (1 vCPU, 3.75GB RAM) |
| **Storage** | 10GB SSD |
| **Public IP** | Enabled |
| **Authorized Networks** | IP `backend-dev` VM only |
| **Database Name** | `db_dev_pedeve` |
| **Database User** | `pedeve_user_db` |
| **Connection Method** | Cloud SQL Auth Proxy (port 127.0.0.1:5432) |
| **Connection String** | `pedeve-pertamina-dms:asia-southeast2:postgres-dev=tcp:5432` |

---

## 📦 Cloud Storage Bucket

| Parameter | Value |
|-----------|-------|
| **Bucket Name** | `pedeve-dev-bucket` |
| **Location** | `asia-southeast2` (Jakarta) |
| **Storage Class** | Standard |
| **Public Access** | Not public (Private) |
| **Public Access Prevention** | Not enforced |
| **Protection** | None |
| **Access Control** | Uniform access control |
| **Base Paths** | `logos/`, `documents/` |
| **Public URL Format** | `https://storage.googleapis.com/pedeve-dev-bucket/{path}/{filename}` |

---

## 🔐 Secret Manager Secrets

Semua secrets disimpan di GCP Secret Manager dan diakses oleh Backend VM service account:

| Secret Name | Value | Purpose |
|-------------|-------|---------|
| `db_password` | `***` | Database password |
| `db_user` | `pedeve_user_db` | Database username |
| `db_name` | `db_dev_pedeve` | Database name |
| `db_host` | Cloud SQL Public IP | Database host |
| `db_port` | `5432` | Database port |
| `jwt_secret` | `***` | JWT token secret |
| `encryption_key` | `***` (optional) | Encryption key |

---

## 🔑 Service Accounts & IAM

### Service Account untuk GitHub Actions
| Parameter | Value |
|-----------|-------|
| **Service Account Email** | `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com` |
| **Purpose** | Deployment dari GitHub Actions via Workload Identity Federation |

**IAM Roles yang diberikan:**
- ✅ `roles/compute.instanceAdmin.v1` - Manage VMs
- ✅ `roles/compute.osLogin` - SSH access via OS Login
- ✅ `roles/iam.serviceAccountUser` - Impersonation
- ✅ `roles/storage.objectAdmin` - Upload files to bucket
- ✅ `roles/cloudsql.client` - Access Cloud SQL
- ✅ `roles/secretmanager.secretAccessor` - Access Secret Manager

### Backend VM Service Account
- **Default Compute Service Account** dengan akses ke:
  - Cloud SQL Client (untuk Cloud SQL Auth Proxy)
  - Secret Manager Secret Accessor (untuk membaca secrets)
  - Storage Object Admin (untuk upload file)

---

## 🔄 Workload Identity Federation (WIF)

| Parameter | Value |
|-----------|-------|
| **Project Number** | `1076379007862` |
| **Pool ID** | `github-actions-pool` |
| **Provider ID** | `github-actions-provider` |
| **WIF Provider Path** | `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| **Service Account** | `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com` |
| **Repository** | `repoareta/pedeve-dms-app` |
| **Audience** | `https://github.com/repoareta/pedeve-dms-app` |
| **Type** | OIDC |
| **State** | ACTIVE ✅ |

**Attribute Mapping:**
- `google.subject` = `assertion.sub`
- `attribute.repository` = `assertion.repository`
- `attribute.actor` = `assertion.actor`

---

## 🔥 Firewall Rules

| Rule Name | Type | Ports | Target Tags | Source | Purpose |
|-----------|------|-------|-------------|--------|---------|
| `allow-backend-api` | Ingress | `tcp:8080` | `backend-api-server` | `0.0.0.0/0` | Allow backend API access |
| `allow-https` | Ingress | `tcp:443` | `https-server` | `0.0.0.0/0` | Allow HTTPS traffic |
| Default HTTP | Ingress | `tcp:80` | `http-server` | `0.0.0.0/0` | Allow HTTP traffic |

---

## 🌍 DNS Configuration

| Domain | IP | Purpose | SSL |
|--------|-----|---------|-----|
| `pedeve-dev.aretaamany.com` | `34.128.123.1` | Frontend | ✅ Let's Encrypt |
| `api-pedeve-dev.aretaamany.com` | `34.101.49.147` | Backend API | ✅ Let's Encrypt |

---

## 🔒 SSL/TLS Configuration

| Domain | Certificate | Auto-renewal | Expires |
|--------|-------------|--------------|---------|
| `pedeve-dev.aretaamany.com` | Let's Encrypt | ✅ Enabled | 2026-02-25 |
| `api-pedeve-dev.aretaamany.com` | Let's Encrypt | ✅ Enabled | 2026-02-25 |

**Setup Method:** Certbot (Let's Encrypt)

---

## 🐳 VM Setup & Software

### Backend VM (`backend-dev`)
- ✅ Docker installed
- ✅ Cloud SQL Auth Proxy running (systemd service)
- ✅ Nginx installed & configured (reverse proxy untuk backend API)
- ✅ SSL certificate configured (Let's Encrypt)
- ✅ Service account dengan akses ke Secret Manager & Storage
- ✅ Port 8080 accessible (backend container)
- ✅ Port 80 & 443 accessible (Nginx)

### Frontend VM (`frontend-dev`)
- ✅ Docker installed (optional, untuk development)
- ✅ Nginx installed & configured (serve static files + SPA routing)
- ✅ SSL certificate configured (Let's Encrypt)
- ✅ Port 80 & 443 accessible (Nginx)

---

## 📝 GitHub Secrets Configuration

Secrets yang sudah dikonfigurasi di GitHub Repository (`Settings → Secrets and variables → Actions`):

| Secret Name | Value | Purpose |
|-------------|-------|---------|
| `GCP_PROJECT_ID` | `pedeve-pertamina-dms` | GCP Project ID |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` | WIF Provider path |
| `GCP_SERVICE_ACCOUNT` | `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com` | Service account email |
| `GCP_BACKEND_VM_IP` | `34.101.49.147` | Backend VM IP |
| `GCP_FRONTEND_VM_IP` | `34.128.123.1` | Frontend VM IP |
| `GCP_SSH_USER` | `info@aretaamany.com` | SSH user (via OS Login) |
| `GHCR_TOKEN` | Personal Access Token dengan `write:packages` permission | GitHub Container Registry token |
| `GCP_PROJECT_NUMBER` | `1076379007862` | GCP Project Number |

---

## 🔄 CI/CD Configuration

### GitHub Actions Workflow
- **Workflow File:** `.github/workflows/ci-cd.yml`
- **Trigger Branch:** `development` (auto-deploy)
- **Build Job:** `build-and-push`
- **Deploy Job:** `deploy-gcp`
- **Authentication:** Workload Identity Federation (WIF)

### Deployment Flow
1. Push ke branch `development`
2. GitHub Actions build backend Docker image
3. Build frontend static files
4. Push images ke GHCR
5. Authenticate via WIF
6. Deploy backend ke `backend-dev` VM
7. Deploy frontend ke `frontend-dev` VM
8. Health check verification

---

## 🌐 Environment Variables

### Backend Container
```bash
# Database (dari GCP Secret Manager)
DATABASE_URL=postgres://pedeve_user_db:<secret>@127.0.0.1:5432/db_dev_pedeve?sslmode=require
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=pedeve_user_db
DB_PASSWORD=<from-secret-manager>
DB_NAME=db_dev_pedeve

# GCP
GCP_PROJECT_ID=pedeve-pertamina-dms
GCP_SECRET_MANAGER_ENABLED=true
GCP_STORAGE_BUCKET=pedeve-dev-bucket
GCP_STORAGE_ENABLED=true

# Secrets (from Secret Manager)
JWT_SECRET=<from-secret-manager>
ENCRYPTION_KEY=<from-secret-manager>

# Application
PORT=8080
ENV=production
CORS_ORIGIN=https://pedeve-dev.aretaamany.com
```

### Frontend Build
```bash
VITE_API_URL=https://api-pedeve-dev.aretaamany.com/api/v1
NODE_ENV=production
```

---

## 📁 File Storage Configuration

| Parameter | Value |
|-----------|-------|
| **Provider** | GCP Cloud Storage |
| **Bucket** | `pedeve-dev-bucket` |
| **Base Paths** | `logos/`, `documents/` |
| **Public URL Format** | `https://storage.googleapis.com/pedeve-dev-bucket/{path}/{filename}` |
| **CORS Configuration** | Configured untuk allow access dari frontend domain |

---

## 🔐 Authentication & Access

### SSH Access
- **Method:** OS Login (tidak perlu private key)
- **User:** `info@aretaamany.com`
- **Command:** `gcloud compute ssh info@aretaamany.com@backend-dev --zone=asia-southeast2-a`

### Service Account Access
- **No JSON Keys Required** ✅
- Semua authentication menggunakan Workload Identity Federation
- VM menggunakan default compute service account dengan IAM roles

---

## 📋 Checklist Setup yang Sudah Selesai

### ✅ GCP Resources
- [x] Cloud SQL instance: `postgres-dev`
- [x] Storage bucket: `pedeve-dev-bucket`
- [x] Backend VM: `backend-dev` (34.101.49.147)
- [x] Frontend VM: `frontend-dev` (34.128.123.1)
- [x] Workload Identity Federation configured
- [x] Service Account dengan semua permissions yang diperlukan

### ✅ GCP Secret Manager Secrets
- [x] `db_password` - Database password
- [x] `db_user` - `pedeve_user_db`
- [x] `db_name` - `db_dev_pedeve`
- [x] `db_host` - Cloud SQL Public IP
- [x] `db_port` - `5432`
- [x] `jwt_secret` - JWT secret key
- [x] `encryption_key` - Encryption key (optional)

### ✅ VM Setup
- [x] Docker installed di kedua VM
- [x] Cloud SQL Auth Proxy running di backend VM
- [x] Nginx configured di kedua VM
- [x] SSL certificates configured (Let's Encrypt)
- [x] Service account dengan akses ke Secret Manager & Storage
- [x] Firewall rules configured
- [x] Network tags applied

### ✅ DNS Configuration
- [x] `pedeve-dev.aretaamany.com` → `34.128.123.1` (Frontend)
- [x] `api-pedeve-dev.aretaamany.com` → `34.101.49.147` (Backend)

### ✅ GitHub Secrets
- [x] Semua 7 secrets sudah dikonfigurasi

### ✅ CI/CD
- [x] GitHub Actions workflow configured
- [x] WIF authentication working
- [x] Auto-deploy ke branch `development`

---

## 🔗 Quick Reference URLs

| Service | URL |
|---------|-----|
| Frontend | `https://pedeve-dev.aretaamany.com` |
| Backend API | `https://api-pedeve-dev.aretaamany.com/api/v1` |
| Health Check | `https://api-pedeve-dev.aretaamany.com/health` |
| Swagger Docs | `https://api-pedeve-dev.aretaamany.com/swagger/index.html` |

---

## 📝 Important Notes

1. **Authentication:** Semua menggunakan Workload Identity Federation, tidak perlu JSON keys
2. **SSH Access:** Menggunakan OS Login, tidak perlu private key
3. **Secrets:** Semua application secrets disimpan di GCP Secret Manager
4. **Storage:** File uploads disimpan di GCP Cloud Storage bucket
5. **Database:** Connection via Cloud SQL Auth Proxy di `127.0.0.1:5432`
6. **Deployment:** Otomatis via GitHub Actions ketika push ke branch `development`
7. **SSL:** Let's Encrypt dengan auto-renewal

---

**Last Updated:** 2025-01-27  
**Environment:** Development  
**Status:** ✅ All configurations tested and working
