# Spesifikasi Teknis DMS App

## 📋 Daftar Isi
1. [Backend (BE)](#backend-be)
2. [Frontend (FE)](#frontend-fe)
3. [Database](#database)
4. [Infrastructure & DevOps](#infrastructure--devops)
5. [Security Features](#security-features)
6. [API Endpoints](#api-endpoints)
7. [Development Environment](#development-environment)
8. [Production Environment](#production-environment)

---

## Backend (BE)

### Technology Stack
- **Language**: Go (Golang) 1.25
- **Framework/Router**: Fiber v2.52.10 (High-performance HTTP framework dengan fasthttp)
- **ORM**: GORM v1.31.1
- **Database Drivers**:
  - SQLite (Development): `gorm.io/driver/sqlite v1.6.0`
  - PostgreSQL (Production): `gorm.io/driver/postgres v1.6.0`
- **Authentication**: JWT (JSON Web Token) - `github.com/golang-jwt/jwt/v5 v5.3.0`
- **Password Hashing**: Bcrypt via `golang.org/x/crypto v0.44.0`
- **API Documentation**: Swagger/OpenAPI - `github.com/gofiber/swagger v1.1.1` + `github.com/swaggo/swag v1.16.6`
- **CORS**: `github.com/gofiber/cors v0.2.2`
- **UUID Generation**: `github.com/google/uuid v1.6.0`

### Security Libraries
- **Input Validation**: `github.com/asaskevich/govalidator`
- **HTML Sanitization**: `github.com/microcosm-cc/bluemonday v1.0.27`
- **Rate Limiting**: `golang.org/x/time v0.14.0`
- **2FA (TOTP)**: `github.com/pquerna/otp v1.5.0`
- **Barcode/QR Code**: `github.com/boombuler/barcode`

### Server Configuration
- **Port**: 8080 (default, configurable via `PORT` env)
- **Base Path**: `/api/v1`
- **Swagger UI**: `/swagger/index.html`
- **Health Check**: `/health`

### Key Features
- ✅ JWT Authentication dengan httpOnly Cookies
- ✅ Role-Based Access Control (RBAC)
- ✅ Two-Factor Authentication (2FA/TOTP)
- ✅ Rate Limiting (Auth, General, Strict)
- ✅ CSRF Protection (Double-submit cookie pattern)
- ✅ Input Validation & Sanitization
- ✅ Audit Logging (User actions + Technical errors)
- ✅ Audit Log Optimization (Retention policy, auto-cleanup)
- ✅ Security Headers
- ✅ CORS Configuration
- ✅ Swagger/OpenAPI Documentation
- ✅ Clean Architecture (cmd/, internal/)

### File Structure (Clean Architecture)
```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── domain/
│   │   └── models.go           # Domain models (User, Document, AuditLog, etc.)
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── database.go     # Database connection & initialization
│   │   ├── jwt/
│   │   │   └── jwt.go          # JWT generation & validation
│   │   ├── password/
│   │   │   └── password.go     # Password hashing utilities
│   │   ├── uuid/
│   │   │   └── uuid.go         # UUID generation
│   │   ├── cookie/
│   │   │   └── cookie.go        # Secure cookie management
│   │   ├── validation/
│   │   │   └── validation.go   # Input validation & sanitization
│   │   ├── audit/
│   │   │   └── audit.go        # Audit logger initialization
│   │   └── seed/
│   │       └── seed.go          # Database seeding
│   ├── delivery/
│   │   └── http/
│   │       ├── auth_handler.go      # Authentication handlers (Login, Profile, Logout)
│   │       ├── twofa_handler.go     # 2FA handlers
│   │       ├── audit_handler.go     # Audit log handlers
│   │       ├── document_handler.go   # Document handlers
│   │       └── csrf_handler.go      # CSRF token handler
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   ├── security.go          # Security headers middleware
│   │   ├── csrf.go              # CSRF protection middleware
│   │   ├── error.go             # Error logging middleware
│   │   ├── rate_limit.go        # Rate limiting middleware
│   │   └── rbac.go              # RBAC middleware
│   ├── repository/
│   │   └── audit_repository.go  # Audit log repository
│   └── usecase/
│       ├── twofa_usecase.go         # 2FA business logic
│       └── audit_cleanup_usecase.go # Audit cleanup logic
├── go.mod                 # Go dependencies
├── Dockerfile             # Production Docker image
└── docs/                  # Swagger generated docs
```

---

## Frontend (FE)

### Technology Stack
- **Framework**: Vue 3.5.22
- **Language**: TypeScript 5.9.0
- **Build Tool**: Vite 7.1.11
- **State Management**: Pinia 3.0.3
- **Routing**: Vue Router 4.6.3
- **UI Framework**: Ant Design Vue 4.2.6
- **HTTP Client**: Axios 1.13.2
- **Charts**: Chart.js 4.5.1 + vue-chartjs 5.3.3
- **Icons**: Iconify Vue 5.0.0
- **Styling**: SCSS (sass-embedded 1.93.3)

### Development Tools
- **Linter**: ESLint 9.37.0
- **Type Checking**: vue-tsc 3.1.1
- **Testing**: Vitest 3.2.4
- **Node Version**: ^20.19.0 || >=22.12.0

### Key Features
- ✅ JWT Token Management
- ✅ Protected Routes (Route Guards)
- ✅ Responsive Design (Desktop, Tablet, Mobile)
- ✅ Hot Module Replacement (HMR)
- ✅ TypeScript Support
- ✅ Component-based Architecture
- ✅ Global SCSS Styling
- ✅ Theme Configuration (Primary: #035CAB, Secondary: #DB241B)

### File Structure
```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts      # Axios configuration, interceptors
│   │   └── auth.ts        # Auth API functions
│   ├── assets/
│   │   ├── global.scss    # Global SCSS styles
│   │   └── main.css       # Main CSS
│   ├── components/
│   │   ├── DashboardHeader.vue
│   │   ├── KPICard.vue
│   │   ├── RevenueChart.vue
│   │   └── SubsidiariesList.vue
│   ├── stores/
│   │   └── auth.ts        # Pinia auth store
│   ├── views/
│   │   ├── LoginView.vue
│   │   ├── RegisterView.vue
│   │   ├── DashboardView.vue
│   │   └── NotFoundView.vue
│   ├── router/
│   │   └── index.ts      # Vue Router configuration
│   ├── App.vue
│   └── main.ts
├── public/
│   ├── logo.png
│   └── imgLogin.png
├── package.json
├── Dockerfile
└── vite.config.ts
```

### Port Configuration
- **Development**: 5173
- **Production**: 80 (via Nginx)

---

## Database

### Development Database
- **Type**: SQLite 3
- **File Location**: `backend/dms.db`
- **Driver**: `gorm.io/driver/sqlite`
- **Connection**: File-based (no server required)
- **CGO Required**: Yes (CGO_ENABLED=1)
- **Alasan Penggunaan**: 
  - ✅ Tidak perlu install database server terpisah (zero-configuration)
  - ✅ File-based, mudah untuk development lokal
  - ✅ Cepat untuk testing dan prototyping
  - ✅ Database dibuat otomatis saat aplikasi pertama kali dijalankan
  - ⚠️ **BUKAN untuk mobile apps** - ini untuk development environment server-side

**Catatan Penting**: SQLite memang populer digunakan di mobile apps (iOS, Android), tapi dalam proyek ini kita menggunakan SQLite **hanya untuk development environment** karena kemudahannya. SQLite cocok untuk development karena tidak perlu setup PostgreSQL server terpisah.

### Production Database
- **Type**: PostgreSQL
- **Driver**: `gorm.io/driver/postgres`
- **Connection**: Via `DATABASE_URL` environment variable
- **Format**: `postgres://user:password@host:port/dbname`
- **Alasan Penggunaan**:
  - ✅ Multi-user support (banyak concurrent connections)
  - ✅ ACID compliance untuk transaksi kompleks
  - ✅ Scalability untuk production workload
  - ✅ Advanced features (full-text search, JSON support, dll)
  - ✅ Replication dan backup yang lebih baik

### Database Selection Logic

```go
// File: backend/internal/infrastructure/database/database.go
// Use SQLite for development if DATABASE_URL not set
if dbURL == "" {
    log.Println("Using SQLite database (development)")
    dialector = sqlite.Open("dms.db")
} else {
    log.Println("Using PostgreSQL database")
    dialector = postgres.Open(dbURL)
}
```

**Cara Kerja**:
- Jika `DATABASE_URL` environment variable **TIDAK** di-set → Gunakan **SQLite** (development)
- Jika `DATABASE_URL` environment variable **DI-SET** → Gunakan **PostgreSQL** (production)

### Database Schema

#### Users Table
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    role TEXT DEFAULT 'user',
    password TEXT NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);
```

#### Two Factor Auth Table
```sql
CREATE TABLE two_factor_auths (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    secret TEXT NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    backup_codes TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

#### Audit Logs Table
```sql
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    username TEXT,
    action TEXT NOT NULL,
    resource TEXT,
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    details TEXT,  -- JSON string untuk detail tambahan
    status TEXT NOT NULL,  -- success, failure, error
    log_type TEXT DEFAULT 'user_action',  -- user_action atau technical_error
    created_at DATETIME
);
-- Indexes untuk performa
CREATE INDEX idx_audit_logs_created_at_log_type ON audit_logs(created_at, log_type);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_log_type ON audit_logs(log_type);
```

### Database Management
- **ORM**: GORM (Auto-migration enabled)
- **Connection Pool**: Dikonfigurasi eksplisit untuk PostgreSQL
  - Max Open Connections: 25
  - Max Idle Connections: 5
  - Connection Max Lifetime: 5 menit
  - Connection Max Idle Time: 10 menit
- **SSL/TLS**: 
  - Development: `sslmode=disable` (aman di local network)
  - Production: `sslmode=require` atau `sslmode=verify-full` (wajib)
- **Backup**: Automated backup script tersedia di `backend/scripts/backup-db.sh`
  - Manual backup: `./scripts/backup-db.sh`
  - Automated backup: Setup cron job (contoh: setiap hari jam 2 pagi)
  - Retention: 7 hari (configurable)
- **Query Builder**: GORM Query Builder
- **Migrations**: Auto-migrate on startup

### DBeaver Connection (SQLite)
- **Database Type**: SQLite
- **Path**: `/path/to/dms-app/backend/dms.db`
- **JDBC URL**: `jdbc:sqlite:/path/to/dms-app/backend/dms.db`
- **Driver**: SQLite JDBC Driver

---

## Infrastructure & DevOps

### Docker
- **Backend Image**: `golang:1.25-alpine` (dev), `alpine:latest` (prod)
- **Frontend Image**: `node:20-alpine` (build), `nginx:alpine` (prod)
- **Multi-stage Build**: Yes (both frontend & backend)

### Docker Compose
- **Development File**: `docker-compose.dev.yml`
- **Production File**: `docker-compose.yml`
- **Network**: `dms-network` (bridge)

### CI/CD Pipeline
- **Platform**: GitHub Actions
- **Workflow File**: `.github/workflows/ci-cd.yml`
- **Triggers**:
  - Push to `main` branch
  - Push version tags (`v*.*.*`)
  - Manual workflow dispatch

### CI/CD Features
- ✅ Lint & Test (Frontend & Backend)
- ✅ Security Scanning (Trivy)
- ✅ Docker Image Build
- ✅ Push to GitHub Container Registry (GHCR)
- ✅ Automatic Version Tagging
- ✅ Changelog Generation
- ✅ GitHub Release Creation (on tag push)
- ✅ SARIF Upload for Security Results

### Container Registry
- **Registry**: GitHub Container Registry (GHCR)
- **Images**:
  - `ghcr.io/fajarriswandi/dms-frontend:latest`
  - `ghcr.io/fajarriswandi/dms-backend:latest`
- **Version Tags**: Auto-generated from Git tags or commit SHA

### Build Tools
- **Go Version**: 1.25
- **Node Version**: 20
- **Build System**: Docker Buildx

---

## Security Features

### Authentication & Authorization
1. **JWT Authentication**
   - Token expiration: 24 hours
   - Algorithm: HS256
   - Secret: Environment variable or default

2. **Password Security**
   - Hashing: Bcrypt (cost: 10)
   - Never returned in API responses

3. **Role-Based Access Control (RBAC)**
   - Roles: `superadmin`, `admin`, `user`
   - Permissions: Read, Write, Delete, Admin
   - Middleware: `RBACMiddleware` (ready, commented out)

4. **Two-Factor Authentication (2FA)**
   - Method: TOTP (Time-based One-Time Password)
   - QR Code generation
   - Backup codes support
   - Library: `github.com/pquerna/otp`

### Rate Limiting
1. **Auth Rate Limiter**
   - Rate: 5 requests per minute per IP
   - Burst: 5
   - Applied to: `/auth/login` (public endpoint)

2. **General Rate Limiter**
   - Rate: 200 requests per second per IP
   - Burst: 200
   - Applied to: All routes (untuk development, bisa dikurangi di production)

3. **Strict Rate Limiter**
   - Rate: 20 requests per minute per IP
   - Burst: 20
   - Ready for sensitive endpoints

### Input Validation & Sanitization
1. **Validation**
   - Email format validation
   - Username validation (alphanumeric, underscore, dash)
   - Password strength validation (min 8 chars)
   - Library: `github.com/asaskevich/govalidator`

2. **Sanitization**
   - HTML sanitization (XSS prevention)
   - Library: `github.com/microcosm-cc/bluemonday`

### Security Headers
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY` (SAMEORIGIN for Swagger)
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy: default-src 'self'` (relaxed for Swagger)

### CSRF Protection
- **Method**: Double-submit cookie pattern
- **Token Header**: `X-CSRF-Token`
- **Cookie**: `csrf_token` (httpOnly, Secure, SameSite=Strict)
- **Applied to**: State-changing methods (POST, PUT, DELETE, PATCH)
- **Excluded**: GET, HEAD, OPTIONS methods

### CORS Configuration
- **Allowed Origins**: `http://localhost:5173`, `http://localhost:3000`
- **Allowed Methods**: GET, POST, PUT, DELETE, OPTIONS, PATCH
- **Allowed Headers**: Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With
- **Credentials**: Enabled
- **Max Age**: 300 seconds

### Audit Logging
- **Actions Tracked**: 
  - User Actions: Login Success, Login Failed, Logout, Create, Update, Delete, 2FA Enable/Disable
  - Technical Errors: System errors, Database errors, Validation errors, Panics
- **Information Logged**: User ID, Username, Action, Resource, Resource ID, IP Address, User Agent, Status, Details (JSON), Log Type, Timestamp
- **Storage**: Database table `audit_logs`
- **Optimization**:
  - Retention Policy: User actions (90 days default), Technical errors (30 days default)
  - Auto-cleanup: Every 24 hours
  - Indexing: Composite indexes on `created_at` + `log_type` for performance

---

## API Endpoints

### Base URL
- **Development**: `http://localhost:8080/api/v1`
- **Production**: `https://your-domain.com/api/v1`

### Public Endpoints

#### 1. Root
- **GET** `/`
- **Description**: API information
- **Response**: `{"message": "DMS Backend API", "version": "1.0.0", "swagger": "/swagger/index.html"}`

#### 2. Health Check
- **GET** `/health`
- **Description**: Health status
- **Response**: `{"status": "OK", "service": "dms-backend"}`

#### 3. API Info
- **GET** `/api/v1`
- **Description**: API version and endpoints
- **Response**: JSON with API info

#### 4. CSRF Token
- **GET** `/api/v1/csrf-token`
- **Description**: Get CSRF token untuk form submissions
- **Response**: `{"csrf_token": "string"}`
- **Note**: Public endpoint, tidak memerlukan authentication

#### 5. Login
- **POST** `/api/v1/auth/login`
- **Description**: Authenticate user (username or email). Mendukung 2FA jika diaktifkan.
- **Body**:
  ```json
  {
    "username": "string",
    "password": "string",
    "code": "string"  // Opsional: 2FA code jika 2FA aktif
  }
  ```
- **Response**: 
  - Jika 2FA tidak aktif: `AuthResponse` (token + user)
  - Jika 2FA aktif dan code belum diberikan: `{"requires_2fa": true, "message": "..."}`
  - Jika 2FA aktif dan code diberikan: `AuthResponse` (token + user)
- **Rate Limit**: 5 req/min
- **Note**: Token disimpan dalam httpOnly cookie (`auth_token`) untuk keamanan XSS

### Protected Endpoints (Require JWT)

#### 6. Get Profile
- **GET** `/api/v1/auth/profile`
- **Headers**: `Authorization: Bearer <token>` (opsional, fallback jika cookie tidak tersedia)
- **Cookies**: `auth_token` (httpOnly cookie, preferred method)
- **Response**: `User` object
- **Note**: Menggunakan httpOnly cookie untuk authentication (lebih aman dari XSS)

#### 7. Logout
- **POST** `/api/v1/auth/logout`
- **Headers**: `Authorization: Bearer <token>` atau `auth_token` cookie
- **Headers**: `X-CSRF-Token: <csrf_token>` (required untuk POST)
- **Response**: `{"message": "Logged out successfully"}`
- **Note**: Menghapus httpOnly cookie dan mencatat aksi logout ke audit log

#### 8. 2FA - Generate Secret
- **POST** `/api/v1/auth/2fa/generate`
- **Headers**: `X-CSRF-Token: <csrf_token>` (required)
- **Response**: `{"secret": "string", "qr_code": "base64_image", "url": "string", "message": "string"}`
- **Note**: Generate TOTP secret dan QR code untuk setup 2FA

#### 9. 2FA - Verify & Enable
- **POST** `/api/v1/auth/2fa/verify`
- **Headers**: `X-CSRF-Token: <csrf_token>` (required)
- **Body**: `{"code": "string"}` (6-digit TOTP code)
- **Response**: `{"message": "string", "backup_codes": ["string"]}`
- **Note**: Verifikasi TOTP code dan aktifkan 2FA, return backup codes

#### 10. 2FA - Get Status
- **GET** `/api/v1/auth/2fa/status`
- **Response**: `{"enabled": boolean}`
- **Note**: Cek status 2FA user saat ini

#### 11. 2FA - Disable
- **POST** `/api/v1/auth/2fa/disable`
- **Headers**: `X-CSRF-Token: <csrf_token>` (required)
- **Response**: `{"message": "2FA has been disabled successfully"}`
- **Note**: Nonaktifkan 2FA untuk user

#### 12. Audit Logs - List
- **GET** `/api/v1/audit-logs`
- **Query Parameters**: `page`, `pageSize`, `action`, `resource`, `status`, `logType`
- **Response**: Paginated audit logs
- **Note**: User reguler hanya melihat logs sendiri, admin/superadmin melihat semua

#### 13. Audit Logs - Stats
- **GET** `/api/v1/audit-logs/stats`
- **Response**: Statistics (total records, counts by type, estimated size, retention policy)
- **Note**: Real-time statistics untuk monitoring

#### 14. Documents - List
- **GET** `/api/v1/documents`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: Array of `Document`

#### 15. Documents - Get by ID
- **GET** `/api/v1/documents/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `Document` object

#### 16. Documents - Create
- **POST** `/api/v1/documents`
- **Headers**: `Authorization: Bearer <token>`
- **Body**: `Document` object
- **Response**: Created `Document`

#### 17. Documents - Update
- **PUT** `/api/v1/documents/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Body**: `Document` object
- **Response**: Updated `Document`

#### 18. Documents - Delete
- **DELETE** `/api/v1/documents/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: Success message

### Swagger UI
- **URL**: `http://localhost:8080/swagger/index.html`
- **Features**: Interactive API documentation, test endpoints directly

---

## Development Environment

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (optional, for local frontend dev)
- Go 1.25+ (optional, for local backend dev)

### Quick Start
```bash
# Start all services with hot reload
make dev

# Or using script
./dev.sh

# Or manual
docker-compose -f docker-compose.dev.yml up --build
```

### Environment Variables

#### Backend
- `PORT`: Server port (default: 8080)
- `ENV`: Environment (development/production)
- `DATABASE_URL`: PostgreSQL connection string (optional, defaults to SQLite)
- `CGO_ENABLED`: Enable CGO for SQLite (1 for dev, 0 for prod)
- `JWT_SECRET`: JWT signing secret (optional, has default)

#### Frontend
- `VITE_API_URL`: Backend API URL (default: `http://localhost:8080/api/v1`)
- `NODE_ENV`: Environment (development/production)

### Hot Reload
- **Frontend**: ✅ Vite HMR (automatic on file save)
- **Backend**: ⚠️ Manual restart required (or use local `go run`)

### Development Commands
```bash
# Makefile commands
make dev           # Start all services
make up            # Start in background
make down          # Stop services
make logs          # View logs
make restart       # Restart services
make rebuild       # Rebuild and restart

# Backend (local)
cd backend
go run ./cmd/api
go test ./...
golangci-lint run

# Frontend (local)
cd frontend
npm run dev
npm run build
npm run lint
```

---

## Production Environment

### Docker Images
- **Backend**: Multi-stage build, Alpine-based, ~20MB
- **Frontend**: Multi-stage build, Nginx-based, ~50MB

### Deployment
- **Backend**: Container runs Go binary on port 8080
- **Frontend**: Nginx serves static files on port 80
- **Database**: PostgreSQL (external service recommended)

### Recommended Setup
1. **Backend**: Deploy to container platform (Kubernetes, Docker Swarm, Cloud Run, ECS)
2. **Frontend**: Deploy to CDN or static hosting (Vercel, Netlify, S3+CloudFront)
3. **Database**: Managed PostgreSQL (AWS RDS, Google Cloud SQL, Azure Database)
4. **Reverse Proxy**: Nginx or Cloud Load Balancer

### Production Environment Variables
```bash
# Backend
PORT=8080
ENV=production
# DATABASE_URL dengan SSL/TLS (wajib untuk production)
# Format: postgres://user:pass@host:5432/dbname?sslmode=require
# Atau: postgres://user:pass@host:5432/dbname?sslmode=verify-full (lebih aman, butuh CA cert)
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
JWT_SECRET=your-secret-key-here
CGO_ENABLED=0
AUDIT_LOG_USER_ACTION_RETENTION_DAYS=365
AUDIT_LOG_TECHNICAL_ERROR_RETENTION_DAYS=90

# Frontend (build-time)
VITE_API_URL=https://api.yourdomain.com/api/v1
```

**Catatan Penting**:
- **SSL/TLS**: Pastikan `DATABASE_URL` menggunakan `sslmode=require` atau `sslmode=verify-full` di production
- **Connection Pooling**: Sudah dikonfigurasi otomatis di aplikasi (tidak perlu environment variable)
- **Backup**: Setup automated backup menggunakan cron job atau managed service backup

### Security Checklist
- ✅ Use strong JWT secret
- ✅ Enable HTTPS/TLS
- ✅ Enable SSL/TLS untuk database connections (`sslmode=require`)
- ✅ Configure connection pooling untuk database stability
- ✅ Setup automated database backups
- ✅ Configure CORS for production domains
- ✅ Use managed database with backups
- ✅ Enable rate limiting
- ✅ Monitor audit logs
- ✅ Regular security scans (Trivy)
- ✅ Keep dependencies updated

---

## Version Information

### Current Versions
- **Go**: 1.25
- **Node.js**: 20+
- **Vue**: 3.5.22
- **TypeScript**: 5.9.0
- **Vite**: 7.1.11
- **Ant Design Vue**: 4.2.6

### Version Tagging
- **Format**: `vX.Y.Z` (e.g., `v1.0.0`)
- **Auto-generation**: From Git tags or commit SHA
- **CI/CD**: Automatic on tag push

---

## Sample User Credentials

### Superadmin
- **Username**: `superadmin`
- **Password**: `Pedeve123`
- **Role**: `superadmin`
- **Email**: Auto-generated

---

## Documentation Files

- `README.md`: Project overview and quick start
- `DEVELOPMENT.md`: Development guide dengan hot reload
- `SPESIFIKASI_TEKNIS.md`: This file (technical specifications)
- `URLS_AND_PORTS.md`: Quick reference for URLs and ports

### Backend Documentation (dalam folder backend/)
- `backend/doc/AUTH.md`: Authentication documentation
- `backend/doc/SECURITY_FEATURES.md`: Security features documentation
- `backend/doc/AUDIT_LOG_OPTIMIZATION.md`: Audit log optimization documentation
- `backend/doc/REFACTORING_NOTES.md`: Catatan refactoring ke Clean Architecture

---

## Architecture Notes

### Clean Architecture
Project menggunakan Clean Architecture dengan struktur:
- **cmd/**: Entry points aplikasi (api, migrate, worker, dll)
- **internal/**: Kode internal aplikasi
  - **domain/**: Domain models dan entities
  - **infrastructure/**: External dependencies (database, JWT, password, dll)
  - **delivery/**: Delivery layer (HTTP handlers)
  - **middleware/**: HTTP middleware
  - **repository/**: Data access layer
  - **usecase/**: Business logic layer

### Framework Migration
- **Sebelumnya**: Chi Router v5.2.3
- **Sekarang**: Fiber v2.52.10
- **Alasan**: Requirement project untuk menggunakan Fiber, Gin, atau Echo
- **Keuntungan**: 
  - Performa tinggi dengan fasthttp
  - API yang lebih modern dan intuitif
  - Built-in middleware yang lengkap
  - Tim internal sudah familiar dengan Fiber

---

**Last Updated**: 2025-11-21
**Version**: 1.0.0

