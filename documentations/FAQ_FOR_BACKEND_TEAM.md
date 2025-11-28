# ❓ FAQ untuk Tim Backend Baru

Dokumentasi pertanyaan dan jawaban umum untuk tim backend yang baru bergabung.

---

## 🛠️ Development Setup

### Q1: Bagaimana cara setup development environment?

**A:** 
```bash
# Prerequisites: Docker, Docker Compose, Go 1.25+, Node.js 20+
# Clone repo
git clone https://github.com/repoareta/pedeve-dms-app.git
cd pedeve-dms-app

# Start semua services (PostgreSQL, Backend, Frontend, Vault)
make dev

# Atau manual:
docker-compose -f docker-compose.dev.yml up --build
```

**Services yang running:**
- PostgreSQL: `localhost:5432`
- Backend API: `http://localhost:8080`
- Frontend: `http://localhost:5173`
- Vault: `http://localhost:8200`

---

### Q2: Database apa yang digunakan untuk development?

**A:** 
- **Development:** PostgreSQL via Docker (`postgres:16-alpine`)
- **Production:** Cloud SQL PostgreSQL 16
- **Local fallback:** SQLite (opsional, untuk testing cepat)

**Connection string development:**
```
postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable
```

---

### Q3: Bagaimana cara run backend secara lokal tanpa Docker?

**A:**
```bash
cd backend

# Set environment variables
export DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable"
export PORT=8080
export ENV=development

# Run dengan Air (hot reload)
air

# Atau run langsung
go run ./cmd/api
```

**Note:** Pastikan PostgreSQL running (via Docker atau local installation).

---

### Q4: Bagaimana cara run seeder untuk sample data?

**A:**
```bash
# Via Makefile (paling mudah)
make seed-companies

# Atau manual
cd backend
DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable" go run ./cmd/seed-companies
```

**Seeder akan membuat:**
- 1 Holding company (Pedeve Pertamina)
- 10 Subsidiaries dengan hierarchy 3 layer
- 11 Admin users (password: `admin123`)

---

### Q5: Bagaimana struktur project backend?

**A:**
```
backend/
├── cmd/                    # Application entry points
│   ├── api/               # Main API server
│   ├── seed-companies/    # Database seeder
│   └── ...
├── internal/
│   ├── domain/            # Domain models & interfaces
│   ├── infrastructure/    # External dependencies (DB, Logger, Secrets, Storage)
│   ├── repository/         # Data access layer
│   ├── usecase/           # Business logic
│   ├── delivery/          # HTTP handlers
│   └── middleware/        # HTTP middleware
├── go.mod
└── go.sum
```

**Architecture:** Clean Architecture dengan separation of concerns.

---

## 🚀 Deployment

### Q6: Bagaimana proses deployment ke production?

**A:**
1. **Push ke branch `development`** → Trigger GitHub Actions
2. **Build & Push:** Build Docker image, push ke GHCR
3. **Deploy:** 
   - Copy image ke VM
   - Stop container lama
   - Start container baru dengan env vars
   - Health check

**Semua otomatis via CI/CD, tidak perlu manual deployment.**

---

### Q7: Dimana aplikasi di-deploy?

**A:**
- **Frontend:** GCP Compute Engine VM (`frontend-dev`, IP: `34.128.123.1`)
- **Backend:** GCP Compute Engine VM (`backend-dev`, IP: `34.101.49.147`)
- **Database:** GCP Cloud SQL PostgreSQL (`postgres-dev`)
- **Storage:** GCP Cloud Storage (`pedeve-dev-bucket`)
- **Secrets:** GCP Secret Manager

**Domains:**
- Frontend: `https://pedeve-dev.aretaamany.com`
- Backend: `https://api-pedeve-dev.aretaamany.com`

---

### Q8: Bagaimana cara deploy manual jika CI/CD gagal?

**A:**
```bash
# 1. Build image lokal
cd backend
docker build -t dms-backend:local .

# 2. Save image
docker save dms-backend:local -o backend-image.tar

# 3. Copy ke VM
gcloud compute scp --zone=asia-southeast2-a backend-image.tar backend-dev:~/

# 4. SSH dan deploy
gcloud compute ssh --zone=asia-southeast2-a backend-dev
# Di dalam VM:
sudo docker load -i ~/backend-image.tar
# Restart container dengan script deploy-backend-vm.sh
```

**Atau gunakan script:** `scripts/deploy-backend-vm.sh`

---

### Q9: Bagaimana cara akses database production?

**A:**
**Via Cloud SQL Proxy (dari backend VM):**
```bash
# Cloud SQL Proxy sudah running di backend VM
# Connect via:
psql "postgres://pedeve_user_db:\$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"
```

**Atau via GCP Console:**
- Cloud SQL → Connect → Cloud SQL Proxy

---

## 🔐 Security & Secrets

### Q10: Bagaimana management secrets?

**A:**
**Development:**
- Vault (local Docker) atau Environment Variables
- Default values untuk development

**Production:**
- GCP Secret Manager (primary)
- Secrets diambil saat container start via deployment script
- Tidak ada hardcoded secrets di code

**Secrets yang ada:**
- `db_password` - Database password
- `jwt_secret` - JWT token secret
- `encryption_key` - Encryption key (optional)

---

### Q11: Bagaimana authentication bekerja?

**A:**
- **JWT Token** dalam httpOnly cookie (`auth_token`)
- **Cookie settings:**
  - `HttpOnly: true` (prevent XSS)
  - `Secure: true` (HTTPS only)
  - `SameSite: None` (cross-site support)
- **CSRF Protection:** Double-submit cookie pattern
- **2FA Support:** TOTP-based (optional)

---

## 🗄️ Database

### Q12: Bagaimana schema migration?

**A:**
- **Auto-migration:** GORM AutoMigrate saat aplikasi start
- **Manual migration:** Tidak ada migration files, semua via GORM models
- **Models location:** `backend/internal/domain/models.go`

**Untuk update schema:**
1. Update model di `models.go`
2. Restart aplikasi → AutoMigrate akan update schema

---

### Q13: Bagaimana company hierarchy bekerja?

**A:**
- **Single root holding:** Hanya 1 company dengan `parent_id = NULL` (level 0)
- **Hierarchy:** Parent-child relationship dengan `level` field
- **Level calculation:** Otomatis berdasarkan parent
- **Access control:** User hanya bisa akses company mereka dan descendants

**Business rules:**
- Tidak bisa create 2 root holdings
- Level otomatis dihitung dari parent
- Update parent akan update semua descendants

---

## 🔄 CI/CD

### Q14: Bagaimana CI/CD pipeline bekerja?

**A:**
**Workflow:** `.github/workflows/ci-cd.yml`

**Jobs:**
1. **build-and-push:**
   - Build backend Docker image
   - Build frontend static files
   - Push image ke GHCR
   - Upload frontend artifacts

2. **deploy-gcp:**
   - Authenticate via Workload Identity Federation
   - Deploy backend ke VM
   - Deploy frontend ke VM
   - Health checks

**Trigger:** Push ke branch `development`

---

### Q15: Bagaimana authentication GitHub Actions ke GCP?

**A:**
- **Method:** Workload Identity Federation (WIF) - **Tidak pakai JSON keys**
- **Provider:** OIDC provider untuk GitHub
- **Service Account:** `github-actions-deployer@pedeve-pertamina-dms.iam.gserviceaccount.com`
- **Permissions:** Compute Admin, OS Login, Storage Admin, Cloud SQL Client

**Setup sudah dilakukan, tidak perlu konfigurasi lagi.**

---

## 🐛 Troubleshooting

### Q16: Backend tidak bisa connect ke database?

**A:**
**Cek:**
1. Cloud SQL Proxy running: `ps aux | grep cloud-sql-proxy`
2. Database password benar: `gcloud secrets versions access latest --secret=db_password`
3. Connection string format: `postgres://user:pass@127.0.0.1:5432/db?sslmode=disable`
4. Container network: `--network host` untuk akses Cloud SQL Proxy

**Fix:**
```bash
# Restart Cloud SQL Proxy
sudo systemctl restart cloud-sql-proxy

# Test connection
psql "postgres://..." -c "SELECT 1;"
```

---

### Q17: Frontend tidak bisa akses backend API (CORS error)?

**A:**
**Cek CORS_ORIGIN di container:**
```bash
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN
```

**Expected:**
```
CORS_ORIGIN=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com
```

**Fix:** Restart container dengan CORS_ORIGIN yang benar (lihat `scripts/deploy-backend-vm.sh`)

---

### Q18: Error 429 Too Many Requests?

**A:**
**Development:** Set `DISABLE_RATE_LIMIT=true` di environment variables

**Production:** Adjust rate limit config atau disable jika perlu

**Fix:**
```bash
# Restart container dengan DISABLE_RATE_LIMIT=true
# (lihat scripts/deploy-backend-vm.sh)
```

---

### Q19: Container crash setelah start?

**A:**
**Cek logs:**
```bash
sudo docker logs dms-backend-prod --tail 100
```

**Common causes:**
- Database connection failed → Cek Cloud SQL Proxy
- Missing environment variables → Cek deployment script
- Port conflict → Cek port 8080

---

### Q20: HTTPS tidak bekerja (port 443 tidak listening)?

**A:**
**Cek:**
1. SSL certificate ada: `sudo certbot certificates`
2. Nginx config punya HTTPS block: `sudo cat /etc/nginx/sites-available/backend-api | grep "listen 443"`
3. Firewall rule `allow-https` aktif
4. VM punya tag `https-server`

**Fix:** Update Nginx config dengan HTTPS block (lihat `FIX_HTTPS_NOT_LISTENING.md`)

---

## 📦 Dependencies & Tools

### Q21: Dependencies apa saja yang digunakan?

**A:**
**Backend:**
- **Framework:** Fiber v2 (HTTP framework)
- **ORM:** GORM (database)
- **Logger:** Zap (structured logging)
- **Validation:** go-playground/validator
- **JWT:** Custom implementation
- **Storage:** GCP Cloud Storage client
- **Secrets:** GCP Secret Manager client

**Full list:** `backend/go.mod`

---

### Q22: Bagaimana cara menambah dependency baru?

**A:**
```bash
cd backend
go get github.com/package/name
go mod tidy
```

**Commit:** `go.mod` dan `go.sum` harus di-commit.

---

### Q23: Bagaimana testing?

**A:**
**Saat ini:**
- Manual testing via Swagger UI atau Postman
- Integration testing via seeder data

**Swagger UI:** `https://api-pedeve-dev.aretaamany.com/swagger/index.html`

**Untuk unit tests:** Bisa ditambahkan di `backend/internal/.../..._test.go`

---

## 🏗️ Architecture

### Q24: Mengapa menggunakan Clean Architecture?

**A:**
- **Separation of concerns:** Business logic terpisah dari infrastructure
- **Testability:** Mudah untuk unit test
- **Maintainability:** Code lebih terorganisir
- **Flexibility:** Mudah ganti database/storage tanpa ubah business logic

**Layers:**
1. **Domain:** Models & interfaces (business rules)
2. **Repository:** Data access (database operations)
3. **Usecase:** Business logic
4. **Delivery:** HTTP handlers (API endpoints)
5. **Infrastructure:** External dependencies (DB, Logger, etc)

---

### Q25: Bagaimana flow request dari client ke database?

**A:**
```
Client Request
  ↓
HTTP Handler (delivery/http/)
  ↓
Usecase (usecase/)
  ↓
Repository (repository/)
  ↓
Database (infrastructure/database/)
```

**Example:**
1. Client → `POST /api/v1/companies`
2. `CompanyHandler.CreateCompany` → `CompanyUsecase.CreateCompany`
3. `CompanyUsecase` → `CompanyRepository.Create`
4. `CompanyRepository` → Database (via GORM)

---

## 🔧 Configuration

### Q26: Environment variables apa saja yang digunakan?

**A:**
**Development:**
- `DATABASE_URL` - Database connection
- `PORT` - Server port (default: 8080)
- `ENV` - Environment mode (development/production)
- `VAULT_ADDR` - Vault address (optional)

**Production:**
- `DATABASE_URL` - From Secret Manager
- `JWT_SECRET` - From Secret Manager
- `ENCRYPTION_KEY` - From Secret Manager (optional)
- `GCP_PROJECT_ID` - GCP project ID
- `GCP_STORAGE_ENABLED` - Enable GCP Storage
- `GCP_STORAGE_BUCKET` - Storage bucket name
- `CORS_ORIGIN` - Allowed CORS origins
- `DISABLE_RATE_LIMIT` - Disable rate limiting (dev)

**Full list:** Lihat `DEPLOYMENT_DOCUMENTATION.md`

---

### Q27: Bagaimana logging bekerja?

**A:**
- **Logger:** Zap (structured logging)
- **Level:** Info (default), bisa diubah via config
- **Output:** Console (development), bisa diubah ke file
- **Audit Logging:** Semua user actions dan errors dicatat di database

**Log format:**
```json
{
  "level": "info",
  "ts": 1234567890,
  "caller": "file.go:123",
  "msg": "Message",
  "key": "value"
}
```

---

## 📁 File Storage

### Q28: Bagaimana file upload bekerja?

**A:**
**Development:**
- Local filesystem: `backend/uploads/`
- URL: `/uploads/{path}/{filename}`

**Production:**
- GCP Cloud Storage: `pedeve-dev-bucket`
- URL: `https://storage.googleapis.com/pedeve-dev-bucket/{path}/{filename}`

**Auto-detect:** Backend otomatis pilih storage berdasarkan `GCP_STORAGE_ENABLED`

---

### Q29: Bagaimana cara upload file?

**A:**
**API Endpoint:**
```
POST /api/v1/upload/logo
Content-Type: multipart/form-data
Body: file (binary)
```

**Response:**
```json
{
  "url": "https://storage.googleapis.com/.../logo.png",
  "filename": "logo.png"
}
```

---

## 🔄 Git Workflow

### Q30: Branch strategy?

**A:**
- **`main`:** Production (stable)
- **`development`:** Development (auto-deploy ke GCP dev)
- **Feature branches:** Untuk development fitur baru

**Workflow:**
1. Create feature branch dari `development`
2. Develop & test
3. Merge ke `development` → Auto-deploy
4. Merge ke `main` untuk production release

---

### Q31: Commit message convention?

**A:**
**Format:** Bahasa Indonesia, deskriptif

**Examples:**
- `Fix: perbaiki error CORS di production`
- `Tambah: fitur assign user ke company`
- `Update: dokumentasi deployment`

**Tidak ada strict convention, tapi harus jelas dan deskriptif.**

---

## 🚨 Common Issues & Solutions

### Q32: Go module path error?

**A:**
**Error:** `no required module provides package github.com/Fajarriswandi/...`

**Fix:**
```bash
# Update module path di go.mod
# Pastikan semua import menggunakan path yang benar:
github.com/repoareta/pedeve-dms-app/backend/...
```

---

### Q33: Database connection timeout?

**A:**
**Cek:**
1. Cloud SQL Proxy running
2. Database instance active
3. Authorized networks correct
4. Connection string format correct

**Fix:**
```bash
# Restart Cloud SQL Proxy
sudo systemctl restart cloud-sql-proxy

# Test connection
psql "postgres://..." -c "SELECT 1;"
```

---

### Q34: Docker build error?

**A:**
**Common causes:**
- Go version mismatch
- Missing dependencies
- Build context issues

**Fix:**
```bash
# Clean build
docker system prune -a

# Rebuild
docker build -t dms-backend:test ./backend
```

---

## 📚 Documentation

### Q35: Dimana dokumentasi lengkap?

**A:**
- **Deployment:** `DEPLOYMENT_DOCUMENTATION.md`
- **API Docs:** Swagger UI (`/swagger/index.html`)
- **Backend Docs:** `DOCUMENTATION_BACKEND.md`
- **Frontend Docs:** `DOCUMENTATION_FRONTEND.md`
- **Troubleshooting:** `TROUBLESHOOT_DEPLOYMENT_FAILURE.md`

---

### Q36: Bagaimana cara generate Swagger docs?

**A:**
**Automatic:** Swagger docs auto-generated saat build

**Manual:**
```bash
cd backend
swag init -g cmd/api/main.go
```

**Access:** `https://api-pedeve-dev.aretaamany.com/swagger/index.html`

---

## 🎯 Best Practices

### Q37: Coding standards?

**A:**
- **Linter:** `golangci-lint` (runs in CI/CD)
- **Format:** `go fmt`
- **Error handling:** Always check errors
- **Logging:** Use structured logging (Zap)
- **Type safety:** Avoid `any`, use specific types

**Lint errors akan block deployment.**

---

### Q38: Bagaimana menambah endpoint baru?

**A:**
1. **Define model** (jika perlu) di `internal/domain/models.go`
2. **Create repository method** di `internal/repository/`
3. **Create usecase method** di `internal/usecase/`
4. **Create handler** di `internal/delivery/http/`
5. **Register route** di `cmd/api/main.go`
6. **Add Swagger docs** (comments di handler)

**Example flow:** Lihat existing endpoints sebagai reference.

---

### Q39: Bagaimana menambah middleware?

**A:**
1. **Create middleware** di `internal/middleware/`
2. **Register** di `cmd/api/main.go` dengan `app.Use()`

**Existing middlewares:**
- CORS
- Rate Limiting
- JWT Auth
- CSRF Protection
- Security Headers
- Error Handler
- Logger

---

## 🔗 Quick Reference

### Q40: Command cheat sheet?

**A:**
```bash
# Development
make dev              # Start all services
make down             # Stop all services
make logs             # View logs
make seed-companies   # Seed sample data

# Database
psql "postgres://..." # Connect to database

# Docker
docker ps             # List containers
docker logs <name>    # View logs
docker exec -it <name> sh  # Enter container

# Deployment
git push origin development  # Trigger auto-deploy
```

---

## 📞 Support & Resources

### Q41: Siapa yang bisa ditanya jika ada masalah?

**A:**
- **Code issues:** Check existing code & documentation
- **Deployment issues:** Check `DEPLOYMENT_DOCUMENTATION.md` dan troubleshooting guides
- **Architecture questions:** Check `DOCUMENTATION_BACKEND.md`

**Resources:**
- GitHub Issues (jika ada)
- Internal documentation
- Swagger API docs

---

### Q42: Bagaimana cara contribute code?

**A:**
1. **Fork atau create feature branch** dari `development`
2. **Develop & test** locally
3. **Ensure linting passes:** `golangci-lint run`
4. **Commit** dengan message jelas
5. **Push** ke feature branch
6. **Create PR** ke `development`
7. **Review & merge**

**Important:** Jangan push langsung ke `development` atau `main` tanpa review.

---

**Last Updated:** 2025-11-28  
**Environment:** Development & Production  
**Status:** ✅ All configurations documented

