#!/bin/bash
set -euo pipefail

# Script untuk restart services langsung di VM
# Usage: 
#   - Backend VM: ./restart-services-on-vm.sh backend
#   - Frontend VM: ./restart-services-on-vm.sh frontend

VM_TYPE=${1:-}

if [ -z "$VM_TYPE" ]; then
  echo "❌ ERROR: Please specify VM type: backend or frontend"
  echo "Usage: ./restart-services-on-vm.sh [backend|frontend]"
  exit 1
fi

if [ "$VM_TYPE" = "backend" ]; then
  echo "🔧 Restarting backend services..."
  
  # Check Docker container
  echo "📦 Checking Docker container..."
  if sudo docker ps -a | grep -q dms-backend-prod; then
    echo "🔄 Restarting container..."
    sudo docker restart dms-backend-prod 2>/dev/null || sudo docker start dms-backend-prod
    sleep 5
    echo "✅ Container status:"
    sudo docker ps | grep dms-backend-prod || sudo docker ps -a | grep dms-backend-prod
  else
    echo "⚠️  Container not found!"
    echo "Available containers:"
    sudo docker ps -a | head -10
    echo ""
    echo "Available images:"
    sudo docker images | head -5
    exit 1
  fi
  
  # Check Nginx
  echo ""
  echo "🌐 Checking Nginx..."
  if sudo systemctl is-active --quiet nginx; then
    echo "✅ Nginx is running"
    sudo systemctl restart nginx
  else
    echo "🔄 Starting Nginx..."
    sudo systemctl enable nginx
    sudo systemctl start nginx
    sleep 2
  fi
  
  sudo systemctl status nginx --no-pager | head -10
  
  # Check ports
  echo ""
  echo "🔍 Checking ports..."
  sudo ss -tlnp | grep -E ':(80|443|8080)' || echo "⚠️  No ports listening"
  
  # Show logs
  echo ""
  echo "📋 Container logs (last 20 lines):"
  sudo docker logs --tail 20 dms-backend-prod 2>/dev/null || echo "⚠️  Cannot get logs"
  
  echo ""
  echo "✅ Backend services restarted!"
  echo "🔍 Test: curl http://localhost:8080/health"
  
elif [ "$VM_TYPE" = "frontend" ]; then
  echo "🔧 Restarting frontend services..."
  
  # Check Nginx
  echo "🌐 Checking Nginx..."
  if sudo systemctl is-active --quiet nginx; then
    echo "✅ Nginx is running"
    sudo systemctl restart nginx
  else
    echo "🔄 Starting Nginx..."
    sudo systemctl enable nginx
    sudo systemctl start nginx
    sleep 2
  fi
  
  sudo systemctl status nginx --no-pager | head -10
  
  # Check ports
  echo ""
  echo "🔍 Checking ports..."
  sudo ss -tlnp | grep -E ':(80|443)' || echo "⚠️  No ports listening"
  
  # Check files
  echo ""
  echo "📁 Checking files..."
  if [ -d /var/www/html ]; then
    ls -la /var/www/html/ | head -10
  else
    echo "⚠️  /var/www/html not found!"
    exit 1
  fi
  
  echo ""
  echo "✅ Frontend services restarted!"
  echo "🔍 Test: curl http://localhost"
  
else
  echo "❌ ERROR: Invalid VM type. Use 'backend' or 'frontend'"
  exit 1
fi

