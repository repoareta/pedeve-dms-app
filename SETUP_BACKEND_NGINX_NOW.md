# Setup Nginx Reverse Proxy Sekarang (Manual)

## Masalah
Backend bisa diakses via IP: `http://34.101.49.147:8080` ✅
Tapi domain error: `http://api-pedeve-dev.aretaamany.com` ❌ (default ke port 80, bukan 8080)

## Solusi: Setup Nginx Reverse Proxy

Nginx akan forward port 80 → backend port 8080, sehingga domain bisa diakses tanpa port.

### Cara Setup (Manual - Sekarang)

**SSH ke backend VM:**
```bash
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
```

**Copy script ke VM:**
```bash
# Dari local machine, copy script
gcloud compute scp \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms \
  scripts/setup-backend-nginx.sh \
  backend-dev:~/
```

**Atau buat script langsung di VM:**
```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a

# Install Nginx
sudo apt-get update
sudo apt-get install -y nginx

# Create config
sudo tee /etc/nginx/sites-available/backend-api > /dev/null <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api-pedeve-dev.aretaamany.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api
sudo rm -f /etc/nginx/sites-enabled/default

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

**Test:**
```bash
curl http://api-pedeve-dev.aretaamany.com/health
curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token
```

### Setelah Deployment Berikutnya

Deployment workflow akan otomatis setup Nginx, jadi tidak perlu manual lagi.

## Verifikasi

```bash
# Test via domain (tanpa port)
curl http://api-pedeve-dev.aretaamany.com/health
curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token

# Harus return JSON response
```

## Troubleshooting

**Jika Nginx error:**
```bash
# Cek config
sudo nginx -t

# Cek logs
sudo tail -f /var/log/nginx/backend-api-error.log

# Restart Nginx
sudo systemctl restart nginx
```

**Jika masih tidak bisa:**
- Pastikan firewall rule untuk port 80 sudah ada (allow-http)
- Pastikan VM sudah di-tag dengan `http-server`
- Cek apakah backend container running: `sudo docker ps | grep dms-backend-prod`

