# Notifikasi - Mark as Read

Panduan untuk menandai notifikasi sebagai sudah ditindak lanjuti. Setelah ditandai, notifikasi akan otomatis hilang dari daftar.

## Mark Single Notifikasi

### Cara: Tombol di Notifikasi

1. **Buka Halaman Notifikasi**
   - Klik icon notifikasi di header
   - Atau akses: `/notifications`

2. **Klik Tombol "Tandai sudah ditindak lanjuti"**
   - Tombol berada di setiap notifikasi
   - Klik tombol untuk notifikasi yang ingin ditandai

3. **Konfirmasi**
   - Sistem akan menampilkan dialog konfirmasi
   - Klik **OK** untuk mengonfirmasi
   - Notifikasi akan ditandai sebagai sudah ditindak lanjuti
   - **Notifikasi akan otomatis hilang dari daftar**

### Peringatan untuk Dokumen Expired

Jika Anda menandai notifikasi dokumen expired sebagai "sudah ditindak lanjuti" tetapi **tidak mengupdate tanggal expired** dokumen tersebut, sistem akan menampilkan peringatan:

**Peringatan:**
> "Dokumen akan berpotensi terlewat masa aktifnya jika tidak segera diperbarui"

**Rekomendasi:**
- Sebelum menandai notifikasi dokumen expired sebagai "sudah ditindak lanjuti", pastikan untuk:
  1. Update tanggal expired dokumen di halaman detail dokumen
  2. Atau perpanjang dokumen tersebut
  3. Baru kemudian tandai notifikasi sebagai "sudah ditindak lanjuti"

## Mark All Notifications

**Catatan Penting:**
- Tombol "Hapus Semua" hanya tersedia untuk role **Administrator**
- User reguler dan admin tidak akan melihat tombol ini

Untuk Administrator yang ingin menghapus semua notifikasi:

1. **Buka Halaman Notifikasi**
   - Klik icon notifikasi di header
   - Atau akses: `/notifications`

2. **Klik Tombol "Hapus Semua"**
   - Tombol berada di atas daftar notifikasi (hanya untuk Administrator)
   - Klik untuk menghapus semua notifikasi

3. **Konfirmasi**
   - Semua notifikasi akan dihapus
   - Unread count di header akan menjadi 0

## Perilaku Setelah Mark as Read

Setelah notifikasi ditandai sebagai "sudah ditindak lanjuti":

- ✅ Notifikasi akan **otomatis hilang dari daftar** di halaman notifikasi
- ✅ Push notification untuk notifikasi tersebut **tidak akan muncul lagi**
- ✅ Unread count di header akan berkurang
- ✅ Notifikasi tidak akan muncul lagi di polling berikutnya

**Catatan:**
- Notifikasi yang sudah ditindak lanjuti tidak dapat dilihat lagi di halaman notifikasi
- Jika Anda perlu melihat riwayat notifikasi, hubungi administrator

## Unread Count

- **Icon Notifikasi**: Menampilkan jumlah notifikasi yang belum ditindak lanjuti
- **Update Otomatis**: Jumlah akan update setelah mark as read
- **Administrator/Superadmin**: Melihat total notifikasi semua pengguna
- **Admin**: Melihat notifikasi dari perusahaan mereka dan perusahaan anak
- **User Reguler**: Hanya melihat notifikasi mereka sendiri

## Tips

- **Tindak Lanjuti Segera**: Segera tindak lanjuti notifikasi dokumen expired untuk menghindari dokumen terlewat masa aktifnya
- **Update Dokumen**: Sebelum menandai notifikasi dokumen expired, pastikan untuk mengupdate tanggal expired dokumen tersebut
- **Cek Push Notification**: Perhatikan push notification yang muncul otomatis setiap 2 menit
- **Cek Secara Berkala**: Cek notifikasi secara berkala untuk tidak melewatkan informasi penting

## Troubleshooting

### Tombol Tidak Berfungsi

- Refresh halaman dan coba lagi
- Cek koneksi internet
- Pastikan Anda memiliki akses untuk menandai notifikasi tersebut
- Hubungi administrator jika masalah berlanjut

### Unread Count Tidak Update

- Refresh halaman setelah mark as read
- Cek apakah notifikasi benar-benar sudah di-mark as read
- Unread count akan update otomatis setelah beberapa detik
- Clear cache browser jika perlu

### Notifikasi Masih Muncul Setelah Mark as Read

- Notifikasi seharusnya otomatis hilang dari daftar
- Refresh halaman jika notifikasi masih muncul
- Pastikan konfirmasi mark as read berhasil
- Hubungi administrator jika masalah berlanjut

## Referensi

- [Daftar Notifikasi](./notifications) - Kembali ke daftar notifikasi
- [Settings](./settings) - Pelajari pengaturan notifikasi
