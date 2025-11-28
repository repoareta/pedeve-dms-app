# 🔧 Fix HTTPS Port 443 Not Listening

## Masalah
- Frontend: HTTP ✅, HTTPS ❌ (port 443 tidak listening)
- Backend: HTTP ✅, HTTPS ❌ (port 443 tidak listening)
- Container backend running ✅
- Static files ada ✅

## Penyebab
Nginx config tidak memiliki block untuk port 443 (HTTPS), atau SSL certificate belum terpasang dengan benar.

## Solusi: Fix Nginx Config untuk HTTPS

### Backend VM - Fix HTTPS Config

**SSH ke backend VM dan jalankan:**

```bash
# 1. Cek Nginx config saat ini
sudo cat /etc/nginx/sites-available/backend-api

# 2. Cek SSL certificate
sudo certbot certificates

# 3. Jika certificate ada tapi config tidak ada HTTPS block, update config:
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    # SSL certificate
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
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;

    # Timeout settings
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    # Forward all requests to backend on port 8080
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# 4. Test config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verifikasi port 443 listening
sudo ss -tlnp | grep 443

# 7. Test HTTPS
curl https://127.0.0.1/health
```

### Frontend VM - Fix HTTPS Config

**SSH ke frontend VM dan jalankan:**

```bash
# 1. Cek SSL certificate
sudo certbot certificates

# 2. Cek Nginx config
sudo cat /etc/nginx/sites-available/default

# 3. Jika certificate ada tapi config tidak ada HTTPS block, update config:
sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name pedeve-dev.aretaamany.com _;

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name pedeve-dev.aretaamany.com;

    # SSL certificate
    ssl_certificate /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    root /var/www/html;
    index index.html;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # SPA routing - semua request ke index.html kecuali static files
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF

# 4. Test config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verifikasi port 443 listening
sudo ss -tlnp | grep 443

# 7. Test HTTPS
curl https://127.0.0.1/
```

## Quick Fix Script

**Backend VM:**

```bash
# Cek certificate path
CERT_PATH=$(sudo certbot certificates 2>/dev/null | grep -A 2 "api-pedeve-dev" | grep "Certificate Path" | awk '{print $3}' | head -1 | xargs dirname 2>/dev/null || echo "/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com")

# Update config dengan HTTPS block (lihat script di atas)
# Lalu:
sudo nginx -t && sudo systemctl reload nginx
sudo ss -tlnp | grep 443
```

**Frontend VM:**

```bash
# Cek certificate path
CERT_PATH=$(sudo certbot certificates 2>/dev/null | grep -A 2 "pedeve-dev" | grep "Certificate Path" | awk '{print $3}' | head -1 | xargs dirname 2>/dev/null || echo "/etc/letsencrypt/live/pedeve-dev.aretaamany.com")

# Update config dengan HTTPS block (lihat script di atas)
# Lalu:
sudo nginx -t && sudo systemctl reload nginx
sudo ss -tlnp | grep 443
```

