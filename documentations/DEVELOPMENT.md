# 🚀 Development Guide - Hot Reload & Simple Commands

## Quick Start - Satu Perintah

```bash
# Start semua service dengan satu perintah
make dev

# Atau
./dev.sh

# Atau manual
docker-compose -f docker-compose.dev.yml up --build
```

## Hot Reload Strategy

### Frontend (Vue + Vite)
✅ **Hot reload otomatis** - Save file, langsung terlihat di browser!
- Tidak perlu restart
- Vite HMR bekerja sempurna di Docker

### Backend (Go)
⚠️ **Perlu restart manual** untuk perubahan Go code
- Docker: `make restart` atau `docker-compose -f docker-compose.dev.yml restart backend`
- **Recommended**: Gunakan local development untuk hot reload yang lebih baik

## Development Workflow

### Option 1: Docker (Simple, tapi perlu restart untuk Go)

```bash
# Start semua
make dev

# Edit file Go → Restart backend
make restart

# Edit file Vue → Auto reload (tidak perlu restart)
```

### Option 2: Hybrid (Recommended untuk Development Cepat)

**Terminal 1 - Backend (Local):**
```bash
cd backend
go run main.go
# Auto-reload dengan go run (atau install air untuk true hot reload)
```

**Terminal 2 - Frontend (Docker atau Local):**
```bash
# Docker
docker-compose -f docker-compose.dev.yml up frontend

# Atau Local (lebih cepat)
cd frontend
npm run dev
```

**Keuntungan:**
- ✅ Backend: Hot reload instant (go run)
- ✅ Frontend: Hot reload instant (Vite)
- ✅ Tidak perlu restart Docker

### Option 3: Full Local (Paling Cepat)

**Terminal 1:**
```bash
cd backend
go run main.go
```

**Terminal 2:**
```bash
cd frontend
npm run dev
```

**Keuntungan:**
- ✅ Paling cepat
- ✅ Hot reload sempurna
- ✅ Tidak perlu Docker untuk development

## Makefile Commands

```bash
make dev           # Start all services (foreground)
make up            # Start all services (background)
make down          # Stop all services
make restart       # Restart all services
make logs          # View all logs
make logs-backend  # Backend logs only
make logs-frontend # Frontend logs only
make status        # Check service status
make clean         # Clean everything
make rebuild       # Rebuild and restart
make help          # Show help
```

## Tips untuk Development Cepat

1. **Gunakan Local Development untuk Backend**
   - Lebih cepat
   - Hot reload instant
   - Debug lebih mudah

2. **Gunakan Docker untuk Frontend** (atau local juga OK)
   - Vite HMR bekerja sempurna
   - Tidak perlu restart

3. **Swagger UI**
   - Buka: http://localhost:8080/swagger/index.html
   - Auto-update saat backend restart

4. **Database**
   - SQLite file: `backend/dms.db`
   - Data persist (tidak hilang saat restart)
   - Bisa di-reset dengan hapus file

## Troubleshooting

### Backend tidak reload otomatis
- **Docker**: Perlu restart manual (`make restart`)
- **Local**: `go run main.go` akan auto-reload saat file berubah

### Frontend tidak reload
- Pastikan Vite dev server running
- Cek browser console untuk error
- Hard refresh: Cmd+Shift+R (Mac) atau Ctrl+Shift+R (Windows)

### Port sudah digunakan
```bash
# Cek port
lsof -i :8080
lsof -i :5173

# Kill process
kill -9 <PID>
```

### Database error
```bash
# Reset database
rm backend/dms.db
make restart
```

## Recommended Setup untuk Development

**Best Practice:**
1. Backend: Local (`go run main.go`)
2. Frontend: Local (`npm run dev`)
3. Database: SQLite (otomatis)

**Kenapa?**
- ✅ Paling cepat
- ✅ Hot reload sempurna
- ✅ Debug lebih mudah
- ✅ Tidak perlu Docker overhead

**Docker digunakan untuk:**
- ✅ Testing production build
- ✅ CI/CD
- ✅ Deployment
- ✅ Consistency check

