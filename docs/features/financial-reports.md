# Laporan Keuangan - Daftar Laporan

Halaman ini menampilkan status laporan keuangan semua perusahaan dalam format matrix per bulan.

## Akses Halaman

- Klik menu **Laporan** di navigasi atas
- Atau akses langsung: `/reports`

## Tampilan

Halaman menampilkan tabel dengan struktur matrix:

- **Company Name**: Nama perusahaan dengan logo/icon
- **12 Kolom Bulan**: Jan, Feb, Mar, Apr, May, Jun, Jul, Aug, Sep, Oct, Nov, Dec
  - Setiap kolom menunjukkan status laporan untuk bulan tersebut
  - **"Ada"** (hijau): Laporan sudah ada untuk bulan tersebut
  - **"Belum"** (abu-abu): Laporan belum ada untuk bulan tersebut
- **Actions**: Tombol untuk melihat detail perusahaan

## Fitur

### Bulk Upload

Di bagian atas halaman, terdapat card **Bulk Upload** yang memungkinkan Anda:

- Upload banyak laporan keuangan sekaligus via file Excel
- Download template Excel untuk diisi
- Validasi data sebelum upload
- Upload otomatis setelah validasi berhasil

**Catatan**: Lihat panduan lengkap di [Bulk Upload](./bulk-upload-reports)

### Pencarian

- Gunakan search box untuk mencari perusahaan berdasarkan nama
- Pencarian bersifat real-time (otomatis saat mengetik)
- Klik icon X untuk menghapus pencarian

### Filter Tahun

- Gunakan dropdown **Tahun** untuk memilih tahun laporan
- Tersedia tahun saat ini dan 5 tahun sebelumnya
- Tabel akan otomatis diperbarui saat tahun berubah

### Sorting

- Klik header kolom **Company Name** untuk sorting berdasarkan nama perusahaan (A-Z atau Z-A)
- Klik header kolom bulan untuk sorting berdasarkan status laporan (Ada/Belum)

### Pagination

- Gunakan pagination di bawah tabel untuk navigasi
- Pilih jumlah item per halaman: 10, 20, 50, atau 100
- Total jumlah perusahaan ditampilkan di pagination

## Aksi

### Lihat Detail Perusahaan

1. Klik tombol **View** di kolom Actions
2. Anda akan diarahkan ke halaman detail perusahaan
3. Di halaman detail, Anda dapat melihat:
   - Informasi lengkap perusahaan
   - Daftar laporan keuangan per periode
   - Input laporan baru
   - Edit laporan yang sudah ada

## Cara Membaca Tabel

Tabel ini menggunakan format **matrix** untuk memudahkan melihat status laporan:

- **Setiap baris** = 1 perusahaan
- **Setiap kolom bulan** = status laporan untuk bulan tersebut dalam tahun yang dipilih
- **Tag hijau "Ada"** = laporan sudah diinput untuk bulan tersebut
- **Tag abu-abu "Belum"** = laporan belum diinput untuk bulan tersebut

### Contoh

Jika Anda melihat:
- **PT ABC** - Jan: Ada, Feb: Ada, Mar: Belum, Apr: Belum, ...

Ini berarti:
- PT ABC sudah memiliki laporan untuk Januari dan Februari
- PT ABC belum memiliki laporan untuk Maret, April, dan seterusnya

## Tips

- Gunakan filter tahun untuk melihat status laporan tahun tertentu
- Gunakan search untuk menemukan perusahaan tertentu dengan cepat
- Klik "View" untuk melihat detail dan input/edit laporan di halaman detail perusahaan
- Gunakan bulk upload untuk input banyak laporan sekaligus (lebih efisien)
- Sorting berdasarkan bulan membantu melihat perusahaan mana yang masih kurang laporannya

## Status Laporan

Sistem mendukung dua jenis laporan:

- **RKAP**: Rencana Kerja dan Anggaran Perusahaan (planning)
- **Realisasi**: Realisasi aktual (actual)

Kedua jenis laporan dapat diinput untuk periode yang sama, dan sistem akan menampilkan status "Ada" jika minimal salah satu jenis laporan sudah ada.

## Referensi

- [Bulk Upload](./bulk-upload-reports) - Pelajari cara upload banyak laporan via Excel
- [Tambah Laporan](./add-report) - Pelajari cara menambah laporan baru (via halaman detail perusahaan)
- [Edit Laporan](./edit-report) - Pelajari cara edit laporan yang sudah ada
- [Detail Perusahaan](./company-detail) - Pelajari fitur-fitur di halaman detail perusahaan
