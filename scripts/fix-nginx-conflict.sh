#!/bin/bash
set -euo pipefail

# Script untuk memperbaiki conflicting server name di Nginx
# Usage: ./fix-nginx-conflict.sh

DOMAIN="api-pedeve-dev.aretaamany.com"

echo "🔧 Fixing Nginx conflicting server name..."

# Cek file config yang aktif
echo "📋 Checking active Nginx configs..."
sudo ls -la /etc/nginx/sites-enabled/

# Backup config saat ini
echo "💾 Backing up current config..."
sudo cp /etc/nginx/sites-available/backend-api /etc/nginx/sites-available/backend-api.backup.$(date +%Y%m%d_%H%M%S)

# Cek apakah ada file default yang masih aktif
if [ -f /etc/nginx/sites-enabled/default ]; then
  echo "⚠️  Found default config, removing..."
  sudo rm -f /etc/nginx/sites-enabled/default
fi

# Cek apakah ada file lain yang menggunakan server_name yang sama
echo "🔍 Checking for duplicate server_name..."
sudo grep -r "server_name.*${DOMAIN}" /etc/nginx/sites-available/ || true
sudo grep -r "server_name.*${DOMAIN}" /etc/nginx/sites-enabled/ || true

# Buat config yang bersih tanpa duplikasi
echo "📝 Creating clean Nginx config..."

sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};

    # SSL certificate (set by Certbot)
    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

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

# Pastikan hanya backend-api yang aktif
echo "🔗 Ensuring only backend-api is enabled..."
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# Hapus semua symlink lain yang mungkin konflik
for file in /etc/nginx/sites-enabled/*; do
  if [ -L "$file" ] && [ "$(basename "$file")" != "backend-api" ]; then
    echo "⚠️  Removing conflicting config: $(basename "$file")"
    sudo rm -f "$file"
  fi
done

# Test Nginx config
echo "🧪 Testing Nginx configuration..."
if sudo nginx -t; then
  echo "✅ Nginx config is valid!"
else
  echo "❌ Nginx config has errors!"
  exit 1
fi

# Reload Nginx
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx

echo ""
echo "✅ Nginx conflict fixed!"
echo ""
echo "📋 Summary:"
echo "   - Removed duplicate server_name configurations"
echo "   - HTTP redirect to HTTPS: ✅"
echo "   - HTTPS server: ✅"
echo ""
echo "🧪 Test commands:"
echo "   curl -I http://${DOMAIN}/health"
echo "   curl https://${DOMAIN}/health"
echo "   curl https://${DOMAIN}/api/v1/csrf-token"

