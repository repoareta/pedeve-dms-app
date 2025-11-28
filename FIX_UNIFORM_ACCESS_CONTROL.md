# 🔓 Fix Uniform Access Control untuk Public Access

## Masalah

Bucket menggunakan **Uniform access control**, sehingga ACL tidak bisa digunakan:
```
CommandException: Failed to set acl for gs://pedeve-dev-bucket/logos/...
Please ensure you have OWNER-role access to this resource.
```

## Solusi

### Opsi 1: Set IAM Policy untuk Folder Logos (Recommended)

**Untuk Uniform access control, perlu set IAM policy dengan condition:**

```bash
# Di Cloud Shell, jalankan:

# 1. Create IAM policy JSON dengan condition untuk folder logos
cat > /tmp/logos-iam-policy.json <<'EOF'
{
  "bindings": [
    {
      "members": ["allUsers"],
      "role": "roles/storage.objectViewer",
      "condition": {
        "title": "Allow public read access to logos folder",
        "description": "Allow public read access to objects in logos folder",
        "expression": "resource.name.startsWith('projects/_/buckets/pedeve-dev-bucket/objects/logos/')"
      }
    }
  ]
}
EOF

# 2. Apply IAM policy ke bucket
gcloud storage buckets update gs://pedeve-dev-bucket \
  --iam-config-file=/tmp/logos-iam-policy.json \
  --project=pedeve-pertamina-dms

# 3. Cleanup
rm -f /tmp/logos-iam-policy.json
```

**⚠️ Note:** IAM condition untuk folder prefix mungkin tidak didukung langsung. Alternatif: set IAM untuk seluruh bucket dengan condition yang lebih spesifik, atau ubah ke fine-grained access control.

### Opsi 2: Set IAM Policy untuk Entire Bucket (Simple, tapi kurang secure)

**Jika condition tidak work, bisa set IAM untuk seluruh bucket:**

```bash
# Berikan public read access ke seluruh bucket
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket
```

**⚠️ Warning:** Ini membuat semua file di bucket bisa diakses public. Pastikan tidak ada data sensitif.

### Opsi 3: Ubah ke Fine-Grained Access Control (Complex)

**Ubah bucket ke fine-grained access control agar bisa pakai ACL:**

```bash
# 1. Disable uniform access control
gsutil uniformbucketlevelaccess set off gs://pedeve-dev-bucket

# 2. Setelah itu, bisa pakai ACL
gsutil -m acl ch -u AllUsers:R gs://pedeve-dev-bucket/logos/*

# 3. Set default ACL untuk future uploads
gsutil defacl ch -u AllUsers:R gs://pedeve-dev-bucket
```

**⚠️ Warning:** Mengubah access control mode bisa mempengaruhi existing permissions. Pastikan backup dulu.

### Opsi 4: Pakai Signed URLs (Most Secure, but Complex)

**Generate signed URLs di backend untuk akses temporary:**

Ini memerlukan modifikasi code untuk generate signed URLs dengan service account private key. Lebih kompleks tapi lebih aman.

## Quick Fix (Recommended untuk Development)

**Untuk development, pakai Opsi 2 (set IAM untuk entire bucket):**

```bash
# Di Cloud Shell
gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket

# Verify
curl -I https://storage.googleapis.com/pedeve-dev-bucket/logos/1764315232_pertamina%20icon.png
# Harus return HTTP 200
```

**Untuk production, pertimbangkan:**
- Opsi 3 (fine-grained access control) untuk kontrol lebih detail
- Opsi 4 (signed URLs) untuk security maksimal

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
1. Run `gsutil iam ch allUsers:objectViewer gs://pedeve-dev-bucket` untuk quick fix
2. Atau ubah ke fine-grained access control untuk kontrol lebih detail
3. Test upload logo di frontend

