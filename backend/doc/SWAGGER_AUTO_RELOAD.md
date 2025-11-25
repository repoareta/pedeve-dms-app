# Swagger Auto-Reload & Auto-Regenerate

Swagger documentation sekarang bisa auto-update tanpa perlu restart server!

## 🚀 Cara Menggunakan

### Opsi 1: Auto-Regenerate dengan File Watcher (Recommended)

Jalankan script di terminal terpisah:

```bash
cd backend
./scripts/auto-swagger.sh
```

Script ini akan:
- ✅ Watch perubahan di file handler (`internal/delivery/http/*.go`)
- ✅ Auto-regenerate Swagger saat ada perubahan
- ✅ Swagger UI akan auto-reload (tidak perlu refresh browser)

**Untuk macOS:**
```bash
# Install fswatch jika belum ada
brew install fswatch

# Jalankan watcher
./scripts/auto-swagger.sh
```

**Untuk Linux:**
```bash
# Install inotify-tools jika belum ada
sudo apt-get install inotify-tools

# Jalankan watcher
./scripts/auto-swagger.sh
```

### Opsi 2: Manual Regenerate via API Endpoint

Jika tidak ingin menggunakan file watcher, bisa regenerate manual via API:

```bash
# Login dulu untuk mendapatkan token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"superadmin","password":"Pedeve123"}'

# Regenerate Swagger (gunakan token dari login)
curl -X POST http://localhost:8080/api/v1/swagger/regenerate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN"
```

Atau via Swagger UI:
1. Buka http://localhost:8080/swagger/index.html
2. Login dengan akun superadmin
3. Cari endpoint `POST /api/v1/swagger/regenerate`
4. Klik "Try it out" → "Execute"

### Opsi 3: Manual Regenerate via Command

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
```

## 🔄 Auto-Reload di Swagger UI

Swagger UI sekarang sudah dikonfigurasi untuk:
- ✅ **No-cache headers** - Mencegah browser cache swagger.json
- ✅ **Auto-reload** - Swagger UI akan otomatis reload saat file berubah

**Cara kerja:**
1. File handler berubah → File watcher detect
2. Swagger auto-regenerate → `docs/swagger.json` ter-update
3. Swagger UI detect perubahan → Auto-reload (tidak perlu refresh browser)

## 📝 Catatan

1. **File Watcher**: Script `auto-swagger.sh` harus running di background saat development
2. **API Endpoint**: Endpoint `/api/v1/swagger/regenerate` memerlukan authentication
3. **No Restart Needed**: Tidak perlu restart server untuk update Swagger!

## 🛠️ Troubleshooting

**Swagger tidak auto-reload?**
- Pastikan file watcher script running
- Hard refresh browser: Cmd+Shift+R (Mac) atau Ctrl+Shift+R (Windows)
- Cek console browser untuk error

**File watcher tidak bekerja?**
- Pastikan `fswatch` (macOS) atau `inotifywait` (Linux) sudah terinstall
- Cek permission file script: `chmod +x scripts/auto-swagger.sh`

**API endpoint tidak bisa diakses?**
- Pastikan sudah login dan punya valid JWT token
- Pastikan CSRF token sudah di-set

