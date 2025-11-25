#!/bin/bash

# Script untuk watch file changes dan auto-regenerate Swagger
# Usage: ./scripts/watch-swagger.sh

set -e

echo "=========================================="
echo "Swagger Auto-Regenerate Watcher"
echo "=========================================="
echo ""
echo "Watching for changes in handler files..."
echo "Swagger will be auto-regenerated when handlers change"
echo "Press Ctrl+C to stop"
echo ""

# Install inotify-tools jika belum ada (untuk Linux)
# Untuk macOS, gunakan fswatch
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - gunakan fswatch
    if ! command -v fswatch &> /dev/null; then
        echo "⚠️  fswatch not found. Installing via brew..."
        echo "   Run: brew install fswatch"
        exit 1
    fi
    
    echo "Using fswatch (macOS)..."
    fswatch -o internal/delivery/http/*.go cmd/api/main.go | while read f; do
        echo ""
        echo "🔄 File changed, regenerating Swagger..."
        go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
        echo "✅ Swagger regenerated!"
        echo ""
    done
else
    # Linux - gunakan inotifywait
    if ! command -v inotifywait &> /dev/null; then
        echo "⚠️  inotifywait not found. Installing..."
        sudo apt-get install -y inotify-tools || sudo yum install -y inotify-tools
    fi
    
    echo "Using inotifywait (Linux)..."
    while inotifywait -e modify -r internal/delivery/http/ cmd/api/main.go; do
        echo ""
        echo "🔄 File changed, regenerating Swagger..."
        go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
        echo "✅ Swagger regenerated!"
        echo ""
    done
fi

