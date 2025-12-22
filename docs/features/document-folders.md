# Manajemen Dokumen - Folder & Kategori

Panduan untuk mengorganisir dokumen menggunakan folder.

## Konsep Folder

Folder digunakan untuk mengorganisir dokumen dalam kategori yang logis:
- Setiap dokumen harus berada di dalam folder
- Folder dapat memiliki subfolder (hierarki)
- Folder membantu pencarian dan navigasi

## Membuat Folder

### Folder Baru

1. **Dari Halaman Dokumen**
   - Klik tombol **Add Folder** atau **Add new folder"
   - Masukkan nama folder
   - Klik **Buat**

2. **Dari Halaman Folder Detail**
   - Klik tombol **Add Subfolder**
   - Masukkan nama subfolder
   - Klik **Buat**

### Subfolder

- Subfolder adalah folder di dalam folder lain
- Membentuk struktur hierarki
- Tidak ada batasan level kedalaman

**Contoh Struktur:**
```
Documents/
├── Finance/
│   ├── Invoices/
│   └── Reports/
├── Legal/
│   └── Contracts/
└── HR/
    └── Policies/
```

## Mengelola Folder

### Rename Folder

1. Klik menu (3 dots) di folder
2. Pilih **Rename**
3. Masukkan nama baru
4. Klik **Simpan**

**Catatan**: Hanya administrator yang dapat rename

### Delete Folder

1. Klik menu (3 dots) di folder
2. Pilih **Delete**
3. Konfirmasi penghapusan

**Catatan**: 
- Hanya administrator yang dapat delete
- Folder yang berisi dokumen tidak bisa dihapus (pindahkan dokumen dulu)

### Membuka Folder

1. **Single Click**: Pilih folder (highlight)
2. **Double Click**: Buka folder dan lihat isinya
3. **Klik Nama Folder**: Buka folder di halaman detail

## Upload ke Folder

### Upload Single File

1. Buka folder yang diinginkan
2. Klik tombol **Upload File**
3. Pilih folder saat upload
4. Upload file

### Batch Upload

1. Buka folder detail
2. Di area upload (label "Multiple files supported")
3. Drag & drop atau browse beberapa file
4. Klik **Upload**
5. Semua file akan tersimpan di folder tersebut

## Navigasi Folder

### Breadcrumb

Di halaman folder detail, breadcrumb menampilkan:
- Path folder saat ini
- Klik folder di breadcrumb untuk navigasi cepat

**Contoh:**
```
Documents > Finance > Invoices
```

### Kembali ke Parent

- Klik tombol **Back** atau folder parent di breadcrumb
- Atau klik "Documents" untuk kembali ke root

## Tips Organisasi

### Naming Convention

- Gunakan nama yang deskriptif dan konsisten
- Contoh: "2025-Invoices", "Q1-Reports", "Legal-Contracts"

### Struktur Hierarki

- Jangan terlalu dalam (maksimal 3-4 level)
- Group berdasarkan kategori utama
- Gunakan subfolder untuk sub-kategori

### Best Practices

- Satu folder untuk satu kategori dokumen
- Gunakan folder untuk grouping, bukan untuk setiap dokumen
- Review dan cleanup folder secara berkala

## Troubleshooting

### Tidak Bisa Buat Folder

- Pastikan Anda memiliki akses (admin/administrator)
- Cek apakah nama folder sudah ada
- Hubungi administrator jika masalah berlanjut

### Folder Tidak Bisa Dihapus

- Pastikan folder kosong (tidak ada dokumen)
- Pindahkan atau hapus dokumen terlebih dahulu
- Hanya administrator yang dapat delete

### Tidak Bisa Rename

- Pastikan Anda memiliki akses (admin/administrator)
- Cek apakah nama baru sudah digunakan folder lain
- Hubungi administrator jika masalah berlanjut

## Langkah Selanjutnya

- [Daftar Dokumen](./document-management) - Kembali ke daftar dokumen
- [Upload Dokumen](./upload-document) - Upload dokumen ke folder
