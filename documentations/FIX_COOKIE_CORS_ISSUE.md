# 🔧 Fix Cookie & CORS Issue Setelah Logout/Login

## Masalah
- Login pertama berhasil
- Setelah logout dan login lagi, error CORS
- Request menggunakan `credentials: "include"`

## Kemungkinan Penyebab

1. **Cookie tidak ter-set dengan benar setelah login kedua**
2. **CORS credentials tidak konsisten**
3. **Cookie domain/path tidak sesuai**

## Solusi: Verifikasi CORS dan Cookie Configuration

### Step 1: Cek CORS Configuration di Backend

**SSH ke backend VM:**

```bash
# Cek CORS_ORIGIN di container
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN

# Cek backend logs untuk CORS
sudo docker logs dms-backend-prod --tail 50 | grep -i cors
```

**Expected CORS_ORIGIN:**
```
CORS_ORIGIN=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com
```

### Step 2: Test CORS dari Browser Console

**Buka browser console (F12) di `https://pedeve-dev.aretaamany.com` dan jalankan:**

```javascript
// Test CORS preflight
fetch('https://api-pedeve-dev.aretaamany.com/api/v1/auth/profile', {
  method: 'GET',
  credentials: 'include',
  headers: {
    'Accept': 'application/json'
  }
})
.then(r => r.json())
.then(console.log)
.catch(console.error);
```

**Cek error detail di console.**

### Step 3: Cek Cookie di Browser

**Di browser console, cek cookie:**

```javascript
// Cek semua cookies
document.cookie

// Cek cookie untuk domain
console.log('Cookies:', document.cookie);
```

**Atau di DevTools:**
- Application tab → Cookies → `https://api-pedeve-dev.aretaamany.com`
- Cek apakah `auth_token` cookie ada
- Cek domain, path, Secure, SameSite attributes

### Step 4: Verifikasi Backend Cookie Settings

**Cek apakah backend set cookie dengan domain yang benar:**

```bash
# Cek backend logs saat login
sudo docker logs dms-backend-prod --tail 100 | grep -i "set.*cookie\|auth_token"
```

**Expected cookie attributes:**
- `HttpOnly: true`
- `Secure: true` (untuk HTTPS)
- `SameSite: Lax` atau `Strict`
- `Domain: .aretaamany.com` atau tidak set (untuk exact domain)

### Step 5: Test Manual Login Flow

**1. Clear cookies di browser:**
- DevTools → Application → Cookies → Clear all

**2. Login lagi dan cek:**
- Apakah cookie ter-set?
- Apakah request ke `/auth/profile` berhasil?

**3. Logout dan cek:**
- Apakah cookie ter-clear?
- Apakah request berikutnya masih menggunakan cookie lama?

## Troubleshooting

### Issue: Cookie Tidak Ter-Set

**Cek backend cookie configuration:**

```go
// Di backend, cookie harus di-set dengan:
c.Cookie(&fiber.Cookie{
    Name:     "auth_token",
    Value:    token,
    HTTPOnly: true,
    Secure:   true,  // true untuk HTTPS
    SameSite: "Lax",
    Path:     "/",
    // Domain tidak perlu di-set jika frontend dan backend di domain berbeda
    // Atau set ke ".aretaamany.com" untuk share cookie antar subdomain
})
```

### Issue: CORS Preflight Fails

**Cek apakah backend handle OPTIONS request dengan benar:**

```bash
# Test OPTIONS request
curl -X OPTIONS \
  -H "Origin: https://pedeve-dev.aretaamany.com" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  https://api-pedeve-dev.aretaamany.com/api/v1/auth/profile -v
```

**Expected response headers:**
```
Access-Control-Allow-Origin: https://pedeve-dev.aretaamany.com
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET,POST,PUT,DELETE,OPTIONS,PATCH
```

### Issue: Cookie Domain Mismatch

**Jika frontend di `pedeve-dev.aretaamany.com` dan backend di `api-pedeve-dev.aretaamany.com`:**

- Cookie harus di-set tanpa `Domain` attribute (default ke exact domain)
- Atau set `Domain: .aretaamany.com` untuk share cookie antar subdomain
- Pastikan `SameSite: Lax` atau `None` (jika `Secure: true`)

## Quick Fix: Restart Backend dengan CORS yang Benar

**Jika CORS_ORIGIN belum benar, restart container:**

```bash
# Cek CORS_ORIGIN
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN

# Jika belum benar, restart dengan script sebelumnya
```

