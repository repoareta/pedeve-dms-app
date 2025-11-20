#!/bin/bash

# Script untuk fix DBeaver connection issue
# Masalah: "FATAL: role 'dms_user' does not exist"

echo "🔧 Fixing DBeaver Connection Issue..."
echo ""

# Cek apakah postgres container running
if ! docker ps | grep -q dms-postgres-dev; then
    echo "❌ PostgreSQL container tidak running"
    exit 1
fi

echo "📊 Checking PostgreSQL configuration..."
echo ""

# Cek user dari dalam container
echo "1. Checking user from inside container..."
USER_EXISTS=$(docker exec dms-postgres-dev psql -U dms_user -d postgres -t -c "SELECT 1 FROM pg_roles WHERE rolname = 'dms_user';" 2>&1 | tr -d '[:space:]')

if [ "$USER_EXISTS" = "1" ]; then
    echo "   ✅ User 'dms_user' exists inside container"
else
    echo "   ❌ User 'dms_user' NOT found inside container"
    echo ""
    echo "   🔄 Recreating PostgreSQL container..."
    docker-compose -f docker-compose.dev.yml down -v postgres
    docker-compose -f docker-compose.dev.yml up -d postgres
    sleep 10
    echo "   ✅ PostgreSQL recreated"
fi

echo ""
echo "2. Checking database..."
DB_EXISTS=$(docker exec dms-postgres-dev psql -U dms_user -d postgres -t -c "SELECT 1 FROM pg_database WHERE datname = 'db_pedeve_dms';" 2>&1 | tr -d '[:space:]')

if [ "$DB_EXISTS" = "1" ]; then
    echo "   ✅ Database 'db_pedeve_dms' exists"
else
    echo "   ⚠️  Database 'db_pedeve_dms' not found, creating..."
    docker exec dms-postgres-dev psql -U dms_user -d postgres -c "CREATE DATABASE db_pedeve_dms;" 2>&1
    echo "   ✅ Database created"
fi

echo ""
echo "3. Testing connection from inside container..."
TEST_CONN=$(docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user, current_database();" 2>&1 | grep -i "dms_user" | head -1)

if [ -n "$TEST_CONN" ]; then
    echo "   ✅ Connection from inside container: OK"
else
    echo "   ❌ Connection from inside container: FAILED"
fi

echo ""
echo "4. Checking PostgreSQL port accessibility..."
PORT_CHECK=$(docker port dms-postgres-dev 2>&1 | grep 5432)

if [ -n "$PORT_CHECK" ]; then
    echo "   ✅ Port 5432 is exposed: $PORT_CHECK"
else
    echo "   ❌ Port 5432 not exposed"
fi

echo ""
echo "5. Verifying user permissions..."
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "\du dms_user" 2>&1 | grep dms_user

echo ""
echo "📋 DBeaver Connection Settings:"
echo ""
echo "   Host:        localhost"
echo "   Port:        5432"
echo "   Database:    db_pedeve_dms"
echo "   Username:    dms_user"
echo "   Password:    dms_password"
echo ""
echo "   Connection URL:"
echo "   jdbc:postgresql://localhost:5432/db_pedeve_dms"
echo ""

echo "💡 If connection still fails:"
echo "   1. Restart PostgreSQL: docker-compose -f docker-compose.dev.yml restart postgres"
echo "   2. Wait 10 seconds"
echo "   3. Try DBeaver connection again"
echo "   4. If still fails, recreate PostgreSQL:"
echo "      docker-compose -f docker-compose.dev.yml down -v postgres"
echo "      docker-compose -f docker-compose.dev.yml up -d postgres"
echo ""

echo "✅ Diagnostic complete!"

