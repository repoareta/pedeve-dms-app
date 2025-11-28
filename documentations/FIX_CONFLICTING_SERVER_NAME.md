# 🔧 Fix: Conflicting Server Name Warning

## Masalah

Nginx warning:
```
nginx: [warn] conflicting server name "api-pedeve-dev.aretaamany.com" on 0.0.0.0:80, ignored
nginx: [warn] conflicting server name "api-pedeve-dev.aretaamany.com" on [::]:80, ignored
```

## Penyebab

Ada duplicate `server_name` di beberapa config file. Nginx mengabaikan duplicate, tapi warning tetap muncul.

## Solusi

### 1. Cek Semua Nginx Config Files

```bash
# Cek semua enabled sites
sudo ls -la /etc/nginx/sites-enabled/

# Cek semua available sites
sudo ls -la /etc/nginx/sites-available/

# Cek apakah ada duplicate server_name
sudo grep -r "server_name api-pedeve-dev.aretaamany.com" /etc/nginx/sites-enabled/
```

### 2. Hapus Duplicate Config

**Jika ada file lain selain backend-api:**
```bash
# List semua enabled sites
sudo ls -la /etc/nginx/sites-enabled/

# Hapus yang tidak perlu (selain backend-api)
sudo rm -f /etc/nginx/sites-enabled/default
sudo rm -f /etc/nginx/sites-enabled/backend-api.backup  # jika ada
```

### 3. Pastikan Hanya backend-api yang Enabled

```bash
# Pastikan symlink benar
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# Verify
sudo ls -la /etc/nginx/sites-enabled/
```

**Expected output:**
```
total 0
lrwxrwxrwx 1 root root 45 Nov 28 06:09 backend-api -> /etc/nginx/sites-available/backend-api
```

### 4. Test dan Reload

```bash
sudo nginx -t
sudo systemctl reload nginx
```

**Jika masih ada warning, cek apakah ada config di `/etc/nginx/conf.d/` atau `/etc/nginx/nginx.conf`:**
```bash
# Cek conf.d
sudo ls -la /etc/nginx/conf.d/

# Cek nginx.conf untuk include yang mungkin duplicate
sudo grep -n "include" /etc/nginx/nginx.conf
```

### 5. Jika Masih Ada Warning

**Cek apakah ada default config yang masih aktif:**
```bash
# Cek semua file yang mungkin punya server_name
sudo grep -r "api-pedeve-dev.aretaamany.com" /etc/nginx/
```

**Hapus atau comment out duplicate:**
```bash
# Edit file yang punya duplicate
sudo nano /etc/nginx/sites-available/[file-name]

# Comment out atau hapus duplicate server block
```

## Catatan

**Warning ini tidak critical** - Nginx tetap berfungsi normal, hanya mengabaikan duplicate config. Tapi lebih baik di-fix untuk clean logs.

## Quick Fix

```bash
# 1. Hapus semua enabled sites kecuali backend-api
sudo rm -f /etc/nginx/sites-enabled/default
sudo rm -f /etc/nginx/sites-enabled/*.backup

# 2. Pastikan hanya backend-api enabled
sudo ln -sf /etc/nginx/sites-available/backend-api /etc/nginx/sites-enabled/backend-api

# 3. Test dan reload
sudo nginx -t && sudo systemctl reload nginx

# 4. Verify no warnings
sudo nginx -t 2>&1 | grep -i warn || echo "✅ No warnings!"
```

