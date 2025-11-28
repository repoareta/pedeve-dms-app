# 🔍 Quick Check Services Status

## Masalah

Setelah deployment, services tidak bisa diakses. Untuk check status, kita perlu SSH ke VM dan jalankan command langsung (tidak bisa pakai `gcloud compute ssh` dari dalam VM).

## Solusi: Script Status Check

Saya sudah buat script yang bisa dijalankan langsung di VM untuk check status services.

### Backend VM

**SSH ke backend VM:**
```bash
# Dari local machine
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
```

**Di dalam backend VM, jalankan:**
```bash
# Copy script ke VM (jika belum ada)
# Atau jalankan command langsung:

# Check Nginx
sudo systemctl status nginx --no-pager -l | head -15

# Check Docker container
sudo docker ps -a | grep dms-backend-prod

# Check listening ports
sudo ss -tlnp | grep -E ':(80|443|8080)'

# Check container health
curl -s http://127.0.0.1:8080/health

# Check Nginx config
sudo nginx -t

# Check Nginx error log
sudo tail -20 /var/log/nginx/error.log
```

**Atau gunakan script lengkap:**
```bash
# Copy script dari local ke VM
gcloud compute scp --zone=asia-southeast2-a --project=pedeve-pertamina-dms \
  scripts/check-backend-status.sh backend-dev:~/

# SSH ke VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Di dalam VM, jalankan:
bash ~/check-backend-status.sh
```

### Frontend VM

**SSH ke frontend VM:**
```bash
# Dari local machine
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
```

**Di dalam frontend VM, jalankan:**
```bash
# Check Nginx
sudo systemctl status nginx --no-pager -l | head -15

# Check listening ports
sudo ss -tlnp | grep -E ':(80|443)'

# Check frontend files
ls -la /var/www/html/ | head -10

# Check local health
curl -s http://127.0.0.1/health

# Check Nginx config
sudo nginx -t

# Check Nginx error log
sudo tail -20 /var/log/nginx/error.log
```

**Atau gunakan script lengkap:**
```bash
# Copy script dari local ke VM
gcloud compute scp --zone=asia-southeast2-a --project=pedeve-pertamina-dms \
  scripts/check-frontend-status.sh frontend-dev:~/

# SSH ke VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Di dalam VM, jalankan:
bash ~/check-frontend-status.sh
```

## Quick Fix Commands

### Jika Nginx tidak running:

**Backend:**
```bash
sudo systemctl enable nginx
sudo systemctl start nginx
sudo systemctl status nginx
```

**Frontend:**
```bash
sudo systemctl enable nginx
sudo systemctl start nginx
sudo systemctl status nginx
```

### Jika Container tidak running:

**Backend:**
```bash
sudo docker ps -a | grep dms-backend-prod
sudo docker start dms-backend-prod
# Atau restart
sudo docker restart dms-backend-prod
sudo docker logs --tail 50 dms-backend-prod
```

### Jika Port 443 tidak listening:

**Cek SSL certificate:**
```bash
sudo certbot certificates
```

**Cek Nginx config untuk HTTPS:**
```bash
sudo cat /etc/nginx/sites-enabled/backend-api | grep -E 'listen.*443|ssl_certificate'
# Atau untuk frontend:
sudo cat /etc/nginx/sites-enabled/default | grep -E 'listen.*443|ssl_certificate'
```

**Jika HTTPS config tidak ada, reload Nginx config:**
```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Expected Output

### Backend VM - Ports Listening:
```
LISTEN  0  128  0.0.0.0:80  0.0.0.0:*  users:(("nginx",pid=1234,fd=6))
LISTEN  0  128  0.0.0.0:443  0.0.0.0:*  users:(("nginx",pid=1234,fd=7))
LISTEN  0  128  0.0.0.0:8080  0.0.0.0:*  users:(("docker-proxy",pid=5678,fd=4))
```

### Frontend VM - Ports Listening:
```
LISTEN  0  128  0.0.0.0:80  0.0.0.0:*  users:(("nginx",pid=1234,fd=6))
LISTEN  0  128  0.0.0.0:443  0.0.0.0:*  users:(("nginx",pid=1234,fd=7))
```

### Container Status:
```
CONTAINER ID   IMAGE                                    STATUS         PORTS   NAMES
abc123def456   ghcr.io/repoareta/dms-backend:latest     Up 5 minutes           dms-backend-prod
```

## Troubleshooting

**Jika port 443 tidak listening:**
1. Cek SSL certificate ada: `sudo certbot certificates`
2. Cek Nginx config punya HTTPS block
3. Reload Nginx: `sudo systemctl reload nginx`

**Jika container tidak running:**
1. Cek logs: `sudo docker logs --tail 50 dms-backend-prod`
2. Start container: `sudo docker start dms-backend-prod`
3. Cek database connection (Cloud SQL Proxy)

**Jika Nginx tidak running:**
1. Enable: `sudo systemctl enable nginx`
2. Start: `sudo systemctl start nginx`
3. Cek error log: `sudo tail -20 /var/log/nginx/error.log`

