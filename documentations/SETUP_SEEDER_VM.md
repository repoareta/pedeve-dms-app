# 🌱 Setup Seeder di VM - Tanpa gcloud

## Opsi 1: Clone Repo dari GitHub (Recommended)

**Di dalam VM backend, jalankan:**

```bash
# Install git jika belum ada
sudo apt-get update
sudo apt-get install -y git

# Clone repo (ganti dengan URL repo Anda jika berbeda)
git clone https://github.com/repoareta/pedeve-dms-app.git ~/pedeve-dms-app

# Masuk ke directory seeder
cd ~/pedeve-dms-app/backend/cmd/seed-companies

# Copy go.mod dan go.sum
cp ~/pedeve-dms-app/backend/go.mod ~/pedeve-dms-app/backend/go.sum .

# Setup environment dan jalankan
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")
export DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Install Go jika belum ada
if ! command -v go &> /dev/null; then
  wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
fi

# Download dependencies dan jalankan
go mod download
go run .
```

## Opsi 2: Download File dari GitHub Raw URL

**Jika repo public, download langsung:**

```bash
# Buat directory
mkdir -p ~/seed-companies
cd ~/seed-companies

# Download main.go dari GitHub (ganti dengan URL yang benar)
curl -o main.go https://raw.githubusercontent.com/repoareta/pedeve-dms-app/development/backend/cmd/seed-companies/main.go

# Download go.mod dan go.sum
curl -o ../go.mod https://raw.githubusercontent.com/repoareta/pedeve-dms-app/development/backend/go.mod
curl -o ../go.sum https://raw.githubusercontent.com/repoareta/pedeve-dms-app/development/backend/go.sum

# Copy go.mod dan go.sum ke directory seeder
cp ../go.mod ../go.sum .

# Setup environment dan jalankan (sama seperti Opsi 1)
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")
export DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Install Go jika belum ada
if ! command -v go &> /dev/null; then
  wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
fi

# Download dependencies dan jalankan
go mod download
go run .
```

## Opsi 3: Copy Manual via SCP (jika punya akses SSH)

**Dari local machine dengan SSH access:**

```bash
# Jika punya SSH key untuk VM
scp -r backend/cmd/seed-companies info@34.101.49.147:~/seed-companies
scp backend/go.mod backend/go.sum info@34.101.49.147:~/
```

## Quick Script (All-in-One)

**Copy-paste script ini di VM:**

```bash
#!/bin/bash
set -e

echo "🌱 Setting up seeder..."

# Install dependencies
sudo apt-get update
sudo apt-get install -y git curl

# Clone repo
if [ ! -d ~/pedeve-dms-app ]; then
  git clone https://github.com/repoareta/pedeve-dms-app.git ~/pedeve-dms-app
fi

cd ~/pedeve-dms-app/backend/cmd/seed-companies
cp ~/pedeve-dms-app/backend/go.mod ~/pedeve-dms-app/backend/go.sum .

# Install Go if needed
if ! command -v go &> /dev/null; then
  echo "📦 Installing Go..."
  wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
  export PATH=$PATH:/usr/local/go/bin
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# Setup database connection
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")
export DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Download dependencies and run
echo "📥 Downloading dependencies..."
go mod download

echo "🚀 Running seeder..."
go run .
```

**Jalankan:**
```bash
# Save script
cat > ~/setup-and-run-seeder.sh << 'SCRIPT_EOF'
# ... paste script di atas ...
SCRIPT_EOF

chmod +x ~/setup-and-run-seeder.sh
~/setup-and-run-seeder.sh
```

