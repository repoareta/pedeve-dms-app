#!/bin/bash

# Script untuk cek dan test password user
# Usage: ./scripts/check-user-password.sh <email_or_username> <password>

set -e

if [ $# -lt 2 ]; then
  echo "Usage: $0 <email_or_username> <password>"
  echo "Example: $0 dhani@pertamina.com Pedeve123"
  exit 1
fi

EMAIL_OR_USERNAME=$1
PASSWORD=$2

echo "🔍 Checking user: $EMAIL_OR_USERNAME"
echo ""

# Connect to database and check user
docker-compose -f docker-compose.dev.yml exec -T postgres psql -U postgres -d db_dms_pedeve <<EOF
SELECT 
  id,
  username,
  email,
  role,
  is_active,
  role_id,
  company_id,
  LENGTH(password) as password_length,
  SUBSTRING(password, 1, 10) as password_preview,
  created_at,
  updated_at
FROM users 
WHERE username = '$EMAIL_OR_USERNAME' OR email = '$EMAIL_OR_USERNAME';
EOF

echo ""
echo "💡 Tips:"
echo "  - Password harus di-hash dengan bcrypt (panjang ~60 karakter)"
echo "  - Jika password_preview tidak dimulai dengan \$2a\$ atau \$2b\$, password tidak di-hash dengan benar"
echo "  - Gunakan fitur 'Reset Password' dari superadmin untuk memperbaiki password"
echo ""

