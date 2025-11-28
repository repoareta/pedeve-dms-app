# 🔧 Fix 429 Too Many Requests - Rate Limiting

## Masalah
Error `429 Too Many Requests` saat akses `/api/v1/auth/profile`.

## Penyebab
Rate limiting aktif di production dan terlalu ketat untuk development environment.

## Solusi: Disable Rate Limiting untuk Development

### Opsi 1: Restart Container dengan DISABLE_RATE_LIMIT (Quick Fix)

**SSH ke backend VM dan jalankan:**

```bash
# Stop container
sudo docker stop dms-backend-prod
sudo docker rm dms-backend-prod

# Get secrets
export GCP_PROJECT_ID=pedeve-pertamina-dms
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=${GCP_PROJECT_ID})
JWT_SECRET=$(gcloud secrets versions access latest --secret=jwt_secret --project=${GCP_PROJECT_ID})
ENCRYPTION_KEY=$(gcloud secrets versions access latest --secret=encryption_key --project=${GCP_PROJECT_ID} 2>/dev/null || echo '')

# URL-encode password
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Construct DATABASE_URL
DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Get backend image
BACKEND_IMAGE=$(sudo docker images --format "{{.Repository}}:{{.Tag}}" | grep -E "dms-backend|ghcr.*dms-backend" | head -1)

# Start container dengan DISABLE_RATE_LIMIT=true
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
  -e DISABLE_RATE_LIMIT=true \
  -e CORS_ORIGIN="https://pedeve-dev.aretaamany.com,http://34.128.123.1,http://pedeve-dev.aretaamany.com" \
  ${BACKEND_IMAGE}

# Verifikasi
sudo docker exec dms-backend-prod env | grep DISABLE_RATE_LIMIT
sudo docker logs dms-backend-prod --tail 20 | grep -i "rate limit"
```

### Opsi 2: Re-deploy via CI/CD (Permanent Fix)

**Push ke development branch:**

```bash
git push origin development
```

CI/CD akan otomatis deploy dengan `DISABLE_RATE_LIMIT=true`.

## Verifikasi

Setelah restart, test lagi:

```bash
# Test dari browser
# Buka https://pedeve-dev.aretaamany.com
# Login dan cek apakah error 429 hilang
```

## Catatan

- `DISABLE_RATE_LIMIT=true` akan disable rate limiting sepenuhnya
- Untuk production, sebaiknya set rate limit yang lebih tinggi, bukan disable
- Atau gunakan `ENV=development` untuk auto-bypass rate limiting

