# Database Renamed: db_pedeve_dms

## ✅ Database Telah Di-rename

Database PostgreSQL telah di-rename dari `dms_db` menjadi `db_pedeve_dms`.

### Konfigurasi Baru

**Connection Details:**
```
Type:       PostgreSQL
Host:       localhost
Port:       5432
Database:   db_pedeve_dms  ⬅️ NAMA BARU
Username:   dms_user
Password:   dms_password
```

**Connection String:**
```
postgres://dms_user:dms_password@postgres:5432/db_pedeve_dms?sslmode=disable
```

**JDBC URL:**
```
jdbc:postgresql://localhost:5432/db_pedeve_dms
```

---

## 📋 File yang Telah Di-update

1. ✅ `docker-compose.dev.yml` - POSTGRES_DB dan DATABASE_URL
2. ✅ `docker-compose.postgres.yml` - POSTGRES_DB dan DATABASE_URL
3. ✅ `DBEAVER_POSTGRESQL.md` - Semua referensi database name
4. ✅ `migrate-to-postgres.sh` - Database name di script

---

## 🔍 Verifikasi

### 1. Cek Database Exists

```bash
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "\dt"
```

**Expected output:**
```
             List of relations
 Schema |      Name       | Type  |  Owner
--------+-----------------+-------+----------
 public | users           | table | dms_user
 public | two_factor_auths| table | dms_user
 public | audit_logs      | table | dms_user
```

### 2. Cek Backend Connection

```bash
docker-compose -f docker-compose.dev.yml logs backend | grep -i "postgres\|database"
```

**Expected output:**
```
Using PostgreSQL database
Database connected and migrated successfully
```

### 3. Test API

```bash
curl http://localhost:8080/health
```

**Expected:**
```json
{"status":"OK","service":"dms-backend"}
```

---

## 🔧 DBeaver Connection

Update connection di DBeaver dengan setting baru:

**Main Tab:**
- Host: `localhost`
- Port: `5432`
- Database: `db_pedeve_dms` ⬅️ UPDATE INI
- Username: `dms_user`
- Password: `dms_password`

Jika connection sudah ada di DBeaver:
1. Right-click connection → **Edit Connection**
2. Update **Database** field menjadi `db_pedeve_dms`
3. Klik **Test Connection**
4. Klik **Finish**

---

## 📝 Next Steps

1. ✅ Database sudah dibuat dengan nama baru
2. ✅ Backend sudah connect ke database baru
3. ✅ Schema sudah ter-migrate otomatis
4. ⏳ Update DBeaver connection (jika sudah ada)
5. ⏳ Test semua API endpoints
6. ⏳ Migrate data dari database lama (jika ada data penting di `dms_db`)

---

## 💡 Migrate Data dari Database Lama

Jika masih ada data di database lama (`dms_db`), Anda bisa migrate dengan:

```bash
# Export dari database lama
docker exec dms-postgres-dev pg_dump -U dms_user dms_db > old_database_backup.sql

# Import ke database baru
docker exec -i dms-postgres-dev psql -U dms_user -d db_pedeve_dms < old_database_backup.sql
```

Atau gunakan script rename yang akan menawarkan opsi rename database.

---

**Last Updated**: $(date)

