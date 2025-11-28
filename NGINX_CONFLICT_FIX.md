# 🔧 Fix Nginx "Conflicting Server Name" Warning

## Masalah

Saat menjalankan `sudo nginx -t`, muncul warning:

```
nginx: [warn] conflicting server name "api-pedeve-dev.aretaamany.com" on 0.0.0.0:80, ignored
nginx: [warn] conflicting server name "api-pedeve-dev.aretaamany.com" on [::]:80, ignored
```

**Catatan:** Warning ini **tidak berbahaya** dan tidak mempengaruhi fungsi Nginx. HTTPS tetap bekerja dengan baik. Namun, jika ingin menghilangkan warning, ikuti langkah di bawah.

## Penyebab

Warning ini muncul karena ada beberapa `server` block dengan `server_name` yang sama pada port yang sama. Kemungkinan penyebab:
- Certbot menambahkan config tambahan
- Ada file config lain yang juga menggunakan `server_name` yang sama
- Config default masih aktif

## Solusi Cepat: Gunakan Script

### Dari Local Machine

```bash
# Copy script ke VM
gcloud compute scp --zone=asia-southeast2-a \
  scripts/fix-nginx-conflict.sh \
  backend-dev:~/fix-nginx-conflict.sh

# SSH ke VM dan jalankan
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "chmod +x ~/fix-nginx-conflict.sh && sudo ~/fix-nginx-conflict.sh"
```

### Dari Dalam VM

```bash
# Jika script sudah ada di VM
chmod +x fix-nginx-conflict.sh
sudo ./fix-nginx-conflict.sh
```

## Solusi Manual

### Step 1: Cek File Config yang Aktif

```bash
sudo ls -la /etc/nginx/sites-enabled/
```

**Expected:** Hanya ada `backend-api` yang aktif.

### Step 2: Hapus Config Default (jika ada)

```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

### Step 3: Cek Isi Config untuk Duplikasi

```bash
sudo cat /etc/nginx/sites-available/backend-api
```

**Expected:** Hanya ada 2 `server` block:
- 1 untuk HTTP (port 80) dengan redirect ke HTTPS
- 1 untuk HTTPS (port 443) dengan proxy ke backend

### Step 4: Cek File Lain yang Mungkin Konflik

```bash
# Cek semua file di sites-available
sudo grep -r "server_name.*api-pedeve-dev.aretaamany.com" /etc/nginx/sites-available/

# Cek semua file di sites-enabled
sudo grep -r "server_name.*api-pedeve-dev.aretaamany.com" /etc/nginx/sites-enabled/
```

Jika ada file lain yang menggunakan `server_name` yang sama, hapus atau nonaktifkan.

### Step 5: Pastikan Config Backend-API Benar

Config `/etc/nginx/sites-available/backend-api` harus seperti ini:

```nginx
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    # SSL certificate (set by Certbot)
    ssl_certificate /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/backend-api-access.log;
    error_log /var/log/nginx/backend-api-error.log;

    # Proxy settings
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Timeout settings
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    # Forward all requests to backend on port 8080
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Step 6: Test dan Reload

```bash
# Test config
sudo nginx -t

# Jika berhasil (tanpa warning), reload
sudo systemctl reload nginx
```

## Verifikasi

Setelah fix, test lagi:

```bash
# Test config (harus tanpa warning)
sudo nginx -t

# Expected output:
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful
# ✅ Tidak ada warning!

# Test HTTP redirect
curl -I http://api-pedeve-dev.aretaamany.com/health
# Expected: HTTP/1.1 301 Moved Permanently

# Test HTTPS
curl https://api-pedeve-dev.aretaamany.com/health
# Expected: {"service":"pedeve-backend","status":"OK"}
```

## Catatan

- Warning ini **tidak mempengaruhi fungsi** Nginx. HTTPS tetap bekerja dengan baik.
- Jika warning masih muncul setelah fix, kemungkinan ada file config lain yang perlu dicek.
- Pastikan hanya ada **satu** `server` block untuk setiap port (80 dan 443) dengan `server_name` yang sama.

