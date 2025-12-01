# Analisis Waktu Deployment ke GCP

## Masalah
Deployment ke GCP memakan waktu **~25-39 menit**, yang terlalu lama untuk proses deployment rutin.

## Penyebab Waktu Deployment Lama

### 1. Health Check dengan Retry Mechanism (Paling Signifikan)
**Backend Health Check:**
- Initial sleep: 30 detik
- MAX_RETRIES: 10 kali
- RETRY_DELAY: 10 detik per retry
- **Potensi maksimal**: 30 + (10 × 10) = **130 detik (~2 menit 10 detik)**
- **Rata-rata**: Jika backend ready di attempt ke-3, total = 30 + (2 × 10) = **50 detik**

**Frontend Health Check:**
- Initial sleep: 10 detik
- MAX_RETRIES: 5 kali
- RETRY_DELAY: 5 detik per retry
- **Potensi maksimal**: 10 + (5 × 5) = **35 detik**
- **Rata-rata**: Jika frontend ready di attempt ke-2, total = 10 + (1 × 5) = **15 detik**

**Total Health Check**: ~1-2.5 menit

### 2. SSL Certificate Generation (Jika Certificate Belum Ada)
**Certbot Process:**
- Certificate generation: **1-2 menit** (jika certificate belum ada)
- Certbot renew --dry-run: **30-60 detik**
- Nginx config update oleh Certbot: **10-20 detik**
- **Total**: **~2-3 menit** (hanya saat certificate pertama kali dibuat)

**Catatan**: Setelah certificate ada, script akan skip dan hanya memakan waktu **<5 detik**

### 3. Docker Image Transfer
**Backend Image:**
- docker save: **1-2 menit** (tergantung ukuran image)
- gcloud compute scp: **3-5 menit** (tergantung ukuran file dan network)
- docker load di VM: **1-2 menit**
- **Total**: **~5-9 menit**

**Catatan**: Ini adalah proses yang tidak bisa dihindari, tapi bisa dioptimasi dengan:
- Menggunakan image yang lebih kecil
- Menggunakan GCP Artifact Registry (lebih cepat dari SCP)
- Menggunakan docker pull langsung di VM (jika image sudah di registry)

### 4. Multiple SSH Connections
**Setiap gcloud compute ssh/scp:**
- Connection establishment: **2-5 detik**
- Command execution: **1-10 detik** (tergantung command)
- **Total SSH calls dalam workflow**: ~15-20 calls
- **Total overhead**: **~1-3 menit**

### 5. Script Execution dengan Retry
**ensure-services-running.sh:**
- Container retry: MAX_RETRIES=5, sleep 5 detik = **~25 detik** (jika perlu retry)
- Port check retry: MAX_RETRIES=10, sleep 3 detik = **~30 detik** (jika perlu retry)
- Nginx retry: MAX_RETRIES=5, sleep 3 detik = **~15 detik** (jika perlu retry)
- **Total**: **~1-2 menit** (jika semua perlu retry)

### 6. Backend Container Startup
**Backend Container:**
- Container start: **10-20 detik**
- Application initialization: **10-30 detik** (tergantung database connection, seeding, dll)
- **Total**: **~20-50 detik**

## Total Waktu Breakdown (Estimasi)

| Komponen | Waktu (Best Case) | Waktu (Worst Case) |
|----------|------------------|-------------------|
| Build & Push | 3-5 menit | 5-8 menit |
| Docker Image Transfer | 5 menit | 9 menit |
| SSL Setup (jika baru) | 5 detik | 3 menit |
| SSL Setup (sudah ada) | 5 detik | 5 detik |
| Script Execution | 2 menit | 5 menit |
| Health Check | 1 menit | 2.5 menit |
| **TOTAL** | **~11-13 menit** | **~25-30 menit** |

## Solusi Optimasi

### 1. Optimasi Health Check (Menghemat ~1-2 menit)
**Masalah**: Retry terlalu banyak dan delay terlalu lama

**Solusi**:
- Kurangi initial sleep backend dari 30 detik menjadi 15 detik
- Kurangi MAX_RETRIES backend dari 10 menjadi 5
- Kurangi RETRY_DELAY dari 10 detik menjadi 5 detik
- Gunakan exponential backoff untuk retry

**Expected savings**: ~1-2 menit

### 2. Skip SSL Setup Jika Certificate Sudah Ada (Menghemat ~2-3 menit)
**Masalah**: Script SSL setup masih menjalankan certbot renew --dry-run meskipun certificate sudah ada

**Solusi**:
- Skip certbot renew --dry-run jika certificate masih valid (>30 hari sebelum expiry)
- Hanya jalankan dry-run saat certificate mendekati expiry

**Expected savings**: ~30-60 detik per deployment

### 3. Optimasi Docker Image Transfer (Menghemat ~2-4 menit)
**Masalah**: docker save + scp lambat

**Solusi**:
- Gunakan docker pull langsung di VM dari GitHub Container Registry
- Atau gunakan GCP Artifact Registry (lebih cepat)
- Atau gunakan docker image yang sudah di-cache di VM

**Expected savings**: ~2-4 menit

### 4. Parallel Execution (Menghemat ~2-3 menit)
**Masalah**: Backend dan Frontend deployment sequential

**Solusi**:
- Deploy backend dan frontend secara parallel (jika tidak ada dependency)
- Health check bisa dilakukan parallel juga

**Expected savings**: ~2-3 menit

### 5. Reduce SSH Overhead (Menghemat ~1-2 menit)
**Masalah**: Terlalu banyak SSH calls

**Solusi**:
- Combine multiple commands dalam satu SSH call
- Gunakan script yang lebih comprehensive untuk mengurangi jumlah SSH calls

**Expected savings**: ~1-2 menit

## Implementasi Optimasi

### Priority 1: Quick Wins (Mudah, Impact Besar)
1. ✅ Optimasi Health Check (kurangi retry dan delay)
2. ✅ Skip SSL dry-run jika certificate masih valid
3. ✅ Combine SSH commands

**Expected total savings**: ~3-5 menit

### Priority 2: Medium Effort (Sedang, Impact Sedang)
1. ⚠️ Optimasi Docker Image Transfer (gunakan docker pull langsung)
2. ⚠️ Reduce initial sleep times

**Expected total savings**: ~2-4 menit

### Priority 3: Long Term (Sulit, Impact Besar)
1. 🔄 Parallel deployment (backend + frontend)
2. 🔄 Use GCP Artifact Registry
3. 🔄 Implement deployment caching

**Expected total savings**: ~3-5 menit

## Target Waktu Deployment

**Current**: ~25-39 menit
**After Priority 1**: ~20-30 menit
**After Priority 1+2**: ~15-25 menit
**After All Optimizations**: ~10-15 menit

## Catatan Penting

1. **SSL Certificate Generation**: Hanya terjadi saat certificate pertama kali dibuat. Setelah itu, hanya memakan waktu <5 detik.

2. **Health Check Retry**: Retry mechanism penting untuk handle case dimana backend/frontend butuh waktu untuk startup. Tapi bisa dioptimasi dengan mengurangi retry yang tidak perlu.

3. **Docker Image Transfer**: Ini adalah bottleneck terbesar. Solusi terbaik adalah menggunakan docker pull langsung di VM atau GCP Artifact Registry.

4. **Network Latency**: Latency ke GCP dari GitHub Actions bisa bervariasi. Ini tidak bisa dikontrol, tapi bisa dioptimasi dengan menggunakan GCP Artifact Registry.

## Rekomendasi

**Untuk deployment berikutnya:**
1. Implement Priority 1 optimizations (quick wins)
2. Monitor waktu deployment setelah optimasi
3. Jika masih >20 menit, implement Priority 2
4. Jika masih >15 menit, consider Priority 3

**Expected improvement**: Dari ~25-39 menit menjadi ~15-25 menit (improvement ~40-50%)

