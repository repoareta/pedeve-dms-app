# 🌱 Menjalankan Seeder di Production (GCP)

## Prerequisites

- Backend VM sudah terhubung ke Cloud SQL via Cloud SQL Proxy
- Database sudah ada dan schema sudah di-migrate
- Akses SSH ke backend VM

## Cara Menjalankan Seeder

### Opsi 1: Copy Script ke VM dan Jalankan (Recommended)

**Dari local machine:**

```bash
# 1. Copy seeder source code ke backend VM
gcloud compute scp \
  --zone=asia-southeast2-a \
  --recurse \
  backend/cmd/seed-companies \
  backend-dev:~/seed-companies

# 2. Copy go.mod dan go.sum (untuk dependencies)
gcloud compute scp \
  --zone=asia-southeast2-a \
  backend/go.mod \
  backend/go.sum \
  backend-dev:~/seed-companies/

# 3. SSH ke VM dan jalankan seeder
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "cd ~/seed-companies && \
   GCP_PROJECT_ID=pedeve-pertamina-dms \
   DATABASE_URL=\"postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable\" \
   go run ."
```

### Opsi 2: Jalankan dari Dalam Container Backend

**SSH ke backend VM:**

```bash
gcloud compute ssh --zone=asia-southeast2-a backend-dev
```

**Di dalam VM, jalankan seeder di container:**

```bash
# Masuk ke container backend
sudo docker exec -it dms-backend-prod sh

# Di dalam container, jalankan seeder
cd /app
go run ./cmd/seed-companies
```

**Atau langsung tanpa masuk container:**

```bash
# Jalankan seeder langsung di container
sudo docker exec -it dms-backend-prod sh -c "cd /app && go run ./cmd/seed-companies"
```

### Opsi 3: Build Binary dan Copy ke VM

**Dari local machine:**

```bash
# 1. Build binary untuk Linux
cd backend
GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

# 2. Copy binary ke VM
gcloud compute scp \
  --zone=asia-southeast2-a \
  seed-companies \
  backend-dev:~/seed-companies

# 3. SSH dan jalankan
gcloud compute ssh --zone=asia-southeast2-a backend-dev -- \
  "chmod +x ~/seed-companies && \
   GCP_PROJECT_ID=pedeve-pertamina-dms \
   DATABASE_URL=\"postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable\" \
   ~/seed-companies"
```

## Verifikasi

Setelah seeder selesai, verifikasi data:

```bash
# SSH ke backend VM
gcloud compute ssh --zone=asia-southeast2-a backend-dev

# Connect ke database via Cloud SQL Proxy
psql "postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Di dalam psql, cek companies
SELECT id, name, code, level, parent_id FROM companies ORDER BY level, name;

# Cek users
SELECT id, username, email, company_id FROM users WHERE username LIKE 'admin.%';

# Exit
\q
```

## Troubleshooting

### Error: "connection refused" atau "dial tcp 127.0.0.1:5432"

**Penyebab:** Cloud SQL Proxy tidak running.

**Fix:**
```bash
# Cek apakah Cloud SQL Proxy running
ps aux | grep cloud-sql-proxy

# Jika tidak running, start Cloud SQL Proxy
# (biasanya sudah di-setup sebagai systemd service)
sudo systemctl status cloud-sql-proxy
```

### Error: "password authentication failed"

**Penyebab:** Password di Secret Manager tidak sesuai.

**Fix:**
```bash
# Cek password dari Secret Manager
gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms

# Pastikan password sesuai dengan Cloud SQL user
```

### Error: "no such file or directory" saat go run

**Penyebab:** Dependencies belum di-download.

**Fix:**
```bash
# Di dalam container atau VM, download dependencies
cd /app  # atau ~/seed-companies
go mod download
```

## Expected Output

Setelah seeder berhasil, akan muncul output seperti:

```
🌱 Seeding Companies and Users

✅ Database initialized (GORM logging disabled for cleaner output)
✅ Connected to database
✅ Roles loaded
✅ Holding company created/updated: Pedeve Pertamina
✅ Level 1 company created/updated: PT Energi Nusantara
...
✅ Seeding completed!
```

## Catatan

- Seeder akan **skip** jika company dengan code yang sama sudah ada (idempotent)
- Jika ingin re-seed, hapus companies terlebih dahulu atau gunakan fresh database
- Semua user memiliki password default: `admin123`

