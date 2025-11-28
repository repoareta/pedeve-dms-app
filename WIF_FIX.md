# 🔧 Perbaikan Workload Identity Federation Error

## ❌ Error yang Terjadi

```
Error: failed to generate Google Cloud federated token: 
{"error":"invalid_grant","error_description":"The audience in ID Token [https://iam.googleapis.com/***] 
does not match the expected audience https://github.com/repoareta/pedeve-dms-app."}
```

## 🔍 Penyebab

WIF Provider di GCP dikonfigurasi dengan **audience** yang salah. Untuk GitHub Actions, audience harus sesuai dengan repository GitHub (`https://github.com/repoareta/pedeve-dms-app`), bukan `iam.googleapis.com`.

**Error menunjukkan:** GitHub Actions mengirim ID token dengan audience `https://iam.googleapis.com/***`, tetapi WIF Provider mengharapkan `https://github.com/repoareta/pedeve-dms-app`.

**Solusi:** Tambahkan parameter `audience` secara eksplisit di GitHub Actions workflow untuk memastikan token menggunakan audience yang benar.

**Kemungkinan penyebab:**
1. WIF Provider masih menggunakan "Default audience" atau audience yang salah
2. Konfigurasi belum benar-benar ter-update di GCP
3. Perlu refresh/update ulang konfigurasi WIF Provider

## ✅ Solusi: Update WIF Provider Configuration + Workflow

**PENTING:** Perlu 2 langkah:
1. Update WIF Provider di GCP (sudah dilakukan)
2. Update GitHub Actions workflow untuk set audience secara eksplisit (perlu dilakukan)

Jalankan command berikut di GCP Console atau gcloud CLI untuk memperbaiki konfigurasi:

### 1. Update WIF Provider dengan Audience yang Benar

```bash
gcloud iam workload-identity-pools providers update-oidc github-actions-provider \
  --workload-identity-pool=github-actions-pool \
  --location=global \
  --project=pedeve-pertamina-dms \
  --issuer-uri=https://token.actions.githubusercontent.com \
  --allowed-audiences="https://github.com/repoareta/pedeve-dms-app" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository == 'repoareta/pedeve-dms-app'"
```

### 2. Atau Update via Console

1. Buka **IAM & Admin → Workload Identity Federation** di GCP Console
2. Pilih pool: `github-actions-pool`
3. Pilih provider: `github-actions-provider`
4. Klik **EDIT**
5. Pastikan konfigurasi berikut:

**OIDC Configuration:**
- **Issuer URL:** `https://token.actions.githubusercontent.com`
- **Allowed audiences:** `https://github.com/repoareta/pedeve-dms-app`

**Attribute Mapping:**
- `google.subject` = `assertion.sub` (WAJIB - untuk identity)
- `attribute.actor` = `assertion.actor` (opsional - untuk audit)
- `attribute.repository` = `assertion.repository` (opsional - untuk filtering)

**⚠️ PENTING:** 
- Jangan tambahkan mapping `assertion.repository` di kolom Google (itu bukan Google attribute)
- Hapus mapping Google 5 → OIDC 5 jika ada
- Cukup 3 mapping di atas sudah cukup

**Attribute Conditions (WAJIB):**
Tambahkan condition untuk membatasi hanya repository yang diizinkan:
```
assertion.repository == 'repoareta/pedeve-dms-app'
```

6. Klik **SAVE**

### 3. Verifikasi Konfigurasi WIF Provider

Jalankan command berikut untuk memverifikasi konfigurasi:

```bash
gcloud iam workload-identity-pools providers describe github-actions-provider \
  --workload-identity-pool=github-actions-pool \
  --location=global \
  --project=pedeve-pertamina-dms
```

**Pastikan output menunjukkan:**
- `issuerUri: https://token.actions.githubusercontent.com`
- `allowedAudiences` berisi `https://github.com/repoareta/pedeve-dms-app` (BUKAN `iam.googleapis.com`)

**Jika `allowedAudiences` masih menunjukkan `iam.googleapis.com`, update dengan command:**

```bash
gcloud iam workload-identity-pools providers update-oidc github-actions-provider \
  --workload-identity-pool=github-actions-pool \
  --location=global \
  --project=pedeve-pertamina-dms \
  --allowed-audiences="https://github.com/repoareta/pedeve-dms-app"
```

**Atau via Console:**
1. Buka WIF Provider di GCP Console
2. Pastikan di tab "Provider details", bagian "Audiences":
   - Pilih **"Allowed audiences"** (bukan "Default audience")
   - Audience 1: `https://github.com/repoareta/pedeve-dms-app`
   - **HAPUS** audience lain jika ada (terutama yang berisi `iam.googleapis.com`)
3. Simpan perubahan

## 📝 Catatan Penting

1. **Repository Path:** Pastikan repository path di audience sesuai dengan format `https://github.com/OWNER/REPO`
2. **Attribute Condition:** Pastikan condition membatasi hanya untuk repository yang diizinkan
3. **Service Account Binding:** Pastikan service account `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com` sudah di-bind dengan WIF Provider

### Cek Service Account Binding (PENTING!)

**Ini adalah langkah kritis yang sering terlewat!** Service account harus di-bind dengan WIF Provider agar bisa digunakan.

```bash
# Cek binding saat ini
gcloud iam service-accounts get-iam-policy github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com \
  --project=pedeve-pertamina-dms
```

**Jika belum ada binding, tambahkan dengan command:**

```bash
gcloud iam service-accounts add-iam-policy-binding github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com \
  --project=pedeve-pertamina-dms \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/attribute.repository/repoareta/pedeve-dms-app"
```

**Atau via Console:**
1. Buka **IAM & Admin → Service Accounts**
2. Pilih service account: `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`
3. Klik tab **PERMISSIONS**
4. Klik **GRANT ACCESS**
5. Di **New principals**, masukkan:
   ```
   principalSet://iam.googleapis.com/projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/attribute.repository/repoareta/pedeve-dms-app
   ```
6. Di **Role**, pilih: `Service Account User` (roles/iam.workloadIdentityUser)
7. Klik **SAVE**

**Pastikan output menunjukkan binding seperti:**
```
bindings:
- members:
  - principalSet://iam.googleapis.com/projects/1076379007862/locations/global/workloadIdentityPools/github-actions-pool/attribute.repository/repoareta/pedeve-dms-app
  role: roles/iam.workloadIdentityUser
```

## 🔄 Setelah Update

Setelah memperbaiki konfigurasi WIF Provider, jalankan kembali GitHub Actions workflow. Error seharusnya teratasi.

**✅ Checklist Verifikasi:**
- [x] WIF Provider `allowedAudiences` = `https://github.com/repoareta/pedeve-dms-app` ✅
- [x] WIF Provider `issuerUri` = `https://token.actions.githubusercontent.com` ✅
- [x] Attribute mapping sudah benar (3 mapping) ✅
- [x] Attribute condition = `assertion.repository == "repoareta/pedeve-dms-app"` ✅
- [x] Service Account binding sudah benar ✅

**⚠️ Jika Error Masih Terjadi Setelah 15+ Menit:**

1. **Verifikasi Ulang WIF Provider (PENTING!):**
   ```bash
   # Cek detail lengkap WIF Provider
   gcloud iam workload-identity-pools providers describe github-actions-provider \
     --workload-identity-pool=github-actions-pool \
     --location=global \
     --project=pedeve-pertamina-dms \
     --format=json
   ```
   
   **Pastikan output menunjukkan:**
   - `oidc.allowedAudiences` HANYA berisi `["https://github.com/repoareta/pedeve-dms-app"]`
   - TIDAK ada field `oidc.defaultAudience` di output (atau jika ada, harus kosong/null)
   - TIDAK ada `iam.googleapis.com` di `allowedAudiences`
   - `state: "ACTIVE"`
   
   **Jika masih ada `defaultAudience` di output, berarti masih menggunakan "Default audience".**

2. **Hapus Default Audience (Jika Ada):**
   Jika masih ada default audience, update provider dengan hanya menggunakan allowed audiences:
   ```bash
   # Update provider dengan hanya allowed audiences (ini akan otomatis menghapus default audience)
   gcloud iam workload-identity-pools providers update-oidc github-actions-provider \
     --workload-identity-pool=github-actions-pool \
     --location=global \
     --project=pedeve-pertamina-dms \
     --allowed-audiences="https://github.com/repoareta/pedeve-dms-app"
   ```
   
   **Catatan:** Ketika Anda set `--allowed-audiences`, default audience akan otomatis di-nonaktifkan. Pastikan di GCP Console, radio button "Allowed audiences" dipilih (bukan "Default audience").

3. **Verifikasi di GCP Console:**
   - Buka **IAM & Admin → Workload Identity Federation**
   - Pilih pool: `github-actions-pool`
   - Pilih provider: `github-actions-provider`
   - Klik **EDIT**
   - Di tab **"Provider details"**, bagian **"Audiences"**:
     - **WAJIB:** Pilih **"Allowed audiences"** (bukan "Default audience")
     - **WAJIB:** Pastikan hanya ada 1 audience: `https://github.com/repoareta/pedeve-dms-app`
     - **WAJIB:** Hapus semua audience lain jika ada
   - Klik **SAVE**

4. **Clear GitHub Actions Cache:**
   - Cancel semua workflow yang sedang berjalan
   - Klik **"Re-run all jobs"** atau buat commit baru untuk trigger workflow baru
   - GitHub Actions mungkin masih menggunakan token lama

5. **Cek Logs di GCP untuk Detail Error:**
   ```bash
   # Cek audit logs untuk melihat detail error
   gcloud logging read "resource.type=workload_identity_pool_provider" \
     --project=pedeve-pertamina-dms \
     --limit=20 \
     --format=json \
     --freshness=1h
   ```

6. **Last Resort - Recreate Provider:**
   Jika semua langkah di atas tidak berhasil setelah 30+ menit, coba hapus dan buat ulang provider:
   ```bash
   # Hapus provider lama
   gcloud iam workload-identity-pools providers delete github-actions-provider \
     --workload-identity-pool=github-actions-pool \
     --location=global \
     --project=pedeve-pertamina-dms
   
   # Buat ulang provider dengan konfigurasi yang benar
   gcloud iam workload-identity-pools providers create-oidc github-actions-provider \
     --workload-identity-pool=github-actions-pool \
     --location=global \
     --project=pedeve-pertamina-dms \
     --issuer-uri=https://token.actions.githubusercontent.com \
     --allowed-audiences="https://github.com/repoareta/pedeve-dms-app" \
     --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
     --attribute-condition="assertion.repository == \"repoareta/pedeve-dms-app\""
   ```

