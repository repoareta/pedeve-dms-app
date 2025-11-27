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

# Construct DATABASE_URL
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Debug: Verify DATABASE_URL format (without showing password)
echo "✅ DATABASE_URL length: ${#DATABASE_URL} characters"

# Start new container with all environment variables
# Use --network host so container can access Cloud SQL Proxy on 127.0.0.1:5432
echo "🚀 Starting new container..."
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
  -e CORS_ORIGIN=https://pedeve-dev.aretaamany.com \
  ${BACKEND_IMAGE}

echo "✅ Backend container started successfully!"

