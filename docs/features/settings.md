# Pengaturan

Halaman ini berisi berbagai pengaturan aplikasi sesuai dengan peran Anda.

## Akses Halaman

- Klik menu **Pengaturan** di navigasi atas
- Atau akses langsung: `/settings`

## Tampilan

Halaman terdiri dari beberapa tab sesuai dengan peran:

### Tab Umum (Semua User)

- **Profil**: Edit informasi profil
- **Security**: Pengaturan keamanan (2FA, password)
- **Pengaturan Notifikasi**: Konfigurasi notifikasi dan threshold expiry

### Tab Master Data (Administrator)

- **Tipe Dokumen**: Kelola tipe dokumen
- **Jenis Pemegang Saham**: Kelola jenis pemegang saham
- **Jabatan Direktur**: Kelola jabatan direktur

### Tab Audit Logs (Administrator)

- **Audit Logs**: Log aktivitas sistem
- **User Activity**: Log aktivitas pengguna (permanent)


## Fitur per Tab

### Profil

- Edit informasi profil
- Lihat informasi akun
- Lihat perusahaan yang di-assign

### Security

- **Two-Factor Authentication (2FA)**
  - Enable/disable 2FA
  - Generate QR code
  - Backup codes

- **Password**
  - Ubah password
  - Validasi password lama

### Pengaturan Notifikasi

Halaman ini memungkinkan Anda mengatur preferensi notifikasi dan threshold untuk notifikasi dokumen yang akan expired.

#### In-App Notifications

- **Aktifkan/Nonaktifkan Push Notification**
  - Toggle untuk mengaktifkan atau menonaktifkan push notification yang muncul otomatis di pojok kanan atas
  - Jika dinonaktifkan, push notification tidak akan muncul, tetapi icon notifikasi dan halaman notifikasi tetap berfungsi normal
  - Default: Aktif

#### Jumlah Hari Sebelum Expired

- **Threshold untuk Notifikasi**
  - Atur berapa hari sebelum dokumen atau masa jabatan expired, sistem akan membuat notifikasi pertama kali
  - Default: 14 hari
  - Range: 1-365 hari

**Cara Kerja:**
- Jika threshold diatur ke 14 hari, sistem akan membuat notifikasi untuk dokumen yang akan expired dalam 14 hari ke depan
- Notifikasi dibuat sekali per dokumen/jabatan
- Push notification akan muncul berulang setiap 2 menit sampai notifikasi ditandai sebagai "sudah ditindak lanjuti"

**Contoh:**
- Threshold: 30 hari
- Dokumen akan expired: 25 hari lagi → **Notifikasi dibuat**
- Dokumen akan expired: 35 hari lagi → **Notifikasi belum dibuat** (masih di luar threshold)

**Catatan:**
- Threshold ini berlaku untuk semua dokumen dan masa jabatan direktur
- Semakin besar threshold, semakin awal notifikasi dibuat
- Dokumen yang sudah expired akan tetap mendapat notifikasi sampai ditindak lanjuti

### Master Data

#### Tipe Dokumen

- **Tambah Tipe Dokumen**
  1. Klik "Tambah Tipe Dokumen"
  2. Masukkan nama tipe
  3. Simpan

- **Edit Tipe Dokumen**
  1. Klik tombol edit di tipe dokumen
  2. Ubah nama
  3. Simpan

- **Hapus Tipe Dokumen**
  1. Klik tombol hapus
  2. Konfirmasi penghapusan

**Catatan**: Tipe dokumen yang sudah digunakan tidak bisa dihapus

#### Jenis Pemegang Saham

- Kelola jenis pemegang saham (Perusahaan, Individu, dll)
- Tambah, edit, atau hapus jenis

#### Jabatan Direktur

- Kelola jabatan direktur (Direktur Utama, Direktur, Komisaris, dll)
- Tambah, edit, atau hapus jabatan

### Audit Logs

#### Audit Logs Tab

- **Filter**
  - Filter berdasarkan action, resource, status
  - Filter berdasarkan log type (user action / technical error)

- **Tabel**
  - Menampilkan log dengan kolom: Waktu, User, Action, Resource, Status
  - Klik log untuk melihat detail

- **Pagination**
  - Navigasi halaman log
  - Pilih jumlah item per halaman

#### User Activity Tab

- **Permanent Logs**
  - Log aktivitas untuk data penting (Report, Document, Company, User)
  - Disimpan permanen tanpa retention policy

- **Filter**
  - Filter berdasarkan action, resource, status

- **Detail**
  - Klik log untuk melihat detail lengkap


## Tips

- Gunakan master data untuk konsistensi data
- Cek audit logs untuk tracking aktivitas

## Troubleshooting

### Tidak Bisa Akses Tab Tertentu

- Pastikan Anda memiliki role yang sesuai
- Beberapa tab hanya untuk administrator
- Hubungi administrator jika seharusnya memiliki akses

### Master Data Tidak Bisa Dihapus

- Pastikan data tidak sedang digunakan
- Data yang sudah digunakan tidak bisa dihapus
- Hapus penggunaan data terlebih dahulu

## Referensi

- [Two-Factor Authentication](./2fa) - Setup 2FA
- [Profil](./profile) - Edit profil
