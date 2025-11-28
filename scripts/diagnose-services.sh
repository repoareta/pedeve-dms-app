#!/bin/bash
set -euo pipefail

# Script untuk diagnose services di VM
# Usage: ./diagnose-services.sh <vm-name> <vm-zone> <project-id>

VM_NAME=$1
VM_ZONE=$2
PROJECT_ID=$3

echo "🔍 Diagnosing services on $VM_NAME..."
echo ""

# Check Nginx status
echo "=== Nginx Status ==="
gcloud compute ssh $VM_NAME \
  --zone=$VM_ZONE \
  --project=$PROJECT_ID \
  --command="
    echo 'Nginx service status:'
    sudo systemctl status nginx --no-pager -l || echo '❌ Nginx service not found'
    
    echo ''
    echo 'Nginx is-active:'
    sudo systemctl is-active nginx || echo '❌ Nginx is not active'
    
    echo ''
    echo 'Nginx is-enabled:'
    sudo systemctl is-enabled nginx || echo '❌ Nginx is not enabled'
    
    echo ''
    echo 'Nginx process:'
    ps aux | grep nginx | grep -v grep || echo '❌ No Nginx process found'
    
    echo ''
    echo 'Nginx listening ports:'
    sudo ss -tlnp | grep nginx || echo '❌ Nginx not listening on any port'
    
    echo ''
    echo 'Nginx config test:'
    sudo nginx -t 2>&1 || echo '❌ Nginx config has errors'
  "

# Check Docker container (for backend VM)
if [ "$VM_NAME" = "backend-dev" ]; then
  echo ""
  echo "=== Docker Container Status ==="
  gcloud compute ssh $VM_NAME \
    --zone=$VM_ZONE \
    --project=$PROJECT_ID \
    --command="
      echo 'Docker containers:'
      sudo docker ps -a | grep dms-backend || echo '❌ No backend container found'
      
      echo ''
      echo 'Backend container logs (last 30 lines):'
      sudo docker logs --tail 30 dms-backend-prod 2>&1 || echo '❌ Cannot get container logs'
      
      echo ''
      echo 'Port 8080 status:'
      sudo ss -tlnp | grep 8080 || echo '❌ Port 8080 not listening'
      
      echo ''
      echo 'Container health check:'
      curl -s -m 5 http://127.0.0.1:8080/health || echo '❌ Container health check failed'
    "
fi

# Check frontend files (for frontend VM)
if [ "$VM_NAME" = "frontend-dev" ]; then
  echo ""
  echo "=== Frontend Files ==="
  gcloud compute ssh $VM_NAME \
    --zone=$VM_ZONE \
    --project=$PROJECT_ID \
    --command="
      echo 'Frontend files:'
      ls -la /var/www/html/ | head -10 || echo '❌ No files found'
      
      echo ''
      echo 'Frontend index.html:'
      head -20 /var/www/html/index.html 2>/dev/null || echo '❌ index.html not found'
      
      echo ''
      echo 'Port 80 status:'
      sudo ss -tlnp | grep ':80 ' || echo '❌ Port 80 not listening'
      
      echo ''
      echo 'Local health check:'
      curl -s -m 5 http://127.0.0.1/health || echo '❌ Local health check failed'
    "
fi

# Check firewall and network
echo ""
echo "=== Network & Firewall ==="
gcloud compute ssh $VM_NAME \
  --zone=$VM_ZONE \
  --project=$PROJECT_ID \
  --command="
    echo 'VM External IP:'
    curl -s ifconfig.me || echo 'Cannot get external IP'
    
    echo ''
    echo 'Listening ports:'
    sudo ss -tlnp | grep -E ':(80|443|8080)' || echo 'No relevant ports listening'
    
    echo ''
    echo 'Nginx error log (last 20 lines):'
    sudo tail -20 /var/log/nginx/error.log 2>/dev/null || echo 'No error log found'
  "

echo ""
echo "✅ Diagnosis completed!"

