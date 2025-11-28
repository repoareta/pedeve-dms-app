# 🚀 Deployment Guide - GCP CI/CD

## 📋 Overview

Sistem CI/CD sudah dikonfigurasi untuk deployment otomatis ke GCP ketika push ke branch `development`.

## ✅ Prerequisites

Semua prerequisites sudah disiapkan:
- ✅ GCP Project: `pedeve-pertamina-dms`
- ✅ Cloud SQL: `postgres-dev`
- ✅ Storage Bucket: `pedeve-dev-bucket`
- ✅ VMs: `backend-dev` & `frontend-dev`
- ✅ Workload Identity Federation configured
- ✅ GitHub Secrets configured (6 secrets)

## 🔄 Deployment Flow

1. **Push ke branch `development`**
   ```bash
   git checkout development
   git add .
   git commit -m "feat: your changes"
   git push origin development
   ```

2. **GitHub Actions akan otomatis:**
   - Build backend & frontend
   - Run tests & linting
   - Build Docker images
   - Push images ke GHCR
   - Deploy ke GCP VMs
   - Run health checks

## 🔧 Manual Deployment (Optional)

Jika perlu deploy manual, gunakan scripts:

```bash
# Deploy backend
./scripts/deploy-backend.sh [image-tag]

# Deploy frontend
./scripts/deploy-frontend.sh [image-tag]
```

## 📝 Environment Variables

### Backend Container
Backend akan otomatis mengambil secrets dari GCP Secret Manager:
- `GCP_PROJECT_ID` - Project ID GCP
- `GCP_SECRET_MANAGER_ENABLED=true` - Enable GCP Secret Manager
- `GCP_STORAGE_ENABLED=true` - Enable GCP Storage
- `GCP_STORAGE_BUCKET=pedeve-dev-bucket` - Storage bucket name
- `PORT=8080` - Backend port
- `ENV=production` - Environment
- `CORS_ORIGIN=https://pedeve-dev.aretaamany.com` - CORS origin

**Secrets yang diambil dari GCP Secret Manager:**
- `database_url` atau `DATABASE_URL` env var
- `jwt_secret`
- `encryption_key`

### Frontend Container
Frontend hanya perlu port mapping, tidak ada env vars.

## 🔍 Monitoring & Troubleshooting

### Check Deployment Status
```bash
# Backend health
curl https://api-pedeve-dev.aretaamany.com/health

# Frontend
curl https://pedeve-dev.aretaamany.com
```

### Check Container Logs
```bash
# SSH ke backend VM
gcloud compute ssh info@aretaamany.com@backend-dev --zone=asia-southeast2-a

# Check logs
docker logs dms-backend-prod
docker logs dms-frontend-prod

# Check running containers
docker ps
```

### Common Issues

1. **Build fails**
   - Check GitHub Actions logs
   - Verify dependencies

2. **Deployment fails**
   - Check WIF authentication
   - Verify service account permissions
   - Check VM SSH access

3. **Health check fails**
   - Check container logs
   - Verify Cloud SQL Auth Proxy is running
   - Check environment variables

## 📚 Related Documentation

- `DEPLOYMENT_CONFIG.md` - Detailed deployment configuration
- `GITHUB_SECRETS_SETUP.md` - GitHub Secrets setup guide
- `PLANNING_CICD_GCP.md` - Planning document

