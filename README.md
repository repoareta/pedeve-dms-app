# Pedeve DMS App

Document Management System untuk manajemen dokumen dan perusahaan dengan hierarki multi-level.

## Deskripsi

Pedeve DMS App adalah aplikasi manajemen dokumen yang dirancang untuk mengelola dokumen dan data perusahaan dalam struktur hierarki multi-level. Aplikasi ini menyediakan fitur lengkap untuk manajemen perusahaan, dokumen, laporan keuangan, pengguna, dan sistem notifikasi.

## Environment Production

| Layanan | URL |
|---|---|
| Frontend | https://dms.pertamina-pedeve.co.id |
| Backend API | https://api-reports.pertamina-pedeve.co.id |
| Health Check | https://api-reports.pertamina-pedeve.co.id/health |

**Catatan production:**
- Swagger UI **dinonaktifkan** (`/swagger/*` → 404)
- Route `/development/*` **tidak diregister**
- Static `/uploads` **dinonaktifkan**
- File serving (`/api/v1/files/*`) memerlukan JWT (httpOnly cookie) + RBAC
- Database: Google Cloud SQL (PostgreSQL) via Cloud SQL Auth Proxy — **tidak dibuka ke publik**
- Auth: JWT di httpOnly cookie (`Secure`, `SameSite=None`); frontend tidak menyimpan JWT di `localStorage`

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

**Akses lokal (development):**
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html *(hanya non-production)*
- Health Check: http://localhost:8080/health
- API Base: http://localhost:8080/api/v1

**Akses production:** lihat tabel [Environment Production](#environment-production) di atas.

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
│   └── workflows/             # CI/CD pipelines (ci-cd.yml, repair-frontend-prod.yml)
├── scripts/                   # Deployment & SSL/Nginx scripts
│   ├── deploy-prod-fast.sh    # Deploy production cepat via terminal (opsional)
│   ├── deploy-backend-vm.sh   # Deploy container backend di VM
│   └── lib/resolve-domain.sh  # Resolver DOMAIN (wajib untuk production)
├── docs/                      # User guideline + checklist keamanan
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

### Branch & Target Environment

| Branch | Environment | Frontend | Backend |
|---|---|---|---|
| `main` | **Production** | `dms.pertamina-pedeve.co.id` | `api-reports.pertamina-pedeve.co.id` |
| `development` | Development | `pedeve-dev.aretaamany.com` | `api-pedeve-dev.aretaamany.com` |

### CI/CD Pipeline

Pipeline otomatis berjalan saat:
- Push ke branch `main` (deploy production) atau `development` (deploy development)
- Push tag versi (v1.0.0, v2.1.3, dll)
- Manual trigger via `workflow_dispatch`

**Fitur CI/CD:**
- Lint & Test: Frontend (ESLint + Vitest + type-check) & Backend (golangci-lint + Go test)
- Security Scan: Trivy vulnerability scanner untuk Docker images
- Build: Docker image backend + static files frontend
- Deploy: Otomatis ke GCP VM sesuai branch
- Registry: GitHub Container Registry (`ghcr.io`)
- Versioning / Changelog / GitHub Release (saat push tag)
- Health check HTTPS (frontend & backend) setelah deploy production

### Deploy Cepat via Terminal (Opsional)

Selain CI/CD (~25 menit), production bisa di-deploy langsung dari mesin lokal:

```bash
export GCP_PROJECT_ID_PROD=<gcp-project-id-prod>
./scripts/deploy-prod-fast.sh              # backend + frontend
./scripts/deploy-prod-fast.sh backend      # backend saja
./scripts/deploy-prod-fast.sh frontend     # frontend saja
```

Prasyarat: `gcloud` terautentikasi dengan akses ke VM `backend-prod-1` / `frontend-prod-2`, Docker, dan Node.js 20+.

### Deployment Automation

Setelah deployment selesai, layanan langsung siap tanpa langkah manual tambahan.

**SSL Certificate Management:**
- Deteksi sertifikat yang sudah ada; generate via Certbot (Let's Encrypt) jika belum
- Idempotent; auto-renewal via Certbot timer
- Production domain wajib di-set (`DOMAIN` / `DEPLOY_TARGET=prod`) — tidak boleh silent-fallback ke domain development

**Nginx Configuration:**
- Setup HTTPS bila sertifikat ada; fallback HTTP jika belum
- Preserve config yang sudah benar; reload setelah update
- Frontend health check memverifikasi HTTP **dan** HTTPS

**Service Management:**
- Backend: Docker container `dms-backend-prod` (non-root user, resource limits memory/CPU/pids)
- Frontend: Nginx static files di `/var/www/html`
- Health check + retry setelah deploy

**Deployment Flow:**
1. Build & Test
2. Deploy image backend / static frontend ke VM
3. Setup SSL (jika perlu)
4. Setup Nginx
5. Ensure services running
6. Health check
7. Ready

### Infrastruktur Production (GCP)

- **Frontend VM:** `frontend-prod-2` (Nginx + static build)
- **Backend VM:** `backend-prod-1` (Docker + Nginx reverse proxy)
- **Database:** Cloud SQL PostgreSQL (`db_prod_pedeve`) via Cloud SQL Proxy di `127.0.0.1:5432`
- **Storage:** GCS bucket `pedeve-prod-bucket`
- **Secrets:** GCP Secret Manager (`db_password_prod`, `jwt_secret`, `encryption_key`, dll.)
- **Region/Zone:** `asia-southeast2-a`

### Referensi Operasional Production

Nilai rahasia (password, key, token) disimpan di GCP Secret Manager / GitHub Actions secrets — **jangan** ditulis di repo.

| Item | Keterangan |
|---|---|
| GCP project ID (prod & dev) | Project tempat VM, Cloud SQL, Secret Manager, GCS |
| IAM / akun admin GCP | Akses Compute, Secret Manager, Cloud SQL |
| GitHub Actions secrets | `GCP_PROJECT_ID_PROD`, WIF, service account, Sonar, dll. |
| VM | `backend-prod-1`, `frontend-prod-2` (zone `asia-southeast2-a`) |
| Domain / DNS | `dms.pertamina-pedeve.co.id`, `api-reports.pertamina-pedeve.co.id` |
| DB | Nama DB `db_prod_pedeve`, user `pedeve_user_db_prod` (password di Secret Manager) |
| Akun aplikasi | Administrator production (`pedeve@pertamina-pedeve.co.id`) |
| Checklist keamanan | `docs/PENTEST_SECURITY_CHECKLIST.md` |

Database **tidak punya URL publik**. Integrasi data eksternal (mis. Data Lake) dilakukan via export / pipeline terkontrol, bukan membuka port 5432 ke internet.

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
- `GET /api/v1/files/*` - Serve files (GCP Storage / local) — **wajib autentikasi JWT + RBAC** (bucket whitelist: `logos`, `documents`)

**Audit Logs:**
- `GET /api/v1/audit-logs` - Get audit logs (retention: 90 hari user actions, 30 hari technical errors)
- `GET /api/v1/audit-logs/stats` - Get audit log statistics
- `GET /api/v1/user-activity-logs` - Get user activity logs (permanent: report, document, company, user)

**Notifications:**
- `GET /api/v1/notifications` - Get all notifications
- `GET /api/v1/notifications/unread-count` - Get unread notification count
- `PUT /api/v1/notifications/{id}/read` - Mark notification as read
- `PUT /api/v1/notifications/read-all` - Mark all notifications as read

**Development tools (Superadmin Only — hanya non-production):**
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
- Rate Limiting: ~100 req/s (general), ~5 req/min (auth endpoints); aktif di production
- Security Headers: HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy
- 2FA Support: TOTP-based dengan backup codes (secret terenkripsi AES-256-GCM)
- JWT Security: httpOnly cookie saja (tidak disimpan di `localStorage`)
- XSS mitigasi frontend: DOMPurify / `escapeHtml` pada render `v-html` sensitif
- File access RBAC + bucket whitelist
- Docker production: non-root user (`appuser`) + resource limits (`--memory`, `--cpus`, `--pids-limit`)
- Transparent Data Encryption (TDE):
  - SQLite: SQLCipher (development)
  - PostgreSQL: enkripsi at-rest Cloud SQL (GCP managed)
  - Key management via GCP Secret Manager / environment variables
- Audit Logging:
  - User actions (retention 90 hari) + technical errors (30 hari)
  - Permanent User Activity Logs untuk data penting (Report, Document, Company, User)
- Input Validation & sanitization; password bcrypt
- Production-Safe Logging di frontend

### Infrastructure
- Container: Docker, Docker Compose
- CI/CD: GitHub Actions (lint, test, Trivy, build, deploy)
- Deployment: GCP Compute Engine VM + automated SSL & Nginx
- Database: Cloud SQL PostgreSQL + Cloud SQL Auth Proxy (private)
- Web Server: Nginx (HTTPS via Let's Encrypt)
- Storage: Google Cloud Storage
- Secrets: GCP Secret Manager
- Security Scan: Trivy
- API Docs: Swagger UI **hanya non-production**

## Fitur Utama

### Authentication & Authorization
- Autentikasi JWT via httpOnly cookie (`withCredentials`); UI state user di `localStorage` (bukan token)
- Two-Factor Authentication (2FA) TOTP + backup codes
- Role-Based Access Control (RBAC): superadmin, administrator, admin, manager, staff
- Kontrol akses berbasis hierarki perusahaan
- Proteksi CSRF untuk request yang mengubah state (POST/PUT/DELETE/PATCH)

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
- Kategorisasi dokumen menggunakan struktur folder
- Upload/penyimpanan via GCP Storage atau lokal; preview via authenticated blob URL
- Batch upload (PDF, gambar, dokumen Office)
- Validasi ukuran: gambar maksimal 10MB; dokumen Office/PDF mengikuti kebijakan upload
- Pelacakan tanggal kedaluwarsa + status dokumen
- Preview gambar/PDF; preview Office (docx/xlsx) dengan sanitasi HTML

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

### Development Tools *(tidak tersedia di production)*
- Reset data anak perusahaan / laporan keuangan (superadmin)
- Menjalankan company seeder melalui UI + cek status seeder
- Trigger notifikasi manual untuk testing (superadmin/administrator)

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
6. Setelah merge ke `development`, CI/CD deploy ke environment **development**
7. Untuk production: merge/cherry-pick ke `main` (setelah review) — CI/CD deploy ke **production**
8. Jangan push langsung ke `main` tanpa konfirmasi / review tim

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

- **API Documentation (local/dev):** http://localhost:8080/swagger/index.html *(tidak tersedia di production)*
- **User Guideline:** `docs/` (VitePress) — di-build ke `frontend/public/user-guideline/`
- **Pentest Security Checklist:** `docs/PENTEST_SECURITY_CHECKLIST.md`
- **UAT Document:** `docs/UAT_DOCUMENT.md`
- **UAT Coverage Analysis:** `docs/UAT_COVERAGE_ANALYSIS.md`
- **Production:** https://dms.pertamina-pedeve.co.id · https://api-reports.pertamina-pedeve.co.id
