# Manajemen Pengguna - Edit Pengguna

Panduan untuk mengedit data pengguna yang sudah ada.

## Akses Halaman

- Dari halaman daftar pengguna, klik tombol **Edit** di kolom Aksi
- Atau dari halaman detail pengguna, klik tombol **Edit**
- Atau akses langsung: `/users/{id}/edit`
- **Catatan**: Hanya admin/administrator yang dapat akses

## Form Edit

Form edit sama dengan form tambah, tetapi sudah terisi dengan data yang ada.

### Field yang Bisa Diedit

- **Username**: Bisa diubah asalkan belum digunakan pengguna lain
- **Email**: Bisa diubah asalkan belum digunakan pengguna lain
- **Password**: Bisa diubah (kosongkan jika tidak ingin mengubah)
- **Role**: Bisa diubah
- **Nama Lengkap**: Bisa diubah
- **Telepon**: Bisa diubah
- **Status**: Bisa diubah (aktif/nonaktif)

### Assignment Perusahaan

**Menambah Assignment:**
1. Klik "Tambah Assignment"
2. Pilih perusahaan dan role
3. Simpan assignment

**Edit Assignment:**
1. Klik pada assignment yang ingin diedit
2. Ubah perusahaan atau role
3. Simpan perubahan

**Menghapus Assignment:**
1. Klik tombol hapus (X) di assignment
2. Konfirmasi penghapusan

## Validasi

Sama seperti form tambah:
- Username harus unik (jika diubah)
- Email harus unik (jika diubah)
- Password harus memenuhi kriteria (jika diubah)
- Company assignment harus valid

## Menyimpan Perubahan

1. **Review Perubahan**
   - Pastikan semua perubahan sudah benar
   - Cek assignment perusahaan

2. **Klik "Update"**
   - Perubahan akan disimpan
   - Jika berhasil, Anda akan diarahkan ke halaman detail pengguna

3. **Jika Ada Error**
   - Pesan error akan ditampilkan
   - Perbaiki field yang error
   - Coba simpan lagi

## Reset Password

Untuk reset password pengguna:

1. **Kosongkan Field Password**
   - Kosongkan field password di form edit
   - Password tidak akan diubah

2. **Atau Isi Password Baru**
   - Isi password baru di field password
   - Password akan diubah ke password baru

**Catatan**: 
- Password tidak ditampilkan di form (untuk keamanan)
- Kosongkan field jika tidak ingin mengubah password
- Isi password baru jika ingin mengubah

## Tips

- Pastikan username dan email masih valid setelah edit
- Update assignment jika pengguna pindah perusahaan
- Nonaktifkan pengguna yang tidak lagi aktif (jangan hapus)
- Backup data penting sebelum melakukan perubahan besar

## Troubleshooting

### Tidak Bisa Edit Username

- Username yang sudah digunakan pengguna lain tidak bisa diubah
- Jika perlu mengubah, hubungi administrator

### Tidak Bisa Edit Email

- Email yang sudah digunakan pengguna lain tidak bisa diubah
- Jika perlu mengubah, hubungi administrator

### Password Tidak Berubah

- Pastikan field password sudah diisi dengan password baru
- Kosongkan field jika tidak ingin mengubah password
- Pastikan password memenuhi kriteria

## Referensi

- [Daftar Pengguna](./user-management) - Kembali ke daftar pengguna
