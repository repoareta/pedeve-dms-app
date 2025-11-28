# 🌱 Build Seeder di Local Machine

## Masalah
- Go 1.25 tidak tersedia di VM
- `go.mod` require Go 1.25
- Build gagal di VM

## Solusi: Build di Local, Copy Binary ke VM

Karena local machine sudah punya Go yang sesuai, build di local lalu copy binary.

### Step 1: Build Binary di Local

**Dari local machine (di directory project):**

```bash
cd backend

# Build binary seeder untuk Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

# Verifikasi
ls -lh seed-companies
file seed-companies
```

### Step 2: Copy Binary ke VM

**Opsi A: Menggunakan scp (jika punya SSH access)**

```bash
# Copy binary ke VM
scp backend/seed-companies info@34.101.49.147:~/seed-companies
```

**Opsi B: Upload via Cloud Console atau cara lain**

Atau jika punya akses lain, upload file `backend/seed-companies` ke VM.

### Step 3: Copy ke Container dan Jalankan

**SSH ke VM dan jalankan:**

```bash
# Copy binary ke container
sudo docker cp ~/seed-companies dms-backend-prod:/root/seed-companies

# Set executable
sudo docker exec dms-backend-prod chmod +x /root/seed-companies

# Verifikasi
sudo docker exec dms-backend-prod ls -lh /root/seed-companies

# Jalankan seeder
sudo docker exec dms-backend-prod /root/seed-companies
```

## Alternative: Update go.mod untuk Support Go 1.21

**Jika ingin build di VM, update go.mod:**

```bash
# Di VM, edit go.mod
cd ~/pedeve-dms-app/backend
sed -i 's/go 1.25/go 1.21/' go.mod

# Build lagi
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies
```

Tapi cara ini bisa menyebabkan masalah jika ada fitur Go 1.25 yang digunakan.

## Recommended: Build di Local

**Cara termudah dan paling aman adalah build di local machine yang sudah punya Go 1.25.**

