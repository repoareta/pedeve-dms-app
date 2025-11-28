#!/bin/bash
set -euo pipefail

# Script untuk check backend status
# Usage: ./check-backend-status.sh

echo "🔍 Checking backend services..."

# Check Docker container
echo ""
echo "📦 Docker Container:"
if sudo docker ps | grep -q dms-backend-prod; then
  echo "✅ Container is running"
  sudo docker ps | grep dms-backend-prod
else
  echo "❌ Container is NOT running"
  echo "Checking stopped containers:"
  sudo docker ps -a | grep dms-backend-prod || echo "⚠️  Container not found"
fi

# Check Nginx
echo ""
echo "🌐 Nginx:"
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is running"
  sudo systemctl status nginx --no-pager | head -10
else
  echo "❌ Nginx is NOT running"
  sudo systemctl status nginx --no-pager | head -10
fi

# Check ports
echo ""
echo "🔍 Ports:"
sudo ss -tlnp | grep -E ':(80|443|8080)' || echo "⚠️  No ports listening"

# Check container logs
echo ""
echo "📋 Container logs (last 20 lines):"
sudo docker logs --tail 20 dms-backend-prod 2>/dev/null || echo "⚠️  Cannot get logs"

# Test health endpoint
echo ""
echo "🏥 Health check:"
curl -s http://localhost:8080/health || echo "❌ Health check failed"

echo ""
echo "✅ Status check complete!"
