# 🚀 Quick Restart Services

Jika backend atau frontend mati setelah deployment, gunakan script berikut:

## Opsi 1: Restart dari Local (dengan gcloud)

```bash
# Restart backend dan frontend
./scripts/restart-services-manual.sh both

# Atau hanya backend
./scripts/restart-services-manual.sh backend

# Atau hanya frontend
./scripts/restart-services-manual.sh frontend
```

## Opsi 2: Restart Langsung di VM (Recommended)

### Backend VM

```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Copy script ke VM
# (dari local machine)
gcloud compute scp --zone=asia-southeast2-a --project=pedeve-pertamina-dms \
  scripts/restart-services-on-vm.sh backend-dev:~/

# Jalankan di VM
chmod +x ~/restart-services-on-vm.sh
~/restart-services-on-vm.sh backend
```

### Frontend VM

```bash
# SSH ke frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms

# Copy script ke VM
# (dari local machine)
gcloud compute scp --zone=asia-southeast2-a --project=pedeve-pertamina-dms \
  scripts/restart-services-on-vm.sh frontend-dev:~/

# Jalankan di VM
chmod +x ~/restart-services-on-vm.sh
~/restart-services-on-vm.sh frontend
```

## Opsi 3: Manual Commands

### Backend VM

```bash
# Restart Docker container
sudo docker restart dms-backend-prod

# Restart Nginx
sudo systemctl restart nginx

# Check status
sudo docker ps | grep dms-backend-prod
sudo systemctl status nginx
sudo ss -tlnp | grep -E ':(80|443|8080)'
```

### Frontend VM

```bash
# Restart Nginx
sudo systemctl restart nginx

# Check status
sudo systemctl status nginx
sudo ss -tlnp | grep -E ':(80|443)'
ls -la /var/www/html/ | head -10
```

## Verify Services

```bash
# Backend
curl http://34.101.49.147:8080/health
curl https://api-pedeve-dev.aretaamany.com/health

# Frontend
curl http://34.128.123.1
curl https://pedeve-dev.aretaamany.com
```

## Troubleshooting

### Backend container tidak running

```bash
# Check container status
sudo docker ps -a | grep dms-backend-prod

# Check logs
sudo docker logs --tail 50 dms-backend-prod

# Start container
sudo docker start dms-backend-prod

# Jika container tidak ada, check images
sudo docker images | grep dms-backend
```

### Nginx tidak running

```bash
# Check status
sudo systemctl status nginx

# Check config
sudo nginx -t

# Restart
sudo systemctl restart nginx

# Enable on boot
sudo systemctl enable nginx
```

### Port tidak listening

```bash
# Check ports
sudo ss -tlnp | grep -E ':(80|443|8080)'

# Check firewall
sudo ufw status
```

