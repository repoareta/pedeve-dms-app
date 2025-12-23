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
# docker-compose.dev.yml sudah include PostgreSQL
# Set DATABASE_URL di docker-compose.dev.yml jika perlu
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

### Menjalankan Semua Automated Tests

Untuk menjalankan semua automated tests (backend + frontend) sekaligus, gunakan command:

```bash
make test
```

Command ini akan:
- Menjalankan semua backend tests (Go test)
- Menjalankan semua frontend tests (Vitest)
- Menampilkan summary hasil test

### Frontend (Vitest)

```bash
cd frontend
npm run test:unit        # Run unit tests
```

**Framework:** Vitest (Vite-native test runner)  
**Environment:** jsdom (browser-like environment)  
**Test Utils:** Vue Test Utils  
**Coverage:** Integrated dengan Vitest

**Apa yang di-test:**
- **Logika Bisnis:**
  - Perhitungan persentase kepemilikan saham (kalau ada 2 pemegang saham dengan modal berbeda, persentasenya harus benar)
  - Penentuan perusahaan induk (perusahaan mana yang jadi parent berdasarkan kepemilikan terbesar)
  - Perhitungan persentase kepemilikan perusahaan sendiri
  - Penanganan kasus khusus (misalnya modal = 0, modal perusahaan sendiri lebih besar dari total modal pemegang saham)
  - Data yang diinput di form tersimpan dengan benar (termasuk ownership_percent dan parent_id)
  - File attachment untuk direktur bisa di-upload dan ditampilkan
- **Logika Komponen:**
  - Validasi form (field wajib harus diisi, format harus benar)
  - Data binding (kalau input berubah, data otomatis ter-update)
  - Reactive updates (persentase kepemilikan otomatis terhitung ulang saat modal berubah)
  - Computed properties untuk perhitungan dinamis

### Backend (Go Test)

```bash
cd backend
go test ./...            # Run all tests
go test ./... -v         # Verbose output
go test ./... -cover      # With coverage report
```

**Framework:** Go built-in testing  
**Coverage:** Integrated dengan `go tool cover`  
**CI/CD:** Otomatis dijalankan di GitHub Actions

**Apa yang di-test:**
- **Logika Bisnis:**
  - CRUD laporan keuangan (buat, baca, update, hapus laporan)
  - Validasi data laporan:
    - Company ID harus ada dan valid
    - Inputter ID harus ada dan valid (kalau diisi)
    - Period harus format benar (YYYY-MM)
    - Tidak boleh ada duplicate period untuk perusahaan yang sama
    - Field wajib harus diisi
  - Upload banyak file sekaligus (bulk upload) dan validasinya
  - Baca data dari file Excel dan ekstrak datanya
  - Perhitungan rasio keuangan
  - Perbandingan RKAP vs Realisasi (apakah perhitungannya benar)
  - RBAC (Role-Based Access Control):
    - Superadmin bisa akses semua laporan
    - Admin hanya bisa akses laporan perusahaan mereka
    - User reguler hanya bisa akses laporan perusahaan mereka
    - Validasi akses berdasarkan role dan company assignment
- **API Endpoints:**
  - Response dari API endpoint (apakah return data yang benar)
  - Validasi request dan response
  - Error handling (kalau ada error, apakah pesannya jelas)
  - Upload banyak file via API (bulk upload Excel)
  - Export laporan ke Excel (dengan filter period, company, multiple companies)
  - Export laporan ke PDF (dengan filter period, company, multiple companies)
  - Generate template Excel untuk bulk upload
  - Route ordering (export routes tidak conflict dengan parameterized routes)
- **Database Operations:**
  - Operasi database (simpan, baca, update, hapus data)
  - Filter dan pagination data
  - Relasi antar data (perusahaan, user, laporan)
  - Query berdasarkan company ID dengan RBAC filtering

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

**Catatan Penting:** Swagger UI hanya tersedia dan dapat diakses di environment **development**. Di environment **production**, Swagger UI tidak diaktifkan untuk alasan keamanan.

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
- Autentikasi berbasis JWT dengan httpOnly cookies untuk meningkatkan keamanan
- Two-Factor Authentication (2FA) menggunakan TOTP sebagai perlindungan tambahan
- Role-Based Access Control (RBAC) untuk mengatur akses berdasarkan peran pengguna
- Kontrol akses berbasis hierarki perusahaan untuk membatasi akses data sesuai level perusahaan
- Proteksi CSRF untuk mencegah serangan pada request yang mengubah state

### Company Management
- Hierarki perusahaan multi-level (Holding → Level 1 → Level 2 → Level 3)
- Operasi CRUD perusahaan dengan validasi hierarki untuk memastikan struktur organisasi tetap konsisten
- Detail perusahaan mencakup informasi pemegang saham, bidang usaha, dan dewan direksi
- Upload logo perusahaan dengan penyimpanan di GCP Storage atau sistem lokal
- Tampilan "My Company" untuk melihat perusahaan yang di-assign kepada pengguna

### User Management
- Operasi CRUD pengguna dengan kontrol akses berbasis RBAC
- Assignment perusahaan ganda per pengguna menggunakan junction table untuk fleksibilitas
- Penugasan peran yang fleksibel per perusahaan, memungkinkan satu pengguna memiliki peran berbeda di perusahaan berbeda
- Manajemen status pengguna (aktif/nonaktif) untuk mengontrol akses
- Fungsi reset password untuk pemulihan akses
- Pengguna standby yang belum memiliki assignment perusahaan atau peran

### Document Management
- Operasi CRUD dokumen untuk mengelola dokumen secara lengkap
- Kategorisasi dokumen menggunakan struktur folder untuk organisasi yang lebih baik
- Upload dan penyimpanan file dengan dukungan GCP Storage atau sistem lokal
- **Dukungan batch upload** - Memungkinkan upload beberapa file sekaligus (PDF, gambar, dokumen)
- **Validasi ukuran file** - File dokumen (.docx, .xlsx, .xls, .pptx, .ppt, .pdf) tanpa batasan ukuran, file gambar (.jpg, .jpeg, .png) maksimal 10MB
- Pelacakan tanggal kedaluwarsa dokumen untuk manajemen siklus hidup dokumen
- Manajemen status dokumen untuk mengontrol visibilitas dan akses
- Preview dokumen untuk gambar dan PDF dengan modal fullscreen

### Financial Reports Management
- Operasi CRUD laporan keuangan untuk mengelola data finansial
- Dukungan untuk RKAP (Rencana Kerja dan Anggaran Perusahaan) dan Realisasi
- Upload massal laporan keuangan melalui Excel untuk efisiensi input data
- Generasi template Excel untuk memudahkan proses bulk upload
- Validasi Excel sebelum upload untuk memastikan data sesuai format
- Mekanisme upsert (update jika sudah ada, insert jika baru) untuk menghindari duplikasi data
- Perbandingan RKAP vs Realisasi YTD (Year-to-Date) untuk analisis performa
- Pelacakan status laporan bulanan per anak perusahaan
- Perhitungan dan validasi rasio keuangan untuk analisis finansial

### Notification System
- Notifikasi dalam aplikasi untuk berbagai event dan aktivitas sistem
- Notifikasi kedaluwarsa dokumen yang dikelompokkan berdasarkan folder
- Notifikasi kedaluwarsa masa jabatan direktur untuk manajemen kepengurusan
- Pesan notifikasi dinamis berdasarkan waktu real-time untuk informasi yang relevan
- Penghitungan jumlah notifikasi yang belum dibaca
- Fungsi mark as read untuk menandai notifikasi yang sudah ditindaklanjuti
- Scheduler otomatis untuk memeriksa item yang akan kedaluwarsa (setiap 24 jam)
- Konfigurasi threshold hari melalui environment variable untuk fleksibilitas

### Development Tools
- Reset data anak perusahaan (hanya untuk superadmin)
- Reset semua laporan keuangan (hanya untuk superadmin)
- Menjalankan company seeder melalui UI (hanya untuk superadmin)
- Pengecekan status seeder untuk memantau proses seeding
- Trigger notifikasi manual untuk keperluan testing (hanya untuk superadmin/administrator)

### Security & Monitoring
- Audit logging komprehensif dengan retention policy untuk pelacakan aktivitas
- Permanent Audit Log: User Activity Logs untuk data penting (Report, Document, Company, User) yang disimpan permanen tanpa retention policy
- Rate limiting per tipe endpoint untuk mencegah abuse dan memastikan stabilitas sistem
- Security headers (CSP, HSTS, XSS protection) untuk meningkatkan keamanan aplikasi
- Validasi dan sanitasi input untuk mencegah serangan injection
- Error logging dengan stack trace untuk debugging dan monitoring
- UI audit log dengan tab terpisah untuk "Audit Logs" dan "User Activity" untuk kemudahan navigasi

## Contributing

### Workflow Development

1. **Buat branch** dari `development` (untuk fitur baru) atau `main` (untuk hotfix)
2. **Develop fitur** dengan mengikuti Clean Architecture pattern
3. **Write tests:** Frontend (Vitest) dan Backend (Go test)
4. **Wajib menjalankan lint dan test sebelum commit:**
   
   **Frontend:**
   ```bash
   cd frontend
   npm run lint          # Lint code untuk memastikan code quality dan consistency
   npm run test:unit     # Run unit tests untuk memastikan tidak ada regression
   ```
   
   **Backend:**
   ```bash
   cd backend
   golangci-lint run     # Lint code untuk memastikan code quality, best practices, dan security
   go test ./...         # Run semua tests untuk memastikan business logic masih benar
   ```
   
   **Atau gunakan Makefile untuk menjalankan semua:**
   ```bash
   make lint             # Lint frontend + backend
   make test             # Test frontend + backend
   ```

5. **Push dan buat PR** ke branch `development`
6. Setelah merge, CI/CD akan otomatis:
   - Run tests (frontend & backend)
   - Build dan deploy ke GCP
   - Setup SSL & Nginx otomatis
   - Verify services running

### Mengapa Wajib Menjalankan Lint dan Test?

**Lint (Code Quality):**
- **Konsistensi kode:** Memastikan semua developer mengikuti style guide yang sama
- **Best practices:** Mendeteksi pola kode yang tidak optimal atau berpotensi error
- **Security:** Mendeteksi vulnerability dan security issues
- **Maintainability:** Kode yang konsisten lebih mudah di-maintain dan di-review

**Test (Business Logic Validation):**
- **Regression prevention:** Memastikan perubahan kode tidak merusak fitur yang sudah ada
- **Business logic verification:** Memastikan perhitungan dan logika bisnis masih benar
- **Confidence:** Memberikan confidence bahwa kode yang diubah masih berfungsi dengan benar
- **Documentation:** Test cases berfungsi sebagai dokumentasi hidup tentang bagaimana fitur seharusnya bekerja

**Penting:** Jangan push kode yang belum di-lint dan di-test, karena:
- CI/CD akan gagal jika ada lint errors atau test failures
- Review process akan lebih lama jika ada banyak issues
- Risiko tinggi untuk introduce bugs ke production

## Dokumentasi Tambahan

- **API Documentation:** http://localhost:8080/swagger/index.html (hanya tersedia di development)
