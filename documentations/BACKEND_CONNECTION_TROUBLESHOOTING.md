# Backend Connection Troubleshooting

## Error: `ERR_CONNECTION_REFUSED` saat akses API

### Kemungkinan Masalah

1. **Domain belum pointing ke Backend VM IP**
2. **Backend container tidak running atau crash**
3. **Firewall GCP memblokir akses**
4. **HTTPS vs HTTP issue** (domain pakai HTTPS tapi backend hanya HTTP)
5. **CORS configuration belum update**

## Quick Checks

### 1. Cek Backend Container Status

```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Cek container status
sudo docker ps | grep dms-backend-prod

# Cek logs jika container crash
sudo docker logs --tail 50 dms-backend-prod
```

### 2. Test Backend dari VM

```bash
# Test health endpoint dari dalam VM
curl http://127.0.0.1:8080/health

# Test dari external IP
curl http://34.101.49.147:8080/health
```

### 3. Cek Domain DNS

```bash
# Cek apakah domain sudah pointing ke IP yang benar
dig api-pedeve-dev.aretaamany.com
nslookup api-pedeve-dev.aretaamany.com

# Harus resolve ke: 34.101.49.147
```

### 4. Test API Endpoint

```bash
# Test via IP (HTTP)
curl http://34.101.49.147:8080/api/v1/csrf-token

# Test via domain (HTTP - jika belum ada SSL)
curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token

# Test via domain (HTTPS - jika sudah ada SSL)
curl https://api-pedeve-dev.aretaamany.com/api/v1/csrf-token
```

## Solutions

### Solution 1: Fix Domain DNS

**Jika domain belum pointing:**

1. Buka DNS provider (dimana `aretaamany.com` di-manage)
2. Tambah A record:
   - **Name:** `api-pedeve-dev`
   - **Type:** `A`
   - **Value:** `34.101.49.147`
   - **TTL:** `300` (5 menit)

3. Tunggu DNS propagation (bisa beberapa menit)

### Solution 2: Fix CORS Configuration

**Backend sekarang sudah support CORS_ORIGIN environment variable.**

**Cek apakah CORS_ORIGIN sudah di-set:**
```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a

# Cek environment variable di container
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN
```

**Jika belum di-set, update deployment script sudah include:**
- `CORS_ORIGIN=https://pedeve-dev.aretaamany.com` (dari deployment script)

**Setelah deployment berikutnya, CORS akan otomatis benar.**

### Solution 3: Fix HTTPS Issue

**Masalah:** Frontend akses via `https://api-pedeve-dev.aretaamany.com` tapi backend hanya expose HTTP.

**Solusi sementara:**
- Gunakan HTTP dulu: `http://api-pedeve-dev.aretaamany.com`
- Atau setup SSL certificate di backend

**Update frontend API URL (temporary):**
- Rebuild frontend dengan `VITE_API_URL=http://api-pedeve-dev.aretaamany.com/api/v1` (HTTP, bukan HTTPS)

**Solusi permanen:**
- Setup SSL certificate untuk backend (Let's Encrypt atau GCP Load Balancer)
- Atau gunakan GCP Load Balancer dengan managed SSL

### Solution 4: Fix Firewall Rules

**Cek firewall rules:**
```bash
# List firewall rules
gcloud compute firewall-rules list --project=pedeve-pertamina-dms

# Cek apakah ada rule untuk allow HTTP/HTTPS
gcloud compute firewall-rules describe allow-http --project=pedeve-pertamina-dms
gcloud compute firewall-rules describe allow-https --project=pedeve-pertamina-dms
```

**Buat firewall rule jika belum ada:**
```bash
# Allow HTTP
gcloud compute firewall-rules create allow-http \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server \
  --description "Allow HTTP traffic" \
  --project pedeve-pertamina-dms

# Allow HTTPS
gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic" \
  --project pedeve-pertamina-dms

# Apply tags ke backend VM
gcloud compute instances add-tags backend-dev \
  --tags http-server,https-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

### Solution 5: Restart Backend Container

**Jika container crash atau tidak running:**
```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a

# Restart container
sudo docker restart dms-backend-prod

# Cek logs
sudo docker logs --tail 100 dms-backend-prod
```

## Temporary Workaround

**Sementara, gunakan IP langsung untuk test:**

1. **Update frontend API URL (temporary):**
   - Rebuild dengan `VITE_API_URL=http://34.101.49.147:8080/api/v1`
   - Atau gunakan browser console untuk override:
     ```javascript
     // Di browser console
     window.VITE_API_URL = 'http://34.101.49.147:8080/api/v1'
     location.reload()
     ```

2. **Test via IP:**
   ```
   http://34.101.49.147:8080/api/v1/csrf-token
   ```

## Checklist

- [ ] Backend container running (`sudo docker ps`)
- [ ] Backend health check berhasil (`curl http://127.0.0.1:8080/health`)
- [ ] Domain DNS pointing ke `34.101.49.147`
- [ ] Firewall rules allow HTTP/HTTPS
- [ ] CORS_ORIGIN environment variable sudah di-set
- [ ] Backend logs tidak ada error
- [ ] Test API endpoint berhasil

## After Fixes

Setelah semua fix:
1. **Rebuild frontend** dengan API URL yang benar
2. **Test dari browser** dengan domain yang benar
3. **Check browser console** untuk memastikan tidak ada CORS error

