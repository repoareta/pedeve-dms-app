# Frontend Browser Access Troubleshooting

## Status: Frontend sudah bisa diakses via IP

✅ **Curl berhasil:** `curl http://34.128.123.1` return HTML
❌ **Browser tidak bisa:** `http://34.128.123.1` atau `https://pedeve-dev.aretaamany.com`

## Kemungkinan Masalah

### 1. HTTPS Redirect Issue

**Masalah:** Browser otomatis redirect ke HTTPS, tapi belum ada SSL certificate.

**Solusi:**
- **Gunakan HTTP dulu:** `http://pedeve-dev.aretaamany.com` (bukan https)
- **Atau akses via IP:** `http://34.128.123.1`
- **Clear browser cache** dan coba lagi

**Cara force HTTP di browser:**
- Chrome: Ketik `http://` secara eksplisit di address bar
- Atau gunakan incognito mode untuk test

### 2. Browser Cache

**Clear cache:**
- Chrome: `Ctrl+Shift+Delete` (Windows) atau `Cmd+Shift+Delete` (Mac)
- Pilih "Cached images and files"
- Atau gunakan **Incognito/Private mode**

### 3. API URL Configuration

**Masalah:** Frontend masih menggunakan `localhost:8080` untuk API karena `VITE_API_URL` tidak diset saat build.

**Solusi:** Rebuild frontend dengan environment variable yang benar.

**Setelah deployment berikutnya:**
- Frontend akan di-build dengan `VITE_API_URL=https://api-pedeve-dev.aretaamany.com/api/v1`
- API calls akan otomatis ke backend yang benar

**Untuk test sekarang (temporary fix):**
1. Buka browser console (F12)
2. Check Network tab
3. Lihat apakah API calls ke `localhost:8080` (salah) atau `api-pedeve-dev.aretaamany.com` (benar)

### 4. CORS Issues

**Jika ada CORS error di console:**
- Backend perlu allow origin dari frontend domain
- Cek `CORS_ORIGIN` environment variable di backend

**Current config:**
- Backend `CORS_ORIGIN=https://pedeve-dev.aretaamany.com`
- Jika akses via IP, mungkin perlu tambah IP juga

### 5. DNS Propagation

**Cek DNS:**
```bash
# Cek apakah domain sudah resolve ke IP yang benar
dig pedeve-dev.aretaamany.com
nslookup pedeve-dev.aretaamany.com

# Harus return: 34.128.123.1
```

**Jika belum resolve:**
- Tunggu beberapa menit (DNS propagation)
- Atau flush DNS cache:
  - Windows: `ipconfig /flushdns`
  - Mac/Linux: `sudo dscacheutil -flushcache`

## Quick Test Steps

### Step 1: Test via IP (HTTP)
```
http://34.128.123.1
```

### Step 2: Test via Domain (HTTP)
```
http://pedeve-dev.aretaamany.com
```
**JANGAN pakai HTTPS** jika belum ada SSL certificate.

### Step 3: Check Browser Console
1. Buka browser (F12)
2. Check Console tab untuk errors
3. Check Network tab untuk failed requests

### Step 4: Check API Calls
1. Buka Network tab
2. Filter by "XHR" atau "Fetch"
3. Lihat apakah API calls ke:
   - ❌ `localhost:8080` (salah - perlu rebuild)
   - ✅ `api-pedeve-dev.aretaamany.com` (benar)

## Temporary Workaround

**Jika API URL masih salah (localhost):**

1. **Akses via IP dengan HTTP:**
   ```
   http://34.128.123.1
   ```

2. **Buka browser console dan override API URL (temporary):**
   ```javascript
   // Di browser console
   window.VITE_API_URL = 'https://api-pedeve-dev.aretaamany.com/api/v1'
   // Reload page
   location.reload()
   ```

**Note:** Ini hanya temporary. Setelah deployment berikutnya, API URL akan otomatis benar.

## Setup SSL Certificate (Untuk HTTPS)

**Opsi 1: Let's Encrypt dengan Certbot**
```bash
# SSH ke frontend VM
gcloud compute ssh frontend-dev --zone=asia-southeast2-a

# Install Certbot
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Generate certificate
sudo certbot --nginx -d pedeve-dev.aretaamany.com

# Auto-renewal sudah di-setup otomatis
```

**Opsi 2: GCP Load Balancer dengan Managed SSL**
- Lebih kompleks tapi lebih robust
- Cocok untuk production

## Checklist

- [ ] Akses via IP HTTP: `http://34.128.123.1` ✅ (sudah berhasil)
- [ ] Akses via domain HTTP: `http://pedeve-dev.aretaamany.com`
- [ ] Clear browser cache
- [ ] Check browser console untuk errors
- [ ] Check Network tab untuk API calls
- [ ] Rebuild frontend dengan VITE_API_URL yang benar (setelah deployment berikutnya)
- [ ] Setup SSL certificate untuk HTTPS (opsional)

