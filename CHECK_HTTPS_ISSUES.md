# Troubleshooting HTTPS Connection Refused

## Masalah: `Connection refused` pada port 443

### Kemungkinan Penyebab

1. **Firewall rule untuk port 443 belum dibuat atau tag belum di-apply**
2. **Nginx belum listen di port 443** (meskipun certbot sudah setup)
3. **Nginx config belum benar**

## Quick Checks

### 1. Cek Nginx Config

```bash
# Cek apakah Nginx listen di port 443
sudo ss -tlnp | grep 443

# Cek Nginx config
sudo cat /etc/nginx/sites-enabled/backend-api

# Test Nginx config
sudo nginx -t

# Restart Nginx jika perlu
sudo systemctl restart nginx
```

### 2. Cek Firewall Rule

**Pastikan firewall rule untuk port 443 sudah dibuat dan tag sudah di-apply:**

Via GCP Console:
- Firewall rule `allow-https` sudah ada
- VM `backend-dev` sudah di-tag dengan `https-server`

Via gcloud (dari local machine):
```bash
# Cek firewall rule
gcloud compute firewall-rules describe allow-https --project pedeve-pertamina-dms

# Cek VM tags
gcloud compute instances describe backend-dev \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms \
  --format="get(tags.items)"
```

### 3. Test dari Dalam VM

```bash
# Test localhost HTTPS
curl -k https://127.0.0.1/health

# Test dengan IP external
curl -k https://34.101.49.147/health
```

Jika `-k` (ignore SSL) berhasil, berarti masalahnya di firewall atau DNS.

## Solusi

### Step 1: Pastikan Firewall Rule Sudah Dibuat

**Via GCP Console:**
1. Go to: https://console.cloud.google.com/networking/firewalls?project=pedeve-pertamina-dms
2. Cek apakah `allow-https` sudah ada
3. Jika belum, buat dengan:
   - Name: `allow-https`
   - Direction: `Ingress`
   - Targets: `Specified target tags`
   - Target tags: `https-server`
   - Source IP ranges: `0.0.0.0/0`
   - Protocols: `tcp:443`

### Step 2: Pastikan Tag Sudah Di-Apply

**Via GCP Console:**
1. Go to: https://console.cloud.google.com/compute/instances?project=pedeve-pertamina-dms
2. Klik VM: `backend-dev`
3. Klik "EDIT"
4. Scroll ke "Network tags"
5. Pastikan tag `https-server` sudah ada
6. Jika belum, tambah dan "SAVE"

### Step 3: Verifikasi Nginx Config

```bash
# Cek config
sudo cat /etc/nginx/sites-enabled/backend-api | grep -A 5 "listen 443"

# Harus ada:
# listen 443 ssl http2;
# listen [::]:443 ssl http2;
```

### Step 4: Restart Nginx

```bash
sudo systemctl restart nginx
sudo systemctl status nginx
```

### Step 5: Test Lagi

```bash
# Test HTTPS
curl https://api-pedeve-dev.aretaamany.com/health

# Test dengan verbose untuk debug
curl -v https://api-pedeve-dev.aretaamany.com/health
```

## Expected Result

Setelah semua fix:
- ✅ Firewall rule `allow-https` sudah ada
- ✅ VM sudah di-tag dengan `https-server`
- ✅ Nginx listen di port 443
- ✅ HTTPS endpoint bisa diakses

---

## Memperbaiki Warning "Conflicting Server Name"

### Masalah: Warning saat `nginx -t`

```
nginx: [warn] conflicting server name "api-pedeve-dev.aretaamany.com" on 0.0.0.0:80, ignored
```

**Penyebab:** Ada beberapa `server` block dengan `server_name` yang sama pada port yang sama. Ini biasanya terjadi karena:
- Certbot menambahkan config tambahan
- Ada file config lain yang juga menggunakan `server_name` yang sama
- Config default masih aktif

### Solusi: Jalankan Script Fix

**Dari local machine (akan copy script ke VM):**

```bash
# Copy script ke VM
gcloud compute scp --zone=asia-southeast2-a \
  scripts/fix-nginx-conflict.sh \
  backend-dev:~/fix-nginx-conflict.sh

# SSH ke VM dan jalankan
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "chmod +x ~/fix-nginx-conflict.sh && sudo ~/fix-nginx-conflict.sh"
```

**Atau langsung dari dalam VM:**

```bash
# Download script dari repo atau copy manual
# Lalu jalankan:
chmod +x fix-nginx-conflict.sh
sudo ./fix-nginx-conflict.sh
```

### Manual Fix (jika script tidak tersedia)

1. **Cek file config yang aktif:**
   ```bash
   sudo ls -la /etc/nginx/sites-enabled/
   ```

2. **Hapus config default jika masih ada:**
   ```bash
   sudo rm -f /etc/nginx/sites-enabled/default
   ```

3. **Pastikan hanya `backend-api` yang aktif:**
   ```bash
   sudo ls -la /etc/nginx/sites-enabled/
   # Harus hanya ada: backend-api -> /etc/nginx/sites-available/backend-api
   ```

4. **Cek isi config untuk duplikasi:**
   ```bash
   sudo cat /etc/nginx/sites-available/backend-api
   # Harus hanya ada 2 server block:
   # - 1 untuk HTTP (port 80) dengan redirect
   # - 1 untuk HTTPS (port 443) dengan proxy
   ```

5. **Test dan reload:**
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```

### Expected Result Setelah Fix

```bash
$ sudo nginx -t
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
# ✅ Tidak ada warning lagi!
```

