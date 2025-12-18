#!/bin/bash

# Script untuk deploy backend ke GCP VM
# Usage: ./scripts/deploy-backend.sh [image-tag]

set -e

IMAGE_TAG=${1:-latest}
REPO_OWNER=${GITHUB_REPOSITORY_OWNER:-$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\).*/\1/')}
BACKEND_IMAGE="ghcr.io/${REPO_OWNER}/dms-backend:${IMAGE_TAG}"
VM_NAME="backend-dev"
VM_ZONE="asia-southeast2-a"
PROJECT_ID=${GCP_PROJECT_ID:-"pedeve-pertamina-dms"}
SSH_USER=${GCP_SSH_USER:-"info@aretaamany.com"}

echo "🚀 Deploying Backend to GCP"
echo "   Image: ${BACKEND_IMAGE}"
echo "   VM: ${VM_NAME}"
echo "   Zone: ${VM_ZONE}"
echo ""

# Pull image dari GHCR
echo "📥 Pulling Docker image..."
docker pull ${BACKEND_IMAGE}

# Save image to tar
echo "💾 Saving Docker image to tar..."
docker save ${BACKEND_IMAGE} -o /tmp/backend-image.tar

# Copy to VM
echo "📤 Copying image to VM..."
gcloud compute scp /tmp/backend-image.tar ${SSH_USER}@${VM_NAME}:~/backend-image.tar \
  --zone=${VM_ZONE} \
  --project=${PROJECT_ID}

# Deploy on VM
echo "🔧 Deploying on VM..."
gcloud compute ssh ${SSH_USER}@${VM_NAME} \
  --zone=${VM_ZONE} \
  --project=${PROJECT_ID} \
  --command="
    # Load Docker image
    docker load -i ~/backend-image.tar
    
    # Stop old container
    docker stop dms-backend-prod 2>/dev/null || true
    docker rm dms-backend-prod 2>/dev/null || true
    
    # Get secrets from GCP Secret Manager
    DB_PASSWORD=\$(gcloud secrets versions access latest --secret=db_password --project=${PROJECT_ID} 2>/dev/null || echo '')
    JWT_SECRET=\$(gcloud secrets versions access latest --secret=jwt_secret --project=${PROJECT_ID} 2>/dev/null || echo '')
    ENCRYPTION_KEY=\$(gcloud secrets versions access latest --secret=encryption_key --project=${PROJECT_ID} 2>/dev/null || echo '')
    
    # Start new container
    docker run -d \
      --name dms-backend-prod \
      --restart unless-stopped \
      -p 8080:8080 \
      -e GCP_PROJECT_ID=${PROJECT_ID} \
      -e GCP_SECRET_MANAGER_ENABLED=true \
      -e GCP_STORAGE_ENABLED=true \
      -e GCP_STORAGE_BUCKET=pedeve-dev-bucket \
      -e DATABASE_URL=postgres://pedeve_user_db:\${DB_PASSWORD}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable \
      -e PORT=8080 \
      -e ENV=production \
      -e CORS_ORIGIN=https://pedeve-dev.aretaamany.com \
      ${BACKEND_IMAGE}
    
    # Cleanup
    rm -f ~/backend-image.tar
    docker image prune -f
  "

# Cleanup local tar
rm -f /tmp/backend-image.tar

echo ""
echo "✅ Backend deployed successfully!"
echo "   Health check: https://api-pedeve-dev.aretaamany.com/health"

