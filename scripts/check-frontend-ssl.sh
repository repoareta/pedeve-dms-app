#!/bin/bash
set -euo pipefail

# Script untuk check frontend SSL dan ports
# Usage: ./check-frontend-ssl.sh

echo "🔍 Checking frontend SSL and ports..."

# Check ports
echo ""
echo "🔌 Ports listening:"
sudo ss -tlnp | grep -E ':(80|443)' || echo "⚠️  No ports listening"

# Check SSL certificates
echo ""
echo "🔐 SSL Certificates:"
if [ -d /etc/letsencrypt/live/pedeve-dev.aretaamany.com ]; then
  echo "✅ SSL certificate directory exists"
  ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/
  
  echo ""
  echo "📄 Certificate files:"
  if [ -f /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem ]; then
    echo "✅ fullchain.pem exists"
    ls -lh /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem
  else
    echo "❌ fullchain.pem NOT found"
  fi
  
  if [ -f /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem ]; then
    echo "✅ privkey.pem exists"
    ls -lh /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem
  else
    echo "❌ privkey.pem NOT found"
  fi
else
  echo "❌ SSL certificate directory NOT found"
  echo "Available certificates:"
  ls -la /etc/letsencrypt/live/ 2>/dev/null || echo "⚠️  No certificates found"
fi

# Check Nginx config
echo ""
echo "🌐 Nginx configuration:"
if [ -f /etc/nginx/sites-available/default ]; then
  echo "✅ Config file exists"
  echo ""
  echo "📋 Config content (HTTPS section):"
  sudo grep -A 20 "listen.*443" /etc/nginx/sites-available/default || echo "⚠️  No HTTPS config found"
else
  echo "❌ Config file NOT found"
  echo "Available configs:"
  ls -la /etc/nginx/sites-available/ 2>/dev/null || echo "⚠️  No configs found"
fi

# Check enabled sites
echo ""
echo "🔗 Enabled sites:"
ls -la /etc/nginx/sites-enabled/ 2>/dev/null || echo "⚠️  No enabled sites"

# Test Nginx config
echo ""
echo "🧪 Nginx config test:"
sudo nginx -t

# Test HTTP
echo ""
echo "🌐 HTTP test (port 80):"
curl -I http://localhost 2>&1 | head -5

# Test HTTPS
echo ""
echo "🔒 HTTPS test (port 443):"
curl -I https://localhost 2>&1 | head -5 || echo "❌ HTTPS not accessible"

# Test external
echo ""
echo "🌍 External HTTP test:"
curl -I http://34.128.123.1 2>&1 | head -5

echo ""
echo "🌍 External HTTPS test:"
curl -I https://34.128.123.1 2>&1 | head -5 || echo "❌ External HTTPS not accessible"

echo ""
echo "✅ Check complete!"

