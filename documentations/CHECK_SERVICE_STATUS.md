# 🔍 Check Service Status - Quick Commands

## Backend VM - Cek Container

**Jalankan di backend VM:**

```bash
# 1. Cek container status
sudo docker ps -a | grep dms-backend-prod

# 2. Cek container logs
sudo docker logs dms-backend-prod --tail 50

# 3. Cek port 443 (SSL)
sudo ss -tlnp | grep 443

# 4. Test backend
curl http://127.0.0.1:8080/health
curl http://127.0.0.1/health
curl https://127.0.0.1/health
```

## Frontend VM - Cek Static Files & SSL

**Jalankan di frontend VM:**

```bash
# 1. Cek static files
ls -la /var/www/html/ | head -10

# 2. Cek port 443
sudo ss -tlnp | grep 443

# 3. Test frontend
curl http://127.0.0.1/
curl https://127.0.0.1/

# 4. Cek SSL certificate
sudo certbot certificates
```

## Kemungkinan Masalah

### 1. Container Backend Tidak Running
**Fix:** Start container atau restart dengan script deploy

### 2. Port 443 Tidak Listening (SSL)
**Fix:** Setup SSL certificate dengan Certbot

### 3. Frontend Static Files Tidak Ada
**Fix:** Re-deploy frontend atau copy files manual

