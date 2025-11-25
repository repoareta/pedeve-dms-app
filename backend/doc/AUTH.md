# Authentication & Security Documentation

## Overview

Backend menggunakan JWT (JSON Web Token) untuk authentication dengan security best practices.

## Security Features

### 1. JWT Authentication
- Token-based authentication
- Token expires dalam 24 jam
- Secure token signing dengan HS256

### 2. Password Security
- Password di-hash menggunakan bcrypt
- Default cost: 10 rounds
- Password tidak pernah dikembalikan dalam response

### 3. Security Headers
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: default-src 'self'`

### 4. CORS Configuration
- Allowed origins: `localhost:5173`, `localhost:3000`
- Allowed methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
- Credentials: enabled
- Max age: 300 seconds

## API Endpoints

### Public Endpoints

#### Register
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "admin",
  "email": "admin@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "admin",
    "email": "admin@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Protected Endpoints

Semua protected endpoints memerlukan JWT token di Authorization header.

#### Get Profile
```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

#### Documents API
Semua endpoints di `/api/v1/documents/*` memerlukan authentication.

## Using JWT Token

### 1. Get Token
Login atau register untuk mendapatkan token.

### 2. Use Token
Tambahkan token di Authorization header:
```http
Authorization: Bearer <your-token-here>
```

### 3. Token Expiration
Token expires dalam 24 jam. Login lagi untuk mendapatkan token baru.

## Example Usage

### Using curl

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"test123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}'

# Get Profile (with token)
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <your-token>"

# Get Documents (with token)
curl -X GET http://localhost:8080/api/v1/documents \
  -H "Authorization: Bearer <your-token>"
```

### Using Swagger UI

1. Buka http://localhost:8080/swagger/index.html
2. Klik "Authorize" button di atas
3. Masukkan: `Bearer <your-token>`
4. Klik "Authorize"
5. Sekarang semua protected endpoints bisa di-test

## Environment Variables

### JWT_SECRET
Secret key untuk signing JWT tokens.

**Default:** `your-secret-key-change-in-production-min-32-chars`

**Production:** Set environment variable dengan secret yang kuat (minimal 32 karakter).

```bash
export JWT_SECRET="your-very-secure-secret-key-min-32-chars-long"
```

## Security Best Practices

1. **Always use HTTPS in production**
2. **Change JWT_SECRET** dari default value
3. **Use strong passwords** (minimal 8 karakter, kombinasi huruf, angka, simbol)
4. **Store tokens securely** di frontend (httpOnly cookies atau secure storage)
5. **Implement token refresh** untuk better UX
6. **Rate limiting** untuk prevent brute force attacks
7. **Input validation** untuk semua user input

## Current Implementation Notes

- **User Storage:** In-memory (data hilang saat restart)
- **Production:** Replace dengan database (PostgreSQL, MySQL, dll)
- **Token Refresh:** Belum diimplementasikan (bisa ditambahkan)
- **Rate Limiting:** Belum diimplementasikan (recommended)

## Next Steps

1. Add database integration (PostgreSQL/MySQL)
2. Implement token refresh mechanism
3. Add rate limiting
4. Add email verification
5. Add password reset functionality
6. Add role-based access control (RBAC)

