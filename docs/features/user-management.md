# Manajemen Pengguna - Daftar Pengguna

Halaman ini menampilkan semua pengguna yang dapat Anda akses (sesuai dengan role Anda).

## Akses Halaman

- Klik menu **Pengguna** di navigasi atas
- Atau akses langsung: `/users`
- **Catatan**: Hanya admin/administrator yang dapat akses

## Tampilan

Halaman menampilkan tabel dengan kolom:

- **Username**: Username pengguna
- **Email**: Email pengguna
- **Role**: Peran pengguna (Administrator, Admin, Manager, Staff)
- **Perusahaan**: Perusahaan yang di-assign
- **Status**: Status aktif/nonaktif
- **Aksi**: Tombol untuk melihat detail atau edit

## Fitur

### Pencarian

- Gunakan search box untuk mencari pengguna
- Pencarian berdasarkan username atau email

### Filter

- **Filter Role**: Filter berdasarkan peran
- **Filter Status**: Filter berdasarkan status (aktif/nonaktif)
- **Filter Perusahaan**: Filter berdasarkan perusahaan

### Sorting

- Klik header kolom untuk sorting
- Sorting berdasarkan username, email, role, dll

### Pagination

- Gunakan pagination untuk navigasi
- Pilih jumlah item per halaman

## Aksi

### Lihat Detail

1. Klik username di tabel
2. Atau klik tombol **Detail** di kolom Aksi
3. Anda akan diarahkan ke halaman detail pengguna

### Edit Pengguna

1. Klik tombol **Edit** di kolom Aksi
2. Anda akan diarahkan ke halaman edit
3. **Catatan**: Lihat panduan [Edit Pengguna](./edit-user)

### Tambah Pengguna Baru

1. Klik tombol **Tambah Pengguna** di atas tabel
2. Anda akan diarahkan ke halaman form tambah pengguna
3. **Catatan**: Lihat panduan [Tambah Pengguna](./add-user)

## Peran Pengguna

| Peran | Deskripsi | Akses |
|-------|-----------|-------|
| **Administrator** | Administrator | Akses penuh ke semua fitur |
| **Admin** | Admin perusahaan | Akses terbatas pada perusahaan yang di-assign |
| **Manager** | Manager | Akses untuk melihat dan mengelola data perusahaan |
| **Staff** | Staff | Akses terbatas untuk input data |

## Tips

- Gunakan filter untuk melihat pengguna spesifik
- Gunakan search untuk menemukan pengguna dengan cepat
- Cek status pengguna sebelum assign ke perusahaan

## Langkah Selanjutnya

- [Tambah Pengguna](./add-user) - Pelajari cara menambah pengguna baru
- [Edit Pengguna](./edit-user) - Pelajari cara edit pengguna
