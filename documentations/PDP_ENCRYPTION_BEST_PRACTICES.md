# Best Practices Enkripsi Data Pribadi untuk Compliance UU PDP

Dokumen ini menjelaskan strategi dan best practices untuk mengamankan data pribadi sesuai dengan UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi.

## 1. Apakah Enkripsi di Database Sudah Mencukupi?

### Jawaban Singkat: **TIDAK, enkripsi di database saja TIDAK CUKUP**

### Penjelasan Lengkap:

Enkripsi di database adalah **salah satu lapisan keamanan**, tetapi tidak cukup sendiri. Kita perlu menggunakan pendekatan **Defense in Depth** (Pertahanan Berlapis):

#### Lapisan Keamanan yang Diperlukan:

1. **Enkripsi di Database** (Data at Rest)
   - Melindungi data ketika database disimpan
   - Tetapi tidak melindungi data ketika sedang diproses oleh aplikasi

2. **Enkripsi di Transport Layer** (Data in Transit)
   - HTTPS/TLS untuk komunikasi antara client dan server
   - Enkripsi koneksi antara aplikasi dan database

3. **Access Control** (Kontrol Akses)
   - Role-Based Access Control (RBAC)
   - Principle of Least Privilege
   - Authentication & Authorization

4. **Audit Logging**
   - Mencatat semua akses ke data pribadi
   - Mencatat semua perubahan data pribadi
   - Deteksi aktivitas mencurigakan

5. **Data Masking di UI**
   - Menyembunyikan sebagian data sensitif di tampilan
   - Contoh: "12******45" untuk KTP/NPWP

6. **Application-Level Security**
   - Input validation & sanitization
   - SQL injection prevention
   - XSS prevention

7. **Key Management**
   - Penyimpanan kunci enkripsi yang aman
   - Rotasi kunci secara berkala

---

## 2. Best Practice Enkripsi: Backend vs Database?

### Jawaban: **Kombinasi Keduanya dengan Tujuan yang Berbeda**

### Perbandingan:

#### A. Database-Level Encryption (TDE - Transparent Data Encryption)

**Kelebihan:**
- ✅ **Transparent** - Aplikasi tidak perlu tahu bahwa data di-enkripsi
- ✅ **Performance** - Biasanya lebih cepat karena dilakukan di level storage/disk
- ✅ **Backup Protection** - Backup database otomatis ter-enkripsi
- ✅ **Mudah diimplementasikan** - Hanya perlu konfigurasi database
- ✅ **Tidak perlu modifikasi code aplikasi**

**Kekurangan:**
- ❌ **Semua data di-enkripsi** - Tidak bisa selective encryption per field
- ❌ **Database admin masih bisa melihat data** - Jika punya akses ke database
- ❌ **Tidak fleksibel** - Sulit untuk enkripsi dengan algoritma yang berbeda per field

**Best Practice:**
- Gunakan untuk **enkripsi seluruh database** (disk-level)
- Cocok untuk melindungi dari **physical access** atau **backup theft**
- PostgreSQL: `pgcrypto` extension atau filesystem encryption
- SQLite: Encryption dengan SQLCipher

#### B. Application-Level Encryption (Backend Encryption)

**Kelebihan:**
- ✅ **Selective Encryption** - Hanya field tertentu yang di-enkripsi (KTP, NPWP, dll)
- ✅ **Granular Control** - Bisa menggunakan algoritma berbeda per field
- ✅ **Application Context** - Bisa enkripsi berdasarkan user role atau permission
- ✅ **Key Management** - Bisa menggunakan external key management (GCP Secret Manager, HashiCorp Vault)
- ✅ **Flexibility** - Mudah untuk rotasi kunci atau perubahan algoritma

**Kekurangan:**
- ❌ **Performance Overhead** - Setiap read/write perlu enkripsi/dekripsi
- ❌ **Complexity** - Perlu modifikasi code aplikasi
- ❌ **Search Limitation** - Tidak bisa langsung search data ter-enkripsi di database
- ❌ **Backup** - Backup sudah dalam bentuk ter-enkripsi (bukan plaintext)

**Best Practice:**
- Gunakan untuk **field-level encryption** untuk data sensitif (KTP, NPWP, Password)
- Cocok untuk data yang memerlukan **granular access control**
- Implement di backend dengan library seperti `golang.org/x/crypto` atau `crypto/aes`

#### C. Hybrid Approach (RECOMMENDED)

**Kombinasi Keduanya:**

```
┌─────────────────────────────────────────────┐
│  Client (HTTPS/TLS)                         │
│  ↓                                          │
│  Backend Application                        │
│  ├─ Application-Level Encryption            │
│  │  └─ Field-level: KTP, NPWP, Password    │
│  │  └─ Key Management: GCP Secret Manager  │
│  ↓                                          │
│  Database Connection (TLS)                  │
│  ↓                                          │
│  Database (PostgreSQL/SQLite)               │
│  ├─ TDE (Transparent Data Encryption)       │
│  │  └─ Disk-level encryption                │
│  └─ Backup Encryption                       │
└─────────────────────────────────────────────┘
```

**Rekomendasi untuk Aplikasi Ini:**

1. **Database-Level (TDE)**: 
   - Enkripsi seluruh database di level disk
   - Untuk melindungi dari physical access dan backup theft

2. **Application-Level (Field Encryption)**:
   - Enkripsi field sensitif: `KTP`, `NPWP`, `IdentityNumber`
   - Enkripsi kredensial: `Password` (hashing), `2FA Secret` (encryption)
   - Menggunakan AES-256-GCM atau ChaCha20-Poly1305

3. **Key Management**:
   - Simpan encryption keys di GCP Secret Manager atau HashiCorp Vault
   - Jangan pernah hardcode keys di source code
   - Implementasi key rotation

---

## 3. Enkripsi yang Tidak Mengganggu Performance: Di Sisi Mana?

### Jawaban: **Database-Level (TDE) untuk Performance, Application-Level untuk Granular Control**

### Analisis Performance:

#### Database-Level Encryption (TDE)

**Performance Impact:**
- ⚡ **Sangat Ringan** - Biasanya hanya 2-5% overhead
- ⚡ Dilakukan di level **storage/disk**, bukan di application layer
- ⚡ Database engine sudah di-optimize untuk ini
- ⚡ Transparent untuk aplikasi - tidak ada perubahan code

**Kapan Cocok:**
- Untuk **enkripsi seluruh database**
- Ketika performance adalah prioritas utama
- Untuk melindungi dari physical access

**Contoh Implementasi:**
- PostgreSQL: Filesystem encryption (LUKS, dm-crypt) atau `pgcrypto` untuk column-level
- SQLite: SQLCipher
- Cloud: GCP Cloud SQL dengan encryption at rest (automatic)

#### Application-Level Encryption

**Performance Impact:**
- ⚠️ **Sedang** - 10-20% overhead untuk field yang di-enkripsi
- ⚠️ Setiap read/write perlu enkripsi/dekripsi di application code
- ⚠️ Overhead lebih besar jika banyak field yang di-enkripsi

**Optimasi Performance:**
- ✅ Gunakan **AES-256-GCM** (lebih cepat dari CBC)
- ✅ Gunakan **hardware acceleration** jika tersedia (AES-NI)
- ✅ **Lazy encryption** - Hanya enkripsi saat write, dekripsi saat read
- ✅ **Caching** - Cache hasil dekripsi untuk session tertentu (dengan hati-hati)
- ✅ **Batch operations** - Kurangi jumlah enkripsi/dekripsi

**Kapan Cocok:**
- Untuk **field-level encryption** yang selective
- Ketika memerlukan granular access control
- Ketika ingin encrypt hanya data tertentu

### Rekomendasi Performance:

1. **Untuk Data Pribadi Spesifik (KTP, NPWP, IdentityNumber)**:
   - Gunakan **Application-Level Encryption** dengan AES-256-GCM
   - Field ini tidak sering di-query (bukan primary key atau index)
   - Overhead acceptable karena tidak high-frequency access

2. **Untuk Seluruh Database**:
   - Gunakan **Database-Level Encryption (TDE)**
   - Minimal performance impact
   - Otomatis melindungi semua data termasuk backup

3. **Untuk Password dan Kredensial**:
   - Password: **Hashing** (bcrypt) - sudah diimplementasi ✅
   - 2FA Secret: **Application-Level Encryption** (AES-256-GCM)

---

## 4. Hal-Hal Lain yang Perlu Dipertimbangkan

### A. Key Management

**Masalah yang Sering Terjadi:**
- ❌ Hardcode encryption keys di source code
- ❌ Menyimpan keys di environment variables yang tidak aman
- ❌ Tidak ada rotasi kunci

**Best Practice:**
- ✅ Gunakan **External Key Management Service**:
  - GCP Secret Manager
  - HashiCorp Vault
  - AWS KMS
- ✅ Implementasi **Key Rotation** (rotasi berkala)
- ✅ **Key Versioning** - Support multiple key versions selama rotasi
- ✅ **Key Access Logging** - Audit semua akses ke encryption keys

### B. Search dan Query Data Ter-Enkripsi

**Masalah:**
- Field yang ter-enkripsi tidak bisa langsung di-search di database
- Tidak bisa menggunakan index pada field ter-enkripsi

**Solusi:**
- ✅ **Encrypted Search**: Gunakan deterministic encryption untuk field yang perlu di-search (dengan trade-off keamanan)
- ✅ **Hash-based Search**: Buat hash dari field untuk exact match search
- ✅ **Application-Level Search**: Dekripsi di application layer kemudian filter (untuk dataset kecil)
- ✅ **Avoid Search**: Jika mungkin, hindari search pada field sensitif

### C. Data Masking di UI

**Pentning:**
- Data ter-enkripsi di database tetap perlu di-mask di UI
- Contoh: KTP "1234567890123456" → tampilkan "12******3456"
- Hanya user authorized yang bisa melihat data lengkap

### D. Audit Logging untuk Data Pribadi

**Wajib:**
- Log semua **read access** ke data pribadi spesifik (KTP, NPWP)
- Log semua **write access** (create, update, delete)
- Log **who, what, when, where** (user, action, timestamp, IP address)
- Store audit logs secara terpisah dan ter-enkripsi

### E. Data Retention dan Deletion

**Compliance:**
- Tentukan **retention policy** untuk data pribadi
- Implementasi **secure deletion** (bukan hanya soft delete)
- Data yang sudah tidak diperlukan harus dihapus atau di-anonymize
- Pastikan backup juga dihapus sesuai retention policy

### F. Data Breach Response

**Wajib:**
- Buat **Incident Response Plan**
- Implementasi **detection mechanism** (monitoring, alerting)
- Siapkan **notification procedure** jika terjadi data breach
- Sesuai UU PDP: Wajib memberitahu subjek data pribadi dan otoritas

### G. Transport Security (Data in Transit)

**Wajib:**
- ✅ **HTTPS/TLS** untuk semua komunikasi client-server
- ✅ **TLS untuk database connection** (jika database remote)
- ✅ **Certificate Management** - Gunakan valid SSL certificates
- ✅ **TLS 1.2 atau lebih baru** (jangan gunakan TLS 1.0/1.1)

### H. Backup Encryption

**Pentning:**
- Backup database juga harus ter-enkripsi
- Jika menggunakan TDE, backup otomatis ter-enkripsi
- Jika tidak, enkripsi backup secara manual sebelum disimpan

### I. Compliance Documentation

**Wajib:**
- Dokumentasi semua langkah perlindungan data
- Dokumentasi data yang diproses dan tujuannya
- Dokumentasi legal basis untuk pemrosesan
- Dokumentasi data retention policy
- Dokumentasi data breach response procedure

---

## Rekomendasi Implementasi untuk Aplikasi Ini

### Phase 1: Immediate (High Priority)

1. ✅ **Password Hashing** - Sudah diimplementasi dengan bcrypt
2. 🔄 **HTTPS/TLS** - Pastikan sudah digunakan (seharusnya sudah)
3. 🔄 **Audit Logging** - Sudah ada, perlu ditambahkan logging khusus untuk akses data pribadi spesifik
4. 🔄 **Data Masking di UI** - Implementasi masking untuk KTP, NPWP, IdentityNumber

### Phase 2: Short Term (1-2 bulan)

1. **Application-Level Encryption untuk Field Sensitif**:
   - Enkripsi `directors.KTP`, `directors.NPWP`
   - Enkripsi `shareholders.IdentityNumber` (untuk individu)
   - Enkripsi `two_factor_auths.Secret`, `two_factor_auths.BackupCodes`

2. **Key Management**:
   - Setup GCP Secret Manager atau HashiCorp Vault
   - Implementasi key storage dan retrieval
   - Implementasi key rotation mechanism

3. **Database-Level Encryption (TDE)**:
   - Enable encryption at rest untuk PostgreSQL (jika menggunakan Cloud SQL)
   - Atau setup filesystem encryption untuk SQLite/PostgreSQL

### Phase 3: Medium Term (3-6 bulan)

1. **Enhanced Audit Logging**:
   - Log khusus untuk akses data pribadi spesifik
   - Alert untuk aktivitas mencurigakan

2. **Data Retention Policy**:
   - Tentukan retention period untuk setiap jenis data
   - Implementasi automatic deletion/anonymization

3. **Compliance Documentation**:
   - Dokumentasi lengkap semua langkah perlindungan
   - Privacy policy
   - Data processing agreement

### Phase 4: Long Term (Ongoing)

1. **Regular Security Audits**
2. **Penetration Testing**
3. **Key Rotation Schedule**
4. **Training untuk Development Team**

---

## Contoh Implementasi Enkripsi di Go (Backend)

```go
// encryption.go
package infrastructure

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

type EncryptionService struct {
    key []byte // Key dari GCP Secret Manager atau Vault
}

func NewEncryptionService(key []byte) *EncryptionService {
    return &EncryptionService{key: key}
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, aesGCM.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := aesGCM.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

---

## Kesimpulan

1. **Enkripsi di database saja TIDAK CUKUP** - Perlu defense in depth
2. **Best practice: Kombinasi Database-Level + Application-Level** encryption
3. **Performance: Database-Level lebih ringan**, Application-Level untuk granular control
4. **Penting juga: Key Management, Audit Logging, Data Masking, Transport Security**

---

**Dokumen ini harus direview secara berkala untuk memastikan alignment dengan best practices terbaru dan regulasi yang berlaku.**

