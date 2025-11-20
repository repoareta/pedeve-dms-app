#!/bin/bash

# Script untuk rename database dari dms_db ke db_pedeve_dms
# atau membuat database baru jika belum ada

echo "🔄 Renaming PostgreSQL Database..."
echo ""

# Cek apakah postgres container running
if ! docker ps | grep -q dms-postgres-dev; then
    echo "❌ PostgreSQL container tidak running"
    echo "   Start dengan: docker-compose -f docker-compose.dev.yml up -d postgres"
    exit 1
fi

echo "📊 Checking current database..."
EXISTING_DB=$(docker exec dms-postgres-dev psql -U dms_user -t -c "SELECT datname FROM pg_database WHERE datname = 'dms_db';" 2>/dev/null | tr -d '[:space:]')

if [ "$EXISTING_DB" = "dms_db" ]; then
    echo "✅ Found database: dms_db"
    echo ""
    echo "⚠️  Option 1: Rename existing database (preserves data)"
    echo "⚠️  Option 2: Create new database (fresh start)"
    echo ""
    read -p "Pilih opsi (1=rename, 2=new): " choice
    
    if [ "$choice" = "1" ]; then
        echo ""
        echo "🔄 Renaming database from dms_db to db_pedeve_dms..."
        
        # Rename database (need to disconnect all connections first)
        docker exec dms-postgres-dev psql -U dms_user -d postgres <<EOF
-- Terminate all connections to dms_db
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'dms_db' AND pid <> pg_backend_pid();

-- Rename database
ALTER DATABASE dms_db RENAME TO db_pedeve_dms;
EOF
        
        if [ $? -eq 0 ]; then
            echo "✅ Database renamed successfully!"
        else
            echo "❌ Failed to rename database"
            exit 1
        fi
    else
        echo ""
        echo "📦 Creating new database: db_pedeve_dms..."
        docker exec dms-postgres-dev psql -U dms_user -d postgres -c "CREATE DATABASE db_pedeve_dms;" 2>/dev/null
        
        if [ $? -eq 0 ]; then
            echo "✅ New database created!"
            echo ""
            echo "💡 Old database 'dms_db' still exists."
            echo "   If you want to migrate data, see POSTGRESQL_MIGRATION.md"
        else
            echo "❌ Failed to create database (might already exist)"
        fi
    fi
else
    echo "📦 Database 'dms_db' not found. Creating new database: db_pedeve_dms..."
    docker exec dms-postgres-dev psql -U dms_user -d postgres -c "CREATE DATABASE db_pedeve_dms;" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo "✅ Database created successfully!"
    else
        echo "❌ Failed to create database"
        exit 1
    fi
fi

echo ""
echo "🔄 Restarting backend to connect to new database..."
docker-compose -f docker-compose.dev.yml restart backend

echo ""
echo "⏳ Waiting for backend to connect..."
sleep 5

echo ""
echo "✅ Done!"
echo ""
echo "📋 Verify database:"
echo "   docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c '\\dt'"
echo ""
echo "📋 Check backend logs:"
echo "   docker-compose -f docker-compose.dev.yml logs backend | grep -i 'postgres\|database'"
echo ""

