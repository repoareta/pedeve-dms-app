# 🔧 Setup CORS dari VM Frontend via SSH

## Prerequisites

1. **gcloud CLI** sudah terinstall di VM Frontend
2. **User sudah login** dengan akun yang punya permission
3. **Project sudah di-set** dengan benar

## Langkah Setup

### 1. SSH ke VM Frontend

```bash
# Dari local machine
gcloud compute ssh frontend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms
```

**Atau jika pakai OS Login:**
```bash
gcloud compute ssh frontend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms \
  --tunnel-through-iap
```

### 2. Cek gcloud CLI

```bash
# Cek apakah gcloud terinstall
gcloud --version

# Cek apakah sudah login
gcloud auth list

# Cek project yang aktif
gcloud config get-value project
```

### 3. Set Project (jika belum)

```bash
gcloud config set project pedeve-pertamina-dms
```

### 4. Create CORS Config File

```bash
# Create CORS config file
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
```

### 5. Apply CORS Configuration

```bash
# Apply CORS ke bucket
gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=/tmp/cors-config.json \
  --project=pedeve-pertamina-dms
```

**Expected output:**
```
Updated gs://pedeve-dev-bucket.
```

### 6. Verify CORS Configuration

```bash
# Cek CORS config yang sudah di-apply
gcloud storage buckets describe gs://pedeve-dev-bucket \
  --format="json(cors)" \
  --project=pedeve-pertamina-dms
```

**Expected output:**
```json
[
  {
    "maxAgeSeconds": 3600,
    "method": [
      "GET",
      "HEAD",
      "OPTIONS"
    ],
    "origin": [
      "https://pedeve-dev.aretaamany.com",
      "http://pedeve-dev.aretaamany.com",
      "http://34.128.123.1",
      "http://localhost:5173",
      "http://localhost:8080"
    ],
    "responseHeader": [
      "Content-Type",
      "Access-Control-Allow-Origin",
      "Access-Control-Allow-Methods",
      "Access-Control-Allow-Headers"
    ]
  }
]
```

### 7. Cleanup

```bash
# Hapus file temporary
rm -f /tmp/cors-config.json
```

## Quick Setup (All-in-One)

```bash
# SSH ke VM Frontend
gcloud compute ssh frontend-dev \
  --zone=asia-southeast2-a \
  --project=pedeve-pertamina-dms

# Setelah masuk ke VM, jalankan:
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

## Troubleshooting

### Error: "gcloud: command not found"

**Install gcloud CLI di VM:**
```bash
# Download dan install gcloud CLI
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
gcloud init
```

### Error: "Permission denied"

**Cek IAM permissions:**
```bash
# Cek akun yang sedang login
gcloud auth list

# Cek permissions
gcloud projects get-iam-policy pedeve-pertamina-dms \
  --flatten="bindings[].members" \
  --filter="bindings.members:$(gcloud config get-value account)"
```

**Jika perlu, minta admin untuk berikan role:**
- `roles/storage.admin` atau
- `roles/storage.buckets.update`

### Error: "Project not found"

**Set project dengan benar:**
```bash
gcloud config set project pedeve-pertamina-dms
gcloud config get-value project
```

### Error: "Bucket not found"

**Cek bucket name:**
```bash
# List semua buckets
gcloud storage buckets list --project=pedeve-pertamina-dms
```

## Alternative: Setup dari Local Machine

Jika ada masalah dengan VM Frontend, bisa setup dari **local machine**:

```bash
# Dari local machine (lebih mudah)
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

gcloud storage buckets update gs://pedeve-dev-bucket \
  --cors-file=cors-config.json \
  --project=pedeve-pertamina-dms

rm -f cors-config.json
```

## Catatan

- **Setup CORS bisa dari mana saja** (local, VM frontend, VM backend)
- **Yang penting:** gcloud CLI terinstall dan user punya permission
- **Lebih mudah dari local machine** karena tidak perlu SSH
- **Setelah setup, tunggu beberapa detik** untuk propagation
- **Test dari frontend** untuk memastikan gambar bisa diakses

## Status

**Date:** 2025-11-28

**Next Steps:**
1. SSH ke VM Frontend
2. Setup CORS via CLI
3. Verify CORS config
4. Test akses gambar dari frontend

