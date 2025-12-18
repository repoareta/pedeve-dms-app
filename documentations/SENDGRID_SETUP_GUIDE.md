# 📧 Panduan Setup SendGrid untuk Testing Notifikasi Email

## 📋 Daftar Isi
1. [Setup Akun SendGrid](#1-setup-akun-sendgrid)
2. [Membuat API Key](#2-membuat-api-key)
3. [Verifikasi Sender Email](#3-verifikasi-sender-email-opsional)
4. [Konfigurasi Environment Variables](#4-konfigurasi-environment-variables)
5. [Install Dependencies](#5-install-dependencies)
6. [Testing Email](#6-testing-email)

---

## 1. Setup Akun SendGrid

### Langkah-langkah:
1. **Kunjungi SendGrid**: https://sendgrid.com
2. **Daftar Akun Baru**:
   - Klik "Start for Free"
   - Isi form pendaftaran
   - Pilih plan **Free** (100 email/hari gratis)
3. **Verifikasi Email**:
   - Cek inbox email yang didaftarkan
   - Klik link verifikasi dari SendGrid

### Catatan:
- **Free Tier**: 100 email per hari (cukup untuk testing)
- **Paid Tier**: Mulai dari $15/bulan untuk 40,000 email
- Untuk production, pertimbangkan upgrade ke paid tier

---

## 2. Membuat API Key

### Langkah-langkah:
1. **Login ke SendGrid Dashboard**
2. **Navigasi ke Settings → API Keys**:
   - Klik menu **Settings** (ikon gear di sidebar kiri)
   - Pilih **API Keys**
3. **Create API Key**:
   - Klik tombol **"Create API Key"** (hijau di kanan atas)
   - **Name**: `Pedeve DMS Development` (atau nama lain yang mudah diingat)
   - **API Key Permissions**: 
     - **Untuk Testing**: Pilih **"Full Access"** (lebih mudah)
     - **Untuk Production**: Pilih **"Restricted Access"** → Centang **"Mail Send"** permission saja
4. **Simpan API Key**:
   - ⚠️ **PENTING**: API key hanya ditampilkan **sekali saja**!
   - **Copy dan simpan** API key di tempat yang aman (password manager, dll)
   - Format: `SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### Catatan:
- API key tidak bisa dilihat lagi setelah dibuat
- Jika lupa, harus buat API key baru
- Satu akun bisa punya multiple API keys

---

## 3. Verifikasi Sender Email (Opsional)

### Untuk Testing:
- **Tidak wajib** - SendGrid free tier bisa kirim email tanpa verifikasi domain
- Email akan dikirim dari `noreply@sendgrid.net` atau email yang sudah terverifikasi

### Untuk Production:
1. **Single Sender Verification** (Mudah):
   - Settings → Sender Authentication → Single Sender Verification
   - Klik "Create New Sender"
   - Isi form dengan email yang akan digunakan (misal: `noreply@pedeve-dev.aretaamany.com`)
   - Verifikasi email yang dikirim SendGrid

2. **Domain Authentication** (Recommended untuk Production):
   - Settings → Sender Authentication → Domain Authentication
   - Tambahkan domain (misal: `aretaamany.com`)
   - Ikuti instruksi untuk menambahkan DNS records
   - Tunggu verifikasi (bisa beberapa jam)

---

## 4. Konfigurasi Environment Variables

### Untuk Development (Local):

Tambahkan ke file `.env` di root backend atau set sebagai environment variable:

```bash
# SendGrid Configuration
SENDGRID_API_KEY=SG.your-api-key-here
EMAIL_FROM=noreply@pedeve-dev.aretaamany.com

# ATAU jika menggunakan SMTP (alternatif)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
EMAIL_FROM=noreply@pedeve-dev.aretaamany.com
```

### Untuk Docker Compose:

Tambahkan ke `docker-compose.dev.yml`:

```yaml
backend:
  environment:
    - SENDGRID_API_KEY=${SENDGRID_API_KEY}
    - EMAIL_FROM=${EMAIL_FROM:-noreply@pedeve-dev.aretaamany.com}
```

### Untuk Production (GCP):

Tambahkan ke **GCP Secret Manager**:

```bash
# Via gcloud CLI
gcloud secrets create sendgrid-api-key \
  --data-file=- \
  --project=pedeve-pertamina-dms

# Atau via Console:
# 1. Buka GCP Console → Secret Manager
# 2. Create Secret: `sendgrid-api-key`
# 3. Paste API key
# 4. Tambahkan ke environment variables di VM atau Cloud Run
```

Atau set sebagai environment variable di VM:

```bash
# SSH ke backend VM
sudo nano /etc/environment

# Tambahkan:
SENDGRID_API_KEY=SG.your-api-key-here
EMAIL_FROM=noreply@pedeve-dev.aretaamany.com
```

---

## 5. Install Dependencies

### Backend (Go):

SendGrid Go SDK belum diinstall. Tambahkan ke `backend/go.mod`:

```bash
cd backend
go get github.com/sendgrid/sendgrid-go
go mod tidy
```

Atau tambahkan manual ke `go.mod`:

```go
require (
    // ... dependencies lainnya
    github.com/sendgrid/sendgrid-go v3.14.0+incompatible
)
```

Lalu jalankan:
```bash
go mod download
go mod tidy
```

---

## 6. Testing Email

### Manual Test via API:

Setelah email service diimplementasikan, test dengan:

1. **Via Swagger UI**:
   - Buka: `http://localhost:8080/swagger/index.html`
   - Cari endpoint untuk test email
   - Atau gunakan endpoint notification yang sudah ada

2. **Via cURL**:
```bash
curl -X POST http://localhost:8080/api/v1/notifications/test-email \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "your-email@example.com",
    "subject": "Test Email",
    "body": "This is a test email from Pedeve DMS"
  }'
```

3. **Via Frontend**:
   - Upload dokumen dengan `expiry_date` yang akan expired dalam beberapa hari
   - Tunggu scheduler/cron job check expiring documents
   - Cek inbox email

### Checklist Testing:

- [ ] Email terkirim ke inbox (bukan spam)
- [ ] Format email benar (HTML rendering)
- [ ] Link di email berfungsi
- [ ] Email masuk ke folder yang benar (Inbox, bukan Spam)
- [ ] Rate limit tidak terlampaui (100 email/hari untuk free tier)

---

## 🔍 Troubleshooting

### Email Tidak Terkirim:

1. **Cek API Key**:
   - Pastikan API key benar dan aktif
   - Cek di SendGrid Dashboard → API Keys

2. **Cek Logs**:
   ```bash
   # Backend logs
   docker-compose logs backend | grep -i email
   
   # Atau jika run langsung
   tail -f backend.log | grep -i email
   ```

3. **Cek SendGrid Activity**:
   - Buka SendGrid Dashboard → Activity
   - Lihat apakah email terkirim atau ada error

4. **Cek Spam Folder**:
   - Email mungkin masuk ke spam folder
   - Verifikasi sender email untuk mengurangi risiko spam

### Error: "API key is invalid"

- Pastikan API key di-copy dengan benar (tidak ada spasi)
- Pastikan API key masih aktif (tidak di-delete)
- Buat API key baru jika perlu

### Error: "Rate limit exceeded"

- Free tier: 100 email/hari
- Tunggu hingga reset (setiap hari)
- Atau upgrade ke paid tier

---

## 📚 Referensi

- **SendGrid Documentation**: https://docs.sendgrid.com/
- **SendGrid Go SDK**: https://github.com/sendgrid/sendgrid-go
- **SendGrid API Reference**: https://docs.sendgrid.com/api-reference
- **Email Best Practices**: https://docs.sendgrid.com/ui/sending-email/best-practices

---

## ✅ Checklist Setup

- [ ] Akun SendGrid dibuat dan terverifikasi
- [ ] API Key dibuat dan disimpan dengan aman
- [ ] Environment variables diset (`SENDGRID_API_KEY`, `EMAIL_FROM`)
- [ ] Dependencies diinstall (`github.com/sendgrid/sendgrid-go`)
- [ ] Email service diimplementasikan di backend
- [ ] Test email berhasil terkirim
- [ ] Email masuk ke inbox (bukan spam)

---

**Catatan**: Email notification service belum diimplementasikan di backend. Setelah setup SendGrid selesai, perlu implementasi email service di `backend/internal/infrastructure/email/email.go` sesuai dengan dokumentasi di `NOTIFICATION_SYSTEM_DESIGN.md`.

