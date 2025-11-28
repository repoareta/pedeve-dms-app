#!/bin/bash
set -euo pipefail

# Script untuk setup SSL di backend VM
# Usage: ./setup-backend-ssl-now.sh

DOMAIN="api-pedeve-dev.aretaamany.com"

echo "🔐 Setting up SSL for backend..."

# Check if Certbot is installed
if ! command -v certbot &> /dev/null; then
  echo "📦 Installing Certbot..."
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
else
  echo "✅ Certbot already installed"
fi

# Check current Nginx config
echo ""
echo "📋 Current Nginx config:"
if [ -f /etc/nginx/sites-available/backend-api ]; then
  sudo cat /etc/nginx/sites-available/backend-api | head -30
else
  echo "⚠️  Config file not found"
fi

# Run Certbot
echo ""
echo "🔐 Running Certbot for ${DOMAIN}..."
echo "⚠️  Make sure domain ${DOMAIN} points to this VM's IP (34.101.49.147)"
echo ""

# Certbot will automatically:
# 1. Obtain SSL certificate
# 2. Update Nginx config to include HTTPS
# 3. Set up HTTP to HTTPS redirect
sudo certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos --email info@aretaamany.com || {
  echo "❌ Certbot failed. Trying interactive mode..."
  sudo certbot --nginx -d ${DOMAIN}
}

# Verify certificate
echo ""
echo "🔍 Verifying certificate:"
if [ -d /etc/letsencrypt/live/${DOMAIN} ]; then
  echo "✅ Certificate directory exists"
  sudo ls -la /etc/letsencrypt/live/${DOMAIN}/
else
  echo "❌ Certificate directory not found"
  exit 1
fi

# Test Nginx config
echo ""
echo "🧪 Testing Nginx config:"
sudo nginx -t

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
curl -I https://localhost/health 2>&1 | head -10 || echo "⚠️  HTTPS test failed"

echo ""
echo "✅ SSL setup complete!"
echo "🌐 Test: curl -I https://${DOMAIN}/health"

