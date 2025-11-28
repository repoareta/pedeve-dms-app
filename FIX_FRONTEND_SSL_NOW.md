# 🔧 Fix Frontend SSL - Step by Step

Certificate sudah ada di `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`, tapi Nginx config belum benar.

## Jalankan di Frontend VM:

```bash
# 1. Backup config
sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup3

# 2. Update Nginx config dengan SSL
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
        try_files $uri $uri/ /index.html;
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

# 3. Test Nginx config
sudo nginx -t

# 4. Reload Nginx
sudo systemctl reload nginx

# 5. Check port 443
sudo ss -tlnp | grep 443

# 6. Test HTTPS
curl -I https://localhost

# 7. Test external
curl -I https://pedeve-dev.aretaamany.com
```

## Verify:

```bash
# Check Nginx status
sudo systemctl status nginx

# Check ports
sudo ss -tlnp | grep -E ':(80|443)'

# Test HTTP (should redirect to HTTPS)
curl -I http://pedeve-dev.aretaamany.com

# Test HTTPS
curl -I https://pedeve-dev.aretaamany.com
```

