# Pedeve DMS App

Document Management System untuk manajemen dokumen dan perusahaan dengan hierarki multi-level.

## Deskripsi

Pedeve DMS App adalah aplikasi manajemen dokumen yang dirancang untuk mengelola dokumen dan data perusahaan dalam struktur hierarki multi-level. Aplikasi ini menyediakan fitur lengkap untuk manajemen perusahaan, dokumen, laporan keuangan, pengguna, dan sistem notifikasi.

## Persyaratan Sistem

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (untuk development frontend)
- Go 1.25+ (untuk development backend)

## Pengaturan Development

### Menggunakan Docker Compose (Recommended)

**Dengan SQLite (Default):**
```bash
make dev

# Atau menggunakan script
./dev.sh

# Atau manual
docker-compose -f docker-compose.dev.yml up --build
```

**Dengan PostgreSQL:**
```bash
docker-compose -f docker-compose.postgres.yml up --build

# Atau set DATABASE_URL di docker-compose.dev.yml
```

**Hot Reload:**
- Backend: Auto-reload saat file `.go` berubah (menggunakan Air)
- Frontend: Auto-reload saat file Vue/TS berubah (Vite HMR)
- Tidak perlu down/up manual - cukup save file dan refresh browser

**Akses:**
- Frontend: http://localhost:5173 (development) atau http://localhost:3000 (production)
- Backend API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health
- API Base: http://localhost:8080/api/v1

### Development Lokal (tanpa Docker)

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

## Struktur Projek

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

## Perintah Development

### Perintah Cepat (Makefile)

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

### Perintah Manual

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
npm run test:unit       # Run unit tests dengan Vitest
```

## Testing

### Frontend (Vitest)

```bash
cd frontend
npm run test:unit        # Run unit tests
```

Framework: Vitest (Vite-native test runner)
Environment: jsdom (browser-like environment)
Test Utils: Vue Test Utils
Coverage: Integrated dengan Vitest

### Backend (Go Test)

```bash
cd backend
go test ./...            # Run all tests
go test ./... -v         # Verbose output
go test ./... -cover      # With coverage report
```

Framework: Go built-in testing
Coverage: Integrated dengan `go tool cover`
CI/CD: Otomatis dijalankan di GitHub Actions

## CI/CD & Deployment

### CI/CD Pipeline

Pipeline otomatis berjalan saat:
- Push ke branch `main` atau `development`
- Push tag versi (v1.0.0, v2.1.3, dll)
- Manual trigger via `workflow_dispatch`

**Fitur CI/CD:**
- Lint & Test: Frontend (ESLint + Vitest) & Backend (golangci-lint + Go test)
- Security Scan: Trivy vulnerability scanner untuk Docker images
- Build: Docker images untuk backend, static files untuk frontend
- Deploy: Otomatis deploy ke GCP VM saat push ke `development`
- Registry: Push images ke GitHub Container Registry
- Versioning: Automatic version tagging
- Changelog: Generate changelog otomatis
- Release: Create GitHub Release (saat push tag)

### Deployment Automation

Setelah deployment selesai, server langsung bisa diakses tanpa perlu menjalankan script manual. Semua proses otomatis:

**SSL Certificate Management:**
- Otomatis detect apakah SSL certificate sudah ada
- Jika belum ada, otomatis membuat via Certbot (Let's Encrypt)
- Idempotent: aman dipanggil berkali-kali, tidak akan error jika certificate sudah ada
- Auto-renewal: Certbot timer otomatis setup untuk renewal

**Nginx Configuration:**
- Otomatis setup Nginx config dengan HTTPS (jika SSL certificate ada)
- Fallback ke HTTP config jika SSL belum tersedia
- Preserve existing config yang sudah benar (tidak di-overwrite)
- Otomatis reload Nginx setelah config update

**Service Management:**
- Otomatis ensure Docker container running (backend)
- Otomatis ensure Nginx service running (frontend & backend)
- Health check otomatis setelah deployment
- Retry mechanism jika service belum ready

**Deployment Flow:**
1. Build & Test
2. Deploy Files/Images
3. Setup SSL (if needed)
4. Setup Nginx
5. Ensure Services Running
6. Health Check
7. Ready

### Release Process

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

## Dokumentasi API

### Swagger UI

Akses dokumentasi API lengkap di: http://localhost:8080/swagger/index.html

Swagger UI menyediakan:
- Dokumentasi semua endpoint
- Test API langsung dari browser
- Request/Response examples
- Schema definitions

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

**Financial Reports:**
- `GET /api/v1/financial-reports` - Get all financial reports
- `GET /api/v1/financial-reports/{id}` - Get financial report by ID
- `POST /api/v1/financial-reports` - Create financial report
- `PUT /api/v1/financial-reports/{id}` - Update financial report
- `DELETE /api/v1/financial-reports/{id}` - Delete financial report
- `GET /api/v1/financial-reports/company/{company_id}` - Get all financial reports for a company
- `GET /api/v1/financial-reports/bulk-upload/template` - Download bulk upload template Excel
- `POST /api/v1/financial-reports/bulk-upload/validate` - Validate bulk upload Excel file
- `POST /api/v1/financial-reports/bulk-upload` - Upload bulk financial reports (upsert)
- `GET /api/v1/financial-reports/compare` - Get comparison RKAP vs Realisasi YTD
- `GET /api/v1/financial-reports/rkap-years/{company_id}` - Get RKAP years for a company

**Documents:**
- `GET /api/v1/documents` - Get all documents
- `GET /api/v1/documents/{id}` - Get document by ID
- `POST /api/v1/documents` - Create new document
- `PUT /api/v1/documents/{id}` - Update document
- `DELETE /api/v1/documents/{id}` - Delete document

**File Upload:**
- `POST /api/v1/upload/logo` - Upload company logo
- `GET /api/v1/files/*` - Serve files (proxy dari GCP Storage atau local)

**Audit Logs:**
- `GET /api/v1/audit-logs` - Get audit logs (dengan retention policy: 90 hari user actions, 30 hari technical errors)
- `GET /api/v1/audit-logs/stats` - Get audit log statistics
- `GET /api/v1/user-activity-logs` - Get user activity logs (permanent storage untuk data penting: report, document, company, user)

**Notifications:**
- `GET /api/v1/notifications` - Get all notifications
- `GET /api/v1/notifications/unread-count` - Get unread notification count
- `PUT /api/v1/notifications/{id}/read` - Mark notification as read
- `PUT /api/v1/notifications/read-all` - Mark all notifications as read

**Development (Superadmin Only):**
- `POST /api/v1/development/reset-subsidiary` - Reset subsidiary data
- `POST /api/v1/development/reset-all-financial-reports` - Reset all financial reports
- `POST /api/v1/development/run-subsidiary-seeder` - Run company seeder
- `GET /api/v1/development/check-seeder-status` - Check seeder status
- `POST /api/v1/development/check-expiring-documents` - Manual trigger check expiring documents
- `POST /api/v1/development/check-expiring-director-terms` - Manual trigger check expiring director terms
- `POST /api/v1/development/check-all-expiring-notifications` - Manual trigger check all expiring notifications

## Troubleshooting

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

## Tech Stack

### Frontend
- Framework: Vue 3 (Composition API)
- Language: TypeScript
- Build Tool: Vite 7
- State Management: Pinia
- Routing: Vue Router 4
- UI Library: Ant Design Vue 4
- HTTP Client: Axios
- Charts: Chart.js + Vue-ChartJS
- Icons: Iconify Vue
- Date: Day.js
- Testing: Vitest + Vue Test Utils
- Logging: Custom logger utility (production-safe, hanya debug/info muncul di development)

### Backend
- Language: Go 1.25
- Web Framework: Fiber v2 (fasthttp-based, high performance)
- Architecture: Clean Architecture (Domain, Infrastructure, Delivery, Usecase, Repository)
- ORM: GORM
- Database: PostgreSQL (production) / SQLite (development)
- Authentication: JWT (golang-jwt/jwt/v5) dengan httpOnly cookies
- 2FA: TOTP (pquerna/otp)
- Password: bcrypt (golang.org/x/crypto)
- Logging: Zap (go.uber.org/zap)
- Validation: go-playground/validator
- Storage: Google Cloud Storage / Local filesystem
- Secrets: GCP Secret Manager / HashiCorp Vault
- API Docs: Swagger/OpenAPI (swaggo/swag)
- Excel Processing: Excelize

### Security Features
- CSRF Protection: Double-submit cookie pattern
- Rate Limiting: 100 req/s (general), 5 req/min (auth endpoints)
- Security Headers: X-Content-Type-Options, X-XSS-Protection, CSP, HSTS
- 2FA Support: TOTP-based dengan backup codes
- Transparent Data Encryption (TDE):
  - SQLite: SQLCipher untuk encryption at rest (development)
  - PostgreSQL: Automatic encryption at rest (GCP Cloud SQL) atau filesystem encryption (self-hosted)
  - Key management via GCP Secret Manager / HashiCorp Vault / Environment variables
- Audit Logging:
  - Comprehensive audit logging untuk semua aksi user dan error teknis
  - Retention policy: 90 hari untuk user actions, 30 hari untuk technical errors
  - Permanent Audit Log: Data penting (Report, Document, Company, User Management) disimpan permanen tanpa retention policy untuk compliance
- JWT Security: httpOnly cookies untuk mencegah XSS
- Input Validation: Comprehensive validation dengan sanitization
- Password Security: bcrypt hashing
- Production-Safe Logging: Frontend menggunakan logger utility yang hanya menampilkan debug/info di development, error/warn tetap muncul di production

### Infrastructure
- Container: Docker, Docker Compose
- CI/CD: GitHub Actions dengan automated testing
- Deployment: Google Cloud Platform (GCP) dengan automated SSL & Nginx setup
- Web Server: Nginx dengan automatic HTTPS/HTTP configuration
- SSL: Let's Encrypt dengan automatic certificate management
- Storage: Google Cloud Storage
- Secrets: GCP Secret Manager
- Security Scan: Trivy Scanner
- API Docs: Swagger UI dengan auto-reload

## Fitur Utama

### Authentication & Authorization
- JWT-based authentication dengan httpOnly cookies
- Two-Factor Authentication (2FA) dengan TOTP
- Role-Based Access Control (RBAC)
- Company hierarchy-based access control
- CSRF protection untuk state-changing requests

### Company Management
- Multi-level company hierarchy (Holding → Level 1 → Level 2 → Level 3)
- Company CRUD dengan validasi hierarki
- Company detail dengan shareholders, business fields, directors
- Company logo upload (GCP Storage / Local)
- "My Company" view untuk melihat company user yang di-assign

### User Management
- User CRUD dengan RBAC
- Multiple company assignments per user (junction table)
- Flexible role assignment per company
- User status management (active/inactive)
- Password reset functionality
- Standby users (tanpa company/role assignment)

### Document Management
- Document CRUD operations
- Document categorization dengan folder structure
- File upload dan storage (GCP Storage / Local)
- Document expiry tracking
- Document status management

### Financial Reports Management
- Financial report CRUD operations
- Support untuk RKAP (Rencana Kerja dan Anggaran Perusahaan) dan Realisasi
- Bulk upload financial reports via Excel
- Template Excel generation untuk bulk upload
- Excel validation sebelum upload
- Upsert mechanism (update if exists, insert if new)
- Comparison RKAP vs Realisasi YTD
- Monthly report status tracking per subsidiary
- Financial ratios calculation dan validation

### Notification System
- In-app notifications untuk berbagai events
- Document expiry notifications (grouped by folder)
- Director term expiry notifications
- Dynamic notification messages berdasarkan waktu real-time
- Unread notification count
- Mark as read functionality
- Automated scheduler untuk check expiring items (every 24 hours)
- Configurable threshold days via environment variable

### Development Tools
- Reset subsidiary data (superadmin only)
- Reset all financial reports (superadmin only)
- Run company seeder via UI (superadmin only)
- Seeder status check
- Manual notification trigger untuk testing (superadmin/administrator only)

### Security & Monitoring
- Comprehensive audit logging dengan retention policy
- Permanent Audit Log: User Activity Logs untuk data penting (Report, Document, Company, User) - disimpan permanen tanpa retention
- Rate limiting (per endpoint type)
- Security headers (CSP, HSTS, XSS protection)
- Input validation & sanitization
- Error logging dengan stack trace
- Audit log UI dengan tab terpisah untuk "Audit Logs" dan "User Activity"

## Contributing

1. Buat branch dari `development` (untuk fitur baru) atau `main` (untuk hotfix)
2. Develop fitur dengan mengikuti Clean Architecture pattern
3. Write tests: Frontend (Vitest) dan Backend (Go test)
4. Test & lint:
   - Frontend: `npm run lint && npm run test:unit`
   - Backend: `golangci-lint run && go test ./...`
5. Push dan buat PR ke branch `development`
6. Setelah merge, CI/CD akan otomatis:
   - Run tests (frontend & backend)
   - Build dan deploy ke GCP
   - Setup SSL & Nginx otomatis
   - Verify services running

## Dokumentasi Tambahan

- API Documentation: http://localhost:8080/swagger/index.html
- Seeder Documentation: `backend/cmd/seed-companies/README.md`
- Manual Fixes: `documentations/MANUAL_FIXES_DOCUMENTATION.md`
- Backend Architecture: Clean Architecture dengan struktur `cmd/`, `internal/`
- Testing Guide: `backend/doc/TESTING_USER_MANAGEMENT.md` (untuk manual testing)
- Deployment Scripts: `scripts/` folder berisi semua deployment automation scripts
- **TDE (Transparent Data Encryption)**: 
  - `documentations/TDE_IMPLEMENTATION_GUIDE.md` - Panduan lengkap implementasi TDE untuk PostgreSQL dan SQLite
- **PDP Compliance**: 
  - `documentations/PDP_DATA_CLASSIFICATION.md` - Klasifikasi data pribadi berdasarkan UU No. 27 Tahun 2022
  - `documentations/PDP_ENCRYPTION_BEST_PRACTICES.md` - Best practices enkripsi untuk compliance
