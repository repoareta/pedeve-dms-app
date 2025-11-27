#!/bin/bash

# Script untuk menjalankan company seeder
# Usage: ./seed.sh

set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$( cd "$SCRIPT_DIR/../.." && pwd )"

# Set default DATABASE_URL if not set
if [ -z "$DATABASE_URL" ]; then
    export DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable"
fi

echo "🌱 Running Company Seeder..."
echo "📂 Backend directory: $BACKEND_DIR"
echo "🔗 Database: $DATABASE_URL"
echo ""

cd "$BACKEND_DIR"
go run ./cmd/seed-companies

