# Laporan Keuangan - Bulk Upload

Panduan untuk upload banyak laporan keuangan sekaligus via Excel.

## Akses Fitur

- Dari halaman daftar laporan, klik tombol **Bulk Upload**
- Atau akses langsung: `/reports` dan klik tab "Bulk Upload"

## Langkah-langkah

### 1. Download Template

1. Klik tombol **Download Template**
2. File Excel akan terdownload
3. Template sudah berisi header kolom yang diperlukan

### 2. Isi Template Excel

Template memiliki kolom berikut:

| Kolom | Deskripsi | Format | Wajib |
|-------|-----------|--------|-------|
| Period (YYYY-MM) | Periode laporan | 2025-01 | Ya |
| Company Code | Kode perusahaan | PDV-001 | Ya |
| Revenue | Total pendapatan | Angka | Ya |
| OPEX | Total biaya operasional | Angka | Ya |
| NPAT | Net Profit After Tax | Angka | Ya |
| Dividend | Dividen | Angka | Ya |
| Financial Ratio (%) | Rasio keuangan | Angka | Ya |
| Remark | Catatan | Teks | Tidak |

**Contoh Data:**
```
Period (YYYY-MM) | Company Code | Revenue | OPEX | NPAT | Dividend | Financial Ratio (%) | Remark
2025-01          | PDV-001      | 1000000 | 500000 | 300000 | 50000 | 2.0 | Laporan bulan Januari
2025-02          | PDV-001      | 1200000 | 600000 | 400000 | 60000 | 2.0 | Laporan bulan Februari
```

### 3. Validasi Template

Sebelum upload, sistem akan memvalidasi:

1. **Format File**
   - Harus file Excel (.xlsx atau .xls)
   - Harus sesuai dengan template

2. **Header Kolom**
   - Semua kolom wajib harus ada
   - Nama kolom harus sesuai template

3. **Data**
   - Period harus format YYYY-MM
   - Company Code harus valid (ada di sistem)
   - Semua angka harus valid
   - Tidak boleh ada duplicate period untuk perusahaan yang sama

### 4. Upload File

1. **Klik "Browse" atau Drag & Drop**
   - Pilih file Excel yang sudah diisi
   - Atau drag file ke area upload

2. **Validasi Otomatis**
   - Sistem akan memvalidasi file
   - Hasil validasi ditampilkan:
     - ✅ Data valid: Siap untuk upload
     - ❌ Ada error: Daftar error ditampilkan

3. **Review Hasil Validasi**
   - Lihat preview data yang akan diupload
   - Cek error jika ada
   - Perbaiki error di Excel jika diperlukan

4. **Upload**
   - Jika validasi berhasil, klik **Upload**
   - Sistem akan memproses upload
   - Progress ditampilkan

### 5. Hasil Upload

Setelah upload selesai, sistem menampilkan:

- **Success**: Jumlah laporan yang berhasil diupload
- **Failed**: Jumlah laporan yang gagal
- **Errors**: Daftar error untuk laporan yang gagal

## Mekanisme Upsert

Sistem menggunakan mekanisme **upsert**:
- **Update**: Jika laporan untuk periode dan perusahaan sudah ada, akan diupdate
- **Insert**: Jika laporan belum ada, akan dibuat baru

**Keuntungan:**
- Tidak perlu cek manual apakah laporan sudah ada
- Bisa update banyak laporan sekaligus
- Tidak akan ada duplicate

## Tips

### Persiapan Data

- Pastikan Company Code benar dan sesuai dengan sistem
- Pastikan format Period benar (YYYY-MM)
- Pastikan semua angka valid (tidak ada teks di kolom angka)
- Cek duplicate period untuk perusahaan yang sama

### Validasi Sebelum Upload

- Gunakan fitur validasi untuk cek error sebelum upload
- Perbaiki semua error sebelum upload
- Review preview data untuk memastikan benar

### Best Practices

- Upload dalam batch kecil (50-100 baris) untuk menghindari timeout
- Backup data penting sebelum bulk update
- Simpan file Excel sebagai backup

## Troubleshooting

### Error: "Company Code tidak ditemukan"

- Pastikan Company Code sesuai dengan kode di sistem
- Cek apakah perusahaan sudah ada di sistem
- Hubungi administrator jika perusahaan seharusnya ada

### Error: "Format Period tidak valid"

- Pastikan format YYYY-MM (contoh: 2025-01, bukan 2025/01 atau 01-2025)
- Pastikan tidak ada spasi di awal/akhir
- Cek format cell di Excel (harus Text atau Custom: YYYY-MM)

### Error: "Duplicate period untuk perusahaan"

- Tidak boleh ada duplicate period untuk perusahaan yang sama
- Hapus baris duplicate di Excel
- Atau gunakan fitur edit untuk update laporan yang sudah ada

### Error: "Data tidak valid"

- Cek apakah semua kolom wajib sudah diisi
- Pastikan angka valid (tidak ada teks di kolom angka)
- Pastikan Financial Ratio sesuai (Revenue/OPEX)

### Upload Gagal

- Cek koneksi internet
- Coba upload file yang lebih kecil
- Hubungi administrator jika masalah berlanjut

## Referensi

- [Daftar Laporan](./financial-reports) - Lihat laporan yang sudah diupload
- [Tambah Laporan](./add-report) - Tambah laporan manual jika diperlukan
