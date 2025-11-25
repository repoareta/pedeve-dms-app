# 🔥 Hot Reload Backend dengan Air

## Overview

Backend Go saat ini **mendukung hot reload otomatis** menggunakan [Air](https://github.com/air-verse/air). Saat Anda mengubah file `.go`, Air akan otomatis:
1. Detect perubahan file
2. Rebuild aplikasi
3. Restart server

**Tidak perlu restart manual!** ✨

## Status Hot Reload

### ✅ Dengan Air (Recommended)
- **Hot reload otomatis** saat file `.go` berubah
- Tidak perlu restart manual
- Perubahan langsung terlihat

### ⚠️ Tanpa Air (Fallback)
- Perlu restart manual: `make restart-backend`
- Atau restart semua: `make restart`

## Setup Air

### Di Docker (Otomatis)
Air sudah dikonfigurasi di `docker-compose.dev.yml`. Jika Air tersedia, akan digunakan otomatis. Jika tidak, akan fallback ke `go run`.

### Di Local Development
```bash
# Install Air
go install github.com/air-verse/air@latest

# Atau dengan brew (macOS)
brew install air

# Jalankan dengan Air
cd backend
air

# Atau dengan go run (tanpa hot reload)
go run ./cmd/api
```

## Perintah Restart

### Restart Backend Saja
```bash
# Menggunakan Makefile (recommended)
make restart-backend

# Atau langsung dengan docker-compose
docker-compose -f docker-compose.dev.yml restart backend
```

### Restart Frontend Saja
```bash
make restart-frontend
```

### Restart Semua Service
```bash
make restart
```

## Kapan Perlu Restart Manual?

### ✅ Tidak Perlu Restart (Hot Reload Otomatis)
- Perubahan file `.go` (handler, usecase, repository, dll)
- Perubahan konfigurasi di `.air.toml`
- Perubahan di `internal/` folder

### ⚠️ Perlu Restart Manual
- Perubahan environment variables di `docker-compose.dev.yml`
- Perubahan di `go.mod` atau `go.sum` (perlu rebuild)
- Perubahan di `cmd/api/main.go` (route registration)
- Perubahan di database schema (migration)

## Troubleshooting

### Air Tidak Terdeteksi di Docker
```bash
# Install Air di dalam container
docker-compose -f docker-compose.dev.yml exec backend sh -c "go install github.com/air-verse/air@latest"

# Atau rebuild container
make rebuild
```

### Hot Reload Tidak Berfungsi
1. **Cek Air terinstall**: `air -v`
2. **Cek file `.air.toml`**: Pastikan path `cmd` benar
3. **Cek logs**: `make logs-backend`
4. **Restart manual**: `make restart-backend`

### Build Error
- Cek `build-errors.log` di folder `backend/tmp/`
- Pastikan semua dependencies terinstall: `go mod download`
- Cek syntax error di file `.go`

## Best Practices

1. **Gunakan Air untuk Development**: Lebih cepat dan efisien
2. **Restart Manual untuk Production-like Testing**: Test dengan restart manual untuk memastikan aplikasi start dengan benar
3. **Monitor Logs**: Gunakan `make logs-backend` untuk melihat perubahan real-time
4. **Clean Build**: Jika ada masalah, coba `make clean` lalu `make dev`

## Perbandingan

| Method | Hot Reload | Speed | Setup |
|--------|-----------|-------|-------|
| **Air** | ✅ Otomatis | ⚡ Sangat Cepat | Mudah |
| **go run** | ❌ Manual | 🐢 Perlu restart | Sangat Mudah |
| **Docker restart** | ❌ Manual | 🐢 Lambat | Mudah |

## Referensi

- [Air Documentation](https://github.com/air-verse/air)
- [Air Configuration](https://github.com/air-verse/air/blob/master/air_example.toml)
- [Makefile Commands](../Makefile)

