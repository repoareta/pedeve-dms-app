#!/bin/bash

# DMS App - Development Script
# Simple command to run all services
# Usage: ./dev.sh [--background|-d]

set -e

# Parse arguments
BACKGROUND_MODE=false
if [[ "$1" == "--background" ]] || [[ "$1" == "-d" ]]; then
    BACKGROUND_MODE=true
fi

echo "🚀 Starting DMS App Development Environment..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose.dev.yml exists
if [ ! -f "docker-compose.dev.yml" ]; then
    echo "❌ Error: docker-compose.dev.yml not found!"
    echo "   Please run this script from the project root directory."
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    docker-compose -f docker-compose.dev.yml down
    exit 0
}

trap cleanup SIGINT SIGTERM

# Start services
if [ "$BACKGROUND_MODE" = true ]; then
    echo "📦 Starting all services in background..."
    docker-compose -f docker-compose.dev.yml up -d --build
    echo ""
    echo "✅ Services started in background!"
    echo "   Use 'docker-compose -f docker-compose.dev.yml logs -f' to view logs"
    echo "   Use 'docker-compose -f docker-compose.dev.yml down' to stop"
else
    echo "📦 Starting all services..."
    docker-compose -f docker-compose.dev.yml up --build
fi

