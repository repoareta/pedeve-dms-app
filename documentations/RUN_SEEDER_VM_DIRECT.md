# 🌱 Menjalankan Seeder di VM Backend - Langsung

## Masalah
- Container tidak punya source code seeder di `/app`
- `gcloud compute ssh` tidak bisa dijalankan dari dalam VM

## Solusi: Copy Seeder ke VM dan Jalankan Langsung

### Step 1: Copy Seeder dari Local Machine ke VM

**Dari local machine (bukan dari dalam VM):**

```bash
# Copy seeder directory ke VM
gcloud compute scp \
  --zone=asia-southeast2-a \
  --recurse \
  backend/cmd/seed-companies \
  backend-dev:~/seed-companies

# Copy go.mod dan go.sum untuk dependencies
gcloud compute scp \
  --zone=asia-southeast2-a \
  backend/go.mod \
  backend/go.sum \
  backend-dev:~/
```

### Step 2: SSH ke VM dan Setup

**SSH ke VM:**
```bash
gcloud compute ssh --zone=asia-southeast2-a backend-dev
```

**Di dalam VM, install Go jika belum ada:**
```bash
# Cek apakah Go sudah terinstall
go version

# Jika belum, install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### Step 3: Setup Environment dan Jalankan Seeder

**Di dalam VM:**

```bash
# Masuk ke directory seeder
cd ~/seed-companies

# Copy go.mod dan go.sum ke directory seeder
cp ~/go.mod ~/go.sum .

# Get database password dari Secret Manager
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})

# URL-encode password (untuk handle special characters)
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Set DATABASE_URL (menggunakan Cloud SQL Proxy di 127.0.0.1:5432)
export DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Download dependencies
go mod download

# Jalankan seeder
go run .
```

## Alternative: Buat Script di VM

**Atau buat script untuk memudahkan:**

```bash
# Di dalam VM, buat script
cat > ~/run-seeder.sh << 'EOF'
#!/bin/bash
set -e

cd ~/seed-companies
cp ~/go.mod ~/go.sum .

export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")
export DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

go mod download
go run .
EOF

chmod +x ~/run-seeder.sh

# Jalankan
~/run-seeder.sh
```

## Verifikasi

Setelah seeder selesai:

```bash
# Connect ke database
psql "${DATABASE_URL}"

# Di dalam psql:
SELECT id, name, code, level FROM companies ORDER BY level, name;
SELECT username, email, company_id FROM users WHERE username LIKE 'admin.%';
\q
```

