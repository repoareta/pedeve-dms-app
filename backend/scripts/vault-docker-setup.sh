#!/bin/bash

# Script untuk setup Vault setelah container running
# Usage: ./scripts/vault-docker-setup.sh

set -e

echo "=========================================="
echo "HashiCorp Vault Docker Setup"
echo "=========================================="
echo ""

# Vault configuration untuk Docker
VAULT_ADDR="http://127.0.0.1:8200"
VAULT_TOKEN="dev-root-token-12345"

export VAULT_ADDR
export VAULT_TOKEN

echo "Vault Configuration:"
echo "  Address: $VAULT_ADDR"
echo "  Token: $VAULT_TOKEN"
echo ""

# Wait for Vault to be ready
echo "Waiting for Vault to be ready..."
for i in {1..30}; do
    if vault status &> /dev/null; then
        echo "✅ Vault is ready!"
        break
    fi
    echo "  Attempt $i/30..."
    sleep 1
done

if ! vault status &> /dev/null; then
    echo "❌ Vault is not ready. Please check if Vault container is running:"
    echo "   docker-compose -f docker-compose.dev.yml ps vault"
    exit 1
fi

echo ""
echo "Vault Status:"
vault status
echo ""

# Enable KV secrets engine if not already enabled
if ! vault secrets list | grep -q "^kv/"; then
    echo "Enabling KV secrets engine (v2)..."
    vault secrets enable -version=2 kv
    echo "✅ KV secrets engine enabled"
else
    echo "✅ KV secrets engine already enabled"
fi

echo ""

# Generate encryption key if not provided
if [ -z "$ENCRYPTION_KEY" ]; then
    echo "Generating 32-byte encryption key..."
    # Generate random 32-byte key (base64 encoded, then take first 32 chars)
    ENCRYPTION_KEY=$(openssl rand -base64 32 | tr -d '\n' | head -c 32)
    # Pad to 32 bytes if needed (shouldn't happen, but just in case)
    while [ ${#ENCRYPTION_KEY} -lt 32 ]; do
        ENCRYPTION_KEY="${ENCRYPTION_KEY}A"
    done
    ENCRYPTION_KEY="${ENCRYPTION_KEY:0:32}"
    echo "✅ Generated encryption key (32 bytes)"
else
    echo "Using provided ENCRYPTION_KEY"
    # Validate key length
    if [ ${#ENCRYPTION_KEY} -ne 32 ]; then
        echo "⚠️  WARNING: ENCRYPTION_KEY must be exactly 32 bytes (got ${#ENCRYPTION_KEY})"
        echo "   Padding or truncating to 32 bytes..."
        if [ ${#ENCRYPTION_KEY} -lt 32 ]; then
            while [ ${#ENCRYPTION_KEY} -lt 32 ]; do
                ENCRYPTION_KEY="${ENCRYPTION_KEY}A"
            done
        else
            ENCRYPTION_KEY="${ENCRYPTION_KEY:0:32}"
        fi
    fi
fi

echo ""

# Store encryption key in Vault
VAULT_SECRET_PATH="secret/data/dms-app"
echo "Storing encryption key in Vault..."
echo "  Path: $VAULT_SECRET_PATH"

vault kv put "$VAULT_SECRET_PATH" encryption_key="$ENCRYPTION_KEY"

if [ $? -eq 0 ]; then
    echo "✅ Encryption key stored successfully"
else
    echo "❌ Failed to store encryption key"
    exit 1
fi

echo ""
echo "Verifying stored key..."
vault kv get "$VAULT_SECRET_PATH"
echo ""

echo "=========================================="
echo "✅ Vault Setup Complete!"
echo "=========================================="
echo ""
echo "Vault Web UI: http://127.0.0.1:8200/ui"
echo "  Login with token: $VAULT_TOKEN"
echo ""
echo "Environment variables for backend:"
echo ""
echo "export VAULT_ADDR=\"$VAULT_ADDR\""
echo "export VAULT_TOKEN=\"$VAULT_TOKEN\""
echo "export VAULT_SECRET_PATH=\"$VAULT_SECRET_PATH\""
echo ""
echo "Or add to docker-compose.dev.yml backend environment:"
echo "  - VAULT_ADDR=$VAULT_ADDR"
echo "  - VAULT_TOKEN=$VAULT_TOKEN"
echo "  - VAULT_SECRET_PATH=$VAULT_SECRET_PATH"
echo ""

