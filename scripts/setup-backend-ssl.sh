#!/bin/bash
set -euo pipefail

# Script untuk setup SSL certificate untuk backend API
# Usage: ./setup-backend-ssl.sh

DOMAIN="api-pedeve-dev.aretaamany.com"
EMAIL="info@aretaamany.com"  # Email untuk Let's Encrypt

echo "🔒 Setting up SSL certificate for ${DOMAIN}..."

# Install Certbot if not exists
if ! command -v certbot &> /dev/null; then
  echo "📦 Installing Certbot..."
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
fi

# Update Nginx config untuk support SSL
echo "📝 Updating Nginx config for SSL..."

sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    # SSL certificate (will be set by Certbot)
    # ssl_certificate /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/privkey.pem;

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

# Reload Nginx
sudo systemctl reload nginx

# Generate SSL certificate with Certbot
echo "🔐 Generating SSL certificate..."
sudo certbot --nginx \
  -d ${DOMAIN} \
  --email ${EMAIL} \
  --agree-tos \
  --non-interactive \
  --redirect

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
echo "   curl https://${DOMAIN}/api/v1/csrf-token"

