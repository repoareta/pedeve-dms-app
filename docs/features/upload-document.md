# Manajemen Dokumen - Upload Dokumen

Panduan untuk mengupload dokumen ke dalam sistem.

## Akses Halaman

- Dari halaman dokumen, klik tombol **Upload File**
- Atau akses langsung: `/documents/upload`

## Form Upload

Form terdiri dari beberapa bagian:

### Informasi Dokumen

**Field Wajib:**
- **Judul Dokumen** *: Judul atau nama dokumen
- **Tipe Dokumen** *: Pilih tipe dokumen dari dropdown
- **Folder** *: Pilih folder tempat menyimpan dokumen
- **Status** *: Aktif atau Nonaktif

**Field Opsional:**
- **Nomor Referensi**: Nomor referensi dokumen (akan ter-generate otomatis jika kosong)
- **Unit**: Unit atau departemen terkait
- **Tanggal Terbit**: Tanggal dokumen diterbitkan
- **Tanggal Efektif**: Tanggal dokumen mulai efektif
- **Tanggal Kedaluwarsa**: Tanggal dokumen kedaluwarsa
- **Status Aktif**: Status aktif dokumen

### Upload File

**Cara Upload:**

1. **Drag & Drop**
   - Drag file dari komputer ke area upload
   - Drop file di area yang ditandai

2. **Browse File**
   - Klik area upload atau tombol "Browse file to upload"
   - Pilih file dari dialog

**Format File yang Didukung:**
- **Dokumen**: PDF, DOCX, XLSX, XLS, PPTX, PPT
- **Gambar**: JPG, JPEG, PNG

**Batasan Ukuran:**
- **Dokumen** (PDF, DOCX, XLSX, dll): **Tidak ada batasan**
- **Gambar** (JPG, PNG): **Maksimal 10MB**

**Catatan:**
- Upload **satu file** per form
- Untuk upload banyak file sekaligus, gunakan fitur batch upload di halaman folder detail

### Generate Nomor Referensi Otomatis

Nomor referensi dapat di-generate otomatis:
- Format: `{TIPE_DOKUMEN}/{TAHUN}/{BULAN}/{NOMOR_URUT}`
- Contoh: `INVOICE/2025/01/001`
- Generate otomatis jika field "Nomor Referensi" kosong

## Validasi

Sistem akan memvalidasi:
- Judul dokumen wajib diisi
- Tipe dokumen wajib dipilih
- Folder wajib dipilih
- File wajib diupload
- Format file harus sesuai
- Ukuran file gambar tidak melebihi 10MB

## Menyimpan Dokumen

1. **Isi Form**
   - Isi semua field yang diperlukan
   - Upload file dokumen

2. **Review**
   - Pastikan semua data sudah benar
   - Cek file yang akan diupload

3. **Klik "Upload"**
   - File akan diupload dan metadata disimpan
   - Jika berhasil, Anda akan diarahkan ke halaman daftar dokumen

4. **Jika Ada Error**
   - Pesan error akan ditampilkan
   - Perbaiki field yang error
   - Coba upload lagi

## Batch Upload (Upload Banyak File)

Untuk upload banyak file sekaligus:

1. **Buka Folder Detail**
   - Klik folder di halaman dokumen
   - Atau akses: `/documents/folders/{folder-id}`

2. **Upload Area**
   - Di halaman folder detail, ada area upload dengan label "Multiple files supported"
   - Drag & drop atau browse beberapa file sekaligus

3. **Pilih File**
   - Pilih beberapa file (PDF, gambar, dll)
   - File akan muncul di list preview

4. **Klik "Upload"**
   - Sistem akan upload semua file sekaligus
   - Setiap file akan dibuat sebagai dokumen terpisah

**Catatan Batch Upload:**
- File akan menggunakan nama file sebagai judul
- Folder akan otomatis ter-set ke folder yang sedang dibuka
- Metadata lain perlu dilengkapi setelah upload (via edit dokumen)

## Tips

- Gunakan judul yang deskriptif untuk memudahkan pencarian
- Pilih folder yang sesuai untuk organisasi yang lebih baik
- Isi tanggal kedaluwarsa untuk tracking otomatis
- Generate nomor referensi otomatis untuk konsistensi
- Gunakan batch upload untuk efisiensi saat upload banyak file

## Troubleshooting

### File Tidak Bisa Upload

**Error: "File terlalu besar"**
- Untuk gambar: Pastikan ukuran tidak melebihi 10MB
- Untuk dokumen: Tidak ada batasan, cek koneksi internet
- Coba kompres gambar jika terlalu besar

**Error: "Format file tidak didukung"**
- Pastikan format file sesuai (PDF, DOCX, XLSX, JPG, PNG)
- Cek ekstensi file (case-sensitive)

**Error: "Upload gagal"**
- Cek koneksi internet
- Coba refresh halaman dan upload lagi
- Hubungi administrator jika masalah berlanjut

### Nomor Referensi Tidak Ter-generate

- Pastikan tipe dokumen sudah dipilih
- Pastikan field "Nomor Referensi" kosong (jika ingin auto-generate)
- Nomor referensi akan di-generate saat klik Upload

## Referensi

- [Detail Dokumen](./document-detail) - Lihat detail dokumen yang sudah diupload
- [Daftar Dokumen](./document-management) - Kembali ke daftar dokumen
