#!/bin/bash
# Deploy production langsung dari terminal (bypass CI queue ~25 menit).
# Versioning tetap via git — build dari working tree / branch saat ini.
#
# Prasyarat:
#   - gcloud CLI terautentikasi dengan akses ke project prod
#   - Docker terpasang lokal
#   - Node.js 20+ untuk build frontend
#
# Usage:
#   export GCP_PROJECT_ID_PROD=pedeve-production   # wajib
#   ./scripts/deploy-prod-fast.sh                  # deploy backend + frontend
#   ./scripts/deploy-prod-fast.sh backend          # backend saja
#   ./scripts/deploy-prod-fast.sh frontend         # frontend saja
#
# Opsi env:
#   GCP_ZONE=asia-southeast2-a
#   BACKEND_VM=backend-prod-1
#   FRONTEND_VM=frontend-prod-2
#   SKIP_BUILD=1          # skip docker/npm build (gunakan image/dist yang ada)
#   BACKEND_IMAGE_TAG=latest

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

TARGET="${1:-all}"
GCP_PROJECT="${GCP_PROJECT_ID_PROD:-}"
GCP_ZONE="${GCP_ZONE:-asia-southeast2-a}"
BACKEND_VM="${BACKEND_VM:-backend-prod-1}"
FRONTEND_VM="${FRONTEND_VM:-frontend-prod-2}"
REPO_OWNER="${GITHUB_REPOSITORY_OWNER:-repoareta}"
BACKEND_IMAGE_TAG="${BACKEND_IMAGE_TAG:-latest}"
BACKEND_IMAGE="ghcr.io/${REPO_OWNER}/pedeve-dms-backend:${BACKEND_IMAGE_TAG}"
FRONTEND_DOMAIN="${FRONTEND_DOMAIN:-dms.pertamina-pedeve.co.id}"
BACKEND_DOMAIN="${BACKEND_DOMAIN:-api-reports.pertamina-pedeve.co.id}"

if [[ -z "${GCP_PROJECT}" ]]; then
  echo "❌ Set GCP_PROJECT_ID_PROD terlebih dahulu, contoh:"
  echo "   export GCP_PROJECT_ID_PROD=pedeve-production"
  exit 1
fi

if [[ ! "${GCP_PROJECT}" =~ ^[a-z0-9-]{1,30}$ ]]; then
  echo "❌ GCP_PROJECT_ID_PROD format tidak valid"
  exit 1
fi

echo "🚀 Fast deploy PROD"
echo "   Project : ${GCP_PROJECT}"
echo "   Zone    : ${GCP_ZONE}"
echo "   Target  : ${TARGET}"
echo "   Branch  : $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
echo "   Commit  : $(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
echo ""

preflight() {
  command -v gcloud >/dev/null || { echo "❌ gcloud tidak ditemukan"; exit 1; }
  command -v docker >/dev/null || { echo "❌ docker tidak ditemukan"; exit 1; }

  echo "🔍 Verifikasi akses GCP..."
  if ! gcloud compute instances describe "${BACKEND_VM}" \
    --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" --format="value(name)" >/dev/null 2>&1; then
    echo "❌ Tidak bisa akses VM ${BACKEND_VM} di project ${GCP_PROJECT}"
    echo "   Jalankan: gcloud auth login && gcloud config set project ${GCP_PROJECT}"
    exit 1
  fi
  echo "✅ Akses GCP OK"
}

deploy_backend() {
  echo ""
  echo "═══════════════════════════════════════"
  echo "  BACKEND"
  echo "═══════════════════════════════════════"

  if [[ "${SKIP_BUILD:-0}" != "1" ]]; then
    echo "🐳 Building backend image..."
    docker build -t "${BACKEND_IMAGE}" -f backend/Dockerfile backend/
  else
    echo "⏭️  SKIP_BUILD=1 — menggunakan image lokal: ${BACKEND_IMAGE}"
    if ! docker image inspect "${BACKEND_IMAGE}" >/dev/null 2>&1; then
      echo "❌ Image ${BACKEND_IMAGE} tidak ditemukan lokal"
      exit 1
    fi
  fi

  echo "📦 Saving image..."
  docker save "${BACKEND_IMAGE}" -o /tmp/backend-image.tar

  echo "📤 Upload ke ${BACKEND_VM}..."
  gcloud compute scp --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    /tmp/backend-image.tar "${BACKEND_VM}:~/backend-image.tar"

  gcloud compute scp --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    scripts/deploy-backend-vm.sh \
    scripts/setup-backend-nginx.sh \
    scripts/setup-backend-ssl.sh \
    scripts/verify-backend-deployment.sh \
    scripts/ensure-services-running.sh \
    "${BACKEND_VM}:~/"

  echo "🚀 Menjalankan deploy di VM..."
  gcloud compute ssh "${BACKEND_VM}" \
    --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    --command="
      set -euo pipefail
      chmod +x ~/deploy-backend-vm.sh ~/setup-backend-nginx.sh ~/setup-backend-ssl.sh \
        ~/verify-backend-deployment.sh ~/ensure-services-running.sh

      export DB_SECRET_SUFFIX='_prod'
      export DB_NAME='db_prod_pedeve'
      export DB_USER='pedeve_user_db_prod'
      export STORAGE_BUCKET='pedeve-prod-bucket'
      export CORS_ORIGIN='https://${FRONTEND_DOMAIN}'
      export DISABLE_RATE_LIMIT='false'
      export DOCKER_MEMORY_LIMIT='1g'
      export DOCKER_CPUS='1.5'
      export DOCKER_PIDS_LIMIT='200'

      ~/deploy-backend-vm.sh '${GCP_PROJECT}' '${BACKEND_IMAGE}'

      export DOMAIN='${BACKEND_DOMAIN}'
      bash ~/setup-backend-ssl.sh \${DOMAIN} || true
      bash ~/setup-backend-nginx.sh \${DOMAIN}
      bash ~/ensure-services-running.sh backend
      bash ~/verify-backend-deployment.sh

      rm -f ~/backend-image.tar
    "

  rm -f /tmp/backend-image.tar
  echo "✅ Backend deploy selesai"
}

deploy_frontend() {
  echo ""
  echo "═══════════════════════════════════════"
  echo "  FRONTEND"
  echo "═══════════════════════════════════════"

  if [[ "${SKIP_BUILD:-0}" != "1" ]]; then
    echo "📦 Building frontend..."
    (
      cd frontend
      export VITE_API_URL="https://${BACKEND_DOMAIN}/api/v1"
      npm ci --prefer-offline 2>/dev/null || npm install
      npm run build-only
    )
    tar -czf /tmp/frontend-dist.tar.gz -C frontend/dist .
  else
    if [[ ! -d frontend/dist ]]; then
      echo "❌ frontend/dist tidak ada. Build dulu atau hapus SKIP_BUILD=1"
      exit 1
    fi
    echo "⏭️  SKIP_BUILD=1 — menggunakan frontend/dist yang ada"
    tar -czf /tmp/frontend-dist.tar.gz -C frontend/dist .
  fi

  echo "📤 Upload ke ${FRONTEND_VM}..."
  gcloud compute scp --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    /tmp/frontend-dist.tar.gz "${FRONTEND_VM}:~/frontend-dist.tar.gz"

  gcloud compute scp --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    scripts/setup-nginx-frontend.sh \
    scripts/setup-frontend-ssl.sh \
    scripts/verify-frontend-deployment.sh \
    scripts/ensure-services-running.sh \
    "${FRONTEND_VM}:~/"

  gcloud compute ssh "${FRONTEND_VM}" \
    --zone="${GCP_ZONE}" --project="${GCP_PROJECT}" \
    --command="
      set -euo pipefail
      if ! command -v nginx &>/dev/null; then
        sudo apt-get update
        sudo apt-get install -y nginx
        sudo systemctl enable nginx
      fi

      sudo rm -rf /var/www/html/*
      sudo mkdir -p /tmp/frontend-dist
      sudo tar -xzf ~/frontend-dist.tar.gz -C /tmp/frontend-dist/
      sudo cp -r /tmp/frontend-dist/* /var/www/html/
      sudo chown -R www-data:www-data /var/www/html
      sudo chmod -R 755 /var/www/html

      chmod +x ~/setup-nginx-frontend.sh ~/setup-frontend-ssl.sh \
        ~/verify-frontend-deployment.sh ~/ensure-services-running.sh
      export DOMAIN='${FRONTEND_DOMAIN}'
      bash ~/setup-frontend-ssl.sh \${DOMAIN} || true
      bash ~/setup-nginx-frontend.sh \${DOMAIN}
      bash ~/ensure-services-running.sh frontend
      bash ~/verify-frontend-deployment.sh

      rm -f ~/frontend-dist.tar.gz
      sudo rm -rf /tmp/frontend-dist
    "

  rm -f /tmp/frontend-dist.tar.gz
  echo "✅ Frontend deploy selesai"
}

health_check() {
  echo ""
  echo "🏥 Health check..."
  sleep 5
  HTTP_CODE=$(curl -s -o /dev/null -w '%{http_code}' -m 15 "https://${BACKEND_DOMAIN}/health" || echo "000")
  if [[ "${HTTP_CODE}" == "200" ]]; then
    echo "✅ Backend health: 200"
  else
    echo "⚠️  Backend health: ${HTTP_CODE} (cek manual)"
  fi
  FE_CODE=$(curl -s -o /dev/null -w '%{http_code}' -m 15 "https://${FRONTEND_DOMAIN}/" || echo "000")
  if [[ "${FE_CODE}" == "200" ]]; then
    echo "✅ Frontend: 200"
  else
    echo "⚠️  Frontend: ${FE_CODE} (cek manual)"
  fi
}

preflight

case "${TARGET}" in
  backend)  deploy_backend; health_check ;;
  frontend) deploy_frontend ;;
  all)
    deploy_backend
    deploy_frontend
    health_check
    ;;
  *)
    echo "❌ Target tidak dikenal: ${TARGET} (gunakan: backend, frontend, all)"
    exit 1
    ;;
esac

echo ""
echo "🎉 Fast deploy selesai (~5-10 menit vs ~25 menit CI/CD)"
