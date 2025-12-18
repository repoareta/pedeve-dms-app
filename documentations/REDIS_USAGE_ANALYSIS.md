# 📊 Analisis Penggunaan Redis di Aplikasi

Analisis tentang apakah Redis digunakan untuk session, cache permission/token, dan rate limiting.

---

## ❌ Kesimpulan: Redis BELUM Digunakan

**Status:** Redis **TIDAK dipakai** sama sekali di aplikasi saat ini.

**Bukti:**
- ❌ Tidak ada dependency Redis di `go.mod`
- ❌ Tidak ada import Redis client library
- ❌ Tidak ada konfigurasi Redis
- ✅ Hanya ada komentar "bisa diganti dengan Redis untuk production" di beberapa tempat

---

## 🔍 Detail Implementasi Saat Ini

### 1. Session Management

**Status:** ❌ Tidak menggunakan Redis

**Implementasi Saat Ini:**
- ✅ **JWT Token di Cookie (Stateless)**
- ✅ Token disimpan di **httpOnly cookie** (`auth_token`)
- ✅ Token berlaku **24 jam** (cookieMaxAge = 24 * 60 * 60)
- ✅ Tidak ada session storage (stateless authentication)

**File:** `backend/internal/infrastructure/cookie/cookie.go`

```go
const (
    authTokenCookie = "auth_token"
    cookieMaxAge    = 24 * 60 * 60 // 24 jam
)
```

**Keuntungan:**
- ✅ Stateless (tidak perlu storage)
- ✅ Scalable (tidak perlu shared session store)
- ✅ Simple (tidak perlu Redis)

**Kekurangan:**
- ⚠️ Tidak bisa revoke token sebelum expire (kecuali blacklist)
- ⚠️ Token size terbatas (JWT bisa besar jika banyak claims)

---

### 2. Cache Permission / Token

**Status:** ❌ Tidak menggunakan Redis

**Implementasi Saat Ini:**
- ❌ **Tidak ada cache** untuk permission
- ❌ Permission check dilakukan **langsung dari database** setiap request
- ❌ JWT token **tidak di-cache**, hanya divalidasi setiap request

**File:** `backend/internal/middleware/rbac.go`

```go
// RequirePermission middleware checks if user has required permission
func RequirePermission(permission Permission) fiber.Handler {
    // ...
    // Get user from database to check role
    var userModel domain.UserModel
    result := database.GetDB().First(&userModel, "id = ?", userID)
    // ...
    // Check permission
    if !HasPermission(userModel.Role, permission) {
        // ...
    }
}
```

**Masalah:**
- ⚠️ **Query database setiap request** untuk check permission
- ⚠️ Tidak ada caching, bisa lambat jika banyak request
- ⚠️ Load database lebih tinggi

**Solusi yang Bisa Diterapkan:**
- ✅ Cache permission di Redis dengan TTL (misalnya 5-10 menit)
- ✅ Cache user role di memory dengan invalidation
- ✅ Cache di JWT token claims (tapi tidak bisa update real-time)

---

### 3. Rate Limiting

**Status:** ❌ Tidak menggunakan Redis

**Implementasi Saat Ini:**
- ✅ **In-Memory Rate Limiter** (map[string]*visitor)
- ✅ Menggunakan `sync.RWMutex` untuk thread-safety
- ✅ Cleanup otomatis setiap 3 menit untuk visitor yang tidak aktif

**File:** `backend/internal/middleware/rate_limit.go`

```go
type RateLimiter struct {
    visitors map[string]*visitor  // In-memory map
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}
```

**Masalah:**
- ⚠️ **Tidak shared** antar instance (jika multiple backend instances)
- ⚠️ Data hilang saat restart
- ⚠️ Tidak scalable untuk multiple servers

**Solusi yang Bisa Diterapkan:**
- ✅ Redis untuk shared rate limiting (semua instance share counter)
- ✅ Distributed rate limiting dengan Redis
- ✅ Lebih akurat untuk multiple backend instances

---

### 4. CSRF Token Storage

**Status:** ❌ Tidak menggunakan Redis

**Implementasi Saat Ini:**
- ✅ **In-Memory Map** untuk CSRF tokens
- ✅ Menggunakan `sync.RWMutex` untuk thread-safety
- ✅ Cleanup otomatis setiap 1 jam

**File:** `backend/internal/middleware/csrf.go`

```go
// CSRF token store (in-memory, bisa diganti dengan Redis untuk production)
var csrfTokens = make(map[string]time.Time)
var csrfMutex sync.RWMutex
```

**Catatan di Code:**
- Ada komentar: `// CSRF token store (in-memory, bisa diganti dengan Redis untuk production)`

**Masalah:**
- ⚠️ Tidak shared antar instance
- ⚠️ Data hilang saat restart

---

## 📋 Ringkasan Pengganti yang Digunakan

| Feature | Redis? | Pengganti Saat Ini | Masalah |
|---------|--------|-------------------|---------|
| **Session** | ❌ Tidak | JWT Token di Cookie (stateless) | Tidak bisa revoke sebelum expire |
| **Cache Permission** | ❌ Tidak | Query database langsung | Lambat, load database tinggi |
| **Cache Token** | ❌ Tidak | Tidak ada cache | Validasi JWT setiap request |
| **Rate Limiting** | ❌ Tidak | In-memory map | Tidak shared antar instance |
| **CSRF Token** | ❌ Tidak | In-memory map | Tidak shared antar instance |

---

## 🎯 Rekomendasi untuk Production

### Prioritas Tinggi (Jika Multiple Instances)

1. **Rate Limiting dengan Redis** ⭐⭐⭐
   - **Kenapa:** Jika ada multiple backend instances, rate limiting harus shared
   - **Impact:** Tanpa Redis, rate limit tidak akurat (setiap instance punya counter sendiri)

2. **CSRF Token dengan Redis** ⭐⭐
   - **Kenapa:** CSRF token harus shared antar instance
   - **Impact:** User bisa dapat error jika request ke instance berbeda

### Prioritas Sedang (Performance)

3. **Cache Permission dengan Redis** ⭐⭐
   - **Kenapa:** Mengurangi query database
   - **Impact:** Performance lebih baik, load database lebih rendah
   - **TTL:** 5-10 menit (balance antara performance dan real-time updates)

### Prioritas Rendah (Optional)

4. **Session dengan Redis** ⭐
   - **Kenapa:** JWT stateless sudah cukup untuk kebanyakan kasus
   - **Impact:** Hanya perlu jika butuh revoke token real-time atau session management kompleks

---

## 💡 Implementasi Redis (Jika Diperlukan)

### Setup Redis di GCP

**Opsi 1: Memorystore (Managed Redis)**
```bash
# Create Memorystore Redis instance
gcloud redis instances create redis-prod \
  --size=1 \
  --region=asia-southeast2 \
  --tier=basic \
  --project=pedeve-production
```

**Opsi 2: Redis di VM (Self-managed)**
```bash
# Install Redis di backend VM
sudo apt-get update
sudo apt-get install -y redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

### Library yang Diperlukan

```go
// Tambahkan ke go.mod
require (
    github.com/redis/go-redis/v9 v9.x.x
)
```

### Contoh Implementasi

**Rate Limiting dengan Redis:**
```go
import "github.com/redis/go-redis/v9"

func (rl *RateLimiter) GetVisitor(ip string) *rate.Limiter {
    // Check Redis untuk rate limit counter
    key := fmt.Sprintf("ratelimit:%s", ip)
    count, err := redisClient.Incr(ctx, key).Result()
    if err == nil {
        redisClient.Expire(ctx, key, time.Minute)
        if count > rl.rate {
            return nil // Rate limit exceeded
        }
    }
    // ...
}
```

---

## 📝 Kesimpulan

**Status Saat Ini:**
- ✅ Redis **BELUM digunakan** sama sekali
- ✅ Semua menggunakan **in-memory** atau **stateless** (JWT)

**Untuk Development:**
- ✅ **Cukup** - tidak perlu Redis untuk single instance

**Untuk Production (Multiple Instances):**
- ⚠️ **Rate Limiting** perlu Redis (shared counter)
- ⚠️ **CSRF Token** perlu Redis (shared storage)
- ⭐ **Cache Permission** recommended (performance)
- ⭐ **Session** optional (JWT sudah cukup)

**Rekomendasi:**
- Jika hanya **1 backend instance**: Tidak perlu Redis
- Jika **multiple backend instances**: Setup Redis untuk rate limiting & CSRF
- Jika butuh **performance**: Tambahkan cache permission

---

**Last Updated:** 2025-01-27  
**Status:** ✅ Analysis Complete
