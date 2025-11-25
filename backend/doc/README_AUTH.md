# Authentication - Quick Start

## Default Admin Account

**Tidak ada akun default!** Anda perlu membuat akun pertama kali melalui form register di frontend.

## Cara Membuat Akun

### Option 1: Via Frontend (Recommended)

1. Buka http://localhost:5173
2. Klik "Register" atau akses http://localhost:5173/register
3. Isi form:
   - Username: `admin` (atau username pilihan Anda)
   - Email: `admin@example.com`
   - Password: `admin123` (atau password kuat)
   - Confirm Password: (ulangi password)
4. Klik "Register"
5. Otomatis login dan redirect ke home

### Option 2: Via Swagger UI

1. Buka http://localhost:8080/swagger/index.html
2. Cari endpoint `POST /api/v1/auth/register`
3. Klik "Try it out"
4. Isi request body:
```json
{
  "username": "admin",
  "email": "admin@example.com",
  "password": "admin123"
}
```
5. Klik "Execute"
6. Copy token dari response
7. Klik "Authorize" di atas, masukkan: `Bearer <token>`

### Option 3: Via curl

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

## Login

Setelah register, gunakan credentials yang sama untuk login:

**Via Frontend:**
- Buka http://localhost:5173/login
- Masukkan username dan password
- Klik "Login"

**Via Swagger:**
- Endpoint: `POST /api/v1/auth/login`
- Request body sama seperti register

**Via curl:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

## Database Location

- **Development**: SQLite database di `backend/dms.db`
- **Production**: PostgreSQL (set via `DATABASE_URL` env var)

## Reset Database

Jika ingin reset database:

```bash
# Hapus SQLite database
rm backend/dms.db

# Restart backend (akan auto-create database baru)
docker-compose -f docker-compose.dev.yml restart backend
```

## Tips

1. **Gunakan password kuat** untuk production
2. **Jangan commit** `dms.db` ke git (sudah di .gitignore)
3. **Backup database** sebelum deploy ke production
4. **Set JWT_SECRET** environment variable untuk production

