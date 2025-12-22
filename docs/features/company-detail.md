# Manajemen Perusahaan - Detail Perusahaan

Halaman detail menampilkan informasi lengkap tentang perusahaan.

## Akses Halaman

- Klik nama perusahaan di daftar perusahaan
- Atau akses langsung: `/subsidiaries/{id}`

## Tampilan Detail

Halaman detail terdiri dari beberapa tab:

### Tab 1: Profile

Menampilkan informasi umum perusahaan:

#### Informasi Dasar
- **Nama Perusahaan**: Nama lengkap
- **Kode**: Kode unik perusahaan
- **Perusahaan Induk**: Nama perusahaan induk (jika ada)
- **Level**: Level hierarki
- **Bidang Usaha**: Bidang usaha perusahaan
- **Status**: Aktif atau Nonaktif
- **Logo**: Logo perusahaan (jika ada)

#### Struktur Kepemilikan

Tabel menampilkan informasi lengkap tentang struktur kepemilikan saham perusahaan:

- **Nama Pemegang Saham**: Nama pemegang saham (perusahaan atau individu)
- **Modal Dasar**: Modal dasar pemegang saham (authorized capital - batas maksimal modal yang diizinkan)
- **Modal Disetor**: Modal yang sudah benar-benar disetor dan dibayar oleh pemegang saham (paid-up capital)
- **Persentase Kepemilikan**: Persentase kepemilikan saham yang dihitung berdasarkan Modal Disetor (dengan presisi 10 digit desimal untuk akurasi maksimal)

**Cara Membaca Persentase Kepemilikan:**

Persentase kepemilikan menunjukkan proporsi kepemilikan saham dari setiap pemegang saham terhadap total modal perusahaan. Persentase ini dihitung dengan rumus:

```
Persentase Kepemilikan = (Modal Disetor Pemegang Saham ÷ Total Modal Disetor) × 100%
```

Dimana:
- **Total Modal Disetor** = Modal Disetor Perusahaan Sendiri + Total Modal Disetor Semua Pemegang Saham

**Contoh Interpretasi:**

Jika Anda melihat:
- **PT ABC (Perusahaan Sendiri)**: 50.0000000000%
- **PT XYZ**: 30.0000000000%
- **PT DEF**: 20.0000000000%

Ini berarti:
- PT ABC memiliki 50% saham dari perusahaan ini
- PT XYZ memiliki 30% saham
- PT DEF memiliki 20% saham
- Total kepemilikan: 100%

**Fitur:**
- **Klik Nama Pemegang Saham**: Jika pemegang saham adalah perusahaan yang terdaftar di sistem, klik nama untuk melihat detail lengkap perusahaan tersebut (akan terbuka di tab baru browser)
- **Perusahaan Sendiri**: Tabel selalu menampilkan baris untuk perusahaan sendiri dengan persentase kepemilikannya. Ini penting untuk menunjukkan proporsi kepemilikan yang dimiliki oleh perusahaan tersebut.
- **Presisi Tinggi**: Persentase ditampilkan dengan 10 digit desimal untuk memastikan akurasi perhitungan, terutama untuk pembagian yang tidak bulat (misalnya: 33.3333333333% untuk pembagian tiga sama besar)
- **Hover Effect**: Arahkan mouse ke baris tabel untuk melihat highlight, memudahkan membaca data

#### Pengurus/Dewan Direksi

Tabel menampilkan:
- **Nama**: Nama direktur
- **Jabatan**: Jabatan direktur
- **Tanggal Mulai**: Tanggal mulai menjabat
- **Tanggal Berakhir**: Tanggal berakhir masa jabatan
- **Dokumen**: File attachment (jika ada)

**Fitur:**
- Klik row untuk expand dan melihat dokumen attachment
- Klik icon mata (👁️) untuk preview dokumen (gambar/PDF)
- Klik icon download untuk download dokumen
- Preview dokumen dalam modal fullscreen

### Tab 2: Laporan Keuangan

Menampilkan daftar laporan keuangan perusahaan:
- Filter berdasarkan periode
- Tabel dengan kolom: Periode, Revenue, Opex, NPAT, dll
- Tombol untuk tambah/edit laporan

### Tab 3: Dokumen

Menampilkan dokumen terkait perusahaan:
- Filter berdasarkan folder
- Tabel dengan kolom: Nama, Tipe, Tanggal Upload, dll
- Tombol untuk upload dokumen baru

### Tab 4: Informasi Lainnya

Informasi tambahan:
- Alamat lengkap
- Kontak (telepon, email, website)
- Catatan tambahan

## Aksi

### Edit Perusahaan

1. Klik tombol **Edit** di pojok kanan atas
2. Anda akan diarahkan ke halaman edit
3. **Catatan**: Hanya admin/administrator yang dapat edit

### Lihat Perusahaan Induk

1. Klik nama perusahaan induk di bagian "Perusahaan Induk"
2. Anda akan diarahkan ke detail perusahaan induk

### Lihat Detail Pemegang Saham

1. Klik nama pemegang saham di tabel "Struktur Kepemilikan"
2. Detail perusahaan pemegang saham akan terbuka di tab baru

### Preview Dokumen Direktur

1. Klik row direktur yang memiliki attachment
2. Row akan expand menampilkan dokumen
3. Klik icon mata (👁️) untuk preview
4. Modal fullscreen akan muncul dengan preview dokumen

**Catatan:**
- Preview hanya untuk gambar (JPG, PNG) dan PDF
- File lain (DOCX, XLSX) hanya bisa di-download

### Download Dokumen

1. Klik icon download di dokumen attachment
2. File akan terdownload ke komputer Anda

## Tips

- Gunakan tab untuk navigasi cepat ke informasi yang diinginkan
- Klik nama pemegang saham untuk melihat struktur kepemilikan lebih detail
- Preview dokumen sebelum download untuk memastikan file yang benar
- Gunakan filter di tab Laporan dan Dokumen untuk mencari data spesifik

## Troubleshooting

### Dokumen Tidak Bisa Di-preview

- Pastikan file adalah gambar (JPG, PNG) atau PDF
- File DOCX/XLSX hanya bisa di-download, tidak bisa di-preview
- Cek koneksi internet jika preview tidak muncul

### Persentase Kepemilikan Tidak Tampil atau Tidak Benar

Jika persentase kepemilikan tidak tampil atau tampil sebagai 0%, ikuti langkah-langkah berikut:

1. **Cek Data Modal Disetor**
   - Pastikan Modal Disetor perusahaan sendiri sudah diisi dengan benar di halaman edit perusahaan
   - Pastikan Modal Disetor untuk setiap pemegang saham (jika pemegang saham adalah perusahaan) sudah diisi
   - Jika Modal Disetor belum diisi, sistem akan menghitung persentase sebagai 0%

2. **Cek Data Pemegang Saham**
   - Pastikan data pemegang saham sudah lengkap dan benar
   - Untuk pemegang saham yang merupakan perusahaan, pastikan perusahaan tersebut sudah memiliki data Modal Disetor di sistem
   - Jika perusahaan pemegang saham belum memiliki Modal Disetor, persentase akan ditampilkan sebagai 0%

3. **Refresh Halaman**
   - Refresh halaman browser untuk memastikan data terbaru dimuat
   - Sistem akan menghitung ulang persentase setiap kali data dimuat

4. **Edit dan Simpan Ulang**
   - Jika persentase masih tidak benar, coba edit perusahaan dan simpan ulang
   - Sistem akan menghitung ulang semua persentase saat menyimpan data

5. **Verifikasi Perhitungan Manual**
   - Jika perlu, verifikasi perhitungan secara manual:
     - Total Modal = Modal Disetor Perusahaan Sendiri + Total Modal Disetor Semua Pemegang Saham
     - Persentase = (Modal Disetor Pemegang Saham ÷ Total Modal) × 100%
   - Pastikan total semua persentase mencapai 100%

## Langkah Selanjutnya

- [Edit Perusahaan](./edit-company) - Pelajari cara edit perusahaan
- [My Company](./my-company) - Lihat perusahaan yang di-assign ke Anda
