#!/bin/bash

# DMS App - Development Script
# Simple command to run all services

set -e

echo "🚀 Starting DMS App Development Environment..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
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
echo "📦 Starting all services..."
docker-compose -f docker-compose.dev.yml up --build

