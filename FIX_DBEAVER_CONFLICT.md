# Fix DBeaver: Database Conflict dengan PostgreSQL Host

## 🔍 Root Cause

**Masalah**: Error `FATAL: database "db_dms_pedeve" does not exist` di DBeaver.

**Penyebab**: Ada **2 PostgreSQL instance** yang berbeda:

1. **PostgreSQL di Docker container** (PostgreSQL 16) - Port 5432
   - Database: `db_dms_pedeve` ✅ (ada)
   - Tables: `users`, `two_factor_auths`, `audit_logs` ✅

2. **PostgreSQL di Host** (PostgreSQL 14 Homebrew) - Port 5432
   - Database: `db_dms_pedeve` ❌ (tidak ada)
   - Running di localhost:5432

Saat DBeaver connect ke `127.0.0.1:5432`, dia connect ke **PostgreSQL di host**, bukan ke **Docker container**.

## ✅ Solution Options

### Option 1: Stop PostgreSQL di Host (Recommended)

Stop PostgreSQL Homebrew yang running di host:

```bash
# Stop PostgreSQL Homebrew
brew services stop postgresql@14
# atau
brew services stop postgresql

# Verify stopped
lsof -i :5432
```

Setelah stop, DBeaver akan connect ke Docker PostgreSQL.

### Option 2: Change Docker PostgreSQL Port

Ubah port Docker PostgreSQL ke port lain (misal 5433):

**Update `docker-compose.dev.yml`:**
```yaml
postgres:
  ports:
    - "5433:5432"  # ⬅️ GUNAKAN PORT 5433
```

**Update DBeaver:**
```
Host:        127.0.0.1
Port:        5433          ⬅️ GUNAKAN PORT 5433
Database:    db_dms_pedeve
Username:    postgres
Password:    dms_password
```

### Option 3: Create Database di PostgreSQL Host

Jika ingin tetap menggunakan PostgreSQL di host:

```bash
# Connect ke PostgreSQL host
psql -U postgres -d postgres

# Create database
CREATE DATABASE db_dms_pedeve;

# Create user (jika perlu)
CREATE USER postgres WITH PASSWORD 'dms_password' SUPERUSER;
```

Tapi ini tidak recommended karena akan ada 2 database terpisah.

## ✅ Recommended: Option 1 (Stop PostgreSQL Host)

### Step 1: Stop PostgreSQL Homebrew

```bash
# Stop service
brew services stop postgresql@14

# Atau jika tidak tahu versi
brew services stop postgresql

# Verify
lsof -i :5432
```

**Expected**: Hanya Docker container yang listen di port 5432.

### Step 2: Test Connection di DBeaver

**DBeaver Settings:**
```
Host:        127.0.0.1
Port:        5432
Database:    db_dms_pedeve
Username:    postgres
Password:    dms_password
```

**Test Connection** → Harus **Connected** ✅

### Step 3: Verify Tables

```sql
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;
```

**Expected**:
```
table_name
----------
audit_logs
two_factor_auths
users
```

## 🔧 Quick Fix Script

```bash
# Stop PostgreSQL Homebrew
brew services stop postgresql@14 2>/dev/null || brew services stop postgresql

# Verify only Docker PostgreSQL running
lsof -i :5432

# Test connection
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U postgres -d db_dms_pedeve -c "SELECT current_database();"
```

## ⚠️ Important Notes

1. **Jika perlu PostgreSQL di host**, gunakan **Option 2** (ubah port Docker)
2. **Jika tidak perlu PostgreSQL di host**, gunakan **Option 1** (stop service)
3. **Pastikan hanya 1 PostgreSQL** yang listen di port 5432 saat DBeaver connect

## ✅ Verification

Setelah fix, verify:

```bash
# Check port 5432
lsof -i :5432

# Should only show Docker container:
# com.docke ... TCP *:postgresql (LISTEN)

# Test connection
PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U postgres -d db_dms_pedeve -c "\dt"
```

**Expected**: Tables visible ✅

---

**Last Updated**: 2025-01-XX

