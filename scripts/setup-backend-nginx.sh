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

# Create Nginx config for backend API reverse proxy
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

# Enable site
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# Remove default site if exists
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx config
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Reload Nginx
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx || sudo systemctl restart nginx

# Check Nginx status
echo "📊 Nginx status:"
sudo systemctl status nginx --no-pager -l || true

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

