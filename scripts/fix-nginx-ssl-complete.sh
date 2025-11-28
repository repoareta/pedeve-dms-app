#!/bin/bash
set -euo pipefail

# Script untuk fix Nginx SSL config secara lengkap
# Usage: ./fix-nginx-ssl-complete.sh

DOMAIN="pedeve-dev.aretaamany.com"

echo "🔍 Checking certificate location..."

# Find certificate
CERT_DIR=""
if [ -d /etc/letsencrypt/live/${DOMAIN} ]; then
  CERT_DIR="/etc/letsencrypt/live/${DOMAIN}"
  echo "✅ Certificate found at: ${CERT_DIR}"
elif [ -d /etc/letsencrypt/live ]; then
  echo "📁 Available certificates:"
  sudo ls -la /etc/letsencrypt/live/
  # Try to find any certificate
  CERT_DIR=$(sudo ls -d /etc/letsencrypt/live/*/ 2>/dev/null | head -1 | sed 's|/$||')
  if [ -n "$CERT_DIR" ]; then
    echo "⚠️  Using certificate from: ${CERT_DIR}"
  fi
else
  echo "❌ No certificates found. Please run certbot first."
  exit 1
fi

# Check certificate files
if [ -f "${CERT_DIR}/fullchain.pem" ] && [ -f "${CERT_DIR}/privkey.pem" ]; then
  echo "✅ Certificate files exist"
  sudo ls -lh ${CERT_DIR}/
else
  echo "❌ Certificate files not found"
  exit 1
fi

# Backup existing config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)
  echo "✅ Config backed up"
fi

# Get certificate directory name (last part of path)
CERT_NAME=$(basename ${CERT_DIR})

echo ""
echo "📝 Creating Nginx config with SSL..."
echo "   Domain: ${DOMAIN}"
echo "   Certificate: ${CERT_DIR}"

# Create Nginx config
sudo tee /etc/nginx/sites-available/default > /dev/null <<NGINX_EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} _;

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};

    # SSL certificate
    ssl_certificate ${CERT_DIR}/fullchain.pem;
    ssl_certificate_key ${CERT_DIR}/privkey.pem;

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
NGINX_EOF

# Test Nginx config
echo ""
echo "🧪 Testing Nginx config:"
if sudo nginx -t; then
  echo "✅ Nginx config is valid"
else
  echo "❌ Nginx config has errors"
  exit 1
fi

# Reload Nginx
echo ""
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx

# Wait a bit
sleep 2

# Check ports
echo ""
echo "🔌 Checking ports:"
sudo ss -tlnp | grep -E ':(80|443)' || echo "⚠️  Ports not listening"

# Test HTTPS
echo ""
echo "🔒 Testing HTTPS:"
curl -I https://localhost 2>&1 | head -10 || echo "⚠️  HTTPS test failed"

echo ""
echo "✅ Nginx SSL config fixed!"
echo "🌐 Test: curl -I https://${DOMAIN}"

