# 🔐 Setup GitHub Container Registry Token (GHCR_TOKEN)

## ⚠️ Penting: Keamanan Package

**JANGAN** membuat package visibility menjadi "Public" karena:
- Docker image bisa berisi informasi sensitif
- Kode aplikasi bisa di-expose
- Security risk tinggi

## ✅ Solusi Aman: Gunakan Personal Access Token (PAT)

### Langkah 1: Buat Personal Access Token

1. Buka GitHub → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Klik **Generate new token** → **Generate new token (classic)**
3. Isi:
   - **Note**: `GHCR Token for CI/CD`
   - **Expiration**: Sesuai kebutuhan (recommended: 90 days atau 1 year)
   - **Scopes**: Centang **hanya**:
     - ✅ `write:packages` - Untuk push Docker images
     - ✅ `read:packages` - Untuk pull Docker images
     - ✅ `delete:packages` - Optional, untuk delete old images
4. Klik **Generate token**
5. **COPY TOKEN SEKARANG** (hanya muncul sekali!)

### Langkah 2: Tambahkan Token ke GitHub Secrets

1. Buka repository → **Settings** → **Secrets and variables** → **Actions**
2. Klik **New repository secret**
3. Isi:
   - **Name**: `GHCR_TOKEN`
   - **Value**: Paste token yang sudah di-copy
4. Klik **Add secret**

### Langkah 3: Verifikasi Package Visibility

1. Buka repository → **Packages** (di sidebar kanan)
2. Klik package `dms-backend` atau `pedeve-dms-backend`
3. Pastikan visibility adalah **Private** (bukan Public)
4. Jika perlu, ubah ke Private:
   - Klik **Package settings**
   - Scroll ke **Danger Zone**
   - Klik **Change visibility** → Pilih **Private**

## 🔍 Verifikasi

Setelah setup, workflow akan:
- ✅ Menggunakan `GHCR_TOKEN` jika ada (lebih aman)
- ✅ Fallback ke `GITHUB_TOKEN` jika `GHCR_TOKEN` tidak ada
- ✅ Package tetap **Private** (aman)

## 📝 Notes

- **PAT lebih aman** karena scope terbatas (hanya packages)
- **GITHUB_TOKEN** memiliki scope lebih luas (bisa akses semua)
- Package visibility **harus Private** untuk keamanan
- Token akan expire sesuai setting, perlu regenerate jika expired

## 🔄 Rotate Token (Jika Perlu)

Jika token ter-compromise atau expired:
1. Generate token baru (ikuti Langkah 1)
2. Update secret `GHCR_TOKEN` dengan token baru
3. Delete token lama di GitHub Settings
