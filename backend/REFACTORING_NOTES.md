# Catatan Refactoring Struktur Folder

## Status: IN PROGRESS

Refactoring ini mengikuti struktur Clean Architecture yang umum digunakan di Go projects.

## Struktur Folder Baru

```
backend/
├── cmd/
│   └── api/              # Entry point untuk API server
│       └── main.go       # Main application entry point
├── internal/
│   ├── domain/           # Domain models dan entities
│   │   └── models.go
│   ├── infrastructure/   # External dependencies
│   │   ├── database/     # Database connection
│   │   ├── jwt/          # JWT utilities
│   │   ├── password/     # Password hashing
│   │   ├── uuid/         # UUID generation
│   │   ├── cookie/       # Cookie management
│   │   └── validation/   # Input validation
│   ├── delivery/         # Delivery layer (HTTP handlers)
│   │   └── http/
│   ├── middleware/       # HTTP middleware
│   │   ├── auth.go
│   │   ├── csrf.go
│   │   ├── security.go
│   │   ├── error.go
│   │   └── rate_limit.go
│   ├── config/           # Configuration
│   └── pkg/              # Shared packages
├── migrations/           # Database migrations
├── scripts/              # Utility scripts
└── docs/                 # Documentation
```

## File yang Sudah Dipindahkan

### ✅ Domain Models
- `models.go` → `internal/domain/models.go`

### ✅ Infrastructure
- `database.go` → `internal/infrastructure/database/database.go`
- `utils.go` → Split ke:
  - `internal/infrastructure/jwt/jwt.go`
  - `internal/infrastructure/password/password.go`
  - `internal/infrastructure/uuid/uuid.go`
- `cookies.go` → `internal/infrastructure/cookie/cookie.go`
- `validation.go` → `internal/infrastructure/validation/validation.go`

## File yang Perlu Dipindahkan (TODO)

### Handlers (internal/delivery/http/)
- `auth.go` → `internal/delivery/http/auth_handler.go`
- `2fa.go` → `internal/delivery/http/twofa_handler.go`
- `audit.go` → `internal/delivery/http/audit_handler.go`
- `audit_cleanup.go` → `internal/delivery/http/audit_cleanup_handler.go`

### Middleware (internal/middleware/)
- `middleware.go` → `internal/middleware/auth.go` & `internal/middleware/security.go`
- `csrf.go` → `internal/middleware/csrf.go`
- `error_logger.go` → `internal/middleware/error.go`
- `ratelimit.go` → `internal/middleware/rate_limit.go`
- `rbac.go` → `internal/middleware/rbac.go`

### Entry Point
- `main.go` → `cmd/api/main.go`

## Update Imports yang Diperlukan

Semua file perlu diupdate untuk menggunakan package paths baru:
- `github.com/Fajarriswandi/dms-app/backend/internal/domain`
- `github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/database`
- `github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/jwt`
- dll.

## Cara Melanjutkan

1. **Update imports di semua file handler dan middleware**
2. **Pindahkan handlers ke internal/delivery/http/**
3. **Pindahkan middleware ke internal/middleware/**
4. **Update cmd/api/main.go dengan semua dependencies**
5. **Test kompilasi setelah setiap perubahan**

## Catatan

- File lama masih ada di root untuk reference
- Hapus file lama setelah semua imports sudah diupdate
- Swagger docs perlu regenerate setelah refactoring selesai

