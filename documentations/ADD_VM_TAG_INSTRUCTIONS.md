# Cara Menambahkan Network Tag ke VM di GCP Console

## ⚠️ Penting: Network Tags vs Resource Hierarchy Tags

Ada **2 jenis tags** di GCP:
1. **Network Tags** - Untuk firewall rules (yang kita butuhkan)
2. **Resource Hierarchy Tags** - Untuk IAM/policy management (bukan yang kita butuhkan)

## Cara Menambahkan Network Tag (Yang Benar)

### Opsi 1: Via GCP Console - Edit VM

1. **Buka halaman VM:**
   - Go to: https://console.cloud.google.com/compute/instances?project=pedeve-pertamina-dms
   - Klik VM: `backend-dev`

2. **Edit VM:**
   - Klik tombol **"EDIT"** (di bagian atas)

3. **Cari bagian "Networking":**
   - Scroll ke bawah sampai bagian **"Networking"** atau **"Network interfaces"**
   - Di bagian ini, ada field **"Network tags"**

4. **Tambah tag:**
   - Di field "Network tags", ketik: `backend-api-server`
   - Atau klik dropdown dan pilih dari existing tags
   - Tag akan muncul sebagai chip/badge

5. **Save:**
   - Scroll ke bawah, klik **"SAVE"**

### Opsi 2: Via gcloud CLI (Dari Local Machine)

```bash
# Pastikan sudah login dan set project
gcloud auth login
gcloud config set project pedeve-pertamina-dms

# Tambah network tag ke backend VM
gcloud compute instances add-tags backend-dev \
  --tags backend-api-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

### Opsi 3: Via gcloud CLI (Dari VM - Jika Punya Permission)

```bash
# SSH ke backend VM
gcloud compute ssh backend-dev --zone=asia-southeast2-a

# Tambah tag (jika service account punya permission)
gcloud compute instances add-tags backend-dev \
  --tags backend-api-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

## Verifikasi Tag Sudah Ditambahkan

### Via GCP Console:
1. Buka VM detail page
2. Scroll ke bagian "Network interfaces"
3. Cek apakah tag `backend-api-server` muncul

### Via gcloud CLI:
```bash
gcloud compute instances describe backend-dev \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms \
  --format="get(tags.items)"
```

**Output yang diharapkan:**
```
backend-api-server
```

## Troubleshooting

**Jika tidak menemukan "Network tags" field:**
- Pastikan Anda di halaman **"EDIT"** VM, bukan detail view
- Scroll ke bagian **"Networking"** atau **"Network interfaces"**
- Network tags biasanya ada di bagian bawah form edit

**Jika tag tidak muncul setelah save:**
- Tunggu beberapa detik (propagation delay)
- Refresh halaman
- Cek via gcloud CLI untuk verifikasi

## Setelah Tag Ditambahkan

Test apakah firewall rule bekerja:

```bash
# Test dari external
curl http://34.101.49.147:8080/health
curl http://34.101.49.147:8080/api/v1/csrf-token
```

Jika berhasil, berarti firewall rule sudah aktif!

