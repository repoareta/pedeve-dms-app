#!/bin/bash
# Script untuk fix frontend Nginx config dengan SSL certificate path yang benar
# Jalankan di frontend VM: sudo bash fix-frontend-nginx.sh

echo "🔧 Fixing frontend Nginx config..."

# Update Nginx config dengan path certificate yang BENAR (pedeve-dev)
sudo tee /etc/nginx/sites-available/default > /dev/null <<'EOF'
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name pedeve-dev.aretaamany.com _;

    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name pedeve-dev.aretaamany.com;

    # PASTIKAN PATH INI BENAR: pedeve-dev (frontend), BUKAN api-pedeve-dev (backend)
    ssl_certificate /etc/letsencrypt/live/pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/pedeve-dev.aretaamany.com/privkey.pem;

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

    location / {
        try_files $uri $uri/ /index.html;
    }

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

echo "✅ Nginx config updated"

# Verify certificate path exists (with sudo)
echo "🔍 Verifying certificate path..."
sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/ | head -5

# Test config
echo "🧪 Testing Nginx config..."
if sudo nginx -t; then
    echo "✅ Nginx config test passed"
    
    # Reload Nginx
    echo "🔄 Reloading Nginx..."
    sudo systemctl reload nginx
    
    # Wait a moment
    sleep 2
    
    # Verify port 443 listening
    echo "📡 Checking port 443..."
    if sudo ss -tlnp | grep -q ':443'; then
        echo "✅ Port 443 is listening"
        sudo ss -tlnp | grep 443
    else
        echo "❌ Port 443 is not listening"
    fi
    
    # Test HTTPS
    echo "🧪 Testing HTTPS..."
    if curl -k -s -m 5 https://127.0.0.1/health > /dev/null; then
        echo "✅ HTTPS is working"
        curl -k https://127.0.0.1/health
    else
        echo "❌ HTTPS test failed"
    fi
else
    echo "❌ Nginx config test failed!"
    echo "Check error above"
    exit 1
fi

echo ""
echo "✅ Frontend Nginx fix completed!"

