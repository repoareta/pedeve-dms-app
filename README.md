# DMS App - Document Management System

Aplikasi Document Management System dengan stack:
- **Frontend**: Vue 3 + TypeScript + Vite
- **Backend**: Go (Golang)
- **CI/CD**: GitHub Actions dengan Docker

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (untuk development frontend)
- Go 1.23+ (untuk development backend)

### Development Setup

#### 🚀 Quick Start - Satu Perintah untuk Semua

```bash
# Cara termudah - run semua service dengan hot reload
make dev

# Atau menggunakan script
./dev.sh

# Atau manual
docker-compose -f docker-compose.dev.yml up --build
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
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## 📁 Project Structure

```
dms-app/
├── backend/          # Go backend API
│   ├── main.go      # Entry point
│   ├── go.mod       # Go dependencies
│   └── Dockerfile   # Production Docker image
├── frontend/         # Vue 3 frontend
│   ├── src/         # Source code
│   ├── package.json # Node dependencies
│   └── Dockerfile   # Production Docker image
├── .github/
│   └── workflows/   # CI/CD pipelines
└── docker-compose.yml # Local development setup
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
go run main.go          # Run server (local, tanpa Docker)
go test ./...           # Run tests
golangci-lint run       # Lint code

# Generate Swagger docs (setelah update annotations)
go run github.com/swaggo/swag/cmd/swag@latest init
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

### Health Check

```bash
# Backend health check
curl http://localhost:8080/health

# Expected response: {"status": "OK", "service": "dms-backend"}
```

### API Endpoints

**Documents API:**
- `GET /api/v1/documents` - Get all documents
- `GET /api/v1/documents/{id}` - Get document by ID
- `POST /api/v1/documents` - Create new document
- `PUT /api/v1/documents/{id}` - Update document
- `DELETE /api/v1/documents/{id}` - Delete document

**Test dengan curl:**
```bash
# Get all documents
curl http://localhost:8080/api/v1/documents

# Get single document
curl http://localhost:8080/api/v1/documents/1

# Health check
curl http://localhost:8080/health
```

## 📦 Port Configuration

- **Frontend (Dev)**: 5173
- **Frontend (Prod)**: 3000
- **Backend API**: 8080

**Note**: Pastikan port-port ini tidak digunakan oleh aplikasi lain.

## 🛠️ Troubleshooting

### Port sudah digunakan
```bash
# Cek port yang digunakan
lsof -i :8080
lsof -i :5173

# Atau ubah port di docker-compose.yml
```

### Docker build error
```bash
# Clean build
docker-compose down
docker system prune -f
docker-compose up --build
```

### Frontend tidak connect ke backend
- Pastikan `VITE_API_URL` di frontend sesuai dengan backend URL
- Cek CORS settings di backend jika diperlukan

## 📚 Tech Stack

- **Frontend**: Vue 3, TypeScript, Vite, Pinia, Vue Router
- **Backend**: Go 1.23, Chi Router, Swagger/OpenAPI
- **Container**: Docker, Docker Compose
- **CI/CD**: GitHub Actions
- **Security**: Trivy Scanner
- **API Docs**: Swagger UI

## 🤝 Contributing

1. Buat branch dari `main`
2. Develop fitur
3. Test & lint
4. Push dan buat PR
5. Setelah merge, CI/CD akan otomatis build

## 📄 License

[Your License Here]

