# 🔧 Fix HTTPS Sekarang - Langsung di VM

## Masalah

- SSL certificate ada ✅
- Tapi port 443 tidak listening ❌
- Nginx config tidak punya HTTPS block

## Fix Backend VM

**SSH ke backend VM dan jalankan:**

```bash
# 1. Update Nginx config dengan HTTPS
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

# 2. Fix conflicting server name (hapus default jika ada)
sudo rm -f /etc/nginx/sites-enabled/default

# 3. Pastikan backend-api enabled
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# 4. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 5. Verify port 443 listening
sudo ss -tlnp | grep 443

# 6. Test HTTPS
curl -k https://127.0.0.1/health
```

## Fix Frontend VM

**SSH ke frontend VM dan jalankan:**

```bash
# 1. Update Nginx config dengan HTTPS
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

# 2. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 3. Verify port 443 listening
sudo ss -tlnp | grep 443

# 4. Test HTTPS
curl -k https://127.0.0.1/health
```

## Expected Result

Setelah fix:
- ✅ Port 443 listening di kedua VM
- ✅ HTTPS accessible: `https://api-pedeve-dev.aretaamany.com/health`
- ✅ HTTPS accessible: `https://pedeve-dev.aretaamany.com/`
- ✅ HTTP redirect to HTTPS

