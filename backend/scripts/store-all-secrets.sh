#!/bin/bash

# Script untuk store semua secrets ke Vault
# Usage: ./scripts/store-all-secrets.sh

set -e

echo "=========================================="
echo "Store All Secrets to Vault"
echo "=========================================="
echo ""

# Vault configuration
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-dev-root-token-12345}"
VAULT_SECRET_PATH="${VAULT_SECRET_PATH:-secret/dms-app}"

export VAULT_ADDR
export VAULT_TOKEN

echo "Vault Configuration:"
echo "  Address: $VAULT_ADDR"
echo "  Path: $VAULT_SECRET_PATH"
echo ""

# Check if running in Docker
if [ -f /.dockerenv ] || [ -n "$DOCKER_CONTAINER" ]; then
    VAULT_CMD="vault"
else
    if docker ps | grep -q dms-vault-dev; then
        VAULT_CMD="docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=$VAULT_TOKEN dms-vault-dev vault"
        export VAULT_ADDR="http://127.0.0.1:8200"
    elif command -v vault &> /dev/null; then
        VAULT_CMD="vault"
    else
        echo "❌ Vault container not running and vault CLI not found!"
        exit 1
    fi
fi

# Secrets configuration (dengan default values)
ENCRYPTION_KEY="${ENCRYPTION_KEY:-default-encryption-key-32-chars!}"
JWT_SECRET="${JWT_SECRET:-your-secret-key-change-in-production-min-32-chars}"
DATABASE_URL="${DATABASE_URL:-postgres://postgres:dms_password@postgres:5432/db_dms_pedeve?sslmode=disable}"
DATABASE_PASSWORD="${DATABASE_PASSWORD:-dms_password}"
CSRF_SECRET="${CSRF_SECRET:-csrf-secret-key-for-token-generation-32!}"
SUPERADMIN_PASSWORD="${SUPERADMIN_PASSWORD:-Pedeve123}"

# SonarCloud configuration (optional, bisa di-set via environment)
SONARCLOUD_URL="${SONARCLOUD_URL:-https://sonarcloud.io}"
SONARCLOUD_TOKEN="${SONARCLOUD_TOKEN:-}"
SONARCLOUD_PROJECT_KEY="${SONARCLOUD_PROJECT_KEY:-}"

# Rate limit config (default)
RATE_LIMIT_CONFIG='{"general":{"rps":500,"burst":500},"auth":{"rpm":5,"burst":5},"strict":{"rpm":50,"burst":50}}'

echo "Secrets to store:"
echo "  - encryption_key: ${ENCRYPTION_KEY:0:10}... (32 bytes)"
echo "  - jwt_secret: ${JWT_SECRET:0:10}... (min 32 chars)"
echo "  - database_url: ${DATABASE_URL:0:30}..."
echo "  - database_password: ${DATABASE_PASSWORD:0:5}..."
echo "  - csrf_secret: ${CSRF_SECRET:0:10}... (32 bytes)"
echo "  - superadmin_password: ${SUPERADMIN_PASSWORD:0:5}... (min 8 chars)"
echo "  - rate_limit: (JSON config)"
if [ -n "$SONARCLOUD_TOKEN" ]; then
    echo "  - SONARCLOUD_URL: $SONARCLOUD_URL"
    echo "  - SONARCLOUD_TOKEN: ${SONARCLOUD_TOKEN:0:10}... (hidden)"
    echo "  - SONARCLOUD_PROJECT_KEY: $SONARCLOUD_PROJECT_KEY"
else
    echo "  - SonarCloud secrets: (not set, skipping)"
fi
echo ""

# Store all secrets in one command
echo "Storing all secrets to Vault..."
echo "  Path: $VAULT_SECRET_PATH"
echo ""

# Store all secrets - KV v2 format uses /data/ in path
# Vault kv put automatically handles KV v2 format when path contains /data/
VAULT_PUT_CMD="$VAULT_CMD kv put \"$VAULT_SECRET_PATH\" \
    encryption_key=\"$ENCRYPTION_KEY\" \
    jwt_secret=\"$JWT_SECRET\" \
    database_url=\"$DATABASE_URL\" \
    database_password=\"$DATABASE_PASSWORD\" \
    csrf_secret=\"$CSRF_SECRET\" \
    superadmin_password=\"$SUPERADMIN_PASSWORD\" \
    rate_limit=\"$RATE_LIMIT_CONFIG\""

# Add SonarCloud secrets if provided
if [ -n "$SONARCLOUD_TOKEN" ] && [ -n "$SONARCLOUD_PROJECT_KEY" ]; then
    VAULT_PUT_CMD="$VAULT_PUT_CMD \
    SONARCLOUD_URL=\"$SONARCLOUD_URL\" \
    SONARCLOUD_TOKEN=\"$SONARCLOUD_TOKEN\" \
    SONARCLOUD_PROJECT_KEY=\"$SONARCLOUD_PROJECT_KEY\""
fi

eval $VAULT_PUT_CMD 2>&1

if [ $? -eq 0 ]; then
    echo "✅ All secrets stored successfully"
else
    echo "❌ Failed to store secrets"
    exit 1
fi

echo ""
echo "Verifying stored secrets..."
$VAULT_CMD kv get "$VAULT_SECRET_PATH" 2>&1 | head -30

echo ""
echo "=========================================="
echo "✅ All Secrets Stored!"
echo "=========================================="
echo ""
echo "Stored secrets:"
echo "  ✅ encryption_key"
echo "  ✅ jwt_secret"
echo "  ✅ database_url"
echo "  ✅ database_password"
echo "  ✅ csrf_secret"
echo "  ✅ superadmin_password"
echo "  ✅ rate_limit"
if [ -n "$SONARCLOUD_TOKEN" ] && [ -n "$SONARCLOUD_PROJECT_KEY" ]; then
    echo "  ✅ SONARCLOUD_URL"
    echo "  ✅ SONARCLOUD_TOKEN"
    echo "  ✅ SONARCLOUD_PROJECT_KEY"
fi
echo ""
echo "Backend will automatically load these secrets on startup."
echo ""
echo "To add SonarCloud secrets later, run:"
echo "  export SONARCLOUD_URL=https://sonarcloud.io"
echo "  export SONARCLOUD_TOKEN=your-token"
echo "  export SONARCLOUD_PROJECT_KEY=your-project-key"
echo "  ./scripts/store-all-secrets.sh"
echo ""

