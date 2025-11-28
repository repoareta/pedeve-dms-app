# 📋 DMS App - URL & Port Reference

Dokumentasi lengkap semua URL dan port yang digunakan dalam project DMS App.

## 🌐 Development URLs (Local)

### Frontend
- **Development Server**: http://localhost:5173
- **Production Build**: http://localhost:3000
- **Vue DevTools**: http://localhost:5173/__devtools__/

### Backend API
- **API Base URL**: http://localhost:8080
- **API v1**: http://localhost:8080/api/v1
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Swagger JSON**: http://localhost:8080/swagger/doc.json
- **Health Check**: http://localhost:8080/health
- **Root/Info**: http://localhost:8080/

## 🔌 Port Configuration

### Development Mode
| Service | Port | Description |
|---------|------|-------------|
| Frontend (Vite) | **5173** | Vue development server dengan hot reload |
| Backend (Go) | **8080** | Go HTTP server |
| Vue DevTools | **5173** | Built-in Vue DevTools (same as frontend) |

### Production Mode
| Service | Port | Description |
|---------|------|-------------|
| Frontend (Nginx) | **3000** | Production build dengan Nginx |
| Backend (Go) | **8080** | Go HTTP server |

### Docker Container Ports
| Container | Internal Port | External Port | Description |
|-----------|---------------|---------------|-------------|
| `dms-frontend-dev` | 5173 | 5173 | Frontend development |
| `dms-backend-dev` | 8080 | 8080 | Backend API |
| `dms-frontend` (prod) | 80 | 3000 | Frontend production |
| `dms-backend` (prod) | 8080 | 8080 | Backend production |

## 📡 API Endpoints

### General Endpoints
```
GET  /                    → API information
GET  /health              → Health check
GET  /api/v1              → API version info
```

### Documents API
```
GET    /api/v1/documents           → Get all documents
GET    /api/v1/documents/{id}      → Get document by ID
POST   /api/v1/documents           → Create new document
PUT    /api/v1/documents/{id}      → Update document
DELETE /api/v1/documents/{id}      → Delete document
```

### Swagger Documentation
```
GET  /swagger/index.html  → Swagger UI
GET  /swagger/doc.json    → Swagger JSON spec
GET  /swagger/*           → Swagger static files
```

## 🐳 Docker Compose Services

### Development (`docker-compose.dev.yml`)
```yaml
Services:
  - backend:  localhost:8080
  - frontend: localhost:5173
```

### Production (`docker-compose.yml`)
```yaml
Services:
  - backend:  localhost:8080
  - frontend: localhost:3000
```

## 🚀 CI/CD & Registry

### GitHub Container Registry (GHCR)
```
Registry: ghcr.io
Owner: fajarriswandi (lowercase)
Images:
  - ghcr.io/fajarriswandi/dms-frontend:latest
  - ghcr.io/fajarriswandi/dms-frontend:v1.0.0
  - ghcr.io/fajarriswandi/dms-backend:latest
  - ghcr.io/fajarriswandi/dms-backend:v1.0.0
```

### GitHub Releases
```
Releases URL: https://github.com/{owner}/{repo}/releases
Auto-generated saat push tag: v1.0.0, v2.1.3, etc.
```

## 🔍 Quick Access Commands

### Test Backend
```bash
# Health check
curl http://localhost:8080/health

# Get all documents
curl http://localhost:8080/api/v1/documents

# Get single document
curl http://localhost:8080/api/v1/documents/1
```

### Test Frontend
```bash
# Open in browser
open http://localhost:5173  # macOS
xdg-open http://localhost:5173  # Linux
start http://localhost:5173  # Windows
```

### Docker Commands
```bash
# View running containers
docker ps

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Check port usage
lsof -i :8080
lsof -i :5173
lsof -i :3000
```

## 📝 Environment Variables

### Frontend
```env
VITE_API_URL=http://localhost:8080
NODE_ENV=development
```

### Backend
```env
PORT=8080
ENV=development
```

## ⚠️ Port Conflicts Check

### Common Ports to Avoid
- **3000**: Used by many Node.js apps (conflict dengan frontend prod)
- **5173**: Vite default port (frontend dev)
- **8080**: Common for APIs (backend)
- **5432**: PostgreSQL default
- **3306**: MySQL default

### Check Port Availability
```bash
# Check if port is in use
lsof -i :8080
lsof -i :5173
lsof -i :3000

# Or use netstat
netstat -an | grep LISTEN | grep -E ':(8080|5173|3000)'
```

## 🔗 Quick Links Summary

### Development
- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health

### Production (Docker)
- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

### Documentation
- README: `/README.md`
- This file: `/URLS_AND_PORTS.md`
- Swagger UI: http://localhost:8080/swagger/index.html

## 📌 Important Notes

1. **Port 3000** mungkin sudah digunakan oleh aplikasi lain (tidak masalah untuk dev)
2. **Port 5173** adalah default Vite dev server
3. **Port 8080** adalah backend API port
4. **Swagger UI** hanya tersedia saat backend running
5. **CORS** sudah dikonfigurasi untuk `localhost:5173` dan `localhost:3000`
6. **Hot reload** aktif di development mode (docker-compose.dev.yml)

## 🆘 Troubleshooting

### Port Already in Use
```bash
# Find process using port
lsof -i :8080
lsof -i :5173

# Kill process (replace PID)
kill -9 <PID>

# Or change port in docker-compose.yml
```

### Cannot Access Swagger
- Pastikan backend sudah running
- Cek: http://localhost:8080/health
- Pastikan tidak ada firewall blocking

### Frontend Cannot Connect to Backend
- Pastikan backend running di port 8080
- Cek `VITE_API_URL` di frontend
- Cek CORS settings di backend

