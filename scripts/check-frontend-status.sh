#!/bin/bash
# Script untuk check status frontend services
# Jalankan langsung di frontend VM: bash check-frontend-status.sh

echo "🔍 Checking Frontend Services Status..."
echo ""

# Check Nginx status
echo "=== Nginx Status ==="
sudo systemctl status nginx --no-pager -l | head -15 || echo "❌ Nginx service not found"

echo ""
echo "Nginx is-active:"
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is active"
else
  echo "❌ Nginx is not active"
fi

echo ""
echo "Nginx is-enabled:"
if sudo systemctl is-enabled --quiet nginx; then
  echo "✅ Nginx is enabled"
else
  echo "❌ Nginx is not enabled"
fi

echo ""
echo "=== Nginx Listening Ports ==="
sudo ss -tlnp | grep nginx || echo "❌ Nginx not listening on any port"

echo ""
echo "=== All Listening Ports (80, 443) ==="
sudo ss -tlnp | grep -E ':(80|443)' || echo "❌ No relevant ports listening"

echo ""
echo "=== Frontend Files ==="
if [ -d /var/www/html ]; then
  echo "✅ /var/www/html directory exists"
  echo "Files:"
  ls -la /var/www/html/ | head -10
else
  echo "❌ /var/www/html directory not found"
fi

echo ""
echo "=== Frontend index.html ==="
if [ -f /var/www/html/index.html ]; then
  echo "✅ index.html exists"
  head -5 /var/www/html/index.html
else
  echo "❌ index.html not found"
fi

echo ""
echo "=== Local Health Check ==="
if curl -s -m 5 http://127.0.0.1/health > /dev/null; then
  echo "✅ Local health check passed"
  curl -s http://127.0.0.1/health
else
  echo "❌ Local health check failed"
fi

echo ""
echo "=== Nginx Config Test ==="
sudo nginx -t 2>&1 || echo "❌ Nginx config has errors"

echo ""
echo "=== Nginx Error Log (last 10 lines) ==="
sudo tail -10 /var/log/nginx/error.log 2>/dev/null || echo "No error log found"

echo ""
echo "✅ Status check completed!"

