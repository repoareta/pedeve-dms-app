# ✅ Deployment Checklist - GCP CI/CD

## 📋 Pre-Deployment Checklist

### ✅ Code Implementation
- [x] GCP Secret Manager integration
- [x] GCP Cloud Storage integration
- [x] CI/CD workflow dengan WIF authentication
- [x] Deployment scripts
- [x] Build errors fixed
- [x] All dependencies installed

### ✅ GitHub Secrets (6 secrets)
- [x] `GCP_PROJECT_ID` = `pedeve-pertamina-dms`
- [x] `GCP_WORKLOAD_IDENTITY_PROVIDER` = `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- [x] `GCP_SERVICE_ACCOUNT` = `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`
- [x] `GCP_SSH_USER` = `info@aretaamany.com`
- [x] `GHCR_TOKEN` = Personal Access Token dengan `write:packages` permission
- [x] `GCP_PROJECT_NUMBER` = `1076379007862`

### ✅ GCP Resources
- [x] Cloud SQL instance: `postgres-dev`
- [x] Storage bucket: `pedeve-dev-bucket`
- [x] Backend VM: `backend-dev` (34.101.49.147)
- [x] Frontend VM: `frontend-dev` (34.128.123.1)
- [x] Workload Identity Federation configured
- [x] Service Account dengan permissions:
  - [x] Compute Instance Admin (v1)
  - [x] Compute OS Login
  - [x] Service Account User
  - [x] Storage Object Admin
  - [x] Cloud SQL Client
  - [x] Secret Manager Secret Accessor

### ✅ GCP Secret Manager Secrets
- [x] `db_password` - Database password
- [x] `db_user` - `pedeve_user_db`
- [x] `db_name` - `db_dev_pedeve`
- [x] `db_host` - Cloud SQL Public IP
- [x] `db_port` - `5432`
- [x] `jwt_secret` - JWT secret key
- [x] `encryption_key` - Encryption key (optional)

### ✅ VM Setup
#### Backend VM (`backend-dev`)
- [x] Docker installed
- [x] Cloud SQL Auth Proxy running
- [x] Service account dengan akses ke Secret Manager & Storage
- [x] Port 8080 accessible

#### Frontend VM (`frontend-dev`)
- [x] Docker installed
- [x] Nginx configured (jika perlu)
- [x] Port 80 accessible

### ✅ DNS Configuration
- [x] `pedeve-dev.aretaamany.com` → `34.128.123.1` (Frontend)
- [x] `api-pedeve-dev.aretaamany.com` → `34.101.49.147` (Backend)

### ✅ Storage Bucket Configuration
- [x] Bucket `pedeve-dev-bucket` created
- [x] IAM permissions untuk service account
- [ ] Public read access configured (jika perlu untuk file uploads)

## 🚀 Deployment Steps

1. **Push ke branch `development`**
   ```bash
   git checkout development
   git add .
   git commit -m "feat: implement GCP CI/CD deployment"
   git push origin development
   ```

2. **Monitor GitHub Actions**
   - Go to: https://github.com/[owner]/[repo]/actions
   - Watch for `build-and-push` job
   - Watch for `deploy-gcp` job

3. **Verify Deployment**
   - Backend health: `https://api-pedeve-dev.aretaamany.com/health`
   - Frontend: `https://pedeve-dev.aretaamany.com`

## 🔍 Troubleshooting

### Build Fails
- Check GitHub Actions logs
- Verify all dependencies in `go.mod` and `package.json`
- Check Docker build logs

### Deployment Fails
- Check WIF authentication
- Verify service account permissions
- Check VM SSH access
- Verify Docker is running on VMs

### Health Check Fails
- Check backend logs: `docker logs dms-backend-prod`
- Check frontend logs: `docker logs dms-frontend-prod`
- Verify Cloud SQL Auth Proxy is running
- Check environment variables in container

## 📝 Notes

- Database connection akan otomatis diambil dari GCP Secret Manager
- File uploads akan disimpan ke GCP Cloud Storage
- Semua secrets diambil dari GCP Secret Manager, bukan environment variables
- Deployment hanya trigger pada branch `development`

