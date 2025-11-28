#!/bin/bash
set -euo pipefail

# Script untuk make folder logos public di GCP Storage
# Usage: ./make-logos-public.sh <PROJECT_ID> <BUCKET_NAME>

PROJECT_ID=${1:-pedeve-pertamina-dms}
BUCKET_NAME=${2:-pedeve-dev-bucket}

echo "🔓 Making logos folder public in gs://${BUCKET_NAME}..."

# Make all objects in logos folder public
gsutil -m acl ch -u AllUsers:R gs://${BUCKET_NAME}/logos/*

# Also make future uploads to logos folder public by default
# (This sets default ACL for new objects in logos folder)
gsutil defacl ch -u AllUsers:R gs://${BUCKET_NAME}/logos/

echo "✅ Logos folder is now public!"
echo "📋 Test dengan:"
echo "   curl https://storage.googleapis.com/${BUCKET_NAME}/logos/<filename>"

