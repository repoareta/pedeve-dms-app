# 🔧 Fix Upload Logo URL Error

## Masalah

Saat upload logo di aplikasi development di GCP, terjadi error:

```
fetch("https://api-pedeve-dev.aretaamany.comhttps//storage.googleapis.com/pedeve-dev-bucket/logos/...")
```

**Error:** URL double dengan format salah (`https://api-pedeve-dev.aretaamany.comhttps//storage.googleapis.com/...`)

## Penyebab

1. **Backend mengembalikan full URL** dari GCP Storage: `https://storage.googleapis.com/pedeve-dev-bucket/logos/...`
2. **Frontend menambahkan `baseURL`** ke semua response URL tanpa cek apakah sudah full URL
3. **Hasil:** URL double dan format salah

## Solusi

**Fix di `SubsidiaryFormView.vue` - `handleLogoUpload` function:**

```typescript
// SEBELUM (SALAH):
const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const baseURL = apiURL.replace(/\/api\/v1$/, '')
logoFileList.value = [{
  url: `${baseURL}${response.url}`, // ❌ Selalu tambahkan baseURL
}]

// SESUDAH (BENAR):
let logoUrl: string
if (response.url.startsWith('http://') || response.url.startsWith('https://')) {
  // Full URL dari GCP Storage, langsung pakai
  logoUrl = response.url
} else {
  // Relative URL dari local storage, tambahkan baseURL
  const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const baseURL = apiURL.replace(/\/api\/v1$/, '')
  logoUrl = `${baseURL}${response.url}`
}
logoFileList.value = [{
  url: logoUrl, // ✅ Cek dulu apakah sudah full URL
}]
```

## Penjelasan

**Backend Storage Behavior:**

1. **GCP Storage (Production):**
   - Mengembalikan full URL: `https://storage.googleapis.com/bucket/path/file.png`
   - Frontend harus langsung pakai tanpa modifikasi

2. **Local Storage (Development):**
   - Mengembalikan relative URL: `/logos/file.png`
   - Frontend harus tambahkan `baseURL` untuk jadi full URL

**Fix Logic:**

- **Cek apakah URL dimulai dengan `http://` atau `https://`**
  - ✅ Jika ya → Full URL, langsung pakai
  - ❌ Jika tidak → Relative URL, tambahkan `baseURL`

## File yang Sudah Benar

Fungsi `getCompanyLogo` di file berikut sudah benar (sudah cek `startsWith('http')`):

- ✅ `SubsidiariesView.vue` - `getCompanyLogo`
- ✅ `SubsidiaryDetailView.vue` - `getCompanyLogo`
- ✅ `SubsidiaryFormView.vue` - `loadCompanyData` (untuk display logo existing)

## Testing

**Test di Development (Local Storage):**
1. Upload logo → Harus bisa tampil dengan URL: `http://localhost:8080/logos/...`

**Test di Production (GCP Storage):**
1. Upload logo → Harus bisa tampil dengan URL: `https://storage.googleapis.com/pedeve-dev-bucket/logos/...`
2. Logo harus bisa diakses langsung dari browser

## Status

✅ **Fixed** - Frontend sekarang cek apakah URL sudah full URL sebelum menambahkan `baseURL`

**Date:** 2025-11-28

