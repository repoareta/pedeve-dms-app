# 🔧 Fix CORS Config yang Return null

Jika `gcloud storage buckets describe` return `null`, berarti CORS config belum ter-apply dengan benar.

## Solusi: Apply CORS Lagi

### Di Cloud Shell, jalankan:

```bash
# 1. Create CORS config file dengan benar
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

# 2. Verify file sudah benar
cat /tmp/cors-config.json

# 3. Apply CORS
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=/tmp/cors-config.json \
  --project=pedeve-pertamina-dms

# 4. Verify lagi
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms

# 5. Atau dengan format yaml untuk lebih jelas
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="yaml(cors)" \
  --project=pedeve-pertamina-dms
```

## Alternative: Pakai gsutil (jika gcloud storage tidak work)

```bash
# Pakai gsutil untuk set CORS
gsutil cors set /tmp/cors-config.json gs://pedeve-dev-bucket

# Verify
gsutil cors get gs://pedeve-dev-bucket
```

## Troubleshooting

### Jika masih null setelah apply:

1. **Cek apakah file JSON valid:**
```bash
cat /tmp/cors-config.json | python3 -m json.tool
```

2. **Cek permission:**
```bash
gcloud projects get-iam-policy pedeve-pertamina-dms \
  --flatten="bindings[].members" \
  --filter="bindings.members:$(gcloud config get-value account)"
```

3. **Coba dengan gsutil:**
```bash
gsutil cors set /tmp/cors-config.json gs://pedeve-dev-bucket
gsutil cors get gs://pedeve-dev-bucket
```

