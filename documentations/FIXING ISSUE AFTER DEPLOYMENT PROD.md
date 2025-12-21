# 📝 Dokumentasi Perbaikan Manual BE dan FE (Production)

Dokumentasi lengkap tentang semua perbaikan manual yang dilakukan untuk backend dan frontend production after deployment.

## 🔧 Frontend Manual Fixes (Production)

### Issue 1: SSL Certificate Tidak Terpasang

**Problem:**
- Certificate belum ada di `/etc/letsencrypt/live/dms.pertamina-pedeve.co.id/`
- Nginx config tidak punya HTTPS block
- Port 443 tidak listening

**Solution:**
```bash
# SSH ke frontend production VM
gcloud compute ssh frontend-prod-2 \
  --zone=asia-southeast2-a \
  --project=pedeve-production

# 1. Pastikan Nginx sudah running dan domain pointing ke IP VM ini
# Cek DNS: nslookup dms.pertamina-pedeve.co.id
# Harus resolve ke IP VM frontend-prod-2

# 2. Buat config HTTP dulu (diperlukan untuk certbot validation)
sudo tee /etc/nginx/sites-available/default > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name dms.pertamina-pedeve.co.id _;

    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Test dan reload Nginx
sudo nginx -t
sudo systemctl reload nginx

# 3. Install Certbot (jika belum ada)
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# 4. Buat SSL certificate menggunakan Certbot
# PENTING: Pastikan domain sudah pointing ke IP VM ini sebelum run certbot
# Certbot akan otomatis update Nginx config dengan SSL
sudo certbot --nginx -d dms.pertamina-pedeve.co.id --non-interactive --agree-tos --email info@aretaamany.com

# 5. Check certificate location (setelah certbot berhasil)
sudo ls -la /etc/letsencrypt/live/dms.pertamina-pedeve.co.id/

# 6. Verify Nginx config (certbot biasanya sudah otomatis update)
sudo nginx -t

# Jika certbot belum otomatis update config, atau perlu manual update, gunakan config berikut:
sudo tee /etc/nginx/sites-available/default > /dev/null <<'EOF'
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name dms.pertamina-pedeve.co.id _;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name dms.pertamina-pedeve.co.id;

    ssl_certificate /etc/letsencrypt/live/dms.pertamina-pedeve.co.id/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dms.pertamina-pedeve.co.id/privkey.pem;

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

# 7. Reload Nginx (jika ada perubahan config)
sudo systemctl reload nginx

# 8. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost
curl -I https://dms.pertamina-pedeve.co.id
```


## 🔧 Backend Manual Fixes (Production)

### Issue 1: SSL Certificate Tidak Ada
**Problem:**
- Backend tidak punya SSL certificate
- Port 443 tidak listening
- Frontend tidak bisa akses via HTTPS

**Solution:**
```bash
# SSH ke backend production VM
gcloud compute ssh backend-prod-1 \
  --zone=asia-southeast2-a \
  --project=pedeve-production

# 1. Install Certbot (jika belum ada)
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# 2. Run Certbot untuk mendapatkan SSL certificate
sudo certbot --nginx -d api-reports.pertamina-pedeve.co.id --non-interactive --agree-tos --email info@aretaamany.com

# 3. Verify certificate
sudo ls -la /etc/letsencrypt/live/api-reports.pertamina-pedeve.co.id/

# 4. Test Nginx config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost/health
curl -I https://api-reports.pertamina-pedeve.co.id/health
```
