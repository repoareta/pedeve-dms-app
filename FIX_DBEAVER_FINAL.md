# Fix DBeaver Connection - Final Solution

## 🔍 Problem Identified

**Error**: `FATAL: role "dms_user" does not exist`

**Root Cause**:
- Connection dari dalam container (docker exec) ✅ OK
- Connection dari host (localhost/DBeaver) ❌ FAILED
- Masalah di **host authentication** atau **IPv6/IPv4**

## ✅ Solution

### Option 1: Use IP Address Instead of localhost (Recommended)

Di DBeaver, **gunakan IP address** `127.0.0.1` **bukan** `localhost`:

**Tab "Main":**
```
Host:        127.0.0.1    ⬅️ GANTI INI (bukan localhost)
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://127.0.0.1:5432/db_pedeve_dms
```

### Option 2: Force IPv4 in DBeaver

1. Di DBeaver connection settings, **Tab "Main"**
2. Ganti **Host** dari `localhost` ke `127.0.0.1`
3. Atau di **Connection URL**, gunakan `127.0.0.1` bukan `localhost`

### Option 3: Check PostgreSQL Container Port Binding

Pastikan port benar-benar exposed:

```bash
docker port dms-postgres-dev
```

Expected:
```
5432/tcp -> 0.0.0.0:5432
5432/tcp -> [::]:5432
```

## 🔧 Alternative: Create User Manually (Jika Masih Gagal)

Jika masih gagal, create user secara manual:

```bash
# Connect sebagai postgres user (default)
docker exec -it dms-postgres-dev psql -U postgres -d postgres

# Atau jika postgres user tidak ada, connect dengan dms_user
docker exec -it dms-postgres-dev psql -U dms_user -d postgres

# Create user baru (jika perlu)
CREATE USER dms_user WITH PASSWORD 'dms_password';
ALTER USER dms_user WITH SUPERUSER;
GRANT ALL PRIVILEGES ON DATABASE db_pedeve_dms TO dms_user;

# Verify
\du dms_user
```

## ✅ Verification Steps

### 1. Test dari Terminal (IPv4)

```bash
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U dms_user -d db_pedeve_dms -c "SELECT current_user;"
```

**Expected**: `dms_user`

### 2. Test dari DBeaver

1. **Host**: `127.0.0.1` (bukan localhost)
2. **Port**: `5432`
3. **Database**: `db_pedeve_dms`
4. **Username**: `dms_user`
5. **Password**: `dms_password`
6. Klik **Test Connection**

**Expected**: ✅ **Connected**

## 🎯 Quick Fix Script

Jalankan script ini untuk test:

```bash
./test-dbeaver-connection.sh
```

## ⚠️ Important Notes

1. **Gunakan `127.0.0.1` bukan `localhost`** di DBeaver
   - `localhost` bisa resolve ke IPv6 (::1) yang mungkin bermasalah
   - `127.0.0.1` adalah IPv4 yang lebih reliable

2. **Clear DBeaver Cache**:
   - Delete connection lama
   - Buat connection baru dengan IP `127.0.0.1`

3. **Jika masih gagal**:
   - Restart DBeaver
   - Atau restart PostgreSQL: `docker-compose -f docker-compose.dev.yml restart postgres`

## 📋 Final DBeaver Settings

**Connection Name**: `DMS App - PostgreSQL Dev`

**Tab "Main":**
```
Host:        127.0.0.1         ⬅️ PASTIKAN INI
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://127.0.0.1:5432/db_pedeve_dms
```

---

**Last Updated**: 2025-01-XX

