# Troubleshooting DBeaver Connection ke PostgreSQL

## 🔍 Error: "FATAL: role 'dms_user' does not exist"

### Penyebab
User `dms_user` sebenarnya sudah ada di PostgreSQL container, tapi mungkin ada masalah dengan:
1. PostgreSQL container belum fully initialized
2. Authentication method tidak sesuai
3. Password tidak match
4. Host/port connection salah

### ✅ Verifikasi User Exists

Jalankan command ini untuk cek apakah user ada:

```bash
# Dari terminal
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "\du"
```

**Expected output:**
```
                             List of roles
 Role name |                         Attributes                         
-----------+------------------------------------------------------------
 dms_user  | Superuser, Create role, Create DB, Replication, Bypass RLS
```

### 🔧 Solusi

#### Option 1: Restart PostgreSQL Container

Jika user tidak terdeteksi oleh DBeaver:

```bash
# Restart PostgreSQL
docker-compose -f docker-compose.dev.yml restart postgres

# Wait sampai healthy
sleep 10

# Cek status
docker ps | grep postgres
```

#### Option 2: Recreate PostgreSQL (Fresh Start)

Jika masih bermasalah, recreate dengan fresh volume:

```bash
# Stop dan remove volume
docker-compose -f docker-compose.dev.yml down -v postgres

# Start ulang
docker-compose -f docker-compose.dev.yml up -d postgres

# Wait sampai healthy
sleep 15

# Restart backend
docker-compose -f docker-compose.dev.yml restart backend
```

#### Option 3: Manual Create User (Jika Perlu)

Jika user benar-benar tidak ada:

```bash
# Connect sebagai superuser (jika ada)
docker exec -it dms-postgres-dev psql -U postgres -d postgres

# Atau jika postgres user tidak ada, perlu recreate container
```

**JANGAN** lakukan manual create user karena PostgreSQL container sudah otomatis create user dari `POSTGRES_USER` environment variable.

### ✅ Settings DBeaver yang Benar

**Tab "Main":**
```
Host:        localhost
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**Connection URL:**
```
jdbc:postgresql://localhost:5432/db_pedeve_dms
```

### 🔍 Test Connection dari Terminal

Test connection dari terminal untuk memastikan credentials benar:

```bash
# Test dengan psql
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user, current_database();"

# Expected output:
#  current_user | current_database 
# --------------+------------------
#  dms_user     | db_pedeve_dms
```

### ⚠️ Common Issues

#### 1. Port 5432 Tidak Accessible

```bash
# Cek apakah port 5432 terbuka
lsof -i :5432

# Atau test dari host
psql -h localhost -p 5432 -U dms_user -d db_pedeve_dms
```

#### 2. PostgreSQL Container Tidak Running

```bash
# Cek status
docker ps | grep postgres

# Start jika tidak running
docker-compose -f docker-compose.dev.yml up -d postgres
```

#### 3. Password Salah

Pastikan password di DBeaver adalah `dms_password` (sesuai `POSTGRES_PASSWORD` di docker-compose).

#### 4. Database Tidak Ada

```bash
# Cek database exists
docker exec dms-postgres-dev psql -U dms_user -d postgres -c "\l" | grep db_pedeve_dms

# Jika tidak ada, restart backend untuk auto-create
docker-compose -f docker-compose.dev.yml restart backend
```

### 🔄 Quick Fix Script

Gunakan script fix otomatis:

```bash
./fix-postgres-user.sh
```

Script ini akan:
1. Cek connection
2. Jika ada masalah, tanya apakah ingin recreate PostgreSQL
3. Recreate jika diperlukan

### ✅ Verifikasi Setelah Fix

1. **Test dari Terminal:**
   ```bash
   docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT version();"
   ```

2. **Test dari DBeaver:**
   - Klik "Test Connection"
   - Harus muncul "Connected" dengan checklist hijau

3. **Cek Tables:**
   ```sql
   SELECT tablename FROM pg_tables WHERE schemaname = 'public';
   ```
   
   **Expected:**
   ```
   tablename
   -----------
   users
   two_factor_auths
   audit_logs
   ```

---

**Last Updated**: 2025-01-XX

