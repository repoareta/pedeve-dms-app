# 📝 Review & Koreksi Roadmap Production Setup

Dokumen ini berisi review dan koreksi penting yang sudah diterapkan ke roadmap production setup.

---

## ✅ Koreksi yang Sudah Diterapkan

### 1. Cloud SQL Flags - PostgreSQL ✅

**Masalah:** `--enable-bin-log` adalah flag MySQL, tidak valid untuk PostgreSQL.

**Solusi:**
- ❌ Hapus `--enable-bin-log`
- ✅ Gunakan `--enable-point-in-time-recovery` untuk PostgreSQL
- ✅ Enable automated backups via Console (recommended) atau CLI
- ✅ Dokumentasi sudah diupdate dengan catatan penting

---

### 2. Cloud SQL Security - Public IP vs Private IP ✅

**Masalah:** Untuk production, Public IP kurang aman.

**Solusi:**
- ✅ Tambahkan opsi: Public IP + Authorized Networks (quick, sama seperti dev)
- ✅ Tambahkan opsi: Private IP (recommended untuk production, lebih secure)
- ✅ Dokumentasi menjelaskan trade-off antara kecepatan vs security
- ✅ Untuk quick setup: Public IP + authorized networks hanya IP backend-prod

---

### 3. WIF Attribute Mapping & Condition ✅

**Masalah:** Format attribute mapping dan condition perlu lebih ketat untuk security.

**Solusi:**
- ✅ Update attribute mapping: tambahkan `attribute.ref=assertion.ref`
- ✅ Update attribute condition: restrict ke branch `main`:
  ```
  assertion.repository=='repoareta/pedeve-dms-app' && attribute.ref.startsWith('refs/heads/main')
  ```
- ✅ Format sudah benar: `attribute.<name>` (huruf kecil, angka, underscore)

---

### 4. WIF Audience ✅

**Masalah:** Jangan pakai URL repo sebagai audience (akan error mismatch).

**Solusi:**
- ✅ Dokumentasi menjelaskan: gunakan provider resource name (default)
- ✅ Format: `https://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/...`
- ✅ GitHub Actions `google-github-actions/auth` otomatis menggunakan provider resource name
- ✅ Catatan penting ditambahkan di roadmap

---

### 5. Service Account Separation ✅

**Masalah:** VM runtime dan GitHub Actions deployment harus pakai SA terpisah (blast radius lebih kecil).

**Solusi:**
- ✅ Tambahkan Fase 6: Setup VM Service Accounts
- ✅ Pisahkan:
  - `github-actions-deployer@...` → untuk deployment (SSH, compute ops)
  - `vm-backend-prod@...` → untuk runtime backend (Secret Manager, Storage, Cloud SQL)
  - `vm-frontend-prod@...` → untuk runtime frontend (jika perlu)
- ✅ IAM roles terpisah sesuai kebutuhan masing-masing
- ✅ VM creation sudah diupdate untuk menggunakan VM SA

---

### 6. Firewall Tags Consistency ✅

**Masalah:** Firewall rule `allow-http-prod` pakai tag `http-server`, tapi VM belum diberi tag itu.

**Solusi:**
- ✅ Update backend VM: tambahkan tag `http-server`
- ✅ Update frontend VM: tambahkan tag `http-server`
- ✅ Firewall rules sudah konsisten dengan VM tags
- ✅ Dokumentasi menjelaskan pentingnya konsistensi tags

---

### 7. Cloud SQL Proxy v2 Format ✅

**Masalah:** Format systemd service perlu lebih jelas dan menggunakan format v2 yang benar.

**Solusi:**
- ✅ Update format ke Cloud SQL Proxy v2:
  ```
  ExecStart=/usr/local/bin/cloud-sql-proxy \
    --address 127.0.0.1 \
    --port 5432 \
    pedeve-production:asia-southeast2:postgres-prod
  ```
- ✅ Lebih bersih, mengurangi kesalahan format
- ✅ Tidak pakai format v1 `tcp:127.0.0.1:5432`

---

### 8. Secrets Consistency ✅

**Masalah:** `db_host_prod` tidak perlu jika pakai proxy (host selalu 127.0.0.1).

**Solusi:**
- ✅ Hapus `db_host_prod` dari secrets (tidak diperlukan)
- ✅ Dokumentasi menjelaskan: host selalu 127.0.0.1 via proxy
- ✅ Environment variables sudah diupdate dengan catatan penting
- ✅ Opsi: simpan `DATABASE_URL` lengkap sebagai secret (alternatif)

---

### 9. GitHub Environments (Best Practice) ✅

**Masalah:** Secrets dev vs prod perlu dipisah dengan lebih baik.

**Solusi:**
- ✅ Tambahkan opsi: GitHub Environments (RECOMMENDED)
- ✅ Environment `production` untuk secrets production
- ✅ Environment `development` untuk secrets development
- ✅ Alternatif: Repository secrets dengan suffix `_PROD` (jika tidak pakai Environments)
- ✅ Dokumentasi menjelaskan keuntungan GitHub Environments

---

## 📋 Urutan Eksekusi yang Disarankan (Lebih Aman)

Berdasarkan review, urutan eksekusi yang lebih aman:

1. **Project Number + Enable APIs**
   - Get project number production
   - Enable required APIs
   - Pastikan policy "SA key creation disabled" (enforce WIF)

2. **WIF + GitHub Actions SA**
   - Setup Workload Identity Federation
   - Create GitHub Actions Service Account
   - Grant IAM roles untuk deployment

3. **VM Service Accounts (Runtime)**
   - Create VM backend SA
   - Create VM frontend SA
   - Grant IAM roles untuk runtime

4. **VM Creation**
   - Create backend-prod dengan VM SA
   - Create frontend-prod dengan VM SA
   - Setup OS Login

5. **Cloud SQL Production**
   - Create instance (via Console recommended)
   - Setup authorized networks atau Private IP
   - Create database & user

6. **Secrets + Permissions**
   - Create secrets di Secret Manager
   - Grant access ke VM SA (bukan GitHub SA)

7. **Storage Bucket + CORS**
   - Create bucket
   - Configure CORS
   - Grant access ke VM SA

8. **Cloud SQL Proxy**
   - Install di backend VM
   - Setup systemd service dengan format v2
   - Verify connection

9. **CI/CD Setup**
   - Setup GitHub Environment `production`
   - Update workflow untuk branch `main`
   - Test deployment

---

## 🔒 Security Best Practices yang Diterapkan

1. ✅ **Service Account Separation:** Deploy vs Runtime terpisah
2. ✅ **WIF Branch Restriction:** Hanya branch `main` bisa deploy ke production
3. ✅ **Cloud SQL Security:** Opsi Private IP atau minimal authorized networks ketat
4. ✅ **Secrets Management:** Tidak simpan informasi yang tidak diperlukan
5. ✅ **Firewall Tags:** Konsisten antara rules dan VM tags
6. ✅ **GitHub Environments:** Memisahkan secrets per environment

---

## 📝 Catatan Penting

1. **Cloud SQL:** Untuk PostgreSQL, gunakan Console atau pastikan flags valid
2. **WIF Audience:** Jangan pakai URL repo, gunakan provider resource name
3. **Service Accounts:** Pisahkan untuk security (blast radius lebih kecil)
4. **Secrets:** Host database selalu 127.0.0.1 via proxy, tidak perlu di secret
5. **Environments:** GitHub Environments lebih baik daripada suffix `_PROD`

---

**Last Updated:** 2025-01-27  
**Status:** ✅ All corrections applied to roadmap
