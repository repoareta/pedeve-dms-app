# SQLite to PostgreSQL Migration Tool

Script untuk migrate data dari SQLite database ke PostgreSQL.

## Cara Menggunakan

### 1. Pastikan PostgreSQL sudah running

```bash
# Start PostgreSQL dengan Docker
docker-compose -f ../../docker-compose.dev.yml up -d postgres

# Atau jika sudah running, skip langkah ini
```

### 2. Set Environment Variables (Optional)

```bash
# Path ke SQLite database (default: dms.db)
export SQLITE_PATH=../../dms.db

# PostgreSQL connection string (default: postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable)
export DATABASE_URL=postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable
```

### 3. Jalankan Migration

```bash
# Dari root backend directory
cd backend
go run ./cmd/migrate-sqlite-to-postgres

# Atau dengan environment variables
SQLITE_PATH=./dms.db DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable" go run ./cmd/migrate-sqlite-to-postgres
```

### 4. Verifikasi Data

Setelah migration selesai, verifikasi data di PostgreSQL:

```bash
# Connect ke PostgreSQL
docker exec -it dms-postgres-dev psql -U postgres -d db_dms_pedeve

# Cek jumlah data
SELECT 'users' as table_name, COUNT(*) FROM users
UNION ALL
SELECT 'roles', COUNT(*) FROM roles
UNION ALL
SELECT 'permissions', COUNT(*) FROM permissions
UNION ALL
SELECT 'role_permissions', COUNT(*) FROM role_permissions
UNION ALL
SELECT 'companies', COUNT(*) FROM companies
UNION ALL
SELECT 'two_factor_auths', COUNT(*) FROM two_factor_auths
UNION ALL
SELECT 'audit_logs', COUNT(*) FROM audit_logs;
```

## Urutan Migration

Script akan migrate data dalam urutan berikut (menghormati foreign keys):

1. **roles** - Tidak ada dependency
2. **permissions** - Tidak ada dependency
3. **role_permissions** - Butuh roles dan permissions
4. **companies** - Tidak ada dependency
5. **users** - Butuh roles dan companies
6. **two_factor_auths** - Butuh users
7. **audit_logs** - Butuh users (tapi user_id bisa NULL)

## Catatan

- Script menggunakan `ON CONFLICT DO NOTHING` untuk menghindari duplicate entries
- Jika ada data yang sudah ada, akan di-skip (tidak akan error)
- Pastikan PostgreSQL sudah di-migrate schema-nya terlebih dahulu (akan otomatis saat backend start)

