#!/bin/bash

# Script untuk auto-regenerate Swagger saat development
# Bisa dijalankan di background atau terminal terpisah
# Usage: ./scripts/auto-swagger.sh

set -e

echo "🚀 Starting Swagger Auto-Regenerate..."
echo ""
echo "This script will:"
echo "  1. Watch for changes in handler files"
echo "  2. Auto-regenerate Swagger docs"
echo "  3. Swagger UI will auto-reload (no need to refresh)"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Function to regenerate swagger
regenerate_swagger() {
    echo ""
    echo "🔄 Changes detected, regenerating Swagger..."
    go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs 2>&1 | grep -E "(Generate|warning|Error|create)" || true
    echo "✅ Swagger regenerated! Swagger UI will auto-reload."
    echo ""
}

# Check OS and use appropriate method
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - try fswatch first, fallback to polling
    if command -v fswatch &> /dev/null; then
        echo "✅ Using fswatch (macOS)"
        fswatch -o internal/delivery/http/*.go cmd/api/main.go | while read f; do
            regenerate_swagger
        done
    else
        echo "⚠️  fswatch not found, using polling method (checks every 2 seconds)"
        echo "   Install fswatch for better performance: brew install fswatch"
        echo ""
        
        LAST_MODIFIED=0
        while true; do
            CURRENT_MODIFIED=$(find internal/delivery/http/*.go cmd/api/main.go -type f -exec stat -f "%m" {} \; 2>/dev/null | sort -n | tail -1)
            
            if [ "$CURRENT_MODIFIED" != "$LAST_MODIFIED" ] && [ "$CURRENT_MODIFIED" != "" ]; then
                if [ "$LAST_MODIFIED" != "0" ]; then
                    regenerate_swagger
                fi
                LAST_MODIFIED=$CURRENT_MODIFIED
            fi
            
            sleep 2
        done
    fi
else
    # Linux - try inotifywait first, fallback to polling
    if command -v inotifywait &> /dev/null; then
        echo "✅ Using inotifywait (Linux)"
        while inotifywait -e modify -r internal/delivery/http/ cmd/api/main.go 2>/dev/null; do
            regenerate_swagger
        done
    else
        echo "⚠️  inotifywait not found, using polling method (checks every 2 seconds)"
        echo "   Install inotifywait for better performance: sudo apt-get install inotify-tools"
        echo ""
        
        LAST_MODIFIED=0
        while true; do
            CURRENT_MODIFIED=$(find internal/delivery/http/*.go cmd/api/main.go -type f -exec stat -c "%Y" {} \; 2>/dev/null | sort -n | tail -1)
            
            if [ "$CURRENT_MODIFIED" != "$LAST_MODIFIED" ] && [ "$CURRENT_MODIFIED" != "" ]; then
                if [ "$LAST_MODIFIED" != "0" ]; then
                    regenerate_swagger
                fi
                LAST_MODIFIED=$CURRENT_MODIFIED
            fi
            
            sleep 2
        done
    fi
fi
