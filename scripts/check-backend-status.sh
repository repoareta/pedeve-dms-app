#!/bin/bash
# Script untuk check status backend services
# Jalankan langsung di backend VM: bash check-backend-status.sh

echo "🔍 Checking Backend Services Status..."
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
echo "=== All Listening Ports (80, 443, 8080) ==="
sudo ss -tlnp | grep -E ':(80|443|8080)' || echo "❌ No relevant ports listening"

echo ""
echo "=== Docker Container Status ==="
sudo docker ps -a | grep dms-backend-prod || echo "❌ No backend container found"

echo ""
echo "=== Container Logs (last 20 lines) ==="
sudo docker logs --tail 20 dms-backend-prod 2>/dev/null || echo "❌ Cannot get container logs"

echo ""
echo "=== Port 8080 Status ==="
if sudo ss -tlnp | grep -q ':8080'; then
  echo "✅ Port 8080 is listening"
  sudo ss -tlnp | grep ':8080'
else
  echo "❌ Port 8080 is not listening"
fi

echo ""
echo "=== Container Health Check ==="
if curl -s -m 5 http://127.0.0.1:8080/health > /dev/null; then
  echo "✅ Container health check passed"
  curl -s http://127.0.0.1:8080/health
else
  echo "❌ Container health check failed"
fi

echo ""
echo "=== Nginx Config Test ==="
sudo nginx -t 2>&1 || echo "❌ Nginx config has errors"

echo ""
echo "=== Nginx Error Log (last 10 lines) ==="
sudo tail -10 /var/log/nginx/error.log 2>/dev/null || echo "No error log found"

echo ""
echo "=== Nginx Access Log (last 5 lines) ==="
sudo tail -5 /var/log/nginx/backend-api-access.log 2>/dev/null || echo "No access log found"

echo ""
echo "✅ Status check completed!"

