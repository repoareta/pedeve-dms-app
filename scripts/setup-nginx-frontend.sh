#!/bin/bash
set -euo pipefail

# Script untuk setup Nginx di frontend VM
# Usage: ./setup-nginx-frontend.sh

echo "🔧 Setting up Nginx for frontend..."

# Backup default config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup
fi

# Check if HTTPS config already exists (SSL already setup)
if [ -f /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem ]; then
  echo "✅ SSL certificate found, creating config with HTTPS..."
  
  # Create Nginx config with HTTPS
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
else
  echo "⚠️  SSL certificate not found, creating HTTP-only config..."
  
  # Create Nginx config for SPA (HTTP only)
  sudo tee /etc/nginx/sites-available/default > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name _;

    root /var/www/html;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # SPA routing - semua request ke index.html kecuali static files
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint (optional)
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF
fi

# Hapus semua enabled sites untuk avoid conflict
echo "🧹 Cleaning up enabled sites..."
sudo rm -f /etc/nginx/sites-enabled/*

# Hapus config backend jika ter-copy (pastikan tidak ada conflict)
sudo rm -f /etc/nginx/sites-available/backend-api

# Enable default site
sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default

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

echo "✅ Nginx setup completed!"
echo ""
echo "📋 Verification:"
echo "   - Config file: /etc/nginx/sites-available/default"
echo "   - Web root: /var/www/html"
echo "   - Test: curl http://localhost/health"

