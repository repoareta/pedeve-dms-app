# Fix DBeaver: Database "db_dms_pedeve" does not exist

## 🔍 Problem

Error: `FATAL: database "db_dms_pedeve" does not exist` saat test connection di DBeaver.

**Fakta**:
- Database `db_dms_pedeve` **ADA** di dalam container ✅
- Connection dari **dalam container** OK ✅
- Connection dari **host (DBeaver)** FAILED ❌

## ✅ Solution: Recreate PostgreSQL

Database ada di container tapi tidak accessible dari host. Solusinya adalah recreate PostgreSQL dengan fresh initialization.

### Step 1: Stop dan Remove PostgreSQL

```bash
docker-compose -f docker-compose.dev.yml down -v postgres
```

### Step 2: Start PostgreSQL Fresh

```bash
docker-compose -f docker-compose.dev.yml up -d postgres
```

### Step 3: Wait for Initialization

```bash
# Wait 25 seconds
sleep 25

# Verify ready
docker exec dms-postgres-dev pg_isready -U postgres
```

### Step 4: Verify Database Created

```bash
docker exec dms-postgres-dev psql -U postgres -d postgres -c "\l" | grep db_dms_pedeve
```

**Expected**: `db_dms_pedeve` ✅

### Step 5: Test Connection from Host

```bash
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U postgres -d db_dms_pedeve -c "SELECT current_database();"
```

**Expected**: `db_dms_pedeve` ✅

### Step 6: Restart Backend untuk Migrate Schema

```bash
docker-compose -f docker-compose.dev.yml restart backend
```

### Step 7: Verify Tables

```bash
docker exec dms-postgres-dev psql -U postgres -d db_dms_pedeve -c "\dt"
```

**Expected**:
```
             List of relations
 Schema |      Name       | Type  |  Owner  
--------+-----------------+-------+----------
 public | users           | table | postgres
 public | two_factor_auths| table | postgres
 public | audit_logs      | table | postgres
```

## ✅ DBeaver Connection Settings

**Tab "Main":**
```
Host:        127.0.0.1
Port:        5432
Database:    db_dms_pedeve
Username:    postgres
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://127.0.0.1:5432/db_dms_pedeve
```

## 🔧 Quick Fix Script

Jalankan script ini untuk fix otomatis:

```bash
# Stop dan remove PostgreSQL
docker-compose -f docker-compose.dev.yml down -v postgres

# Start fresh
docker-compose -f docker-compose.dev.yml up -d postgres

# Wait
sleep 25

# Verify
docker exec dms-postgres-dev psql -U postgres -d postgres -c "\l" | grep db_dms_pedeve

# Test from host
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U postgres -d db_dms_pedeve -c "SELECT current_database();"

# Restart backend
docker-compose -f docker-compose.dev.yml restart backend
```

## ⚠️ Important Notes

1. **Gunakan `127.0.0.1` bukan `localhost`** di DBeaver
2. **Wait cukup lama** (25 detik) setelah PostgreSQL start untuk initialization
3. **Verify database exists** sebelum test connection di DBeaver
4. **Clear DBeaver cache** (delete connection lama, buat baru)

## ✅ Verification Checklist

- [ ] PostgreSQL container running dan healthy
- [ ] Database `db_dms_pedeve` exists (dari docker exec)
- [ ] Database visible dari host connection (psql -h 127.0.0.1)
- [ ] Backend connected dan migrated schema
- [ ] Tables exist: users, two_factor_auths, audit_logs
- [ ] DBeaver test connection: Connected ✅

---

**Last Updated**: 2025-01-XX

