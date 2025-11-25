#!/bin/bash

# Script untuk list semua users di database
# Usage: ./scripts/list-users.sh

set -e

echo "📋 Daftar Semua Users di Database:"
echo ""

docker-compose -f docker-compose.dev.yml exec -T postgres psql -U postgres -d db_dms_pedeve <<EOF
SELECT 
  id,
  username,
  email,
  role,
  is_active,
  role_id,
  company_id,
  created_at
FROM users 
ORDER BY created_at DESC;
EOF

echo ""
echo "💡 Tips:"
echo "  - Gunakan email atau username yang tepat untuk login"
echo "  - Pastikan is_active = true untuk bisa login"
echo ""

