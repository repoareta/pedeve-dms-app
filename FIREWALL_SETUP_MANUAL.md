# Manual Firewall Setup (Jika Deployment Script Gagal)

## Masalah
VM service account tidak punya permission untuk create firewall rules atau modify VM tags.

## Solusi: Setup dari GCP Console atau Local Machine

### Opsi 1: Via GCP Console (Paling Mudah)

1. **Buka GCP Console:**
   - Go to: https://console.cloud.google.com/networking/firewalls
   - Project: `pedeve-pertamina-dms`

2. **Create Firewall Rule:**
   - Klik **"CREATE FIREWALL RULE"**
   - **Name:** `allow-backend-api`
   - **Direction:** `Ingress`
   - **Targets:** `Specified target tags`
   - **Target tags:** `backend-api-server`
   - **Source IP ranges:** `0.0.0.0/0`
   - **Protocols and ports:** 
     - ✅ `tcp`
     - Port: `8080`
   - Klik **"CREATE"**

3. **Apply Tag ke Backend VM:**
   - Go to: https://console.cloud.google.com/compute/instances
   - Klik VM: `backend-dev`
   - Klik **"EDIT"**
   - Scroll ke **"Network tags"**
   - Tambah tag: `backend-api-server`
   - Klik **"SAVE"**

### Opsi 2: Via Local Machine (gcloud CLI)

**Pastikan Anda sudah login dengan akun yang punya permission:**

```bash
# Login ke GCP
gcloud auth login

# Set project
gcloud config set project pedeve-pertamina-dms

# Create firewall rule
gcloud compute firewall-rules create allow-backend-api \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags backend-api-server \
  --description "Allow Backend API traffic on port 8080" \
  --project pedeve-pertamina-dms

# Apply tag to backend VM
gcloud compute instances add-tags backend-dev \
  --tags backend-api-server \
  --zone asia-southeast2-a \
  --project pedeve-pertamina-dms
```

### Opsi 3: Via GitHub Actions (Otomatis)

**Deployment workflow sekarang sudah include firewall setup.**

Setelah push commit berikutnya, GitHub Actions akan otomatis:
1. Create firewall rule `allow-backend-api`
2. Apply tag `backend-api-server` ke backend VM

**Tidak perlu manual setup jika menggunakan deployment workflow.**

## Verifikasi

Setelah setup, test:

```bash
# Test dari external
curl http://34.101.49.147:8080/health
curl http://34.101.49.147:8080/api/v1/csrf-token

# Test via domain
curl http://api-pedeve-dev.aretaamany.com/api/v1/csrf-token
```

## Troubleshooting

**Jika masih tidak bisa akses:**

1. **Cek firewall rule:**
   ```bash
   gcloud compute firewall-rules describe allow-backend-api --project pedeve-pertamina-dms
   ```

2. **Cek VM tags:**
   ```bash
   gcloud compute instances describe backend-dev \
     --zone asia-southeast2-a \
     --project pedeve-pertamina-dms \
     --format="get(tags.items)"
   ```

3. **Cek apakah port listening:**
   ```bash
   # SSH ke backend VM
   gcloud compute ssh backend-dev --zone=asia-southeast2-a
   
   # Cek port
   sudo ss -tlnp | grep 8080
   ```

4. **Cek backend logs:**
   ```bash
   sudo docker logs --tail 50 dms-backend-prod
   ```

