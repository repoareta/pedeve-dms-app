#!/bin/bash
set -euo pipefail

# Script untuk setup Nginx di frontend VM
# Usage: ./setup-nginx-frontend.sh [DOMAIN]
# 
# Jika DOMAIN tidak diberikan, akan menggunakan default untuk development
# Atau bisa set via environment variable: DOMAIN=dms.pertamina-pedeve.co.id

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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/resolve-domain.sh
source "${SCRIPT_DIR}/lib/resolve-domain.sh"

export DEV_DOMAIN_DEFAULT="pedeve-dev.aretaamany.com"
DOMAIN=$(resolve_deploy_domain "${1:-}")

# Security: Validate domain
validate_domain "${DOMAIN}"

echo "🔧 Setting up Nginx for frontend..."
echo "   Domain: ${DOMAIN}"

# Backup default config
if [ -f /etc/nginx/sites-available/default ]; then
  sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup
fi

# Check if config already exists and is correct
CONFIG_EXISTS=false
CONFIG_CORRECT=false
SSL_CERT_EXISTS=false

ssl_cert_exists_for_domain() {
  sudo test -f "/etc/letsencrypt/live/${DOMAIN}/fullchain.pem" \
    && sudo test -f "/etc/letsencrypt/live/${DOMAIN}/privkey.pem"
}

# Check if SSL certificate exists
# IMPORTANT: Preserve existing SSL certificate - DO NOT OVERWRITE
# Also check if port 443 is listening - if cert exists but port not listening, we need to fix config
if ssl_cert_exists_for_domain "${DOMAIN}"; then
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

# Check if default config already exists
if [ -f /etc/nginx/sites-available/default ]; then
  CONFIG_EXISTS=true
  echo "✅ Frontend Nginx config already exists"
  
  # Check if config is correct (has root /var/www/html and SPA routing)
  if sudo grep -q "root /var/www/html" /etc/nginx/sites-available/default && \
     sudo grep -q "try_files.*index.html" /etc/nginx/sites-available/default; then
    CONFIG_CORRECT=true
    echo "✅ Frontend Nginx config is correct"
    
    # If SSL exists, check if config has HTTPS block
    # IMPORTANT: Preserve existing SSL configuration - DO NOT OVERWRITE if correct
    if [ "$SSL_CERT_EXISTS" = true ]; then
      # Comprehensive check for HTTPS config
      if sudo grep -q "ssl_certificate.*${DOMAIN}" /etc/nginx/sites-available/default && \
         sudo grep -q "listen.*443.*ssl" /etc/nginx/sites-available/default && \
         sudo grep -q "server_name.*${DOMAIN}" /etc/nginx/sites-available/default && \
         sudo grep -q "ssl_certificate_key.*${DOMAIN}" /etc/nginx/sites-available/default; then
        echo "✅ HTTPS config already present and correct"
        
        # CRITICAL: Validate config syntax before skipping
        echo "🧪 Validating existing Nginx config syntax..."
        if sudo nginx -t 2>/dev/null; then
          echo "✅ Nginx config syntax is valid"
          echo "⏭️  SKIPPING config update - preserving existing SSL configuration"
          echo "   - SSL certificate: /etc/letsencrypt/live/${DOMAIN}/"
          echo "   - Port 443: configured"
          echo "   - Server name: ${DOMAIN}"
          echo "   - Config file: /etc/nginx/sites-available/default"
          echo ""
          echo "🔒 PRESERVATION MODE: Config will NOT be overwritten"
          
          # Just ensure it's enabled and reload
          sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default
          
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
            if sudo ss -tlnp | grep -q ':80 '; then
              echo "✅ Port 80 is listening - HTTP redirect is active"
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
      # No SSL detected in cert path — but double-check before keeping HTTP-only
      if ssl_cert_exists_for_domain "${DOMAIN}"; then
        echo "⚠️  SSL certificate found via sudo check, will configure HTTPS..."
        CONFIG_CORRECT=false
      elif ! sudo grep -q "ssl_certificate" /etc/nginx/sites-available/default; then
        echo "✅ HTTP-only config is correct (no SSL)"
        
        # CRITICAL: Validate config syntax before skipping
        echo "🧪 Validating existing Nginx config syntax..."
        if sudo nginx -t 2>/dev/null; then
          echo "✅ Nginx config syntax is valid"
          echo "⏭️  SKIPPING config update - existing config is correct"
          echo "   - Server name: (default or configured)"
          echo "   - Root: /var/www/html"
          echo "   - Config file: /etc/nginx/sites-available/default"
          echo ""
          echo "🔒 PRESERVATION MODE: Config will NOT be overwritten"
          
          # Just ensure it's enabled and reload
          sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default
          
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
  if [ -f ~/setup-frontend-ssl.sh ]; then
    echo "🔒 Running SSL setup script..."
    chmod +x ~/setup-frontend-ssl.sh
    if ~/setup-frontend-ssl.sh; then
      echo "✅ SSL setup completed"
      # Wait a moment for Certbot to finish
      sleep 2
    else
      echo "⚠️  SSL setup script returned non-zero, but checking if certificate exists anyway..."
    fi
    
    # Always re-check if certificate exists (Certbot might have created it)
    if ssl_cert_exists_for_domain "${DOMAIN}"; then
      SSL_CERT_EXISTS=true
      echo "✅ SSL certificate found after SSL setup"
    else
      echo "⚠️  SSL certificate not found after setup attempt"
      echo "   This might be normal if DNS is not configured or Let's Encrypt rate limit"
    fi
  else
    echo "⚠️  SSL setup script not found, will create HTTP-only config"
    echo "   To enable HTTPS, run setup-frontend-ssl.sh manually"
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
    sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)
    echo "📦 Backed up existing config"
  fi
  
  # Create Nginx config with HTTPS
  # Temporarily disable unbound variable check for heredoc (Nginx variables will be evaluated by Nginx, not bash)
  set +u
  sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} _;

    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};

    ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    root /var/www/html;
    index index.html;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # index.html should NEVER be cached - it contains references to hashed assets
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
        try_files \$uri /index.html;
    }

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Static assets with hash in filename can be cached aggressively
    # Vite generates files like index-abc123.js and index-xyz789.css
    # These hashes change on every build, so caching is safe
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF
  set -u
else
  echo "⚠️  SSL certificate not found, creating/updating HTTP-only config..."
  
  # Backup existing config if it exists
  if [ "$CONFIG_EXISTS" = true ]; then
    sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup.$(date +%Y%m%d_%H%M%S)
    echo "📦 Backed up existing config"
  fi
  
  # Create Nginx config for SPA (HTTP only)
  # Temporarily disable unbound variable check for heredoc (Nginx variables will be evaluated by Nginx, not bash)
  set +u
  sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name _;

    root /var/www/html;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # index.html should NEVER be cached - it contains references to hashed assets
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
        try_files \$uri /index.html;
    }

    # SPA routing - semua request ke index.html kecuali static files
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Static assets with hash in filename can be cached aggressively
    # Vite generates files like index-abc123.js and index-xyz789.css
    # These hashes change on every build, so caching is safe
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint (optional)
    location /health {
        access_log off;
        return 200 "OK\n";
        add_header Content-Type text/plain;
    }
}
EOF
  set -u
fi

# Only remove conflicting enabled sites (not all)
# Keep default if it's already enabled
if [ -L /etc/nginx/sites-enabled/default ]; then
  echo "✅ default site already enabled"
else
  # Remove only backend-api and other conflicting sites
  sudo rm -f /etc/nginx/sites-enabled/backend-api
  # Enable default site
  sudo ln -sf /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default
  echo "✅ Enabled default site"
fi

# Remove backend config if it exists (shouldn't be on frontend VM)
if [ -f /etc/nginx/sites-available/backend-api ]; then
  echo "🧹 Removing backend config from frontend VM..."
  sudo rm -f /etc/nginx/sites-available/backend-api
fi

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

echo "✅ Nginx setup completed!"
echo ""
echo "📋 Verification:"
echo "   - Config file: /etc/nginx/sites-available/default"
echo "   - Web root: /var/www/html"
echo "   - Test: curl http://localhost/health"

