#!/bin/bash

# Script to build VitePress docs and copy to frontend/public

set -e

echo "📚 Building VitePress documentation..."

# Navigate to docs directory
cd "$(dirname "$0")/../docs"

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
  echo "📦 Installing VitePress dependencies..."
  npm install
fi

# Build VitePress
echo "🔨 Building documentation..."
npm run build

# Copy build output to frontend/public/user-guideline
echo "📋 Copying build output to frontend/public/user-guideline..."
cd ..
rm -rf frontend/public/user-guideline
mkdir -p frontend/public/user-guideline
cp -r docs/.vitepress/dist/* frontend/public/user-guideline/

echo "✅ Documentation built and copied successfully!"
echo "📍 Access at: http://localhost:5173/user-guideline/"
