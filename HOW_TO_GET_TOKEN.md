# Cara Mendapatkan Token Autentikasi

Untuk menggunakan curl commands, Anda perlu mendapatkan JWT token terlebih dahulu melalui login.

## Cara 1: Menggunakan Script Otomatis (Paling Mudah)

Jalankan script yang sudah dibuat:

```bash
# Edit file terlebih dahulu untuk set username dan password
nano get-token-and-send-data.sh

# Atau langsung edit di editor favorit Anda
# Ganti USERNAME dan PASSWORD di dalam file

# Jalankan script
./get-token-and-send-data.sh
```

Script ini akan:
1. Login otomatis dengan username/password yang Anda set
2. Mendapatkan token
3. Mengirim data November, Desember, dan RKAP secara otomatis

## Cara 2: Login Manual dan Copy Token

### Step 1: Login untuk mendapatkan token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
```

**Response akan seperti ini:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "...",
    "username": "admin",
    "email": "admin@example.com",
    "role": "superadmin"
  }
}
```

### Step 2: Copy token dari response

Copy nilai dari field `"token"` (yang panjang, dimulai dengan `eyJ...`)

### Step 3: Dapatkan CSRF Token

Untuk request POST/PUT/DELETE, Anda juga perlu CSRF token:

```bash
curl -X GET http://localhost:8080/api/v1/csrf-token \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Response akan berisi:
```json
{
  "csrf_token": "abc123..."
}
```

### Step 4: Gunakan token dan CSRF token di curl command

Ganti `YOUR_AUTH_TOKEN` dan `YOUR_CSRF_TOKEN` dengan token yang Anda dapatkan:

```bash
curl -X POST http://localhost:8080/api/v1/financial-reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-CSRF-Token: abc123..." \
  -d '{...}'
```

## Cara 3: Menggunakan Script untuk Extract Token dan CSRF Token

Buat script sederhana untuk mendapatkan token dan CSRF token:

```bash
#!/bin/bash
USERNAME="admin"
PASSWORD="admin"

# Login dan extract token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"${USERNAME}\", \"password\": \"${PASSWORD}\"}" \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo "Token: $TOKEN"

# Dapatkan CSRF token
CSRF_TOKEN=$(curl -s -X GET http://localhost:8080/api/v1/csrf-token \
  -H "Authorization: Bearer ${TOKEN}" \
  | grep -o '"csrf_token":"[^"]*' | cut -d'"' -f4)

echo "CSRF Token: $CSRF_TOKEN"

# Gunakan token dan CSRF token untuk request berikutnya
curl -X POST http://localhost:8080/api/v1/financial-reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -d '{...}'
```

## Catatan Penting

1. **CSRF Token Required**: Semua request POST, PUT, DELETE, PATCH memerlukan CSRF token dalam header `X-CSRF-Token`. Dapatkan CSRF token dari endpoint `/api/v1/csrf-token`.

2. **Token Expiry**: Token JWT memiliki waktu kadaluarsa. Jika token expired, Anda perlu login lagi untuk mendapatkan token baru. CSRF token juga memiliki expiry (24 jam).

2. **Default Credentials**: 
   - Username default: `admin`
   - Password default: `admin` (atau sesuai dengan yang di-set saat seeding)

3. **2FA**: Jika user memiliki 2FA aktif, Anda perlu mengirim kode 2FA pada request login kedua:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "username": "admin",
       "password": "admin",
       "code": "123456"
     }'
   ```

4. **Cookie vs Header**: 
   - API juga mendukung token dalam httpOnly cookie (lebih aman)
   - Untuk curl, gunakan `Authorization: Bearer <token>` header
   - Untuk browser/frontend, token otomatis disimpan dalam cookie

## Troubleshooting

### Error: "Invalid or expired token"
- Token sudah expired, login lagi untuk mendapatkan token baru
- Pastikan format header benar: `Authorization: Bearer <token>` (dengan spasi setelah "Bearer")

### Error: "Invalid credentials"
- Username/password salah
- Pastikan user sudah terdaftar di database
- Cek apakah user aktif (is_active = true)

### Error: "2FA verification required"
- User memiliki 2FA aktif
- Kirim request login kedua dengan field `code` berisi kode 2FA
