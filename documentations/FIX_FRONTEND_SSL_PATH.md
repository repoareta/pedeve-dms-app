# 🔧 Fix: Frontend SSL Certificate Path Error

## Masalah

Frontend error:
```
nginx: [emerg] cannot load certificate "/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem": 
BIO_new_file() failed (SSL: error:80000002:system library::No such file or directory)
```

**Penyebab:** Frontend config menggunakan path certificate backend (`api-pedeve-dev`) padahal seharusnya `pedeve-dev`.

## Fix

**SSH ke frontend VM dan jalankan:**

```bash
# 1. Cek certificate yang benar
sudo certbot certificates

# 2. Update Nginx config dengan path certificate yang benar (pedeve-dev, BUKAN api-pedeve-dev)
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

    # PASTIKAN PATH INI BENAR: pedeve-dev, BUKAN api-pedeve-dev
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

# 3. Verify certificate path exists
ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

# 4. Test config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verify port 443 listening
sudo ss -tlnp | grep 443

# 7. Test HTTPS
curl -k https://127.0.0.1/health
```

## Verifikasi

**Cek certificate path:**
```bash
# Frontend certificate (harus ada)
ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

# Backend certificate (tidak ada di frontend VM)
ls -la /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/  # Akan error, ini normal
```

## Catatan

- **Frontend certificate:** `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- **Backend certificate:** `/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/`
- **Jangan tertukar!** Frontend harus pakai `pedeve-dev`, backend pakai `api-pedeve-dev`

