# 🚀 Roadmap Setup Server Production GCP

Roadmap lengkap untuk implementasi setup server production di Google Cloud Platform. Setup ini akan mengikuti pola yang sama dengan server development, dengan perbedaan utama: **menggunakan branch `main` untuk deployment**.

---

## 📋 Overview

**⚠️ PENTING: Production menggunakan akun GCP terpisah (milik client), berbeda dengan development!**

Setup production akan menggunakan:
- **Project GCP:** `pedeve-production` ⚠️ **BERBEDA dengan development**
- **Project Number:** `<PROJECT_NUMBER_PRODUCTION>` ⚠️ **BERBEDA dengan development** (akan diisi saat setup)
- **Region:** `asia-southeast2` (Jakarta) - bisa berbeda sesuai kebutuhan
- **Branch Deployment:** `main` (bukan `development`)
- **Naming Convention:** Semua resources menggunakan suffix `-prod` (bukan `-dev`)
- **Domain:** Akan berbeda dengan development (akan diisi saat setup)
- **SSH User:** Mungkin berbeda dengan development (akan dikonfirmasi saat setup)

**Perbedaan Utama dengan Development:**
- ✅ Project GCP terpisah (akun client)
- ✅ Service Account terpisah
- ✅ WIF setup baru (tidak bisa reuse dari development)
- ✅ Domain berbeda
- ✅ Semua resources di project GCP yang berbeda

---

## 🎯 Fase 1: Persiapan & Planning

### 1.1 Identifikasi Resources yang Perlu Dibuat

**Virtual Machines:**
- [ ] `backend-prod` - Backend VM untuk production
- [ ] `frontend-prod` - Frontend VM untuk production

**Cloud SQL:**
- [ ] `postgres-prod` - PostgreSQL instance untuk production

**Cloud Storage:**
- [ ] `pedeve-prod-bucket` - Storage bucket untuk production

**Secret Manager:**
- [ ] Semua secrets dengan prefix `prod-` atau suffix `-prod`

**DNS:**
- [ ] Domain untuk production:
  - Frontend: `reports.pertamina-pedeve.co.id`
  - Backend: `api-reports.pertamina-pedeve.co.id`

### 1.2 Tentukan Spesifikasi Resources

**Backend VM:**
- Machine Type: `e2-medium` atau lebih besar (sesuai kebutuhan production)
- Region: `asia-southeast2-a`
- External IP: Akan di-assign otomatis

**Frontend VM:**
- Machine Type: `e2-micro` atau `e2-small` (sesuai kebutuhan)
- Region: `asia-southeast2-a`
- External IP: Akan di-assign otomatis

**Cloud SQL:**
- Engine: PostgreSQL 16
- Machine Type: Custom (sesuaikan dengan beban production)
- Storage: Minimal 20GB SSD (lebih besar dari development)

**Storage Bucket:**
- Location: `asia-southeast2`
- Storage Class: Standard
- Public Access: Not public (Private)

---

## 🎯 Fase 2: Setup GCP Resources

### 2.1 Create Cloud SQL Instance

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# ⚠️ CATATAN: --enable-bin-log adalah flag MySQL, TIDAK VALID untuk PostgreSQL!
# Untuk PostgreSQL, gunakan automated backups + point-in-time recovery (PITR)

# Opsi 1: Via Console (RECOMMENDED untuk PostgreSQL)
# 1. Buka GCP Console → Cloud SQL → Create Instance
# 2. Pilih PostgreSQL 16
# 3. Enable "Automated backups" + "Point-in-time recovery"
# 4. Set backup window: 02:00
# 5. Maintenance window: Sunday 03:00

# Opsi 2: Via CLI (pastikan flags valid untuk PostgreSQL)
gcloud sql instances create postgres-prod \
  --database-version=POSTGRES_16 \
  --tier=db-custom-1-3840 \
  --region=asia-southeast2 \
  --root-password=<strong-password> \
  --storage-type=SSD \
  --storage-size=20GB \
  --storage-auto-increase \
  --backup-start-time=02:00 \
  --enable-point-in-time-recovery \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=03 \
  --project=pedeve-production

# ⚠️ SECURITY: Untuk production, pertimbangkan Private IP (lebih aman)
# Opsi A: Public IP + Authorized Networks (sama seperti dev, cepat tapi kurang aman)
gcloud sql instances patch postgres-prod \
  --assign-ip \
  --project=pedeve-production

# Setelah backend VM dibuat, tambahkan authorized network:
# gcloud sql instances patch postgres-prod \
#   --authorized-networks=<backend-prod-ip>/32 \
#   --project=pedeve-production

# Opsi B: Private IP (RECOMMENDED untuk production, lebih aman)
# Perlu setup VPC peering atau Private Service Connect
# Lebih kompleks tapi lebih secure
```

**Todo:**
- [ ] Create Cloud SQL instance `postgres-prod` (via Console recommended untuk PostgreSQL)
- [ ] Enable automated backups + Point-in-time recovery
- [ ] Pilih: Public IP + Authorized Networks (quick) ATAU Private IP (secure)
- [ ] Jika Public IP: Setup authorized networks hanya IP backend-prod (setelah VM dibuat)
- [ ] Create database `db_prod_pedeve`
- [ ] Create database user `pedeve_user_db_prod`

---

### 2.2 Create Storage Bucket

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Create storage bucket untuk production
gsutil mb -p pedeve-production \
  -c STANDARD \
  -l asia-southeast2 \
  gs://pedeve-prod-bucket

# Set uniform access control
gsutil uniformbucketlevelaccess set on gs://pedeve-prod-bucket

# Set public access prevention (optional, untuk security)
gsutil pap set enforced gs://pedeve-prod-bucket
```

**Todo:**
- [ ] Create bucket `pedeve-prod-bucket`
- [ ] Set uniform access control
- [ ] Configure CORS (setelah frontend domain diketahui)
- [ ] Setup IAM permissions untuk service account

---

### 2.3 Create Backend VM

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Create backend VM untuk production
gcloud compute instances create backend-prod \
  --zone=asia-southeast2-a \
  --machine-type=e2-medium \
  --network-interface=network-tier=PREMIUM,subnet=default \
  --maintenance-policy=MIGRATE \
  --provisioning-model=STANDARD \
  --service-account=<service-account-email> \
  --scopes=https://www.googleapis.com/auth/cloud-platform \
  --tags=backend-api-server,https-server,http-server \
  --create-disk=auto-delete=yes,boot=yes,device-name=backend-prod,image=projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20240111,mode=rw,size=20,type=projects/pedeve-production/zones/asia-southeast2-a/diskTypes/pd-standard \
  --no-shielded-secure-boot \
  --shielded-vtpm \
  --shielded-integrity-monitoring \
  --labels=env=production,role=backend \
  --reservation-affinity=any \
  --project=pedeve-production
```

**Todo:**
- [ ] Create Service Account untuk VM runtime: `vm-backend-prod@pedeve-production.iam.gserviceaccount.com`
- [ ] Grant IAM roles ke VM SA: Secret Manager Secret Accessor, Storage Object Admin, Cloud SQL Client
- [ ] Create backend VM `backend-prod` dengan VM SA (bukan GitHub Actions SA)
- [ ] Assign external IP (akan otomatis)
- [ ] Setup OS Login untuk SSH access
- [ ] Install Docker
- [ ] Install Cloud SQL Auth Proxy
- [ ] Setup systemd service untuk Cloud SQL Auth Proxy
- [ ] Install Nginx

---

### 2.4 Create Frontend VM

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Create frontend VM untuk production
gcloud compute instances create frontend-prod \
  --zone=asia-southeast2-a \
  --machine-type=e2-micro \
  --network-interface=network-tier=PREMIUM,subnet=default \
  --maintenance-policy=MIGRATE \
  --provisioning-model=STANDARD \
  --service-account=vm-frontend-prod@pedeve-production.iam.gserviceaccount.com \
  --scopes=https://www.googleapis.com/auth/cloud-platform \
  --tags=https-server,http-server \
  --create-disk=auto-delete=yes,boot=yes,device-name=frontend-prod,image=projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20240111,mode=rw,size=10,type=projects/pedeve-production/zones/asia-southeast2-a/diskTypes/pd-standard \
  --no-shielded-secure-boot \
  --shielded-vtpm \
  --shielded-integrity-monitoring \
  --labels=env=production,role=frontend \
  --reservation-affinity=any \
  --project=pedeve-production
```

**Todo:**
- [ ] Create Service Account untuk VM runtime: `vm-frontend-prod@pedeve-production.iam.gserviceaccount.com`
- [ ] Grant IAM roles ke VM SA: Storage Object Viewer (jika perlu akses storage)
- [ ] Create frontend VM `frontend-prod` dengan VM SA
- [ ] Assign external IP (akan otomatis)
- [ ] Setup OS Login untuk SSH access
- [ ] Install Nginx

---

### 2.5 Setup Firewall Rules

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Backend API (port 8080)
gcloud compute firewall-rules create allow-backend-api-prod \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags backend-api-server \
  --description "Allow Backend API traffic on port 8080 (Production)" \
  --project=pedeve-production

# HTTPS (port 443)
gcloud compute firewall-rules create allow-https-prod \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic (Production)" \
  --project=pedeve-production

# HTTP (port 80) - untuk redirect ke HTTPS
# ⚠️ PENTING: Pastikan VM sudah diberi tag http-server
gcloud compute firewall-rules create allow-http-prod \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server \
  --description "Allow HTTP traffic (Production)" \
  --project=pedeve-production
```

**Todo:**
- [ ] Create firewall rule `allow-backend-api-prod` (port 8080, tag: backend-api-server)
- [ ] Create firewall rule `allow-https-prod` (port 443, tag: https-server)
- [ ] Create firewall rule `allow-http-prod` (port 80, tag: http-server)
- [ ] Pastikan VM backend-prod punya tags: backend-api-server, https-server, http-server
- [ ] Pastikan VM frontend-prod punya tags: https-server, http-server
- [ ] Test connectivity

---

## 🎯 Fase 3: Setup Secret Manager

### 3.1 Create Secrets untuk Production

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Database secrets
# ⚠️ PENTING: Jika menggunakan Cloud SQL Proxy, host SELALU 127.0.0.1
# Jangan simpan db_host_prod dengan IP public (tidak akan dipakai, malah bikin bingung)

echo -n "<db-password>" | gcloud secrets create db_password_prod \
  --data-file=- \
  --project=pedeve-production

echo -n "pedeve_user_db_prod" | gcloud secrets create db_user_prod \
  --data-file=- \
  --project=pedeve-production

echo -n "db_prod_pedeve" | gcloud secrets create db_name_prod \
  --data-file=- \
  --project=pedeve-production

# Host dan port tidak perlu di secret (selalu 127.0.0.1:5432 via proxy)
# Atau simpan sebagai DATABASE_URL lengkap jika lebih praktis:
# echo -n "postgres://pedeve_user_db_prod:<password>@127.0.0.1:5432/db_prod_pedeve?sslmode=require" | \
#   gcloud secrets create database_url_prod --data-file=- --project=pedeve-production

echo -n "5432" | gcloud secrets create db_port_prod \
  --data-file=- \
  --project=pedeve-production

# Application secrets
echo -n "<jwt-secret>" | gcloud secrets create jwt_secret_prod \
  --data-file=- \
  --project=pedeve-production

echo -n "<encryption-key>" | gcloud secrets create encryption_key_prod \
  --data-file=- \
  --project=pedeve-production
```

**Todo:**
- [ ] Create secret `db_password_prod`
- [ ] Create secret `db_user_prod`
- [ ] Create secret `db_name_prod`
- [ ] Create secret `db_port_prod` (atau skip, selalu 5432)
- [ ] **SKIP** `db_host_prod` (tidak perlu, host selalu 127.0.0.1 via proxy)
- [ ] Create secret `jwt_secret_prod`
- [ ] Create secret `encryption_key_prod`
- [ ] Grant access ke VM backend service account (`vm-backend-prod@...`)

---

## 🎯 Fase 4: Setup Database

### 4.1 Create Database & User

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# Connect ke Cloud SQL instance
gcloud sql connect postgres-prod \
  --user=postgres \
  --project=pedeve-production

# Di dalam SQL console:
CREATE DATABASE db_prod_pedeve;
CREATE USER pedeve_user_db_prod WITH PASSWORD '<strong-password>';
GRANT ALL PRIVILEGES ON DATABASE db_prod_pedeve TO pedeve_user_db_prod;
\q
```

**Todo:**
- [ ] Create database `db_prod_pedeve`
- [ ] Create user `pedeve_user_db_prod`
- [ ] Grant permissions
- [ ] Run database migrations
- [ ] Run database seeder (jika perlu)

---

## 🎯 Fase 5: Setup Workload Identity Federation (WIF)

### 5.1 Setup WIF Baru untuk Production

**⚠️ PENTING: Production menggunakan project GCP terpisah, jadi WIF harus di-setup baru!**

WIF development tidak bisa digunakan untuk production karena:
- Project GCP berbeda
- Service Account berbeda
- IAM permissions terpisah

**Langkah Setup WIF untuk Production:**

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# 1. Enable required APIs
gcloud services enable iamcredentials.googleapis.com \
  --project=pedeve-production

gcloud services enable sts.googleapis.com \
  --project=pedeve-production

# 2. Get project number
PROJECT_NUMBER=$(gcloud projects describe pedeve-production --format="value(projectNumber)")
echo "Project Number: $PROJECT_NUMBER"

# 3. Create Workload Identity Pool
gcloud iam workload-identity-pools create github-actions-pool \
  --location=global \
  --project=pedeve-production \
  --display-name="GitHub Actions Pool"

# 4. Create Workload Identity Provider
# ⚠️ PENTING: Format attribute mapping harus benar (key: attribute.<name>)
# ⚠️ PENTING: Attribute condition harus restrict branch untuk security
gcloud iam workload-identity-pools providers create-oidc github-actions-provider \
  --workload-identity-pool=github-actions-pool \
  --location=global \
  --project=pedeve-production \
  --display-name="GitHub Actions Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.actor=assertion.actor,attribute.ref=assertion.ref" \
  --attribute-condition="assertion.repository=='repoareta/pedeve-dms-app' && attribute.ref.startsWith('refs/heads/main')" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# 5. Create Service Account untuk GitHub Actions (untuk deployment)
gcloud iam service-accounts create github-actions-deployer \
  --display-name="GitHub Actions Deployer" \
  --project=pedeve-production

# ⚠️ PENTING: Pisahkan Service Account:
# - github-actions-deployer: untuk deployment (SSH, compute ops)
# - vm-backend-prod: untuk runtime backend (Secret Manager, Storage, Cloud SQL)
# - vm-frontend-prod: untuk runtime frontend (jika perlu akses storage)

# 6. Grant IAM roles ke Service Account
gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/compute.instanceAdmin.v1"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/compute.osLogin"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:github-actions-deployer@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# 7. Allow GitHub Actions to impersonate Service Account
gcloud iam service-accounts add-iam-policy-binding \
  github-actions-deployer@pedeve-production.iam.gserviceaccount.com \
  --project=pedeve-production \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions-pool/attribute.repository/repoareta/pedeve-dms-app"
```

**WIF Provider Path untuk GitHub Secrets:**
```
projects/<PROJECT_NUMBER_PRODUCTION>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider
```

**⚠️ PENTING tentang Audience:**
- Jangan pakai URL repo sebagai audience (akan error mismatch)
- Gunakan provider resource name (default) atau format: `https://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- Di GitHub Actions, `google-github-actions/auth` akan otomatis menggunakan provider resource name

**Todo:**
- [ ] Enable required APIs
- [ ] Get project number production
- [ ] Create Workload Identity Pool
- [ ] Create Workload Identity Provider
- [ ] Create Service Account untuk GitHub Actions
- [ ] Grant IAM roles ke Service Account
- [ ] Allow GitHub Actions to impersonate Service Account
- [ ] Test authentication dari GitHub Actions

---

## 🎯 Fase 6: Setup Service Accounts untuk VM Runtime

### 6.1 Create VM Service Accounts (Terpisah dari GitHub Actions SA)

**⚠️ PENTING: Pisahkan Service Account untuk security (blast radius lebih kecil):**
- **GitHub Actions SA:** Hanya untuk deployment (SSH, compute ops)
- **VM Runtime SA:** Untuk runtime aplikasi (Secret Manager, Storage, Cloud SQL)

```bash
# ⚠️ PASTIKAN: Set project production terlebih dahulu
gcloud config set project pedeve-production

# 1. Create Service Account untuk Backend VM
gcloud iam service-accounts create vm-backend-prod \
  --display-name="Backend VM Runtime Service Account" \
  --project=pedeve-production

# 2. Grant IAM roles untuk Backend VM SA
gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:vm-backend-prod@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:vm-backend-prod@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:vm-backend-prod@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# 3. Create Service Account untuk Frontend VM (jika perlu akses storage)
gcloud iam service-accounts create vm-frontend-prod \
  --display-name="Frontend VM Runtime Service Account" \
  --project=pedeve-production

# 4. Grant IAM roles untuk Frontend VM SA (jika perlu)
gcloud projects add-iam-policy-binding pedeve-production \
  --member="serviceAccount:vm-frontend-prod@pedeve-production.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"
```

**Todo:**
- [ ] Create Service Account `vm-backend-prod@pedeve-production.iam.gserviceaccount.com`
- [ ] Grant roles: Secret Manager Secret Accessor, Storage Object Admin, Cloud SQL Client
- [ ] Create Service Account `vm-frontend-prod@pedeve-production.iam.gserviceaccount.com` (jika perlu)
- [ ] Grant roles: Storage Object Viewer (jika perlu)

---

## 🎯 Fase 7: Setup GitHub Secrets & Environments untuk Production

### 7.1 Setup GitHub Environments (RECOMMENDED - Best Practice)

**⚠️ PENTING: Gunakan GitHub Environments untuk memisahkan dev vs prod secrets!**

**Langkah:**
1. Buka GitHub Repository → **Settings → Environments**
2. Create Environment: `production`
3. Tambahkan secrets di Environment `production` (bukan Repository secrets)

**Secrets di Environment `production`:**

| Secret Name | Value | Notes |
|-------------|-------|-------|
| `GCP_PROJECT_ID` | `pedeve-production` | Project ID production |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` | WIF Provider path |
| `GCP_SERVICE_ACCOUNT` | `github-actions-deployer@pedeve-production.iam.gserviceaccount.com` | GitHub Actions SA |
| `GCP_BACKEND_VM_IP` | `<backend-prod-ip>` | Backend VM IP |
| `GCP_FRONTEND_VM_IP` | `<frontend-prod-ip>` | Frontend VM IP |
| `GCP_SSH_USER` | `<ssh-user>` | SSH user |
| `GCP_PROJECT_NUMBER` | `<PROJECT_NUMBER>` | Project number |
| `GHCR_TOKEN` | (existing) | Bisa reuse dari repository secrets |

**Keuntungan GitHub Environments:**
- ✅ Secrets terpisah per environment
- ✅ Protection rules (required reviewers, deployment branches)
- ✅ Audit trail lebih jelas
- ✅ Tidak perlu suffix `_PROD` (otomatis terpisah)

### 7.2 Alternatif: Repository Secrets dengan Suffix `_PROD`

Jika tidak menggunakan Environments, gunakan suffix `_PROD`:

| Secret Name | Value | Notes |
|-------------|-------|-------|
| `GCP_PROJECT_ID_PROD` | `pedeve-production` | **NEW** - Project ID production |
| `GCP_WORKLOAD_IDENTITY_PROVIDER_PROD` | `projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` | **NEW** - WIF Provider path |
| `GCP_SERVICE_ACCOUNT_PROD` | `github-actions-deployer@pedeve-production.iam.gserviceaccount.com` | **NEW** - Service account |
| `GCP_BACKEND_VM_IP_PROD` | `<backend-prod-ip>` | **NEW** - IP backend production VM |
| `GCP_FRONTEND_VM_IP_PROD` | `<frontend-prod-ip>` | **NEW** - IP frontend production VM |
| `GCP_SSH_USER_PROD` | `<ssh-user>` | **NEW** - SSH user |
| `GCP_PROJECT_NUMBER_PROD` | `<PROJECT_NUMBER>` | **NEW** - Project number |

**Todo:**
- [ ] Setup GitHub Environment `production` (RECOMMENDED)
- [ ] Atau tambahkan secrets dengan suffix `_PROD` di Repository secrets
- [ ] Add semua secrets production (setelah resources dibuat)

---

## 🎯 Fase 8: Update CI/CD Workflow

### 8.1 Update GitHub Actions Workflow

File: `.github/workflows/ci-cd.yml`

**Perubahan yang diperlukan:**
1. Tambahkan job `deploy-gcp-prod` yang trigger pada branch `main`
2. **Jika pakai GitHub Environments:** Gunakan `environment: production` (secrets otomatis terpisah)
3. **Jika pakai Repository secrets:** Gunakan secrets dengan suffix `_PROD`
4. Update environment variables untuk production (project ID, bucket, domain)
5. Update VM IPs untuk production
6. Update bucket name untuk production: `pedeve-prod-bucket`
7. Update domain untuk production:
   - Frontend: `reports.pertamina-pedeve.co.id`
   - Backend: `api-reports.pertamina-pedeve.co.id`
8. Update WIF authentication untuk production
9. Update `VITE_API_URL` untuk production build

**Todo:**
- [ ] Setup GitHub Environment `production` (jika pakai Environments)
- [ ] Update workflow untuk support deployment ke production
- [ ] Add condition untuk branch `main`
- [ ] Add `environment: production` ke job deploy (jika pakai Environments)
- [ ] Update backend VM IP
- [ ] Update frontend VM IP
- [ ] Update bucket name ke `pedeve-prod-bucket`
- [ ] Update domain untuk production
- [ ] Update CORS origin untuk production domain
- [ ] Update `VITE_API_URL` untuk production build: `https://api-reports.pertamina-pedeve.co.id/api/v1`

---

## 🎯 Fase 9: Setup DNS

### 9.1 Configure DNS Records

| Domain | IP | Type | Purpose |
|--------|-----|------|---------|
| `reports.pertamina-pedeve.co.id` | `<frontend-prod-ip>` | A | Frontend production |
| `api-reports.pertamina-pedeve.co.id` | `<backend-prod-ip>` | A | Backend API production |

**Todo:**
- [ ] Create A record untuk `reports.pertamina-pedeve.co.id`
- [ ] Create A record untuk `api-reports.pertamina-pedeve.co.id`
- [ ] Verify DNS propagation
- [ ] Test connectivity

---

## 🎯 Fase 10: Setup SSL/TLS

### 10.1 Install SSL Certificates

**Di Backend VM:**
```bash
# SSH ke backend VM
gcloud compute ssh info@aretaamany.com@backend-prod --zone=asia-southeast2-a

# Install Certbot
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d api-reports.pertamina-pedeve.co.id

# Test auto-renewal
sudo certbot renew --dry-run
```

**Di Frontend VM:**
```bash
# SSH ke frontend VM
gcloud compute ssh info@aretaamany.com@frontend-prod --zone=asia-southeast2-a

# Install Certbot
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d reports.pertamina-pedeve.co.id

# Test auto-renewal
sudo certbot renew --dry-run
```

**Todo:**
- [ ] Install Certbot di backend VM
- [ ] Get SSL certificate untuk `api-reports.pertamina-pedeve.co.id`
- [ ] Install Certbot di frontend VM
- [ ] Get SSL certificate untuk `reports.pertamina-pedeve.co.id`
- [ ] Setup auto-renewal
- [ ] Verify SSL certificates

---

## 🎯 Fase 11: Setup Nginx

### 11.1 Configure Nginx di Backend VM

**File:** `/etc/nginx/sites-available/backend-api`

**Todo:**
- [ ] Create Nginx config untuk backend API
- [ ] Setup reverse proxy ke port 8080
- [ ] Configure SSL
- [ ] Test Nginx configuration
- [ ] Reload Nginx

---

### 11.2 Configure Nginx di Frontend VM

**File:** `/etc/nginx/sites-available/default`

**Todo:**
- [ ] Create Nginx config untuk frontend
- [ ] Setup static file serving
- [ ] Configure SPA routing
- [ ] Configure SSL
- [ ] Test Nginx configuration
- [ ] Reload Nginx

---

## 🎯 Fase 12: Setup Cloud SQL Auth Proxy

### 12.1 Install & Configure di Backend VM

```bash
# SSH ke backend VM
gcloud compute ssh info@aretaamany.com@backend-prod --zone=asia-southeast2-a

# Download Cloud SQL Auth Proxy
wget https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.linux.amd64 -O cloud-sql-proxy
chmod +x cloud-sql-proxy
sudo mv cloud-sql-proxy /usr/local/bin/

# Create systemd service
sudo nano /etc/systemd/system/cloud-sql-proxy.service
```

**Systemd service file (Cloud SQL Proxy v2 - format terbaik):**
```ini
[Unit]
Description=Cloud SQL Auth Proxy
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/cloud-sql-proxy \
  --address 127.0.0.1 \
  --port 5432 \
  pedeve-production:asia-southeast2:postgres-prod
Restart=always

[Install]
WantedBy=multi-user.target
```

**⚠️ CATATAN:**
- Cloud SQL Proxy v2 menggunakan format `--address` dan `--port` (lebih bersih)
- Instance connection name tanpa `tcp:127.0.0.1:5432` style v1
- Format ini lebih jelas dan mengurangi kesalahan konfigurasi

**Todo:**
- [ ] Download Cloud SQL Auth Proxy
- [ ] Create systemd service
- [ ] Enable & start service
- [ ] Verify connection

---

## 🎯 Fase 13: Setup Environment Variables

### 13.1 Backend Environment Variables

**Update deployment script atau docker-compose untuk production:**

```bash
# Database (via Cloud SQL Proxy, host SELALU 127.0.0.1)
DATABASE_URL=postgres://pedeve_user_db_prod:<secret>@127.0.0.1:5432/db_prod_pedeve?sslmode=require
DB_HOST=127.0.0.1  # SELALU 127.0.0.1 (via proxy)
DB_PORT=5432        # SELALU 5432 (via proxy)
DB_USER=pedeve_user_db_prod
DB_PASSWORD=<from-secret-manager>
DB_NAME=db_prod_pedeve

# GCP
GCP_PROJECT_ID=pedeve-production
GCP_SECRET_MANAGER_ENABLED=true
GCP_STORAGE_BUCKET=pedeve-prod-bucket
GCP_STORAGE_ENABLED=true

# Secrets (from Secret Manager)
JWT_SECRET=<from-secret-manager>
ENCRYPTION_KEY=<from-secret-manager>

# Application
PORT=8080
ENV=production
CORS_ORIGIN=https://reports.pertamina-pedeve.co.id
```

**Todo:**
- [ ] Update backend environment variables
- [ ] Update secret names untuk production
- [ ] Verify Secret Manager access

---

### 13.2 Frontend Environment Variables

**Update build-time variables:**

```bash
VITE_API_URL=https://api-reports.pertamina-pedeve.co.id/api/v1
NODE_ENV=production
```

**Todo:**
- [ ] Update `VITE_API_URL` untuk production
- [ ] Update CI/CD workflow untuk production build

---

## 🎯 Fase 14: Setup CORS untuk Storage Bucket

### 14.1 Configure CORS

```bash
# Create CORS config file
cat > cors-config-prod.json <<'EOF'
[
  {
    "origin": [
      "https://reports.pertamina-pedeve.co.id",
      "http://reports.pertamina-pedeve.co.id",
      "https://api-reports.pertamina-pedeve.co.id"
    ],
    "method": ["GET", "HEAD", "OPTIONS"],
    "responseHeader": [
      "Content-Type",
      "Access-Control-Allow-Origin",
      "Access-Control-Allow-Methods",
      "Access-Control-Allow-Headers"
    ],
    "maxAgeSeconds": 3600
  }
]
EOF

# Apply CORS config
gcloud storage buckets update gs://pedeve-prod-bucket \
  --cors-file=cors-config-prod.json \
  --project=pedeve-production
```

**Todo:**
- [ ] Create CORS config file
- [ ] Apply CORS config ke bucket
- [ ] Verify CORS configuration

---

## 🎯 Fase 15: Initial Deployment

### 15.1 Deploy Backend

**Todo:**
- [ ] Push code ke branch `main`
- [ ] Monitor GitHub Actions workflow
- [ ] Verify backend deployment
- [ ] Check backend health endpoint
- [ ] Verify database connection

---

### 15.2 Deploy Frontend

**Todo:**
- [ ] Verify frontend build successful
- [ ] Verify frontend deployment
- [ ] Test frontend access
- [ ] Verify API connection dari frontend

---

## 🎯 Fase 16: Verification & Testing

### 16.1 Health Checks

**Todo:**
- [ ] Test backend health: `https://api-reports.pertamina-pedeve.co.id/health`
- [ ] Test frontend: `https://reports.pertamina-pedeve.co.id`
- [ ] Test API endpoints
- [ ] Test file uploads
- [ ] Test authentication flow
- [ ] Test SSL certificates

---

### 16.2 Monitoring

**Todo:**
- [ ] Setup monitoring & alerting (optional)
- [ ] Setup log aggregation (optional)
- [ ] Document production URLs
- [ ] Create runbook untuk troubleshooting

---

## 📋 Checklist Lengkap

### Phase 1: Planning ✅
- [x] Identifikasi resources yang perlu dibuat
- [x] Tentukan spesifikasi resources

### Phase 2: GCP Resources
- [ ] Create Cloud SQL instance `postgres-prod`
- [ ] Create storage bucket `pedeve-prod-bucket`
- [ ] Create backend VM `backend-prod`
- [ ] Create frontend VM `frontend-prod`
- [ ] Setup firewall rules

### Phase 3: Secret Manager
- [ ] Create semua secrets untuk production
- [ ] Grant access ke service account

### Phase 4: Database
- [ ] Create database `db_prod_pedeve`
- [ ] Create user `pedeve_user_db_prod`
- [ ] Run migrations
- [ ] Run seeder (jika perlu)

### Phase 5: WIF
- [ ] Setup WIF baru untuk production
- [ ] Test authentication

### Phase 6: VM Service Accounts
- [ ] Create VM backend SA
- [ ] Create VM frontend SA
- [ ] Grant IAM roles

### Phase 7: GitHub Secrets & Environments
- [ ] Add `GCP_BACKEND_VM_IP_PROD`
- [ ] Add `GCP_FRONTEND_VM_IP_PROD`
- [ ] Verify semua secrets

### Phase 8: CI/CD
- [ ] Update GitHub Actions workflow
- [ ] Test deployment flow

### Phase 9: DNS
- [ ] Create DNS records
- [ ] Verify DNS propagation

### Phase 10: SSL
- [ ] Install SSL di backend VM
- [ ] Install SSL di frontend VM
- [ ] Setup auto-renewal

### Phase 11: Nginx
- [ ] Configure Nginx di backend VM
- [ ] Configure Nginx di frontend VM

### Phase 12: Cloud SQL Proxy
- [ ] Install Cloud SQL Auth Proxy
- [ ] Setup systemd service
- [ ] Verify connection

### Phase 13: Environment Variables
- [ ] Update backend env vars
- [ ] Update frontend env vars

### Phase 14: CORS
- [ ] Configure CORS untuk storage bucket

### Phase 15: Deployment
- [ ] Deploy backend
- [ ] Deploy frontend

### Phase 16: Verification
- [ ] Health checks
- [ ] End-to-end testing
- [ ] Monitoring setup

---

## 🔗 Quick Reference

### Production URLs (Setelah Setup)
- Frontend: `https://reports.pertamina-pedeve.co.id`
- Backend API: `https://api-reports.pertamina-pedeve.co.id/api/v1`
- Health Check: `https://api-reports.pertamina-pedeve.co.id/health`
- Swagger Docs: `https://api-reports.pertamina-pedeve.co.id/swagger/index.html`

### Key Differences: Development vs Production

| Aspect | Development | Production |
|--------|-------------|------------|
| Branch | `development` | `main` |
| Backend VM | `backend-dev` | `backend-prod` |
| Frontend VM | `frontend-dev` | `frontend-prod` |
| Cloud SQL | `postgres-dev` | `postgres-prod` |
| Storage Bucket | `pedeve-dev-bucket` | `pedeve-prod-bucket` |
| Database | `db_dev_pedeve` | `db_prod_pedeve` |
| Domain | `pedeve-dev.aretaamany.com` | `reports.pertamina-pedeve.co.id` (berbeda) |
| API Domain | `api-pedeve-dev.aretaamany.com` | `api-reports.pertamina-pedeve.co.id` (berbeda) |
| Secrets | `db_password`, `jwt_secret`, etc. | `db_password_prod`, `jwt_secret_prod`, etc. |
| Project GCP | `pedeve-pertamina-dms` | `pedeve-production` (berbeda) |
| Project Number | `1076379007862` | `<PROJECT_NUMBER_PRODUCTION>` (berbeda, akan diisi saat setup) |
| Service Account | `github-actions-deployer@pedeve-pertamina-dms...` | `github-actions-deployer@pedeve-production.iam.gserviceaccount.com` (berbeda) |
| WIF Provider | Development WIF | Production WIF (setup baru) |

---

## 📝 Notes

1. **⚠️ Project GCP Terpisah:** Production menggunakan akun GCP terpisah (milik client), berbeda dengan development
2. **Naming Convention:** Semua resources production menggunakan suffix `-prod`
3. **Secrets:** Semua secrets production menggunakan suffix `_prod` atau `-prod`
4. **GitHub Secrets:** Production menggunakan secrets dengan suffix `_PROD` (tidak mengubah secrets development)
5. **Branch:** Deployment production trigger dari branch `main`
6. **WIF Setup:** WIF harus di-setup baru di project production (tidak bisa reuse dari development)
7. **Service Account:** Service Account production berbeda dengan development
8. **Domain:** Domain production berbeda dengan development (akan dikonfirmasi dengan client)
9. **SSH User:** SSH user mungkin berbeda dengan development (akan dikonfirmasi dengan client)
10. **Security:** Production harus lebih secure (stronger passwords, monitoring, etc.)
11. **Backup:** Setup backup strategy untuk production database
12. **Monitoring:** Pertimbangkan setup monitoring & alerting untuk production

## 🔑 Informasi yang Perlu Dikonfirmasi dengan Client

Sebelum mulai setup, pastikan informasi berikut sudah dikonfirmasi:

- [x] **Project ID Production:** `pedeve-production` ✅
- [ ] **Project Number Production:** `<PROJECT_NUMBER_PRODUCTION>`
- [x] **Domain Frontend Production:** `reports.pertamina-pedeve.co.id` ✅
- [x] **Domain Backend Production:** `api-reports.pertamina-pedeve.co.id` ✅
- [ ] **SSH User Production:** `<ssh-user>` (mungkin berbeda dengan development)
- [ ] **Service Account Policy:** Pastikan "SA key creation disabled" (untuk enforce WIF)
- [ ] **Region/Zone:** Apakah tetap `asia-southeast2` atau berbeda?
- [ ] **Machine Types:** Apakah spesifikasi VM sama dengan development atau perlu lebih besar?
- [ ] **Database Size:** Apakah storage database perlu lebih besar dari development?

---

**Last Updated:** 2025-01-27  
**Status:** 📋 Planning Phase
