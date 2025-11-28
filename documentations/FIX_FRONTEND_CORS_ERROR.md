# 🔧 Fix Frontend CORS Error

## Masalah
Frontend error saat fetch ke `/api/v1/auth/profile` dari `https://pedeve-dev.aretaamany.com`.

## Kemungkinan Penyebab
Backend belum di-redeploy dengan CORS_ORIGIN yang baru yang include `https://pedeve-dev.aretaamany.com`.

## Solusi: Verifikasi dan Update CORS

### Step 1: Cek CORS_ORIGIN di Container

**SSH ke backend VM dan cek:**

```bash
# Cek CORS_ORIGIN di container
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN
```

**Expected output:**
```
CORS_ORIGIN=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com
```

### Step 2A: Jika CORS_ORIGIN Sudah Benar

**Cek backend logs untuk error detail:**

```bash
# Cek backend logs
sudo docker logs dms-backend-prod --tail 50 | grep -i cors
```

**Cek apakah request sampai ke backend:**

```bash
# Monitor backend logs real-time
sudo docker logs -f dms-backend-prod
```

Lalu coba akses frontend lagi dan lihat error di logs.

### Step 2B: Jika CORS_ORIGIN Belum Benar

**Restart container dengan CORS_ORIGIN yang benar:**

```bash
# Stop container
sudo docker stop dms-backend-prod
sudo docker rm dms-backend-prod

# Get secrets
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
JWT_SECRET=$(gcloud secrets versions access latest --secret=jwt_secret --project=${GCP_PROJECT_ID})
ENCRYPTION_KEY=$(gcloud secrets versions access latest --secret=encryption_key --project=${GCP_PROJECT_ID})

# URL-encode password
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Construct DATABASE_URL
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Get backend image
BACKEND_IMAGE=$(sudo docker images | grep dms-backend | head -1 | awk '{print $1":"$2}')

# Start container dengan CORS_ORIGIN yang benar
sudo docker run -d \
  --name dms-backend-prod \
  --restart unless-stopped \
  --network host \
  -e GCP_PROJECT_ID=${GCP_PROJECT_ID} \
  -e GCP_SECRET_MANAGER_ENABLED=false \
  -e GCP_STORAGE_ENABLED=true \
  -e GCP_STORAGE_BUCKET=pedeve-dev-bucket \
  -e DATABASE_URL="${DATABASE_URL}" \
  -e JWT_SECRET="${JWT_SECRET}" \
  -e ENCRYPTION_KEY="${ENCRYPTION_KEY}" \
  -e PORT=8080 \
  -e ENV=production \
  -e CORS_ORIGIN="https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com" \
  ${BACKEND_IMAGE}

# Verifikasi
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN
```

### Step 3: Re-deploy via CI/CD (Recommended)

**Cara terbaik adalah push ke development branch untuk trigger auto-deploy:**

```bash
# Di local machine, push perubahan
git push origin development
```

CI/CD akan otomatis re-deploy backend dengan CORS_ORIGIN yang benar.

## Troubleshooting

### Cek Browser Console Error

Buka browser console (F12) dan cek error detail:
- CORS error biasanya: `Access to fetch at ... from origin ... has been blocked by CORS policy`
- Error lain bisa jadi authentication atau network issue

### Test CORS dari Command Line

```bash
# Test CORS dari VM
curl -H "Origin: https://pedeve-dev.aretaamany.com" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://api-pedeve-dev.aretaamany.com/api/v1/auth/profile -v
```

**Expected:** Response dengan header `Access-Control-Allow-Origin: https://pedeve-dev.aretaamany.com`

### Cek Backend CORS Configuration

**Cek apakah backend log menunjukkan CORS origin:**

```bash
sudo docker logs dms-backend-prod | grep -i "cors origin"
```

**Expected:** `CORS origin configured origin=https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com`

