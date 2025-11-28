# 🔧 Fix CORS Permission Error di VM Frontend

## Error yang Terjadi

```
ERROR: [1076379007862-compute@developer.gserviceaccount.com] does not have permission to access b instance [pedeve-dev-bucket]
Permission 'storage.buckets.get' denied on resource
```

## Penyebab

VM Frontend menggunakan **default compute service account** yang tidak punya permission untuk update bucket.

## Solusi

### Opsi 1: Login dengan Akun User (Recommended)

**Di VM Frontend, jalankan:**

```bash
# Login dengan akun user yang punya permission
gcloud auth login

# Set project
gcloud config set project pedeve-pertamina-dms

# Cek akun yang aktif
gcloud auth list

# Sekarang coba lagi setup CORS
cat > /tmp/cors-config.json <<'EOF'
[
  {
    "origin": [
      "https://pedeve-dev.aretaamany.com",
      "http://pedeve-dev.aretaamany.com",
      "http://34.128.123.1",
      "http://localhost:5173",
      "http://localhost:8080"
    ],
    "method": ["GET", "HEAD", "OPTIONS"],
    "responseHeader": [
      "Content-Type",
      "Access-Control-Allow-Origin",
      "Access-Control-Allow-Methods",
      "Access-Control-Allow-Headers"
    ],
    "maxAgeSeconds": 3600
  }
]
EOF

gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=/tmp/cors-config.json \
  --project=pedeve-pertamina-dms

gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms

rm -f /tmp/cors-config.json
```

**Catatan:** `gcloud auth login` akan membuka browser untuk login. Jika tidak ada browser di VM, gunakan `gcloud auth login --no-browser` dan copy URL ke local machine.

### Opsi 2: Setup dari Local Machine (Paling Mudah)

**Dari local machine Anda, jalankan:**

```bash
# Pastikan sudah login
gcloud auth list

# Set project
gcloud config set project pedeve-pertamina-dms

# Create CORS config
cat > cors-config.json <<'EOF'
[
  {
    "origin": [
      "https://pedeve-dev.aretaamany.com",
      "http://pedeve-dev.aretaamany.com",
      "http://34.128.123.1",
      "http://localhost:5173",
      "http://localhost:8080"
    ],
    "method": ["GET", "HEAD", "OPTIONS"],
    "responseHeader": [
      "Content-Type",
      "Access-Control-Allow-Origin",
      "Access-Control-Allow-Methods",
      "Access-Control-Allow-Headers"
    ],
    "maxAgeSeconds": 3600
  }
]
EOF

# Apply CORS
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms

# Verify
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms

# Cleanup
rm -f cors-config.json
```

### Opsi 3: Berikan Permission ke Service Account (Untuk Future)

**Jika ingin VM Frontend bisa akses bucket, berikan permission:**

```bash
# Dari local machine atau GCP Console
# Berikan role Storage Admin ke service account VM Frontend

gcloud projects add-iam-policy-binding pedeve-pertamina-dms \
  --member="serviceAccount:1076379007862-compute@developer.gserviceaccount.com" \
  --role="roles/storage.admin"
```

**⚠️ Warning:** Ini memberikan full access ke semua buckets. Lebih baik gunakan custom role atau berikan permission ke bucket tertentu saja.

## Quick Fix (Recommended)

**Gunakan Opsi 2 (Setup dari Local Machine)** karena:
- ✅ Tidak perlu login di VM
- ✅ Lebih cepat dan mudah
- ✅ Tidak perlu ubah IAM permissions
- ✅ Langsung bisa dijalankan

## Verifikasi Setelah Setup

**Test dari browser console:**

```javascript
fetch('https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png')
  .then(r => {
    console.log('Status:', r.status);
    console.log('CORS Headers:', r.headers.get('Access-Control-Allow-Origin'));
    return r.blob();
  })
  .then(blob => {
    console.log('✅ Success! Image loaded:', blob.size, 'bytes');
  })
  .catch(e => console.error('❌ Error:', e))
```

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Setup CORS dari local machine (Opsi 2 - Recommended)
2. Verify CORS config
3. Test akses gambar dari frontend

