#!/bin/bash
set -euo pipefail

# Script untuk verify frontend deployment setelah deploy
# Usage: ./verify-frontend-deployment.sh

echo "🔍 Verifying frontend deployment..."

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

# Verify files exist
if [ ! -f /var/www/html/index.html ]; then
  echo '❌ ERROR: Frontend files not found!'
  ls -la /var/www/html/ | head -10
  exit 1
fi

echo '✅ Nginx is running and enabled'
echo '✅ Frontend files deployed'

# Copy status check script untuk debugging
echo '#!/bin/bash' > ~/check-frontend-status.sh
echo 'sudo systemctl status nginx --no-pager -l | head -10' >> ~/check-frontend-status.sh
echo 'sudo ss -tlnp | grep 80' >> ~/check-frontend-status.sh
echo 'sudo ss -tlnp | grep 443' >> ~/check-frontend-status.sh
echo 'ls -la /var/www/html/ | head -5' >> ~/check-frontend-status.sh
chmod +x ~/check-frontend-status.sh

echo '✅ Frontend deployment verification completed!'

