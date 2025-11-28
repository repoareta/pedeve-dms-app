#!/bin/bash
set -euo pipefail

# Script untuk verify backend deployment setelah deploy
# Usage: ./verify-backend-deployment.sh

echo "🔍 Verifying backend deployment..."

# Ensure Nginx is enabled and running
sudo systemctl enable nginx
sudo systemctl daemon-reload
sudo systemctl start nginx || sudo systemctl restart nginx

# Wait for services to stabilize
sleep 5

# Verify Nginx is running with retry
MAX_RETRIES=3
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if sudo systemctl is-active --quiet nginx; then
    echo '✅ Nginx is running'
    break
  else
    RETRY_COUNT=$(expr $RETRY_COUNT + 1)
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
      echo "⚠️  Nginx not running, retrying... attempt $RETRY_COUNT of $MAX_RETRIES"
      sudo systemctl restart nginx
      sleep 3
    else
      echo '❌ ERROR: Nginx failed to start after retries!'
      sudo systemctl status nginx --no-pager -l
      sudo tail -20 /var/log/nginx/error.log 2>/dev/null || true
      exit 1
    fi
  fi
done

# Ensure Nginx will auto-start on boot
if ! sudo systemctl is-enabled --quiet nginx; then
  echo '⚠️  WARNING: Nginx is not enabled for auto-start, enabling now...'
  sudo systemctl enable nginx
fi

# Verify listening ports
echo 'Checking listening ports...'
if ! sudo ss -tlnp | grep -q ':80 '; then
  echo '❌ ERROR: Port 80 is not listening!'
  exit 1
fi
if ! sudo ss -tlnp | grep -q ':443 '; then
  echo '⚠️  WARNING: Port 443 is not listening (HTTPS may not be configured)'
  echo 'Checking SSL certificate...'
  sudo certbot certificates 2>/dev/null | head -10 || true
else
  echo '✅ Port 443 is listening'
fi
if ! sudo ss -tlnp | grep -q ':8080'; then
  echo '❌ ERROR: Port 8080 is not listening!'
  exit 1
fi

# Verify backend container is running
if ! sudo docker ps | grep -q dms-backend-prod; then
  echo '❌ ERROR: Backend container is not running!'
  sudo docker ps -a | grep dms-backend-prod
  sudo docker logs --tail 30 dms-backend-prod 2>/dev/null || true
  exit 1
fi

echo '✅ Backend container is running'
echo '✅ Nginx is running and enabled'

# Copy status check script untuk debugging
echo '#!/bin/bash' > ~/check-backend-status.sh
echo 'sudo systemctl status nginx --no-pager -l | head -10' >> ~/check-backend-status.sh
echo 'sudo docker ps | grep dms-backend-prod' >> ~/check-backend-status.sh
echo 'sudo ss -tlnp | grep 80' >> ~/check-backend-status.sh
echo 'sudo ss -tlnp | grep 443' >> ~/check-backend-status.sh
echo 'sudo ss -tlnp | grep 8080' >> ~/check-backend-status.sh
chmod +x ~/check-backend-status.sh

echo '✅ Backend deployment verification completed!'

