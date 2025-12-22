# Laporan Keuangan - Tambah Laporan

Panduan untuk menambahkan laporan keuangan baru.

## Akses Halaman

- Dari halaman daftar laporan, klik tombol **Tambah Laporan**
- Atau akses langsung: `/reports/new`

## Form Tambah Laporan

### Informasi Dasar

**Field Wajib:**
- **Periode** *: Pilih periode laporan (format: YYYY-MM)
- **Perusahaan** *: Pilih perusahaan
- **Revenue** *: Total pendapatan
- **OPEX** *: Total biaya operasional
- **NPAT** *: Net Profit After Tax
- **Dividend** *: Dividen
- **Financial Ratio** *: Rasio keuangan (Revenue/OPEX)

**Field Opsional:**
- **Inputter**: User yang menginput (default: Anda)
- **Remark**: Catatan tambahan
- **Attachment**: File pendukung (Excel, PDF, dll)

### Validasi Otomatis

Sistem akan memvalidasi:
- Periode harus format YYYY-MM (contoh: 2025-01)
- Tidak boleh ada duplicate periode untuk perusahaan yang sama
- Financial Ratio akan terhitung otomatis (Revenue/OPEX)
- Company ID harus valid
- Inputter ID harus valid (jika diisi)

### Perhitungan Otomatis

- **Financial Ratio**: Otomatis terhitung = Revenue / OPEX
- Sistem akan menampilkan perhitungan di form

## Menyimpan Laporan

1. **Isi Form**
   - Isi semua field wajib
   - Pastikan data sudah benar

2. **Review**
   - Cek perhitungan financial ratio
   - Pastikan periode benar

3. **Klik "Simpan"**
   - Sistem akan menyimpan laporan
   - Jika berhasil, Anda akan diarahkan ke halaman daftar laporan

4. **Jika Ada Error**
   - Sistem akan menampilkan pesan error
   - Perbaiki field yang error
   - Coba simpan lagi

## Tips

- Pastikan periode benar sebelum simpan
- Cek apakah laporan untuk periode tersebut sudah ada
- Gunakan remark untuk catatan penting
- Upload attachment jika ada file pendukung

## Troubleshooting

### Periode Sudah Ada

- Tidak boleh ada duplicate periode untuk perusahaan yang sama
- Jika perlu update, gunakan fitur edit
- Atau hapus laporan lama terlebih dahulu

### Financial Ratio Tidak Valid

- Pastikan OPEX tidak 0 (akan menyebabkan error division by zero)
- Sistem akan menghitung otomatis, pastikan Revenue dan OPEX benar

### Perusahaan Tidak Muncul di Dropdown

- Pastikan Anda memiliki akses ke perusahaan tersebut
- Hubungi administrator jika perusahaan seharusnya muncul

## Langkah Selanjutnya

- [Edit Laporan](./edit-report) - Pelajari cara edit laporan
- [Bulk Upload](./bulk-upload-reports) - Pelajari bulk upload untuk input banyak laporan
