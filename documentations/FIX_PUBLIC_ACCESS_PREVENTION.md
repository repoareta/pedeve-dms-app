# 🔓 Fix Public Access Prevention untuk GCP Storage

## Masalah

Bucket memiliki **Public Access Prevention** yang di-enforce:
```
PreconditionException: 412 The member bindings allUsers and allAuthenticatedUsers 
are not allowed since public access prevention is enforced.
```

## Solusi

### Opsi 1: Disable Public Access Prevention (Quick Fix)

**Di GCP Console:**

1. Buka **Cloud Storage** → **Buckets**
2. Klik bucket **`pedeve-dev-bucket`**
3. Klik tab **Configuration**
4. Scroll ke section **Permissions**
5. Cari **"Public access prevention"**
6. Klik **"Edit"** (pencil icon)
7. Pilih **"Enforced"** → ubah ke **"Not enforced"** atau **"Inherited"**
8. Klik **"Save"**

**Atau via gcloud CLI (di Cloud Shell):**

```bash
# Disable public access prevention
gcloud storage buckets update gs://pedeve-dev-bucket \
  --public-access-prevention=inherited \
  --project=pedeve-pertamina-dms

# Setelah itu, bisa set IAM untuk public access
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket
```

**⚠️ Warning:** Ini akan membuat bucket bisa diakses public. Pastikan tidak ada data sensitif.

### Opsi 2: Gunakan Signed URLs (More Secure, but Complex)

**Implementasi Signed URLs di backend:**

1. Backend generate signed URL dengan expiry time (e.g., 1 hour)
2. Frontend menggunakan signed URL untuk akses file
3. Tidak perlu disable Public Access Prevention

**Ini memerlukan modifikasi code untuk generate signed URLs dengan service account credentials.**

### Opsi 3: Backend Proxy (Alternative)

**Backend serve file sebagai proxy:**

1. Frontend request ke backend: `/api/v1/files/logo/{filename}`
2. Backend fetch file dari GCP Storage
3. Backend return file ke frontend
4. Tidak perlu public access

**Ini memerlukan modifikasi code untuk add endpoint baru.**

## Quick Fix (Recommended untuk Development)

**Disable Public Access Prevention via GCP Console:**

1. Buka GCP Console → Cloud Storage → Buckets
2. Klik `pedeve-dev-bucket`
3. Tab **Configuration** → **Permissions**
4. **Public access prevention** → **Edit** → **Not enforced** → **Save**
5. Setelah itu, jalankan:

```bash
# Set IAM untuk public read
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket

# Verify
curl -I https://storage.googleapis.com/pedeve-dev-bucket/logos/1764315232_pertamina%20icon.png
# Harus return HTTP 200
```

## Verifikasi

**Test akses:**

```bash
# Test dengan curl
curl -I https://storage.googleapis.com/pedeve-dev-bucket/logos/1764315232_pertamina%20icon.png

# Harus return HTTP 200, bukan 403
```

**Test dari browser:**

```javascript
fetch('https://storage.googleapis.com/pedeve-dev-bucket/logos/1764315232_pertamina%20icon.png')
  .then(r => {
    console.log('Status:', r.status);
    if (r.status === 200) {
      console.log('✅ Public access working!');
      return r.blob();
    } else {
      throw new Error('Access denied');
    }
  })
  .then(blob => {
    console.log('✅ Success!', blob.size, 'bytes');
    const img = document.createElement('img');
    img.src = URL.createObjectURL(blob);
    document.body.appendChild(img);
  })
  .catch(e => console.error('❌ Error:', e))
```

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Disable Public Access Prevention via GCP Console
2. Set IAM untuk public read: `gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket`
3. Test upload logo di frontend

