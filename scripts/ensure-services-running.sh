#!/bin/bash
set -euo pipefail

# Script untuk memastikan semua services running
# Usage: ./ensure-services-running.sh [backend|frontend]

VM_TYPE=${1:-backend}

if [ "$VM_TYPE" = "backend" ]; then
  echo "🔍 Checking backend services..."
  
  # Check Docker container
  if ! sudo docker ps | grep -q dms-backend-prod; then
    echo "⚠️  Backend container not running, starting..."
    sudo docker start dms-backend-prod || {
      echo "❌ Failed to start container. Checking status..."
      sudo docker ps -a | grep dms-backend-prod
      sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
      exit 1
    }
    sleep 5
  fi
  
  # Check if container is healthy
  if ! sudo docker ps | grep -q dms-backend-prod; then
    echo "❌ ERROR: Backend container failed to start!"
    sudo docker ps -a | grep dms-backend-prod
    sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
    exit 1
  fi
  
  # Check port 8080
  if ! sudo ss -tlnp | grep -q ':8080'; then
    echo "⚠️  Port 8080 not listening, container may be starting..."
    sleep 5
    if ! sudo ss -tlnp | grep -q ':8080'; then
      echo "❌ ERROR: Port 8080 still not listening!"
      sudo docker logs --tail 30 dms-backend-prod 2>/dev/null || true
      exit 1
    fi
  fi
  
  # Check Nginx
  if ! sudo systemctl is-active --quiet nginx; then
    echo "⚠️  Nginx not running, starting..."
    sudo systemctl enable nginx
    sudo systemctl start nginx || sudo systemctl restart nginx
    sleep 3
  fi
  
  if ! sudo systemctl is-active --quiet nginx; then
    echo "❌ ERROR: Nginx failed to start!"
    sudo systemctl status nginx --no-pager -l
    exit 1
  fi
  
  echo "✅ Backend services are running"
  echo "   - Container: $(sudo docker ps | grep dms-backend-prod | awk '{print $1}')"
  echo "   - Port 8080: listening"
  echo "   - Nginx: active"
  
elif [ "$VM_TYPE" = "frontend" ]; then
  echo "🔍 Checking frontend services..."
  
  # Check Nginx
  if ! sudo systemctl is-active --quiet nginx; then
    echo "⚠️  Nginx not running, starting..."
    sudo systemctl enable nginx
    sudo systemctl start nginx || sudo systemctl restart nginx
    sleep 3
  fi
  
  if ! sudo systemctl is-active --quiet nginx; then
    echo "❌ ERROR: Nginx failed to start!"
    sudo systemctl status nginx --no-pager -l
    exit 1
  fi
  
  # Check files
  if [ ! -f /var/www/html/index.html ]; then
    echo "❌ ERROR: Frontend files not found!"
    ls -la /var/www/html/ | head -10
    exit 1
  fi
  
  # Check ports
  if ! sudo ss -tlnp | grep -q ':80 '; then
    echo "⚠️  Port 80 not listening, restarting Nginx..."
    sudo systemctl restart nginx
    sleep 3
    if ! sudo ss -tlnp | grep -q ':80 '; then
      echo "❌ ERROR: Port 80 still not listening!"
      sudo systemctl status nginx --no-pager -l
      exit 1
    fi
  fi
  
  echo "✅ Frontend services are running"
  echo "   - Nginx: active"
  echo "   - Port 80: listening"
  echo "   - Files: present"
else
  echo "❌ ERROR: Invalid VM type. Use 'backend' or 'frontend'"
  exit 1
fi

echo "✅ All services are running!"

