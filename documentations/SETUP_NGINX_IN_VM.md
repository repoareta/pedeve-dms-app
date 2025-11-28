# Setup Nginx di Backend VM (Anda Sudah di Dalam VM)

## Anda sudah di dalam VM backend-dev!

Tidak perlu SSH lagi. Langsung jalankan command berikut di VM:

## Setup Nginx Reverse Proxy

```bash
# 1. Install Nginx
sudo apt-get update
sudo apt-get install -y nginx

# 2. Create Nginx config untuk backend API
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    # Logging
    access_log /var/log/nginx/backend-api-access.log;
    error_log /var/log/nginx/backend-api-error.log;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

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

# 3. Enable site
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# 4. Remove default site (optional)
sudo rm -f /etc/nginx/sites-enabled/default

# 5. Test Nginx config
sudo nginx -t

# 6. Reload Nginx
sudo systemctl reload nginx || sudo systemctl restart nginx

# 7. Check status
sudo systemctl status nginx
```

## Test Setelah Setup

```bash
# Test via domain (tanpa port)
curl http://api-pedeve-dev.aretaamany.com/health
curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token
```

**Harus return JSON response!**

## Troubleshooting

**Jika Nginx error:**
```bash
# Cek config syntax
sudo nginx -t

# Cek logs
sudo tail -f /var/log/nginx/backend-api-error.log

# Restart Nginx
sudo systemctl restart nginx
```

**Jika masih tidak bisa:**
- Pastikan backend container running: `sudo docker ps | grep dms-backend-prod`
- Test backend langsung: `curl http://127.0.0.1:8080/health`
- Cek apakah port 80 listening: `sudo ss -tlnp | grep 80`

