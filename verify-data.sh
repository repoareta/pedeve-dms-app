#!/bin/bash

# ============================================
# Script untuk Verifikasi Data Financial Report
# Company ID: bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d
# ============================================

API_BASE_URL="http://localhost:8080/api/v1"

# ============================================
# KONFIGURASI LOGIN
# ============================================
USERNAME="administrator@pertamina.com"
PASSWORD="Pedeve123"

echo "🔐 Login untuk mendapatkan token..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

# Extract token dari response
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Gagal login. Response:"
  echo "$LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Login berhasil!"
echo ""

COMPANY_ID="bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d"

echo "📊 Memeriksa data financial reports untuk company: $COMPANY_ID"
echo ""

# Get all financial reports for this company
REPORTS_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/financial-reports/company/${COMPANY_ID}" \
  -H "Authorization: Bearer ${TOKEN}")

echo "📋 Daftar Financial Reports yang tersimpan:"
echo "$REPORTS_RESPONSE" | grep -o '"period":"[^"]*' | cut -d'"' -f4 | while read period; do
  echo "   - Period: $period"
done

echo ""
echo "🔍 Detail per periode:"
echo ""

# Check for 2024 data
echo "=== TAHUN 2024 ==="
echo "RKAP 2024:"
RKAP_2024=$(curl -s -X GET "${API_BASE_URL}/financial-reports/compare?company_id=${COMPANY_ID}&year=2024&month=12" \
  -H "Authorization: Bearer ${TOKEN}")
if echo "$RKAP_2024" | grep -q '"rkap"'; then
  echo "   ✅ RKAP 2024 ditemukan"
else
  echo "   ❌ RKAP 2024 tidak ditemukan"
fi

echo "Realisasi YTD Desember 2024:"
if echo "$RKAP_2024" | grep -q '"realisasi_ytd"'; then
  echo "   ✅ Realisasi YTD Desember 2024 ditemukan"
else
  echo "   ❌ Realisasi YTD Desember 2024 tidak ditemukan"
fi

echo ""

# Check for 2025 data
echo "=== TAHUN 2025 ==="
echo "RKAP 2025:"
RKAP_2025=$(curl -s -X GET "${API_BASE_URL}/financial-reports/compare?company_id=${COMPANY_ID}&year=2025&month=12" \
  -H "Authorization: Bearer ${TOKEN}")
if echo "$RKAP_2025" | grep -q '"rkap"'; then
  echo "   ✅ RKAP 2025 ditemukan"
else
  echo "   ❌ RKAP 2025 tidak ditemukan"
fi

echo "Realisasi YTD Desember 2025:"
if echo "$RKAP_2025" | grep -q '"realisasi_ytd"'; then
  echo "   ✅ Realisasi YTD Desember 2025 ditemukan"
else
  echo "   ❌ Realisasi YTD Desember 2025 tidak ditemukan"
fi

echo ""
echo "✅ Verifikasi selesai!"
