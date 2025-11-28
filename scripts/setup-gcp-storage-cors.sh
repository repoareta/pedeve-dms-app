#!/bin/bash
# Script untuk setup CORS di GCP Storage bucket
# Usage: ./setup-gcp-storage-cors.sh <PROJECT_ID> <BUCKET_NAME>

PROJECT_ID=$1
BUCKET_NAME=$2

if [ -z "$PROJECT_ID" ] || [ -z "$BUCKET_NAME" ]; then
  echo "Usage: ./setup-gcp-storage-cors.sh <PROJECT_ID> <BUCKET_NAME>"
  exit 1
fi

echo "🔧 Setting up CORS for GCP Storage bucket: $BUCKET_NAME"

# Create CORS config file
cat > /tmp/cors-config.json <<EOF
[
  {
    "origin": [
      "https://pedeve-dev.aretaamany.com",
      "http://pedeve-dev.aretaamany.com",
      "http://34.128.123.1",
      "https://pedeve-dev.aretaamany.com:443",
      "http://localhost:5173",
      "http://localhost:8080"
    ],
    "method": ["GET", "HEAD", "OPTIONS"],
    "responseHeader": [
      "Content-Type",
      "Access-Control-Allow-Origin",
      "Access-Control-Allow-Methods",
      "Access-Control-Allow-Headers"
    ],
    "maxAgeSeconds": 3600
  }
]
EOF

# Apply CORS configuration
echo "📤 Applying CORS configuration..."
gcloud storage buckets update gs://${BUCKET_NAME} \
  --cors-file=/tmp/cors-config.json \
  --project=${PROJECT_ID}

if [ $? -eq 0 ]; then
  echo "✅ CORS configuration applied successfully!"
  echo ""
  echo "📋 CORS Configuration:"
  cat /tmp/cors-config.json
  echo ""
  echo "🧪 Test: Try accessing an image from frontend now"
else
  echo "❌ Failed to apply CORS configuration"
  exit 1
fi

# Cleanup
rm -f /tmp/cors-config.json

echo "✅ CORS setup completed!"

