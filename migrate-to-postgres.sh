#!/bin/bash

# Script untuk migrasi dari SQLite ke PostgreSQL
# Menjalankan proses migrasi database

echo "🚀 Starting PostgreSQL Migration..."
echo ""

# Stop services jika sedang running
echo "📦 Stopping existing services..."
docker-compose -f docker-compose.dev.yml down 2>/dev/null

# Start PostgreSQL service dulu
echo "🐘 Starting PostgreSQL..."
docker-compose -f docker-compose.dev.yml up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

MAX_ATTEMPTS=30
ATTEMPT=0
while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  if docker exec dms-postgres-dev pg_isready -U dms_user > /dev/null 2>&1; then
    echo "✅ PostgreSQL is ready!"
    break
  fi
  ATTEMPT=$((ATTEMPT + 1))
  echo "  Attempt $ATTEMPT/$MAX_ATTEMPTS..."
  sleep 2
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
  echo "❌ PostgreSQL failed to start"
  exit 1
fi

# Check if SQLite database exists and has data
SQLITE_DB="./backend/dms.db"
if [ -f "$SQLITE_DB" ]; then
  echo ""
  echo "📊 SQLite database found: $SQLITE_DB"
  echo "💡 Backend will auto-migrate schema to PostgreSQL"
  echo "💡 If you have data in SQLite, you may need to export it manually"
  echo ""
  echo "To migrate existing data:"
  echo "1. Export data from SQLite (see POSTGRESQL_MIGRATION.md)"
  echo "2. Import to PostgreSQL after backend starts"
  echo ""
fi

# Start backend (will auto-connect to PostgreSQL and migrate schema)
echo "🔄 Starting backend with PostgreSQL..."
docker-compose -f docker-compose.dev.yml up -d backend

# Wait a bit for backend to initialize
sleep 3

# Start frontend
echo "🎨 Starting frontend..."
docker-compose -f docker-compose.dev.yml up -d frontend

echo ""
echo "✅ Migration process started!"
echo ""
echo "📋 Services status:"
docker-compose -f docker-compose.dev.yml ps

echo ""
echo "📝 Check backend logs to verify PostgreSQL connection:"
echo "   docker-compose -f docker-compose.dev.yml logs backend | grep -i postgres"
echo ""
echo "📝 View all logs:"
echo "   docker-compose -f docker-compose.dev.yml logs -f"
echo ""
echo "🔍 Test connection:"
echo "   docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c '\\dt'"
echo ""

