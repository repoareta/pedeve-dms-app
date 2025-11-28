#!/bin/bash
set -euo pipefail

# Script untuk setup Nginx reverse proxy di backend VM
# Usage: ./setup-backend-nginx.sh

echo "🔧 Setting up Nginx reverse proxy for backend..."

# Install Nginx if not exists
if ! command -v nginx &> /dev/null; then
  echo "📦 Installing Nginx..."
  sudo apt-get update
  sudo apt-get install -y nginx
  sudo systemctl enable nginx
fi

# Backup default config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup
fi

# Check if HTTPS config already exists (SSL already setup)
if [ -f /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem ]; then
  echo "✅ SSL certificate found, creating config with HTTPS..."
  
  # Create Nginx config with HTTPS
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
else
  echo "⚠️  SSL certificate not found, creating HTTP-only config..."
  
  # Create Nginx config for backend API reverse proxy (HTTP only)
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
fi

# Hapus semua enabled sites untuk avoid conflict
sudo rm -f /etc/nginx/sites-enabled/*

# Enable backend-api site
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# Hapus config frontend jika ter-copy (pastikan tidak ada conflict)
# (tidak perlu, karena frontend pakai default)

# Test Nginx config
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Enable Nginx to start on boot
echo "🔧 Enabling Nginx to start on boot..."
sudo systemctl enable nginx

# Reload Nginx (reload is safer than restart, preserves connections)
echo "🔄 Reloading Nginx..."
if sudo nginx -t 2>/dev/null; then
  sudo systemctl reload nginx || sudo systemctl restart nginx
else
  echo "⚠️  Nginx config test failed, trying restart..."
  sudo systemctl restart nginx
fi

# Ensure Nginx is running
echo "▶️  Starting Nginx if not running..."
sudo systemctl start nginx || sudo systemctl restart nginx

# Wait a moment for Nginx to fully start
sleep 2

# Check Nginx status
echo "📊 Nginx status:"
sudo systemctl status nginx --no-pager -l || true

# Verify Nginx is active
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is running and enabled"
  
  # Verify listening ports
  echo "📡 Checking listening ports..."
  if sudo ss -tlnp | grep -q ':80 '; then
    echo "✅ Port 80 is listening"
  else
    echo "⚠️  Port 80 is not listening"
  fi
  
  if sudo ss -tlnp | grep -q ':443 '; then
    echo "✅ Port 443 is listening"
  else
    echo "⚠️  Port 443 is not listening (HTTPS may not be configured)"
  fi
else
  echo "❌ ERROR: Nginx failed to start!"
  echo "Nginx error log:"
  sudo tail -20 /var/log/nginx/error.log 2>/dev/null || true
  exit 1
fi

echo "✅ Nginx reverse proxy setup completed!"
echo ""
echo "📋 Configuration:"
echo "   - Listen: port 80"
echo "   - Server name: api-pedeve-dev.aretaamany.com"
echo "   - Proxy to: http://127.0.0.1:8080"
echo ""
echo "🧪 Test commands:"
echo "   curl http://api-pedeve-dev.aretaamany.com/health"
echo "   curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token"

