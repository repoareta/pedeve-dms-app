# Laporan Keuangan - Edit Laporan

Panduan untuk mengedit laporan keuangan yang sudah ada.

## Akses Halaman

- Dari halaman daftar laporan, klik tombol **Edit** di kolom Aksi
- Atau dari halaman detail laporan, klik tombol **Edit**
- Atau akses langsung: `/reports/{id}/edit`

## Form Edit

Form edit sama dengan form tambah, tetapi sudah terisi dengan data yang ada.

### Field yang Dapat Diedit

- **Periode**: Dapat diubah (jika belum ada duplicate)
- **Perusahaan**: Dapat diubah
- **Revenue**: Dapat diubah
- **OPEX**: Dapat diubah
- **NPAT**: Dapat diubah
- **Dividend**: Dapat diubah
- **Financial Ratio**: Akan terhitung ulang otomatis
- **Inputter**: Dapat diubah
- **Remark**: Dapat diubah
- **Attachment**: Dapat diubah atau dihapus

## Validasi

Sama seperti form tambah:
- Periode harus format YYYY-MM
- Tidak boleh ada duplicate periode untuk perusahaan yang sama
- Financial Ratio akan terhitung ulang otomatis

## Menyimpan Perubahan

1. **Review Perubahan**
   - Pastikan semua perubahan sudah benar
   - Cek perhitungan financial ratio

2. **Klik "Update"**
   - Sistem akan menyimpan perubahan
   - Jika berhasil, Anda akan diarahkan ke halaman detail laporan

3. **Jika Ada Error**
   - Sistem akan menampilkan pesan error
   - Perbaiki field yang error
   - Coba simpan lagi

## Tips

- Pastikan periode masih valid setelah edit
- Cek perhitungan financial ratio setelah perubahan
- Backup data penting sebelum melakukan perubahan besar

## Troubleshooting

### Tidak Bisa Edit Periode

- Jika periode baru sudah digunakan perusahaan lain, tidak bisa diubah
- Gunakan periode yang berbeda atau hapus laporan lama

### Financial Ratio Tidak Update

- Pastikan Revenue dan OPEX sudah diubah
- Sistem akan menghitung ulang otomatis saat simpan
- Refresh halaman jika perhitungan tidak muncul

## Langkah Selanjutnya

- [Detail Laporan](./financial-reports) - Lihat perubahan yang sudah disimpan
- [Daftar Laporan](./financial-reports) - Kembali ke daftar laporan
