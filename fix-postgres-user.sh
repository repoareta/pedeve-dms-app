#!/bin/bash

# Script untuk fix/create user dms_user di PostgreSQL
# Jika user tidak ada atau ada masalah

echo "🔧 Fixing PostgreSQL User..."
echo ""

# Cek apakah postgres container running
if ! docker ps | grep -q dms-postgres-dev; then
    echo "❌ PostgreSQL container tidak running"
    echo "   Start dengan: docker-compose -f docker-compose.dev.yml up -d postgres"
    exit 1
fi

# Cek apakah bisa connect dengan dms_user
echo "📊 Checking current connection..."
CONNECTION_TEST=$(docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user;" 2>&1 | grep -i "FATAL\|error" | head -1)

if [ -n "$CONNECTION_TEST" ]; then
    echo "⚠️  Connection issue detected: $CONNECTION_TEST"
    echo ""
    echo "💡 User 'dms_user' mungkin belum ada atau tidak punya akses."
    echo ""
    echo "🔧 Solusi:"
    echo "1. Stop PostgreSQL container"
    echo "2. Hapus volume PostgreSQL (akan hilang data!)"
    echo "3. Start ulang PostgreSQL (akan create user baru)"
    echo ""
    read -p "Apakah Anda ingin menghapus volume dan restart PostgreSQL? (y/n): " confirm
    
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        echo ""
        echo "🛑 Stopping PostgreSQL..."
        docker-compose -f docker-compose.dev.yml stop postgres
        
        echo "🗑️  Removing PostgreSQL volume..."
        docker-compose -f docker-compose.dev.yml down -v postgres
        
        echo "🔄 Starting PostgreSQL dengan konfigurasi baru..."
        docker-compose -f docker-compose.dev.yml up -d postgres
        
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
        
        echo ""
        echo "✅ PostgreSQL restarted with new user!"
        echo ""
        echo "📋 Test connection:"
        echo "   docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c 'SELECT current_user;'"
    else
        echo "❌ Cancelled"
        exit 1
    fi
else
    echo "✅ Connection OK! User 'dms_user' exists."
    echo ""
    docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "\du" 2>&1
fi

echo ""
echo "✅ Done!"

