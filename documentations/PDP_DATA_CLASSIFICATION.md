# Klasifikasi Data Pribadi Berdasarkan UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi

Dokumen ini mengidentifikasi data apa saja yang perlu dilindungi berdasarkan ketentuan UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi yang berlaku di Indonesia.

Referensi: [UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi](https://peraturan.bpk.go.id/Details/229798/uu-no-27-tahun-2022)

## Kategori Data Pribadi Menurut UU PDP

Menurut UU No. 27 Tahun 2022, data pribadi dibagi menjadi 2 kategori:

### 1. Data Pribadi yang Bersifat Spesifik
Data yang memerlukan perlindungan **EKSTRA** karena sifatnya yang sangat sensitif:
- Data dan informasi kesehatan
- Data biometrik
- Data genetika
- Catatan kejahatan
- Data anak
- **Data keuangan pribadi**
- Dan/atau data lainnya sesuai dengan ketentuan peraturan perundang-undangan

### 2. Data Pribadi yang Bersifat Umum
Data yang dapat mengidentifikasi seseorang:
- **Nama lengkap**
- Jenis kelamin
- Kewarganegaraan
- Agama
- Status perkawinan
- Dan/atau Data Pribadi yang dikombinasikan untuk mengidentifikasi seseorang

---

## Klasifikasi Data di Aplikasi Pedeve DMS

### DATA PRIBADI SPESIFIK (Perlindungan Ekstra)

#### 1. Data Keuangan Pribadi

**Tabel: `directors` (DirectorModel)**
- `NPWP` (Nomor Pokok Wajib Pajak) - **SPESIFIK: Data Keuangan Pribadi**
- `KTP` (Nomor KTP) - **SPESIFIK: Identitas Resmi** (meskipun tidak disebutkan eksplisit, KTP termasuk identitas resmi yang sangat sensitif)

**Tabel: `shareholders` (ShareholderModel)**
- `IdentityNumber` (KTP/NPWP untuk individu) - **SPESIFIK: Data Keuangan Pribadi & Identitas Resmi**

**Tabel: `companies` (CompanyModel)**
- `NPWP` (Nomor Pokok Wajib Pajak perusahaan) - Meskipun data perusahaan, NPWP perusahaan tetap sensitif karena dapat digunakan untuk identifikasi

**Catatan**: Data keuangan pribadi memerlukan perlindungan ekstra karena dapat digunakan untuk tujuan yang tidak sah (penipuan, identitas palsu, dll).

#### 2. Data Identitas Resmi

**Tabel: `directors` (DirectorModel)**
- `KTP` - **SPESIFIK: Identitas Resmi** (dapat digunakan untuk identifikasi unik seseorang)
- `NPWP` - **SPESIFIK: Data Keuangan Pribadi**

**Tabel: `shareholders` (ShareholderModel)**
- `IdentityNumber` (KTP/NPWP) - **SPESIFIK: Data Keuangan Pribadi & Identitas Resmi**

---

### DATA PRIBADI UMUM (Perlindungan Standar)

#### 1. Data yang Dapat Mengidentifikasi Individu

**Tabel: `users` (UserModel)**
- `Username` - **UMUM: Identitas Pengguna** (dapat mengidentifikasi seseorang)
- `Email` - **UMUM: Identitas Kontak** (dapat mengidentifikasi seseorang, juga data kontak pribadi)
- `Password` - **SENSITIF: Kredensial** (harus di-hash, tidak boleh disimpan dalam bentuk plaintext)
- `UserID` - **UMUM: Identitas Pengguna**

**Tabel: `directors` (DirectorModel)**
- `FullName` - **UMUM: Nama Lengkap** (seperti yang disebutkan eksplisit dalam UU PDP)
- `DomicileAddress` - **UMUM: Alamat Domisili** (dapat mengidentifikasi seseorang jika dikombinasikan dengan data lain)
- `StartDate`, `EndDate` - **UMUM: Data Temporal** (dapat mengidentifikasi jika dikombinasikan dengan data lain)

**Tabel: `shareholders` (ShareholderModel)**
- `Name` - **UMUM: Nama** (untuk individu, dapat mengidentifikasi seseorang)

**Tabel: `audit_logs` (AuditLog)**
- `UserID` - **UMUM: Identitas Pengguna**
- `Username` - **UMUM: Identitas Pengguna**
- `IPAddress` - **UMUM: Data Lokasi/Tracking** (dapat mengidentifikasi lokasi dan perangkat pengguna)
- `UserAgent` - **UMUM: Data Teknis** (dapat digunakan untuk identifikasi perangkat)

**Tabel: `user_activity_logs` (UserActivityLog)**
- `UserID` - **UMUM: Identitas Pengguna**
- `Username` - **UMUM: Identitas Pengguna**
- `IPAddress` - **UMUM: Data Lokasi/Tracking**
- `UserAgent` - **UMUM: Data Teknis**

**Tabel: `notifications` (NotificationModel)**
- `UserID` - **UMUM: Identitas Pengguna**

**Tabel: `two_factor_auths` (TwoFactorAuth)**
- `UserID` - **UMUM: Identitas Pengguna**
- `Secret` - **SENSITIF: Kredensial** (harus di-enkripsi)
- `BackupCodes` - **SENSITIF: Kredensial** (harus di-enkripsi)

**Tabel: `documents` (DocumentModel)**
- `UploaderID` - **UMUM: Identitas Pengguna**
- `DirectorID` - **UMUM: Identitas Pengurus** (jika dokumen terkait dengan director, dapat mengidentifikasi individu)
- `FileName`, `FilePath` - Dapat mengandung informasi sensitif jika nama file mengidentifikasi individu

**Tabel: `document_folders` (DocumentFolderModel)**
- `CreatedBy` - **UMUM: Identitas Pengguna**

---

### DATA YANG DAPAT DIKOMBINASIKAN UNTUK IDENTIFIKASI

Beberapa data meskipun secara individual tidak langsung mengidentifikasi seseorang, tetapi jika dikombinasikan dapat mengidentifikasi seseorang:

#### Kombinasi Data Director:
- `FullName` + `KTP` + `NPWP` + `DomicileAddress` → **Sangat kuat mengidentifikasi seseorang**
- `FullName` + `CompanyID` + `Position` → Dapat mengidentifikasi seseorang dalam konteks perusahaan

#### Kombinasi Data Shareholder:
- `Name` + `IdentityNumber` → Dapat mengidentifikasi seseorang
- `Name` + `OwnershipPercent` + `CompanyID` → Dapat mengidentifikasi kepemilikan individu

#### Kombinasi Data User:
- `Username` + `Email` → Dapat mengidentifikasi seseorang
- `UserID` + `IPAddress` + timestamp → Dapat melacak aktivitas seseorang

---

## Ringkasan Klasifikasi

### Data Pribadi Spesifik (Perlindungan Ekstra) - **PRIORITAS TINGGI**

1. **`directors.KTP`** - Nomor KTP
2. **`directors.NPWP`** - Nomor NPWP (Data Keuangan Pribadi)
3. **`shareholders.IdentityNumber`** - KTP/NPWP untuk individu (Data Keuangan Pribadi & Identitas Resmi)
4. **`companies.NPWP`** - NPWP perusahaan (meskipun data perusahaan, tetap sensitif)

### Data Pribadi Umum (Perlindungan Standar) - **PRIORITAS MENENGAH**

1. **`users.Email`** - Email pengguna
2. **`users.Username`** - Username pengguna
3. **`directors.FullName`** - Nama lengkap pengurus
4. **`directors.DomicileAddress`** - Alamat domisili
5. **`shareholders.Name`** - Nama pemegang saham (untuk individu)
6. **`audit_logs.IPAddress`** - Alamat IP (data lokasi/tracking)
7. **`user_activity_logs.IPAddress`** - Alamat IP (data lokasi/tracking)
8. **`audit_logs.UserAgent`** - User agent (data teknis)
9. **`user_activity_logs.UserAgent`** - User agent (data teknis)

### Data Sensitif (Kredensial) - **PRIORITAS SANGAT TINGGI**

1. **`users.Password`** - Password (harus di-hash dengan bcrypt)
2. **`two_factor_auths.Secret`** - Secret TOTP (harus di-enkripsi)
3. **`two_factor_auths.BackupCodes`** - Backup codes 2FA (harus di-enkripsi)

---

## Rekomendasi Perlindungan Data

### 1. Data Pribadi Spesifik (KTP, NPWP, IdentityNumber)
- **Enkripsi di Database**: Gunakan enkripsi field-level untuk data sensitif
- **Access Control**: Hanya role tertentu yang dapat mengakses (misalnya: superadmin, admin)
- **Audit Logging**: Log semua akses, read, update, delete terhadap data ini
- **Masking**: Di UI, tampilkan data ter-mask (contoh: "12******45") kecuali untuk user yang authorized
- **Data Retention Policy**: Tentukan kebijakan penyimpanan data sesuai kebutuhan hukum
- **Data Minimization**: Hanya simpan data yang benar-benar diperlukan

### 2. Data Pribadi Umum (Email, Nama, Alamat)
- **Access Control**: Implementasi RBAC untuk membatasi akses
- **Audit Logging**: Log akses dan perubahan data
- **Anonymization**: Pertimbangkan anonymization untuk data yang sudah tidak diperlukan
- **Data Retention Policy**: Tentukan kebijakan penyimpanan data

### 3. Data Kredensial (Password, Secret, BackupCodes)
- **Password Hashing**: Sudah diimplementasikan dengan bcrypt
- **Secret Encryption**: Secret TOTP dan backup codes harus di-enkripsi (bukan hanya hashing)
- **Never Log**: Jangan pernah log kredensial dalam bentuk apapun
- **Secure Storage**: Pastikan database dan backup ter-enkripsi

### 4. Data Tracking (IPAddress, UserAgent)
- **Anonymization**: Anonymize IP address setelah periode tertentu (misalnya: hapus 3 digit terakhir)
- **Retention Policy**: Sesuai dengan audit log retention policy (90 hari untuk user actions, 30 hari untuk technical errors)
- **Access Control**: Hanya role tertentu yang dapat melihat IP address lengkap

---

## Compliance Checklist

- [ ] Identifikasi semua data pribadi di database (✅ Sudah dilakukan di dokumen ini)
- [ ] Implementasi enkripsi untuk data pribadi spesifik (KTP, NPWP)
- [ ] Implementasi access control yang ketat untuk data sensitif
- [ ] Implementasi audit logging untuk semua akses data pribadi
- [ ] Implementasi data masking di UI untuk data sensitif
- [ ] Tentukan dan implementasi data retention policy
- [ ] Implementasi anonymization untuk data lama
- [ ] Dokumentasi prosedur penanganan data pribadi
- [ ] Training untuk tim tentang penanganan data pribadi
- [ ] Regular audit terhadap akses data pribadi
- [ ] Implementasi mekanisme untuk hak subjek data pribadi (akses, perbaikan, penghapusan, portabilitas)
- [ ] Implementasi consent management (jika diperlukan)
- [ ] Dokumentasi legal basis untuk pemrosesan data pribadi
- [ ] Implementasi data breach notification procedure

---

## Catatan Penting

1. **UU No. 27 Tahun 2022 sudah berlaku sejak 17 Oktober 2022**, sehingga semua aplikasi yang memproses data pribadi wajib mematuhi ketentuan ini.

2. **Sanksi**: UU PDP mengatur sanksi administratif dan pidana bagi pelanggaran pelindungan data pribadi.

3. **Data Anak**: Jika aplikasi ini akan memproses data pribadi anak di bawah umur, diperlukan perlindungan ekstra dan persetujuan dari orang tua/wali.

4. **Transfer Data**: Transfer data pribadi ke luar negeri harus mematuhi ketentuan yang diatur dalam UU PDP.

5. **Data Breach**: Dalam hal terjadi kebocoran data pribadi, pengendali data pribadi wajib memberitahukan kepada subjek data pribadi dan/atau otoritas yang berwenang sesuai dengan ketentuan peraturan perundang-undangan.

---

## Referensi

- [UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi](https://peraturan.bpk.go.id/Details/229798/uu-no-27-tahun-2022)
- Peraturan pelaksana UU PDP (jika ada)
- Panduan dari Otoritas Pelindungan Data Pribadi (jika sudah ditetapkan)

---

**Dokumen ini dibuat untuk keperluan internal dan harus direview secara berkala untuk memastikan compliance dengan regulasi terbaru.**

