# Troubleshooting Guide

## Masalah Umum

### 1. 401 Unauthorized untuk `/auth/logout` atau `/auth/profile`

**Penyebab:**
- User belum login atau token expired
- Cookie tidak terkirim dengan benar
- CSRF token tidak valid (untuk POST request)

**Solusi:**
1. Pastikan user sudah login terlebih dahulu
2. Cek apakah cookie `auth_token` terkirim di browser (DevTools > Application > Cookies)
3. Untuk POST request (logout), pastikan CSRF token terkirim:
   - Ambil CSRF token dari `/api/v1/csrf-token`
   - Kirim di header `X-CSRF-Token`
4. Pastikan `credentials: 'include'` di Axios/fetch request

**Contoh request dengan CSRF token:**
```javascript
// 1. Ambil CSRF token
const csrfResponse = await fetch('http://localhost:8080/api/v1/csrf-token', {
  credentials: 'include'
});
const { csrf_token } = await csrfResponse.json();

// 2. Gunakan CSRF token untuk POST request
await fetch('http://localhost:8080/api/v1/auth/logout', {
  method: 'POST',
  credentials: 'include',
  headers: {
    'X-CSRF-Token': csrf_token
  }
});
```

### 2. 429 Too Many Requests

**Penyebab:**
- Terlalu banyak request dalam waktu singkat
- Rate limiter terlalu ketat (sudah diperbaiki untuk development)

**Solusi:**
1. Rate limiter sudah ditingkatkan untuk development:
   - General API: 500 req/s, burst 500
   - Auth endpoints: 5 req/min, burst 5
2. Jika masih terjadi, tunggu beberapa detik sebelum request berikutnya
3. Pastikan tidak ada infinite loop di frontend yang melakukan polling

**Rate Limiter Settings (Development):**
- General API: 500 requests/second, burst: 500
- Auth endpoints: 5 requests/minute, burst: 5
- Strict endpoints: 50 requests/minute, burst: 50

### 3. Cookie tidak terkirim

**Penyebab:**
- CORS configuration tidak mengizinkan credentials
- Cookie domain/path tidak sesuai
- Browser blocking third-party cookies

**Solusi:**
1. Pastikan CORS mengizinkan credentials:
   ```go
   AllowCredentials: true
   ```
2. Pastikan frontend mengirim credentials:
   ```javascript
   axios.defaults.withCredentials = true;
   // atau
   fetch(url, { credentials: 'include' })
   ```
3. Cek browser console untuk error CORS

### 4. CSRF Token Invalid

**Penyebab:**
- CSRF token expired (24 jam)
- CSRF token tidak terkirim di header
- CSRF token tidak valid

**Solusi:**
1. Ambil CSRF token baru dari `/api/v1/csrf-token`
2. Pastikan token terkirim di header `X-CSRF-Token` (bukan di body)
3. CSRF token valid selama 24 jam

## Testing di Local

### Test Login Flow:
```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"superadmin","password":"Pedeve123"}' \
  -c cookies.txt

# 2. Ambil CSRF token
curl -X GET http://localhost:8080/api/v1/csrf-token \
  -b cookies.txt \
  -c cookies.txt

# 3. Get Profile (GET tidak perlu CSRF)
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -b cookies.txt

# 4. Logout (POST perlu CSRF token)
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "X-CSRF-Token: <csrf_token_dari_step_2>" \
  -b cookies.txt
```

## Debug Tips

1. **Cek Logs**: Lihat log aplikasi untuk detail error
2. **Browser DevTools**: 
   - Network tab: cek request/response headers
   - Application tab: cek cookies
   - Console: cek error messages
3. **Postman/Insomnia**: Test API secara langsung tanpa frontend

