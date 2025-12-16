# Panduan Implementasi TDE (Transparent Data Encryption)

Dokumen ini menjelaskan cara mengimplementasikan TDE untuk aplikasi Pedeve DMS sesuai dengan compliance UU PDP.

## Overview

Aplikasi ini mendukung 2 database:
1. **PostgreSQL** - untuk production/development dengan docker
2. **SQLite** - untuk development lokal

Implementasi TDE berbeda untuk masing-masing database.

---

## 1. PostgreSQL TDE Implementation

### Option A: Cloud SQL (GCP) - **RECOMMENDED untuk Production**

Jika menggunakan GCP Cloud SQL, **encryption at rest sudah otomatis enabled** dan tidak perlu konfigurasi tambahan.

**Keuntungan:**
- ✅ Zero configuration
- ✅ Automatic key management
- ✅ Compliant dengan berbagai standar keamanan

### Option B: Self-hosted PostgreSQL dengan Filesystem Encryption

Untuk PostgreSQL yang self-hosted, gunakan **filesystem-level encryption** menggunakan LUKS (Linux Unified Key Setup).

#### Prerequisites:
- Linux host dengan LUKS support
- Root access ke server

#### Langkah Implementasi:

1. **Setup LUKS pada volume database:**

```bash
# 1. Stop PostgreSQL
sudo systemctl stop postgresql

# 2. Backup data existing
sudo cp -r /var/lib/postgresql/data /var/lib/postgresql/data.backup

# 3. Buat encrypted volume (contoh: 20GB)
sudo cryptsetup luksFormat /dev/sdX  # Ganti sdX dengan device yang sesuai

# 4. Buka encrypted volume
sudo cryptsetup luksOpen /dev/sdX postgres-encrypted

# 5. Format filesystem
sudo mkfs.ext4 /dev/mapper/postgres-encrypted

# 6. Mount encrypted volume
sudo mkdir /mnt/postgres-encrypted
sudo mount /dev/mapper/postgres-encrypted /mnt/postgres-encrypted

# 7. Copy data ke encrypted volume
sudo cp -r /var/lib/postgresql/data/* /mnt/postgres-encrypted/

# 8. Update fstab untuk auto-mount
echo "/dev/mapper/postgres-encrypted /var/lib/postgresql/data ext4 defaults 0 2" | sudo tee -a /etc/fstab

# 9. Update /etc/crypttab untuk auto-open LUKS
echo "postgres-encrypted /dev/sdX none luks" | sudo tee -a /etc/crypttab

# 10. Start PostgreSQL
sudo systemctl start postgresql
```

#### Untuk Docker Compose:

Jika menggunakan Docker Compose, bisa setup encrypted volume di host:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    volumes:
      - /mnt/postgres-encrypted:/var/lib/postgresql/data  # Mount dari encrypted volume di host
    # ... rest of config
```

### Option C: PostgreSQL dengan pgcrypto Extension (Column-Level Encryption)

**Catatan:** Option ini lebih kompleks dan memerlukan perubahan kode aplikasi. Tidak direkomendasikan untuk quick implementation.

Jika tetap ingin menggunakan pgcrypto:

```sql
-- Enable extension
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enkripsi column (contoh)
UPDATE directors SET ktp = pgp_sym_encrypt(ktp, 'encryption-key') WHERE ktp IS NOT NULL;
```

**Tidak direkomendasikan** karena:
- Perlu perubahan kode aplikasi untuk enkripsi/dekripsi
- Lebih kompleks dari filesystem encryption
- Performance overhead lebih besar

---

## 2. SQLite TDE Implementation

Untuk SQLite, gunakan **SQLCipher** yang merupakan drop-in replacement untuk SQLite dengan encryption built-in.

### Langkah Implementasi:

1. **Install SQLCipher dependencies:**

```bash
# macOS
brew install sqlcipher

# Ubuntu/Debian
sudo apt-get install sqlcipher libsqlcipher-dev

# Atau compile from source
git clone https://github.com/sqlcipher/sqlcipher.git
cd sqlcipher
./configure --enable-tempstore=yes CFLAGS="-DSQLITE_HAS_CODEC" LDFLAGS="-lcrypto"
make && sudo make install
```

2. **Update Go dependencies:**

```bash
cd backend
go get github.com/mutecomm/go-sqlcipher/v4
go mod tidy
```

3. **Update database.go untuk menggunakan SQLCipher:**

**File:** `backend/internal/infrastructure/database/database.go`

Ganti driver SQLite dengan SQLCipher:

```go
import (
    // Ganti import ini:
    // "gorm.io/driver/sqlite"
    
    // Dengan:
    sqlcipher "github.com/mutecomm/go-sqlcipher/v4"
    "gorm.io/driver/sqlite"
)

// Di fungsi InitDB(), ganti:
if dbURL == "" {
    zapLog.Info("Using SQLite database (development)")
    
    // Ganti ini:
    // dialector = sqlite.Open("dms.db")
    
    // Dengan SQLCipher:
    encryptionKey := getEncryptionKey() // Dapatkan dari environment atau secret manager
    dbPath := "dms.db"
    
    // SQLCipher connection string dengan encryption
    dsn := dbPath + "?_pragma_key=" + encryptionKey + "&_pragma_cipher_page_size=4096"
    dialector = sqlite.Open(dsn)
}
```

4. **Setup encryption key:**

Tambahkan fungsi untuk mendapatkan encryption key:

```go
// getEncryptionKey mendapatkan encryption key untuk SQLCipher
func getEncryptionKey() string {
    // Priority 1: Dari secret manager
    key, err := secrets.GetSecretWithFallback("sqlcipher_key", "SQLCIPHER_KEY", "")
    if err == nil && key != "" {
        return key
    }
    
    // Priority 2: Dari environment variable
    key = os.Getenv("SQLCIPHER_KEY")
    if key != "" {
        return key
    }
    
    // Priority 3: Generate default key (HANYA untuk development, tidak aman untuk production!)
    zapLog.Warn("SQLCIPHER_KEY not set, using default key (NOT SECURE for production!)")
    return "default-encryption-key-change-in-production"
}
```

5. **Migrate existing database (jika sudah ada data):**

```bash
# Backup database existing
cp dms.db dms.db.backup

# Convert plain SQLite ke SQLCipher
sqlcipher dms.db.backup
> PRAGMA key = 'your-encryption-key';
> ATTACH DATABASE 'dms_encrypted.db' AS encrypted KEY 'your-encryption-key';
> SELECT sqlcipher_export('encrypted');
> DETACH DATABASE encrypted;
> .exit

# Replace database
mv dms_encrypted.db dms.db
```

---

## 3. Docker Compose Configuration

### Untuk PostgreSQL dengan Filesystem Encryption:

Update `docker-compose.dev.yml` atau `docker-compose.postgres.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    volumes:
      # Option 1: Mount dari encrypted volume di host
      - /mnt/postgres-encrypted:/var/lib/postgresql/data
      
      # Option 2: Gunakan named volume (tidak encrypted, perlu setup di host level)
      # - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=dms_password
      - POSTGRES_DB=db_dms_pedeve
    # ... rest of config
```

**Catatan:** Untuk encryption di Docker, setup harus dilakukan di **host level** (filesystem encryption), bukan di container level.

### Untuk SQLite dengan SQLCipher:

Tidak ada perubahan di docker-compose, hanya perlu update code seperti dijelaskan di bagian SQLite TDE Implementation.

---

## 4. Environment Variables

Tambahkan environment variables berikut:

### PostgreSQL (Filesystem Encryption):
- Tidak ada env var khusus (encryption di level filesystem)

### SQLite (SQLCipher):
```bash
# Development
export SQLCIPHER_KEY="your-encryption-key-min-32-chars"

# Production (gunakan secret manager)
# Store di GCP Secret Manager atau HashiCorp Vault dengan key: "sqlcipher_key"
```

### Untuk Secret Manager:

Jika menggunakan GCP Secret Manager atau Vault:

```bash
# GCP Secret Manager
gcloud secrets create sqlcipher_key --data-file=key.txt

# HashiCorp Vault
vault kv put secret/dms-app sqlcipher_key="your-encryption-key"
```

---

## 5. Verifikasi Enkripsi

### PostgreSQL:

```bash
# Check apakah volume ter-enkripsi
lsblk
# Look for /dev/mapper/postgres-encrypted

# Check LUKS status
sudo cryptsetup status postgres-encrypted
```

### SQLite (SQLCipher):

```bash
# Try buka database tanpa key (harus error)
sqlite3 dms.db
> .tables
# Should fail: "Error: file is encrypted or is not a database"

# Try buka dengan key (harus berhasil)
sqlcipher dms.db
> PRAGMA key = 'your-encryption-key';
> .tables
# Should show tables successfully
```

---

## 6. Backup dengan Enkripsi

### PostgreSQL:

Backup otomatis ter-enkripsi jika menggunakan filesystem encryption:

```bash
# Backup
pg_dump -U postgres db_dms_pedeve > backup.sql

# Backup file akan tersimpan di encrypted filesystem
# Pastikan backup location juga encrypted
```

### SQLite (SQLCipher):

```bash
# Backup dengan encryption
sqlcipher dms.db
> PRAGMA key = 'your-encryption-key';
> .backup backup_encrypted.db
> PRAGMA key = 'your-encryption-key';
> .exit
```

---

## 7. Rekomendasi Implementasi

### Untuk Development (Lokal):

**SQLite dengan SQLCipher:**
- ✅ Mudah setup
- ✅ Cocok untuk development
- ⚠️ Pastikan encryption key tidak di-commit ke git

### Untuk Production:

**Option 1: GCP Cloud SQL (RECOMMENDED)**
- ✅ Encryption at rest otomatis enabled
- ✅ Zero configuration
- ✅ Managed service

**Option 2: Self-hosted PostgreSQL dengan Filesystem Encryption**
- ✅ Full control
- ✅ Encryption di level filesystem (transparent)
- ⚠️ Perlu setup manual

---

## 8. Checklist Implementasi

### Phase 1: Setup Encryption

- [ ] Tentukan database yang digunakan (PostgreSQL atau SQLite)
- [ ] Untuk PostgreSQL: Setup filesystem encryption atau gunakan Cloud SQL
- [ ] Untuk SQLite: Install SQLCipher dan update code
- [ ] Setup encryption keys di secret manager (jangan hardcode)
- [ ] Test enkripsi/dekripsi database

### Phase 2: Migration Existing Data

- [ ] Backup database existing
- [ ] Migrate database ke encrypted format
- [ ] Verify data integrity setelah migration
- [ ] Test aplikasi dengan encrypted database

### Phase 3: Documentation & Monitoring

- [ ] Dokumentasi encryption keys location
- [ ] Setup backup dengan encryption
- [ ] Setup monitoring untuk encryption status
- [ ] Document recovery procedure

---

## 9. Troubleshooting

### PostgreSQL:

**Error: Permission denied pada encrypted volume**
```bash
# Fix permissions
sudo chown -R postgres:postgres /var/lib/postgresql/data
```

**Error: Cannot mount encrypted volume**
```bash
# Check LUKS status
sudo cryptsetup status postgres-encrypted

# Re-open encrypted volume
sudo cryptsetup luksOpen /dev/sdX postgres-encrypted
```

### SQLite (SQLCipher):

**Error: "file is encrypted or is not a database"**
- Pastikan encryption key benar
- Pastikan database sudah di-convert ke SQLCipher format

**Error: "sqlcipher: command not found"**
- Install SQLCipher: `brew install sqlcipher` (macOS) atau `apt-get install sqlcipher` (Ubuntu)

---

## 10. Catatan Penting

1. **Encryption Key Management:**
   - JANGAN hardcode encryption keys di source code
   - Gunakan secret manager (GCP Secret Manager / HashiCorp Vault)
   - Rotate keys secara berkala (best practice: setiap 90 hari)

2. **Backup:**
   - Pastikan backup juga ter-enkripsi
   - Store backup keys secara terpisah dari backup data

3. **Performance:**
   - Filesystem encryption (LUKS): ~2-5% overhead
   - SQLCipher: ~5-10% overhead
   - Overhead ini acceptable untuk compliance

4. **Compliance:**
   - TDE sudah memenuhi requirement dasar UU PDP untuk data at rest encryption
   - Kombinasikan dengan HTTPS/TLS untuk data in transit
   - Implementasi audit logging untuk complete compliance

---

**Dokumen ini harus direview dan diupdate secara berkala untuk memastikan alignment dengan best practices terbaru.**

