# Fix Nginx HTTPS Config Sekarang

## Masalah
Nginx config hanya listen di port 80, tidak ada port 443. Certbot sudah generate certificate tapi config belum di-update.

## Solusi: Update Nginx Config Manual

Jalankan di VM backend-dev:

```bash
# Backup config dulu
sudo cp /etc/nginx/sites-enabled/backend-api /etc/nginx/sites-enabled/backend-api.backup

# Update config dengan HTTPS block
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    # SSL certificate (dari Certbot)
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
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# Test config
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx

# Verify port 443 listening
sudo ss -tlnp | grep 443
```

## Test Setelah Fix

```bash
# Test HTTPS
curl https://api-pedeve-dev.aretaamany.com/health
curl https://api-pedeve-dev.aretaamany.com/api/v1/csrf-token

# Test HTTP redirect
curl -I http://api-pedeve-dev.aretaamany.com/health
# Harus return 301 redirect ke HTTPS
```

## Expected Result

Setelah fix:
- ✅ Nginx listen di port 443
- ✅ HTTPS endpoint bisa diakses
- ✅ HTTP otomatis redirect ke HTTPS
- ✅ SSL certificate terpasang dengan benar

