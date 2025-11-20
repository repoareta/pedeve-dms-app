# Security Features Documentation

Dokumentasi lengkap untuk semua security features yang telah diimplementasikan di DMS Backend.

## Daftar Security Features

### ✅ 1. JWT Authentication & Authorization
**Status:** ✅ FULLY IMPLEMENTED

**Lokasi File:**
- `backend/utils.go` - JWT generation & validation
- `backend/middleware.go` - JWTAuthMiddleware

**Penggunaan:**
```go
// Apply middleware to protected routes
r.Group(func(r chi.Router) {
    r.Use(JWTAuthMiddleware)
    r.Get("/auth/profile", GetProfile)
})
```

**Fitur:**
- Token expiry: 24 jam
- Secure token signing dengan HS256
- Bearer token authentication
- Context-based user info injection

---

### ✅ 2. Two-Factor Authentication (2FA)
**Status:** ✅ FULLY IMPLEMENTED (Standby)

**Lokasi File:**
- `backend/2fa.go` - 2FA implementation

**Fitur:**
- TOTP (Time-based One-Time Password) generation
- QR code generation untuk authenticator apps
- Backup codes untuk recovery
- API endpoints:
  - `POST /api/v1/auth/2fa/generate` - Generate 2FA secret
  - `POST /api/v1/auth/2fa/verify` - Verify and enable 2FA

**Database:**
- Table: `two_factor_auths`
- Fields: `user_id`, `secret`, `enabled`, `backup_codes`

**Penggunaan (untuk masa depan):**
```go
// Di login handler, setelah password verified:
if user2FA.Enabled {
    // Require 2FA code
    if !Verify2FALogin(userID, code) {
        return error
    }
}
```

---

### ✅ 3. Role-Based Access Control (RBAC)
**Status:** ✅ FULLY IMPLEMENTED (Standby)

**Lokasi File:**
- `backend/rbac.go` - RBAC implementation

**Roles yang tersedia:**
- `user` - Basic user permissions
- `admin` - Admin permissions
- `superadmin` - All permissions

**Permissions:**
- `user:read`, `user:write`, `user:delete`
- `document:read`, `document:write`, `document:delete`
- `admin:read`, `admin:write`, `admin:delete`

**Penggunaan:**
```go
// Middleware untuk check permission
r.Group(func(r chi.Router) {
    r.Use(JWTAuthMiddleware)
    r.Use(RequirePermission(PermissionDocumentDelete))
    r.Delete("/documents/{id}", deleteDocumentHandler)
})

// Middleware untuk check role
r.Group(func(r chi.Router) {
    r.Use(JWTAuthMiddleware)
    r.Use(RequireRole("admin", "superadmin"))
    r.Post("/admin/users", createUserHandler)
})
```

**Konfigurasi:**
Lihat `RolePermissions` map di `rbac.go` untuk mengubah permissions per role.

---

### ✅ 4. Rate Limiting
**Status:** ✅ FULLY IMPLEMENTED

**Lokasi File:**
- `backend/ratelimit.go` - Rate limiting implementation

**Rate Limiters:**
1. **General Rate Limiter** - 100 requests/second, burst 10
   - Applied globally di main.go
2. **Auth Rate Limiter** - 5 requests/minute, burst 2
   - Applied ke `/auth/login` dan `/auth/register`
3. **Strict Rate Limiter** - 10 requests/minute, burst 3
   - Dapat digunakan untuk sensitive endpoints

**Penggunaan:**
```go
// Apply rate limiting
r.Use(RateLimitMiddleware(generalRateLimiter))

// Atau untuk specific routes
r.Group(func(r chi.Router) {
    r.Use(AuthRateLimitMiddleware)
    r.Post("/auth/login", Login)
})
```

**Fitur:**
- IP-based rate limiting
- Automatic cleanup of old visitors
- Configurable rate limits per endpoint

---

### ✅ 5. Input Validation & Sanitization
**Status:** ✅ FULLY IMPLEMENTED

**Lokasi File:**
- `backend/validation.go` - Validation & sanitization

**Fitur:**
- Email validation & sanitization
- Username validation (alphanumeric + underscore, 3-50 chars)
- Password strength validation (min 8 chars, letter + number)
- HTML sanitization (mencegah XSS)
- SQL injection prevention (additional layer)

**Penggunaan:**
```go
// Validate register input
if err := ValidateRegisterInput(&req); err != nil {
    return error
}

// Sanitize string input
sanitized := SanitizeString(userInput)

// Sanitize email
email, err := SanitizeEmail(userEmail)
```

**Library yang digunakan:**
- `github.com/asaskevich/govalidator` - Email validation
- `github.com/microcosm-cc/bluemonday` - HTML sanitization

**Note:** GORM sudah menggunakan parameterized queries untuk SQL injection prevention.

---

### ✅ 6. SQL Injection Prevention
**Status:** ✅ FULLY IMPLEMENTED

**Lokasi:**
- Semua database queries menggunakan GORM
- GORM menggunakan parameterized queries secara otomatis

**Contoh:**
```go
// Safe - menggunakan parameterized query
DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&user)

// Tidak pernah menggunakan string concatenation untuk SQL
```

**Additional Layer:**
- `SanitizeSQLInput()` function di `validation.go` sebagai tambahan

---

### ✅ 7. Audit Logging
**Status:** ✅ FULLY IMPLEMENTED (Standby)

**Lokasi File:**
- `backend/audit.go` - Audit logging system

**Database:**
- Table: `audit_logs`
- Fields: `user_id`, `username`, `action`, `resource`, `resource_id`, `ip_address`, `user_agent`, `details`, `status`, `created_at`

**Action Types:**
- `login`, `logout`, `register`
- `create_user`, `update_user`, `delete_user`
- `create_document`, `update_document`, `delete_document`, `view_document`
- `enable_2fa`, `disable_2fa`
- `failed_login`, `password_reset`

**Penggunaan:**
```go
// Log action
LogAction(
    userID, 
    username, 
    ActionLogin, 
    ResourceAuth, 
    "", 
    ipAddress, 
    userAgent, 
    StatusSuccess, 
    nil,
)

// Get audit logs
logs, total, err := GetAuditLogs(userID, action, resource, status, limit, offset)
```

**Fitur:**
- Async logging (non-blocking)
- Detailed tracking (who, what, when, where)
- Filtering capabilities
- Pagination support

---

## Integrasi dengan Existing Code

### Login Handler
- ✅ Input validation & sanitization
- ✅ Rate limiting (AuthRateLimitMiddleware)
- ✅ Audit logging (success/failure)
- ✅ 2FA ready (dapat diintegrasikan)

### Register Handler
- ✅ Input validation & sanitization
- ✅ Rate limiting (AuthRateLimitMiddleware)

### Protected Routes
- ✅ JWT authentication
- ✅ Rate limiting (general)
- 🔄 RBAC ready (dapat diaktifkan dengan menambahkan middleware)

---

## Cara Mengaktifkan Fitur yang Standby

### Mengaktifkan 2FA di Login:
```go
// Di auth.go, setelah password verified:
var twoFA TwoFactorAuth
DB.Where("user_id = ? AND enabled = ?", userModel.ID, true).First(&twoFA)
if twoFA.ID != "" {
    // Require 2FA code dari request
    if !Verify2FALogin(userModel.ID, req.OTPCode) {
        return error
    }
}
```

### Mengaktifkan RBAC:
```go
// Tambahkan middleware ke routes yang perlu role check
r.Group(func(r chi.Router) {
    r.Use(JWTAuthMiddleware)
    r.Use(RequireRole("admin", "superadmin"))
    // routes here
})
```

### Mengaktifkan Audit Logging untuk specific actions:
```go
// Di handler function
LogAction(
    userID, 
    username, 
    ActionCreateDocument, 
    ResourceDocument, 
    documentID, 
    ipAddress, 
    userAgent, 
    StatusSuccess, 
    map[string]interface{}{
        "title": document.Title,
    },
)
```

---

## Environment Variables

### JWT
```bash
JWT_SECRET=your-very-secure-secret-key-min-32-chars-long
```

### Database
```bash
DATABASE_URL=postgres://user:pass@localhost/dbname
```

---

## Best Practices

1. **Selalu gunakan JWT middleware untuk protected routes**
2. **Gunakan rate limiting untuk prevent brute force**
3. **Validate dan sanitize semua user input**
4. **Log semua security-relevant actions**
5. **Gunakan RBAC untuk fine-grained access control**
6. **Enable 2FA untuk sensitive operations (admin/superadmin)**

---

## Testing Security Features

### Test Rate Limiting:
```bash
# Test auth rate limit (5 req/min)
for i in {1..10}; do 
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}'
done
```

### Test Input Validation:
```bash
# Test email validation
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"invalid-email","password":"123"}'
```

### Test RBAC:
```bash
# Test permission check (setelah implementasi)
curl -X DELETE http://localhost:8080/api/v1/documents/1 \
  -H "Authorization: Bearer <user-token>"
```

---

## Future Enhancements

- [ ] IP whitelisting/blacklisting
- [ ] Session management
- [ ] Password policy enforcement
- [ ] Account lockout after failed attempts
- [ ] Email verification
- [ ] Password reset with secure tokens
- [ ] API key authentication
- [ ] OAuth2 integration

