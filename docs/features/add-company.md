# Manajemen Perusahaan - Tambah Perusahaan

Panduan untuk menambahkan perusahaan baru ke dalam sistem.

## Akses Halaman

- Dari halaman daftar perusahaan, klik tombol **Tambah Perusahaan**
- Atau akses langsung: `/subsidiaries/new`

## Form Tambah Perusahaan

Form terdiri dari beberapa tab:

### Tab 1: Informasi Umum

**Field Wajib:**
- **Nama Perusahaan** *: Nama lengkap perusahaan
- **Kode Perusahaan** *: Kode unik perusahaan (harus unik)
- **Perusahaan Induk**: Pilih perusahaan induk (opsional untuk holding)
- **Bidang Usaha**: Pilih bidang usaha perusahaan
- **Status**: Aktif atau Nonaktif

**Field Opsional:**
- **Alamat**: Alamat lengkap perusahaan
- **Telepon**: Nomor telepon
- **Email**: Email perusahaan
- **Website**: Website perusahaan
- **Logo**: Upload logo perusahaan

### Tab 2: Pemegang Saham

Menambahkan pemegang saham perusahaan:

1. **Klik "Tambah Pemegang Saham"**
2. **Isi Form:**
   - **Nama Pemegang Saham**: Pilih dari daftar perusahaan atau masukkan nama individu
   - **Jenis Pemegang Saham**: Pilih jenis (Perusahaan, Individu, dll)
   - **Nomor Identitas**: NPWP untuk perusahaan, NIK untuk individu
   - **Modal Dasar**: Modal dasar pemegang saham
   - **Modal Disetor**: Modal yang sudah disetor
   - **Persentase Kepemilikan**: Akan terhitung otomatis berdasarkan modal

3. **Persentase Kepemilikan**

   Persentase kepemilikan saham dihitung secara otomatis oleh sistem berdasarkan **Modal Disetor** dari setiap pemegang saham. Sistem menggunakan perhitungan yang presisi hingga 10 digit desimal untuk memastikan akurasi.

   **Konsep Dasar:**
   
   - **Modal Dasar**: Modal yang diizinkan oleh anggaran dasar perusahaan (authorized capital). Ini adalah batas maksimal modal yang dapat dikeluarkan oleh perusahaan.
   - **Modal Disetor**: Modal yang benar-benar sudah disetor dan dibayar oleh pemegang saham (paid-up capital). Ini adalah modal yang sebenarnya dimiliki perusahaan.
   - **Persentase Kepemilikan**: Proporsi kepemilikan saham yang dihitung berdasarkan Modal Disetor, bukan Modal Dasar.

   **Cara Perhitungan:**

   Sistem menghitung persentase kepemilikan dengan rumus berikut:

   ```
   Total Modal = Modal Disetor Perusahaan Sendiri + Total Modal Disetor Semua Pemegang Saham
   
   Persentase Kepemilikan Pemegang Saham = (Modal Disetor Pemegang Saham ÷ Total Modal) × 100%
   
   Persentase Kepemilikan Perusahaan Sendiri = (Modal Disetor Perusahaan Sendiri ÷ Total Modal) × 100%
   ```

   **Contoh Perhitungan:**

   Misalkan Anda memiliki perusahaan dengan struktur kepemilikan berikut:
   
   - **Perusahaan Sendiri (PT ABC)**:
     - Modal Dasar: Rp 1.000.000.000
     - Modal Disetor: Rp 500.000.000
   
   - **Pemegang Saham 1 (PT XYZ)**:
     - Modal Disetor: Rp 300.000.000
   
   - **Pemegang Saham 2 (PT DEF)**:
     - Modal Disetor: Rp 200.000.000

   **Perhitungan:**
   
   1. Total Modal = Rp 500.000.000 + Rp 300.000.000 + Rp 200.000.000 = Rp 1.000.000.000
   
   2. Persentase Kepemilikan PT ABC (Perusahaan Sendiri):
      = (Rp 500.000.000 ÷ Rp 1.000.000.000) × 100%
      = 50%
   
   3. Persentase Kepemilikan PT XYZ:
      = (Rp 300.000.000 ÷ Rp 1.000.000.000) × 100%
      = 30%
   
   4. Persentase Kepemilikan PT DEF:
      = (Rp 200.000.000 ÷ Rp 1.000.000.000) × 100%
      = 20%
   
   5. **Total Persentase**: 50% + 30% + 20% = 100% ✓

   **Catatan Penting:**

   - **Perhitungan Otomatis**: Untuk pemegang saham yang merupakan perusahaan (bukan individu), persentase akan dihitung otomatis oleh sistem berdasarkan Modal Disetor perusahaan tersebut. Anda tidak perlu mengisi persentase secara manual.
   
   - **Input Manual untuk Individu**: Jika pemegang saham adalah individu atau entitas eksternal (bukan perusahaan di sistem), Anda perlu mengisi persentase secara manual karena sistem tidak memiliki data Modal Disetor untuk entitas tersebut.
   
   - **Total Harus 100%**: Sistem akan memastikan bahwa total persentase kepemilikan selalu 100%. Jika total tidak mencapai 100%, sistem akan menampilkan peringatan dan tidak akan mengizinkan penyimpanan data.
   
   - **Perhitungan Real-time**: Setiap kali Anda menambah, mengubah, atau menghapus pemegang saham, atau mengubah Modal Disetor, sistem akan secara otomatis menghitung ulang semua persentase kepemilikan secara real-time.
   
   - **Presisi Tinggi**: Sistem menggunakan presisi 10 digit desimal untuk perhitungan, sehingga dapat menangani perhitungan yang sangat detail dan akurat. Contoh: 33.3333333333% untuk pembagian yang tidak bulat.
   
   - **Perusahaan Sendiri Termasuk**: Perusahaan yang sedang Anda input juga dihitung sebagai pemegang saham dengan persentase kepemilikannya sendiri. Ini penting untuk menunjukkan proporsi kepemilikan yang dimiliki oleh perusahaan tersebut.
   
   - **Modal Dasar vs Modal Disetor**: Perhitungan persentase hanya menggunakan Modal Disetor, bukan Modal Dasar. Modal Dasar hanya digunakan sebagai informasi referensi dan tidak mempengaruhi perhitungan persentase.

4. **Hapus Pemegang Saham**
   - Klik tombol hapus (X) untuk menghapus pemegang saham

### Tab 3: Dewan Direksi

Menambahkan dewan direksi:

1. **Klik "Tambah Direktur"**
2. **Isi Form:**
   - **Nama**: Nama lengkap direktur
   - **Jabatan**: Pilih jabatan (Direktur Utama, Direktur, Komisaris, dll)
   - **Tanggal Mulai**: Tanggal mulai menjabat
   - **Tanggal Berakhir**: Tanggal berakhir masa jabatan
   - **Dokumen**: Upload dokumen pendukung (opsional)

3. **Upload Dokumen Pendukung**

   Anda dapat mengupload dokumen pendukung untuk direktur, seperti surat keputusan pengangkatan, dokumen identitas, atau dokumen pendukung lainnya.

   **Cara Upload:**
   1. Klik icon attachment (📎) di kolom "Aksi" pada row direktur
   2. Modal "Upload Dokumen Individu" akan terbuka
   3. Pilih kategori dokumen dari dropdown (wajib)
   4. Klik "Pilih File" atau drag & drop file ke area upload
   5. Anda dapat memilih multiple files sekaligus
   6. Klik tombol untuk menyimpan file ke daftar
   7. File akan diupload saat Anda menyimpan form perusahaan

   **Format File yang Diizinkan:**
   - **Dokumen**: DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF
   - **Gambar**: JPG, JPEG, PNG

   **Ukuran Maksimal:**
   - **Semua file**: Maksimal 50MB per file
   - Tidak ada batasan jumlah file yang dapat diupload

   **Catatan Penting:**
   - File yang dipilih akan disimpan sementara di daftar "pending files"
   - File akan diupload ke server saat Anda menyimpan form perusahaan (klik "Simpan")
   - Setelah form disimpan, file akan terhubung dengan direktur yang bersangkutan
   - File dapat dilihat di halaman detail perusahaan di tab "Profile" > "Pengurus/Dewan Direksi"

4. **Hapus Direktur**
   - Klik tombol hapus untuk menghapus direktur

## Validasi

Sistem akan memvalidasi:

- Kode perusahaan harus unik
- Nama perusahaan wajib diisi
- Persentase kepemilikan harus total 100%
- Tanggal berakhir harus setelah tanggal mulai
- Format file dokumen harus sesuai

## Menyimpan Data

1. **Review Data**
   - Pastikan semua data sudah benar
   - Cek persentase kepemilikan

2. **Klik "Simpan"**
   - Sistem akan menyimpan data
   - Jika berhasil, Anda akan diarahkan ke halaman detail perusahaan

3. **Jika Ada Error**
   - Sistem akan menampilkan pesan error
   - Perbaiki field yang error
   - Coba simpan lagi

## Tips

- Pastikan kode perusahaan unik sebelum submit
- Gunakan search untuk mencari perusahaan induk dengan cepat
- Upload dokumen pendukung untuk direktur jika tersedia
- Cek persentase kepemilikan sebelum simpan

## Troubleshooting

### Kode Perusahaan Sudah Ada

- Gunakan kode yang berbeda
- Cek daftar perusahaan untuk melihat kode yang sudah digunakan

### Persentase Tidak 100%

Jika total persentase kepemilikan tidak mencapai 100%, berikut langkah-langkah untuk memperbaikinya:

1. **Cek Modal Disetor Perusahaan Sendiri**
   - Pastikan field "Modal Disetor" di Tab 1 (Informasi Umum) sudah diisi
   - Modal Disetor perusahaan sendiri harus diisi untuk menghitung persentase kepemilikan sendiri

2. **Cek Modal Disetor Pemegang Saham**
   - Untuk pemegang saham yang merupakan perusahaan, pastikan perusahaan tersebut sudah memiliki data Modal Disetor di sistem
   - Jika perusahaan pemegang saham belum memiliki Modal Disetor, sistem akan menghitung persentase sebagai 0%
   - Anda perlu mengisi Modal Disetor perusahaan pemegang saham terlebih dahulu, atau input persentase secara manual

3. **Cek Input Manual untuk Individu**
   - Untuk pemegang saham individu atau eksternal, pastikan Anda sudah mengisi persentase secara manual
   - Pastikan total persentase dari semua pemegang saham (termasuk perusahaan sendiri) mencapai 100%

4. **Verifikasi Perhitungan**
   - Sistem akan menampilkan peringatan jika total persentase tidak 100%
   - Gunakan popover informasi di kolom "Persentase Kepemilikan" untuk melihat detail perhitungan
   - Pastikan semua angka Modal Disetor sudah benar sebelum menyimpan

5. **Contoh Masalah dan Solusi**
   
   **Masalah**: Total persentase hanya 80%
   
   **Kemungkinan Penyebab**:
   - Modal Disetor perusahaan sendiri belum diisi
   - Salah satu pemegang saham perusahaan belum memiliki Modal Disetor
   - Persentase untuk pemegang saham individu belum diisi
   
   **Solusi**:
   - Isi Modal Disetor perusahaan sendiri
   - Pastikan semua perusahaan pemegang saham sudah memiliki Modal Disetor
   - Atau, untuk pemegang saham individu, isi persentase secara manual hingga total mencapai 100%

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

- [Detail Perusahaan](./company-detail) - Lihat detail perusahaan yang baru dibuat
- [Edit Perusahaan](./edit-company) - Pelajari cara edit perusahaan
