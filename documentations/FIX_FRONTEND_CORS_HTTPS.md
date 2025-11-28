# 🔧 Fix Frontend CORS & HTTPS Issues

## Masalah yang Ditemukan

1. **Frontend HTTPS tidak bisa diakses**: `https://pedeve-dev.aretaamany.com/` masih belum bisa diakses
2. **CORS Error**: Frontend diakses via IP (`http://34.128.123.1`) tapi backend CORS hanya mengizinkan domain HTTPS

## Solusi

### 1. Update CORS di Backend

Backend perlu mengizinkan multiple origins:
- `https://pedeve-dev.aretaamany.com` (domain HTTPS)
- `http://34.128.123.1` (IP untuk testing)
- `http://pedeve-dev.aretaamany.com` (domain HTTP, akan redirect ke HTTPS)

**File yang diupdate:**
- `scripts/deploy-backend-vm.sh`: CORS_ORIGIN sekarang mengizinkan multiple origins

### 2. Setup SSL di Frontend VM

Frontend VM perlu setup SSL certificate untuk domain `pedeve-dev.aretaamany.com`.

**Script yang dibuat:**
- `scripts/setup-frontend-ssl.sh`: Script untuk setup SSL certificate dengan Certbot

## Langkah-langkah Fix

### Step 1: Update Backend CORS (Sudah diupdate di script)

CORS_ORIGIN sekarang mengizinkan:
```
https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com
```

**Untuk apply perubahan:**
1. Re-deploy backend (akan otomatis update CORS)
2. Atau restart container dengan CORS_ORIGIN yang baru

### Step 2: Setup SSL di Frontend VM

**Dari local machine:**

```bash
# Copy script ke VM
gcloud compute scp --zone=asia-southeast2-a \
  scripts/setup-frontend-ssl.sh \
  frontend-dev:~/setup-frontend-ssl.sh

# SSH ke VM dan jalankan
gcloud compute ssh --zone=asia-southeast2-a frontend-dev -- \
  "chmod +x ~/setup-frontend-ssl.sh && sudo ~/setup-frontend-ssl.sh"
```

**Atau dari dalam VM:**

```bash
# Download script atau copy manual
chmod +x setup-frontend-ssl.sh
sudo ./setup-frontend-ssl.sh
```

### Step 3: Setup Firewall untuk Frontend HTTPS

**Pastikan firewall rule untuk port 443 sudah dibuat:**

Via GCP Console:
1. Go to: https://console.cloud.google.com/networking/firewalls?project=pedeve-pertamina-dms
2. Cek apakah `allow-https` sudah ada
3. Jika belum, buat dengan:
   - Name: `allow-https`
   - Direction: `Ingress`
   - Targets: `Specified target tags`
   - Target tags: `https-server`
   - Source IP ranges: `0.0.0.0/0`
   - Protocols: `tcp:443`

**Pastikan tag sudah di-apply ke frontend VM:**

Via GCP Console:
1. Go to: https://console.cloud.google.com/compute/instances?project=pedeve-pertamina-dms
2. Klik VM: `frontend-dev`
3. Klik "EDIT"
4. Scroll ke "Network tags"
5. Pastikan tag `https-server` sudah ada
6. Jika belum, tambah dan "SAVE"

### Step 4: Verifikasi

**Test Frontend HTTPS:**
```bash
curl https://pedeve-dev.aretaamany.com/health
# Expected: OK

curl -I http://pedeve-dev.aretaamany.com/health
# Expected: HTTP/1.1 301 Moved Permanently (redirect to HTTPS)
```

**Test CORS dari Browser:**
1. Buka `http://34.128.123.1` di browser
2. Buka Developer Console (F12)
3. Cek Network tab
4. Request ke `https://api-pedeve-dev.aretaamany.com/api/v1/csrf-token` seharusnya berhasil tanpa CORS error

**Test Frontend via Domain HTTPS:**
1. Buka `https://pedeve-dev.aretaamany.com` di browser
2. Seharusnya frontend bisa diakses
3. API calls seharusnya berhasil tanpa CORS error

## Troubleshooting

### CORS masih error

**Cek CORS_ORIGIN di backend container:**
```bash
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "sudo docker exec dms-backend-prod env | grep CORS_ORIGIN"
```

**Expected output:**
```
CORS_ORIGIN=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com
```

**Jika tidak sesuai, restart container dengan CORS_ORIGIN yang benar:**
```bash
# Stop container
sudo docker stop dms-backend-prod

# Start dengan CORS_ORIGIN yang benar
sudo docker run -d \
  --name dms-backend-prod \
  --restart unless-stopped \
  --network host \
  -e CORS_ORIGIN="https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com" \
  # ... other env vars ...
  ghcr.io/repoareta/dms-backend:latest
```

### Frontend HTTPS masih tidak bisa diakses

**Cek Nginx config:**
```bash
gcloud compute ssh --zone=asia-southeast2-a frontend-dev -- \
  "sudo cat /etc/nginx/sites-available/default | grep -A 5 'listen 443'"
```

**Cek SSL certificate:**
```bash
gcloud compute ssh --zone=asia-southeast2-a frontend-dev -- \
  "sudo certbot certificates"
```

**Cek port 443 listening:**
```bash
gcloud compute ssh --zone=asia-southeast2-a frontend-dev -- \
  "sudo ss -tlnp | grep 443"
```

**Cek firewall rule:**
```bash
gcloud compute firewall-rules describe allow-https \
  --project=pedeve-pertamina-dms

gcloud compute instances describe frontend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms \
  --format="get(tags.items)"
```

## Expected Result

Setelah semua fix:
- ✅ Frontend bisa diakses via `https://pedeve-dev.aretaamany.com`
- ✅ Frontend bisa diakses via `http://34.128.123.1` (untuk testing)
- ✅ HTTP redirect ke HTTPS: `http://pedeve-dev.aretaamany.com` → `https://pedeve-dev.aretaamany.com`
- ✅ CORS error hilang, API calls berhasil
- ✅ Backend CORS mengizinkan semua origin yang diperlukan

