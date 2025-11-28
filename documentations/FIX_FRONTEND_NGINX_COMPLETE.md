# 🔧 Fix Frontend Nginx - Complete Fix

## Masalah

Nginx masih error karena menggunakan path certificate backend (`api-pedeve-dev`) padahal seharusnya `pedeve-dev`.

## Langkah Fix Lengkap

**Jalankan command berikut di frontend VM (satu per satu):**

### 1. Cek Semua Nginx Config Files

```bash
# Cek config yang aktif
sudo cat /etc/nginx/sites-enabled/default

# Cek apakah ada config lain
sudo ls -la /etc/nginx/sites-enabled/
sudo ls -la /etc/nginx/sites-available/

# Cek apakah ada include di nginx.conf
sudo grep -n "include" /etc/nginx/nginx.conf
```

### 2. Hapus Semua Config yang Salah

```bash
# Hapus semua enabled sites
sudo rm -f /etc/nginx/sites-enabled/*

# Hapus config backend jika ada
sudo rm -f /etc/nginx/sites-available/backend-api
```

### 3. Update Config dengan Path yang Benar

```bash
# Update default config dengan path certificate FRONTEND (pedeve-dev)
sudo tee /etc/nginx/sites-available/default > /dev/null <<'EOF'
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name pedeve-dev.aretaamany.com _;

    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name pedeve-dev.aretaamany.com;

    # PASTIKAN: pedeve-dev (FRONTEND), BUKAN api-pedeve-dev
    ssl_certificate /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    root /var/www/html;
    index index.html;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF
```

### 4. Enable Config

```bash
# Enable default config
sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default

# Verify hanya default yang enabled
sudo ls -la /etc/nginx/sites-enabled/
```

### 5. Verify Certificate Path

```bash
# Verify certificate path exists (dengan sudo)
sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

# Verify certificate readable
sudo test -r /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem && echo "✅ Certificate readable" || echo "❌ Certificate not readable"
sudo test -r /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem && echo "✅ Private key readable" || echo "❌ Private key not readable"
```

### 6. Test dan Reload

```bash
# Test config
sudo nginx -t

# Jika test berhasil, reload
sudo systemctl reload nginx

# Jika reload gagal, restart
sudo systemctl restart nginx

# Check status
sudo systemctl status nginx --no-pager -l | head -20
```

### 7. Verify Port 443

```bash
# Check port 443 listening
sudo ss -tlnp | grep 443

# Test HTTPS
curl -k https://127.0.0.1/health
```

## Jika Masih Error

**Cek error log:**
```bash
# Cek Nginx error log
sudo tail -50 /var/log/nginx/error.log

# Cek systemd log
sudo journalctl -xeu nginx.service --no-pager | tail -30
```

**Cek apakah ada config lain:**
```bash
# Cek semua file yang mention api-pedeve-dev
sudo grep -r "api-pedeve-dev" /etc/nginx/

# Cek apakah ada symlink yang salah
sudo find /etc/nginx -type l -ls
```

## Quick Fix (All-in-One)

```bash
# 1. Hapus semua enabled sites
sudo rm -f /etc/nginx/sites-enabled/*

# 2. Update config (copy config di step 3)

# 3. Enable config
sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default

# 4. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 5. Verify
sudo ss -tlnp | grep 443
curl -k https://127.0.0.1/health
```

