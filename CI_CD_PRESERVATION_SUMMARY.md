# 📋 CI/CD Preservation Summary

Dokumentasi lengkap tentang apa yang di-preserve oleh CI/CD deployment dan apa yang tidak.

## ✅ Yang DI-PRESERVE (Tidak Akan Di-Reset)

### 1. SSL Certificates
- **Frontend**: `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- **Backend**: `/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/`
- **Action**: Tidak pernah dihapus atau di-overwrite
- **Script**: `setup-nginx-frontend.sh` dan `setup-backend-nginx.sh` check certificate existence sebelum update

### 2. Nginx Config dengan SSL
- **Frontend**: `/etc/nginx/sites-available/default`
- **Backend**: `/etc/nginx/sites-available/backend-api`
- **Action**: Skip update jika:
  - SSL certificate exists
  - Config sudah punya HTTPS block (port 443)
  - Config sudah punya `server_name` yang benar
  - Config sudah punya `ssl_certificate` dan `ssl_certificate_key` yang benar
- **Script**: Idempotent checks di `setup-nginx-frontend.sh` dan `setup-backend-nginx.sh`

### 3. Docker Container Network Mode
- **Backend**: `--network host`
- **Action**: Tidak pernah diubah
- **Reason**: Required untuk akses Cloud SQL Proxy di `127.0.0.1:5432`
- **Script**: `deploy-backend-vm.sh` selalu menggunakan `--network host`

### 4. Firewall Rules
- **Action**: Tidak pernah di-reset
- **Note**: Firewall rules dikelola manual di GCP Console

### 5. Domain DNS Settings
- **Action**: Tidak pernah diubah
- **Note**: DNS dikelola manual

### 6. Cloud SQL Proxy Configuration
- **Action**: Tidak pernah di-reset
- **Note**: Proxy dikonfigurasi manual di VM

### 7. GCP Secret Manager Secrets
- **Action**: Tidak pernah diubah oleh deployment
- **Note**: Secrets dikelola manual di GCP Console

## ⚠️ Yang BOLEH Di-Update (Expected Behavior)

### 1. Docker Container Image
- **Backend**: `ghcr.io/repoareta/dms-backend:latest`
- **Action**: Selalu di-update dengan image baru setiap deployment
- **Reason**: Ini adalah expected behavior - kita ingin deploy versi terbaru
- **Script**: `deploy-backend-vm.sh` selalu stop old container dan start new one

### 2. Frontend Static Files
- **Location**: `/var/www/html/`
- **Action**: Selalu di-update dengan build terbaru setiap deployment
- **Reason**: Ini adalah expected behavior - kita ingin deploy versi terbaru
- **Command**: `sudo rm -rf /var/www/html/*` (hanya static files, bukan config)

### 3. Nginx Config Tanpa SSL
- **Action**: Boleh di-update jika belum ada SSL certificate
- **Reason**: Perlu setup config untuk pertama kali atau jika SSL belum ada
- **Script**: `setup-nginx-frontend.sh` dan `setup-backend-nginx.sh` akan update jika perlu

### 4. Environment Variables di Container
- **Action**: Boleh di-update untuk container baru
- **Reason**: Perlu update env vars untuk container baru
- **Script**: `deploy-backend-vm.sh` selalu pass env vars ke container baru

## 🚫 Yang TIDAK BOLEH Di-Hapus

### 1. SSL Certificates
- **Location**: `/etc/letsencrypt/live/*/`
- **Protection**: Scripts check existence sebelum update config
- **CI/CD**: Tidak ada command yang menghapus certificates

### 2. Nginx Config Files
- **Location**: `/etc/nginx/sites-available/*`
- **Protection**: Scripts check config correctness sebelum update
- **CI/CD**: Tidak ada command yang menghapus config files

### 3. Docker Network Mode
- **Protection**: Hardcoded di `deploy-backend-vm.sh` sebagai `--network host`
- **CI/CD**: Tidak pernah diubah

## 📝 CI/CD Workflow Behavior

### Backend Deployment:
1. ✅ Copy scripts ke VM
2. ✅ Run `deploy-backend-vm.sh`:
   - Stop old container
   - Load new image
   - Start new container dengan `--network host` (preserved)
   - Pass environment variables
3. ✅ Run `setup-backend-nginx.sh`:
   - Check SSL certificate existence
   - Check Nginx config correctness
   - **Skip update jika SSL dan config sudah benar**
   - Update hanya jika perlu
4. ✅ Run `ensure-services-running.sh`:
   - Check container status
   - Check Nginx status
   - Start jika tidak running
5. ✅ Cleanup: Hanya hapus temporary files (scripts, tar files)
   - **TIDAK menghapus config atau certificates**

### Frontend Deployment:
1. ✅ Copy static files ke VM
2. ✅ Extract static files ke `/var/www/html/`:
   - `sudo rm -rf /var/www/html/*` (hanya static files, bukan config)
   - Copy new files
3. ✅ Run `setup-nginx-frontend.sh`:
   - Check SSL certificate existence
   - Check Nginx config correctness
   - **Skip update jika SSL dan config sudah benar**
   - Update hanya jika perlu
4. ✅ Run `ensure-services-running.sh`:
   - Check Nginx status
   - Start jika tidak running
5. ✅ Cleanup: Hanya hapus temporary files (scripts, tar files, /tmp/dist)
   - **TIDAK menghapus config atau certificates**

## 🔍 Verification

### Check Preservation Logic:
```bash
# Frontend
grep -A 5 "SSL_CERT_EXISTS" scripts/setup-nginx-frontend.sh
grep -A 10 "preserving existing" scripts/setup-nginx-frontend.sh

# Backend
grep -A 5 "SSL_CERT_EXISTS" scripts/setup-backend-nginx.sh
grep -A 10 "preserving existing" scripts/setup-backend-nginx.sh
grep "network host" scripts/deploy-backend-vm.sh
```

### Check CI/CD Comments:
```bash
grep -A 5 "IMPORTANT" .github/workflows/ci-cd.yml
grep -A 5 "preserves" .github/workflows/ci-cd.yml
```

## 📋 Checklist Preservation

- [x] SSL certificates tidak pernah dihapus
- [x] Nginx config dengan SSL di-preserve jika sudah benar
- [x] Docker network mode (host) tidak pernah diubah
- [x] Firewall rules tidak pernah di-reset
- [x] Domain DNS tidak pernah diubah
- [x] Cloud SQL Proxy config tidak pernah di-reset
- [x] GCP Secret Manager secrets tidak pernah diubah
- [x] Scripts idempotent dengan checks yang strict
- [x] CI/CD workflow punya comments yang jelas
- [x] Cleanup hanya temporary files, bukan configs

## 🎯 Summary

**Deployment scripts sekarang:**
- ✅ **Preserve** semua konfigurasi manual yang sudah benar
- ✅ **Skip update** jika config sudah correct dengan SSL
- ✅ **Never delete** SSL certificates atau config files
- ✅ **Only update** jika perlu (misalnya SSL ada tapi config belum benar)
- ✅ **Cleanup** hanya temporary files, bukan configs

**CI/CD workflow:**
- ✅ Punya comments yang jelas tentang preservation
- ✅ Tidak ada command yang menghapus config penting
- ✅ Hanya cleanup temporary files
- ✅ Scripts yang dipanggil sudah idempotent

