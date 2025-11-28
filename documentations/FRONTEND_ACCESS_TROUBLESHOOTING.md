# Frontend Access Troubleshooting

## Masalah: Frontend tidak bisa diakses via domain atau IP

### 1. Akses via IP Address

**Frontend VM IP:** `34.128.123.1` (dari DEPLOYMENT_CONFIG.md)

**Cara test:**
```bash
# Test dari local
curl http://34.128.123.1

# Atau buka di browser
http://34.128.123.1
```

**Jika tidak bisa akses, cek:**

#### A. Firewall Rules di GCP
1. Buka GCP Console → VPC Network → Firewall
2. Pastikan ada rule yang allow:
   - **HTTP (port 80)** dari `0.0.0.0/0` (atau IP tertentu)
   - **HTTPS (port 443)** dari `0.0.0.0/0` (jika pakai SSL)

**Cara buat firewall rule:**
```bash
# Allow HTTP
gcloud compute firewall-rules create allow-http \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server \
  --description "Allow HTTP traffic"

# Allow HTTPS
gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic"
```

**Apply tag ke VM:**
```bash
gcloud compute instances add-tags frontend-dev \
  --tags http-server,https-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

#### B. Nginx Status
SSH ke frontend VM dan cek:
```bash
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Cek Nginx status
sudo systemctl status nginx

# Cek apakah port 80 listening
sudo ss -tlnp | grep 80

# Test dari dalam VM
curl http://localhost
curl http://127.0.0.1
```

#### C. File di /var/www/html
```bash
# Cek apakah file sudah di-deploy
ls -la /var/www/html

# Harus ada index.html
ls -la /var/www/html/index.html
```

### 2. Akses via Domain

**Domain:** `https://pedeve-dev.aretaamany.com/`

**Cek domain pointing:**
```bash
# Cek DNS record
dig pedeve-dev.aretaamany.com
nslookup pedeve-dev.aretaamany.com

# Harus resolve ke IP: 34.128.123.1
```

**Jika domain belum pointing:**
1. Buka DNS provider (dimana aretaamany.com di-manage)
2. Tambah A record:
   - **Name:** `pedeve-dev`
   - **Type:** `A`
   - **Value:** `34.128.123.1`
   - **TTL:** `300` (5 menit)

**SSL Certificate:**
- Jika pakai HTTPS, perlu setup SSL certificate
- Bisa pakai Let's Encrypt dengan Certbot
- Atau pakai GCP Load Balancer dengan managed SSL

### 3. Nginx Configuration

**Cek config:**
```bash
# SSH ke frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a

# Cek config
sudo cat /etc/nginx/sites-available/default

# Test config
sudo nginx -t

# Reload jika perlu
sudo systemctl reload nginx
```

**Config yang benar untuk SPA:**
```nginx
server {
    listen 80;
    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### 4. Manual Setup Nginx

Jika deployment script belum setup Nginx dengan benar:

```bash
# SSH ke frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a

# Copy setup script
# (script sudah di-copy oleh deployment, tapi bisa manual juga)

# Run setup script
chmod +x ~/setup-nginx-frontend.sh
~/setup-nginx-frontend.sh
```

### 5. Test dari VM

```bash
# SSH ke frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a

# Test localhost
curl http://localhost
curl http://127.0.0.1

# Test dengan IP external
curl http://34.128.123.1

# Cek Nginx logs jika ada error
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### 6. Quick Fix Commands

```bash
# 1. Setup firewall (jika belum)
gcloud compute firewall-rules create allow-http \
  --allow tcp:80 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server

gcloud compute instances add-tags frontend-dev \
  --tags http-server \
  --zone asia-southeast2-a

# 2. Restart Nginx
gcloud compute ssh frontend-dev --zone=asia-southeast2-a \
  --command="sudo systemctl restart nginx"

# 3. Test
curl http://34.128.123.1
```

## Checklist

- [ ] Firewall rule untuk HTTP (port 80) sudah dibuat
- [ ] VM sudah di-tag dengan `http-server`
- [ ] Nginx sudah installed dan running
- [ ] File sudah di-deploy ke `/var/www/html`
- [ ] Nginx config sudah benar (SPA routing)
- [ ] Domain sudah pointing ke IP yang benar
- [ ] SSL certificate sudah di-setup (jika pakai HTTPS)

