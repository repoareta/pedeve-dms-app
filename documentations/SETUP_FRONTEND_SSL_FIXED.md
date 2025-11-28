# 🔒 Setup SSL Frontend - Fixed Version

## Masalah
Error: `no "ssl_certificate" is defined for the "listen ... ssl" directive`

Ini terjadi karena config Nginx sudah include SSL block tapi certificate belum ada.

## Solusi: Setup HTTP dulu, baru generate SSL

### Step 1: Setup Nginx Config untuk HTTP (tanpa SSL block)

```bash
sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name pedeve-dev.aretaamany.com _;

    root /var/www/html;
    index index.html;

    # Security headers
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

# Test config
sudo nginx -t

# Reload jika test berhasil
sudo systemctl reload nginx
```

### Step 2: Generate SSL Certificate dengan Certbot

Certbot akan otomatis update Nginx config untuk HTTPS:

```bash
# Install Certbot jika belum ada
if ! command -v certbot &> /dev/null; then
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
fi

# Generate SSL certificate
sudo certbot --nginx \
  -d pedeve-dev.aretaamany.com \
  --email info@aretaamany.com \
  --agree-tos \
  --non-interactive \
  --redirect
```

Certbot akan:
- Generate SSL certificate
- Otomatis update Nginx config untuk HTTPS
- Setup HTTP to HTTPS redirect

### Step 3: Setup Auto-renewal

```bash
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Test renewal
sudo certbot renew --dry-run
```

### Step 4: Verifikasi

```bash
# Test HTTPS
curl https://pedeve-dev.aretaamany.com/health

# Test HTTP redirect
curl -I http://pedeve-dev.aretaamany.com/health

# Cek SSL certificate
sudo certbot certificates

# Cek port 443 listening
sudo ss -tlnp | grep 443
```

## Expected Result

Setelah setup:
- ✅ HTTP redirect ke HTTPS: `http://pedeve-dev.aretaamany.com` → `https://pedeve-dev.aretaamany.com`
- ✅ HTTPS berfungsi: `https://pedeve-dev.aretaamany.com` bisa diakses
- ✅ SSL certificate auto-renewal aktif

