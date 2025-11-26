# 📚 Dokumentasi Backend - DMS App

Dokumentasi lengkap untuk tim Backend Development.

---

## 🎯 Quick Reference

### URLs & Ports
- **API Base URL**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Health Check**: `http://localhost:8080/health`
- **Port**: `8080` (configurable via `PORT` env)

### Tech Stack
- **Language**: Go 1.25
- **Framework**: Fiber v2.52.10
- **ORM**: GORM v1.31.1
- **Database**: SQLite (dev) / PostgreSQL (prod)
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Docs**: Swagger/OpenAPI

---

## 🏗️ Architecture

### Clean Architecture Structure
```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── domain/
│   │   └── models.go           # Domain models
│   ├── infrastructure/
│   │   ├── database/           # DB connection
│   │   ├── jwt/                # JWT utilities
│   │   ├── password/           # Password hashing
│   │   ├── cookie/             # Cookie management
│   │   ├── validation/         # Input validation
│   │   ├── audit/              # Audit logging
│   │   └── seed/               # Database seeding
│   ├── delivery/
│   │   └── http/               # HTTP handlers
│   ├── middleware/             # HTTP middleware
│   ├── repository/             # Data access layer
│   └── usecase/                # Business logic
├── go.mod
└── Dockerfile
```

### Key Principles
- **Separation of Concerns**: Domain, Infrastructure, Delivery terpisah
- **Dependency Injection**: Interfaces untuk testability
- **Clean Code**: Single responsibility, clear naming

---

## 🔐 Authentication & Security

### JWT Authentication
- **Token Expiry**: 24 jam
- **Algorithm**: HS256
- **Storage**: httpOnly Cookie (preferred) atau Authorization header
- **Secret**: Environment variable `JWT_SECRET` (min 32 chars)

### Security Features Implemented
1. ✅ **JWT Authentication** - Token-based dengan httpOnly cookies
2. ✅ **Password Hashing** - Bcrypt (cost: 10)
3. ✅ **2FA (TOTP)** - Ready, dapat diaktifkan
4. ✅ **RBAC** - Role-based access control (ready)
5. ✅ **Rate Limiting** - Auth (5/min), General (200/sec), Strict (20/min)
6. ✅ **CSRF Protection** - Double-submit cookie pattern
7. ✅ **Input Validation** - Email, username, password validation
8. ✅ **HTML Sanitization** - XSS prevention
9. ✅ **Security Headers** - CSP, XSS Protection, HSTS, dll
10. ✅ **Audit Logging** - User actions & technical errors
11. ✅ **CORS** - Configured untuk localhost:5173 dan localhost:3000

### Rate Limiting Configuration
```go
// Auth endpoints (login, register)
- Rate: 5 requests/minute
- Burst: 5

// General endpoints
- Rate: 200 requests/second
- Burst: 200

// Strict endpoints (sensitive operations)
- Rate: 20 requests/minute
- Burst: 20
```

---

## 🗄️ Database

### Development (SQLite)
- **File**: `backend/dms.db`
- **Driver**: `gorm.io/driver/sqlite`
- **Auto-created**: Ya, saat pertama kali run
- **CGO Required**: `CGO_ENABLED=1`

### Production (PostgreSQL)
- **Connection**: Via `DATABASE_URL` env variable
- **Format**: `postgres://user:pass@host:port/dbname?sslmode=require`
- **Driver**: `gorm.io/driver/postgres`
- **Connection Pool**:
  - Max Open: 25
  - Max Idle: 5
  - Max Lifetime: 5 menit
  - Max Idle Time: 10 menit

### Database Selection Logic
```go
// Jika DATABASE_URL tidak di-set → SQLite (dev)
// Jika DATABASE_URL di-set → PostgreSQL (prod)
```

### Key Tables
- `users` - User accounts
- `two_factor_auths` - 2FA secrets & backup codes
- `audit_logs` - Audit trail (user actions & errors)
- `companies` - Company/subsidiary data
- `directors` - Director information
- `shareholders` - Shareholder data
- `business_fields` - Business field categories

---

## 📡 API Endpoints

### Public Endpoints
```
GET  /                          → API info
GET  /health                    → Health check
GET  /api/v1                    → API version info
GET  /api/v1/csrf-token         → Get CSRF token
POST /api/v1/auth/login         → Login (rate limited: 5/min)
POST /api/v1/auth/register      → Register (rate limited: 5/min)
```

### Protected Endpoints (Require JWT)
```
GET    /api/v1/auth/profile           → Get user profile
POST   /api/v1/auth/logout            → Logout (requires CSRF token)

# 2FA Endpoints
POST   /api/v1/auth/2fa/generate       → Generate 2FA secret & QR
POST   /api/v1/auth/2fa/verify         → Verify & enable 2FA
GET    /api/v1/auth/2fa/status         → Get 2FA status
POST   /api/v1/auth/2fa/disable        → Disable 2FA

# Audit Logs
GET    /api/v1/audit-logs              → List audit logs (paginated)
GET    /api/v1/audit-logs/stats        → Audit log statistics

# Documents (example)
GET    /api/v1/documents               → List documents
GET    /api/v1/documents/{id}          → Get document by ID
POST   /api/v1/documents               → Create document
PUT    /api/v1/documents/{id}          → Update document
DELETE /api/v1/documents/{id}          → Delete document
```

### Authentication Methods
1. **httpOnly Cookie** (Preferred)
   - Cookie name: `auth_token`
   - Automatically sent by browser
   - More secure (XSS protection)

2. **Authorization Header** (Fallback)
   - Format: `Authorization: Bearer <token>`
   - For API clients, mobile apps

### CSRF Protection
- **Required for**: POST, PUT, DELETE, PATCH
- **Header**: `X-CSRF-Token: <token>`
- **Token Endpoint**: `GET /api/v1/csrf-token` (public)
- **Token Expiry**: 24 jam

---

## 🔧 Development Setup

### Prerequisites
- Go 1.25+
- Docker & Docker Compose (optional)
- SQLite (untuk dev) atau PostgreSQL (untuk prod)

### Local Development
```bash
cd backend
go mod download
go run ./cmd/api/main.go
```

### Environment Variables
```bash
# Server
PORT=8080                          # Server port (default: 8080)
ENV=development                    # development/production

# Database
DATABASE_URL=                      # Kosong = SQLite, diisi = PostgreSQL
CGO_ENABLED=1                     # 1 untuk SQLite, 0 untuk PostgreSQL

# JWT
JWT_SECRET=your-secret-key-min-32-chars

# Audit Log Retention (optional)
AUDIT_LOG_USER_ACTION_RETENTION_DAYS=90
AUDIT_LOG_TECHNICAL_ERROR_RETENTION_DAYS=30
```

### Docker Development
```bash
# Start dengan hot reload (jika menggunakan Air)
docker-compose -f docker-compose.dev.yml up backend

# Atau manual
docker-compose -f docker-compose.dev.yml up --build backend
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific test
go test ./internal/delivery/http -v
```

### Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

---

## 📝 Swagger Documentation

### Generate Swagger Docs
```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go

# Docs akan tersimpan di backend/docs/
```

### Access Swagger UI
- **URL**: `http://localhost:8080/swagger/index.html`
- **Auto-update**: Setelah regenerate swagger docs

### Swagger Annotations
```go
// @Summary Login user
// @Description Authenticate user dengan username/email dan password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/login [post]
```

---

## 🛡️ Security Best Practices

### 1. Input Validation
- ✅ Validate semua user input
- ✅ Sanitize HTML input (prevent XSS)
- ✅ Validate email format
- ✅ Validate username (alphanumeric, underscore, dash)
- ✅ Validate password strength (min 8 chars)

### 2. Authentication
- ✅ Use httpOnly cookies untuk JWT (XSS protection)
- ✅ Implement CSRF protection untuk state-changing methods
- ✅ Rate limit auth endpoints (prevent brute force)
- ✅ Never return password dalam response

### 3. Database
- ✅ Use parameterized queries (GORM handles this)
- ✅ Enable SSL/TLS untuk PostgreSQL di production
- ✅ Configure connection pooling
- ✅ Regular backups

### 4. Error Handling
- ✅ Don't expose internal errors ke client
- ✅ Log errors untuk debugging
- ✅ Return generic error messages untuk users

### 5. Headers
- ✅ Security headers (CSP, XSS Protection, HSTS)
- ✅ CORS configuration
- ✅ Content-Type validation

---

## 📊 Audit Logging

### Log Types
1. **User Actions** (`user_action`)
   - Login, Logout, Create, Update, Delete
   - 2FA Enable/Disable
   - Retention: 90 hari (default)

2. **Technical Errors** (`technical_error`)
   - System errors, Database errors
   - Validation errors, Panics
   - Retention: 30 hari (default)

### Usage
```go
import "github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/audit"

// Log user action
audit.LogAction(
    userID,
    username,
    "login",
    "auth",
    "",
    ipAddress,
    userAgent,
    "success",
    nil,
)
```

### Auto Cleanup
- Runs every 24 hours
- Deletes logs based on retention policy
- Configurable via environment variables

---

## 🔄 Middleware Stack

### Order (Important!)
```go
1. SecurityHeadersMiddleware    // Security headers
2. CORSMiddleware               // CORS handling
3. ErrorLoggerMiddleware        // Error logging
4. RateLimitMiddleware          // Rate limiting
5. JWTAuthMiddleware            // JWT authentication (protected routes)
6. CSRFMiddleware              // CSRF protection (state-changing methods)
7. RBACMiddleware              // Role-based access (optional)
```

### Custom Middleware
```go
// Example: Custom middleware
func CustomMiddleware(c *fiber.Ctx) error {
    // Your logic here
    return c.Next()
}
```

---

## 🧪 Testing Guidelines

### Unit Tests
- Test business logic di `usecase/`
- Test utilities di `infrastructure/`
- Mock dependencies dengan interfaces

### Integration Tests
- Test HTTP handlers di `delivery/http/`
- Test database operations di `repository/`
- Use test database (SQLite in-memory)

### Example Test
```go
func TestLoginHandler(t *testing.T) {
    // Setup test database
    // Setup test request
    // Execute handler
    // Assert response
}
```

---

## 🚀 Deployment

### Docker Build
```bash
# Build image
docker build -t dms-backend:latest -f backend/Dockerfile backend/

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e JWT_SECRET=... \
  dms-backend:latest
```

### Production Checklist
- [ ] Set `ENV=production`
- [ ] Set strong `JWT_SECRET` (min 32 chars)
- [ ] Configure `DATABASE_URL` dengan SSL (`sslmode=require`)
- [ ] Set `CGO_ENABLED=0` (jika tidak pakai SQLite)
- [ ] Configure connection pooling
- [ ] Setup automated backups
- [ ] Enable HTTPS/TLS
- [ ] Configure CORS untuk production domains
- [ ] Setup monitoring & logging
- [ ] Review rate limiting settings

---

## 📚 Additional Documentation

### Backend Docs (dalam `backend/doc/`)
- `AUTH.md` - Authentication documentation
- `SECURITY_FEATURES.md` - Security features detail
- `AUDIT_LOG_OPTIMIZATION.md` - Audit log optimization
- `REFACTORING_NOTES.md` - Clean Architecture notes
- `VAULT_SETUP.md` - HashiCorp Vault setup (jika digunakan)

### External Resources
- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [JWT Best Practices](https://jwt.io/introduction)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)

---

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Check port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Error
- Check `DATABASE_URL` format
- Verify database server running
- Check SSL/TLS settings untuk PostgreSQL

### JWT Token Invalid
- Check `JWT_SECRET` consistency
- Verify token expiry (24 hours)
- Check cookie/httpOnly settings

### CORS Issues
- Verify allowed origins di CORS config
- Check `withCredentials` di frontend
- Verify preflight requests (OPTIONS)

---

## 📞 Support & Resources

### Key Files to Know
- `cmd/api/main.go` - Application entry point
- `internal/delivery/http/` - HTTP handlers
- `internal/middleware/` - Middleware stack
- `internal/infrastructure/` - External dependencies
- `internal/usecase/` - Business logic

### Common Commands
```bash
# Run server
go run ./cmd/api/main.go

# Run tests
go test ./...

# Generate Swagger
swag init -g cmd/api/main.go

# Format code
go fmt ./...

# Build
go build -o bin/api ./cmd/api
```

---

**Last Updated**: 2025-01-XX
**Version**: 1.0.0

