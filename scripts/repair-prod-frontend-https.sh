#!/bin/bash
# Perbaiki HTTPS frontend production (tanpa rebuild app).
# Domain production: https://dms.pertamina-pedeve.co.id
#
# Usage (di VM atau via gcloud ssh):
#   export DEPLOY_TARGET=prod
#   bash repair-prod-frontend-https.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export DEPLOY_TARGET=prod
export DOMAIN="dms.pertamina-pedeve.co.id"

echo "🔧 Repair HTTPS frontend production: ${DOMAIN}"

chmod +x "${SCRIPT_DIR}/setup-nginx-frontend.sh" "${SCRIPT_DIR}/setup-frontend-ssl.sh" \
  "${SCRIPT_DIR}/verify-frontend-deployment.sh"

bash "${SCRIPT_DIR}/setup-nginx-frontend.sh" "${DOMAIN}"
bash "${SCRIPT_DIR}/setup-frontend-ssl.sh" "${DOMAIN}" || true
bash "${SCRIPT_DIR}/setup-nginx-frontend.sh" "${DOMAIN}"
bash "${SCRIPT_DIR}/verify-frontend-deployment.sh"

echo "✅ Repair selesai — test: curl -I https://${DOMAIN}/login"
