#!/bin/bash

# Script untuk menjalankan company seeder
# Usage: ./seed.sh

set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$( cd "$SCRIPT_DIR/../.." && pwd )"

# DATABASE_URL must be set via environment variable for security
# Never hardcode database credentials in source code
if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL environment variable is required." >&2
    echo "   Please set it before running this script." >&2
    echo "   Example: export DATABASE_URL='postgres://user:password@host:port/dbname?sslmode=disable'" >&2
    exit 1
fi

echo "🌱 Running Company Seeder..."
echo "📂 Backend directory: $BACKEND_DIR"
echo "🔗 Database: $DATABASE_URL"
echo ""

cd "$BACKEND_DIR"
go run ./cmd/seed-companies

