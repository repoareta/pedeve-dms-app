# 🔧 Fix: GCP Storage CORS Error untuk Logo Images

## Masalah

File logo berhasil di-upload ke GCP Storage, tapi gambar tidak muncul di frontend dengan error CORS:

```
fetch("https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png", {
  "mode": "no-cors",
  ...
})
```

**Error:** Browser tidak bisa akses gambar dari GCP Storage karena CORS policy tidak dikonfigurasi.

## Penyebab

GCP Storage bucket `pedeve-dev-bucket` tidak memiliki CORS configuration, sehingga browser memblokir cross-origin requests dari frontend (`pedeve-dev.aretaamany.com`) ke GCP Storage (`storage.googleapis.com`).

## Solusi: Setup CORS di GCP Storage

### Opsi 1: Setup CORS via Script (Recommended)

**Jalankan script dari local machine:**

```bash
# Pastikan gcloud sudah login dan punya akses
gcloud auth login

# Setup CORS
chmod +x scripts/setup-gcp-storage-cors.sh
./scripts/setup-gcp-storage-cors.sh pedeve-pertamina-dms pedeve-dev-bucket
```

### Opsi 2: Setup CORS Manual via gcloud CLI

**Jalankan command berikut:**

```bash
# Create CORS config file
cat > cors-config.json <<EOF
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

# Apply CORS configuration
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms

# Verify
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="value(cors)" \
  --project=pedeve-pertamina-dms
```

### Opsi 3: Setup CORS via GCP Console

1. Buka **Cloud Storage** di GCP Console
2. Pilih bucket `pedeve-dev-bucket`
3. Klik tab **Configuration**
4. Scroll ke **CORS configuration**
5. Klik **Edit CORS configuration**
6. Paste config berikut:

```json
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
```

7. Klik **Save**

## Verifikasi

**Setelah CORS dikonfigurasi, test dari browser:**

1. Buka browser console
2. Test fetch:
```javascript
fetch('https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png')
  .then(r => r.blob())
  .then(blob => console.log('✅ Success:', blob))
  .catch(e => console.error('❌ Error:', e))
```

**Atau test dengan curl:**
```bash
curl -H "Origin: https://pedeve-dev.aretaamany.com" \
  -H "Access-Control-Request-Method: GET" \
  -X OPTIONS \
  https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png \
  -v
```

**Expected response headers:**
```
Access-Control-Allow-Origin: https://pedeve-dev.aretaamany.com
Access-Control-Allow-Methods: GET, HEAD, OPTIONS
```

## Alternative: Make Bucket Public (Not Recommended)

**Jika CORS tidak bekerja, bisa buat bucket public untuk folder logos:**

```bash
# Make logos folder public
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket/logos

# Atau make entire bucket public (NOT RECOMMENDED for security)
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket
```

**⚠️ Warning:** Making bucket public bisa security risk. Lebih baik pakai CORS.

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Setup CORS di GCP Storage bucket
2. Test akses gambar dari frontend
3. Verify gambar muncul di form

