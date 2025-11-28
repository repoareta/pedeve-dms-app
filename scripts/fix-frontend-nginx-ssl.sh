#!/bin/bash
set -euo pipefail

# Script untuk fix Nginx config dan install SSL certificate
# Usage: ./fix-frontend-nginx-ssl.sh

DOMAIN="pedeve-dev.aretaamany.com"

echo "🔧 Fixing Nginx config for SSL..."

# Backup existing config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)
  echo "✅ Config backed up"
fi

# Check if certificate exists
if [ -d /etc/letsencrypt/live/${DOMAIN} ]; then
  echo "✅ Certificate exists: /etc/letsencrypt/live/${DOMAIN}"
  sudo ls -la /etc/letsencrypt/live/${DOMAIN}/
else
  echo "❌ Certificate not found. Please run certbot first."
  exit 1
fi

# Create Nginx config with SSL
echo "📝 Creating Nginx config with SSL..."

sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
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
    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

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
EOF

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

# Check ports
echo ""
echo "🔌 Checking ports:"
sudo ss -tlnp | grep -E ':(80|443)' || echo "⚠️  Ports not listening"

# Test HTTPS
echo ""
echo "🔒 Testing HTTPS:"
sleep 2
curl -I https://localhost 2>&1 | head -10 || echo "⚠️  HTTPS test failed"

echo ""
echo "✅ Nginx SSL config fixed!"
echo "🌐 Test: curl -I https://${DOMAIN}"

