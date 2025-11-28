# 🔧 Fix: Memastikan Services Tetap Running Setelah Deployment

## Masalah

Setelah deployment selesai di GitHub Actions, server mati dan aplikasi tidak bisa diakses:
- Frontend tidak bisa diakses (`ERR_CONNECTION_REFUSED`)
- Backend container mati
- Nginx tidak running

## Penyebab

1. **Nginx tidak di-enable** untuk auto-start saat boot
2. **Nginx tidak di-start** setelah deployment selesai
3. **Tidak ada verifikasi** bahwa services masih running setelah deployment
4. **Container backend** tidak diverifikasi apakah masih running

## Solusi

### 1. Update Deployment Scripts

**`scripts/setup-nginx-frontend.sh`:**
- ✅ Tambahkan `systemctl enable nginx` untuk auto-start
- ✅ Tambahkan `systemctl start nginx` untuk memastikan running
- ✅ Tambahkan verifikasi bahwa Nginx aktif

**`scripts/setup-backend-nginx.sh`:**
- ✅ Tambahkan `systemctl enable nginx` untuk auto-start
- ✅ Tambahkan `systemctl start nginx` untuk memastikan running
- ✅ Tambahkan verifikasi bahwa Nginx aktif

**`scripts/deploy-backend-vm.sh`:**
- ✅ Tambahkan verifikasi container running setelah start
- ✅ Tambahkan check port 8080 listening
- ✅ Tambahkan error handling jika container gagal start

### 2. Update CI/CD Workflow

**`.github/workflows/ci-cd.yml`:**

**Backend Deployment:**
- ✅ Pastikan Nginx di-enable dan di-start setelah deployment
- ✅ Verifikasi container backend masih running
- ✅ Verifikasi Nginx masih running

**Frontend Deployment:**
- ✅ Pastikan Nginx di-enable dan di-start setelah deployment
- ✅ Verifikasi Nginx masih running

**Health Checks:**
- ✅ Improved frontend health check dengan verifikasi Nginx status
- ✅ Tambahkan final service verification step

### 3. Final Service Verification

**New Step:** `Final Service Verification`
- ✅ Check backend container status
- ✅ Check backend port 8080 listening
- ✅ Check backend Nginx status (active & enabled)
- ✅ Check frontend Nginx status (active & enabled)
- ✅ Check frontend files exist
- ✅ Check frontend port 80 listening

## Perubahan Detail

### Backend Deployment

```bash
# Setelah deploy-backend-vm.sh dan setup-backend-nginx.sh
sudo systemctl enable nginx
sudo systemctl start nginx || sudo systemctl restart nginx

# Verifikasi
if ! sudo systemctl is-active --quiet nginx; then
  echo '❌ ERROR: Nginx failed to start!'
  exit 1
fi

if ! sudo docker ps | grep -q dms-backend-prod; then
  echo '❌ ERROR: Backend container is not running!'
  exit 1
fi
```

### Frontend Deployment

```bash
# Setelah setup-nginx-frontend.sh
sudo systemctl enable nginx
sudo systemctl start nginx || sudo systemctl restart nginx

# Verifikasi
if ! sudo systemctl is-active --quiet nginx; then
  echo '❌ ERROR: Nginx failed to start!'
  exit 1
fi
```

### Container Verification

```bash
# Di deploy-backend-vm.sh
sleep 5  # Wait for container to start

if sudo docker ps | grep -q dms-backend-prod; then
  echo "✅ Backend container is running"
  
  if sudo ss -tlnp | grep -q ':8080'; then
    echo "✅ Backend is listening on port 8080"
  else
    echo "⚠️  WARNING: Port 8080 not listening yet"
    sudo docker logs --tail 20 dms-backend-prod
  fi
else
  echo "❌ ERROR: Container failed to start!"
  sudo docker logs --tail 50 dms-backend-prod
  exit 1
fi
```

## Hasil

Setelah fix ini:

1. ✅ **Nginx auto-start** saat VM boot (via `systemctl enable`)
2. ✅ **Nginx di-start** setelah setiap deployment
3. ✅ **Container backend** diverifikasi masih running
4. ✅ **Final verification** memastikan semua services running
5. ✅ **Pipeline akan fail** jika ada service yang tidak running

## Testing

Setelah deployment selesai, verifikasi manual:

```bash
# Backend VM
sudo systemctl status nginx
sudo docker ps | grep dms-backend-prod
sudo ss -tlnp | grep 8080

# Frontend VM
sudo systemctl status nginx
sudo ss -tlnp | grep ':80 '
ls -la /var/www/html/
```

## Status

✅ **Fixed** - Semua services sekarang dijamin running setelah deployment

**Date:** 2025-11-28

