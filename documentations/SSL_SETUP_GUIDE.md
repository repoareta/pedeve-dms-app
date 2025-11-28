# SSL Certificate Setup Guide untuk Backend API

## Overview

Setup SSL certificate menggunakan Let's Encrypt dengan Certbot untuk domain `api-pedeve-dev.aretaamany.com`.

## Prerequisites

- Domain sudah pointing ke backend VM IP: `34.101.49.147`
- Nginx sudah terinstall dan running
- Port 80 dan 443 sudah di-allow di firewall
- VM sudah di-tag dengan `http-server` dan `https-server`

## Setup SSL Certificate

### Step 1: Pastikan Firewall Rules

```bash
# Allow HTTPS (port 443)
gcloud compute firewall-rules create allow-https \
  --allow tcp:443 \
  --source-ranges 0.0.0.0/0 \
  --target-tags https-server \
  --description "Allow HTTPS traffic" \
  --project pedeve-pertamina-dms

# Apply tag
gcloud compute instances add-tags backend-dev \
  --tags https-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

### Step 2: Setup SSL di Backend VM

**SSH ke backend VM:**
```bash
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
```

**Jalankan script setup SSL:**
```bash
# Copy script ke VM (dari local machine)
gcloud compute scp \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms \
  scripts/setup-backend-ssl.sh \
  backend-dev:~/

# SSH dan jalankan
gcloud compute ssh backend-dev --zone=asia-southeast2-a
chmod +x ~/setup-backend-ssl.sh
~/setup-backend-ssl.sh
```

**Atau manual setup:**
```bash
# Install Certbot
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Generate certificate
sudo certbot --nginx \
  -d api-pedeve-dev.aretaamany.com \
  --email info@aretaamany.com \
  --agree-tos \
  --non-interactive \
  --redirect

# Setup auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

### Step 3: Verifikasi

```bash
# Test HTTPS
curl https://api-pedeve-dev.aretaamany.com/health
curl https://api-pedeve-dev.aretaamany.com/api/v1/csrf-token

# Test HTTP redirect
curl -I http://api-pedeve-dev.aretaamany.com/health
# Harus return 301 redirect ke HTTPS
```

## Auto-Renewal

Certbot sudah di-setup dengan auto-renewal via systemd timer. Certificate akan otomatis di-renew sebelum expired (setiap 90 hari).

**Cek status auto-renewal:**
```bash
sudo systemctl status certbot.timer
```

**Test renewal (dry-run):**
```bash
sudo certbot renew --dry-run
```

## Troubleshooting

### Error: "Failed to connect to domain"

**Penyebab:** Domain belum pointing atau firewall block.

**Solusi:**
1. Cek DNS: `dig api-pedeve-dev.aretaamany.com`
2. Pastikan resolve ke: `34.101.49.147`
3. Pastikan firewall rule untuk port 443 sudah ada

### Error: "Challenge failed"

**Penyebab:** Let's Encrypt tidak bisa verify domain ownership.

**Solusi:**
1. Pastikan domain pointing ke IP yang benar
2. Pastikan port 80 accessible dari internet
3. Pastikan Nginx running dan bisa diakses

### Certificate Expired

**Renewal manual:**
```bash
sudo certbot renew
sudo systemctl reload nginx
```

## After SSL Setup

1. **Frontend sudah di-build dengan HTTPS API URL** ✅
2. **Backend sudah support HTTPS** ✅
3. **HTTP akan otomatis redirect ke HTTPS** ✅
4. **Frontend bisa connect ke backend via HTTPS** ✅

## Security Notes

- Certificate valid untuk 90 hari
- Auto-renewal sudah di-setup
- HSTS header sudah di-set untuk security
- HTTP akan redirect ke HTTPS

