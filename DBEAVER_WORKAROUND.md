# DBeaver Connection Workaround

## 🔍 Masalah

Error: `FATAL: role "dms_user" does not exist` saat connect dari DBeaver.

**Fakta**:
- User `dms_user` **ADA** di dalam container ✅
- Connection dari **dalam container** OK ✅  
- Connection dari **host (DBeaver)** FAILED ❌

## 💡 Workaround: Gunakan `postgres` User

Karena user `dms_user` sepertinya tidak fully accessible dari host connection, gunakan user **`postgres`** sebagai workaround.

### Option 1: Connect dengan postgres User (Quick Fix)

**DBeaver Settings:**

Tab "Main":
```
Host:        127.0.0.1
Port:        5432
Database:    db_pedeve_dms
Username:    postgres          ⬅️ GUNAKAN POSTGRES USER
Password:    dms_password      ⬅️ Coba password yang sama
```

**Note**: Jika `postgres` user tidak ada, lanjut ke Option 2.

### Option 2: Create postgres User (Jika Tidak Ada)

```bash
# Connect sebagai dms_user dari dalam container
docker exec -it dms-postgres-dev psql -U dms_user -d postgres

# Create postgres user
CREATE USER postgres WITH PASSWORD 'dms_password' SUPERUSER;

# Grant privileges
GRANT ALL PRIVILEGES ON DATABASE db_pedeve_dms TO postgres;

# Exit
\q
```

Lalu di DBeaver, gunakan user `postgres` dengan password `dms_password`.

### Option 3: Manual Create dms_user dari postgres User

Jika `postgres` user ada, connect sebagai `postgres` lalu create/verify `dms_user`:

```sql
-- Connect sebagai postgres
-- Di DBeaver atau terminal

-- Check user exists
SELECT rolname FROM pg_roles WHERE rolname = 'dms_user';

-- If not exists, create
CREATE USER dms_user WITH PASSWORD 'dms_password' SUPERUSER;
GRANT ALL PRIVILEGES ON DATABASE db_pedeve_dms TO dms_user;

-- Verify
\du dms_user
```

## ✅ Recommended: Use postgres User for DBeaver

Untuk DBeaver, gunakan **user `postgres`** karena:
1. User `postgres` adalah default superuser yang pasti accessible dari network
2. Tidak ada masalah authentication
3. Punya semua permissions yang dibutuhkan

**DBeaver Connection Settings dengan postgres User:**

```
Host:        127.0.0.1
Port:        5432
Database:    db_pedeve_dms
Username:    postgres
Password:    dms_password    ⬅️ Atau password yang sama dengan POSTGRES_PASSWORD
```

## 🔧 Alternative: Use Default postgres User in docker-compose

Ubah `docker-compose.dev.yml` untuk menggunakan user `postgres` default:

```yaml
postgres:
  image: postgres:16-alpine
  environment:
    - POSTGRES_USER=postgres        ⬅️ GUNAKAN POSTGRES
    - POSTGRES_PASSWORD=dms_password
    - POSTGRES_DB=db_pedeve_dms
```

Lalu:
1. Stop dan remove PostgreSQL
2. Update docker-compose
3. Start fresh
4. Backend akan auto-create `dms_user` jika perlu

## 📋 Quick Fix Steps

1. **Test dengan postgres user**:
   ```bash
   PGPASSWORD=dms_password psql -h 127.0.0.1 -p 5432 -U postgres -d db_pedeve_dms -c "SELECT current_user;"
   ```

2. **Jika berhasil**, gunakan `postgres` user di DBeaver

3. **Jika postgres user tidak ada**, create dari container:
   ```bash
   docker exec -it dms-postgres-dev psql -U dms_user -d postgres -c "CREATE USER postgres WITH PASSWORD 'dms_password' SUPERUSER;"
   ```

---

**Last Updated**: 2025-01-XX

