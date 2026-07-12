#!/bin/bash
set -euo pipefail

# Script untuk setup SSL certificate untuk backend API
# Usage: ./setup-backend-ssl.sh [DOMAIN]
# Script ini idempotent - aman dipanggil berkali-kali
# 
# Jika DOMAIN tidak diberikan, akan menggunakan default untuk development

# Security: Domain validation function
validate_domain() {
  local domain=$1
  # Domain format: alphanumeric, dots, hyphens, max 253 chars
  # Reject path traversal characters
  if [[ "$domain" =~ [\/\\\$\`\;] ]] || [[ "$domain" =~ \.\. ]]; then
    echo "❌ ERROR: Invalid DOMAIN format. Contains dangerous characters"
    exit 1
  fi
  # Basic domain format validation
  if [[ ! "$domain" =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$ ]]; then
    echo "❌ ERROR: Invalid DOMAIN format"
    exit 1
  fi
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/resolve-domain.sh
source "${SCRIPT_DIR}/lib/resolve-domain.sh"

export DEV_DOMAIN_DEFAULT="api-pedeve-dev.aretaamany.com"
DOMAIN=$(resolve_deploy_domain "${1:-}")
EMAIL="info@aretaamany.com"  # Email untuk Let's Encrypt (dev legacy)

# Security: Validate domain
validate_domain "${DOMAIN}"

echo "🔒 Setting up SSL certificate for ${DOMAIN}..."

# Check if SSL certificate already exists (requires sudo — live/ is root-only)
if ssl_cert_exists_for_domain "${DOMAIN}"; then
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
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    # Security headers
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
set -u

# Enable the config (create symlink to sites-enabled)
echo "🔗 Enabling Nginx config..."
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

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
# Security: Quote all variables
if sudo certbot --nginx \
  -d "${DOMAIN}" \
  --email "${EMAIL}" \
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
echo "   curl https://${DOMAIN}/api/v1/csrf-token"

