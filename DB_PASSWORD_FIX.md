# Fix Database Password Authentication

## Status: ✅ FIXED
Password sudah di-reset menjadi: `PedeveDb#2025!`
- ✅ Updated di GCP Secret Manager
- ✅ Updated di GitHub Secrets
- ✅ Updated di Cloud SQL

## Masalah (Resolved)
Error: `password authentication failed for user "pedeve_user_db"`

Ini berarti password di GCP Secret Manager tidak sesuai dengan password di Cloud SQL.
**Sudah diperbaiki dengan reset password.**

## Solusi: Verifikasi dan Sync Password

### Opsi 1: Update Password di Cloud SQL (Recommended)

1. **SSH ke backend VM:**
   ```bash
   gcloud compute ssh backend-dev --zone=asia-southeast2-a --project=pedeve-pertamina-dms
   ```

2. **Test koneksi dengan password dari Secret Manager:**
   ```bash
   # Get password dari Secret Manager
   DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)
   
   # Test dengan psql (jika Cloud SQL Proxy running)
   export PGPASSWORD="${DB_PASSWORD}"
   psql -h 127.0.0.1 -p 5432 -U pedeve_user_db -d db_dev_pedeve -c "SELECT version();"
   ```

3. **Jika gagal, reset password di Cloud SQL:**
   ```bash
   # Connect sebagai postgres user (via Cloud SQL Proxy)
   psql -h 127.0.0.1 -p 5432 -U postgres -d postgres
   
   # Di dalam psql, reset password:
   ALTER USER pedeve_user_db WITH PASSWORD 'BUG84NbNnceIi+k)';
   \q
   ```

4. **Verifikasi password sudah benar:**
   ```bash
   export PGPASSWORD="BUG84NbNnceIi+k)"
   psql -h 127.0.0.1 -p 5432 -U pedeve_user_db -d db_dev_pedeve -c "SELECT version();"
   ```

### Opsi 2: Update Password di Secret Manager

Jika password di database berbeda, update Secret Manager:

```bash
# Update password di Secret Manager
echo -n "PASSWORD_YANG_BENAR_DI_DATABASE" | gcloud secrets versions add db_password \
  --data-file=- \
  --project=pedeve-pertamina-dms
```

### Opsi 3: Reset Password di Cloud SQL via Console

1. **Buka Cloud SQL Console:**
   - Go to: https://console.cloud.google.com/sql/instances
   - Pilih instance: `postgres-dev`

2. **Edit User:**
   - Klik "Users" tab
   - Klik user `pedeve_user_db`
   - Klik "Edit"
   - Set password baru: `BUG84NbNnceIi+k)`
   - Save

3. **Verifikasi:**
   - Pastikan password di Secret Manager sama dengan yang baru di-set

## Verifikasi Setelah Fix

1. **Test dari VM:**
   ```bash
   # Copy test script ke VM
   gcloud compute scp scripts/test-db-connection.sh backend-dev:~/
   
   # SSH dan run
   gcloud compute ssh backend-dev --zone=asia-southeast2-a
   chmod +x ~/test-db-connection.sh
   ~/test-db-connection.sh pedeve-pertamina-dms
   ```

2. **Check container logs:**
   ```bash
   sudo docker logs dms-backend-prod
   ```

3. **Test health endpoint:**
   ```bash
   curl http://127.0.0.1:8080/health
   ```

## Best Practice

1. **Password harus sama di:**
   - Cloud SQL user password
   - GCP Secret Manager `db_password`

2. **Untuk karakter khusus (+, ), dll):**
   - Password di database: simpan as-is
   - Password di Secret Manager: simpan as-is
   - URL encoding dilakukan otomatis di deployment script

3. **Verifikasi sebelum deploy:**
   - Test koneksi dari VM dengan password dari Secret Manager
   - Pastikan bisa connect sebelum deploy container

