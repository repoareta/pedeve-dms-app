# Two-Factor Authentication (2FA)

Panduan lengkap untuk setup dan menggunakan Two-Factor Authentication.

## Apa itu 2FA?

2FA adalah sistem keamanan tambahan yang memerlukan dua faktor untuk login:
1. **Password** (sesuatu yang Anda tahu)
2. **Kode dari Authenticator App** (sesuatu yang Anda miliki)

## Setup 2FA

### Langkah 1: Buka Halaman Settings

1. Klik menu **Pengaturan** di navigasi atas
2. Pilih tab **Security** atau **Keamanan**

### Langkah 2: Generate QR Code

1. **Klik "Enable 2FA" atau "Aktifkan 2FA"**
   - Tombol berada di bagian 2FA
   - QR code akan di-generate

2. **QR Code Akan Muncul**
   - QR code ditampilkan di layar
   - Siapkan aplikasi authenticator di smartphone

### Langkah 3: Install Authenticator App

Jika belum punya, install salah satu aplikasi berikut:
- **Google Authenticator** (iOS/Android)
- **Authy** (iOS/Android)
- **Microsoft Authenticator** (iOS/Android)

### Langkah 4: Scan QR Code

1. **Buka Aplikasi Authenticator**
   - Buka aplikasi di smartphone
   - Pilih opsi "Add Account" atau "Tambah Akun"

2. **Scan QR Code**
   - Pilih "Scan QR Code"
   - Arahkan kamera ke QR code di layar
   - QR code akan ter-scan otomatis

3. **Akun Ditambahkan**
   - Aplikasi akan menambahkan akun "Pedeve DMS"
   - Kode 6 digit akan muncul dan berubah setiap 30 detik

### Langkah 5: Verifikasi Kode

1. **Masukkan Kode**
   - Masukkan kode 6 digit dari aplikasi authenticator
   - Kode berubah setiap 30 detik, pastikan menggunakan kode terbaru

2. **Klik "Verify" atau "Verifikasi"**
   - Kode akan diverifikasi
   - Jika benar, 2FA akan aktif

### Langkah 6: Simpan Backup Codes

1. **Backup Codes Ditampilkan**
   - Setelah verifikasi berhasil, sistem menampilkan backup codes
   - **PENTING**: Simpan backup codes di tempat yang aman

2. **Simpan Backup Codes**
   - Copy atau screenshot backup codes
   - Simpan di tempat yang aman (password manager, file terenkripsi)
   - Backup codes dapat digunakan jika kehilangan akses ke authenticator

## Login dengan 2FA

Setelah 2FA aktif, proses login sedikit berbeda:

1. **Masukkan Username dan Password**
   - Login seperti biasa dengan username dan password

2. **Masukkan Kode 2FA**
   - Setelah password benar, sistem meminta kode 2FA
   - Buka aplikasi authenticator di smartphone
   - Masukkan kode 6 digit yang ditampilkan

3. **Klik "Verify"**
   - Kode akan diverifikasi
   - Jika benar, Anda akan login

## Menggunakan Backup Codes

Jika kehilangan akses ke authenticator:

1. **Login dengan Username dan Password**
   - Login seperti biasa

2. **Klik "Use Backup Code"**
   - Di halaman verifikasi 2FA
   - Pilih opsi menggunakan backup code

3. **Masukkan Backup Code**
   - Masukkan salah satu backup code yang disimpan
   - Setiap backup code hanya bisa digunakan sekali

4. **Login Berhasil**
   - Setelah backup code valid, Anda akan login
   - **PENTING**: Setup 2FA ulang setelah login untuk keamanan

## Disable 2FA

Jika ingin menonaktifkan 2FA:

1. **Buka Halaman Settings → Security**
2. **Klik "Disable 2FA"**
3. **Masukkan Password**
   - Konfirmasi dengan memasukkan password
4. **Konfirmasi**
   - 2FA akan dinonaktifkan
   - **Catatan**: Tidak disarankan untuk menonaktifkan 2FA

## Tips Keamanan

- **Simpan Backup Codes**: Simpan backup codes di tempat yang aman
- **Jangan Bagikan Kode**: Jangan bagikan kode 2FA kepada siapa pun
- **Gunakan Password Manager**: Simpan backup codes di password manager
- **Setup di Multiple Devices**: Beberapa authenticator app support sync (seperti Authy)
- **Jangan Screenshot QR Code**: QR code mengandung secret, jangan screenshot dan bagikan

## Troubleshooting

### Kode 2FA Tidak Valid

**Penyebab:**
- Waktu di smartphone tidak sinkron
- Menggunakan kode yang sudah expired (lebih dari 30 detik)

**Solusi:**
- Pastikan waktu di smartphone sudah sinkron (auto-sync time)
- Gunakan kode terbaru (kode berubah setiap 30 detik)
- Coba kode berikutnya jika kode saat ini tidak valid

### Kehilangan Authenticator App

**Jika Masih Punya Backup Codes:**
1. Login menggunakan backup code
2. Setup 2FA ulang dengan QR code baru
3. Simpan backup codes baru

**Jika Tidak Punya Backup Codes:**
1. Hubungi administrator
2. Administrator akan reset 2FA untuk Anda
3. Setup 2FA ulang setelah reset

### QR Code Tidak Bisa Di-scan

**Solusi:**
- Pastikan kamera smartphone bersih
- Pastikan QR code jelas dan tidak terpotong
- Coba zoom in/out untuk fokus yang lebih baik
- Atau gunakan manual entry (masukkan secret key secara manual)

### Authenticator App Terhapus

**Jika Menggunakan Authy (dengan sync):**
- Install ulang Authy
- Login dengan akun Authy
- Data akan ter-sync otomatis

**Jika Menggunakan Google Authenticator:**
- Gunakan backup codes untuk login
- Setup 2FA ulang setelah login

## Langkah Selanjutnya

- [Profil](./profile) - Edit profil pengguna
- [Settings](./settings) - Pengaturan lainnya
