# User Acceptance Test (UAT) Document
## Pedeve DMS App

**Versi Dokumen:** 1.0  
**Tanggal:** 2024  
**Status:** Draft untuk Review

---

## 1. Informasi Umum

### 1.1 Tujuan Dokumen
Dokumen ini digunakan untuk melakukan User Acceptance Testing (UAT) pada aplikasi Pedeve DMS App. UAT dilakukan untuk memastikan bahwa aplikasi telah memenuhi kebutuhan bisnis dan dapat digunakan oleh end-user sesuai dengan requirement yang telah ditetapkan.

### 1.2 Scope Testing
- **Aplikasi:** Pedeve DMS App (Document Management System)
- **Environment:** Development / Staging / Production
- **Platform:** Web Application (Vue.js Frontend + Go Backend)
- **Browser Support:** Chrome, Firefox, Safari, Edge (latest versions)

### 1.3 Roles yang Akan Ditest
- **Superadmin** - Full access, semua fitur
- **Administrator** - Management access, semua fitur kecuali development tools
- **Admin** - Company-level management
- **Manager** - View dan limited edit
- **Staff** - View only

---

## 2. Test Scenarios

### 2.1 Authentication & Authorization

#### TC-AUTH-001: User Registration
**Tujuan:** Memastikan user baru dapat melakukan registrasi dengan sukses

**Kondisi Awal:**
- User belum memiliki akun
- Browser dapat mengakses aplikasi

**Langkah-langkah Test:**
1. Buka halaman aplikasi
2. Klik tombol "Register" atau navigasi ke `/register`
3. Isi form registrasi:
   - Username: `testuser001`
   - Email: `testuser001@example.com`
   - Password: `Password123!`
   - Confirm Password: `Password123!`
4. Klik tombol "Register"
5. Tunggu proses registrasi selesai

**Hasil yang Diharapkan:**
- ✅ User berhasil terdaftar
- ✅ Menampilkan pesan sukses "Registrasi berhasil"
- ✅ User otomatis login dan redirect ke halaman subsidiaries
- ✅ User dapat melihat dashboard

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-AUTH-002: User Login (Normal - tanpa 2FA)
**Tujuan:** Memastikan user dapat login dengan email/password yang valid

**Kondisi Awal:**
- User sudah terdaftar
- User belum enable 2FA

**Langkah-langkah Test:**
1. Buka halaman login
2. Input email: `testuser001@example.com`
3. Input password: `Password123!`
4. Klik tombol "Login"
5. Tunggu proses login

**Hasil yang Diharapkan:**
- ✅ Login berhasil
- ✅ Menampilkan pesan sukses "Login berhasil!"
- ✅ Redirect ke halaman `/subsidiaries`
- ✅ User dapat melihat dashboard dengan data sesuai role

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-AUTH-003: User Login (dengan 2FA)
**Tujuan:** Memastikan user dengan 2FA enabled dapat login dengan kode 2FA

**Kondisi Awal:**
- User sudah terdaftar
- User sudah enable 2FA
- User memiliki authenticator app (Google Authenticator, Authy, dll)

**Langkah-langkah Test:**
1. Buka halaman login
2. Input email: `testuser2fa@example.com`
3. Input password: `Password123!`
4. Klik tombol "Login"
5. Sistem menampilkan form input 2FA code
6. Buka authenticator app
7. Input 6-digit code dari authenticator app
8. Klik tombol "Verify"

**Hasil yang Diharapkan:**
- ✅ Sistem meminta kode 2FA setelah login
- ✅ Menampilkan pesan "Masukkan kode 2FA dari authenticator app Anda"
- ✅ Setelah input kode 2FA yang valid, login berhasil
- ✅ Redirect ke halaman `/subsidiaries`

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-AUTH-004: Login dengan Credentials Salah
**Tujuan:** Memastikan sistem menolak login dengan credentials yang salah

**Langkah-langkah Test:**
1. Buka halaman login
2. Input email: `wrong@example.com`
3. Input password: `WrongPassword123!`
4. Klik tombol "Login"

**Hasil yang Diharapkan:**
- ✅ Login gagal
- ✅ Menampilkan error message: "Email atau password salah"
- ✅ User tetap di halaman login
- ✅ Form tidak di-reset (user bisa coba lagi)

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-AUTH-005: User Logout
**Tujuan:** Memastikan user dapat logout dengan sukses

**Kondisi Awal:**
- User sudah login

**Langkah-langkah Test:**
1. User sudah login dan berada di halaman manapun
2. Klik tombol/logout icon di header
3. Konfirmasi logout (jika ada)
4. Tunggu proses logout

**Hasil yang Diharapkan:**
- ✅ User berhasil logout
- ✅ Session cleared
- ✅ Redirect ke halaman login
- ✅ User tidak bisa akses halaman yang memerlukan authentication

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.2 Company/Subsidiary Management

#### TC-COMP-001: Create New Subsidiary
**Tujuan:** Memastikan user dapat membuat subsidiary baru dengan lengkap

**Kondisi Awal:**
- User sudah login dengan role: Superadmin, Administrator, atau Admin
- User berada di halaman Subsidiaries

**Langkah-langkah Test:**
1. Klik tombol "Add new Subsidiary"
2. **Step 1 - Identitas Perusahaan:**
   - Isi Nama Lengkap: `PT Test Company UAT`
   - Isi Nama Singkat: `Test Co`
   - Isi NPWP: `123456789012345`
   - Isi NIB: `987654321098765`
   - Isi Deskripsi: `Perusahaan untuk testing UAT`
   - Pilih Status: `Aktif`
   - Pilih Currency: `IDR`
   - Upload logo (optional): Pilih file gambar < 5MB
3. Klik "Next" atau lanjut ke Step 2
4. **Step 2 - Struktur Kepemilikan:**
   - Klik "Tambah Pemegang Saham"
   - Pilih jenis: "Perusahaan" atau "Individu"
   - Jika Perusahaan: Pilih dari dropdown
   - Jika Individu: 
     - Isi Nama: `John Doe`
     - Isi NIK: `1234567890123456`
     - Isi Modal Dasar: `10000000000` (10M)
     - Isi Modal Disetor: `5000000000` (5M)
   - Klik "Simpan" di modal
   - Verifikasi persentase kepemilikan terhitung otomatis
5. Klik "Next" atau lanjut ke Step 3
6. **Step 3 - Bidang Usaha:**
   - Pilih Industry Sector
   - Pilih KBLI
   - Isi Main Business Activity
   - Isi Additional Activities (optional)
   - Isi Start Operation Date
7. Klik "Next" atau lanjut ke Step 4
8. **Step 4 - Pengurus/Dewan Direksi:**
   - Klik "Tambah Pengurus"
   - Isi Nama Lengkap: `Jane Doe`
   - Pilih Jabatan: `Direktur Utama`
   - Isi KTP: `9876543210987654`
   - Isi NPWP (optional)
   - Isi Start Date dan End Date
   - Isi Domicile Address
   - Upload dokumen (optional)
   - Klik "Simpan"
9. Klik "Finish" untuk submit

**Hasil yang Diharapkan:**
- ✅ Form dapat diisi dengan lengkap di semua step
- ✅ Validasi berjalan dengan baik (required fields)
- ✅ Persentase kepemilikan terhitung otomatis
- ✅ Submit berhasil
- ✅ Menampilkan pesan sukses: "Perusahaan berhasil dibuat"
- ✅ Redirect ke halaman list subsidiaries
- ✅ Subsidiary baru muncul di list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-002: Edit Existing Subsidiary
**Tujuan:** Memastikan user dapat mengedit data subsidiary yang sudah ada

**Kondisi Awal:**
- User sudah login dengan role yang memiliki permission edit
- Ada minimal 1 subsidiary yang sudah dibuat

**Langkah-langkah Test:**
1. Navigasi ke halaman Subsidiaries
2. Klik tombol "Edit" pada salah satu subsidiary (atau klik card → menu → Edit)
3. Ubah beberapa data:
   - Ubah Nama Singkat
   - Ubah Deskripsi
   - Ubah Status (jika ada)
4. Klik "Next" melalui semua step
5. Di Step 2, edit atau tambah pemegang saham
6. Di Step 4, edit atau tambah pengurus
7. Klik "Update" untuk submit

**Hasil yang Diharapkan:**
- ✅ Form ter-load dengan data existing
- ✅ User dapat mengubah data
- ✅ Submit berhasil
- ✅ Menampilkan pesan sukses: "Perusahaan berhasil diupdate"
- ✅ Redirect ke halaman detail subsidiary
- ✅ Data ter-update dengan benar

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-003: View Subsidiary Detail
**Tujuan:** Memastikan user dapat melihat detail lengkap subsidiary

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 subsidiary

**Langkah-langkah Test:**
1. Navigasi ke halaman Subsidiaries
2. Klik pada salah satu subsidiary card atau klik "Lihat Detail"
3. Lihat semua tab:
   - Tab "Profile" - Informasi dasar, pemegang saham, bidang usaha, pengurus
   - Tab "Input Laporan" - Financial reports
   - Tab "Dokumen" - Documents terkait subsidiary

**Hasil yang Diharapkan:**
- ✅ Halaman detail terbuka dengan benar
- ✅ Semua informasi ditampilkan dengan lengkap
- ✅ Tab navigation berfungsi
- ✅ Data financial reports (jika ada) ditampilkan
- ✅ Documents list ditampilkan (jika ada)

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-004: Delete Subsidiary
**Tujuan:** Memastikan user dapat menghapus subsidiary (soft delete)

**Kondisi Awal:**
- User sudah login dengan role: Superadmin, Administrator, atau Admin
- Ada minimal 1 subsidiary

**Langkah-langkah Test:**
1. Navigasi ke halaman Subsidiaries
2. Klik menu "..." pada salah satu subsidiary
3. Klik "Hapus"
4. Konfirmasi penghapusan (jika ada dialog konfirmasi)
5. Tunggu proses penghapusan

**Hasil yang Diharapkan:**
- ✅ Sistem meminta konfirmasi sebelum delete
- ✅ Setelah konfirmasi, subsidiary terhapus
- ✅ Menampilkan pesan sukses
- ✅ Subsidiary hilang dari list
- ✅ Subsidiary tidak muncul di list (soft delete)

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-005: Search & Filter Subsidiaries
**Tujuan:** Memastikan fitur search dan filter berfungsi dengan baik

**Kondisi Awal:**
- User sudah login
- Ada minimal 3 subsidiaries dengan nama berbeda

**Langkah-langkah Test:**
1. Navigasi ke halaman Subsidiaries
2. Di search box, ketik nama salah satu subsidiary
3. Lihat hasil filtered
4. Clear search (klik X atau hapus text)
5. Lihat semua subsidiaries muncul lagi
6. Test dengan view mode berbeda (Grid/List)

**Hasil yang Diharapkan:**
- ✅ Search box dapat digunakan
- ✅ Hasil filtered sesuai dengan keyword
- ✅ Clear search mengembalikan semua data
- ✅ Search bekerja di kedua view mode (Grid/List)
- ✅ Search case-insensitive

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-006: View My Company
**Tujuan:** Memastikan user dapat melihat detail perusahaan yang di-assign kepadanya

**Kondisi Awal:**
- User sudah login
- User sudah di-assign ke minimal 1 company

**Langkah-langkah Test:**
1. Klik menu "My Company" atau navigasi ke `/my-company`
2. Lihat detail perusahaan
3. Lihat financial data (jika ada)
4. Lihat documents (jika ada)

**Hasil yang Diharapkan:**
- ✅ Halaman My Company terbuka
- ✅ Menampilkan detail perusahaan yang di-assign
- ✅ Financial data ditampilkan dengan benar
- ✅ Documents list ditampilkan
- ✅ User hanya melihat company yang di-assign kepadanya

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-COMP-007: Comparison Feature (Period Comparison)
**Tujuan:** Memastikan fitur comparison periode berfungsi dengan baik

**Kondisi Awal:**
- User sudah login
- Ada subsidiary dengan minimal 2 periode financial report
- Fitur ENABLE_COMPARISON_FEATURE sudah diaktifkan

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Aktifkan "Compare Mode" (toggle atau checkbox)
4. Pilih Period Range 1: `2024-01` sampai `2024-03`
5. Pilih Period Range 2: `2024-04` sampai `2024-06`
6. Lihat perbandingan data di charts dan tables

**Hasil yang Diharapkan:**
- ✅ Compare mode dapat diaktifkan
- ✅ User dapat memilih 2 period range berbeda
- ✅ Charts menampilkan perbandingan (P1 vs P2)
- ✅ Tables menampilkan data perbandingan
- ✅ Label "(P1)" dan "(P2)" muncul di charts/tables

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.3 Document Management

#### TC-DOC-001: Upload Single Document
**Tujuan:** Memastikan user dapat upload dokumen tunggal dengan metadata

**Kondisi Awal:**
- User sudah login
- User memiliki permission untuk upload document
- File dokumen siap (PDF, DOCX, XLSX, dll)

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik tombol "Upload Document" atau navigasi ke upload page
3. Pilih file: `test-document.pdf`
4. Isi metadata:
   - Nama Dokumen: `Test Document UAT`
   - Kategori: Pilih dari dropdown
   - Folder: Pilih atau buat folder baru
   - Expiry Date (optional): Pilih tanggal
   - Description (optional): `Dokumen untuk testing UAT`
5. Klik "Upload" atau "Submit"

**Hasil yang Diharapkan:**
- ✅ File dapat dipilih
- ✅ Metadata form dapat diisi
- ✅ Upload berhasil
- ✅ Menampilkan pesan sukses
- ✅ Document muncul di list documents
- ✅ File dapat di-download

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-002: Upload Multiple Documents (Batch Upload)
**Tujuan:** Memastikan user dapat upload multiple files sekaligus

**Kondisi Awal:**
- User sudah login
- User memiliki permission untuk upload
- Multiple files siap (minimal 3 files)

**Langkah-langkah Test:**
1. Navigasi ke Document Upload
2. Klik "Pilih File" atau drag & drop area
3. Pilih multiple files (Ctrl/Cmd + Click):
   - `document1.pdf`
   - `document2.docx`
   - `image1.jpg`
4. Isi metadata umum (akan diterapkan ke semua files)
5. Klik "Upload"

**Hasil yang Diharapkan:**
- ✅ Multiple files dapat dipilih
- ✅ Progress bar menampilkan progress upload
- ✅ Semua files ter-upload dengan sukses
- ✅ Menampilkan pesan sukses dengan jumlah files
- ✅ Semua documents muncul di list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-003: View Document Detail & Preview
**Tujuan:** Memastikan user dapat melihat detail dan preview dokumen

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 document yang sudah di-upload

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik pada salah satu document
3. Lihat detail document (metadata, size, upload date, dll)
4. Klik "Preview" untuk file PDF atau gambar
5. Test preview untuk berbagai format:
   - PDF
   - Image (JPG, PNG)
   - Excel (XLSX, XLS)
   - Word (DOCX)

**Hasil yang Diharapkan:**
- ✅ Detail document ditampilkan dengan lengkap
- ✅ Preview PDF dapat dibuka di modal/fullscreen
- ✅ Preview gambar dapat dibuka
- ✅ Preview Excel menampilkan tabel HTML
- ✅ Preview Word menampilkan konten (jika didukung)
- ✅ Download button berfungsi

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-004: Edit Document Metadata
**Tujuan:** Memastikan user dapat mengedit metadata document

**Kondisi Awal:**
- User sudah login
- User memiliki permission edit
- Ada minimal 1 document

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik "Edit" pada salah satu document
3. Ubah metadata:
   - Ubah nama dokumen
   - Ubah kategori
   - Ubah folder
   - Ubah expiry date
4. Klik "Update" atau "Save"

**Hasil yang Diharapkan:**
- ✅ Form edit terbuka dengan data existing
- ✅ User dapat mengubah semua field
- ✅ Update berhasil
- ✅ Menampilkan pesan sukses
- ✅ Metadata ter-update di list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-005: Delete Document
**Tujuan:** Memastikan user dapat menghapus document

**Kondisi Awal:**
- User sudah login
- User memiliki permission delete
- Ada minimal 1 document

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik "Delete" pada salah satu document
3. Konfirmasi penghapusan
4. Tunggu proses delete

**Hasil yang Diharapkan:**
- ✅ Sistem meminta konfirmasi
- ✅ Setelah konfirmasi, document terhapus
- ✅ Menampilkan pesan sukses
- ✅ Document hilang dari list
- ✅ File terhapus dari storage

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-006: Organize Documents in Folders
**Tujuan:** Memastikan user dapat mengorganisir documents dalam folder structure

**Kondisi Awal:**
- User sudah login
- User memiliki permission untuk manage folders

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik "Create Folder" atau "New Folder"
3. Isi nama folder: `UAT Test Folder`
4. Pilih parent folder (jika membuat subfolder)
5. Klik "Create"
6. Pindahkan beberapa documents ke folder tersebut:
   - Select documents
   - Klik "Move to Folder"
   - Pilih folder tujuan
   - Klik "Move"

**Hasil yang Diharapkan:**
- ✅ Folder dapat dibuat
- ✅ Folder structure ditampilkan dengan benar
- ✅ Documents dapat dipindahkan ke folder
- ✅ Documents muncul di folder yang benar
- ✅ Navigation folder berfungsi

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-007: Search Documents
**Tujuan:** Memastikan fitur search documents berfungsi

**Kondisi Awal:**
- User sudah login
- Ada minimal 5 documents dengan nama berbeda

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Di search box, ketik nama document
3. Lihat hasil filtered
4. Clear search
5. Test search dengan berbagai keyword (nama, kategori, dll)

**Hasil yang Diharapkan:**
- ✅ Search box dapat digunakan
- ✅ Hasil filtered sesuai keyword
- ✅ Search bekerja real-time atau setelah enter
- ✅ Clear search mengembalikan semua documents
- ✅ Search case-insensitive

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-008: Download Document
**Tujuan:** Memastikan user dapat download document

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 document

**Langkah-langkah Test:**
1. Navigasi ke Document Management
2. Klik "Download" pada salah satu document
3. Tunggu file terdownload
4. Verifikasi file terdownload dengan benar

**Hasil yang Diharapkan:**
- ✅ Download button berfungsi
- ✅ File terdownload dengan nama yang benar
- ✅ File dapat dibuka dan konten sesuai
- ✅ Progress download ditampilkan (jika file besar)

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DOC-009: Upload File Exceeding Size Limit
**Tujuan:** Memastikan sistem menolak upload file yang melebihi batas ukuran

**Kondisi Awal:**
- User sudah login
- File gambar > 10MB siap

**Langkah-langkah Test:**
1. Navigasi ke Document Upload
2. Pilih file gambar > 10MB (misal: `large-image.jpg` 15MB)
3. Klik "Upload"

**Hasil yang Diharapkan:**
- ✅ Sistem mendeteksi file terlalu besar
- ✅ Menampilkan error message: "File gambar maksimal 10MB"
- ✅ Upload dibatalkan
- ✅ User dapat memilih file lain yang valid

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.4 Financial Reports Management

#### TC-REPORT-001: Create Financial Report (Realisasi)
**Tujuan:** Memastikan user dapat membuat laporan keuangan Realisasi

**Kondisi Awal:**
- User sudah login
- User memiliki permission untuk input laporan
- Ada minimal 1 subsidiary

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Pastikan mode "Realisasi" terpilih (bukan RKAP)
4. Pilih Periode: `2024-01` (Januari 2024)
5. Pilih Nama Penginput
6. Isi data **Neraca:**
   - Current Assets: `1000000000`
   - Non Current Assets: `2000000000`
   - Short Term Liabilities: `500000000`
   - Long Term Liabilities: `1000000000`
   - Equity: `1500000000`
7. Isi data **Laba Rugi:**
   - Revenue: `5000000000`
   - Operating Expenses: `3000000000`
   - Operating Profit: `2000000000`
   - Other Income: `100000000`
   - Tax: `500000000`
   - Net Profit: `1600000000`
8. Isi data **Cashflow:**
   - Operating Cashflow: `1800000000`
   - Investing Cashflow: `-500000000`
   - Financing Cashflow: `-300000000`
   - Ending Balance: `1000000000`
9. Isi data **Rasio Keuangan:**
   - ROE: `10.67`
   - ROI: `8.5`
   - Current Ratio: `2.0`
   - Cash Ratio: `1.5`
   - EBITDA: `2100000000`
   - EBITDA Margin: `42`
   - Net Profit Margin: `32`
   - Operating Profit Margin: `40`
   - Debt to Equity: `1.0`
10. Klik "Simpan" atau "Submit"

**Hasil yang Diharapkan:**
- ✅ Form dapat diisi dengan lengkap
- ✅ Semua field validasi berjalan
- ✅ Submit berhasil
- ✅ Menampilkan pesan sukses
- ✅ Data tersimpan dan muncul di list reports
- ✅ Charts ter-update dengan data baru

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-REPORT-002: Create Financial Report (RKAP)
**Tujuan:** Memastikan user dapat membuat laporan RKAP

**Kondisi Awal:**
- User sudah login
- User memiliki permission
- Ada minimal 1 subsidiary

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Pilih mode "RKAP"
4. Pilih Tahun: `2024`
5. Pilih Nama Penginput
6. Isi semua data financial (sama seperti Realisasi)
7. Klik "Simpan"

**Hasil yang Diharapkan:**
- ✅ Mode RKAP dapat dipilih
- ✅ Form menampilkan field untuk RKAP
- ✅ Submit berhasil
- ✅ Data RKAP tersimpan terpisah dari Realisasi
- ✅ RKAP muncul di list dengan label yang jelas

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-REPORT-003: Bulk Upload Financial Reports via Excel
**Tujuan:** Memastikan user dapat upload multiple reports via Excel template

**Kondisi Awal:**
- User sudah login
- User memiliki permission
- Excel template sudah di-download

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Klik "Bulk Upload" atau "Upload Excel"
4. Klik "Download Template" (jika ada)
5. Buka template Excel
6. Isi template dengan data multiple reports (minimal 3 reports)
7. Save template
8. Klik "Choose File" dan pilih template yang sudah diisi
9. Klik "Upload"
10. Tunggu proses validasi dan upload

**Hasil yang Diharapkan:**
- ✅ Template dapat di-download
- ✅ Template memiliki format yang benar
- ✅ Upload berhasil
- ✅ Validasi Excel berjalan (jika ada error, ditampilkan)
- ✅ Semua reports ter-upload
- ✅ Menampilkan summary: "X reports berhasil diupload"
- ✅ Data muncul di list reports

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-REPORT-004: Edit Financial Report
**Tujuan:** Memastikan user dapat mengedit report yang sudah ada

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 financial report

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Klik "Edit" pada salah satu report di list
4. Ubah beberapa nilai (misal: Revenue, Operating Expenses)
5. Klik "Update" atau "Save"

**Hasil yang Diharapkan:**
- ✅ Form edit terbuka dengan data existing
- ✅ User dapat mengubah data
- ✅ Update berhasil
- ✅ Menampilkan pesan sukses
- ✅ Data ter-update di list dan charts

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-REPORT-005: View Financial Charts & Analytics
**Tujuan:** Memastikan charts dan analytics menampilkan data dengan benar

**Kondisi Awal:**
- User sudah login
- Ada minimal 2 financial reports dengan periode berbeda

**Langkah-langkah Test:**
1. Buka halaman Subsidiary Detail
2. Tab "Input Laporan"
3. Lihat semua charts:
   - Balance Sheet Overview Chart
   - Profit Loss Overview Chart
   - Cashflow Overview Chart
   - Ratio Overview Chart
4. Ubah periode filter
5. Lihat charts ter-update
6. Test dengan compare mode (jika enabled)

**Hasil yang Diharapkan:**
- ✅ Semua charts ditampilkan
- ✅ Data di charts sesuai dengan data reports
- ✅ Charts responsive dan dapat di-zoom
- ✅ Filter periode mengupdate charts
- ✅ Compare mode menampilkan perbandingan (jika enabled)
- ✅ Tooltips/info box berfungsi saat hover

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.5 User Management (Admin/Superadmin Only)

#### TC-USER-001: Create New User
**Tujuan:** Memastikan admin dapat membuat user baru

**Kondisi Awal:**
- User login dengan role: Superadmin, Administrator, atau Admin
- User berada di halaman User Management

**Langkah-langkah Test:**
1. Navigasi ke User Management
2. Klik "Add User" atau "Tambah User"
3. Isi form:
   - Username: `newuser001`
   - Email: `newuser001@example.com`
   - Password: `Password123!`
   - Role: Pilih dari dropdown (Staff, Manager, Admin, dll)
4. Assign Company:
   - Pilih company dari dropdown
   - Pilih role untuk company tersebut
5. Klik "Create" atau "Simpan"

**Hasil yang Diharapkan:**
- ✅ Form dapat diisi dengan lengkap
- ✅ User berhasil dibuat
- ✅ Menampilkan pesan sukses
- ✅ User muncul di list
- ✅ User dapat login dengan credentials baru

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-USER-002: Edit User Information
**Tujuan:** Memastikan admin dapat mengedit informasi user

**Kondisi Awal:**
- User login dengan role admin
- Ada minimal 1 user (bukan superadmin)

**Langkah-langkah Test:**
1. Navigasi ke User Management
2. Klik "Edit" pada salah satu user
3. Ubah informasi:
   - Ubah username
   - Ubah email
   - Ubah role
4. Klik "Update"

**Hasil yang Diharapkan:**
- ✅ Form edit terbuka dengan data existing
- ✅ User dapat mengubah data
- ✅ Update berhasil
- ✅ Data ter-update di list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-USER-003: Assign Company to User
**Tujuan:** Memastikan admin dapat assign company ke user

**Kondisi Awal:**
- User login dengan role admin
- Ada user yang belum di-assign company
- Ada minimal 1 company

**Langkah-langkah Test:**
1. Navigasi ke User Management
2. Klik "Assign Company" pada user
3. Pilih company dari dropdown
4. Pilih role untuk company tersebut
5. Klik "Assign" atau "Simpan"

**Hasil yang Diharapkan:**
- ✅ Modal/form assign terbuka
- ✅ Company dapat dipilih
- ✅ Role dapat dipilih
- ✅ Assign berhasil
- ✅ Menampilkan pesan sukses
- ✅ User sekarang memiliki akses ke company tersebut

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-USER-004: Deactivate/Activate User
**Tujuan:** Memastikan admin dapat mengaktifkan/nonaktifkan user

**Kondisi Awal:**
- User login dengan role admin
- Fitur ENABLE_ACTIVATE_DEACTIVATE_FEATURE aktif
- Ada minimal 1 user

**Langkah-langkah Test:**
1. Navigasi ke User Management
2. Lihat status user (Aktif/Nonaktif)
3. Klik toggle atau menu untuk change status
4. Konfirmasi perubahan
5. Verifikasi status berubah

**Hasil yang Diharapkan:**
- ✅ Status dapat diubah
- ✅ User nonaktif tidak bisa login
- ✅ User aktif bisa login
- ✅ Status ter-update di list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-USER-005: Delete User
**Tujuan:** Memastikan admin dapat menghapus user

**Kondisi Awal:**
- User login dengan role admin
- Ada user yang dapat dihapus (bukan superadmin)

**Langkah-langkah Test:**
1. Navigasi ke User Management
2. Klik "Delete" pada user
3. Konfirmasi penghapusan
4. Tunggu proses delete

**Hasil yang Diharapkan:**
- ✅ Sistem meminta konfirmasi
- ✅ Setelah konfirmasi, user terhapus
- ✅ Menampilkan pesan sukses
- ✅ User hilang dari list

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.6 Profile & Settings

#### TC-PROF-001: Update Email
**Tujuan:** Memastikan user dapat mengupdate email

**Kondisi Awal:**
- User sudah login

**Langkah-langkah Test:**
1. Navigasi ke Profile
2. Klik "Update Email" atau tab "Email"
3. Input email baru: `newemail@example.com`
4. Klik "Update" atau "Save"
5. Verifikasi email (jika diperlukan)

**Hasil yang Diharapkan:**
- ✅ Form update email dapat diisi
- ✅ Validasi email format berjalan
- ✅ Update berhasil
- ✅ Menampilkan pesan sukses
- ✅ Email ter-update di profile

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-PROF-002: Change Password
**Tujuan:** Memastikan user dapat mengubah password

**Kondisi Awal:**
- User sudah login

**Langkah-langkah Test:**
1. Navigasi ke Profile
2. Klik "Change Password" atau tab "Password"
3. Isi form:
   - Old Password: `Password123!`
   - New Password: `NewPassword123!`
   - Confirm Password: `NewPassword123!`
4. Klik "Change Password" atau "Update"

**Hasil yang Diharapkan:**
- ✅ Form dapat diisi
- ✅ Validasi password match berjalan
- ✅ Validasi password strength (jika ada)
- ✅ Update berhasil
- ✅ Menampilkan pesan sukses
- ✅ User dapat login dengan password baru

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-PROF-003: Setup 2FA
**Tujuan:** Memastikan user dapat setup Two-Factor Authentication

**Kondisi Awal:**
- User sudah login
- User belum enable 2FA
- User memiliki authenticator app

**Langkah-langkah Test:**
1. Navigasi ke Settings
2. Tab "Security" atau "2FA"
3. Klik "Enable 2FA" atau "Setup 2FA"
4. Scan QR code dengan authenticator app
5. Input 6-digit verification code
6. Klik "Verify" atau "Enable"

**Hasil yang Diharapkan:**
- ✅ QR code ditampilkan
- ✅ QR code dapat di-scan
- ✅ Verification code dapat diinput
- ✅ Setup berhasil
- ✅ Menampilkan pesan sukses dengan backup codes
- ✅ 2FA sekarang aktif untuk user

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-PROF-004: Disable 2FA
**Tujuan:** Memastikan user dapat menonaktifkan 2FA

**Kondisi Awal:**
- User sudah login
- User sudah enable 2FA

**Langkah-langkah Test:**
1. Navigasi ke Settings
2. Tab "Security" atau "2FA"
3. Klik "Disable 2FA"
4. Konfirmasi (input password jika diperlukan)
5. Klik "Disable" atau "Confirm"

**Hasil yang Diharapkan:**
- ✅ Sistem meminta konfirmasi
- ✅ Setelah konfirmasi, 2FA dinonaktifkan
- ✅ Menampilkan pesan sukses
- ✅ User tidak perlu 2FA untuk login berikutnya

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.7 Dashboard & Analytics

#### TC-DASH-001: View Dashboard Metrics
**Tujuan:** Memastikan dashboard menampilkan metrics dengan benar

**Kondisi Awal:**
- User sudah login
- Ada data financial reports

**Langkah-langkah Test:**
1. Setelah login, user diarahkan ke dashboard (atau klik menu Dashboard)
2. Lihat semua KPI Cards:
   - Revenue
   - Opex
   - NPAT
   - Dividend
   - Financial Ratios
3. Lihat semua charts:
   - Revenue Chart
   - Financial Comparison Chart
4. Lihat Subsidiaries List

**Hasil yang Diharapkan:**
- ✅ Dashboard terbuka dengan benar
- ✅ Semua KPI cards menampilkan nilai
- ✅ Charts ditampilkan dengan data
- ✅ Metrics sesuai dengan data reports
- ✅ Loading state ditampilkan saat fetch data

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DASH-002: Filter Dashboard by Period
**Tujuan:** Memastikan filter periode di dashboard berfungsi

**Kondisi Awal:**
- User sudah login
- Ada data reports dengan periode berbeda

**Langkah-langkah Test:**
1. Buka Dashboard
2. Klik dropdown "Periode" atau "Period"
3. Pilih periode: `2024-01`
4. Lihat semua metrics dan charts ter-update
5. Pilih periode lain: `2024-02`
6. Verifikasi data berubah

**Hasil yang Diharapkan:**
- ✅ Dropdown periode dapat digunakan
- ✅ List periode menampilkan 12 bulan terakhir
- ✅ Setelah pilih periode, semua metrics ter-update
- ✅ Charts ter-update sesuai periode
- ✅ Data yang ditampilkan sesuai periode yang dipilih

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DASH-003: Export Dashboard Report (PDF)
**Tujuan:** Memastikan export PDF dashboard berfungsi

**Kondisi Awal:**
- User sudah login
- Dashboard menampilkan data

**Langkah-langkah Test:**
1. Buka Dashboard
2. Klik tombol "Export PDF" atau icon PDF
3. Tunggu proses generate PDF
4. Verifikasi file terdownload

**Hasil yang Diharapkan:**
- ✅ Export PDF berfungsi
- ✅ File PDF terdownload
- ✅ PDF berisi semua data dashboard (KPI, charts, dll)
- ✅ Format PDF rapi dan dapat dibaca

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-DASH-004: Export Dashboard Report (Excel)
**Tujuan:** Memastikan export Excel dashboard berfungsi

**Kondisi Awal:**
- User sudah login
- Dashboard menampilkan data

**Langkah-langkah Test:**
1. Buka Dashboard
2. Klik tombol "Export Excel" atau icon Excel
3. Tunggu proses generate Excel
4. Verifikasi file terdownload

**Hasil yang Diharapkan:**
- ✅ Export Excel berfungsi
- ✅ File Excel terdownload
- ✅ Excel berisi data yang dapat diolah
- ✅ Format Excel sesuai standar

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.8 Notifications

#### TC-NOTIF-001: View Notifications
**Tujuan:** Memastikan user dapat melihat notifications

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 notification (dapat dibuat manual atau otomatis)

**Langkah-langkah Test:**
1. Klik icon notifications di header (bell icon)
2. Lihat list notifications
3. Lihat unread count (badge number)
4. Scroll melalui list notifications

**Hasil yang Diharapkan:**
- ✅ Notification icon dapat diklik
- ✅ List notifications ditampilkan
- ✅ Unread count ditampilkan dengan benar
- ✅ Notifications memiliki informasi lengkap (title, message, type, date)
- ✅ Notifications dapat di-scroll

**Hasil Aktual:**
- [ ] Lolos
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-NOTIF-002: Mark Notification as Read
**Tujuan:** Memastikan user dapat mark notification sebagai read

**Kondisi Awal:**
- User sudah login
- Ada minimal 1 unread notification

**Langkah-langkah Test:**
1. Klik icon notifications
2. Klik pada salah satu unread notification
3. Lihat detail notification
4. Verifikasi notification marked as read

**Hasil yang Diharapkan:**
- ✅ Notification dapat diklik
- ✅ Detail notification ditampilkan
- ✅ Notification marked as read
- ✅ Unread count berkurang
- ✅ Notification tidak lagi ditandai sebagai unread

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-NOTIF-003: Clear All Notifications
**Tujuan:** Memastikan user dapat clear semua notifications

**Kondisi Awal:**
- User sudah login
- Ada minimal 3 notifications

**Langkah-langkah Test:**
1. Klik icon notifications
2. Klik "Clear All" atau "Hapus Semua"
3. Konfirmasi (jika ada)
4. Verifikasi semua notifications cleared

**Hasil yang Diharapkan:**
- ✅ Clear All button berfungsi
- ✅ Sistem meminta konfirmasi
- ✅ Setelah konfirmasi, semua notifications terhapus
- ✅ Unread count menjadi 0
- ✅ List notifications kosong

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.9 Role-Based Access Control (RBAC)

#### TC-RBAC-001: Superadmin - Full Access
**Tujuan:** Memastikan superadmin memiliki akses penuh ke semua fitur

**Kondisi Awal:**
- User login dengan role: Superadmin

**Langkah-langkah Test:**
1. Login sebagai superadmin
2. Verifikasi dapat akses:
   - User Management
   - Company Management (create, edit, delete)
   - Document Management
   - Financial Reports
   - Settings (termasuk Development Tools)
   - Audit Logs
3. Test semua CRUD operations

**Hasil yang Diharapkan:**
- ✅ Superadmin dapat akses semua menu
- ✅ Semua fitur dapat digunakan
- ✅ Development tools tersedia
- ✅ Tidak ada restriction

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-RBAC-002: Administrator - Management Access
**Tujuan:** Memastikan administrator memiliki akses management kecuali development tools

**Kondisi Awal:**
- User login dengan role: Administrator

**Langkah-langkah Test:**
1. Login sebagai administrator
2. Verifikasi dapat akses:
   - User Management ✅
   - Company Management ✅
   - Document Management ✅
   - Financial Reports ✅
   - Settings (tanpa Development Tools) ✅
3. Verifikasi TIDAK dapat akses:
   - Development Tools ❌

**Hasil yang Diharapkan:**
- ✅ Administrator dapat akses semua fitur management
- ✅ Development Tools tidak muncul atau disabled
- ✅ Semua CRUD operations berfungsi

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-RBAC-003: Admin - Company-Level Management
**Tujuan:** Memastikan admin hanya dapat manage company yang di-assign

**Kondisi Awal:**
- User login dengan role: Admin
- Admin di-assign ke specific companies

**Langkah-langkah Test:**
1. Login sebagai admin
2. Verifikasi dapat akses:
   - Company Management (hanya company yang di-assign) ✅
   - Document Management ✅
   - Financial Reports (hanya untuk assigned companies) ✅
3. Verifikasi TIDAK dapat akses:
   - User Management ❌
   - Global company management ❌

**Hasil yang Diharapkan:**
- ✅ Admin hanya melihat companies yang di-assign
- ✅ Admin tidak bisa manage users
- ✅ Admin tidak bisa akses companies lain

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-RBAC-004: Manager - View & Limited Edit
**Tujuan:** Memastikan manager dapat view dan edit terbatas

**Kondisi Awal:**
- User login dengan role: Manager

**Langkah-langkah Test:**
1. Login sebagai manager
2. Verifikasi dapat:
   - View companies ✅
   - View documents ✅
   - Create/Edit documents ✅
   - View financial reports ✅
3. Verifikasi TIDAK dapat:
   - Delete companies ❌
   - Manage users ❌
   - Create companies ❌

**Hasil yang Diharapkan:**
- ✅ Manager dapat view semua data
- ✅ Manager dapat edit documents
- ✅ Manager tidak bisa delete companies
- ✅ Manager tidak bisa manage users

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-RBAC-005: Staff - View Only
**Tujuan:** Memastikan staff hanya dapat view, tidak bisa edit/delete

**Kondisi Awal:**
- User login dengan role: Staff

**Langkah-langkah Test:**
1. Login sebagai staff
2. Verifikasi dapat:
   - View companies ✅
   - View documents ✅
   - View financial reports ✅
   - View dashboard ✅
3. Verifikasi TIDAK dapat:
   - Create/Edit companies ❌
   - Delete companies ❌
   - Manage users ❌
   - Edit documents ❌
   - Input financial reports ❌

**Hasil yang Diharapkan:**
- ✅ Staff hanya dapat view
- ✅ Tidak ada tombol create/edit/delete
- ✅ Tidak bisa akses management pages
- ✅ Read-only access

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

### 2.10 Error Handling & Edge Cases

#### TC-ERROR-001: Submit Form with Missing Required Fields
**Tujuan:** Memastikan validasi form berjalan dengan baik

**Kondisi Awal:**
- User sudah login

**Langkah-langkah Test:**
1. Buka form (misal: Create Subsidiary)
2. Langsung klik "Submit" tanpa isi required fields
3. Lihat error messages

**Hasil yang Diharapkan:**
- ✅ Form tidak bisa di-submit
- ✅ Error messages ditampilkan untuk setiap required field
- ✅ Error messages jelas dan informatif
- ✅ User tahu field mana yang harus diisi

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-ERROR-002: Network Error Handling
**Tujuan:** Memastikan aplikasi menangani network error dengan baik

**Kondisi Awal:**
- User sudah login
- Network dapat di-simulate (offline atau slow connection)

**Langkah-langkah Test:**
1. Disconnect network atau simulate slow connection
2. Lakukan action yang memerlukan API call (misal: submit form)
3. Lihat error handling

**Hasil yang Diharapkan:**
- ✅ Sistem mendeteksi network error
- ✅ Menampilkan error message yang user-friendly
- ✅ Tidak crash atau hang
- ✅ User dapat retry setelah network kembali

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

#### TC-ERROR-003: Access Restricted Page
**Tujuan:** Memastikan user dengan role terbatas tidak bisa akses restricted pages

**Kondisi Awal:**
- User login dengan role: Staff

**Langkah-langkah Test:**
1. Login sebagai staff
2. Coba akses URL langsung: `/user-management` atau `/settings` (development tools)
3. Lihat behavior aplikasi

**Hasil yang Diharapkan:**
- ✅ User di-redirect atau ditolak akses
- ✅ Menampilkan error message: "Access Denied" atau sejenisnya
- ✅ User tidak bisa melihat restricted content

**Hasil Aktual:**
- [ ] Lolos
- [ ] Gagal
- [ ] Terblokir

**Catatan:**
_________________________________________________________________

---

## 3. Ringkasan Test

### 3.1 Ringkasan Eksekusi Test

| Kategori | Total Test Cases | Lolos | Gagal | Terblokir | Belum Ditest |
|----------|------------------|--------|--------|---------|------------|
| Authentication & Authorization | 5 | | | | |
| Company/Subsidiary Management | 7 | | | | |
| Document Management | 9 | | | | |
| Financial Reports Management | 5 | | | | |
| User Management | 5 | | | | |
| Profile & Settings | 4 | | | | |
| Dashboard & Analytics | 4 | | | | |
| Notifications | 3 | | | | |
| Role-Based Access Control | 5 | | | | |
| Error Handling & Edge Cases | 3 | | | | |
| **TOTAL** | **49** | | | | |

### 3.2 Lingkungan Test

**Environment:** [ ] Development [ ] Staging [ ] Production  
**Browser:** [ ] Chrome [ ] Firefox [ ] Safari [ ] Edge  
**Versi Browser:** _______________  
**OS:** [ ] Windows [ ] macOS [ ] Linux  
**Tanggal:** _______________  
**Nama Tester:** _______________  

### 3.3 Hasil Keseluruhan

**Status Keseluruhan:** [ ] LOLOS [ ] GAGAL [ ] LOLOS BERSYARAT

**Catatan LOLOS BERSYARAT:**
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________

### 3.4 Masalah Kritis yang Ditemukan

| ID Issue | Deskripsi | Tingkat Keparahan | Status |
|----------|-------------|----------|--------|
| | | | |
| | | | |

### 3.5 Rekomendasi

_________________________________________________________________
_________________________________________________________________
_________________________________________________________________

---

## 4. Persetujuan

**Tester:**
- Nama: _______________
- Tanda Tangan: _______________
- Tanggal: _______________

**Project Manager:**
- Nama: _______________
- Tanda Tangan: _______________
- Tanggal: _______________

**Business Owner:**
- Nama: _______________
- Tanda Tangan: _______________
- Tanggal: _______________

---

**Riwayat Versi Dokumen:**
- v1.0 - Initial draft - 2024
