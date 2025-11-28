# 🔧 Fix: Port 443 Tidak Listening - Langsung Fix

## Status Saat Ini

✅ Nginx running
✅ Container running  
✅ Port 80 listening
✅ Port 8080 listening
❌ **Port 443 TIDAK listening** (HTTPS tidak aktif)
⚠️ Warning: conflicting server name

## Langkah Fix

### 1. Cek SSL Certificate

```bash
sudo certbot certificates
```

**Expected output:**
```
Found the following certs:
  Certificate Name: api-pedeve-dev.aretaamany.com
    Domains: api-pedeve-dev.aretaamany.com
    Expiry Date: 2026-XX-XX XX:XX:XX+00:00 (VALID: XX days)
    Certificate Path: /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem
    Private Key Path: /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/privkey.pem
```

### 2. Cek Nginx Config Saat Ini

```bash
sudo cat /etc/nginx/sites-enabled/backend-api
```

**Jika tidak ada HTTPS block (listen 443), lanjut ke step 3.**

### 3. Update Nginx Config dengan HTTPS

```bash
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    ssl_certificate /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    access_log /var/log/nginx/backend-api-access.log;
    error_log /var/log/nginx/backend-api-error.log;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF
```

### 4. Fix Conflicting Server Name Warning

**Cek apakah ada duplicate config:**
```bash
sudo ls -la /etc/nginx/sites-enabled/
```

**Hapus default config jika ada:**
```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

**Pastikan hanya backend-api yang enabled:**
```bash
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api
```

### 5. Test dan Reload Nginx

```bash
# Test config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Verify port 443 listening
sudo ss -tlnp | grep 443
```

**Expected output:**
```
LISTEN 0  128  0.0.0.0:443  0.0.0.0:*  users:(("nginx",pid=XXXX,fd=X))
```

### 6. Test HTTPS

```bash
# Test dari dalam VM
curl -k https://127.0.0.1/health

# Atau test dari local
curl -k https://api-pedeve-dev.aretaamany.com/health
```

## Quick Fix Script (All-in-One)

```bash
# 1. Cek SSL certificate
sudo certbot certificates

# 2. Update Nginx config dengan HTTPS (copy config di atas)

# 3. Fix conflicting server name
sudo rm -f /etc/nginx/sites-enabled/default
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# 4. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 5. Verify
sudo ss -tlnp | grep 443
curl -k https://127.0.0.1/health
```

## Jika SSL Certificate Tidak Ada

**Generate SSL certificate:**
```bash
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com
```

**Atau manual:**
```bash
sudo certbot certonly --standalone -d api-pedeve-dev.aretaamany.com
```

## Expected Result

Setelah fix:
- ✅ Port 443 listening
- ✅ HTTPS accessible: `https://api-pedeve-dev.aretaamany.com/health`
- ✅ No conflicting server name warning
- ✅ HTTP redirect to HTTPS

