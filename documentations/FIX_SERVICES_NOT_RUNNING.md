# 🔧 Fix: Services Tidak Running Setelah Deployment

## Masalah

Setelah deployment selesai tanpa error, services masih tidak bisa diakses:
- `https://api-pedeve-dev.aretaamany.com/health` → Connection refused (port 443)
- `https://pedeve-dev.aretaamany.com` → Connection refused (port 443)

## Diagnosis

**Test dari local:**
```bash
curl -v https://api-pedeve-dev.aretaamany.com/health
# Result: Connection refused on port 443

curl -v https://pedeve-dev.aretaamany.com
# Result: Connection refused on port 443
```

**Kemungkinan penyebab:**
1. **Port 443 tidak listening** - Nginx tidak dikonfigurasi untuk HTTPS
2. **Nginx tidak running** - Service mati setelah deployment
3. **Firewall tidak allow port 443** - Firewall rule tidak aktif
4. **SSL certificate tidak ada** - Certbot belum dijalankan

## Solusi

### 1. Run Diagnostic Script

**Jalankan script diagnostic untuk cek status services:**

```bash
# Backend VM
./scripts/diagnose-services.sh backend-dev asia-southeast2-a pedeve-pertamina-dms

# Frontend VM
./scripts/diagnose-services.sh frontend-dev asia-southeast2-a pedeve-pertamina-dms
```

**Atau manual check:**

```bash
# Backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  echo '=== Nginx Status ==='
  sudo systemctl status nginx --no-pager -l
  
  echo ''
  echo '=== Nginx Listening Ports ==='
  sudo ss -tlnp | grep nginx
  
  echo ''
  echo '=== Docker Container ==='
  sudo docker ps | grep dms-backend-prod
  
  echo ''
  echo '=== Port 8080 ==='
  sudo ss -tlnp | grep 8080
  
  echo ''
  echo '=== Nginx Config ==='
  sudo cat /etc/nginx/sites-enabled/backend-api | grep -E 'listen|server_name'
"

# Frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  echo '=== Nginx Status ==='
  sudo systemctl status nginx --no-pager -l
  
  echo ''
  echo '=== Nginx Listening Ports ==='
  sudo ss -tlnp | grep nginx
  
  echo ''
  echo '=== Frontend Files ==='
  ls -la /var/www/html/ | head -5
"
```

### 2. Fix Nginx HTTPS Configuration

**Jika port 443 tidak listening, kemungkinan Nginx tidak dikonfigurasi untuk HTTPS.**

**Backend VM - Update Nginx config untuk HTTPS:**

```bash
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Cek SSL certificate
sudo certbot certificates

# Update Nginx config dengan HTTPS block
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
# HTTP server - redirect to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api-pedeve-dev.aretaamany.com;

    ssl_certificate /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/privkey.pem;

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

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# Test config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Verify port 443 listening
sudo ss -tlnp | grep 443
```

**Frontend VM - Update Nginx config untuk HTTPS:**

```bash
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Cek SSL certificate
sudo certbot certificates

# Update Nginx config dengan HTTPS block
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

# Test config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Verify port 443 listening
sudo ss -tlnp | grep 443
```

### 3. Fix Firewall Rules

**Pastikan firewall rule untuk port 443 aktif:**

```bash
# Cek firewall rules
gcloud compute firewall-rules list --project=pedeve-pertamina-dms | grep -E 'allow-https|443'

# Jika tidak ada, create firewall rule
gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic" \
  --project=pedeve-pertamina-dms

# Apply tag ke VMs
gcloud compute instances add-tags backend-dev \
  --tags=https-server \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms

gcloud compute instances add-tags frontend-dev \
  --tags=https-server \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms
```

### 4. Restart Services

**Jika services tidak running, restart:**

```bash
# Backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  sudo systemctl restart nginx
  sudo docker restart dms-backend-prod
  sleep 5
  sudo systemctl status nginx --no-pager -l
  sudo docker ps | grep dms-backend-prod
"

# Frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  sudo systemctl restart nginx
  sleep 5
  sudo systemctl status nginx --no-pager -l
"
```

## Quick Fix Script

**Jalankan script ini untuk fix semua issues sekaligus:**

```bash
# Backend
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  sudo systemctl enable nginx
  sudo systemctl start nginx
  sudo docker start dms-backend-prod || sudo docker restart dms-backend-prod
  sudo systemctl status nginx --no-pager -l
  sudo docker ps | grep dms-backend-prod
  sudo ss -tlnp | grep -E ':(80|443|8080)'
"

# Frontend
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms --command="
  sudo systemctl enable nginx
  sudo systemctl start nginx
  sudo systemctl status nginx --no-pager -l
  sudo ss -tlnp | grep -E ':(80|443)'
"
```

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Run diagnostic script untuk cek status services
2. Fix Nginx HTTPS configuration jika port 443 tidak listening
3. Verify firewall rules untuk port 443
4. Restart services jika perlu

