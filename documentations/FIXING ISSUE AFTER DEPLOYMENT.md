# 📝 Dokumentasi Perbaikan Manual BE dan FE

Dokumentasi lengkap tentang semua perbaikan manual yang dilakukan untuk backend dan frontend after deployment.

## 🔧 Frontend Manual Fixes

### Issue 1: SSL Certificate Tidak Terpasang

**Problem:**
- Certificate ada di `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- Tapi Nginx config tidak punya HTTPS block
- Port 443 tidak listening

**Solution:**
```bash
# 1. Check certificate location
sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

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

# 3. Test config
sudo nginx -t

# 4. Reload Nginx
sudo systemctl reload nginx

# 5. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost
```


## 🔧 Backend Manual Fixes

### Issue 1: SSL Certificate Tidak Ada
**Problem:**
- Backend tidak punya SSL certificate
- Port 443 tidak listening
- Frontend tidak bisa akses via HTTPS

**Solution:**
```bash
# 1. Install Certbot (jika belum ada)
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# 2. Run Certbot untuk mendapatkan SSL certificate
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com --non-interactive --agree-tos --email info@aretaamany.com

# 3. Verify certificate
sudo ls -la /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/

# 4. Test Nginx config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost/health
curl -I https://api-pedeve-dev.aretaamany.com/health
```
