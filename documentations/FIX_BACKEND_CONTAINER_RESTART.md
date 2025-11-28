# 🔧 Fix Backend Container Restart

## Masalah
1. `encryption_key` secret tidak ditemukan
2. `BACKEND_IMAGE` kosong
3. 502 Bad Gateway - container tidak running

## Solusi: Restart Container dengan Benar

**SSH ke backend VM dan jalankan:**

```bash
# Stop dan remove container lama
sudo docker stop dms-backend-prod 2>/dev/null || true
sudo docker rm dms-backend-prod 2>/dev/null || true

# Get secrets
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
JWT_SECRET=$(gcloud secrets versions access latest --secret=jwt_secret --project=${GCP_PROJECT_ID})

# encryption_key optional - gunakan empty string jika tidak ada
ENCRYPTION_KEY=$(gcloud secrets versions access latest --secret=encryption_key --project=${GCP_PROJECT_ID} 2>/dev/null || echo '')

# URL-encode password
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Construct DATABASE_URL
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Get backend image (cara yang lebih reliable)
BACKEND_IMAGE=$(sudo docker images --format "{{.Repository}}:{{.Tag}}" | grep "dms-backend\|ghcr.io.*dms-backend" | head -1)

# Jika masih kosong, coba cara lain
if [ -z "$BACKEND_IMAGE" ]; then
  BACKEND_IMAGE=$(sudo docker images | grep -E "dms-backend|ghcr" | head -1 | awk '{print $1":"$2}')
fi

# Verifikasi image ada
echo "Backend image: $BACKEND_IMAGE"
sudo docker images | grep -E "dms-backend|ghcr"

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

# Verifikasi container running
sudo docker ps | grep dms-backend-prod

# Verifikasi CORS_ORIGIN
sudo docker exec dms-backend-prod env | grep CORS_ORIGIN

# Cek logs
sudo docker logs dms-backend-prod --tail 20
```

## Alternative: Gunakan Image dari GHCR Langsung

**Jika image tidak ditemukan di local, pull dari GHCR:**

```bash
# Login ke GHCR (jika perlu)
echo "$GITHUB_TOKEN" | sudo docker login ghcr.io -u USERNAME --password-stdin

# Pull image
sudo docker pull ghcr.io/repoareta/dms-backend:latest

# Gunakan image ini
BACKEND_IMAGE="ghcr.io/repoareta/dms-backend:latest"
```

## Troubleshooting

### Cek Container Status

```bash
# Cek apakah container running
sudo docker ps -a | grep dms-backend-prod

# Cek logs jika container tidak start
sudo docker logs dms-backend-prod --tail 50
```

### Cek Backend Health

```bash
# Test dari dalam VM
curl http://127.0.0.1:8080/health

# Test dari luar (via domain)
curl https://api-pedeve-dev.aretaamany.com/health
```

### Fix 502 Bad Gateway

502 biasanya berarti:
- Backend container tidak running
- Backend tidak listen di port 8080
- Nginx tidak bisa connect ke backend

**Cek:**
```bash
# Cek container status
sudo docker ps | grep dms-backend-prod

# Cek apakah backend listen di port 8080
sudo ss -tlnp | grep 8080

# Cek Nginx config
sudo cat /etc/nginx/sites-enabled/backend-api | grep proxy_pass
```

