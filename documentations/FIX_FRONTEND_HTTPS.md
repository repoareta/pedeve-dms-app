# 🔧 Fix: Frontend HTTPS Tidak Aktif

## Masalah

Frontend `https://pedeve-dev.aretaamany.com/` masih tidak bisa diakses.

## Langkah Fix

### 1. SSH ke Frontend VM

```bash
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
```

### 2. Cek Status Services

```bash
# Check Nginx status
sudo systemctl status nginx --no-pager -l | head -15

# Check listening ports
sudo ss -tlnp | grep -E ':(80|443)'

# Check SSL certificate
sudo certbot certificates

# Check frontend files
ls -la /var/www/html/ | head -10
```

### 3. Cek Nginx Config

```bash
sudo cat /etc/nginx/sites-enabled/default | grep -E 'listen|ssl_certificate'
```

**Jika tidak ada HTTPS block (listen 443), lanjut ke step 4.**

### 4. Update Nginx Config dengan HTTPS

**Jika SSL certificate ada:**

```bash
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

### 5. Test dan Reload Nginx

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 6. Verify Port 443 Listening

```bash
sudo ss -tlnp | grep 443
```

**Expected output:**
```
LISTEN 0  128  0.0.0.0:443  0.0.0.0:*  users:(("nginx",pid=XXXX,fd=X))
```

### 7. Test HTTPS

```bash
# Test dari dalam VM
curl -k https://127.0.0.1/health

# Atau test dari local
curl -k https://pedeve-dev.aretaamany.com/
```

## Jika SSL Certificate Tidak Ada

**Generate SSL certificate:**

```bash
sudo certbot --nginx -d pedeve-dev.aretaamany.com
```

**Atau manual:**

```bash
sudo certbot certonly --standalone -d pedeve-dev.aretaamany.com
```

Setelah itu, jalankan kembali step 4-7.

## Quick Fix Script (All-in-One)

```bash
# 1. Cek SSL certificate
sudo certbot certificates

# 2. Update Nginx config dengan HTTPS (copy config di atas)

# 3. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 4. Verify
sudo ss -tlnp | grep 443
curl -k https://127.0.0.1/health
```

## Expected Result

Setelah fix:
- ✅ Port 443 listening
- ✅ HTTPS accessible: `https://pedeve-dev.aretaamany.com/`
- ✅ HTTP redirect to HTTPS
- ✅ Frontend files accessible

