# Login & Autentikasi

Panduan lengkap untuk login dan setup Two-Factor Authentication (2FA).

## Login ke Aplikasi

### Langkah Login

1. **Buka Halaman Login**
   - Akses URL aplikasi
   - Anda akan diarahkan ke halaman login jika belum terautentikasi

2. **Masukkan Kredensial**
   - **Username**: Masukkan username Anda
   - **Password**: Masukkan password Anda

3. **Klik Tombol Login**
   - Sistem akan memverifikasi kredensial Anda
   - Jika berhasil, Anda akan diarahkan ke dashboard

### Jika Lupa Password

Jika Anda lupa password:

1. Hubungi administrator untuk reset password
2. Administrator akan membuat password baru untuk Anda
3. Setelah login, disarankan untuk mengubah password di halaman Settings

## Two-Factor Authentication (2FA)

2FA memberikan keamanan tambahan untuk akun Anda dengan memerlukan kode verifikasi selain password.

### Setup 2FA

1. **Buka Halaman Settings**
   - Klik menu **Pengaturan** di navigasi atas
   - Pilih tab **Security** atau **Keamanan**

2. **Generate QR Code**
   - Klik tombol **Enable 2FA** atau **Aktifkan 2FA**
   - Sistem akan menampilkan QR code

3. **Scan QR Code**
   - Buka aplikasi authenticator di smartphone Anda (Google Authenticator, Authy, dll)
   - Scan QR code yang ditampilkan
   - Aplikasi akan menambahkan akun Pedeve DMS

4. **Verifikasi Kode**
   - Masukkan kode 6 digit dari aplikasi authenticator
   - Klik **Verify** atau **Verifikasi**
   - Jika berhasil, 2FA akan aktif

5. **Simpan Backup Codes**
   - Sistem akan menampilkan backup codes
   - **PENTING**: Simpan backup codes di tempat yang aman
   - Backup codes dapat digunakan jika Anda kehilangan akses ke authenticator

### Login dengan 2FA

Setelah 2FA aktif, proses login akan sedikit berbeda:

1. Masukkan username dan password seperti biasa
2. Setelah password benar, sistem akan meminta kode 2FA
3. Buka aplikasi authenticator di smartphone
4. Masukkan kode 6 digit yang ditampilkan
5. Klik **Verify** untuk melanjutkan

### Disable 2FA

Jika Anda ingin menonaktifkan 2FA:

1. Buka halaman **Settings** → **Security**
2. Klik tombol **Disable 2FA**
3. Masukkan password untuk konfirmasi
4. 2FA akan dinonaktifkan

**Catatan**: Disarankan untuk tetap menggunakan 2FA untuk keamanan akun.

## Logout

Untuk logout dari aplikasi:

1. Klik nama pengguna di pojok kanan atas
2. Pilih **Logout** dari dropdown menu
3. Anda akan diarahkan ke halaman login

## Tips Keamanan

- Jangan bagikan kredensial login Anda
- Gunakan password yang kuat dan unik
- Aktifkan 2FA untuk keamanan tambahan
- Simpan backup codes di tempat yang aman
- Logout jika menggunakan komputer bersama

## Troubleshooting

### Tidak Bisa Login

- Pastikan username dan password benar (case-sensitive)
- Cek apakah Caps Lock aktif
- Clear cache browser dan coba lagi
- Hubungi administrator jika masalah berlanjut

### Kode 2FA Tidak Valid

- Pastikan waktu di smartphone Anda sudah sinkron
- Pastikan Anda menggunakan kode terbaru (kode berubah setiap 30 detik)
- Coba gunakan backup code jika tersedia
- Hubungi administrator jika masalah berlanjut

### Lupa Backup Codes

- Hubungi administrator untuk reset 2FA
- Setelah reset, Anda perlu setup 2FA ulang
