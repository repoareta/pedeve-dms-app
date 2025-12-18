#!/bin/bash

# Script untuk generate document folders untuk semua perusahaan yang sudah ada

cd "$(dirname "$0")/../.."

echo "🚀 Running Generate Company Folders..."
echo ""

go run cmd/generate-company-folders/main.go

echo ""
echo "✅ Done!"
