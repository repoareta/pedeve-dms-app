# 🔒 Setup SSL Frontend - Manual di VM

## Masalah
`gcloud compute scp` tidak bisa dijalankan dari dalam VM karena insufficient authentication scopes.

## Solusi: Buat Script Langsung di VM

Jalankan perintah berikut **di SSH frontend VM** untuk membuat script langsung:

```bash
cat > ~/setup-frontend-ssl.sh << 'SCRIPT_EOF'
#!/bin/bash
set -euo pipefail

# Script untuk setup SSL certificate untuk frontend
DOMAIN="pedeve-dev.aretaamany.com"
EMAIL="info@aretaamany.com"

echo "🔒 Setting up SSL certificate for ${DOMAIN}..."

# Install Certbot if not exists
if ! command -v certbot &> /dev/null; then
  echo "📦 Installing Certbot..."
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
fi

# Update Nginx config untuk support SSL
echo "📝 Updating Nginx config for SSL..."

sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} _;

    # Redirect HTTP to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};

    # SSL certificate (will be set by Certbot)
    # ssl_certificate /etc/letsencrypt/live/${DOMAIN}/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/${DOMAIN}/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    root /var/www/html;
    index index.html;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

    # SPA routing - semua request ke index.html kecuali static files
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Cache static assets
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

# Reload Nginx
sudo systemctl reload nginx

# Generate SSL certificate with Certbot
echo "🔐 Generating SSL certificate..."
sudo certbot --nginx \
  -d ${DOMAIN} \
  --email ${EMAIL} \
  --agree-tos \
  --non-interactive \
  --redirect

# Setup auto-renewal
echo "🔄 Setting up auto-renewal..."
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Test renewal
echo "🧪 Testing certificate renewal..."
sudo certbot renew --dry-run

echo ""
echo "✅ SSL certificate setup completed!"
echo ""
echo "📋 Summary:"
echo "   - Domain: ${DOMAIN}"
echo "   - SSL Certificate: Let's Encrypt"
echo "   - Auto-renewal: Enabled"
echo ""
echo "🧪 Test commands:"
echo "   curl https://${DOMAIN}/health"
echo "   curl -I http://${DOMAIN}/health  # Should redirect to HTTPS"
SCRIPT_EOF

# Set executable permission
chmod +x ~/setup-frontend-ssl.sh

# Jalankan script
sudo ./setup-frontend-ssl.sh
```

## Atau Copy-Paste Langsung

Jika lebih mudah, copy-paste perintah berikut **satu per satu** di SSH frontend VM:

```bash
# 1. Buat script
cat > ~/setup-frontend-ssl.sh << 'SCRIPT_EOF'
#!/bin/bash
set -euo pipefail
DOMAIN="pedeve-dev.aretaamany.com"
EMAIL="info@aretaamany.com"
echo "🔒 Setting up SSL certificate for ${DOMAIN}..."
if ! command -v certbot &> /dev/null; then
  echo "📦 Installing Certbot..."
  sudo apt-get update
  sudo apt-get install -y certbot python3-certbot-nginx
fi
echo "📝 Updating Nginx config for SSL..."
sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN} _;
    return 301 https://\$server_name\$request_uri;
}
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name ${DOMAIN};
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
        try_files \$uri \$uri/ /index.html;
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
sudo systemctl reload nginx
echo "🔐 Generating SSL certificate..."
sudo certbot --nginx -d ${DOMAIN} --email ${EMAIL} --agree-tos --non-interactive --redirect
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
sudo certbot renew --dry-run
echo "✅ SSL certificate setup completed!"
SCRIPT_EOF

# 2. Set permission
chmod +x ~/setup-frontend-ssl.sh

# 3. Jalankan
sudo ./setup-frontend-ssl.sh
```

## Verifikasi Setelah Setup

```bash
# Test HTTPS
curl https://pedeve-dev.aretaamany.com/health

# Test HTTP redirect
curl -I http://pedeve-dev.aretaamany.com/health

# Cek SSL certificate
sudo certbot certificates

# Cek port 443 listening
sudo ss -tlnp | grep 443
```

