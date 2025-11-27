#!/bin/bash

# Script untuk deploy frontend ke GCP VM
# Usage: ./scripts/deploy-frontend.sh [image-tag]

set -e

IMAGE_TAG=${1:-latest}
REPO_OWNER=${GITHUB_REPOSITORY_OWNER:-$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\).*/\1/')}
FRONTEND_IMAGE="ghcr.io/${REPO_OWNER}/dms-frontend:${IMAGE_TAG}"
VM_NAME="frontend-dev"
VM_ZONE="asia-southeast2-a"
PROJECT_ID=${GCP_PROJECT_ID:-"pedeve-pertamina-dms"}
SSH_USER=${GCP_SSH_USER:-"info@aretaamany.com"}

echo "🚀 Deploying Frontend to GCP"
echo "   Image: ${FRONTEND_IMAGE}"
echo "   VM: ${VM_NAME}"
echo "   Zone: ${VM_ZONE}"
echo ""

# Pull image dari GHCR
echo "📥 Pulling Docker image..."
docker pull ${FRONTEND_IMAGE}

# Save image to tar
echo "💾 Saving Docker image to tar..."
docker save ${FRONTEND_IMAGE} -o /tmp/frontend-image.tar

# Copy to VM
echo "📤 Copying image to VM..."
gcloud compute scp /tmp/frontend-image.tar ${SSH_USER}@${VM_NAME}:~/frontend-image.tar \
  --zone=${VM_ZONE} \
  --project=${PROJECT_ID}

# Deploy on VM
echo "🔧 Deploying on VM..."
gcloud compute ssh ${SSH_USER}@${VM_NAME} \
  --zone=${VM_ZONE} \
  --project=${PROJECT_ID} \
  --command="
    # Load Docker image
    docker load -i ~/frontend-image.tar
    
    # Stop old container
    docker stop dms-frontend-prod 2>/dev/null || true
    docker rm dms-frontend-prod 2>/dev/null || true
    
    # Start new container
    docker run -d \
      --name dms-frontend-prod \
      --restart unless-stopped \
      -p 80:80 \
      ${FRONTEND_IMAGE}
    
    # Cleanup
    rm -f ~/frontend-image.tar
    docker image prune -f
  "

# Cleanup local tar
rm -f /tmp/frontend-image.tar

echo ""
echo "✅ Frontend deployed successfully!"
echo "   URL: https://pedeve-dev.aretaamany.com"

