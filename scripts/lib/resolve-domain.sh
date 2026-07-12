#!/bin/bash
# Resolve DOMAIN for deploy scripts.
# Production WAJIB set DOMAIN atau pass arg — tidak boleh fallback ke domain dev.

resolve_deploy_domain() {
  local arg="${1:-}"

  if [ -n "${arg}" ]; then
    printf '%s' "${arg}"
    return 0
  fi

  if [ -n "${DOMAIN:-}" ]; then
    printf '%s' "${DOMAIN}"
    return 0
  fi

  if [ "${DEPLOY_TARGET:-}" = "prod" ]; then
    echo "❌ ERROR: DOMAIN wajib untuk production (contoh: dms.pertamina-pedeve.co.id)" >&2
    return 1
  fi

  printf '%s' "${DEV_DOMAIN_DEFAULT:-localhost}"
}

ssl_cert_exists_for_domain() {
  local domain="$1"
  sudo test -f "/etc/letsencrypt/live/${domain}/fullchain.pem" \
    && sudo test -f "/etc/letsencrypt/live/${domain}/privkey.pem"
}
