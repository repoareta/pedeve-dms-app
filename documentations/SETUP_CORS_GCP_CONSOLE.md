# 🔧 Setup CORS di GCP Console - Step by Step

## Lokasi CORS Configuration

CORS configuration ada di halaman **Configuration** bucket, di bagian bawah halaman.

## Langkah-langkah

### 1. Buka Bucket Configuration

1. Buka **Cloud Storage** di GCP Console
2. Pilih bucket **`pedeve-dev-bucket`**
3. Klik tab **Configuration** (sudah terbuka di screenshot Anda)

### 2. Cari CORS Configuration

**Scroll ke bawah** di halaman Configuration, cari section:
- **"CORS configuration"** atau
- **"Cross-Origin Resource Sharing (CORS)"**

Jika tidak terlihat, mungkin perlu:
- Scroll lebih ke bawah
- Atau klik **"Edit"** di bagian Overview untuk melihat semua settings

### 3. Edit CORS Configuration

1. Klik **"Edit CORS configuration"** atau **"Edit"** button di section CORS
2. Akan muncul text area atau form untuk CORS config
3. Paste config berikut:

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

4. Klik **"Save"** atau **"Update"**

## Alternatif: Via gcloud CLI

Jika tidak menemukan di Console, bisa pakai CLI:

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

# Apply CORS
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms

# Verify
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="value(cors)" \
  --project=pedeve-pertamina-dms
```

## Verifikasi

Setelah setup, test dari browser console:

```javascript
fetch('https://storage.googleapis.com/pedeve-dev-bucket/logos/1764313255_pertamina%20icon.png')
  .then(r => {
    console.log('Status:', r.status);
    console.log('Headers:', [...r.headers.entries()]);
    return r.blob();
  })
  .then(blob => console.log('✅ Success:', blob))
  .catch(e => console.error('❌ Error:', e))
```

## Catatan

- CORS configuration biasanya ada di **bagian bawah** halaman Configuration
- Jika tidak terlihat, coba refresh halaman atau gunakan CLI
- Setelah update, tunggu beberapa detik untuk propagation

