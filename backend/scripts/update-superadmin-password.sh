#!/bin/bash

# Script untuk update password superadmin dari Vault
# Usage: ./scripts/update-superadmin-password.sh

set -e

echo "=========================================="
echo "Update Superadmin Password from Vault"
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

# Get password from Vault
echo "Retrieving superadmin password from Vault..."
if docker ps | grep -q dms-vault-dev; then
    VAULT_CMD="docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=$VAULT_TOKEN dms-vault-dev vault"
else
    if command -v vault &> /dev/null; then
        VAULT_CMD="vault"
    else
        echo "❌ Vault container not running and vault CLI not found!"
        exit 1
    fi
fi

SUPERADMIN_PASSWORD=$($VAULT_CMD kv get -field=superadmin_password "$VAULT_SECRET_PATH" 2>&1)

if [ $? -ne 0 ] || [ -z "$SUPERADMIN_PASSWORD" ]; then
    echo "❌ Failed to retrieve superadmin_password from Vault"
    echo "   Make sure the secret exists: $VAULT_SECRET_PATH"
    exit 1
fi

echo "✅ Password retrieved from Vault"
echo ""

# Check if backend is running
if ! docker ps | grep -q dms-backend-dev; then
    echo "⚠️  Backend container not running!"
    echo "   Password will be updated when backend starts with SUPERADMIN_AUTO_SYNC_PASSWORD=true"
    echo ""
    echo "   Or you can update manually via database:"
    echo "   docker exec -it dms-postgres-dev psql -U postgres -d db_dms_pedeve -c \"UPDATE users SET password = '\$(echo -n '$SUPERADMIN_PASSWORD' | docker exec -i dms-backend-dev go run ./cmd/update-password/main.go)' WHERE username = 'superadmin';\""
    exit 0
fi

# Update via Go command
echo "Updating superadmin password in database..."
echo ""

# Check if backend container is running
if docker ps | grep -q dms-backend-dev; then
    # Run the update command via backend container
    if docker exec -e VAULT_ADDR="$VAULT_ADDR" -e VAULT_TOKEN="$VAULT_TOKEN" -e VAULT_SECRET_PATH="$VAULT_SECRET_PATH" dms-backend-dev go run ./cmd/update-superadmin-password/main.go 2>&1; then
        echo ""
        echo "=========================================="
        echo "✅ Superadmin Password Updated!"
        echo "=========================================="
        echo ""
        echo "Password has been updated from Vault."
        echo "You can now login with the new password."
    else
        echo ""
        echo "❌ Failed to update password"
        echo ""
        echo "Alternative: Enable auto-sync on next startup:"
        echo "  Add to docker-compose.dev.yml backend environment:"
        echo "    - SUPERADMIN_AUTO_SYNC_PASSWORD=true"
        echo ""
        echo "  Then restart backend:"
        echo "    docker-compose -f docker-compose.dev.yml restart backend"
        exit 1
    fi
else
    echo "⚠️  Backend container not running!"
    echo ""
    echo "To update password:"
    echo "  1. Start backend: docker-compose -f docker-compose.dev.yml up -d backend"
    echo "  2. Run this script again"
    echo ""
    echo "Or enable auto-sync on startup:"
    echo "  Add to docker-compose.dev.yml backend environment:"
    echo "    - SUPERADMIN_AUTO_SYNC_PASSWORD=true"
    exit 1
fi

