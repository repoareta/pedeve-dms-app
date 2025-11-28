# 🔧 Setup CORS via gcloud CLI (Recommended)

## Kenapa Pakai CLI?

CORS configuration tidak selalu terlihat di GCP Console UI, terutama untuk bucket dengan Uniform access control. **gcloud CLI adalah cara yang paling reliable.**

## Langkah Setup

### 1. Pastikan gcloud CLI Terinstall dan Login

```bash
# Cek gcloud version
gcloud --version

# Login (jika belum)
gcloud auth login

# Set project
gcloud config set project pedeve-pertamina-dms
```

### 2. Create CORS Config File

**Dari local machine, jalankan:**

```bash
# Create CORS config file
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
```

### 3. Apply CORS Configuration

```bash
# Apply CORS ke bucket
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms
```

**Expected output:**
```
Updated gs://pedeve-dev-bucket.
```

### 4. Verify CORS Configuration

```bash
# Cek CORS config yang sudah di-apply
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="value(cors)" \
  --project=pedeve-pertamina-dms
```

**Atau dengan format JSON:**
```bash
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms
```

### 5. Test CORS

**Dari browser console (setelah setup CORS):**

```javascript
fetch('https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png')
  .then(r => {
    console.log('Status:', r.status);
    console.log('CORS Headers:', r.headers.get('Access-Control-Allow-Origin'));
    return r.blob();
  })
  .then(blob => {
    console.log('✅ Success! Image loaded:', blob.size, 'bytes');
    // Create image element to display
    const img = document.createElement('img');
    img.src = URL.createObjectURL(blob);
    document.body.appendChild(img);
  })
  .catch(e => console.error('❌ Error:', e))
```

## Quick Setup (All-in-One)

```bash
# 1. Create config
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

# 2. Apply CORS
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms

# 3. Verify
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms

# 4. Cleanup
rm -f cors-config.json
```

## Troubleshooting

### Jika Error: "Permission denied"

**Pastikan Anda punya akses:**
```bash
# Cek IAM permissions
gcloud projects get-iam-policy pedeve-pertamina-dms \
  --flatten="bindings[].members" \
  --filter="bindings.members:$(gcloud config get-value account)"
```

**Jika perlu, minta admin untuk berikan role:**
- `roles/storage.admin` atau
- `roles/storage.buckets.update`

### Jika Error: "Bucket not found"

**Pastikan bucket name benar:**
```bash
# List semua buckets
gcloud storage buckets list --project=pedeve-pertamina-dms
```

### Jika CORS Masih Tidak Bekerja

**Cek apakah config ter-apply:**
```bash
# Cek CORS config
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="yaml(cors)" \
  --project=pedeve-pertamina-dms
```

**Pastikan origin frontend sesuai:**
- Frontend domain: `https://pedeve-dev.aretaamany.com`
- Harus exact match dengan yang di CORS config

## Alternative: Make Folder Public (Not Recommended)

Jika CORS masih tidak bekerja, bisa buat folder `logos` public:

```bash
# Make logos folder public (NOT RECOMMENDED for security)
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket/logos
```

**⚠️ Warning:** Ini membuat semua file di folder `logos` bisa diakses publik. Lebih baik fix CORS.

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Setup CORS via CLI (recommended)
2. Test akses gambar dari frontend
3. Verify gambar muncul di form

