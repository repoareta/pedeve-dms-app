#!/bin/bash
set -euo pipefail

# Script untuk memastikan semua services running
# Usage: ./ensure-services-running.sh [backend|frontend]

VM_TYPE=${1:-backend}

if [ "$VM_TYPE" = "backend" ]; then
  echo "🔍 Checking backend services..."
  
  # Ensure Docker container is running with retry mechanism
  MAX_RETRIES=5
  RETRY_COUNT=0
  
  while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if sudo docker ps | grep -q dms-backend-prod; then
      echo "✅ Backend container is running"
      break
    else
      RETRY_COUNT=$((RETRY_COUNT + 1))
      echo "⚠️  Backend container not running (attempt $RETRY_COUNT/$MAX_RETRIES), starting/restarting..."
      
      # Try to start existing container first
      if sudo docker ps -a | grep -q dms-backend-prod; then
        sudo docker start dms-backend-prod || sudo docker restart dms-backend-prod || {
          echo "❌ Failed to start/restart container. Checking status..."
          sudo docker ps -a | grep dms-backend-prod
          sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
          if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
            exit 1
          fi
        }
      else
        echo "❌ ERROR: Container dms-backend-prod does not exist!"
        echo "   Container needs to be deployed first using deploy-backend-vm.sh"
        exit 1
      fi
      
      sleep 5
    fi
  done
  
  if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ ERROR: Backend container failed to start after $MAX_RETRIES attempts!"
    sudo docker ps -a | grep dms-backend-prod
    sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
    exit 1
  fi
  
  # Check port 8080 with retry
  echo "🔍 Checking port 8080..."
  PORT_RETRIES=10
  PORT_COUNT=0
  
  while [ $PORT_COUNT -lt $PORT_RETRIES ]; do
    if sudo ss -tlnp | grep -q ':8080'; then
      echo "✅ Port 8080 is listening"
      break
    else
      PORT_COUNT=$((PORT_COUNT + 1))
      if [ $PORT_COUNT -lt $PORT_RETRIES ]; then
        echo "⏳ Port 8080 not listening yet, waiting... ($PORT_COUNT/$PORT_RETRIES)"
        sleep 3
      fi
    fi
  done
  
  if [ $PORT_COUNT -eq $PORT_RETRIES ]; then
    echo "❌ ERROR: Port 8080 still not listening after $PORT_RETRIES attempts!"
    echo "Container logs:"
    sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
    exit 1
  fi
  
  # Ensure Nginx is running and enabled
  echo "🔍 Checking Nginx..."
  if ! sudo systemctl is-active --quiet nginx; then
    echo "⚠️  Nginx not running, starting..."
    sudo systemctl enable nginx
    sudo systemctl start nginx || sudo systemctl restart nginx
    sleep 3
  else
    # Reload Nginx to ensure latest config is applied
    echo "🔄 Reloading Nginx to apply latest configuration..."
    sudo systemctl reload nginx || sudo systemctl restart nginx
    sleep 2
  fi
  
  # Verify Nginx is running
  if ! sudo systemctl is-active --quiet nginx; then
    echo "❌ ERROR: Nginx failed to start!"
    sudo systemctl status nginx --no-pager -l
    exit 1
  fi
  
  # Final health check
  echo "🏥 Performing final health check..."
  if curl -s -f -m 5 http://127.0.0.1:8080/health > /dev/null 2>&1; then
    echo "✅ Backend health check passed"
  else
    echo "⚠️  WARNING: Backend health check failed, but container is running"
    echo "   This might be normal if backend is still initializing"
  fi
  
  echo "✅ Backend services are running"
  echo "   - Container: $(sudo docker ps | grep dms-backend-prod | awk '{print $1}')"
  echo "   - Port 8080: listening"
  echo "   - Nginx: active"
  
elif [ "$VM_TYPE" = "frontend" ]; then
  echo "🔍 Checking frontend services..."
  
  # Check files first
  if [ ! -f /var/www/html/index.html ]; then
    echo "❌ ERROR: Frontend files not found!"
    ls -la /var/www/html/ | head -10
    exit 1
  fi
  echo "✅ Frontend files are present"
  
  # Ensure Nginx is running and enabled with retry
  MAX_RETRIES=5
  RETRY_COUNT=0
  
  while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if sudo systemctl is-active --quiet nginx; then
      echo "✅ Nginx is running"
      # Reload Nginx to ensure latest config is applied
      echo "🔄 Reloading Nginx to apply latest configuration..."
      sudo systemctl reload nginx || sudo systemctl restart nginx
      sleep 2
      break
    else
      RETRY_COUNT=$((RETRY_COUNT + 1))
      echo "⚠️  Nginx not running (attempt $RETRY_COUNT/$MAX_RETRIES), starting..."
      sudo systemctl enable nginx
      sudo systemctl start nginx || sudo systemctl restart nginx
      sleep 3
    fi
  done
  
  if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ ERROR: Nginx failed to start after $MAX_RETRIES attempts!"
    sudo systemctl status nginx --no-pager -l
    exit 1
  fi
  
  # Check ports with retry
  echo "🔍 Checking port 80..."
  PORT_RETRIES=10
  PORT_COUNT=0
  
  while [ $PORT_COUNT -lt $PORT_RETRIES ]; do
    if sudo ss -tlnp | grep -q ':80 '; then
      echo "✅ Port 80 is listening"
      break
    else
      PORT_COUNT=$((PORT_COUNT + 1))
      if [ $PORT_COUNT -lt $PORT_RETRIES ]; then
        echo "⚠️  Port 80 not listening (attempt $PORT_COUNT/$PORT_RETRIES), restarting Nginx..."
        sudo systemctl restart nginx
        sleep 3
      fi
    fi
  done
  
  if [ $PORT_COUNT -eq $PORT_RETRIES ]; then
    echo "❌ ERROR: Port 80 still not listening after $PORT_RETRIES attempts!"
    sudo systemctl status nginx --no-pager -l
    exit 1
  fi
  
  # Final health check
  echo "🏥 Performing final health check..."
  if curl -s -f -m 5 http://127.0.0.1/health > /dev/null 2>&1; then
    echo "✅ Frontend health check passed"
  else
    echo "⚠️  WARNING: Frontend health check failed, but Nginx is running"
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

