#!/bin/bash
set -euo pipefail

# Script untuk deploy backend di VM
# Usage: ./deploy-backend-vm.sh <PROJECT_ID> <BACKEND_IMAGE>

PROJECT_ID=$1
BACKEND_IMAGE=$2

echo "🚀 Starting backend deployment on VM..."

# Install Docker if not exists
if ! command -v docker &> /dev/null; then
  echo "📦 Installing Docker..."
  curl -fsSL https://get.docker.com -o get-docker.sh
  sudo sh get-docker.sh
  sudo usermod -aG docker $USER || true
  rm -f get-docker.sh
fi

# Load Docker image
echo "🐳 Loading Docker image..."
sudo docker load -i ~/backend-image.tar

# Stop old container
echo "🛑 Stopping old container..."
sudo docker stop dms-backend-prod 2>/dev/null || true
sudo docker rm dms-backend-prod 2>/dev/null || true

# Get secrets from GCP Secret Manager
echo "🔑 Getting secrets from GCP Secret Manager..."
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${PROJECT_ID} 2>/dev/null || echo '')
JWT_SECRET=$(gcloud secrets versions access latest --secret=jwt_secret --project=${PROJECT_ID} 2>/dev/null || echo '')
ENCRYPTION_KEY=$(gcloud secrets versions access latest --secret=encryption_key --project=${PROJECT_ID} 2>/dev/null || echo '')

# Verify secrets were retrieved
if [ -z "${DB_PASSWORD}" ]; then
  echo "❌ ERROR: Failed to retrieve db_password from Secret Manager"
  exit 1
fi

# Debug: Check password length (without showing actual password)
echo "✅ Password retrieved: ${#DB_PASSWORD} characters"

# URL-encode password untuk menghindari masalah dengan karakter khusus (+, ), dll)
# Python urllib.parse.quote meng-encode karakter khusus dengan benar
# Gunakan stdin untuk menghindari masalah dengan single quote atau karakter khusus lainnya
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Construct DATABASE_URL dengan password yang sudah di-encode
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Debug: Verify DATABASE_URL format (without showing password)
echo "✅ DATABASE_URL length: ${#DATABASE_URL} characters"
echo "✅ Password encoded successfully"

# Start new container with all environment variables
# IMPORTANT: Use --network host so container can access Cloud SQL Proxy on 127.0.0.1:5432
# DO NOT CHANGE network mode - it's required for Cloud SQL Proxy access
echo "🚀 Starting new container..."
echo "   - Network mode: host (required for Cloud SQL Proxy)"
echo "   - Container name: dms-backend-prod"
echo "   - Restart policy: unless-stopped"
sudo docker run -d \
  --name dms-backend-prod \
  --restart unless-stopped \
  --network host \
  -e GCP_PROJECT_ID=${PROJECT_ID} \
  -e GCP_SECRET_MANAGER_ENABLED=false \
  -e GCP_STORAGE_ENABLED=true \
  -e GCP_STORAGE_BUCKET=pedeve-dev-bucket \
  -e DATABASE_URL="${DATABASE_URL}" \
  -e JWT_SECRET="${JWT_SECRET}" \
  -e ENCRYPTION_KEY="${ENCRYPTION_KEY}" \
  -e PORT=8080 \
  -e ENV=production \
  -e DISABLE_RATE_LIMIT=true \
  -e CORS_ORIGIN=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com \
  -e BACKEND_DIR=/app/backend \
  ${BACKEND_IMAGE}

# Wait a moment for container to start
echo "⏳ Waiting for container to start..."
sleep 10

# Verify container is running
echo "🔍 Verifying container status..."
MAX_RETRIES=5
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if sudo docker ps | grep -q dms-backend-prod; then
    echo "✅ Backend container is running"
    
    # Check if container is healthy (listening on port 8080)
    if sudo ss -tlnp | grep -q ':8080'; then
      echo "✅ Backend is listening on port 8080"
      
      # Final health check
      if curl -s -f -m 5 http://127.0.0.1:8080/health > /dev/null 2>&1; then
        echo "✅ Backend health check passed"
        break
      else
        echo "⚠️  Backend is running but health check failed, retrying..."
      fi
    else
      echo "⚠️  WARNING: Backend container is running but port 8080 is not listening yet"
      echo "Container logs:"
      sudo docker logs --tail 20 dms-backend-prod
    fi
  else
    echo "❌ ERROR: Backend container is not running!"
    echo "Container logs:"
    sudo docker logs --tail 50 dms-backend-prod 2>/dev/null || true
    
    # Try to start container again
    echo "🔄 Attempting to restart container..."
    sudo docker start dms-backend-prod 2>/dev/null || {
      echo "❌ Failed to restart container. Checking status..."
      sudo docker ps -a | grep dms-backend-prod
      exit 1
    }
    sleep 5
  fi
  
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
    echo "⏳ Retrying in 5 seconds... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 5
  fi
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo "❌ ERROR: Backend container failed to start after $MAX_RETRIES attempts!"
  echo "Container status:"
  sudo docker ps -a | grep dms-backend-prod || echo "Container not found"
  echo "Container logs:"
  sudo docker logs --tail 100 dms-backend-prod 2>/dev/null || true
  exit 1
fi

# Ensure Docker service is enabled for auto-start
echo "🔧 Ensuring Docker service is enabled..."
sudo systemctl enable docker 2>/dev/null || true
sudo systemctl start docker 2>/dev/null || true

# Verify Docker is running
if ! sudo systemctl is-active --quiet docker; then
  echo "❌ ERROR: Docker service is not running!"
  sudo systemctl status docker --no-pager -l
  exit 1
fi

echo "✅ Backend deployment completed successfully!"
echo "✅ Container restart policy: unless-stopped (will auto-restart on VM reboot)"

