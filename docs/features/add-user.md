# Manajemen Pengguna - Tambah Pengguna

Panduan untuk menambahkan pengguna baru ke dalam sistem.

## Akses Halaman

- Dari halaman daftar pengguna, klik tombol **Tambah Pengguna**
- Atau akses langsung: `/users/new`
- **Catatan**: Hanya admin/administrator yang dapat akses

## Form Tambah Pengguna

### Informasi Dasar

**Field Wajib:**
- **Username** *: Username untuk login (harus unik)
- **Email** *: Email pengguna (harus unik)
- **Password** *: Password untuk login
- **Role** *: Pilih peran pengguna

**Field Opsional:**
- **Nama Lengkap**: Nama lengkap pengguna
- **Telepon**: Nomor telepon
- **Status**: Aktif atau Nonaktif (default: Aktif)

### Assignment Perusahaan

**Menambah Assignment:**

1. **Klik "Tambah Assignment"**
2. **Pilih Perusahaan**
   - Pilih perusahaan dari dropdown
   - Bisa pilih beberapa perusahaan

3. **Pilih Role di Perusahaan**
   - Setiap perusahaan bisa memiliki role berbeda
   - Contoh: Admin di Perusahaan A, Manager di Perusahaan B

4. **Simpan Assignment**
   - Assignment akan ditambahkan ke list
   - Bisa menambah multiple assignments

**Menghapus Assignment:**
- Klik tombol hapus (X) di assignment yang ingin dihapus

### Standby User

Pengguna bisa dibuat tanpa assignment perusahaan (standby):
- User akan dibuat tanpa assignment
- Assignment bisa ditambahkan nanti
- Berguna untuk user yang belum ditentukan perusahaannya

## Validasi

Sistem akan memvalidasi:
- Username harus unik
- Email harus unik dan format valid
- Password harus memenuhi kriteria (minimal 8 karakter)
- Role harus valid
- Company assignment harus valid (jika diisi)

## Menyimpan Pengguna

1. **Isi Form**
   - Isi semua field wajib
   - Tambah assignment perusahaan jika diperlukan

2. **Review**
   - Pastikan username dan email benar
   - Cek assignment perusahaan

3. **Klik "Simpan"**
   - Sistem akan menyimpan pengguna
   - Jika berhasil, Anda akan diarahkan ke halaman daftar pengguna

4. **Jika Ada Error**
   - Sistem akan menampilkan pesan error
   - Perbaiki field yang error
   - Coba simpan lagi

## Tips

- Gunakan username yang mudah diingat
- Pastikan email valid dan aktif
- Buat password yang kuat (minimal 8 karakter, kombinasi huruf dan angka)
- Assign ke perusahaan yang sesuai dengan role
- Buat standby user jika perusahaan belum ditentukan

## Troubleshooting

### Username Sudah Ada

- Username harus unik
- Gunakan username yang berbeda
- Cek daftar pengguna untuk melihat username yang sudah digunakan

### Email Sudah Ada

- Email harus unik
- Gunakan email yang berbeda
- Cek apakah email sudah digunakan pengguna lain

### Password Tidak Valid

- Password minimal 8 karakter
- Disarankan kombinasi huruf besar, huruf kecil, dan angka
- Pastikan password tidak terlalu mudah ditebak

### Perusahaan Tidak Muncul di Dropdown

- Pastikan perusahaan sudah ada di sistem
- Pastikan perusahaan status aktif
- Hubungi administrator jika perusahaan seharusnya muncul

## Langkah Selanjutnya

- [Edit Pengguna](./edit-user) - Pelajari cara edit pengguna
- [Daftar Pengguna](./user-management) - Kembali ke daftar pengguna
