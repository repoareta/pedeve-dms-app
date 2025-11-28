# 🔧 Fix Nginx Reload Error

## Error yang Terjadi
```
Job for nginx.service failed.
See "systemctl status nginx.service" and "journalctl -xeu nginx.service" for details.
```

## Langkah Troubleshooting

### 1. Cek Status Nginx
```bash
sudo systemctl status nginx.service
```

### 2. Cek Error Log Detail
```bash
sudo journalctl -xeu nginx.service --no-pager | tail -50
```

### 3. Test Nginx Config
```bash
sudo nginx -t
```

Ini akan menunjukkan syntax error jika ada.

## Kemungkinan Masalah

### A. Syntax Error di Config
Jika `nginx -t` menunjukkan syntax error, kemungkinan ada masalah di config file.

**Fix:** Edit config manual atau perbaiki script.

### B. Port 443 Sudah Digunakan
Jika port 443 sudah digunakan oleh service lain.

**Cek:**
```bash
sudo ss -tlnp | grep 443
```

**Fix:** Stop service yang menggunakan port 443, atau ubah config Nginx.

### C. SSL Certificate Path Tidak Valid
Jika Certbot belum generate certificate tapi config sudah reference path certificate.

**Fix:** Comment out SSL certificate lines dulu, generate certificate, baru uncomment.

## Quick Fix Script

Jika ada syntax error, jalankan script ini untuk fix:

```bash
# Backup config saat ini
sudo cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.backup

# Test config
sudo nginx -t

# Jika ada error, lihat output dan fix manual
# Atau restore backup
sudo cp /etc/nginx/sites-available/default.backup /etc/nginx/sites-available/default
sudo nginx -t
```

## Alternative: Setup SSL Step by Step

Jika script gagal, setup manual step by step:

### Step 1: Install Certbot
```bash
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx
```

### Step 2: Setup Nginx Config untuk HTTP dulu (tanpa SSL)
```bash
sudo tee /etc/nginx/sites-available/default > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name pedeve-dev.aretaamany.com _;

    root /var/www/html;
    index index.html;

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

# Test config
sudo nginx -t

# Reload jika test berhasil
sudo systemctl reload nginx
```

### Step 3: Generate SSL Certificate dengan Certbot
```bash
sudo certbot --nginx \
  -d pedeve-dev.aretaamany.com \
  --email info@aretaamany.com \
  --agree-tos \
  --non-interactive \
  --redirect
```

Certbot akan otomatis update Nginx config untuk HTTPS.

### Step 4: Verifikasi
```bash
# Test HTTPS
curl https://pedeve-dev.aretaamany.com/health

# Test HTTP redirect
curl -I http://pedeve-dev.aretaamany.com/health
```

