# 🔧 Troubleshooting Deployment Failure

## Masalah
Setelah deployment selesai, frontend dan backend tidak bisa diakses.

## Quick Diagnostic Commands

### Step 1: Cek Backend Container Status

**SSH ke backend VM dan jalankan:**

```bash
# Cek container status
sudo docker ps -a | grep dms-backend-prod

# Cek container logs (jika running)
sudo docker logs dms-backend-prod --tail 50

# Cek apakah container crash
sudo docker ps -a

# Cek port 8080 listening
sudo ss -tlnp | grep 8080
```

### Step 2: Cek Nginx Status

**Backend VM:**
```bash
# Cek Nginx status
sudo systemctl status nginx

# Cek Nginx config
sudo nginx -t

# Cek port 80 dan 443 listening
sudo ss -tlnp | grep -E "80|443"
```

**Frontend VM:**
```bash
# Cek Nginx status
sudo systemctl status nginx

# Cek port 80 dan 443 listening
sudo ss -tlnp | grep -E "80|443"

# Cek static files
ls -la /var/www/html/
```

### Step 3: Cek Cloud SQL Proxy

**Backend VM:**
```bash
# Cek Cloud SQL Proxy running
ps aux | grep cloud-sql-proxy

# Test database connection
psql "postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable" -c "SELECT 1;"
```

### Step 4: Test dari VM

**Backend VM:**
```bash
# Test backend langsung
curl http://127.0.0.1:8080/health

# Test via Nginx
curl http://127.0.0.1/health
curl https://127.0.0.1/health
```

**Frontend VM:**
```bash
# Test frontend
curl http://127.0.0.1/
curl https://127.0.0.1/
```

## Common Issues & Fixes

### Issue 1: Container Tidak Running

**Gejala:** `sudo docker ps` tidak menampilkan container

**Fix:**
```bash
# Cek logs container yang stopped
sudo docker ps -a
sudo docker logs dms-backend-prod

# Restart container
sudo docker start dms-backend-prod

# Atau restart dengan script deploy
```

### Issue 2: Container Crash (Exit Code != 0)

**Gejala:** Container status "Exited" dengan exit code bukan 0

**Fix:**
```bash
# Cek logs untuk error
sudo docker logs dms-backend-prod --tail 100

# Common causes:
# - Database connection failed
# - Missing environment variables
# - Port already in use
```

### Issue 3: Nginx Tidak Running

**Gejala:** Port 80/443 tidak listening

**Fix:**
```bash
# Start Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Cek error
sudo nginx -t
sudo journalctl -u nginx --tail 50
```

### Issue 4: Firewall Rules Tidak Aktif

**Gejala:** Bisa akses dari VM, tapi tidak dari luar

**Fix:**
```bash
# Cek firewall rules di GCP Console
# Atau via gcloud:
gcloud compute firewall-rules list --project=pedeve-pertamina-dms

# Cek VM tags
gcloud compute instances describe backend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms \
  --format="get(tags.items)"
```

### Issue 5: Database Connection Failed

**Gejala:** Container logs menunjukkan database error

**Fix:**
```bash
# Cek Cloud SQL Proxy
ps aux | grep cloud-sql-proxy

# Restart Cloud SQL Proxy jika perlu
# (biasanya running sebagai systemd service)

# Test connection
psql "postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable" -c "SELECT 1;"
```

## Quick Recovery Script

**Jalankan di backend VM:**

```bash
#!/bin/bash
set -e

echo "🔍 Diagnosing deployment issues..."

# 1. Cek container
echo "1. Container status:"
sudo docker ps -a | grep dms-backend-prod || echo "Container not found"

# 2. Cek container logs
echo ""
echo "2. Container logs (last 30 lines):"
sudo docker logs dms-backend-prod --tail 30 2>&1 || echo "Cannot get logs"

# 3. Cek port 8080
echo ""
echo "3. Port 8080 listening:"
sudo ss -tlnp | grep 8080 || echo "Port 8080 not listening"

# 4. Cek Nginx
echo ""
echo "4. Nginx status:"
sudo systemctl status nginx --no-pager -l | head -10 || echo "Nginx not running"

# 5. Cek port 80/443
echo ""
echo "5. Port 80/443 listening:"
sudo ss -tlnp | grep -E "80|443" || echo "Ports 80/443 not listening"

# 6. Test backend
echo ""
echo "6. Testing backend:"
curl -s http://127.0.0.1:8080/health || echo "Backend not responding"

# 7. Test via Nginx
echo ""
echo "7. Testing via Nginx:"
curl -s http://127.0.0.1/health || echo "Nginx not responding"
```

