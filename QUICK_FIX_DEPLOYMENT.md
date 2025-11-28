# 🚨 Quick Fix - Service Tidak Bisa Diakses

## Langkah Cepat: Diagnose & Fix

### Step 1: SSH ke Backend VM dan Jalankan Diagnostic

```bash
# Copy script diagnostic ke VM (dari local machine)
gcloud compute scp --zone=asia-southeast2-a \
  scripts/diagnose-deployment.sh \
  backend-dev:~/diagnose-deployment.sh

# SSH ke VM dan jalankan
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "chmod +x ~/diagnose-deployment.sh && sudo ~/diagnose-deployment.sh"
```

**Atau jalankan manual di VM:**

```bash
# Cek container
sudo docker ps -a | grep dms-backend-prod

# Cek logs jika container stopped
sudo docker logs dms-backend-prod --tail 50

# Cek Nginx
sudo systemctl status nginx

# Cek ports
sudo ss -tlnp | grep -E "80|443|8080"
```

### Step 2: Restart Services (Jika Diperlukan)

**Jika container tidak running:**

```bash
# Start container
sudo docker start dms-backend-prod

# Atau restart dengan script deploy
# (lihat scripts/deploy-backend-vm.sh)
```

**Jika Nginx tidak running:**

```bash
# Start Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Reload config
sudo nginx -t && sudo systemctl reload nginx
```

### Step 3: Cek Frontend VM

**SSH ke frontend VM:**

```bash
# Cek Nginx
sudo systemctl status nginx

# Cek static files
ls -la /var/www/html/

# Test
curl http://127.0.0.1/
curl https://127.0.0.1/
```

## Common Fixes

### Fix 1: Container Crash

**Gejala:** Container status "Exited"

**Fix:**
```bash
# Cek logs untuk error
sudo docker logs dms-backend-prod --tail 100

# Common errors:
# - Database connection failed → Cek Cloud SQL Proxy
# - Missing env vars → Restart dengan env vars lengkap
# - Port conflict → Cek port 8080
```

### Fix 2: Nginx Tidak Running

**Fix:**
```bash
sudo systemctl start nginx
sudo systemctl enable nginx
sudo nginx -t
```

### Fix 3: Cloud SQL Proxy Tidak Running

**Fix:**
```bash
# Cek service
sudo systemctl status cloud-sql-proxy

# Start jika perlu
sudo systemctl start cloud-sql-proxy
```

