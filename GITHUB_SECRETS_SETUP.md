# 🔐 GitHub Secrets Setup Guide

## 📋 Secrets yang Perlu Ditambahkan di GitHub Repository

Buka: **Settings → Secrets and variables → Actions → New repository secret**

### 1. GCP_PROJECT_ID
```
Name: GCP_PROJECT_ID
Value: pedeve-pertamina-dms
```

### 2. GCP_WORKLOAD_IDENTITY_PROVIDER
```
Name: GCP_WORKLOAD_IDENTITY_PROVIDER
Value: projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider
```

### 3. GCP_SERVICE_ACCOUNT
```
Name: GCP_SERVICE_ACCOUNT
Value: github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com
```

### 4. GCP_BACKEND_VM_IP
```
Name: GCP_BACKEND_VM_IP
Value: 34.101.49.147
```

### 5. GCP_FRONTEND_VM_IP
```
Name: GCP_FRONTEND_VM_IP
Value: 34.128.123.1
```

### 6. GCP_SSH_USER (Optional - untuk reference)
```
Name: GCP_SSH_USER
Value: info@aretaamany.com
```
*Note: SSH akan menggunakan OS Login, jadi ini hanya untuk reference*

---

## ✅ Checklist

- [ ] Semua 6 secrets sudah ditambahkan di GitHub
- [ ] WIF Provider sudah dikonfigurasi dengan repository GitHub
- [ ] Service Account memiliki semua IAM roles yang diperlukan
- [ ] GCP Storage bucket `pedeve-dev-bucket` sudah dibuat
- [ ] Cloud SQL Auth Proxy sudah running di backend VM

---

## 🔍 Verifikasi WIF Configuration

Pastikan WIF Provider sudah dikonfigurasi untuk repository GitHub Anda:

```bash
# Cek WIF Provider attributes
gcloud iam workload-identity-pools providers describe github-actions-provider \
  --workload-identity-pool=github-actions-pool \
  --location=global \
  --project=pedeve-pertamina-dms

# Pastikan attribute mapping sudah benar:
# - google.subject = assertion.sub
# - attribute.repository = assertion.repository
# - attribute.actor = assertion.actor
```

---

## 📝 Notes

- **Tidak perlu JSON key** - Semua authentication menggunakan Workload Identity Federation
- **OS Login** - SSH access menggunakan OS Login, tidak perlu private key
- **Secrets di GCP** - Semua application secrets disimpan di GCP Secret Manager
- **Storage** - File uploads akan disimpan di `pedeve-dev-bucket`

