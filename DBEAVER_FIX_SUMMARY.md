# Fix DBeaver Connection - Summary

## ✅ Status: Semua Konfigurasi Benar

Dari diagnostic, semua sudah benar:
- ✅ User `dms_user` exists
- ✅ Database `db_pedeve_dms` exists  
- ✅ Port 5432 exposed
- ✅ User punya superuser permissions
- ✅ Connection dari dalam container OK

## 🔧 Solusi untuk Error "role 'dms_user' does not exist"

### Step 1: Restart PostgreSQL

```bash
docker-compose -f docker-compose.dev.yml restart postgres
```

Tunggu 10 detik, lalu coba connect lagi di DBeaver.

### Step 2: Jika Masih Gagal - Recreate PostgreSQL

Jika restart tidak membantu, recreate PostgreSQL dengan fresh volume:

```bash
# Stop dan remove volume (⚠️ HAPUS DATA!)
docker-compose -f docker-compose.dev.yml down -v postgres

# Start ulang
docker-compose -f docker-compose.dev.yml up -d postgres

# Wait sampai healthy
sleep 15

# Restart backend untuk migrate schema
docker-compose -f docker-compose.dev.yml restart backend
```

### Step 3: Verifikasi DBeaver Settings

Pastikan settings di DBeaver **PERSIS** seperti ini:

**Tab "Main":**
```
Host:        localhost
Port:        5432
Database:    db_pedeve_dms
Username:    dms_user
Password:    dms_password
```

**⚠️ PENTING:**
- Jangan ada spasi di username/password
- Pastikan password benar: `dms_password` (bukan `dms_password ` dengan spasi)
- Database name: `db_pedeve_dms` (bukan `dms_db`)

### Step 4: Clear DBeaver Cache (Jika Perlu)

Jika masih error, clear DBeaver connection cache:

1. Di DBeaver, **Delete** connection yang lama
2. Buat **New Connection** dengan settings di atas
3. Test Connection

### Step 5: Test dari Terminal (Verifikasi)

Test connection dari terminal untuk memastikan PostgreSQL accessible:

```bash
# Test dengan docker exec (pasti berhasil)
docker exec dms-postgres-dev psql -U dms_user -d db_pedeve_dms -c "SELECT current_user;"

# Expected: dms_user
```

## 🔍 Troubleshooting Lanjutan

### Issue: Port 5432 Already in Use

Jika port 5432 sudah digunakan:

```bash
# Cek port
lsof -i :5432

# Atau ubah port di docker-compose.dev.yml
ports:
  - "5433:5432"  # Gunakan port 5433 di host
```

Lalu di DBeaver, gunakan port **5433**.

### Issue: Connection Timeout

Jika connection timeout:

1. Cek PostgreSQL container running:
   ```bash
   docker ps | grep postgres
   ```

2. Cek health status:
   ```bash
   docker ps --filter "name=postgres" --format "{{.Status}}"
   ```
   
   Harus muncul: `Up X minutes (healthy)`

3. Restart jika tidak healthy:
   ```bash
   docker-compose -f docker-compose.dev.yml restart postgres
   ```

### Issue: Wrong Password

Jika password salah:

1. Cek password di docker-compose:
   ```bash
   docker exec dms-postgres-dev env | grep POSTGRES_PASSWORD
   ```

2. Pastikan di DBeaver menggunakan password yang sama: `dms_password`

## ✅ Quick Fix Script

Gunakan script otomatis:

```bash
./fix-dbeaver-connection.sh
```

Script akan:
- ✅ Check user exists
- ✅ Check database exists
- ✅ Test connection
- ✅ Show DBeaver settings

## 📋 Final Checklist

Sebelum connect di DBeaver, pastikan:

- [ ] PostgreSQL container running dan healthy
- [ ] Port 5432 accessible
- [ ] User `dms_user` exists (verified dengan script)
- [ ] Database `db_pedeve_dms` exists
- [ ] DBeaver settings benar (host, port, database, username, password)
- [ ] Tidak ada typo di username/password
- [ ] DBeaver connection baru (bukan cache lama)

## 🎯 Expected Result

Setelah fix, DBeaver connection harus:
- ✅ Test Connection: **Connected** (hijau)
- ✅ Bisa lihat database `db_pedeve_dms`
- ✅ Bisa lihat tables: `users`, `two_factor_auths`, `audit_logs`
- ✅ Bisa query data

---

**Last Updated**: 2025-01-XX

