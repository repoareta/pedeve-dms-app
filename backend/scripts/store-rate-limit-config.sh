#!/bin/bash

# Script untuk store rate limit configuration ke Vault
# Usage: ./scripts/store-rate-limit-config.sh

set -e

echo "=========================================="
echo "Store Rate Limit Config to Vault"
echo "=========================================="
echo ""

# Vault configuration
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-dev-root-token-12345}"
VAULT_SECRET_PATH="${VAULT_SECRET_PATH:-secret/data/dms-app}"

export VAULT_ADDR
export VAULT_TOKEN

echo "Vault Configuration:"
echo "  Address: $VAULT_ADDR"
echo "  Path: $VAULT_SECRET_PATH"
echo ""

# Check if running in Docker
if [ -f /.dockerenv ] || [ -n "$DOCKER_CONTAINER" ]; then
    echo "Running inside Docker container"
    VAULT_CMD="vault"
else
    # Check if vault CLI is available
    if ! command -v vault &> /dev/null; then
        echo "❌ Vault CLI not found!"
        echo "   Using Docker exec instead..."
        if docker ps | grep -q dms-vault-dev; then
            VAULT_CMD="docker exec dms-vault-dev vault"
            # Set VAULT_ADDR for docker exec
            export VAULT_ADDR="http://127.0.0.1:8200"
        else
            echo "❌ Vault container not running!"
            echo "   Please start Vault first: docker-compose -f docker-compose.dev.yml up -d vault"
            exit 1
        fi
    else
        VAULT_CMD="vault"
    fi
fi

# Rate limit configuration (default values)
GENERAL_RPS="${RATE_LIMIT_GENERAL_RPS:-500}"
GENERAL_BURST="${RATE_LIMIT_GENERAL_BURST:-500}"
AUTH_RPM="${RATE_LIMIT_AUTH_RPM:-5}"
AUTH_BURST="${RATE_LIMIT_AUTH_BURST:-5}"
STRICT_RPM="${RATE_LIMIT_STRICT_RPM:-50}"
STRICT_BURST="${RATE_LIMIT_STRICT_BURST:-50}"

echo "Rate Limit Configuration:"
echo "  General: ${GENERAL_RPS} req/s, burst: ${GENERAL_BURST}"
echo "  Auth: ${AUTH_RPM} req/min, burst: ${AUTH_BURST}"
echo "  Strict: ${STRICT_RPM} req/min, burst: ${STRICT_BURST}"
echo ""

# Build JSON config
RATE_LIMIT_CONFIG=$(cat <<EOF
{
  "rate_limit": {
    "general": {
      "rps": ${GENERAL_RPS},
      "burst": ${GENERAL_BURST}
    },
    "auth": {
      "rpm": ${AUTH_RPM},
      "burst": ${AUTH_BURST}
    },
    "strict": {
      "rpm": ${STRICT_RPM},
      "burst": ${STRICT_BURST}
    }
  }
}
EOF
)

echo "Storing rate limit config to Vault..."
echo "  Path: $VAULT_SECRET_PATH"
echo ""

# Store config in Vault
# For KV v2, we need to use the data path
if [[ "$VAULT_SECRET_PATH" == *"/data/"* ]]; then
    # KV v2 format - extract the base path
    BASE_PATH=$(echo "$VAULT_SECRET_PATH" | sed 's|/data/.*||')
    SECRET_NAME=$(echo "$VAULT_SECRET_PATH" | sed 's|.*/data/||')
    
    # Store as JSON string in rate_limit_config key
    echo "$RATE_LIMIT_CONFIG" | $VAULT_CMD kv put "${BASE_PATH}/data/${SECRET_NAME}" rate_limit_config=- 2>&1
    
    # Also merge with existing encryption_key if exists
    echo "✅ Rate limit config stored successfully"
    echo ""
    echo "Note: If you have encryption_key in the same path, it will be preserved."
    echo "      Both encryption_key and rate_limit_config can coexist."
else
    # KV v1 format
    echo "$RATE_LIMIT_CONFIG" | $VAULT_CMD kv put "$VAULT_SECRET_PATH" rate_limit_config=- 2>&1
    echo "✅ Rate limit config stored successfully"
fi

echo ""
echo "Verifying stored config..."
$VAULT_CMD kv get "$VAULT_SECRET_PATH" 2>&1 | head -20

echo ""
echo "=========================================="
echo "✅ Rate Limit Config Stored!"
echo "=========================================="
echo ""
echo "Backend will automatically load this config on startup."
echo ""

