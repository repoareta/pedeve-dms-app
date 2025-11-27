#!/bin/bash
set -euo pipefail

# Script untuk test koneksi database dari VM
# Usage: ./test-db-connection.sh <PROJECT_ID>

PROJECT_ID=$1

echo "🔍 Testing database connection from VM..."

# Get password from Secret Manager
echo "📥 Getting password from GCP Secret Manager..."
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${PROJECT_ID} 2>/dev/null || echo '')

if [ -z "${DB_PASSWORD}" ]; then
  echo "❌ ERROR: Failed to retrieve db_password from Secret Manager"
  exit 1
fi

echo "✅ Password retrieved: ${#DB_PASSWORD} characters"

# Show first and last character (for verification, not full password)
FIRST_CHAR="${DB_PASSWORD:0:1}"
LAST_CHAR="${DB_PASSWORD: -1}"
echo "   First char: ${FIRST_CHAR}, Last char: ${LAST_CHAR}"

# URL-encode password
echo "🔐 URL-encoding password..."
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")
echo "   Encoded length: ${#DB_PASSWORD_ENCODED} characters"

# Construct DATABASE_URL
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"
echo "✅ DATABASE_URL constructed"

# Test connection using psql (if available) or Python
echo ""
echo "🧪 Testing connection..."

# Try with psql first
if command -v psql &> /dev/null; then
  echo "   Using psql..."
  export PGPASSWORD="${DB_PASSWORD}"
  if psql -h 127.0.0.1 -p 5432 -U pedeve_user_db -d db_dev_pedeve -c "SELECT version();" 2>&1; then
    echo "✅ Connection successful with psql!"
  else
    echo "❌ Connection failed with psql"
  fi
  unset PGPASSWORD
fi

# Try with Python psycopg2
if python3 -c "import psycopg2" 2>/dev/null; then
  echo "   Using Python psycopg2..."
  python3 << EOF
import urllib.parse
import psycopg2

password = "${DB_PASSWORD}"
password_encoded = urllib.parse.quote(password, safe='')
conn_string = f"postgresql://pedeve_user_db:{password_encoded}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

try:
    conn = psycopg2.connect(conn_string)
    print("✅ Connection successful with Python!")
    conn.close()
except Exception as e:
    print(f"❌ Connection failed: {e}")
EOF
else
  echo "   psycopg2 not available, skipping Python test"
fi

echo ""
echo "📋 Summary:"
echo "   - Password retrieved from Secret Manager: ✅"
echo "   - Password length: ${#DB_PASSWORD} characters"
echo "   - URL-encoded length: ${#DB_PASSWORD_ENCODED} characters"
echo ""
echo "💡 If connection fails, please verify:"
echo "   1. Password in GCP Secret Manager matches password in Cloud SQL"
echo "   2. User 'pedeve_user_db' exists in Cloud SQL"
echo "   3. Cloud SQL Proxy is running on 127.0.0.1:5432"
echo "   4. User has proper permissions on database 'db_dev_pedeve'"

