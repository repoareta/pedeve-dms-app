# 🌱 Build dan Jalankan Seeder Binary

## Solusi: Build Binary Seeder dan Copy ke Container

Karena container hanya punya binary (tidak ada source code), kita perlu:
1. Build binary seeder di local
2. Copy ke VM
3. Copy ke container
4. Jalankan di container

## Step 1: Build Binary Seeder di Local

**Dari local machine (di directory project):**

```bash
cd backend

# Build binary seeder untuk Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o seed-companies ./cmd/seed-companies

# Verifikasi binary
ls -lh seed-companies
file seed-companies
```

## Step 2: Copy Binary ke VM

**Dari local machine:**

```bash
# Copy binary ke VM
gcloud compute scp \
  --zone=asia-southeast2-a \
  backend/seed-companies \
  backend-dev:~/seed-companies
```

**Jika gcloud tidak tersedia, gunakan scp biasa:**

```bash
scp backend/seed-companies info@34.101.49.147:~/seed-companies
```

## Step 3: Copy Binary ke Container dan Jalankan

**SSH ke VM dan jalankan:**

```bash
# SSH ke VM
gcloud compute ssh --zone=asia-southeast2-a backend-dev

# Di dalam VM, copy binary ke container
sudo docker cp ~/seed-companies dms-backend-prod:/root/seed-companies

# Set executable permission
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

# Jalankan seeder (DATABASE_URL sudah ada di container)
sudo docker exec dms-backend-prod /root/seed-companies
```

## Alternative: Build Binary Langsung di VM

**Jika Go sudah terinstall di VM:**

```bash
# SSH ke VM
gcloud compute ssh --zone=asia-southeast2-a backend-dev

# Clone repo (jika belum)
git clone https://github.com/repoareta/pedeve-dms-app.git ~/pedeve-dms-app

# Build binary
cd ~/pedeve-dms-app/backend
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o seed-companies ./cmd/seed-companies

# Copy ke container
sudo docker cp seed-companies dms-backend-prod:/root/seed-companies
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

# Jalankan
sudo docker exec dms-backend-prod /root/seed-companies
```

## Quick Script (All-in-One)

**Dari local machine, jalankan script ini:**

```bash
#!/bin/bash
set -e

echo "🔨 Building seeder binary..."
cd backend
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o seed-companies ./cmd/seed-companies

echo "📤 Copying to VM..."
gcloud compute scp --zone=asia-southeast2-a seed-companies backend-dev:~/seed-companies

echo "🚀 Deploying to container..."
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "sudo docker cp ~/seed-companies dms-backend-prod:/root/seed-companies && \
   sudo docker exec dms-backend-prod chmod +x /root/seed-companies && \
   sudo docker exec dms-backend-prod /root/seed-companies"

echo "✅ Done!"
```

## Verifikasi

Setelah seeder selesai, verifikasi data:

```bash
# Connect ke database via container (jika punya psql)
# Atau dari VM langsung
psql "${DATABASE_URL}"

# Di dalam psql:
SELECT id, name, code, level FROM companies ORDER BY level, name;
SELECT username, email, company_id FROM users WHERE username LIKE 'admin.%';
\q
```

