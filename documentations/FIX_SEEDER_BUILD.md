# 🔧 Fix Seeder Build - Go Version Issue

## Masalah
- `go1.25` tidak tersedia untuk download
- Binary tidak ter-build
- File tidak ada di container

## Solusi: Install Go Manual dan Build

**Jalankan perintah berikut di VM:**

```bash
# Install Go manual (jika belum ada atau versi salah)
if ! command -v go &> /dev/null || go version | grep -q "go1.25"; then
  echo "📦 Installing Go 1.21..."
  wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  rm -f go1.21.5.linux-amd64.tar.gz
fi

# Verifikasi Go
go version

# Masuk ke directory backend
cd ~/pedeve-dms-app/backend

# Download dependencies
go mod download

# Build binary seeder (tanpa CGO untuk compatibility)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

# Verifikasi binary ter-build
ls -lh seed-companies
file seed-companies

# Copy binary ke container
sudo docker cp seed-companies dms-backend-prod:/root/seed-companies

# Set executable permission
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

# Verifikasi file ada di container
sudo docker exec dms-backend-prod ls -lh /root/seed-companies

# Jalankan seeder
echo "🚀 Running seeder..."
sudo docker exec dms-backend-prod /root/seed-companies
```

## Alternative: Build dengan Go yang Sudah Ada di Container

**Jika container punya Go (tapi kemungkinan tidak karena alpine):**

```bash
# Cek apakah container punya Go
sudo docker exec dms-backend-prod go version

# Jika ada, bisa build langsung di container
```

## Quick Fix Script

**Copy-paste semua ini sekaligus:**

```bash
#!/bin/bash
set -e

echo "📦 Installing/Updating Go..."
wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
rm -f go1.21.5.linux-amd64.tar.gz

echo "✅ Go version:"
go version

echo "📥 Downloading dependencies..."
cd ~/pedeve-dms-app/backend
go mod download

echo "🔨 Building seeder binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

echo "✅ Binary built:"
ls -lh seed-companies

echo "📤 Copying to container..."
sudo docker cp seed-companies dms-backend-prod:/root/seed-companies
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

echo "✅ Verifying in container:"
sudo docker exec dms-backend-prod ls -lh /root/seed-companies

echo "🚀 Running seeder..."
sudo docker exec dms-backend-prod /root/seed-companies
```

