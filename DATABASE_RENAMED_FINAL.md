# Database Renamed: db_dms_pedeve

## ✅ Database Telah Di-rename

Database PostgreSQL telah di-rename dari `db_pedeve_dms` menjadi **`db_dms_pedeve`**.

## ✅ Changes Completed

### 1. Updated Configuration Files

- ✅ `docker-compose.dev.yml` - POSTGRES_DB dan DATABASE_URL
- ✅ `docker-compose.postgres.yml` - POSTGRES_DB dan DATABASE_URL

### 2. PostgreSQL Recreated

- ✅ PostgreSQL container di-recreate dengan database baru
- ✅ Database `db_dms_pedeve` sudah dibuat
- ✅ Backend akan auto-migrate schema saat connect

## ✅ DBeaver Connection Settings (Updated)

**Connection Name**: `DMS App - PostgreSQL Dev`

**Tab "Main":**
```
Host:        127.0.0.1
Port:        5432
Database:    db_dms_pedeve      ⬅️ NAMA BARU
Username:    postgres
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://127.0.0.1:5432/db_dms_pedeve
```

## ✅ Verification

### 1. Check Database Exists

```bash
docker exec dms-postgres-dev psql -U postgres -d postgres -c "\l" | grep db_dms_pedeve
```

**Expected**: `db_dms_pedeve` ✅

### 2. Check Tables (After Backend Migrate)

```bash
docker exec dms-postgres-dev psql -U postgres -d db_dms_pedeve -c "\dt"
```

**Expected** (setelah backend migrate):
```
             List of relations
 Schema |      Name       | Type  |  Owner  
--------+-----------------+-------+----------
 public | users           | table | postgres
 public | two_factor_auths| table | postgres
 public | audit_logs      | table | postgres
```

### 3. Test Connection di DBeaver

1. **Edit connection** di DBeaver
2. **Update Database** field menjadi `db_dms_pedeve`
3. **Test Connection** → Harus **Connected** ✅

## 📋 Summary

- ✅ Database name: **`db_dms_pedeve`** (baru)
- ✅ User: `postgres`
- ✅ Password: `dms_password`
- ✅ Port: `5432`
- ✅ Connection URL: `jdbc:postgresql://127.0.0.1:5432/db_dms_pedeve`

## 🔄 Update DBeaver Connection

Jika connection sudah ada di DBeaver:

1. **Right-click connection** → **Edit Connection**
2. **Tab "Main"**:
   - Database: `db_dms_pedeve` ⬅️ UPDATE INI
3. **Test Connection**
4. **Finish**

---

**Last Updated**: 2025-01-XX

