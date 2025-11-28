# 🔧 Fix VM Repo - Checkout Development Branch

## Masalah
- Directory `cmd/seed-companies` tidak ada di repo yang di-clone
- Repo di VM mungkin di branch `main` atau belum ter-update

## Solusi: Checkout Branch Development

**Di VM, jalankan:**

```bash
cd ~/pedeve-dms-app

# Cek branch saat ini
git branch

# Checkout branch development
git checkout development

# Pull latest changes
git pull origin development

# Verifikasi seeder ada
ls -la backend/cmd/seed-companies/

# Kembali ke directory backend
cd backend

# Update go.mod untuk Go 1.21 (jika belum)
sed -i 's/go 1.25/go 1.21/' go.mod

# Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

# Verifikasi
ls -lh seed-companies

# Copy ke container
sudo docker cp seed-companies dms-backend-prod:/root/seed-companies
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

# Jalankan seeder
sudo docker exec dms-backend-prod /root/seed-companies
```

