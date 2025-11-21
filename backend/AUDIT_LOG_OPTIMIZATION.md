# Optimasi Audit Log Storage

## Analisis Potensi Storage

### Ukuran Per Record
- **User Action**: ~200-300 bytes (login, logout, create, update, delete)
- **Technical Error**: 1-5 KB (dengan stack trace lengkap)
- **Rata-rata**: ~500 bytes per record

### Estimasi Pertumbuhan

#### Skenario Kecil (100 user aktif, 50 aksi/hari)
- 5,000 records/hari × 500 bytes = **2.5 MB/hari**
- **~75 MB/bulan**
- **~900 MB/tahun**

#### Skenario Sedang (500 user aktif, 75 aksi/hari)
- 37,500 records/hari × 500 bytes = **18.75 MB/hari**
- **~562 MB/bulan**
- **~6.7 GB/tahun**

#### Skenario Besar (1,000 user aktif, 100 aksi/hari)
- 100,000 records/hari × 500 bytes = **50 MB/hari**
- **~1.5 GB/bulan**
- **~18 GB/tahun**

## Solusi yang Sudah Diimplementasikan

### 1. Retention Policy (Automatic Cleanup)
- **User Actions**: Disimpan selama 90 hari (3 bulan) - default
- **Technical Errors**: Disimpan selama 30 hari (1 bulan) - default
- Cleanup otomatis berjalan setiap 24 jam
- Dapat dikonfigurasi via environment variables:
  - `AUDIT_LOG_USER_ACTION_RETENTION_DAYS` (default: 90)
  - `AUDIT_LOG_TECHNICAL_ERROR_RETENTION_DAYS` (default: 30)

### 2. Indexing untuk Performa
- Index pada `created_at` untuk query berdasarkan waktu
- Index pada `log_type` untuk filter cepat
- Index pada `user_id`, `action`, `resource`, `status` untuk filtering

## Rekomendasi Tambahan (Opsional)

### 1. Archiving ke Cold Storage
Untuk data yang perlu disimpan lebih lama (compliance):
- Pindahkan log > 90 hari ke S3/Cloud Storage
- Kompres dengan gzip sebelum archive
- Hapus dari database setelah archive

### 2. Kompresi Stack Trace
Untuk technical errors:
- Kompres stack trace sebelum disimpan
- Gunakan gzip compression untuk field `details`
- Dapat mengurangi ukuran hingga 70-80%

### 3. Sampling untuk High-Volume Actions
Untuk aksi yang sangat sering (misalnya view document):
- Sample hanya 10% dari aksi view
- Atau log hanya aksi penting (create, update, delete)

### 4. Partitioning (PostgreSQL)
Jika menggunakan PostgreSQL:
- Partition tabel berdasarkan `created_at` (bulanan)
- Memudahkan cleanup dan meningkatkan performa query

### 5. Monitoring Storage Usage
- Gunakan `GetAuditLogStats()` untuk monitoring
- Set alert jika storage melebihi threshold
- Review retention policy secara berkala

## Konfigurasi Environment Variables

```bash
# Retention period (dalam hari)
AUDIT_LOG_USER_ACTION_RETENTION_DAYS=90        # Default: 90 hari
AUDIT_LOG_TECHNICAL_ERROR_RETENTION_DAYS=30    # Default: 30 hari
```

## Monitoring

Endpoint untuk melihat statistik audit log (dapat ditambahkan):
- Total records
- Count by log type
- Estimated size
- Oldest/newest record dates

## Best Practices

1. **Review retention policy** setiap 6 bulan berdasarkan kebutuhan compliance
2. **Monitor storage growth** secara berkala
3. **Archive data penting** sebelum dihapus (jika diperlukan untuk compliance)
4. **Adjust retention** berdasarkan:
   - Kebutuhan compliance/regulasi
   - Volume traffic aplikasi
   - Budget storage

## Perhitungan Storage dengan Retention Policy

Dengan retention policy yang sudah diimplementasikan:

### Skenario Besar (1,000 user, 100 aksi/hari)
- **User Actions** (90 hari retention):
  - 100,000 records/hari × 90 hari = 9,000,000 records
  - 9,000,000 × 300 bytes = **~2.7 GB**

- **Technical Errors** (30 hari retention):
  - Asumsi 1% error rate = 1,000 errors/hari
  - 1,000 × 30 hari = 30,000 records
  - 30,000 × 2 KB = **~60 MB**

- **Total maksimal**: **~2.76 GB** (bukan 18 GB/tahun)

**Penghematan**: ~85% dibanding tanpa retention policy!

