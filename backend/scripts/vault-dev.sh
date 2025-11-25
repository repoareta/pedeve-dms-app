#!/bin/bash

# Script untuk start Vault dev server dan setup secret untuk development
# Usage: ./scripts/vault-dev.sh

set -e

echo "=========================================="
echo "HashiCorp Vault Dev Server Setup"
echo "=========================================="

# Check if Vault is installed
if ! command -v vault &> /dev/null; then
    echo "❌ Vault is not installed!"
    echo ""
    echo "Install Vault:"
    echo "  macOS: brew install vault"
    echo "  Linux: https://developer.hashicorp.com/vault/downloads"
    exit 1
fi

echo "✅ Vault is installed: $(vault version | head -1)"
echo ""

# Check if Vault is already running
if vault status &> /dev/null; then
    echo "⚠️  Vault is already running!"
    echo ""
    echo "Current Vault status:"
    vault status
    echo ""
    read -p "Do you want to continue with existing Vault? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
else
    echo "Starting Vault dev server..."
    echo ""
    echo "⚠️  This will start Vault in dev mode (data stored in memory)"
    echo "⚠️  Press Ctrl+C to stop Vault server"
    echo ""
    echo "Starting Vault in background..."
    
    # Start Vault in background
    vault server -dev > /tmp/vault-dev.log 2>&1 &
    VAULT_PID=$!
    
    # Wait for Vault to start
    sleep 2
    
    # Check if Vault started successfully
    if ! vault status &> /dev/null; then
        echo "❌ Failed to start Vault server"
        cat /tmp/vault-dev.log
        exit 1
    fi
    
    echo "✅ Vault dev server started (PID: $VAULT_PID)"
    echo ""
fi

# Get Vault address and token
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
export VAULT_ADDR

# Try to get token from environment or use dev token
if [ -z "$VAULT_TOKEN" ]; then
    # In dev mode, try to extract token from log or use default
    if [ -f /tmp/vault-dev.log ]; then
        VAULT_TOKEN=$(grep "Root Token:" /tmp/vault-dev.log | awk '{print $NF}' | head -1)
    fi
    
    if [ -z "$VAULT_TOKEN" ]; then
        echo "⚠️  VAULT_TOKEN not set. Please set it manually:"
        echo "   export VAULT_TOKEN=\"<token>\""
        echo ""
        echo "Or start Vault manually and copy the Root Token from output"
        exit 1
    fi
fi

export VAULT_TOKEN

echo "Vault Configuration:"
echo "  Address: $VAULT_ADDR"
echo "  Token: ${VAULT_TOKEN:0:10}..." # Show first 10 chars only
echo ""

# Enable KV secrets engine if not already enabled
if ! vault secrets list | grep -q "^kv/"; then
    echo "Enabling KV secrets engine..."
    vault secrets enable -version=2 kv
    echo "✅ KV secrets engine enabled"
else
    echo "✅ KV secrets engine already enabled"
fi

echo ""

# Generate encryption key if not provided
if [ -z "$ENCRYPTION_KEY" ]; then
    echo "Generating encryption key..."
    ENCRYPTION_KEY=$(openssl rand -base64 32 | tr -d '\n' | head -c 32)
    # Pad to 32 bytes if needed
    while [ ${#ENCRYPTION_KEY} -lt 32 ]; do
        ENCRYPTION_KEY="${ENCRYPTION_KEY}!"
    done
    ENCRYPTION_KEY="${ENCRYPTION_KEY:0:32}"
    echo "✅ Generated encryption key (32 bytes)"
else
    echo "Using provided ENCRYPTION_KEY"
fi

echo ""

# Store encryption key in Vault
VAULT_SECRET_PATH="${VAULT_SECRET_PATH:-secret/data/dms-app}"
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
echo "=========================================="
echo "✅ Vault Setup Complete!"
echo "=========================================="
echo ""
echo "Environment variables for your application:"
echo ""
echo "export VAULT_ADDR=\"$VAULT_ADDR\""
echo "export VAULT_TOKEN=\"$VAULT_TOKEN\""
echo "export VAULT_SECRET_PATH=\"$VAULT_SECRET_PATH\""
echo ""
echo "Now you can run your application:"
echo "  make dev"
echo ""
echo "To stop Vault dev server:"
echo "  pkill vault"
echo ""

