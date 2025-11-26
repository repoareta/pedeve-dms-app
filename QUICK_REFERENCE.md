# ⚡ Quick Reference - DMS App

Dokumentasi cepat untuk referensi sehari-hari.

---

## 🚀 Quick Start

### Development
```bash
# Start semua service (Docker)
make dev

# Atau manual
docker-compose -f docker-compose.dev.yml up --build
```

### Local Development
```bash
# Backend
cd backend && go run ./cmd/api/main.go

# Frontend
cd frontend && npm run dev
```

---

## 🌐 URLs

| Service | URL |
|---------|-----|
| Frontend (Dev) | http://localhost:5173 |
| Frontend (Prod) | http://localhost:3000 |
| Backend API | http://localhost:8080/api/v1 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Health Check | http://localhost:8080/health |

---

## 🔑 Credentials

### Superadmin (Default)
- **Username**: `superadmin`
- **Password**: `Pedeve123`
- **Role**: `superadmin`

---

## 📦 Tech Stack

### Backend
- Go 1.25
- Fiber v2.52.10
- GORM v1.31.1
- SQLite (dev) / PostgreSQL (prod)

### Frontend
- Vue 3.5.22
- TypeScript 5.9.0
- Vite 7.1.11
- Pinia 3.0.3
- Ant Design Vue 4.2.6

---

## 🔐 Authentication

### JWT Token
- **Expiry**: 24 jam
- **Storage**: httpOnly Cookie (preferred) atau Authorization header
- **Secret**: `JWT_SECRET` env variable

### CSRF Token
- **Endpoint**: `GET /api/v1/csrf-token`
- **Header**: `X-CSRF-Token`
- **Required for**: POST, PUT, DELETE, PATCH

---

## 🗄️ Database

### Development (SQLite)
- **File**: `backend/dms.db`
- **Auto-created**: Ya
- **Reset**: Hapus file `dms.db`

### Production (PostgreSQL)
- **Connection**: `DATABASE_URL` env variable
- **Format**: `postgres://user:pass@host:port/dbname?sslmode=require`

---

## 📡 API Endpoints

### Public
```
GET  /health
GET  /api/v1/csrf-token
POST /api/v1/auth/login
POST /api/v1/auth/register
```

### Protected (Require JWT)
```
GET    /api/v1/auth/profile
POST   /api/v1/auth/logout
GET    /api/v1/documents
POST   /api/v1/documents
...
```

---

## 🛠️ Common Commands

### Backend
```bash
# Run server
go run ./cmd/api/main.go

# Run tests
go test ./...

# Generate Swagger
swag init -g cmd/api/main.go

# Lint
golangci-lint run
```

### Frontend
```bash
# Dev server
npm run dev

# Build
npm run build

# Type check
npm run type-check

# Lint
npm run lint

# Test
npm run test:unit
```

### Docker
```bash
# Start
make dev

# Stop
make down

# Logs
make logs

# Restart
make restart
```

---

## 🔧 Environment Variables

### Backend
```bash
PORT=8080
ENV=development
DATABASE_URL=                    # Kosong = SQLite
JWT_SECRET=your-secret-key-min-32-chars
CGO_ENABLED=1                    # 1 untuk SQLite
```

### Frontend
```bash
VITE_API_URL=http://localhost:8080/api/v1
```

---

## 📚 Documentation Files

- `DOCUMENTATION_BACKEND.md` - Dokumentasi lengkap Backend
- `DOCUMENTATION_FRONTEND.md` - Dokumentasi lengkap Frontend
- `SPESIFIKASI_TEKNIS.md` - Spesifikasi teknis lengkap
- `README.md` - Project overview
- `DEVELOPMENT.md` - Development guide
- `URLS_AND_PORTS.md` - URL & port reference

---

## 🐛 Troubleshooting

### Port Already in Use
```bash
lsof -i :8080
kill -9 <PID>
```

### Database Error
```bash
# Reset SQLite
rm backend/dms.db
make restart
```

### Frontend Can't Connect
- Check `VITE_API_URL`
- Verify backend is running
- Check CORS settings

---

## 📞 Quick Links

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080/api/v1

---

**Last Updated**: 2025-01-XX

