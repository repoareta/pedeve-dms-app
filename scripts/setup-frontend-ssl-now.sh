#!/bin/bash
set -euo pipefail

# Script untuk setup SSL di frontend VM
# Usage: ./setup-frontend-ssl-now.sh

echo "🔐 Setting up SSL for frontend..."

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
sudo cat /etc/nginx/sites-available/default | head -30

# Run Certbot
echo ""
echo "🔐 Running Certbot for pedeve-dev.aretaamany.com..."
echo "⚠️  Make sure domain pedeve-dev.aretaamany.com points to this VM's IP (34.128.123.1)"
echo ""

# Certbot will automatically:
# 1. Obtain SSL certificate
# 2. Update Nginx config to include HTTPS
# 3. Set up HTTP to HTTPS redirect
sudo certbot --nginx -d pedeve-dev.aretaamany.com --non-interactive --agree-tos --email info@aretaamany.com || {
  echo "❌ Certbot failed. Trying interactive mode..."
  sudo certbot --nginx -d pedeve-dev.aretaamany.com
}

# Verify certificate
echo ""
echo "🔍 Verifying certificate:"
if [ -d /etc/letsencrypt/live/pedeve-dev.aretaamany.com ]; then
  echo "✅ Certificate directory exists"
  sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/
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
curl -I https://localhost 2>&1 | head -10 || echo "⚠️  HTTPS test failed"

echo ""
echo "✅ SSL setup complete!"
echo "🌐 Test: curl -I https://pedeve-dev.aretaamany.com"

