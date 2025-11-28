# 🔒 Preservation of Manual Configurations

Dokumentasi ini menjelaskan bagaimana deployment scripts mempertahankan konfigurasi manual yang sudah diperbaiki.

## Masalah yang Sudah Diperbaiki

### Frontend Issues:
1. **SSL Certificate** - Certificate sudah ada tapi tidak terpasang
2. **Nginx Config** - Config tidak punya `server_name pedeve-dev.aretaamany.com` dan HTTPS block

### Backend Issues:
1. **Docker Container** - Container tidak running setelah deployment
2. **Nginx Config** - Config sudah di-setup manual dengan SSL (jika ada)

## Perbaikan yang Sudah Diterapkan

### 1. Idempotent Nginx Setup Scripts

**Frontend (`scripts/setup-nginx-frontend.sh`):**
- ✅ Check jika SSL certificate sudah ada
- ✅ Check jika Nginx config sudah punya HTTPS block dengan:
  - `ssl_certificate` untuk domain `pedeve-dev.aretaamany.com`
  - `listen 443` (HTTPS port)
  - `server_name pedeve-dev.aretaamany.com`
- ✅ **Skip update** jika config sudah benar
- ✅ Hanya reload Nginx jika config valid

**Backend (`scripts/setup-backend-nginx.sh`):**
- ✅ Check jika SSL certificate sudah ada
- ✅ Check jika Nginx config sudah punya HTTPS block dengan:
  - `ssl_certificate` untuk domain `api-pedeve-dev.aretaamany.com`
  - `listen 443` (HTTPS port)
  - `server_name api-pedeve-dev.aretaamany.com`
- ✅ **Skip update** jika config sudah benar
- ✅ Hanya reload Nginx jika config valid

### 2. Service Restart Scripts

**`scripts/ensure-services-running.sh`:**
- ✅ Check dan start services jika tidak running
- ✅ Auto-restart container dan Nginx jika mati
- ✅ Verify status setelah restart

**`scripts/restart-services-on-vm.sh`:**
- ✅ Script untuk restart services langsung di VM
- ✅ Support backend dan frontend
- ✅ Include check status dan logs

### 3. Deployment Workflow

**`.github/workflows/ci-cd.yml`:**
- ✅ Run `ensure-services-running.sh` setelah deployment
- ✅ Verify services status sebelum selesai
- ✅ Tidak ada command yang mereset config manual

## Behavior Setelah Perbaikan

### Frontend Deployment:
1. ✅ Extract static files ke `/var/www/html`
2. ✅ Run `setup-nginx-frontend.sh`:
   - **Jika SSL sudah ada dan config benar** → Skip update, hanya reload
   - **Jika SSL tidak ada** → Create HTTP-only config
   - **Jika SSL ada tapi config belum benar** → Update config dengan HTTPS
3. ✅ Run `ensure-services-running.sh` → Pastikan Nginx running
4. ✅ Verify deployment

### Backend Deployment:
1. ✅ Load Docker image
2. ✅ Stop old container
3. ✅ Start new container dengan environment variables
4. ✅ Run `setup-backend-nginx.sh`:
   - **Jika SSL sudah ada dan config benar** → Skip update, hanya reload
   - **Jika SSL tidak ada** → Create HTTP-only config
   - **Jika SSL ada tapi config belum benar** → Update config dengan HTTPS
5. ✅ Run `ensure-services-running.sh` → Pastikan container dan Nginx running
6. ✅ Verify deployment

## Manual Configurations yang Dipertahankan

### ✅ Dipertahankan (Tidak Di-Overwrite):
- SSL certificates di `/etc/letsencrypt/live/`
- Nginx config yang sudah punya HTTPS block dengan SSL certificate yang benar
- Firewall rules
- Domain DNS settings
- Cloud SQL Proxy configuration
- GCP Secret Manager secrets

### ⚠️ Akan Di-Update (Jika Perlu):
- Nginx config yang belum punya HTTPS block (jika SSL certificate sudah ada)
- Docker container (selalu restart dengan image baru)
- Frontend static files (selalu di-update dengan build terbaru)

## Troubleshooting

### Jika Services Masih Mati Setelah Deployment:

1. **Check status:**
   ```bash
   # Frontend
   sudo systemctl status nginx
   sudo ss -tlnp | grep -E ':(80|443)'
   
   # Backend
   sudo docker ps | grep dms-backend-prod
   sudo systemctl status nginx
   sudo ss -tlnp | grep -E ':(80|443|8080)'
   ```

2. **Restart services:**
   ```bash
   # Frontend
   sudo systemctl restart nginx
   
   # Backend
   sudo docker restart dms-backend-prod
   sudo systemctl restart nginx
   ```

3. **Atau gunakan script:**
   ```bash
   # Di VM
   ~/restart-services-on-vm.sh frontend
   ~/restart-services-on-vm.sh backend
   ```

### Jika SSL Config Ter-Overwrite:

1. **Check backup:**
   ```bash
   sudo ls -la /etc/nginx/sites-available/*.backup*
   ```

2. **Restore backup:**
   ```bash
   sudo cp /etc/nginx/sites-available/default.backup.YYYYMMDD_HHMMSS /etc/nginx/sites-available/default
   sudo nginx -t
   sudo systemctl reload nginx
   ```

3. **Re-run SSL setup:**
   ```bash
   # Frontend
   sudo certbot install --cert-name pedeve-dev.aretaamany.com
   
   # Backend
   sudo certbot install --cert-name api-pedeve-dev.aretaamany.com
   ```

## Best Practices

1. **Jangan manual edit config** jika tidak perlu - gunakan scripts
2. **Backup config** sebelum manual edit
3. **Test config** dengan `sudo nginx -t` sebelum reload
4. **Monitor deployment logs** untuk melihat apakah config di-skip atau di-update
5. **Verify services** setelah deployment dengan health checks

## Summary

✅ **SSL certificates** - Tidak pernah di-overwrite  
✅ **Nginx config dengan SSL** - Dipertahankan jika sudah benar  
✅ **Manual fixes** - Dipertahankan oleh idempotent checks  
⚠️ **Docker container** - Selalu restart dengan image baru (expected)  
⚠️ **Frontend files** - Selalu di-update dengan build terbaru (expected)

