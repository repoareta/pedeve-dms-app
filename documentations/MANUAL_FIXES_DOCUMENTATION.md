# 📝 Dokumentasi Perbaikan Manual BE dan FE

Dokumentasi lengkap tentang semua perbaikan manual yang dilakukan untuk backend dan frontend.

## 🎯 Summary Issues yang Ditemukan

### Frontend Issues:
1. **SSL Certificate** - Certificate sudah ada tapi tidak terpasang di Nginx config
2. **Nginx Config** - Config tidak punya `server_name pedeve-dev.aretaamany.com` dan HTTPS block (port 443)
3. **Port 443** - Tidak listening karena config tidak ada HTTPS block

### Backend Issues:
1. **SSL Certificate** - Tidak ada SSL certificate untuk `api-pedeve-dev.aretaamany.com`
2. **Port 443** - Tidak listening karena tidak ada SSL
3. **Docker Container** - Container tidak running setelah deployment
4. **Nginx Config** - Config mungkin sudah di-setup manual dengan SSL (jika ada)

---

## 🔧 Frontend Manual Fixes

### Issue 1: SSL Certificate Tidak Terpasang

**Problem:**
- Certificate ada di `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- Tapi Nginx config tidak punya HTTPS block
- Port 443 tidak listening

**Solution:**
```bash
# 1. Check certificate location
sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

# 2. Update Nginx config dengan SSL
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

# 3. Test config
sudo nginx -t

# 4. Reload Nginx
sudo systemctl reload nginx

# 5. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost
```

**Result:**
- ✅ Port 443 listening
- ✅ HTTPS accessible
- ✅ HTTP redirect to HTTPS

### Frontend Configuration Details:

**VM Info:**
- Name: `frontend-dev`
- IP: `34.128.123.1`
- Domain: `pedeve-dev.aretaamany.com`
- Zone: `asia-southeast2-a`

**Ports:**
- Port 80: HTTP (redirect to HTTPS)
- Port 443: HTTPS (SSL)

**SSL Certificate:**
- Location: `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- Files:
  - `fullchain.pem`
  - `privkey.pem`
- Provider: Let's Encrypt
- Auto-renewal: Enabled

**Nginx Config:**
- File: `/etc/nginx/sites-available/default`
- Enabled: `/etc/nginx/sites-enabled/default`
- Server Name: `pedeve-dev.aretaamany.com`
- Root: `/var/www/html`
- SPA Routing: Enabled (try_files to index.html)

---

## 🔧 Backend Manual Fixes

### Issue 1: SSL Certificate Tidak Ada

**Problem:**
- Backend tidak punya SSL certificate
- Port 443 tidak listening
- Frontend tidak bisa akses via HTTPS

**Solution:**
```bash
# 1. Install Certbot (jika belum ada)
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# 2. Run Certbot untuk mendapatkan SSL certificate
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com --non-interactive --agree-tos --email info@aretaamany.com

# 3. Verify certificate
sudo ls -la /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/

# 4. Test Nginx config
sudo nginx -t

# 5. Reload Nginx
sudo systemctl reload nginx

# 6. Verify
sudo ss -tlnp | grep 443
curl -I https://localhost/health
curl -I https://api-pedeve-dev.aretaamany.com/health
```

**Result:**
- ✅ SSL certificate created
- ✅ Port 443 listening
- ✅ HTTPS accessible
- ✅ HTTP redirect to HTTPS

### Issue 2: Docker Container Tidak Running

**Problem:**
- Container tidak running setelah deployment
- Backend tidak accessible

**Solution:**
```bash
# 1. Check container status
sudo docker ps -a | grep dms-backend-prod

# 2. Start container
sudo docker start dms-backend-prod

# 3. Check logs
sudo docker logs --tail 30 dms-backend-prod

# 4. Verify
sudo docker ps | grep dms-backend-prod
curl http://127.0.0.1:8080/health
```

**Result:**
- ✅ Container running
- ✅ Port 8080 listening
- ✅ Health check OK

### Backend Configuration Details:

**VM Info:**
- Name: `backend-dev`
- IP: `34.101.49.147`
- Domain: `api-pedeve-dev.aretaamany.com`
- Zone: `asia-southeast2-a`

**Ports:**
- Port 8080: Backend application (direct)
- Port 80: HTTP (Nginx reverse proxy, redirect to HTTPS)
- Port 443: HTTPS (Nginx reverse proxy with SSL)

**SSL Certificate:**
- Location: `/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/`
- Files:
  - `fullchain.pem`
  - `privkey.pem`
- Provider: Let's Encrypt
- Auto-renewal: Enabled

**Docker Container:**
- Name: `dms-backend-prod`
- Image: `ghcr.io/repoareta/dms-backend:latest`
- Network: `host` (untuk akses Cloud SQL Proxy di 127.0.0.1:5432)
- Restart Policy: `unless-stopped`

**Nginx Config:**
- File: `/etc/nginx/sites-available/backend-api`
- Enabled: `/etc/nginx/sites-enabled/backend-api`
- Server Name: `api-pedeve-dev.aretaamany.com`
- Proxy Pass: `http://127.0.0.1:8080`
- Reverse Proxy: Yes

**Database Connection:**
- Type: Cloud SQL PostgreSQL
- Connection: Via Cloud SQL Auth Proxy
- Proxy Port: `127.0.0.1:5432`
- Database: `db_dev_pedeve`
- User: `pedeve_user_db`

---

## 📋 Checklist Konfigurasi yang Harus Dipertahankan

### Frontend:
- ✅ SSL Certificate di `/etc/letsencrypt/live/pedeve-dev.aretaamany.com/`
- ✅ Nginx config dengan HTTPS block (port 443)
- ✅ Server name: `pedeve-dev.aretaamany.com`
- ✅ HTTP to HTTPS redirect
- ✅ SPA routing configuration
- ✅ Static files di `/var/www/html`

### Backend:
- ✅ SSL Certificate di `/etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/`
- ✅ Nginx config dengan HTTPS block (port 443)
- ✅ Server name: `api-pedeve-dev.aretaamany.com`
- ✅ HTTP to HTTPS redirect
- ✅ Reverse proxy ke `http://127.0.0.1:8080`
- ✅ Docker container `dms-backend-prod`
- ✅ Cloud SQL Proxy connection

### Network:
- ✅ Firewall rules untuk port 80, 443, 8080
- ✅ Domain DNS pointing ke IP yang benar
- ✅ CORS configuration di backend

---

## 🚫 Yang TIDAK BOLEH Di-Reset oleh Deployment

1. **SSL Certificates** - Jangan pernah hapus atau overwrite
2. **Nginx Config dengan SSL** - Jangan overwrite jika sudah benar
3. **Firewall Rules** - Jangan reset
4. **Domain DNS** - Jangan ubah
5. **Cloud SQL Proxy** - Jangan reset
6. **Docker Container Config** - Jangan reset network mode (host)

---

## ✅ Yang BOLEH Di-Update oleh Deployment

1. **Docker Container Image** - Selalu update dengan image baru
2. **Frontend Static Files** - Selalu update dengan build terbaru
3. **Nginx Config tanpa SSL** - Boleh update jika belum ada SSL
4. **Environment Variables** - Boleh update untuk container

---

## 🔍 Verification Commands

### Frontend:
```bash
# Check SSL certificate
sudo ls -la /etc/letsencrypt/live/pedeve-dev.aretaamany.com/

# Check ports
sudo ss -tlnp | grep -E ':(80|443)'

# Check Nginx config
sudo nginx -t
sudo cat /etc/nginx/sites-available/default | grep -A 5 "server_name"

# Test HTTPS
curl -I https://pedeve-dev.aretaamany.com
```

### Backend:
```bash
# Check SSL certificate
sudo ls -la /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/

# Check ports
sudo ss -tlnp | grep -E ':(80|443|8080)'

# Check container
sudo docker ps | grep dms-backend-prod

# Check Nginx config
sudo nginx -t
sudo cat /etc/nginx/sites-available/backend-api | grep -A 5 "server_name"

# Test HTTPS
curl -I https://api-pedeve-dev.aretaamany.com/health
```

---

## 📝 Notes

- Semua perbaikan manual ini harus dipertahankan oleh deployment scripts
- Deployment scripts harus idempotent dan check konfigurasi yang sudah ada
- Jangan pernah reset SSL certificates atau Nginx config yang sudah benar
- Selalu backup config sebelum update (jika perlu)

