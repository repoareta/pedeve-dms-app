# Halaman Daftar Anak Perusahaan (Subsidiaries)

Halaman Daftar Anak Perusahaan adalah halaman utama yang ditampilkan setelah Anda berhasil login. Halaman ini menampilkan daftar semua anak perusahaan yang dapat Anda akses beserta ringkasan metrik keuangan penting.

## Akses Halaman

Setelah login, Anda akan otomatis diarahkan ke halaman Daftar Anak Perusahaan. Anda juga dapat mengakses halaman ini melalui:

- Klik menu **Subsidiaries** di navigasi atas
- Atau klik logo aplikasi di pojok kiri atas

## Tampilan Halaman

Halaman Daftar Anak Perusahaan memiliki dua mode tampilan:

### 1. Tampilan Grid (Card View)

Tampilan grid menampilkan setiap perusahaan dalam bentuk card yang menampilkan:

- **Logo Perusahaan**: Logo perusahaan atau inisial dengan warna unik
- **Nama Perusahaan**: Nama lengkap perusahaan
- **Nomor Registrasi**: NIB (Nomor Induk Berusaha)
- **Metrik Keuangan Terbaru**:
  - **Net Profit**: Laba bersih dengan persentase perubahan dari periode sebelumnya
  - **Financial Health Score**: Skor kesehatan keuangan (0-100) dengan status (Excellent, Good, Fair, Poor)
  - **Revenue**: Total pendapatan dengan persentase perubahan
  - **Operating Expenses**: Total biaya operasional dengan persentase perubahan
- **Informasi Periode**: Periode data terbaru yang ditampilkan
- **Tombol "Learn more"**: Untuk melihat detail lengkap perusahaan

### 2. Tampilan List (Table View)

Tampilan list menampilkan perusahaan dalam bentuk tabel dengan kolom-kolom:

- **Logo**: Logo atau inisial perusahaan
- **Nama Perusahaan**: Nama lengkap perusahaan
- **Level**: Level perusahaan dalam hierarki (Level 1, 2, 3, dst.)
- **Status**: Status aktif/tidak aktif
- **Aksi**: Menu dropdown dengan opsi:
  - **Lihat Detail**: Membuka halaman detail perusahaan
  - **Edit**: Mengedit data perusahaan (jika memiliki akses)
  - **Assign Role**: Menetapkan role untuk perusahaan (hanya admin/administrator)
  - **Hapus**: Menghapus perusahaan (hanya admin/administrator)

## Fitur Halaman

### Pencarian (Search)

- **Tampilan Grid**: Kotak pencarian di header kanan
- **Tampilan List**: Kotak pencarian di atas tabel
- Pencarian akan memfilter perusahaan berdasarkan nama perusahaan
- Pencarian bersifat real-time (langsung memfilter saat mengetik)

### Switch Tampilan

Anda dapat beralih antara tampilan Grid dan List menggunakan tombol di header:

- **Tombol Grid** (ikon grid): Menampilkan dalam bentuk card
- **Tombol List** (ikon list): Menampilkan dalam bentuk tabel
- Preferensi tampilan akan disimpan di browser Anda

### Tambah Anak Perusahaan Baru

Tombol **"Add new Subsidiary"** tersedia di header (hanya untuk admin atau administrator):

- Klik tombol untuk membuka form tambah perusahaan baru
- Form akan mengarahkan ke halaman form tambah perusahaan

### Pagination

- **Tampilan Grid**: Menampilkan 8 perusahaan per halaman dengan navigasi halaman
- **Tampilan List**: Menampilkan 10 perusahaan per halaman dengan navigasi halaman
- Gunakan tombol navigasi untuk berpindah halaman

## Informasi pada Card Perusahaan

Setiap card perusahaan menampilkan informasi keuangan terbaru:

### Net Profit (Laba Bersih)

- Menampilkan nilai laba bersih dalam format mata uang
- Menampilkan periode data (bulan dan tahun)
- Menampilkan persentase perubahan dari periode sebelumnya:
  - **Hijau dengan tanda +**: Peningkatan laba
  - **Merah dengan tanda -**: Penurunan laba

### Financial Health Score

- Skor kesehatan keuangan dari 0-100
- Status kesehatan:
  - **Excellent** (80-100): Warna hijau
  - **Good** (60-79): Warna biru
  - **Fair** (40-59): Warna kuning
  - **Poor** (0-39): Warna merah
- Hover pada ikon informasi untuk melihat detail perhitungan skor

### Revenue (Pendapatan)

- Total pendapatan dalam format mata uang
- Persentase perubahan dari periode sebelumnya

### Operating Expenses (Biaya Operasional)

- Total biaya operasional dalam format mata uang
- Persentase perubahan dari periode sebelumnya

## Akses Berdasarkan Role

### Administrator / Admin

- Dapat melihat semua perusahaan
- Dapat menambah perusahaan baru
- Dapat mengedit semua perusahaan
- Dapat menghapus perusahaan
- Dapat assign role untuk perusahaan

### Manager

- Dapat melihat perusahaan yang di-assign
- Dapat mengedit perusahaan yang di-assign
- Tidak dapat menambah atau menghapus perusahaan

### Staff

- Dapat melihat perusahaan yang di-assign
- Dapat mengedit perusahaan yang di-assign
- Tidak dapat menambah atau menghapus perusahaan

## Navigasi Cepat

Dari halaman Daftar Anak Perusahaan, Anda dapat:

- **Lihat Detail Perusahaan**: Klik pada card perusahaan atau tombol "Learn more"
- **Edit Perusahaan**: Klik menu "Edit" di dropdown aksi (tampilan list) atau akses dari detail perusahaan
- **Tambah Perusahaan Baru**: Klik tombol "Add new Subsidiary" (jika memiliki akses)

## Tips

- Gunakan pencarian untuk menemukan perusahaan dengan cepat
- Pilih tampilan yang paling nyaman untuk Anda (Grid atau List)
- Hover pada ikon informasi di card untuk melihat detail perhitungan metrik
- Klik pada card perusahaan untuk melihat detail lengkap
- Refresh halaman untuk melihat data terbaru

## Troubleshooting

### Data Tidak Muncul

- Pastikan Anda sudah memiliki akses ke perusahaan
- Cek koneksi internet Anda
- Refresh halaman
- Hubungi administrator jika masalah berlanjut

### Perusahaan Tidak Muncul di Daftar

- Pastikan perusahaan memiliki status aktif
- Pastikan Anda memiliki akses ke perusahaan tersebut
- Coba gunakan fitur pencarian dengan nama perusahaan
- Hubungi administrator untuk memverifikasi akses Anda
