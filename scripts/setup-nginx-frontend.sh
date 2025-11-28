#!/bin/bash
set -euo pipefail

# Script untuk setup Nginx di frontend VM
# Usage: ./setup-nginx-frontend.sh

echo "🔧 Setting up Nginx for frontend..."

# Backup default config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup
fi

# Create Nginx config for SPA
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

# Test Nginx config
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Enable Nginx to start on boot
echo "🔧 Enabling Nginx to start on boot..."
sudo systemctl enable nginx

# Reload Nginx
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx || sudo systemctl restart nginx

# Ensure Nginx is running
echo "▶️  Starting Nginx if not running..."
sudo systemctl start nginx || true

# Check Nginx status
echo "📊 Nginx status:"
sudo systemctl status nginx --no-pager -l || true

# Verify Nginx is active
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is running and enabled"
else
  echo "❌ ERROR: Nginx failed to start!"
  exit 1
fi

echo "✅ Nginx setup completed!"
echo ""
echo "📋 Verification:"
echo "   - Config file: /etc/nginx/sites-available/default"
echo "   - Web root: /var/www/html"
echo "   - Test: curl http://localhost/health"

