# Panduan Migrasi ke PostgreSQL

## ✅ Status: Aplikasi Sudah Support PostgreSQL!

Aplikasi sudah siap menggunakan PostgreSQL. Driver dan konfigurasi sudah ada di kode. Tinggal set environment variable `DATABASE_URL`.

---

## 🚀 Cara Migrasi ke PostgreSQL

### Opsi 1: Menggunakan PostgreSQL di Docker (Recommended untuk Development)

#### 1. Update `docker-compose.dev.yml`

Tambahkan service PostgreSQL:

```yaml
services:
  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    container_name: dms-postgres-dev
    environment:
      - POSTGRES_USER=dms_user
      - POSTGRES_PASSWORD=dms_password
      - POSTGRES_DB=dms_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - dms-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dms_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Backend API (Go) - Development mode
  backend:
    image: golang:1.25-alpine
    container_name: dms-backend-dev
    working_dir: /app
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENV=development
      - CGO_ENABLED=0  # Tidak perlu CGO untuk PostgreSQL
      - DATABASE_URL=postgres://dms_user:dms_password@postgres:5432/dms_db?sslmode=disable
    volumes:
      - ./backend:/app
      - /go/pkg/mod
    command: sh -c "apk add --no-cache git && go mod download && go run ."
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - dms-network

  # Frontend (Vue + Vite) - Development mode with hot reload
  frontend:
    image: node:20-alpine
    container_name: dms-frontend-dev
    working_dir: /app
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - VITE_API_URL=http://localhost:8080/api/v1
      - NODE_ENV=development
    command: sh -c "npm install && npm run dev -- --host 0.0.0.0"
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - dms-network

volumes:
  postgres_data:

networks:
  dms-network:
    driver: bridge
```

#### 2. Start Services

```bash
# Stop services yang sedang berjalan
docker-compose -f docker-compose.dev.yml down

# Start dengan PostgreSQL
docker-compose -f docker-compose.dev.yml up --build
```

#### 3. Verifikasi

Backend akan otomatis:
- ✅ Connect ke PostgreSQL
- ✅ Auto-migrate schema (membuat tables: users, two_factor_auths, audit_logs)
- ✅ Seed superadmin user

Cek logs:
```bash
docker-compose -f docker-compose.dev.yml logs backend | grep -i "postgres\|database"
```

---

### Opsi 2: PostgreSQL Lokal (Tanpa Docker)

#### 1. Install PostgreSQL

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Windows:**
Download dari https://www.postgresql.org/download/windows/

#### 2. Buat Database dan User

```bash
# Login ke PostgreSQL
psql -U postgres

# Atau dengan user default
psql postgres
```

Kemudian di psql prompt:
```sql
-- Buat database
CREATE DATABASE dms_db;

-- Buat user
CREATE USER dms_user WITH PASSWORD 'dms_password';

-- Berikan privileges
GRANT ALL PRIVILEGES ON DATABASE dms_db TO dms_user;

-- Exit
\q
```

#### 3. Set Environment Variable

**Untuk Development Lokal (tanpa Docker):**
```bash
export DATABASE_URL="postgres://dms_user:dms_password@localhost:5432/dms_db?sslmode=disable"
```

**Untuk Docker:**
Tambahkan di `docker-compose.dev.yml`:
```yaml
environment:
  - DATABASE_URL=postgres://dms_user:dms_password@postgres:5432/dms_db?sslmode=disable
```

#### 4. Jalankan Backend

```bash
cd backend
go run main.go
```

Backend akan otomatis connect ke PostgreSQL dan migrate schema.

---

### Opsi 3: PostgreSQL Cloud/Managed (Production)

#### Contoh: AWS RDS, Google Cloud SQL, Azure Database

Set `DATABASE_URL` sesuai dengan connection string dari provider:

```bash
# Format umum
DATABASE_URL=postgres://username:password@host:port/database?sslmode=require

# Contoh AWS RDS
DATABASE_URL=postgres://admin:password@dms-db.xxxxx.us-east-1.rds.amazonaws.com:5432/dms_db?sslmode=require

# Contoh Google Cloud SQL
DATABASE_URL=postgres://user:pass@/dms_db?host=/cloudsql/project:region:instance

# Contoh Azure Database
DATABASE_URL=postgres://user@server:password@server.postgres.database.azure.com:5432/dms_db?sslmode=require
```

---

## 📋 Format DATABASE_URL

Format PostgreSQL connection string:
```
postgres://[user]:[password]@[host]:[port]/[database]?[parameters]
```

**Contoh:**
```
postgres://dms_user:dms_password@localhost:5432/dms_db?sslmode=disable
```

**Parameter SSL:**
- `sslmode=disable` - Untuk development lokal
- `sslmode=require` - Untuk production (wajib SSL)
- `sslmode=prefer` - SSL jika tersedia

---

## 🔄 Migrasi Data dari SQLite ke PostgreSQL

Jika Anda sudah punya data di SQLite dan ingin migrasi:

### Metode 1: Export/Import via SQL (Manual)

#### 1. Export Data dari SQLite

```bash
# Install sqlite3 jika belum ada
# macOS: brew install sqlite3
# Linux: sudo apt install sqlite3

# Export users table
sqlite3 backend/dms.db <<EOF
.mode csv
.headers on
.output users.csv
SELECT * FROM users;
.quit
EOF

# Export two_factor_auths (jika ada)
sqlite3 backend/dms.db <<EOF
.mode csv
.headers on
.output two_factor_auths.csv
SELECT * FROM two_factor_auths;
.quit
EOF

# Export audit_logs (jika ada)
sqlite3 backend/dms.db <<EOF
.mode csv
.headers on
.output audit_logs.csv
SELECT * FROM audit_logs;
.quit
EOF
```

#### 2. Import ke PostgreSQL

```bash
# Connect ke PostgreSQL
psql -U dms_user -d dms_db

# Import users (pastikan table sudah ada dari auto-migrate)
\copy users(id, username, email, role, password, created_at, updated_at) FROM 'users.csv' WITH CSV HEADER;

# Import two_factor_auths (jika ada)
\copy two_factor_auths(id, user_id, secret, enabled, backup_codes, created_at, updated_at) FROM 'two_factor_auths.csv' WITH CSV HEADER;

# Import audit_logs (jika ada)
\copy audit_logs(id, user_id, action, description, ip_address, user_agent, created_at) FROM 'audit_logs.csv' WITH CSV HEADER;
```

### Metode 2: Menggunakan GORM (Programmatic)

Buat script migrasi Go:

```go
// migrate.go
package main

import (
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Connect ke SQLite
    sqliteDB, err := gorm.Open(sqlite.Open("dms.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to SQLite:", err)
    }

    // Connect ke PostgreSQL
    postgresDB, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to PostgreSQL:", err)
    }

    // Migrate schema di PostgreSQL
    postgresDB.AutoMigrate(&UserModel{}, &TwoFactorAuth{}, &AuditLog{})

    // Copy data
    var users []UserModel
    sqliteDB.Find(&users)
    postgresDB.Create(&users)

    log.Println("Migration completed!")
}
```

---

## ✅ Verifikasi Migrasi

### 1. Cek Connection

```bash
# Cek logs backend
docker-compose -f docker-compose.dev.yml logs backend | grep -i "postgres\|database"

# Harus muncul:
# "Using PostgreSQL database"
# "Database connected and migrated successfully"
```

### 2. Test API

```bash
# Health check
curl http://localhost:8080/health

# Register user baru
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "test123456"
  }'
```

### 3. Cek Database

```bash
# Connect ke PostgreSQL
docker exec -it dms-postgres-dev psql -U dms_user -d dms_db

# Atau jika lokal
psql -U dms_user -d dms_db

# Cek tables
\dt

# Cek users
SELECT id, username, email, role, created_at FROM users;

# Exit
\q
```

---

## 🔧 Troubleshooting

### Error: "connection refused"

**Penyebab:** PostgreSQL belum running atau host/port salah.

**Solusi:**
```bash
# Cek PostgreSQL running
docker ps | grep postgres

# Atau untuk lokal
psql -U postgres -c "SELECT version();"
```

### Error: "password authentication failed"

**Penyebab:** Username/password salah di `DATABASE_URL`.

**Solusi:**
- Cek username dan password di `DATABASE_URL`
- Reset password PostgreSQL jika perlu:
```sql
ALTER USER dms_user WITH PASSWORD 'new_password';
```

### Error: "database does not exist"

**Penyebab:** Database belum dibuat.

**Solusi:**
```sql
CREATE DATABASE dms_db;
```

### Error: "relation does not exist"

**Penyebab:** Schema belum di-migrate.

**Solusi:**
- Restart backend (akan auto-migrate)
- Atau manual migrate:
```go
// Di backend, pastikan AutoMigrate dipanggil
DB.AutoMigrate(&UserModel{}, &TwoFactorAuth{}, &AuditLog{})
```

### Error: "CGO_ENABLED" saat build

**Penyebab:** CGO masih enabled padahal tidak perlu untuk PostgreSQL.

**Solusi:**
- Set `CGO_ENABLED=0` di environment variable
- Atau hapus dari docker-compose.yml

---

## 📝 Checklist Migrasi

- [ ] PostgreSQL installed/running
- [ ] Database dan user dibuat
- [ ] `DATABASE_URL` environment variable di-set
- [ ] Backend restart dan connect ke PostgreSQL
- [ ] Schema auto-migrate berhasil
- [ ] Test API endpoints
- [ ] Data migrated (jika ada data lama)
- [ ] Backup database (untuk production)

---

## 🎯 Rekomendasi

### Development
- ✅ Gunakan PostgreSQL di Docker (Opsi 1)
- ✅ Mudah setup dan cleanup
- ✅ Consistent dengan production

### Production
- ✅ Gunakan Managed PostgreSQL (AWS RDS, Google Cloud SQL, dll)
- ✅ Auto-backup dan monitoring
- ✅ High availability
- ✅ SSL/TLS enabled

---

## 📚 Referensi

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- GORM PostgreSQL Driver: https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL
- Connection String Format: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING

---

**Last Updated**: 2025-01-XX

