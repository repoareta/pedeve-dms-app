# 🌱 Cek dan Jalankan Seeder di Container Backend

## Step 1: Cek Apakah Source Code Ada di Container

**SSH ke backend VM dan cek:**

```bash
# Cek apakah container punya source code
sudo docker exec dms-backend-prod ls -la /root/

# Cek apakah ada directory cmd
sudo docker exec dms-backend-prod ls -la /root/cmd/ 2>/dev/null || echo "No cmd directory"

# Cek apakah ada file main.go atau source code
sudo docker exec dms-backend-prod find /root -name "*.go" 2>/dev/null | head -5
```

## Step 2A: Jika Source Code ADA di Container

**Jalankan seeder langsung:**

```bash
# Masuk ke container
sudo docker exec -it dms-backend-prod sh

# Di dalam container, cek struktur
ls -la
cd /root  # atau /app, tergantung WORKDIR di Dockerfile

# Setup DATABASE_URL (container sudah punya env vars, tapi perlu set manual)
export DATABASE_URL="postgres://pedeve_user_db:$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)@127.0.0.1:5432/db_dev_pedeve?sslmode=disable"

# Jalankan seeder
go run ./cmd/seed-companies
```

## Step 2B: Jika Source Code TIDAK ADA (Hanya Binary)

**Kita perlu copy source code ke container atau build binary seeder.**

### Opsi 1: Copy Source Code ke Container (Temporary)

```bash
# Dari VM, copy source code ke container
# (Butuh source code dari local atau clone repo)

# Atau build binary seeder di local, copy ke container
```

### Opsi 2: Build Binary Seeder dan Copy ke Container

**Dari local machine:**

```bash
# Build binary seeder untuk Linux
cd backend
GOOS=linux GOARCH=amd64 go build -o seed-companies ./cmd/seed-companies

# Copy binary ke VM
gcloud compute scp --zone=asia-southeast2-a seed-companies backend-dev:~/seed-companies

# Di VM, copy binary ke container
sudo docker cp ~/seed-companies dms-backend-prod:/root/seed-companies

# Jalankan di container
sudo docker exec dms-backend-prod /root/seed-companies
```

### Opsi 3: Update Dockerfile untuk Include Source Code (Permanent Fix)

**Update `backend/Dockerfile` untuk include source code:**

```dockerfile
# Stage 2: Production
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy source code (untuk seeder dan tools)
COPY --from=builder /app/cmd ./cmd
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/go.mod ./go.mod
COPY --from=builder /app/go.sum ./go.sum

# Install Go untuk bisa run seeder (atau build binary seeder di stage 1)
RUN apk add --no-cache go

EXPOSE 8080

CMD ["./main"]
```

**Lalu re-deploy backend.**

## Step 3: Setup DATABASE_URL di Container

**Container perlu DATABASE_URL yang benar:**

```bash
# Di VM, get password
DB_PASSWORD=$(gcloud secrets versions access latest --secret=db_password --project=pedeve-pertamina-dms)
DB_PASSWORD_ENCODED=$(echo -n "${DB_PASSWORD}" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

# Set di container
sudo docker exec -e DATABASE_URL="postgres://pedeve_user_db:${DB_PASSWORD_ENCODED}@127.0.0.1:5432/db_dev_pedeve?sslmode=disable" dms-backend-prod go run ./cmd/seed-companies
```

## Quick Check Script

**Jalankan ini di VM untuk cek:**

```bash
echo "🔍 Checking container structure..."
echo "1. Root directory:"
sudo docker exec dms-backend-prod ls -la /root/

echo ""
echo "2. Looking for Go files:"
sudo docker exec dms-backend-prod find /root -name "*.go" 2>/dev/null | head -10

echo ""
echo "3. Looking for cmd directory:"
sudo docker exec dms-backend-prod find /root -type d -name "cmd" 2>/dev/null

echo ""
echo "4. Container environment:"
sudo docker exec dms-backend-prod env | grep -E "DATABASE_URL|GCP"
```

