# Final Fix: DBeaver Connection Error

## 🔍 Root Cause

Error: `FATAL: role "dms_user" does not exist`

**Masalah**: PostgreSQL container menggunakan `POSTGRES_USER` yang membuat user, tapi user tersebut mungkin tidak fully configured untuk network connections dari host.

## ✅ Solution: Recreate PostgreSQL dengan Proper Configuration

### Step 1: Stop dan Remove PostgreSQL (⚠️ Hapus Data!)

```bash
docker-compose -f docker-compose.dev.yml down -v postgres
```

### Step 2: Update docker-compose.dev.yml (Sudah Diupdate)

File `docker-compose.dev.yml` sudah diupdate dengan:
```yaml
environment:
  - POSTGRES_USER=dms_user
  - POSTGRES_PASSWORD=dms_password
  - POSTGRES_DB=db_pedeve_dms
  - POSTGRES_INITDB_ARGS=--auth-host=scram-sha-256 --auth-local=trust
```

### Step 3: Start PostgreSQL Fresh

```bash
docker-compose -f docker-compose.dev.yml up -d postgres
```

### Step 4: Wait for PostgreSQL Ready

```bash
# Wait 15 seconds
sleep 15

# Check ready
docker exec dms-postgres-dev pg_isready -U dms_user
```

### Step 5: Restart Backend untuk Migrate Schema

```bash
docker-compose -f docker-compose.dev.yml restart backend
```

### Step 6: Verify User

```bash
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "\du dms_user"
```

**Expected**:
```
                             List of roles
 Role name |                         Attributes                         
-----------+------------------------------------------------------------
 dms_user  | Superuser, Create role, Create DB, Replication, Bypass RLS
```

### Step 7: Test Connection di DBeaver

**DBeaver Settings:**

Tab "Main":
```
Host:        127.0.0.1      ⬅️ GUNAKAN IP, BUKAN localhost
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://127.0.0.1:5432/db_pedeve_dms
```

Klik **Test Connection** → Harus **Connected** ✅

## 🔧 Alternative: Quick Fix Script

Jalankan script ini untuk fix otomatis:

```bash
# Stop dan remove PostgreSQL
docker-compose -f docker-compose.dev.yml down -v postgres

# Start fresh
docker-compose -f docker-compose.dev.yml up -d postgres

# Wait
sleep 15

# Verify
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user;"

# Restart backend
docker-compose -f docker-compose.dev.yml restart backend
```

## ⚠️ Important Notes

1. **Gunakan `127.0.0.1` bukan `localhost`** di DBeaver
   - Lebih reliable untuk network connections

2. **Clear DBeaver Cache**:
   - Delete connection lama
   - Buat connection baru

3. **Verify Password**:
   - Pastikan password: `dms_password` (tidak ada spasi)

4. **If Still Fails**:
   - Restart DBeaver application
   - Atau restart PostgreSQL: `docker-compose -f docker-compose.dev.yml restart postgres`

## ✅ Final Verification

Setelah fix, test dari terminal:

```bash
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U dms_user -d db_pedeve_dms -c "SELECT current_user;"
```

**Expected**: `dms_user`

Jika ini berhasil, DBeaver juga akan berhasil!

---

**Last Updated**: 2025-01-XX

