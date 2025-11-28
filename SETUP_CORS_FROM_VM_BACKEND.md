# 🔧 Setup CORS dari VM Backend

VM Backend biasanya sudah punya permission lebih lengkap karena digunakan untuk deployment.

## Langkah

### 1. SSH ke VM Backend

```bash
# Dari GCP Console, buka Cloud Shell atau SSH langsung
gcloud compute ssh backend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms
```

### 2. Setup CORS

```bash
# Create CORS config
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

# Apply CORS
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=/tmp/cors-config.json \
  --project=pedeve-pertamina-dms

# Verify
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms

# Cleanup
rm -f /tmp/cors-config.json
```

## Alternatif: Via GCP Cloud Shell

Jika VM Backend juga error permission, gunakan **GCP Cloud Shell** (ada di GCP Console):

1. Buka GCP Console
2. Klik icon **Cloud Shell** (di pojok kanan atas)
3. Jalankan perintah setup CORS di Cloud Shell

Cloud Shell sudah punya semua permission yang diperlukan.

