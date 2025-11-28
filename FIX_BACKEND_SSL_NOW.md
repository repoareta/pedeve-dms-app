# 🔧 Fix Backend SSL - Step by Step

Backend tidak punya SSL certificate, jadi HTTPS tidak bisa diakses. Frontend mencoba akses via `https://api-pedeve-dev.aretaamany.com` tapi gagal.

## Diagnosis Results:
- ✅ Container running
- ✅ Port 8080 listening
- ✅ Backend health check (direct) OK
- ✅ Nginx running
- ✅ Port 80 listening (HTTP)
- ❌ SSL certificate NOT found
- ❌ Port 443 NOT listening (HTTPS)
- ❌ HTTPS test failed

## Solution: Setup SSL Certificate

**Jalankan di backend VM:**

```bash
# 1. Install Certbot (jika belum ada)
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# 2. Run Certbot untuk mendapatkan SSL certificate
# Certbot akan otomatis update Nginx config untuk HTTPS
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com --non-interactive --agree-tos --email info@aretaamany.com

# Atau jika perlu interaktif:
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com
```

**Setelah SSL setup, verify:**

```bash
# 1. Check certificate
sudo ls -la /etc/letsencrypt/live/api-pedeve-dev.aretaamany.com/

# 2. Check port 443
sudo ss -tlnp | grep 443

# 3. Test Nginx config
sudo nginx -t

# 4. Reload Nginx
sudo systemctl reload nginx

# 5. Test HTTPS
curl -I https://localhost/health
curl -I https://api-pedeve-dev.aretaamany.com/health
```

**Jika Certbot gagal karena Nginx config tidak punya `server_name`:**

```bash
# 1. Check current config
sudo cat /etc/nginx/sites-available/backend-api | grep server_name

# 2. Update config untuk tambahkan server_name (jika belum ada)
# Pastikan ada: server_name api-pedeve-dev.aretaamany.com;

# 3. Test config
sudo nginx -t

# 4. Reload Nginx
sudo systemctl reload nginx

# 5. Run Certbot lagi
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com
```

**Atau gunakan script:**

```bash
# Copy script ke VM (dari local machine)
# Atau buat langsung di VM dengan script di atas

chmod +x ~/setup-backend-ssl-now.sh
~/setup-backend-ssl-now.sh
```

## Troubleshooting

### Certbot Error: "Could not automatically find a matching server block"

**Fix:** Update Nginx config untuk include `server_name api-pedeve-dev.aretaamany.com;`

```bash
# Check config
sudo cat /etc/nginx/sites-available/backend-api

# Jika tidak ada server_name, update config:
sudo nano /etc/nginx/sites-available/backend-api

# Pastikan ada:
# server_name api-pedeve-dev.aretaamany.com;

# Test dan reload
sudo nginx -t
sudo systemctl reload nginx

# Run Certbot lagi
sudo certbot --nginx -d api-pedeve-dev.aretaamany.com
```

### Certificate Already Exists but Not Installed

**Fix:** Install certificate yang sudah ada:

```bash
sudo certbot install --cert-name api-pedeve-dev.aretaamany.com
```

### Domain Not Pointing to VM

**Fix:** Pastikan domain `api-pedeve-dev.aretaamany.com` pointing ke IP `34.101.49.147`

```bash
# Test DNS
nslookup api-pedeve-dev.aretaamany.com
# Should return: 34.101.49.147
```

## After SSL Setup

✅ Backend akan bisa diakses via HTTPS  
✅ Frontend bisa fetch dari `https://api-pedeve-dev.aretaamany.com`  
✅ CORS akan bekerja dengan baik  
✅ Secure connection untuk semua API calls

