# 🔐 Deployment Configuration - GCP

## 📋 Informasi Deployment

### **VM Information**
- **Backend VM:**
  - Name: `backend-dev`
  - External IP: `34.101.49.147`
  - Region: `asia-southeast2` (Jakarta)
  - Type: `e2-medium` (2 vCPU, 4GB RAM)
  - SSH User: `info@aretaamany.com` (via OS Login)
  - Domain: `api-pedeve-dev.aretaamany.com`

- **Frontend VM:**
  - Name: `frontend-dev`
  - External IP: `34.128.123.1`
  - Region: `asia-southeast2` (Jakarta)
  - Type: `e2-micro` (1 vCPU, 2GB RAM)
  - SSH User: `info@aretaamany.com` (via OS Login)
  - Domain: `pedeve-dev.aretaamany.com`

### **GCP Resources**
- **Project ID:** `pedeve-pertamina-dms`
- **Project Number:** `1076379007862`
- **Cloud SQL Instance:** `postgres-dev`
- **Storage Bucket:** `pedeve-dev-bucket` ✅
  - Location: `asia-southeast2` (Jakarta)
  - Storage Class: Standard
  - Public Access: Not public (Private)
  - Protection: None
- **Service Account:** `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`

### **Authentication**
- **Method:** Workload Identity Federation (WIF)
- **No JSON Keys Required** ✅
- **OS Login Enabled** ✅

### **Secrets di GCP Secret Manager**
- `db_password` - Database password
- `db_user` - `pedeve_user_db`
- `db_name` - `db_dev_pedeve`
- `db_host` - Cloud SQL Public IP
- `db_port` - `5432`
- `jwt_secret` - JWT secret key
- `encryption_key` - Encryption key (jika ada)

### **Environment Variables**

#### Backend
```bash
# Database
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

#### Frontend
```bash
VITE_API_URL=https://api-pedeve-dev.aretaamany.com/api/v1
NODE_ENV=production
```

---

## 🔑 GitHub Secrets Configuration

Setelah WIF setup, tambahkan secrets berikut di GitHub Repository:

1. **GCP_PROJECT_ID**
   - Value: `pedeve-pertamina-dms`

2. **GCP_WORKLOAD_IDENTITY_PROVIDER**
   - Value: `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
   - Pool ID: `github-actions-pool`
   - Provider ID: `github-actions-provider`
   - Type: OIDC
   - State: ACTIVE ✅
   - Description: OIDC pool for GitHub Actions deployment

3. **GCP_SERVICE_ACCOUNT**
   - Value: `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`
   - IAM Roles:
     - `roles/compute.instanceAdmin.v1` ✅
     - `roles/compute.osLogin` ✅
     - `roles/iam.serviceAccountUser` ✅
     - `roles/storage.objectAdmin` ✅
     - `roles/cloudsql.client` ✅

4. **GCP_BACKEND_VM_IP**
   - Value: `34.101.49.147`

5. **GCP_FRONTEND_VM_IP**
   - Value: `34.128.123.1`

---

## 📝 Notes

- SSH access menggunakan OS Login, tidak perlu private key
- Cloud SQL connection via Auth Proxy di `127.0.0.1:5432`
- File uploads akan disimpan di GCP Storage bucket `pedeve-dev-bucket`
- Semua secrets diambil dari GCP Secret Manager saat runtime

