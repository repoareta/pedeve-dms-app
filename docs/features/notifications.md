# Notifikasi

Halaman ini menampilkan semua notifikasi yang relevan untuk Anda. Sistem akan menampilkan push notification otomatis untuk notifikasi penting yang belum ditindak lanjuti.

## Akses Halaman

- Klik icon notifikasi (🔔) di header
- Atau klik menu **Notifikasi** di navigasi atas
- Atau akses langsung: `/notifications`

## Push Notification

Sistem akan menampilkan push notification otomatis di pojok kanan atas layar untuk notifikasi yang belum ditindak lanjuti.

### Karakteristik Push Notification

- **Frekuensi**: Muncul setiap 2 menit sekali untuk notifikasi yang belum ditindak lanjuti
- **Durasi**: Notifikasi akan otomatis hilang setelah 2.5 detik
- **Maksimal**: Maksimal 10 notifikasi ditampilkan per polling cycle
- **Prioritas**: Notifikasi dokumen expired diprioritaskan dan ditampilkan terlebih dahulu
- **Pengulangan**: Push notification akan muncul berulang-ulang sampai notifikasi ditandai sebagai "sudah ditindak lanjuti"

### Styling Push Notification

**Dokumen Sudah Expired:**
- Background warna merah (`#bf4e4e`)
- Border dashed abu-abu
- Text warna putih
- Efek blur pada background

**Dokumen Akan Expired:**
- Background warna default (putih)
- Styling standar Ant Design Vue
- Text warna default

## Tampilan Halaman

Halaman menampilkan daftar notifikasi dengan informasi:

- **Icon**: Icon sesuai jenis notifikasi
- **Pesan**: Pesan notifikasi
- **Waktu**: Waktu notifikasi dibuat
- **Aksi**: Tombol untuk mark as read

**Catatan Penting:**
- Hanya notifikasi yang **belum ditindak lanjuti** yang ditampilkan di halaman
- Notifikasi yang sudah ditandai sebagai "sudah ditindak lanjuti" akan otomatis hilang dari daftar
- Filter "belum ditindak lanjuti" dan "sudah ditindak lanjuti" sudah dihapus karena tidak diperlukan

## Jenis Notifikasi

### Notifikasi Kedaluwarsa Dokumen

**Dokumen Sudah Expired:**
- **Trigger**: Dokumen sudah melewati tanggal kedaluwarsa
- **Pesan**: Menampilkan nama dokumen dan jumlah hari yang sudah expired
- **Styling Push Notification**: Warna merah dengan text putih

**Contoh:**
```
Dokumen 'KTP Rahmat' Sudah Expired
Dokumen 'KTP Rahmat' sudah expired 5 hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.
```

**Dokumen Akan Expired:**
- **Trigger**: Dokumen akan kedaluwarsa dalam X hari (sesuai threshold di Settings)
- **Pesan**: Menampilkan nama dokumen dan jumlah hari tersisa
- **Styling Push Notification**: Warna default (putih)

**Contoh:**
```
Dokumen 'Kontrak Vendor ABC' Akan Expired
Dokumen 'Kontrak Vendor ABC' akan expired dalam 3 hari. Silakan perbarui atau perpanjang dokumen tersebut.
```

**Catatan:**
- Threshold untuk notifikasi "akan expired" dapat diatur di halaman Settings → Pengaturan Notifikasi → "Jumlah Hari Sebelum Expired"
- Default threshold adalah 14 hari
- Notifikasi dibuat sekali per dokumen dan akan muncul berulang sebagai push notification sampai ditindak lanjuti

### Notifikasi Kedaluwarsa Masa Jabatan Direktur

- **Trigger**: Masa jabatan direktur akan berakhir dalam X hari (sesuai threshold)
- **Pesan**: Menampilkan nama direktur, posisi, dan perusahaan

**Contoh:**
```
Masa Jabatan 'John Doe' Akan Berakhir
Masa jabatan John Doe sebagai Direktur Utama di PT ABC akan berakhir dalam 60 hari. Silakan perpanjang atau ganti pengurus tersebut.
```

### Notifikasi Lainnya

- Notifikasi untuk event lainnya sesuai konfigurasi sistem

## Fitur

### Search (Pencarian)

- **Reaktif**: Pencarian berjalan saat Anda mengetik (tidak perlu tekan Enter)
- **Fitur**: Cari notifikasi berdasarkan pesan, judul, atau konten
- **Real-time**: Hasil pencarian update secara real-time

### Mark as Read (Tandai Sudah Ditindak Lanjuti)

**Single Notifikasi:**
1. Klik tombol **Tandai sudah ditindak lanjuti** di notifikasi
2. **Konfirmasi**: Sistem akan menampilkan dialog konfirmasi sebelum menandai notifikasi
3. Klik **OK** untuk konfirmasi
4. Notifikasi akan ditandai sebagai sudah dibaca dan **otomatis hilang dari daftar**

**Peringatan untuk Dokumen Expired:**
- Jika Anda menandai notifikasi dokumen expired sebagai "sudah ditindak lanjuti" tetapi **tidak mengupdate tanggal expired** dokumen tersebut, sistem akan menampilkan peringatan:
  - "Dokumen akan berpotensi terlewat masa aktifnya jika tidak segera diperbarui"
- **Rekomendasi**: Update tanggal expired dokumen sebelum menandai notifikasi sebagai "sudah ditindak lanjuti"

**Catatan:**
- Tombol "Hapus Semua" hanya tersedia untuk role Administrator
- Notifikasi yang sudah ditindak lanjuti akan otomatis hilang dari daftar dan tidak akan muncul lagi

### Unread Count

- Icon notifikasi di header menampilkan jumlah notifikasi yang belum ditindak lanjuti
- Jumlah akan update otomatis setelah mark as read
- **Administrator/Superadmin**: Melihat total notifikasi semua pengguna
- **Admin**: Melihat notifikasi dari perusahaan mereka dan perusahaan anak
- **User Reguler**: Hanya melihat notifikasi mereka sendiri

## Tips

- **Cek Push Notification**: Perhatikan push notification yang muncul otomatis di pojok kanan atas
- **Tindak Lanjuti Segera**: Segera tindak lanjuti notifikasi dokumen expired untuk menghindari dokumen terlewat masa aktifnya
- **Update Dokumen**: Sebelum menandai notifikasi dokumen expired sebagai "sudah ditindak lanjuti", pastikan untuk mengupdate tanggal expired dokumen tersebut
- **Gunakan Search**: Gunakan fitur search untuk menemukan notifikasi spesifik dengan cepat
- **Cek Secara Berkala**: Cek notifikasi secara berkala untuk tidak melewatkan informasi penting

## Troubleshooting

### Push Notification Tidak Muncul

- Pastikan fitur "In-App Notifications" diaktifkan di Settings → Pengaturan Notifikasi
- Pastikan ada notifikasi yang belum ditindak lanjuti
- Push notification muncul setiap 2 menit sekali, tunggu beberapa saat
- Refresh halaman jika push notification tidak muncul setelah beberapa menit

### Notifikasi Tidak Muncul di Halaman

- Pastikan Anda memiliki akses ke data terkait
- Hanya notifikasi yang **belum ditindak lanjuti** yang ditampilkan
- Notifikasi yang sudah ditindak lanjuti akan otomatis hilang dari daftar
- Refresh halaman untuk melihat notifikasi terbaru

### Unread Count Tidak Akurat

- Refresh halaman setelah mark as read
- Cek apakah notifikasi benar-benar sudah di-mark as read
- Unread count akan update otomatis setelah beberapa detik
- Hubungi administrator jika masalah berlanjut

### Push Notification Terlalu Banyak

- Sistem menampilkan maksimal 10 push notification per polling cycle
- Notifikasi dokumen expired diprioritaskan dan ditampilkan terlebih dahulu
- Tindak lanjuti notifikasi yang sudah tidak relevan untuk mengurangi jumlah push notification

## Pengaturan Notifikasi

Anda dapat mengatur notifikasi di halaman Settings → Pengaturan Notifikasi:

- **In-App Notifications**: Aktifkan/nonaktifkan push notification
- **Jumlah Hari Sebelum Expired**: Atur threshold untuk notifikasi dokumen yang akan expired (default: 14 hari)

**Catatan:**
- Threshold ini menentukan berapa hari sebelum dokumen expired, sistem akan membuat notifikasi
- Semakin besar threshold, semakin awal notifikasi dibuat
- Threshold berlaku untuk semua dokumen dan masa jabatan direktur

## Referensi

- [Mark as Read](./mark-notifications-read) - Pelajari cara menandai notifikasi
- [Settings](./settings) - Pelajari pengaturan notifikasi
