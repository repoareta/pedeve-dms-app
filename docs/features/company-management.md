# Manajemen Perusahaan - Daftar Perusahaan

Halaman ini menampilkan daftar semua perusahaan yang dapat Anda akses.

## Akses Halaman

- Klik menu **Anak Perusahaan** di navigasi atas
- Atau akses langsung melalui URL: `/subsidiaries`

## Tampilan Daftar

Halaman menampilkan tabel dengan kolom:

- **Nama Perusahaan**: Nama perusahaan
- **Kode**: Kode unik perusahaan
- **Level**: Level hierarki (Holding, Level 1, Level 2, Level 3)
- **Status**: Status aktif/nonaktif
- **Aksi**: Tombol untuk melihat detail atau edit

## Fitur

### Pencarian

- Gunakan search box di atas tabel untuk mencari perusahaan
- Pencarian berdasarkan nama atau kode perusahaan

### Filter

- Filter berdasarkan level hierarki
- Filter berdasarkan status (aktif/nonaktif)

### Sorting

- Klik header kolom untuk sorting
- Sorting berdasarkan nama, kode, atau level

### Pagination

- Gunakan pagination di bawah tabel untuk navigasi
- Pilih jumlah item per halaman

## Aksi

### Lihat Detail

1. Klik nama perusahaan di tabel
2. Atau klik tombol **Detail** di kolom Aksi
3. Anda akan diarahkan ke halaman detail perusahaan

### Edit Perusahaan

1. Klik tombol **Edit** di kolom Aksi
2. Anda akan diarahkan ke halaman edit
3. **Catatan**: Hanya admin/administrator yang dapat edit

### Tambah Perusahaan Baru

1. Klik tombol **Tambah Perusahaan** di atas tabel
2. Anda akan diarahkan ke halaman form tambah perusahaan
3. **Catatan**: Hanya admin/administrator yang dapat menambah

## Hierarki Perusahaan

Perusahaan diorganisir dalam struktur hierarki:

```
Holding (Level 0)
├── Level 1 Company
│   ├── Level 2 Company
│   │   └── Level 3 Company
│   └── Level 2 Company
└── Level 1 Company
```

- **Holding**: Perusahaan induk utama
- **Level 1**: Anak perusahaan langsung dari holding
- **Level 2**: Anak perusahaan dari level 1
- **Level 3**: Anak perusahaan dari level 2

## Tips

- Gunakan search untuk menemukan perusahaan dengan cepat
- Filter berdasarkan level untuk melihat struktur hierarki
- Klik nama perusahaan untuk melihat detail lengkap

## Langkah Selanjutnya

- [Tambah Perusahaan](./add-company) - Pelajari cara menambah perusahaan baru
- [Detail Perusahaan](./company-detail) - Pelajari detail perusahaan
- [Edit Perusahaan](./edit-company) - Pelajari cara edit perusahaan
