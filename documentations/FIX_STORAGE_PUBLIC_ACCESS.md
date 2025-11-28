# 🔓 Fix Storage Public Access untuk Logo

## Masalah

Object yang di-upload ke GCP Storage tidak public, sehingga frontend tidak bisa akses:
```
<Error>
<Code>AccessDenied</Code>
<Message>Access denied.</Message>
</Error>
```

## Solusi

### Opsi 1: Make Folder Public (Quick Fix)

**Di Cloud Shell, jalankan:**

```bash
# Make semua file di folder logos public
gsutil -m acl ch -u AllUsers:R gs://pedeve-dev-bucket/logos/*

# Make future uploads juga public
gsutil defacl ch -u AllUsers:R gs://pedeve-dev-bucket/logos/
```

**Atau pakai script:**

```bash
chmod +x scripts/make-logos-public.sh
./scripts/make-logos-public.sh pedeve-pertamina-dms pedeve-dev-bucket
```

### Opsi 2: Set IAM Policy untuk Bucket (Recommended untuk Uniform Access)

Jika bucket pakai **Uniform access control**, perlu set IAM policy:

```bash
# Berikan permission public read ke folder logos
# (Uniform access control tidak support object-level ACL)

# Option A: Make entire bucket public (NOT RECOMMENDED)
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket

# Option B: Make hanya folder logos public (via IAM condition)
# Ini lebih kompleks, perlu custom IAM policy dengan condition
```

**Untuk Uniform access control, lebih baik pakai Opsi 1 (ACL) atau modify code.**

### Opsi 3: Modify Code (Long-term Solution)

Code sudah di-update untuk set ACL public saat upload. Tapi jika bucket pakai Uniform access control, perlu set IAM policy.

**Deploy code update:**

```bash
# Code sudah di-update di backend/internal/infrastructure/storage/gcp_storage.go
# Setelah deploy, file baru akan otomatis public
```

## Verifikasi

**Test akses public:**

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

## Catatan

- **Uniform access control**: ACL tidak berlaku, perlu IAM policy
- **Fine-grained access control**: ACL bisa digunakan
- **Security**: Hanya folder `logos` yang public, bukan seluruh bucket
- **Future uploads**: Code sudah di-update untuk set ACL public otomatis

## Status

**Date:** 2025-11-28

**Next Steps:**
1. Run `gsutil -m acl ch -u AllUsers:R gs://pedeve-dev-bucket/logos/*` untuk existing files
2. Deploy code update untuk future uploads
3. Test upload logo di frontend

