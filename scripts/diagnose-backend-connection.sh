#!/bin/bash
set -euo pipefail

# Script untuk diagnose backend connection issues
# Usage: ./diagnose-backend-connection.sh

echo "🔍 Diagnosing backend connection issues..."

# 1. Check Docker container
echo ""
echo "📦 Docker Container:"
if sudo docker ps | grep -q dms-backend-prod; then
  echo "✅ Container is running"
  sudo docker ps | grep dms-backend-prod
else
  echo "❌ Container is NOT running"
  echo "Checking stopped containers:"
  sudo docker ps -a | grep dms-backend-prod || echo "⚠️  Container not found"
  echo ""
  echo "🔄 Attempting to start container..."
  sudo docker start dms-backend-prod 2>/dev/null || echo "❌ Failed to start"
fi

# 2. Check container logs
echo ""
echo "📋 Container logs (last 30 lines):"
sudo docker logs --tail 30 dms-backend-prod 2>/dev/null || echo "⚠️  Cannot get logs"

# 3. Check port 8080
echo ""
echo "🔌 Port 8080:"
if sudo ss -tlnp | grep -q ':8080'; then
  echo "✅ Port 8080 is listening"
  sudo ss -tlnp | grep ':8080'
else
  echo "❌ Port 8080 is NOT listening"
fi

# 4. Test backend health directly
echo ""
echo "🏥 Backend health check (direct):"
curl -s -m 5 http://127.0.0.1:8080/health || echo "❌ Health check failed"

# 5. Check Nginx
echo ""
echo "🌐 Nginx:"
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is running"
  sudo systemctl status nginx --no-pager | head -10
else
  echo "❌ Nginx is NOT running"
  echo "🔄 Attempting to start Nginx..."
  sudo systemctl start nginx 2>/dev/null || echo "❌ Failed to start"
fi

# 6. Check Nginx ports
echo ""
echo "🔌 Nginx ports:"
sudo ss -tlnp | grep -E ':(80|443)' || echo "⚠️  No ports listening"

# 7. Check Nginx config
echo ""
echo "📋 Nginx config test:"
sudo nginx -t 2>&1 || echo "❌ Config test failed"

# 8. Check SSL certificate (if exists)
echo ""
echo "🔐 SSL Certificate:"
if [ -f /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem ]; then
  echo "✅ SSL certificate exists"
  sudo ls -lh /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/
else
  echo "⚠️  SSL certificate not found (HTTPS may not work)"
fi

# 9. Check Nginx config for backend-api
echo ""
echo "📄 Nginx backend-api config:"
if [ -f /etc/nginx/sites-available/backend-api ]; then
  echo "✅ Config file exists"
  echo "Server name:"
  sudo grep "server_name" /etc/nginx/sites-available/backend-api || echo "⚠️  No server_name found"
  echo "Listen ports:"
  sudo grep "listen" /etc/nginx/sites-available/backend-api || echo "⚠️  No listen found"
  echo "Proxy pass:"
  sudo grep "proxy_pass" /etc/nginx/sites-available/backend-api || echo "⚠️  No proxy_pass found"
else
  echo "❌ Config file NOT found"
fi

# 10. Test via Nginx (HTTP)
echo ""
echo "🌐 Test via Nginx (HTTP):"
curl -s -m 5 -I http://127.0.0.1/health 2>&1 | head -10 || echo "❌ HTTP test failed"

# 11. Test via Nginx (HTTPS)
echo ""
echo "🔒 Test via Nginx (HTTPS):"
curl -s -m 5 -I -k https://127.0.0.1/health 2>&1 | head -10 || echo "❌ HTTPS test failed"

# 12. Test external (if domain resolves)
echo ""
echo "🌍 Test external domain:"
curl -s -m 5 -I https://api-pedeve-dev.aretaamany.com/health 2>&1 | head -10 || echo "❌ External test failed"

echo ""
echo "✅ Diagnosis complete!"
echo ""
echo "💡 Common fixes:"
echo "   1. If container not running: sudo docker start dms-backend-prod"
echo "   2. If Nginx not running: sudo systemctl start nginx"
echo "   3. If SSL not configured: sudo certbot --nginx -d api-pedeve-dev.aretaamany.com"
echo "   4. If config wrong: check /etc/nginx/sites-available/backend-api"

