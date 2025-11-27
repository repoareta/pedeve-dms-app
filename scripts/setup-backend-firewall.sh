#!/bin/bash
set -euo pipefail

# Script untuk setup firewall rules untuk backend
# Usage: ./setup-backend-firewall.sh <PROJECT_ID>

PROJECT_ID=${1:-pedeve-pertamina-dms}

echo "🔥 Setting up firewall rules for backend..."

# Allow HTTP (port 80) - untuk frontend
echo "📡 Creating firewall rule for HTTP (port 80)..."
gcloud compute firewall-rules create allow-http \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server \
  --description "Allow HTTP traffic" \
  --project ${PROJECT_ID} 2>/dev/null || echo "   ⚠️  Rule 'allow-http' already exists"

# Allow HTTPS (port 443) - untuk frontend dengan SSL
echo "📡 Creating firewall rule for HTTPS (port 443)..."
gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic" \
  --project ${PROJECT_ID} 2>/dev/null || echo "   ⚠️  Rule 'allow-https' already exists"

# Allow Backend API (port 8080) - untuk backend API
echo "📡 Creating firewall rule for Backend API (port 8080)..."
gcloud compute firewall-rules create allow-backend-api \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags backend-api-server \
  --description "Allow Backend API traffic on port 8080" \
  --project ${PROJECT_ID} 2>/dev/null || echo "   ⚠️  Rule 'allow-backend-api' already exists"

# Apply tags to frontend VM
echo "🏷️  Applying tags to frontend VM..."
gcloud compute instances add-tags frontend-dev \
  --tags http-server,https-server \
  --zone asia-southeast2-a \
  --project ${PROJECT_ID} 2>/dev/null || echo "   ⚠️  Tags already applied or VM not found"

# Apply tags to backend VM
echo "🏷️  Applying tags to backend VM..."
gcloud compute instances add-tags backend-dev \
  --tags backend-api-server \
  --zone asia-southeast2-a \
  --project ${PROJECT_ID} 2>/dev/null || echo "   ⚠️  Tags already applied or VM not found"

echo ""
echo "✅ Firewall rules setup completed!"
echo ""
echo "📋 Summary:"
echo "   - HTTP (port 80): ✅ Allowed"
echo "   - HTTPS (port 443): ✅ Allowed"
echo "   - Backend API (port 8080): ✅ Allowed"
echo ""
echo "🧪 Test commands:"
echo "   curl http://34.101.49.147:8080/health"
echo "   curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token"

