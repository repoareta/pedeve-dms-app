#!/bin/bash
set -euo pipefail

# Script untuk setup Nginx reverse proxy di backend VM
# Usage: ./setup-backend-nginx.sh [DOMAIN]
# 
# Jika DOMAIN tidak diberikan, akan menggunakan default untuk development
# Atau bisa set via environment variable: DOMAIN=api-reports.pertamina-pedeve.co.id

# Security: Domain validation function
validate_domain() {
  local domain=$1
  # Domain format: alphanumeric, dots, hyphens, max 253 chars
  # Reject path traversal characters
  if [[ "$domain" =~ [\/\\\$\`\;] ]] || [[ "$domain" =~ \.\. ]]; then
    echo "❌ ERROR: Invalid DOMAIN format. Contains dangerous characters"
    exit 1
  fi
  # Basic domain format validation
  if [[ ! "$domain" =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$ ]]; then
    echo "❌ ERROR: Invalid DOMAIN format"
    exit 1
  fi
}

DOMAIN=${1:-${DOMAIN:-"api-pedeve-dev.aretaamany.com"}}

# Security: Validate domain
validate_domain "${DOMAIN}"

echo "🔧 Setting up Nginx reverse proxy for backend..."
echo "   Domain: ${DOMAIN}"

# Install Nginx if not exists
if ! command -v nginx &> /dev/null; then
  echo "📦 Installing Nginx..."
  sudo apt-get update
  sudo apt-get install -y nginx
  sudo systemctl enable nginx
fi

# Backup default config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup
fi

# Check if config already exists and is correct
CONFIG_EXISTS=false
CONFIG_CORRECT=false
SSL_CERT_EXISTS=false

# Check if SSL certificate exists
# IMPORTANT: Preserve existing SSL certificate - DO NOT OVERWRITE
# Also check if port 443 is listening - if cert exists but port not listening, we need to fix config
if [ -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem ] && \
   [ -f /etc/letsencrypt/live/${DOMAIN}/privkey.pem ]; then
  SSL_CERT_EXISTS=true
  echo "✅ SSL certificate found (preserving existing certificate)"
  
  # Check if port 443 is listening
  if ! sudo ss -tlnp | grep -q ':443 '; then
    echo "⚠️  WARNING: SSL certificate exists but port 443 is not listening"
    echo "   This means Nginx config needs to be updated with HTTPS block"
    # Force update config even if it exists
    CONFIG_CORRECT=false
  fi
fi

# Check if backend-api config already exists
if [ -f /etc/nginx/sites-available/backend-api ]; then
  CONFIG_EXISTS=true
  echo "✅ Backend Nginx config already exists"
  
  # Check if config is correct (has correct server_name and proxy_pass)
  if sudo grep -q "server_name ${DOMAIN}" /etc/nginx/sites-available/backend-api && \
     sudo grep -q "proxy_pass http://127.0.0.1:8080" /etc/nginx/sites-available/backend-api; then
    CONFIG_CORRECT=true
    echo "✅ Backend Nginx config is correct"
    
    # If SSL exists, check if config has HTTPS block
    # IMPORTANT: Preserve existing SSL configuration - DO NOT OVERWRITE if correct
    if [ "$SSL_CERT_EXISTS" = true ]; then
      # Comprehensive check for HTTPS config
      if sudo grep -q "ssl_certificate.*${DOMAIN}" /etc/nginx/sites-available/backend-api && \
         sudo grep -q "listen.*443.*ssl" /etc/nginx/sites-available/backend-api && \
         sudo grep -q "server_name.*${DOMAIN}" /etc/nginx/sites-available/backend-api && \
         sudo grep -q "ssl_certificate_key.*${DOMAIN}" /etc/nginx/sites-available/backend-api && \
         sudo grep -q "proxy_pass.*127.0.0.1:8080" /etc/nginx/sites-available/backend-api; then
        echo "✅ HTTPS config already present and correct"
        
        # CRITICAL: Validate config syntax before skipping
        echo "🧪 Validating existing Nginx config syntax..."
        if sudo nginx -t 2>/dev/null; then
          echo "✅ Nginx config syntax is valid"
          echo "⏭️  SKIPPING config update - preserving existing SSL configuration"
          echo "   - SSL certificate: /etc/letsencrypt/live/${DOMAIN}/"
          echo "   - Port 443: configured"
          echo "   - Server name: ${DOMAIN}"
          echo "   - Proxy pass: http://127.0.0.1:8080"
          echo "   - Config file: /etc/nginx/sites-available/backend-api"
          echo ""
          echo "🔒 PRESERVATION MODE: Config will NOT be overwritten"
          
          # Just ensure it's enabled and reload
          sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api
          
          # Reload Nginx to ensure config is active
          if sudo nginx -t 2>/dev/null; then
            sudo systemctl reload nginx || sudo systemctl restart nginx
            sleep 2
            
            # Verify port is listening
            if sudo ss -tlnp | grep -q ':443 '; then
              echo "✅ Port 443 is listening - SSL config is active"
            else
              echo "⚠️  WARNING: Port 443 not listening, but config is preserved"
            fi
          fi
          
          echo "✅ Existing configuration preserved successfully"
          exit 0
        else
          echo "⚠️  Config exists but syntax is invalid, will fix..."
          echo "   - This is safe - we will fix config while preserving SSL certificate paths"
        fi
      else
        echo "⚠️  SSL exists but config doesn't have HTTPS block, will update..."
        echo "   - This is safe - we will add HTTPS block without removing existing config"
        CONFIG_CORRECT=false  # Force update
      fi
    else
      # No SSL, check if config is HTTP-only (correct)
      if ! sudo grep -q "ssl_certificate" /etc/nginx/sites-available/backend-api; then
        echo "✅ HTTP-only config is correct (no SSL)"
        
        # CRITICAL: Validate config syntax before skipping
        echo "🧪 Validating existing Nginx config syntax..."
        if sudo nginx -t 2>/dev/null; then
          echo "✅ Nginx config syntax is valid"
          echo "⏭️  SKIPPING config update - existing config is correct"
          echo "   - Server name: ${DOMAIN}"
          echo "   - Proxy pass: http://127.0.0.1:8080"
          echo "   - Config file: /etc/nginx/sites-available/backend-api"
          echo ""
          echo "🔒 PRESERVATION MODE: Config will NOT be overwritten"
          
          # Just ensure it's enabled and reload
          sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api
          
          # Reload Nginx to ensure config is active
          if sudo nginx -t 2>/dev/null; then
            sudo systemctl reload nginx || sudo systemctl restart nginx
            sleep 2
            
            # Verify port is listening
            if sudo ss -tlnp | grep -q ':80 '; then
              echo "✅ Port 80 is listening - HTTP config is active"
            else
              echo "⚠️  WARNING: Port 80 not listening, but config is preserved"
            fi
          fi
          
          echo "✅ Existing configuration preserved successfully"
          exit 0
        else
          echo "⚠️  Config exists but syntax is invalid, will fix..."
        fi
      fi
    fi
  fi
fi

# If SSL certificate doesn't exist, try to setup SSL first
if [ "$SSL_CERT_EXISTS" = false ]; then
  echo "⚠️  SSL certificate not found, attempting to setup SSL..."
  
  # Check if SSL setup script exists
  if [ -f ~/setup-backend-ssl.sh ]; then
    echo "🔒 Running SSL setup script..."
    chmod +x ~/setup-backend-ssl.sh
    if ~/setup-backend-ssl.sh; then
      echo "✅ SSL setup completed"
      # Wait a moment for Certbot to finish
      sleep 2
    else
      echo "⚠️  SSL setup script returned non-zero, but checking if certificate exists anyway..."
    fi
    
    # Always re-check if certificate exists (Certbot might have created it)
    if [ -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem ] && \
       [ -f /etc/letsencrypt/live/${DOMAIN}/privkey.pem ]; then
      SSL_CERT_EXISTS=true
      echo "✅ SSL certificate found after SSL setup"
    else
      echo "⚠️  SSL certificate not found after setup attempt"
      echo "   This might be normal if DNS is not configured or Let's Encrypt rate limit"
    fi
  else
    echo "⚠️  SSL setup script not found, will create HTTP-only config"
    echo "   To enable HTTPS, run setup-backend-ssl.sh manually"
  fi
fi

# Only create/update config if needed
# CRITICAL: If SSL certificate exists but port 443 is not listening, we MUST update config
if [ "$SSL_CERT_EXISTS" = true ]; then
  # Double-check port 443
  PORT_443_LISTENING=false
  if sudo ss -tlnp | grep -q ':443 '; then
    PORT_443_LISTENING=true
  fi
  
  if [ "$PORT_443_LISTENING" = false ]; then
    echo "⚠️  SSL certificate exists but port 443 is not listening"
    echo "   Forcing Nginx config update with HTTPS block..."
    CONFIG_CORRECT=false
  fi
  
  echo "✅ SSL certificate found, creating/updating config with HTTPS..."
  
  # Backup existing config if it exists
  if [ "$CONFIG_EXISTS" = true ]; then
    sudo cp /etc/nginx/sites-available/backend-api /etc/nginx/sites-available/backend-api.backup.$(date +%Y%m%d_%H%M%S)
    echo "📦 Backed up existing config"
  fi
  
  # Create Nginx config with HTTPS
  # Temporarily disable unbound variable check for heredoc (Nginx variables will be evaluated by Nginx, not bash)
  set +u
  sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};
    # Allow uploads up to 10MB (matching Fiber BodyLimit)
    client_max_body_size 10m;

    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    access_log /var/log/nginx/backend-api-access.log;
    error_log /var/log/nginx/backend-api-error.log;

    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;

    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF
  set -u
else
  echo "⚠️  SSL certificate not found, creating/updating HTTP-only config..."
  
  # Backup existing config if it exists
  if [ "$CONFIG_EXISTS" = true ]; then
    sudo cp /etc/nginx/sites-available/backend-api /etc/nginx/sites-available/backend-api.backup.$(date +%Y%m%d_%H%M%S)
    echo "📦 Backed up existing config"
  fi
  
  # Create Nginx config for backend API reverse proxy (HTTP only)
  # Temporarily disable unbound variable check for heredoc (Nginx variables will be evaluated by Nginx, not bash)
  set +u
  sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    # Allow uploads up to 10MB (matching Fiber BodyLimit)
    client_max_body_size 10m;

    # Logging
    access_log /var/log/nginx/backend-api-access.log;
    error_log /var/log/nginx/backend-api-error.log;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Proxy settings
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;

    # Timeout settings
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    # Forward all requests to backend on port 8080
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF
  set -u
fi

# Only remove conflicting enabled sites (not all)
# Keep backend-api if it's already enabled
if [ -L /etc/nginx/sites-enabled/backend-api ]; then
  echo "✅ backend-api already enabled"
else
  # Remove only default and other conflicting sites
  sudo rm -f /etc/nginx/sites-enabled/default
  # Enable backend-api site
  sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api
  echo "✅ Enabled backend-api site"
fi

# Hapus config frontend jika ter-copy (pastikan tidak ada conflict)
# (tidak perlu, karena frontend pakai default)

# Test Nginx config
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# Enable Nginx to start on boot
echo "🔧 Enabling Nginx to start on boot..."
sudo systemctl enable nginx

# Reload Nginx (reload is safer than restart, preserves connections)
echo "🔄 Reloading Nginx..."
if sudo nginx -t 2>/dev/null; then
  sudo systemctl reload nginx || sudo systemctl restart nginx
else
  echo "⚠️  Nginx config test failed, trying restart..."
  sudo systemctl restart nginx
fi

# Ensure Nginx is running
echo "▶️  Starting Nginx if not running..."
sudo systemctl start nginx || sudo systemctl restart nginx

# Wait a moment for Nginx to fully start
sleep 2

# Check Nginx status
echo "📊 Nginx status:"
sudo systemctl status nginx --no-pager -l || true

# Verify Nginx is active
if sudo systemctl is-active --quiet nginx; then
  echo "✅ Nginx is running and enabled"
  
  # Verify listening ports
  echo "📡 Checking listening ports..."
  if sudo ss -tlnp | grep -q ':80 '; then
    echo "✅ Port 80 is listening"
  else
    echo "⚠️  Port 80 is not listening"
  fi
  
  if sudo ss -tlnp | grep -q ':443 '; then
    echo "✅ Port 443 is listening"
  else
    echo "❌ ERROR: Port 443 is not listening after config update!"
    echo "   This might indicate a configuration problem"
    echo "   Checking Nginx error log..."
    sudo tail -20 /var/log/nginx/error.log 2>/dev/null || true
    echo "   Attempting to restart Nginx..."
    sudo systemctl restart nginx
    sleep 3
    if sudo ss -tlnp | grep -q ':443 '; then
      echo "✅ Port 443 is now listening after restart"
    else
      echo "❌ Port 443 still not listening - please check Nginx configuration manually"
    fi
  fi
else
  echo "❌ ERROR: Nginx failed to start!"
  echo "Nginx error log:"
  sudo tail -20 /var/log/nginx/error.log 2>/dev/null || true
  exit 1
fi

echo "✅ Nginx reverse proxy setup completed!"
echo ""
echo "📋 Configuration:"
echo "   - Listen: port 80"
echo "   - Server name: ${DOMAIN}"
echo "   - Proxy to: http://127.0.0.1:8080"
echo ""
echo "🧪 Test commands:"
echo "   curl http://${DOMAIN}/health"
echo "   curl http://${DOMAIN}/api/v1/csrf-token"
