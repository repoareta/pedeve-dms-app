#!/bin/bash

# Script untuk test connection seperti DBeaver
# Simulasi connection dari host ke PostgreSQL container

echo "🔍 Testing DBeaver-like Connection..."
echo ""

# Test dengan psql dari host (jika ada)
if command -v psql &> /dev/null; then
    echo "Testing with host psql..."
    PGPASSWORD=dms_password psql -h localhost -p 5432 -U dms_user -d db_pedeve_dms -c "SELECT current_user, current_database();" 2>&1
else
    echo "⚠️  psql not installed on host"
    echo ""
    echo "Testing with docker exec (should work)..."
    docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user, current_database();" 2>&1
fi

echo ""
echo "📋 Connection Info:"
echo "   Host:        localhost"
echo "   Port:        5432"
echo "   Database:    db_pedeve_dms"
echo "   Username:    dms_user"
echo "   Password:    dms_password"
echo ""
echo "💡 If host psql fails but docker exec works:"
echo "   - PostgreSQL container OK"
echo "   - Issue with host connection (pg_hba.conf or network)"
echo "   - DBeaver might have same issue"
echo ""

