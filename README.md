# Pedeve DMS App - Document Management System

Aplikasi Document Management System untuk manajemen dokumen dan perusahaan dengan hierarki multi-level.

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (untuk development frontend)
- Go 1.25+ (untuk development backend)

### Development Setup

#### 🚀 Quick Start - Satu Perintah untuk Semua

**Dengan SQLite (Default):**
```bash
# Cara termudah - run semua service dengan hot reload
make dev

# Atau menggunakan script
./dev.sh

# Atau manual
docker-compose -f docker-compose.dev.yml up --build
```

**Dengan PostgreSQL:**
```bash
# Gunakan file docker-compose khusus PostgreSQL
docker-compose -f docker-compose.postgres.yml up --build

# Atau set DATABASE_URL di docker-compose.dev.yml
# PostgreSQL sudah dikonfigurasi di docker-compose.dev.yml
```

**Hot Reload:**
- ✅ Backend: Auto-reload saat file `.go` berubah (menggunakan Air)
- ✅ Frontend: Auto-reload saat file Vue/TS berubah (Vite HMR)
- ✅ Tidak perlu down/up manual - cukup save file dan refresh browser!

**Akses:**
- Frontend: http://localhost:5173 (dev) atau http://localhost:3000 (prod)
- Backend API: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health
- API Base: http://localhost:8080/api/v1

#### Option 2: Local Development (Lebih cepat untuk development)

**Backend:**
```bash
cd backend
go mod download
go run ./cmd/api/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## 📁 Project Structure

```
pedeve-dms-app/
├── backend/                    # Go backend API (Clean Architecture)
│   ├── cmd/
│   │   ├── api/               # API server entry point
│   │   │   └── main.go
│   │   └── seed-companies/    # Company seeder
│   ├── internal/
│   │   ├── domain/            # Domain models & entities
│   │   ├── infrastructure/    # External dependencies (DB, JWT, Storage, etc)
│   │   ├── delivery/          # HTTP handlers (Fiber)
│   │   ├── middleware/        # HTTP middleware (Auth, CSRF, Rate limit, etc)
│   │   ├── repository/        # Data access layer
│   │   └── usecase/           # Business logic layer
│   ├── go.mod
│   └── Dockerfile
├── frontend/                   # Vue 3 frontend
│   ├── src/
│   │   ├── api/               # API clients
│   │   ├── components/        # Vue components
│   │   ├── views/             # Page views
│   │   ├── stores/            # Pinia stores
│   │   └── router/            # Vue Router
│   ├── package.json
│   └── Dockerfile
├── .github/
│   └── workflows/             # CI/CD pipelines
├── scripts/                   # Deployment scripts
├── documentations/            # Documentation files
└── docker-compose.dev.yml     # Local development setup
```

## 🔧 Development Commands

### Quick Commands (Makefile)

```bash
make dev           # Start all services dengan hot reload
make up            # Start services in background
make down          # Stop all services
make restart       # Restart services
make logs          # View all logs
make logs-backend  # View backend logs only
make logs-frontend # View frontend logs only
make status        # Check service status
make clean         # Clean everything
make rebuild       # Rebuild and restart
make help          # Show all commands
```

### Manual Commands

**Backend:**
```bash
cd backend
go run ./cmd/api/main.go    # Run server (local, tanpa Docker)
go test ./...               # Run tests
golangci-lint run           # Lint code

# Generate Swagger docs (setelah update annotations)
swag init -g cmd/api/main.go --output docs
```

**Frontend:**
```bash
cd frontend
npm run dev             # Development server (local, tanpa Docker)
npm run build           # Build for production
npm run lint            # Lint code
npm run test:unit       # Run tests
```

## 🐳 Docker Commands

```bash
# Development (dengan hot reload)
make dev                    # Start dengan hot reload
docker-compose -f docker-compose.dev.yml up --build

# Production
docker-compose up --build

# Background
make up                     # Start in background
docker-compose -f docker-compose.dev.yml up -d

# Stop
make down                   # Stop services
docker-compose -f docker-compose.dev.yml down

# Logs
make logs                   # View all logs
make logs-backend           # Backend only
make logs-frontend          # Frontend only
docker-compose -f docker-compose.dev.yml logs -f

# Status
make status                 # Check status
docker-compose -f docker-compose.dev.yml ps
```

## 🚢 CI/CD

Pipeline otomatis berjalan saat:
- Push ke branch `main`
- Push tag versi (v1.0.0, v2.1.3, dll)

**Fitur CI/CD:**
- ✅ Lint & Test (Frontend & Backend)
- ✅ Security Scan (Trivy)
- ✅ Build Docker Images
- ✅ Push ke GitHub Container Registry
- ✅ Automatic Version Tagging
- ✅ Generate Changelog
- ✅ Create GitHub Release (saat push tag)

## 📝 Release Process

```bash
# 1. Buat tag versi
git tag v1.0.0
git push origin v1.0.0

# 2. CI/CD akan otomatis:
#    - Build images dengan tag v1.0.0
#    - Generate changelog
#    - Create GitHub release
#    - Push images ke registry
```

## 🔍 API Documentation

### Swagger UI
Akses dokumentasi API lengkap di: **http://localhost:8080/swagger/index.html**

Swagger UI menyediakan:
- ✅ Dokumentasi semua endpoint
- ✅ Test API langsung dari browser
- ✅ Request/Response examples
- ✅ Schema definitions

### API Endpoints

**Authentication:**
- `POST /api/v1/auth/login` - Login (dengan 2FA support)
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/profile` - Get user profile
- `POST /api/v1/auth/2fa/generate` - Generate 2FA QR code
- `POST /api/v1/auth/2fa/verify` - Verify 2FA code

**Company Management:**
- `GET /api/v1/companies` - Get all companies (dengan hierarki)
- `GET /api/v1/companies/{id}` - Get company detail
- `POST /api/v1/companies` - Create company
- `PUT /api/v1/companies/{id}` - Update company
- `DELETE /api/v1/companies/{id}` - Delete company (soft delete)
- `GET /api/v1/companies/{id}/users` - Get users assigned to company

**User Management:**
- `GET /api/v1/users` - Get all users (dengan RBAC filtering)
- `GET /api/v1/users/{id}` - Get user detail
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/{id}` - Update user
- `POST /api/v1/users/{id}/assign-company` - Assign user to company
- `POST /api/v1/users/{id}/unassign-company` - Unassign user from company

**Development (Superadmin Only):**
- `POST /api/v1/development/reset-subsidiary` - Reset subsidiary data
- `POST /api/v1/development/run-subsidiary-seeder` - Run company seeder
- `GET /api/v1/development/check-seeder-status` - Check seeder status

**Audit Logs:**
- `GET /api/v1/audit-logs` - Get audit logs (dengan retention policy: 90 hari user actions, 30 hari technical errors)
- `GET /api/v1/audit-logs/stats` - Get audit log statistics
- `GET /api/v1/user-activity-logs` - Get user activity logs (permanent storage untuk data penting: report, document, company, user)

**Documents:**
- `GET /api/v1/documents` - Get all documents
- `GET /api/v1/documents/{id}` - Get document by ID
- `POST /api/v1/documents` - Create new document
- `PUT /api/v1/documents/{id}` - Update document
- `DELETE /api/v1/documents/{id}` - Delete document

**File Upload:**
- `POST /api/v1/upload/logo` - Upload company logo
- `GET /api/v1/files/*` - Serve files (proxy dari GCP Storage atau local)



## 🛠️ Troubleshooting

### Port sudah digunakan
```bash
# Cek port yang digunakan
lsof -i :8080
lsof -i :5173

# Atau ubah port di docker-compose.dev.yml
```

### Docker build error
```bash
# Clean build
docker-compose -f docker-compose.dev.yml down
docker system prune -f
docker-compose -f docker-compose.dev.yml up --build
```

### Frontend tidak connect ke backend
- Pastikan `VITE_API_URL` atau `VITE_API_BASE_URL` di frontend sesuai dengan backend URL
- Cek CORS settings di backend (default: localhost:5173, localhost:3000)
- Pastikan backend sudah running di port 8080

### CSRF Token Error
- Pastikan frontend menggunakan `apiClient` dari `frontend/src/api/client.ts`
- `apiClient` otomatis menambahkan CSRF token untuk POST/PUT/DELETE/PATCH
- Jika masih error, coba logout dan login ulang untuk refresh token

### Database Connection Error
- Untuk PostgreSQL: Pastikan `DATABASE_URL` sudah di-set dengan benar
- Untuk SQLite: File database akan dibuat otomatis di `backend/dms.db`
- Cek koneksi database di `backend/internal/infrastructure/database/database.go`

### Seeder tidak jalan
- Pastikan role "admin" sudah ada di database (auto-created saat startup)
- Gunakan fitur "Jalankan Seeder Data Subsidiary" di Settings (superadmin only)
- Atau jalankan manual: `cd backend && go run ./cmd/seed-companies`

## 🛠️ Tech Stack

### Frontend
- **Framework**: Vue 3 (Composition API)
- **Language**: TypeScript
- **Build Tool**: Vite 7
- **State Management**: Pinia
- **Routing**: Vue Router 4
- **UI Library**: Ant Design Vue 4
- **HTTP Client**: Axios
- **Charts**: Chart.js + Vue-ChartJS
- **Icons**: Iconify Vue
- **Date**: Day.js

### Backend
- **Language**: Go 1.25
- **Web Framework**: Fiber v2 (fasthttp-based, high performance)
- **Architecture**: Clean Architecture (Domain, Infrastructure, Delivery, Usecase, Repository)
- **ORM**: GORM
- **Database**: PostgreSQL (production) / SQLite (development)
- **Authentication**: JWT (golang-jwt/jwt/v5) dengan httpOnly cookies
- **2FA**: TOTP (pquerna/otp)
- **Password**: bcrypt (golang.org/x/crypto)
- **Logging**: Zap (go.uber.org/zap)
- **Validation**: go-playground/validator
- **Storage**: Google Cloud Storage / Local filesystem
- **Secrets**: GCP Secret Manager / HashiCorp Vault
- **API Docs**: Swagger/OpenAPI (swaggo/swag)

### Security Features
- ✅ **CSRF Protection**: Double-submit cookie pattern
- ✅ **Rate Limiting**: 100 req/s (general), 5 req/min (auth endpoints)
- ✅ **Security Headers**: X-Content-Type-Options, X-XSS-Protection, CSP, HSTS
- ✅ **2FA Support**: TOTP-based dengan backup codes
- ✅ **Audit Logging**: 
  - Comprehensive audit logging untuk semua aksi user dan error teknis
  - Retention policy: 90 hari untuk user actions, 30 hari untuk technical errors
  - **Permanent Audit Log**: Data penting (Report, Document, Company, User Management) disimpan permanen tanpa retention policy untuk compliance
- ✅ **JWT Security**: httpOnly cookies untuk mencegah XSS
- ✅ **Input Validation**: Comprehensive validation dengan sanitization
- ✅ **Password Security**: bcrypt hashing

### Infrastructure
- **Container**: Docker, Docker Compose
- **CI/CD**: GitHub Actions
- **Deployment**: Google Cloud Platform (GCP)
- **Storage**: Google Cloud Storage
- **Secrets**: GCP Secret Manager
- **Security Scan**: Trivy Scanner
- **API Docs**: Swagger UI dengan auto-reload

## 🎯 Fitur Utama

### Authentication & Authorization
- ✅ JWT-based authentication dengan httpOnly cookies
- ✅ Two-Factor Authentication (2FA) dengan TOTP
- ✅ Role-Based Access Control (RBAC)
- ✅ Company hierarchy-based access control
- ✅ CSRF protection untuk state-changing requests

### Company Management
- ✅ Multi-level company hierarchy (Holding → Level 1 → Level 2 → Level 3)
- ✅ Company CRUD dengan validasi hierarki
- ✅ Company detail dengan shareholders, business fields, directors
- ✅ Company logo upload (GCP Storage / Local)
- ✅ "My Company" view untuk melihat company user yang di-assign

### User Management
- ✅ User CRUD dengan RBAC
- ✅ Multiple company assignments per user (junction table)
- ✅ Flexible role assignment per company
- ✅ User status management (active/inactive)
- ✅ Password reset functionality
- ✅ Standby users (tanpa company/role assignment)

### Development Tools
- ✅ Reset subsidiary data (superadmin only)
- ✅ Run company seeder via UI (superadmin only)
- ✅ Seeder status check

### Security & Monitoring
- ✅ Comprehensive audit logging dengan retention policy
- ✅ **Permanent Audit Log**: User Activity Logs untuk data penting (Report, Document, Company, User) - disimpan permanen tanpa retention
- ✅ Rate limiting (per endpoint type)
- ✅ Security headers (CSP, HSTS, XSS protection)
- ✅ Input validation & sanitization
- ✅ Error logging dengan stack trace
- ✅ Audit log UI dengan tab terpisah untuk "Audit Logs" dan "User Activity"

## 🤝 Contributing

1. Buat branch dari `development` (untuk fitur baru) atau `main` (untuk hotfix)
2. Develop fitur dengan mengikuti Clean Architecture pattern
3. Test & lint (frontend: `npm run lint`, backend: `golangci-lint run`)
4. Push dan buat PR ke branch `development`
5. Setelah merge, CI/CD akan otomatis build dan deploy ke GCP

## 📖 Dokumentasi Tambahan

- **API Documentation**: http://localhost:8080/swagger/index.html
- **Seeder Documentation**: `backend/cmd/seed-companies/README.md`
- **Manual Fixes**: `documentations/MANUAL_FIXES_DOCUMENTATION.md`
- **Backend Architecture**: Clean Architecture dengan struktur `cmd/`, `internal/`

## 📦 Build Version

**Build Date**: 2025-11-30 15:30:00 +0700  
**Commit**: Latest (permanent audit log implementation)  
**Version**: 1.0.0

### Recent Updates
- ✅ Implementasi permanent audit log untuk data penting (Report, Document, Company, User Management)
- ✅ Tab "User Activity" di halaman Settings untuk melihat permanent logs
- ✅ Differentiated retention policy: permanent storage untuk data penting, retention 90/30 hari untuk data biasa
- ✅ Table terpisah `user_activity_logs` untuk permanent storage tanpa retention policy

