#!/bin/bash
set -euo pipefail

# Script untuk setup SSL certificate untuk frontend
# Usage: ./setup-frontend-ssl.sh [DOMAIN]
# Script ini idempotent - aman dipanggil berkali-kali
# 
# Jika DOMAIN tidak diberikan, akan menggunakan default untuk development

DOMAIN=${1:-"pedeve-dev.aretaamany.com"}
EMAIL="info@aretaamany.com"  # Email untuk Let's Encrypt

echo "🔒 Setting up SSL certificate for ${DOMAIN}..."

# Check if SSL certificate already exists
if [ -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem ] && \
   [ -f /etc/letsencrypt/live/${DOMAIN}/privkey.pem ]; then
  echo "✅ SSL certificate already exists for ${DOMAIN}"
  echo "   - Certificate: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
  echo "   - Private key: /etc/letsencrypt/live/${DOMAIN}/privkey.pem"
  echo "⏭️  Skipping SSL certificate generation"
  
  # Just ensure auto-renewal is enabled
  if ! sudo systemctl is-enabled certbot.timer &>/dev/null; then
    echo "🔄 Enabling auto-renewal..."
    sudo systemctl enable certbot.timer
    sudo systemctl start certbot.timer
  fi
  
  echo "✅ SSL certificate setup completed (certificate already exists)"
  exit 0
fi

# Install Certbot if not exists
if ! command -v certbot &> /dev/null; then
  echo "📦 Installing Certbot..."
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
fi

# Ensure Nginx is running (required for Certbot)
if ! sudo systemctl is-active --quiet nginx; then
  echo "⚠️  Nginx is not running, starting it..."
  sudo systemctl start nginx || {
    echo "❌ ERROR: Cannot start Nginx. Please check Nginx configuration first."
    exit 1
  }
fi

# Update Nginx config untuk HTTP-only (Certbot will add HTTPS block automatically)
echo "📝 Updating Nginx config for HTTP (Certbot will add HTTPS automatically)..."

# Temporarily disable unbound variable check for heredoc (Nginx variables will be evaluated by Nginx, not bash)
set +u
sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} _;

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

    # Health check endpoint (optional)
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF
set -u

# Test Nginx config before reloading
echo "🧪 Testing Nginx configuration..."
if ! sudo nginx -t; then
  echo "❌ ERROR: Nginx configuration test failed!"
  exit 1
fi

# Reload Nginx
echo "🔄 Reloading Nginx..."
sudo systemctl reload nginx || sudo systemctl restart nginx

# Generate SSL certificate with Certbot
# Certbot will automatically:
# 1. Create SSL certificate
# 2. Add HTTPS block to Nginx config
# 3. Configure HTTP to HTTPS redirect
echo "🔐 Generating SSL certificate with Certbot..."
if sudo certbot --nginx \
  -d ${DOMAIN} \
  --email ${EMAIL} \
  --agree-tos \
  --non-interactive \
  --redirect; then
  echo "✅ SSL certificate generated successfully"
  echo "✅ Certbot has automatically configured HTTPS in Nginx"
else
  echo "⚠️  WARNING: SSL certificate generation may have failed"
  echo "   This might be normal if:"
  echo "   - Certificate already exists"
  echo "   - DNS is not configured correctly"
  echo "   - Let's Encrypt rate limit reached"
  echo "   - Port 80 is not accessible from internet"
  # Don't exit with error - certificate might already exist
fi

# Setup auto-renewal
echo "🔄 Setting up auto-renewal..."
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Test renewal
echo "🧪 Testing certificate renewal..."
sudo certbot renew --dry-run

echo ""
echo "✅ SSL certificate setup completed!"
echo ""
echo "📋 Summary:"
echo "   - Domain: ${DOMAIN}"
echo "   - SSL Certificate: Let's Encrypt"
echo "   - Auto-renewal: Enabled"
echo ""
echo "🧪 Test commands:"
echo "   curl https://${DOMAIN}/health"
echo "   curl -I http://${DOMAIN}/health  # Should redirect to HTTPS"

