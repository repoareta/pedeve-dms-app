#!/bin/bash
set -euo pipefail

# Script untuk restart services secara manual
# Usage: ./restart-services-manual.sh [backend|frontend|both]

VM_TYPE=${1:-both}

echo "🔄 Restarting services: $VM_TYPE"

if [ "$VM_TYPE" = "backend" ] || [ "$VM_TYPE" = "both" ]; then
  echo ""
  echo "🔧 Backend VM (backend-dev)..."
  gcloud compute ssh backend-dev \
    --zone=asia-southeast2-a \
    --project=pedeve-pertamina-dms \
    --command="
      echo '📦 Checking Docker container...'
      if sudo docker ps -a | grep -q dms-backend-prod; then
        echo '🔄 Restarting container...'
        sudo docker restart dms-backend-prod || sudo docker start dms-backend-prod
        sleep 5
        echo '✅ Container status:'
        sudo docker ps | grep dms-backend-prod || sudo docker ps -a | grep dms-backend-prod
      else
        echo '⚠️  Container not found, checking images...'
        sudo docker images | head -5
      fi
      
      echo ''
      echo '🌐 Checking Nginx...'
      if sudo systemctl is-active --quiet nginx; then
        echo '✅ Nginx is running'
      else
        echo '🔄 Starting Nginx...'
        sudo systemctl enable nginx
        sudo systemctl start nginx
        sleep 2
        sudo systemctl status nginx --no-pager | head -10
      fi
      
      echo ''
      echo '🔍 Checking ports...'
      sudo ss -tlnp | grep -E ':(80|443|8080)' || echo '⚠️  No ports listening'
      
      echo ''
      echo '📋 Container logs (last 20 lines):'
      sudo docker logs --tail 20 dms-backend-prod 2>/dev/null || echo '⚠️  Cannot get logs'
    "
fi

if [ "$VM_TYPE" = "frontend" ] || [ "$VM_TYPE" = "both" ]; then
  echo ""
  echo "🔧 Frontend VM (frontend-dev)..."
  gcloud compute ssh frontend-dev \
    --zone=asia-southeast2-a \
    --project=pedeve-pertamina-dms \
    --command="
      echo '🌐 Checking Nginx...'
      if sudo systemctl is-active --quiet nginx; then
        echo '✅ Nginx is running'
        sudo systemctl restart nginx
      else
        echo '🔄 Starting Nginx...'
        sudo systemctl enable nginx
        sudo systemctl start nginx
        sleep 2
      fi
      sudo systemctl status nginx --no-pager | head -10
      
      echo ''
      echo '🔍 Checking ports...'
      sudo ss -tlnp | grep -E ':(80|443)' || echo '⚠️  No ports listening'
      
      echo ''
      echo '📁 Checking files...'
      ls -la /var/www/html/ | head -10 || echo '⚠️  Files not found'
    "
fi

echo ""
echo "✅ Restart complete!"
echo ""
echo "🔍 Verify services:"
if [ "$VM_TYPE" = "backend" ] || [ "$VM_TYPE" = "both" ]; then
  echo "   Backend: curl http://34.101.49.147:8080/health"
  echo "   Backend API: curl https://api-pedeve-dev.aretaamany.com/health"
fi
if [ "$VM_TYPE" = "frontend" ] || [ "$VM_TYPE" = "both" ]; then
  echo "   Frontend: curl http://34.128.123.1"
  echo "   Frontend HTTPS: curl https://pedeve-dev.aretaamany.com"
fi

