# Panduan Koneksi DBeaver ke PostgreSQL DMS App

## 📋 Konfigurasi Koneksi DBeaver

### Settings untuk Koneksi

Berdasarkan konfigurasi PostgreSQL di `docker-compose.dev.yml`, berikut pengisian di DBeaver:

#### Tab "Main"

**Server Section:**
- **Connect by:** Pilih **Host** (radio button)
- **Host:** `localhost`
- **Port:** `5432`
- **Database:** `db_pedeve_dms` ⚠️ **PENTING: gunakan db_pedeve_dms**
- ✅ **Show all databases:** Bisa dicentang (opsional)

**Authentication Section:**
- **Authentication:** `Database Native` (default)
- **Username:** `dms_user` ⚠️ **BUKAN** `postgres`
- **Password:** `dms_password`
- ✅ **Save password:** Centang jika ingin password tersimpan

**Connection URL (akan auto-generate):**
```
jdbc:postgresql://localhost:5432/db_pedeve_dms
```

#### Tab "Driver properties" (Opsional)

Tidak perlu diubah, gunakan default settings.

---

## 🔍 Langkah-langkah Detail

### 1. Buat Connection Baru

1. Buka DBeaver
2. Klik **"New Database Connection"** (icon konektor) atau **File → New → Database Connection**
3. Pilih **PostgreSQL**
4. Klik **Next**

### 2. Isi Main Settings

Di tab **Main**:

```
Host:        localhost
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**⚠️ PENTING:**
- Database harus `db_pedeve_dms`
- Username harus `dms_user` (bukan `postgres`)

### 3. Test Connection

1. Klik tombol **"Test Connection ..."** di bagian bawah
2. Jika driver belum ada, DBeaver akan download PostgreSQL driver
3. Klik **Download** jika diminta
4. Setelah driver terinstall, klik **Test Connection** lagi
5. Harus muncul: **"Connected"** dengan checklist hijau

### 4. Save Connection

1. Klik **Finish**
2. Connection akan muncul di **Database Navigator** panel kiri

---

## ✅ Verifikasi Koneksi

Setelah connect, cek apakah tables sudah ter-migrate:

### Cek Tables

```sql
-- Di SQL Editor DBeaver
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public';
```

**Expected output:**
```
tablename
-----------
users
two_factor_auths
audit_logs
```

### Cek Users

```sql
SELECT id, username, email, role, created_at 
FROM users;
```

Jika sudah ada superadmin, akan muncul:
```
id          | username    | email | role       | created_at
------------|-------------|-------|------------|------------
<uuid>      | superadmin  | ...   | superadmin | 2025-01-XX
```

---

## 🔧 Troubleshooting

### Error: "Connection refused"

**Penyebab:** PostgreSQL container tidak running.

**Solusi:**
```bash
# Cek status
docker ps | grep postgres

# Start jika tidak running
docker-compose -f docker-compose.dev.yml up -d postgres

# Tunggu sampai healthy
docker ps | grep postgres
```

### Error: "password authentication failed"

**Penyebab:** Username/password salah.

**Solusi:**
- Pastikan:
  - Username: `dms_user` (bukan `postgres`)
  - Password: `dms_password`
  - Database: `db_pedeve_dms`

### Error: "database does not exist"

**Penyebab:** Database belum dibuat atau backend belum migrate schema.

**Solusi:**
1. Pastikan backend sudah running dan connect ke PostgreSQL
2. Cek logs backend:
   ```bash
   docker-compose -f docker-compose.dev.yml logs backend | grep -i "postgres\|database"
   ```
3. Restart backend jika perlu:
   ```bash
   docker-compose -f docker-compose.dev.yml restart backend
   ```

### Error: "Driver not found"

**Solusi:**
1. Saat test connection, DBeaver akan otomatis download driver
2. Klik **Download** dan tunggu selesai
3. Atau manual install:
   - Edit connection → Driver properties
   - Download driver dari: https://jdbc.postgresql.org/download/

### Error: "Connection timeout"

**Penyebab:** Port 5432 tidak accessible atau firewall blocking.

**Solusi:**
```bash
# Test connection dari terminal
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT version();"

# Cek port
lsof -i :5432
```

---

## 📊 Quick Reference

### Connection Details

```
Type:       PostgreSQL
Host:       localhost
Port:       5432
Database:   db_pedeve_dms
Username:   dms_user
Password:   dms_password
```

### Connection String

```
jdbc:postgresql://localhost:5432/db_pedeve_dms
```

### JDBC URL (Full)

```
jdbc:postgresql://localhost:5432/db_pedeve_dms?user=dms_user&password=dms_password
```

---

## 🎯 Tips

1. **Save Password:** Centang "Save password" agar tidak perlu input setiap kali connect

2. **Connection Name:** Beri nama yang jelas, misal: "DMS App - PostgreSQL Dev"

3. **Color Coding:** Di DBeaver bisa set warna untuk connection (right-click → Edit → Appearance)

4. **SQL Editor:** Gunakan SQL Editor untuk query manual:
   - Right-click connection → SQL Editor → New SQL script

5. **View Data:** Double-click table untuk view data
   - Right-click table → View Data

6. **Export Data:** 
   - Right-click table → Export Data → Pilih format (CSV, SQL, dll)

---

## 📝 Sample Queries

### Cek All Tables

```sql
SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;
```

### Cek Table Structure

```sql
-- Cek columns di table users
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'users'
ORDER BY ordinal_position;
```

### Cek Indexes

```sql
SELECT 
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
```

### Cek Row Counts

```sql
SELECT 
    'users' as table_name,
    COUNT(*) as row_count
FROM users
UNION ALL
SELECT 
    'two_factor_auths',
    COUNT(*)
FROM two_factor_auths
UNION ALL
SELECT 
    'audit_logs',
    COUNT(*)
FROM audit_logs;
```

---

## 🔐 Security Note

⚠️ **Untuk Production:**
- Gunakan password yang lebih kuat
- Jangan save password di shared computer
- Gunakan SSL connection (`sslmode=require`)
- Limit access dengan firewall rules

---

**Last Updated**: 2025-01-XX

