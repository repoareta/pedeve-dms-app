#!/bin/bash
set -euo pipefail

# Script untuk deploy backend di VM
# Usage: ./deploy-backend-vm.sh <PROJECT_ID> <BACKEND_IMAGE> [OPTIONS]
# Options (via environment variables):
#   - DB_SECRET_SUFFIX: suffix untuk secret names (default: "", untuk dev: "", untuk prod: "_prod")
#   - DB_NAME: database name (default: "db_dev_pedeve")
#   - DB_USER: database user (default: "pedeve_user_db")
#   - STORAGE_BUCKET: storage bucket name (default: "pedeve-dev-bucket")
#   - CORS_ORIGIN: CORS origin (default: dev domain)
#   - DISABLE_RATE_LIMIT: disable rate limit (default: "true" untuk dev)

PROJECT_ID=$1
BACKEND_IMAGE=$2

# Set defaults untuk development
DB_SECRET_SUFFIX=${DB_SECRET_SUFFIX:-""}
DB_NAME=${DB_NAME:-"db_dev_pedeve"}
DB_USER=${DB_USER:-"pedeve_user_db"}
STORAGE_BUCKET=${STORAGE_BUCKET:-"pedeve-dev-bucket"}
CORS_ORIGIN=${CORS_ORIGIN:-"https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com"}
DISABLE_RATE_LIMIT=${DISABLE_RATE_LIMIT:-"true"}

echo "🚀 Starting backend deployment on VM..."

# Install Docker if not exists
if ! command -v docker &> /dev/null; then
  echo "📦 Installing Docker..."
  curl -fsSL https://get.docker.com -o get-docker.sh
  sudo sh get-docker.sh
  sudo usermod -aG docker $USER || true
  rm -f get-docker.sh
fi

# Check if Cloud SQL Proxy is running (required for database connection)
echo "🔍 Checking Cloud SQL Proxy..."
if ! ps aux | grep -q "[c]loud-sql-proxy"; then
  echo "⚠️  WARNING: Cloud SQL Proxy is not running!"
  echo "   Attempting to start Cloud SQL Proxy service..."
  
  # Try to start Cloud SQL Proxy service if it exists
  if sudo systemctl list-units --type=service | grep -q cloud-sql-proxy; then
    sudo systemctl start cloud-sql-proxy || echo "⚠️  Failed to start Cloud SQL Proxy service"
    sleep 5
  else
    echo "⚠️  Cloud SQL Proxy service not found. Please ensure it's installed and configured."
    echo "   Container will start, but database connection may fail if Cloud SQL Proxy is not running."
  fi
else
  echo "✅ Cloud SQL Proxy is running"
fi

# Load Docker image
echo "🐳 Loading Docker image..."
sudo docker load -i ~/backend-image.tar

# Stop old container
echo "🛑 Stopping old container..."
sudo docker stop dms-backend-prod 2>/dev/null || true
sudo docker rm dms-backend-prod 2>/dev/null || true

# Get secrets from GCP Secret Manager (dengan suffix jika ada)
echo "🔑 Getting secrets from GCP Secret Manager..."
echo "   Using secret suffix: '${DB_SECRET_SUFFIX}'"
echo "   Project: ${PROJECT_ID}"

# Function to get secret with fallback
get_secret() {
  local secret_name=$1
  local secret_name_with_suffix="${secret_name}${DB_SECRET_SUFFIX}"
  local value=""
  
  # Try with suffix first (if suffix is not empty)
  if [ -n "${DB_SECRET_SUFFIX}" ]; then
    echo "   Trying secret: ${secret_name_with_suffix}" >&2
    # Try to get secret, capture both stdout and stderr
    local temp_output=$(mktemp)
    local temp_error=$(mktemp)
    
    if gcloud secrets versions access latest --secret=${secret_name_with_suffix} --project=${PROJECT_ID} >"${temp_output}" 2>"${temp_error}"; then
      if [ -s "${temp_output}" ]; then
        value=$(cat "${temp_output}")
        rm -f "${temp_output}" "${temp_error}"
      else
        rm -f "${temp_output}" "${temp_error}"
        value=""
      fi
    else
      local error_content=$(cat "${temp_error}")
      rm -f "${temp_output}" "${temp_error}"
      value=""
      if echo "${error_content}" | grep -q "was not found\|NOT_FOUND"; then
        echo "   ⚠️  Secret ${secret_name_with_suffix} not found, trying without suffix: ${secret_name}" >&2
      else
        echo "   ⚠️  Error accessing ${secret_name_with_suffix}: ${error_content}" >&2
        echo "   Trying without suffix: ${secret_name}" >&2
      fi
    fi
  fi
  
  # If not found and suffix is not empty, try without suffix as fallback
  # OR if suffix is empty, try without suffix directly
  if [ -z "${value}" ]; then
    echo "   Trying secret: ${secret_name}" >&2
    local temp_output=$(mktemp)
    local temp_error=$(mktemp)
    
    if gcloud secrets versions access latest --secret=${secret_name} --project=${PROJECT_ID} >"${temp_output}" 2>"${temp_error}"; then
      if [ -s "${temp_output}" ]; then
        value=$(cat "${temp_output}")
        rm -f "${temp_output}" "${temp_error}"
      else
        rm -f "${temp_output}" "${temp_error}"
        value=""
      fi
    else
      local error_content=$(cat "${temp_error}")
      rm -f "${temp_output}" "${temp_error}"
      echo "   ❌ Secret ${secret_name_with_suffix} (or ${secret_name}) not found!" >&2
      if [ -n "${error_content}" ]; then
        echo "   Error: ${error_content}" >&2
      fi
      return 1
    fi
  fi
  
  echo "   ✅ Secret retrieved successfully" >&2
  echo "${value}"
  return 0
}

# Get secrets
# Temporarily disable exit on error for command substitution
set +e
DB_PASSWORD=$(get_secret "db_password")
SECRET_EXIT_CODE=$?
set -e

if [ $SECRET_EXIT_CODE -ne 0 ] || [ -z "${DB_PASSWORD}" ]; then
  echo ""
  echo "❌ ERROR: Failed to retrieve db_password from Secret Manager"
  echo ""
  echo "📋 Troubleshooting steps:"
  echo "   1. Verify secret exists in GCP Secret Manager:"
  echo "      gcloud secrets list --project=${PROJECT_ID}"
  echo ""
  echo "   2. If using suffix '${DB_SECRET_SUFFIX}', verify secret name is:"
  if [ -n "${DB_SECRET_SUFFIX}" ]; then
    echo "      - db_password${DB_SECRET_SUFFIX} (with suffix)"
    echo "      - db_password (without suffix - fallback)"
  else
    echo "      - db_password (no suffix)"
  fi
  echo ""
  echo "   3. Verify VM Service Account has Secret Manager Secret Accessor role:"
  echo "      gcloud projects get-iam-policy ${PROJECT_ID} --flatten='bindings[].members' --filter='bindings.members:*@${PROJECT_ID}.iam.gserviceaccount.com'"
  echo ""
  echo "   4. Test secret access manually:"
  if [ -n "${DB_SECRET_SUFFIX}" ]; then
    echo "      gcloud secrets versions access latest --secret=db_password${DB_SECRET_SUFFIX} --project=${PROJECT_ID}"
    echo "      OR:"
  fi
  echo "      gcloud secrets versions access latest --secret=db_password --project=${PROJECT_ID}"
  echo ""
  exit 1
fi

set +e
JWT_SECRET=$(get_secret "jwt_secret")
JWT_EXIT_CODE=$?
set -e

if [ $JWT_EXIT_CODE -ne 0 ] || [ -z "${JWT_SECRET}" ]; then
  echo "⚠️  WARNING: jwt_secret not found, container may fail to start"
  JWT_SECRET=""
fi

set +e
ENCRYPTION_KEY=$(get_secret "encryption_key")
ENCRYPTION_EXIT_CODE=$?
set -e

if [ $ENCRYPTION_EXIT_CODE -ne 0 ] || [ -z "${ENCRYPTION_KEY}" ]; then
  echo "⚠️  WARNING: encryption_key not found, container may fail to start"
  ENCRYPTION_KEY=""
fi

# Debug: Check password length (without showing actual password)
echo "✅ Password retrieved: ${#DB_PASSWORD} characters"

# URL-encode password untuk menghindari masalah dengan karakter khusus (+, ), dll)
# Python urllib.parse.quote meng-encode karakter khusus dengan benar
# Gunakan stdin untuk menghindari masalah dengan single quote atau karakter khusus lainnya
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Construct DATABASE_URL dengan password yang sudah di-encode
# IMPORTANT: Cloud SQL Proxy tidak memerlukan TLS, jadi gunakan sslmode=disable
# Untuk production dengan Private IP atau direct connection, mungkin perlu sslmode=require
# Tapi karena kita pakai Cloud SQL Proxy, selalu gunakan sslmode=disable
SSL_MODE=${SSL_MODE:-"disable"}
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/${DB_NAME}?sslmode=${SSL_MODE}"

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
  -e GCP_STORAGE_BUCKET=${STORAGE_BUCKET} \
  -e DATABASE_URL="${DATABASE_URL}" \
  -e JWT_SECRET="${JWT_SECRET}" \
  -e ENCRYPTION_KEY="${ENCRYPTION_KEY}" \
  -e PORT=8080 \
  -e ENV=production \
  -e DISABLE_RATE_LIMIT=${DISABLE_RATE_LIMIT} \
  -e CORS_ORIGIN="${CORS_ORIGIN}" \
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

