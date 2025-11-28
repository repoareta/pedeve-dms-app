# 📋 Dokumentasi Deployment - Pedeve DMS App

Dokumentasi lengkap semua konfigurasi yang sudah berhasil untuk setup development dan production.

---

## 🌐 GCP Resources Configuration

### Virtual Machines (Compute Engine)

| Resource | Nama | Region | Type | IP External | Domain | Purpose |
|----------|------|--------|------|-------------|--------|---------|
| Frontend VM | `frontend-dev` | `asia-southeast2-a` | `e2-micro` (1 vCPU, 2GB RAM) | `34.128.123.1` | `pedeve-dev.aretaamany.com` | Serve static files via Nginx |
| Backend VM | `backend-dev` | `asia-southeast2-a` | `e2-medium` (2 vCPU, 4GB RAM) | `34.101.49.147` | `api-pedeve-dev.aretaamany.com` | Run Docker container backend + Cloud SQL Proxy |

### Cloud SQL PostgreSQL

| Parameter | Value |
|-----------|-------|
| Instance Name | `postgres-dev` |
| Engine | PostgreSQL 16 |
| Region | `asia-southeast2` (Jakarta) |
| Machine Type | Custom (1 vCPU, 3.75GB RAM) |
| Storage | 10GB SSD |
| Public IP | Enabled |
| Authorized Networks | IP `backend-dev` VM only |
| Database Name | `db_dev_pedeve` |
| Database User | `pedeve_user_db` |
| Connection Method | Cloud SQL Auth Proxy (port 127.0.0.1:5432) |

### Cloud Storage

| Parameter | Value |
|-----------|-------|
| Bucket Name | `pedeve-dev-bucket` |
| Location | `asia-southeast2` (Jakarta) |
| Storage Class | Standard |
| Public Access | Not public (Private) |
| Protection | None |

### Secret Manager

| Secret Name | Purpose | Access |
|-------------|---------|--------|
| `db_password` | Database password | Backend VM service account |
| `db_user` | Database username (`pedeve_user_db`) | Backend VM service account |
| `db_name` | Database name (`db_dev_pedeve`) | Backend VM service account |
| `db_host` | Cloud SQL Public IP | Backend VM service account |
| `db_port` | Database port (`5432`) | Backend VM service account |
| `jwt_secret` | JWT token secret | Backend VM service account |
| `encryption_key` | Encryption key (optional) | Backend VM service account |

### Workload Identity Federation (WIF)

| Parameter | Value |
|-----------|-------|
| Project Number | `1076379007862` |
| Pool ID | `github-actions-pool` |
| Provider ID | `github-actions-provider` |
| WIF Provider | `projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| Service Account | `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com` |
| Repository | `repoareta/pedeve-dms-app` |
| Audience | `https://github.com/repoareta/pedeve-dms-app` |

### IAM Roles & Permissions

| Service Account | Roles | Purpose |
|-----------------|-------|---------|
| Backend VM Service Account | `Cloud SQL Client` | Akses Cloud SQL |
| Backend VM Service Account | Full API Access | Akses GCP services |
| `github-actions-deployer@...` | `Compute Instance Admin (v1)` | Manage VMs |
| `github-actions-deployer@...` | `Compute OS Login` | SSH access via OS Login |
| `github-actions-deployer@...` | `Service Account User` | Impersonation |
| `github-actions-deployer@...` | `Storage Object Admin` | Upload files to bucket |
| `github-actions-deployer@...` | `Cloud SQL Client` | Access Cloud SQL |

### Firewall Rules

| Rule Name | Type | Ports | Target Tags | Source | Purpose |
|-----------|------|-------|-------------|--------|---------|
| `allow-backend-api` | Ingress | `tcp:8080` | `backend-api-server` | `0.0.0.0/0` | Allow backend API access |
| `allow-https` | Ingress | `tcp:443` | `https-server` | `0.0.0.0/0` | Allow HTTPS traffic |
| Default HTTP | Ingress | `tcp:80` | `http-server` | `0.0.0.0/0` | Allow HTTP traffic |

### Network Tags

| VM Name | Network Tags | Purpose |
|---------|--------------|---------|
| `backend-dev` | `backend-api-server`, `https-server` | Allow API and HTTPS access |
| `frontend-dev` | `https-server` | Allow HTTPS access |

---

## 🔧 Local Development Setup

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Docker | Latest | Container runtime |
| Docker Compose | Latest | Multi-container orchestration |
| Go | 1.25+ | Backend development |
| Node.js | 20+ | Frontend development |
| Make | Latest | Build automation |

### Development Services

| Service | Port | Container Name | Image | Purpose |
|---------|------|---------------|-------|---------|
| PostgreSQL | `5432` | `dms-postgres-dev` | `postgres:16-alpine` | Local database |
| Backend API | `8080` | `dms-backend-dev` | `golang:1.25-alpine` | Go API with hot reload (Air) |
| Frontend | `5173` | `dms-frontend-dev` | `node:20-alpine` | Vue dev server with HMR |
| Vault | `8200` | `dms-vault-dev` | `hashicorp/vault:latest` | Secret management (dev) |

### Development Commands

| Command | Purpose |
|---------|---------|
| `make dev` | Start all services (PostgreSQL, Backend, Frontend, Vault) |
| `make up` | Start services in background |
| `make down` | Stop all services |
| `make restart` | Restart all services |
| `make logs` | View all logs |
| `make seed-companies` | Seed sample companies and users |

### Local Database

| Parameter | Value |
|-----------|-------|
| Database Type | PostgreSQL (via Docker) |
| Database Name | `db_dms_pedeve` |
| User | `postgres` |
| Password | `dms_password` |
| Connection | `postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable` |

---

## 🚀 CI/CD Configuration

### GitHub Actions Workflow

| Component | Configuration |
|-----------|---------------|
| Workflow File | `.github/workflows/ci-cd.yml` |
| Trigger Branch | `development` (auto-deploy) |
| Build Job | `build-and-push` |
| Deploy Job | `deploy-gcp` |
| Authentication | Workload Identity Federation (WIF) |

### GitHub Secrets (Required)

| Secret Name | Purpose | Example |
|-------------|---------|---------|
| `GCP_PROJECT_ID` | GCP Project ID | `pedeve-pertamina-dms` |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | WIF Provider path | `projects/.../providers/github-actions-provider` |
| `GCP_SERVICE_ACCOUNT` | Service account email | `github-actions-deployer@...` |
| `GCP_BACKEND_VM_IP` | Backend VM IP | `34.101.49.147` |
| `GCP_FRONTEND_VM_IP` | Frontend VM IP | `34.128.123.1` |
| `GHCR_TOKEN` | GitHub Container Registry token | (auto-generated) |

### Docker Images

| Image | Registry | Purpose |
|-------|----------|---------|
| Backend | `ghcr.io/repoareta/dms-backend:latest` | Backend API container |
| Frontend | Static files (no Docker) | Served via Nginx |

---

## 🌍 Domain & SSL Configuration

### Domains

| Domain | IP | Purpose | SSL |
|--------|-----|---------|-----|
| `pedeve-dev.aretaamany.com` | `34.128.123.1` | Frontend | ✅ Let's Encrypt |
| `api-pedeve-dev.aretaamany.com` | `34.101.49.147` | Backend API | ✅ Let's Encrypt |

### SSL Certificates

| Domain | Certificate | Auto-renewal | Expires |
|--------|-------------|--------------|---------|
| `pedeve-dev.aretaamany.com` | Let's Encrypt | ✅ Enabled | 2026-02-25 |
| `api-pedeve-dev.aretaamany.com` | Let's Encrypt | ✅ Enabled | 2026-02-25 |

### Nginx Configuration

| VM | Config File | Purpose | Ports |
|----|-------------|---------|-------|
| Frontend | `/etc/nginx/sites-available/default` | Serve static files + SPA routing | 80 → 443 |
| Backend | `/etc/nginx/sites-available/backend-api` | Reverse proxy to backend | 80 → 8080, 443 → 8080 |

---

## 🔐 Environment Variables

### Backend (Production)

| Variable | Value | Source | Purpose |
|----------|-------|--------|---------|
| `GCP_PROJECT_ID` | `pedeve-pertamina-dms` | Environment | GCP project identifier |
| `GCP_SECRET_MANAGER_ENABLED` | `false` | Environment | Disable in-container Secret Manager |
| `GCP_STORAGE_ENABLED` | `true` | Environment | Enable GCP Cloud Storage |
| `GCP_STORAGE_BUCKET` | `pedeve-dev-bucket` | Environment | Storage bucket name |
| `DATABASE_URL` | `postgres://...@127.0.0.1:5432/...` | Secret Manager (via script) | Database connection |
| `JWT_SECRET` | `***` | Secret Manager (via script) | JWT token secret |
| `ENCRYPTION_KEY` | `***` (optional) | Secret Manager (via script) | Encryption key |
| `PORT` | `8080` | Environment | Backend API port |
| `ENV` | `production` | Environment | Environment mode |
| `DISABLE_RATE_LIMIT` | `true` | Environment | Disable rate limiting (dev) |
| `CORS_ORIGIN` | `https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com` | Environment | Allowed CORS origins |

### Frontend (Build Time)

| Variable | Value | Purpose |
|----------|-------|---------|
| `VITE_API_URL` | `https://api-pedeve-dev.aretaamany.com/api/v1` | Backend API URL |

### Backend (Local Development)

| Variable | Value | Purpose |
|----------|-------|---------|
| `DATABASE_URL` | `postgres://postgres:dms_password@postgres:5432/db_dms_pedeve?sslmode=disable` | Local database |
| `PORT` | `8080` | Backend port |
| `ENV` | `development` | Development mode |
| `VAULT_ADDR` | `http://vault:8200` | Vault address |
| `VAULT_TOKEN` | `dev-root-token-12345` | Vault token (dev) |

---

## 📦 Deployment Scripts

### Backend Deployment

| Script | Location | Purpose |
|--------|----------|---------|
| `deploy-backend-vm.sh` | `scripts/deploy-backend-vm.sh` | Deploy backend container to VM |
| `setup-backend-nginx.sh` | `scripts/setup-backend-nginx.sh` | Setup Nginx reverse proxy |
| `setup-backend-ssl.sh` | `scripts/setup-backend-ssl.sh` | Setup SSL certificate |

### Frontend Deployment

| Script | Location | Purpose |
|--------|----------|---------|
| `setup-nginx-frontend.sh` | `scripts/setup-nginx-frontend.sh` | Setup Nginx for static files |
| `setup-frontend-ssl.sh` | `scripts/setup-frontend-ssl.sh` | Setup SSL certificate |

### Utility Scripts

| Script | Location | Purpose |
|--------|----------|---------|
| `fix-nginx-conflict.sh` | `scripts/fix-nginx-conflict.sh` | Fix Nginx conflicting server name |
| `setup-backend-firewall.sh` | `scripts/setup-backend-firewall.sh` | Setup firewall rules |

---

## 🔄 Deployment Flow

### Automated (CI/CD)

| Step | Action | Tool |
|------|--------|------|
| 1. Code Push | Push to `development` branch | Git |
| 2. Build | Build backend Docker image | Docker |
| 3. Build Frontend | Build static files | Vite |
| 4. Push Image | Push to GHCR | Docker |
| 5. Authenticate | WIF authentication | `google-github-actions/auth` |
| 6. Deploy Backend | Copy image, run container | `gcloud compute scp/ssh` |
| 7. Deploy Frontend | Copy static files, setup Nginx | `gcloud compute scp/ssh` |
| 8. Health Check | Verify deployment | `curl` |

### Manual Steps (One-time Setup)

| Step | Action | Command/Script |
|------|--------|----------------|
| 1. Setup Firewall | Create firewall rules | `scripts/setup-backend-firewall.sh` or GCP Console |
| 2. Setup Nginx Backend | Configure reverse proxy | `scripts/setup-backend-nginx.sh` |
| 3. Setup SSL Backend | Get SSL certificate | `scripts/setup-backend-ssl.sh` |
| 4. Setup Nginx Frontend | Configure static file serving | `scripts/setup-nginx-frontend.sh` |
| 5. Setup SSL Frontend | Get SSL certificate | `scripts/setup-frontend-ssl.sh` |
| 6. Run Seeder | Seed sample data | Build binary and run in container |

---

## 🗄️ Database Configuration

### Cloud SQL Connection

| Parameter | Value |
|-----------|-------|
| Connection String | `pedeve-pertamina-dms:asia-southeast2:postgres-dev=tcp:5432` |
| Proxy Port | `127.0.0.1:5432` |
| Proxy Status | Running on backend VM |
| IAM Role | Cloud SQL Client |

### Local Database Connection

| Parameter | Value |
|-----------|-------|
| Host | `postgres` (Docker service name) |
| Port | `5432` |
| Database | `db_dms_pedeve` |
| User | `postgres` |
| Password | `dms_password` |

---

## 🔒 Security Configuration

### Cookie Settings (Production)

| Parameter | Value | Reason |
|-----------|-------|--------|
| `HttpOnly` | `true` | Prevent XSS |
| `Secure` | `true` | HTTPS only |
| `SameSite` | `None` | Cross-site requests (frontend/backend different subdomains) |
| `Path` | `/` | Available for all paths |
| `MaxAge` | `86400` (24 hours) | Cookie expiry |

### CORS Configuration

| Parameter | Value |
|-----------|-------|
| `AllowOrigins` | `https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com` |
| `AllowMethods` | `GET,POST,PUT,DELETE,OPTIONS,PATCH` |
| `AllowHeaders` | `Accept,Authorization,Content-Type,X-CSRF-Token,X-Requested-With` |
| `AllowCredentials` | `true` |
| `MaxAge` | `300` seconds |

### Rate Limiting

| Environment | Status | Configuration |
|-------------|--------|---------------|
| Development | Disabled | `DISABLE_RATE_LIMIT=true` |
| Production | Enabled (default) | 100 req/s, burst 50 (general), 5 req/min (auth) |

---

## 📁 File Storage

### Production (GCP Cloud Storage)

| Parameter | Value |
|-----------|-------|
| Provider | GCP Cloud Storage |
| Bucket | `pedeve-dev-bucket` |
| Base Path | `logos/`, `documents/` |
| Public URL Format | `https://storage.googleapis.com/pedeve-dev-bucket/{path}/{filename}` |

### Development (Local)

| Parameter | Value |
|-----------|-------|
| Provider | Local filesystem |
| Base Path | `backend/uploads/` |
| URL Format | `/uploads/{path}/{filename}` |

---

## 🛠️ Tools & Utilities

### Database Seeder

| Script | Location | Purpose |
|--------|----------|---------|
| `seed-companies` | `backend/cmd/seed-companies/main.go` | Seed 11 companies (1 holding + 10 subsidiaries) with admin users |

### Migration Scripts

| Script | Location | Purpose |
|--------|----------|---------|
| `migrate-sqlite-to-postgres` | `backend/cmd/migrate-sqlite-to-postgres/main.go` | Migrate data from SQLite to PostgreSQL |
| `create-schema` | `backend/cmd/create-schema/main.go` | Create database schema |
| `fix-company-levels` | `backend/cmd/fix-company-levels/main.go` | Fix company hierarchy levels |

---

## 📝 Important Notes

### Production Setup Checklist

- [ ] Create VMs (frontend & backend)
- [ ] Setup Cloud SQL PostgreSQL instance
- [ ] Create Cloud Storage bucket
- [ ] Setup Secret Manager secrets
- [ ] Configure Workload Identity Federation
- [ ] Setup GitHub Secrets
- [ ] Configure DNS (A records for domains)
- [ ] Setup firewall rules
- [ ] Apply network tags to VMs
- [ ] Deploy backend (auto via CI/CD)
- [ ] Deploy frontend (auto via CI/CD)
- [ ] Setup Nginx on both VMs
- [ ] Setup SSL certificates (Certbot)
- [ ] Run database seeder (manual)
- [ ] Verify health checks
- [ ] Test login/logout flow

### Key Differences: Development vs Production

| Aspect | Development | Production |
|--------|-------------|------------|
| Database | Local PostgreSQL (Docker) | Cloud SQL PostgreSQL |
| Storage | Local filesystem | GCP Cloud Storage |
| Secrets | Vault / Environment | GCP Secret Manager |
| Frontend | Vite dev server | Static files via Nginx |
| Backend | Hot reload (Air) | Docker container |
| SSL | No | Let's Encrypt |
| Rate Limiting | Disabled | Enabled (configurable) |
| Cookie SameSite | Lax | None (cross-site) |

---

## 🔗 Quick Reference URLs

| Service | URL |
|--------|-----|
| Frontend | `https://pedeve-dev.aretaamany.com` |
| Backend API | `https://api-pedeve-dev.aretaamany.com/api/v1` |
| Health Check | `https://api-pedeve-dev.aretaamany.com/health` |
| Swagger Docs | `https://api-pedeve-dev.aretaamany.com/swagger/index.html` |

---

**Last Updated:** 2025-11-28  
**Environment:** Development  
**Status:** ✅ All configurations tested and working

