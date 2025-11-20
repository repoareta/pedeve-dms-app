# Status Migrasi ke PostgreSQL

## ✅ Migrasi Telah Dimulai

Date: $(date)

### Status Services

Jalankan perintah berikut untuk cek status:

```bash
docker ps --filter "name=dms"
```

### Verifikasi PostgreSQL Connection

#### 1. Cek Backend Logs

```bash
docker-compose -f docker-compose.dev.yml logs backend | grep -i "postgres\|database\|migrate"
```

Harus muncul:
- ✅ "Using PostgreSQL database"
- ✅ "Database connected and migrated successfully"

#### 2. Cek Database Tables

```bash
docker exec dms-postgres-dev psql -U dms_user -d dms_db -c "\dt"
```

Harus muncul tables:
- ✅ `users`
- ✅ `two_factor_auths`
- ✅ `audit_logs`

#### 3. Test API

```bash
# Health check
curl http://localhost:8080/health

# Expected: {"status": "OK", "service": "dms-backend"}
```

#### 4. Cek Superadmin User

```bash
docker exec dms-postgres-dev psql -U dms_user -d dms_db -c "SELECT username, email, role FROM users;"
```

### Migrasi Data dari SQLite (Jika Ada Data Lama)

Jika Anda punya data di SQLite (`backend/dms.db`) yang perlu dimigrate, ikuti panduan di `POSTGRESQL_MIGRATION.md`.

#### Quick Export/Import

```bash
# Export dari SQLite
sqlite3 backend/dms.db ".dump users" > users_export.sql

# Import ke PostgreSQL
docker exec -i dms-postgres-dev psql -U dms_user -d dms_db < users_export.sql
```

### Troubleshooting

#### Backend tidak connect ke PostgreSQL

1. Cek PostgreSQL status:
   ```bash
   docker exec dms-postgres-dev pg_isready -U dms_user
   ```

2. Cek DATABASE_URL di docker-compose:
   ```bash
   docker exec dms-backend-dev env | grep DATABASE_URL
   ```

3. Restart backend:
   ```bash
   docker-compose -f docker-compose.dev.yml restart backend
   ```

#### Schema tidak ter-migrate

1. Backend akan auto-migrate saat startup
2. Jika gagal, cek logs:
   ```bash
   docker-compose -f docker-compose.dev.yml logs backend
   ```

3. Manual migrate (jika perlu):
   - Restart backend container

### Info Koneksi PostgreSQL

```
Host: localhost
Port: 5432
Database: dms_db
User: dms_user
Password: dms_password
```

### Connection String

```
postgres://dms_user:dms_password@localhost:5432/dms_db?sslmode=disable
```

### Next Steps

1. ✅ PostgreSQL service running
2. ✅ Backend connected to PostgreSQL
3. ✅ Schema migrated
4. ⏳ Migrate existing data (jika ada)
5. ⏳ Test all API endpoints
6. ⏳ Update production config (jika perlu)

---

**Last Updated**: $(date)

