# Manajemen Perusahaan - Edit Perusahaan

Panduan untuk mengedit data perusahaan yang sudah ada.

## Akses Halaman

- Dari halaman detail perusahaan, klik tombol **Edit**
- Atau akses langsung: `/subsidiaries/{id}/edit`

## Form Edit

Form edit sama dengan form tambah, tetapi sudah terisi dengan data yang ada.

### Tab 1: Informasi Umum

**Field yang Dapat Diedit:**
- Nama Perusahaan
- Kode Perusahaan (jika belum digunakan perusahaan lain)
- Perusahaan Induk
- Bidang Usaha
- Status
- Alamat, Telepon, Email, Website
- Logo (dapat diubah)

### Tab 2: Pemegang Saham

**Menambah Pemegang Saham:**
1. Klik "Tambah Pemegang Saham"
2. Isi form seperti saat menambah perusahaan baru
3. Persentase akan terhitung ulang otomatis

**Edit Pemegang Saham:**
1. Klik pada row pemegang saham yang ingin diedit
2. Ubah data yang diperlukan
3. **Perhitungan Ulang Otomatis**: Sistem akan secara otomatis menghitung ulang semua persentase kepemilikan setelah Anda mengubah data pemegang saham. Ini termasuk:
   - Jika Anda mengubah pemegang saham dari satu perusahaan ke perusahaan lain, sistem akan mengambil Modal Disetor dari perusahaan baru dan menghitung ulang persentase
   - Jika Anda mengubah pemegang saham dari perusahaan menjadi individu (atau sebaliknya), sistem akan menyesuaikan metode perhitungan (otomatis untuk perusahaan, manual untuk individu)
   - Semua persentase akan dihitung ulang untuk memastikan total tetap 100%

**Hapus Pemegang Saham:**
1. Klik tombol hapus (X) di row pemegang saham
2. Konfirmasi penghapusan
3. **Perhitungan Ulang Otomatis**: Setelah menghapus pemegang saham, sistem akan:
   - Menghitung ulang semua persentase kepemilikan yang tersisa
   - Memastikan total persentase tetap 100%
   - Menyesuaikan persentase perusahaan sendiri jika diperlukan
   - Menampilkan peringatan jika setelah penghapusan, total persentase tidak mencapai 100%

### Tab 3: Dewan Direksi

**Menambah Direktur:**
1. Klik "Tambah Direktur"
2. Isi form dengan data direktur baru
3. Upload dokumen pendukung jika ada

**Edit Direktur:**
1. Klik pada row direktur yang ingin diedit
2. Ubah data yang diperlukan
3. Upload dokumen baru atau hapus dokumen lama

**Hapus Direktur:**
1. Klik tombol hapus di row direktur
2. Konfirmasi penghapusan

**Edit Dokumen Direktur:**

Dokumen direktur adalah file attachment yang terkait dengan data direktur, seperti:
- Surat keputusan pengangkatan
- Dokumen identitas (KTP, NPWP)
- Dokumen pendukung lainnya

**Cara Upload Dokumen:**

1. **Buka Modal Upload**
   - Klik icon attachment (📎) di kolom "Aksi" pada row direktur yang ingin dilampirkan dokumennya
   - Modal "Upload Dokumen Individu" akan terbuka

2. **Pilih Kategori Dokumen**
   - Pilih kategori dokumen dari dropdown (wajib diisi)
   - Kategori membantu mengorganisir dokumen berdasarkan jenisnya

3. **Pilih File untuk Upload**
   - Klik tombol "Pilih File" atau drag & drop file ke area upload
   - Anda dapat memilih multiple files sekaligus
   - Format file yang diizinkan:
     - **Dokumen**: DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF
     - **Gambar**: JPG, JPEG, PNG
   - **Ukuran maksimal**: 50MB per file

4. **Simpan ke Daftar**
   - Setelah memilih file, klik tombol untuk menyimpan file ke daftar
   - File akan ditambahkan ke daftar "Dokumen yang Sudah Di-upload"
   - **Catatan Penting**: File belum langsung diupload ke server. File akan diupload saat Anda menyimpan form perusahaan (klik "Update")

5. **Hapus Dokumen**
   - Klik icon hapus (🗑️) di dokumen yang ingin dihapus
   - Konfirmasi penghapusan
   - Dokumen akan dihapus dari daftar

**Cara Kerja Upload:**

- File yang dipilih akan disimpan sementara di daftar "pending files"
- File akan diupload ke server saat Anda menyimpan form perusahaan
- Setelah form disimpan, file akan terhubung dengan direktur yang bersangkutan
- File dapat dilihat di halaman detail perusahaan di tab "Profile" > "Pengurus/Dewan Direksi"

## Validasi

Sama seperti form tambah, sistem akan memvalidasi:
- Kode perusahaan harus unik (jika diubah)
- Persentase kepemilikan harus total 100%
- Format file dokumen harus sesuai
- Tanggal berakhir harus setelah tanggal mulai

## Menyimpan Perubahan

1. **Review Perubahan**
   - Pastikan semua perubahan sudah benar
   - Cek persentase kepemilikan masih 100%

2. **Klik "Update"**
   - Sistem akan menyimpan perubahan
   - Jika berhasil, Anda akan diarahkan ke halaman detail

3. **Jika Ada Error**
   - Sistem akan menampilkan pesan error
   - Perbaiki field yang error
   - Coba simpan lagi

## Tips

- Pastikan persentase kepemilikan tetap 100% setelah edit
- Backup data penting sebelum melakukan perubahan besar
- Gunakan preview untuk melihat perubahan sebelum simpan
- Simpan dokumen penting sebelum menghapus

## Troubleshooting

### Tidak Bisa Edit Kode Perusahaan

- Kode perusahaan yang sudah digunakan perusahaan lain tidak bisa diubah
- Jika perlu mengubah, hubungi administrator

### Persentase Tidak 100% Setelah Edit

Jika setelah mengedit data perusahaan, total persentase kepemilikan tidak mencapai 100%, ikuti langkah-langkah berikut:

1. **Pahami Perubahan yang Terjadi**
   - Setiap perubahan pada pemegang saham atau Modal Disetor akan memicu perhitungan ulang
   - Jika Anda menghapus pemegang saham, persentase yang tersisa akan dihitung ulang
   - Jika Anda menambah pemegang saham baru, semua persentase akan disesuaikan

2. **Cek Modal Disetor Perusahaan Sendiri**
   - Pastikan Modal Disetor perusahaan sendiri masih terisi dengan benar
   - Jika Anda mengubah Modal Disetor perusahaan sendiri, semua persentase akan dihitung ulang
   - Pastikan nilai yang diisi sudah benar

3. **Cek Modal Disetor Pemegang Saham Perusahaan**
   - Jika Anda mengubah pemegang saham dari satu perusahaan ke perusahaan lain, pastikan perusahaan baru sudah memiliki Modal Disetor
   - Jika perusahaan pemegang saham belum memiliki Modal Disetor, sistem akan menghitung persentase sebagai 0%
   - Anda perlu mengisi Modal Disetor perusahaan pemegang saham terlebih dahulu, atau input persentase secara manual

4. **Cek Persentase untuk Pemegang Saham Individu**
   - Jika ada pemegang saham individu atau eksternal, pastikan persentase mereka sudah diisi dengan benar
   - Persentase untuk individu harus diisi secara manual karena tidak ada perhitungan otomatis

5. **Verifikasi Total**
   - Gunakan popover informasi di kolom "Persentase Kepemilikan" untuk melihat detail perhitungan
   - Pastikan total dari semua persentase (termasuk perusahaan sendiri) mencapai 100%
   - Sistem akan menampilkan peringatan jika total tidak 100%

6. **Contoh Skenario**

   **Skenario 1: Menambah Pemegang Saham Baru**
   - Sebelum: Perusahaan sendiri 50%, Pemegang Saham A 30%, Pemegang Saham B 20% (Total: 100%)
   - Setelah menambah Pemegang Saham C dengan Modal Disetor Rp 100.000.000:
     - Sistem akan menghitung ulang semua persentase berdasarkan total modal baru
     - Jika total modal baru menjadi Rp 1.100.000.000, maka:
       - Perusahaan sendiri: (Rp 500.000.000 ÷ Rp 1.100.000.000) × 100% = 45.45%
       - Pemegang Saham A: (Rp 300.000.000 ÷ Rp 1.100.000.000) × 100% = 27.27%
       - Pemegang Saham B: (Rp 200.000.000 ÷ Rp 1.100.000.000) × 100% = 18.18%
       - Pemegang Saham C: (Rp 100.000.000 ÷ Rp 1.100.000.000) × 100% = 9.09%
     - Total: 45.45% + 27.27% + 18.18% + 9.09% = 99.99% (pembulatan, sebenarnya 100%)

   **Skenario 2: Menghapus Pemegang Saham**
   - Sebelum: Perusahaan sendiri 50%, Pemegang Saham A 30%, Pemegang Saham B 20% (Total: 100%)
   - Setelah menghapus Pemegang Saham B:
     - Sistem akan menghitung ulang berdasarkan total modal yang tersisa
     - Total modal baru = Rp 500.000.000 + Rp 300.000.000 = Rp 800.000.000
     - Perusahaan sendiri: (Rp 500.000.000 ÷ Rp 800.000.000) × 100% = 62.5%
     - Pemegang Saham A: (Rp 300.000.000 ÷ Rp 800.000.000) × 100% = 37.5%
     - Total: 62.5% + 37.5% = 100%

   **Skenario 3: Mengubah Modal Disetor Perusahaan Sendiri**
   - Sebelum: Perusahaan sendiri 50% (Modal Disetor Rp 500.000.000), Pemegang Saham A 30%, Pemegang Saham B 20%
   - Setelah mengubah Modal Disetor perusahaan sendiri menjadi Rp 600.000.000:
     - Total modal baru = Rp 600.000.000 + Rp 300.000.000 + Rp 200.000.000 = Rp 1.100.000.000
     - Perusahaan sendiri: (Rp 600.000.000 ÷ Rp 1.100.000.000) × 100% = 54.55%
     - Pemegang Saham A: (Rp 300.000.000 ÷ Rp 1.100.000.000) × 100% = 27.27%
     - Pemegang Saham B: (Rp 200.000.000 ÷ Rp 1.100.000.000) × 100% = 18.18%
     - Total: 54.55% + 27.27% + 18.18% = 99.99% (pembulatan, sebenarnya 100%)

### Dokumen Direktur Tidak Bisa Di-upload

Jika Anda mengalami masalah saat mengupload dokumen untuk direktur, ikuti langkah-langkah troubleshooting berikut:

**1. Cek Format File**

Sistem hanya menerima format file berikut:
- **Dokumen**: DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF
- **Gambar**: JPG, JPEG, PNG

**Jika file Anda tidak dalam format di atas:**
- Konversi file ke format yang diizinkan terlebih dahulu
- Contoh: Jika file Anda adalah .doc (bukan .docx), simpan ulang sebagai .docx
- Jika file Anda adalah .txt, simpan sebagai .docx atau .pdf

**2. Cek Ukuran File**

- **Ukuran maksimal**: 50MB per file
- Sistem akan menolak file yang melebihi 50MB

**Jika file Anda terlalu besar:**
- **Untuk dokumen**: Kompres file atau split menjadi beberapa file yang lebih kecil
- **Untuk gambar**: Kompres gambar menggunakan tool kompresi online atau software image editor
- **Untuk PDF**: Gunakan tool kompresi PDF untuk mengurangi ukuran file

**3. Cek Koneksi Internet**

- Pastikan koneksi internet Anda stabil
- Upload file besar memerlukan koneksi yang stabil
- Jika koneksi terputus, coba upload ulang

**4. Cek Kategori Dokumen**

- Pastikan Anda sudah memilih kategori dokumen sebelum upload
- Kategori dokumen adalah field wajib, tidak bisa dikosongkan

**5. Cek Browser dan Cache**

- Coba refresh halaman browser
- Clear cache browser jika masalah berlanjut
- Coba gunakan browser lain (Chrome, Firefox, Safari, Edge)

**6. Cek Jumlah File**

- Sistem mendukung upload multiple files sekaligus
- Namun, jika terlalu banyak file sekaligus, coba upload secara bertahap (5-10 file per batch)

**7. Pesan Error yang Mungkin Muncul**

- **"Format file tidak diizinkan"**: File Anda tidak dalam format yang diizinkan. Konversi ke format yang diizinkan.
- **"Ukuran file melebihi 50MB"**: File terlalu besar. Kompres atau split file.
- **"Pilih kategori dokumen"**: Anda belum memilih kategori dokumen. Pilih kategori terlebih dahulu.
- **"Tidak ada file yang valid untuk disimpan"**: File yang dipilih tidak valid. Coba pilih file lain.

**8. Tips untuk Upload yang Sukses**

- **Persiapkan file terlebih dahulu**: Pastikan file sudah dalam format yang benar dan ukurannya sesuai sebelum upload
- **Upload secara bertahap**: Jika banyak file, upload secara bertahap untuk menghindari timeout
- **Gunakan nama file yang jelas**: Beri nama file yang deskriptif untuk memudahkan identifikasi
- **Cek file sebelum upload**: Pastikan file tidak corrupt dan dapat dibuka dengan normal

**9. Jika Masalah Masih Berlanjut**

- Hubungi administrator sistem untuk bantuan lebih lanjut
- Sertakan informasi:
  - Format file yang dicoba diupload
  - Ukuran file
  - Pesan error yang muncul (jika ada)
  - Screenshot error (jika memungkinkan)

## Langkah Selanjutnya

- [Detail Perusahaan](./company-detail) - Lihat perubahan yang sudah disimpan
- [Daftar Perusahaan](./company-management) - Kembali ke daftar perusahaan
