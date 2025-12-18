# 🐳 Docker vs Kubernetes - Penjelasan Sederhana

Penjelasan mudah dipahami tentang Docker, Docker Compose, dan Kubernetes untuk kebutuhan deployment multiple aplikasi.

---

## 📚 Konsep Dasar (Dengan Analogi)

### 1. **Docker Container** = Kamar Hotel
- Setiap aplikasi punya "kamar" sendiri yang terisolasi
- Aplikasi di dalam container tidak saling ganggu
- Contoh: Aplikasi Go Anda berjalan di container sendiri

### 2. **Docker Compose** = Gedung Hotel dengan Beberapa Kamar
- Satu file konfigurasi untuk mengatur beberapa container sekaligus
- Semua container bisa saling "berkomunikasi" jika perlu
- Contoh: Backend, Frontend, dan Database dalam satu `docker-compose.yml`

### 3. **Kubernetes (K8s)** = Kompleks Hotel Besar dengan Banyak Gedung
- Mengatur ratusan/ribuan container di banyak server
- Otomatis mengatur load balancing, scaling, failover
- Lebih kompleks, butuh lebih banyak resources

---

## 🎯 Kebutuhan Anda: 3 Aplikasi di Satu Server

**Skenario:**
- Aplikasi Go (yang sekarang)
- Aplikasi Laravel
- Aplikasi WordPress

**Pertanyaan:** Apakah butuh Kubernetes?

**Jawaban Singkat:** **TIDAK PERLU** untuk saat ini! 🎉

---

## ✅ Solusi yang Tepat: Docker Compose dengan Multiple Services

### Kenapa Docker Compose Cukup?

1. **Satu Server Saja**
   - Kubernetes dirancang untuk banyak server (cluster)
   - Untuk 1 server, Docker Compose lebih sederhana dan efisien

2. **Aplikasi Terpisah**
   - Docker Compose bisa mengatur 3 aplikasi terpisah dengan mudah
   - Masing-masing punya container sendiri
   - Tidak saling terkait = tidak perlu network yang kompleks

3. **Lebih Mudah Dikelola**
   - Satu file `docker-compose.yml` untuk semua
   - Command sederhana: `docker-compose up`
   - Tidak perlu setup cluster, nodes, pods, dll

---

## 🏗️ Arsitektur yang Disarankan

### Opsi 1: Satu Docker Compose File (Recommended untuk Awal)

```
server/
├── docker-compose.yml          # Semua aplikasi dalam satu file
│   ├── go-app (backend + frontend)
│   ├── laravel-app
│   └── wordpress-app
```

**Kelebihan:**
- ✅ Sederhana, mudah dikelola
- ✅ Satu command untuk start/stop semua
- ✅ Mudah untuk development

**Kekurangan:**
- ⚠️ Semua aplikasi restart bersamaan saat update file
- ⚠️ Jika satu aplikasi crash, bisa affect yang lain (tapi jarang)

---

### Opsi 2: Multiple Docker Compose Files (Recommended untuk Production)

```
server/
├── go-app/
│   └── docker-compose.yml      # Aplikasi Go
├── laravel-app/
│   └── docker-compose.yml      # Aplikasi Laravel
└── wordpress-app/
    └── docker-compose.yml      # Aplikasi WordPress
```

**Kelebihan:**
- ✅ Setiap aplikasi benar-benar terpisah
- ✅ Update satu aplikasi tidak affect yang lain
- ✅ Bisa di-deploy secara independen
- ✅ Lebih mudah untuk maintenance

**Kekurangan:**
- ⚠️ Perlu manage 3 file terpisah
- ⚠️ Perlu setup network jika aplikasi perlu komunikasi

---

## 📋 Contoh Konfigurasi

### Contoh: Docker Compose untuk 3 Aplikasi Terpisah

```yaml
version: '3.8'

services:
  # Aplikasi Go Anda (yang sekarang)
  go-backend:
    image: ghcr.io/repoareta/dms-backend:latest
    container_name: go-backend
    ports:
      - "8080:8080"
    networks:
      - go-network
    restart: unless-stopped

  go-frontend:
    image: ghcr.io/repoareta/dms-frontend:latest
    container_name: go-frontend
    ports:
      - "80:80"
    networks:
      - go-network
    depends_on:
      - go-backend

  # Aplikasi Laravel
  laravel-app:
    image: laravel-app:latest
    container_name: laravel-app
    ports:
      - "8081:80"  # Port berbeda
    networks:
      - laravel-network
    volumes:
      - ./laravel:/var/www/html
    restart: unless-stopped

  # Laravel Database (jika perlu)
  laravel-db:
    image: mysql:8.0
    container_name: laravel-db
    environment:
      MYSQL_DATABASE: laravel_db
      MYSQL_USER: laravel_user
      MYSQL_PASSWORD: laravel_pass
    networks:
      - laravel-network

  # Aplikasi WordPress
  wordpress:
    image: wordpress:latest
    container_name: wordpress-app
    ports:
      - "8082:80"  # Port berbeda
    networks:
      - wordpress-network
    volumes:
      - ./wordpress:/var/www/html
    restart: unless-stopped

  # WordPress Database
  wordpress-db:
    image: mysql:8.0
    container_name: wordpress-db
    environment:
      MYSQL_DATABASE: wordpress_db
      MYSQL_USER: wp_user
      MYSQL_PASSWORD: wp_pass
    networks:
      - wordpress-network

networks:
  go-network:
    driver: bridge
  laravel-network:
    driver: bridge
  wordpress-network:
    driver: bridge
```

**Penjelasan:**
- Setiap aplikasi punya **network sendiri** (terpisah)
- Setiap aplikasi punya **port berbeda** (tidak bentrok)
- Setiap aplikasi punya **container sendiri** (isolated)

---

## 🔄 Reverse Proxy dengan Nginx

Karena setiap aplikasi pakai port berbeda, kita perlu **Nginx sebagai reverse proxy** untuk:
- `pedeve.aretaamany.com` → Go App (port 80)
- `laravel.aretaamany.com` → Laravel (port 8081)
- `wordpress.aretaamany.com` → WordPress (port 8082)

**Contoh Nginx Config:**

```nginx
# Go App
server {
    listen 80;
    server_name pedeve.aretaamany.com;
    
    location / {
        proxy_pass http://localhost:80;  # Go frontend
    }
    
    location /api {
        proxy_pass http://localhost:8080;  # Go backend
    }
}

# Laravel App
server {
    listen 80;
    server_name laravel.aretaamany.com;
    
    location / {
        proxy_pass http://localhost:8081;
    }
}

# WordPress App
server {
    listen 80;
    server_name wordpress.aretaamany.com;
    
    location / {
        proxy_pass http://localhost:8082;
    }
}
```

---

## 🤔 Kapan Sebaiknya Pakai Kubernetes?

### Gunakan Kubernetes Jika:

1. **Banyak Server (Cluster)**
   - Punya 3+ server yang perlu di-manage
   - Ingin high availability (jika 1 server down, aplikasi tetap jalan)

2. **Auto Scaling**
   - Traffic naik → otomatis tambah container
   - Traffic turun → otomatis kurangi container

3. **Complex Orchestration**
   - Perlu rolling updates tanpa downtime
   - Perlu canary deployments
   - Perlu service mesh, monitoring kompleks

4. **Enterprise Scale**
   - Ratusan/ribuan aplikasi
   - Multiple teams dengan banyak aplikasi

### JANGAN Pakai Kubernetes Jika:

1. ❌ Hanya 1-2 server
2. ❌ Hanya beberapa aplikasi (3-10 aplikasi)
3. ❌ Tim kecil (1-3 developer)
4. ❌ Tidak perlu auto-scaling
5. ❌ Budget terbatas (Kubernetes butuh lebih banyak resources)

---

## 💰 Perbandingan Biaya & Kompleksitas

| Aspek | Docker Compose | Kubernetes |
|-------|----------------|------------|
| **Setup Time** | 1-2 jam | 1-2 hari |
| **Learning Curve** | Mudah | Sulit (butuh training) |
| **Maintenance** | Sederhana | Kompleks |
| **Resource Usage** | Minimal | Lebih banyak (overhead) |
| **Cost** | Rendah | Lebih tinggi |
| **Suitable For** | 1-10 server, <50 apps | 10+ server, 50+ apps |

---

## 🎯 Rekomendasi untuk Kebutuhan Anda

### Untuk Saat Ini: **Docker Compose dengan Multiple Files**

**Struktur:**
```
/opt/apps/
├── pedeve-dms/
│   ├── docker-compose.yml
│   └── ...
├── laravel-app/
│   ├── docker-compose.yml
│   └── ...
└── wordpress-app/
    ├── docker-compose.yml
    └── ...
```

**Keuntungan:**
- ✅ Setiap aplikasi benar-benar terpisah
- ✅ Update satu aplikasi tidak affect yang lain
- ✅ Mudah untuk backup/restore per aplikasi
- ✅ Bisa di-deploy secara independen
- ✅ Tidak perlu belajar Kubernetes

**Cara Deploy:**
```bash
# Deploy Go App
cd /opt/apps/pedeve-dms
docker-compose up -d

# Deploy Laravel App
cd /opt/apps/laravel-app
docker-compose up -d

# Deploy WordPress App
cd /opt/apps/wordpress-app
docker-compose up -d
```

---

## 🚀 Kapan Harus Migrasi ke Kubernetes?

**Pertimbangkan Kubernetes jika:**

1. **Scale Out** - Butuh lebih dari 5-10 server
2. **High Availability** - Butuh aplikasi tetap jalan meski 1 server down
3. **Auto Scaling** - Traffic sangat fluktuatif, butuh auto scale
4. **Complex Deployments** - Butuh blue-green, canary, dll
5. **Team Growth** - Tim berkembang, banyak aplikasi

**Tapi untuk sekarang:** Docker Compose sudah lebih dari cukup! ✅

---

## 📝 Kesimpulan

### Untuk Kebutuhan Anda (3 Aplikasi di 1 Server):

✅ **Gunakan: Docker Compose dengan Multiple Files**
- Sederhana
- Mudah dikelola
- Cukup powerful
- Tidak perlu Kubernetes

❌ **Jangan Pakai: Kubernetes**
- Overkill untuk 1 server
- Terlalu kompleks
- Butuh lebih banyak resources
- Learning curve tinggi

### Analogi Akhir:

- **Docker Compose** = Apartemen dengan 3 kamar terpisah (cukup untuk kebutuhan Anda)
- **Kubernetes** = Kompleks perumahan dengan ratusan rumah (terlalu besar untuk kebutuhan Anda)

---

## 🔗 Next Steps

1. **Setup struktur folder** untuk 3 aplikasi terpisah
2. **Buat docker-compose.yml** untuk masing-masing aplikasi
3. **Setup Nginx reverse proxy** untuk routing domain
4. **Test deployment** satu per satu
5. **Setup monitoring** (optional, untuk production)

---

**Last Updated:** 2025-01-27  
**Status:** 📋 Reference Document
