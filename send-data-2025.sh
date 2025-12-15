#!/bin/bash

# ============================================
# Script untuk Mengirim Data Financial Report Tahun 2025
# Company ID: 7ebf36d1-9541-4de5-8f80-95768fa00b8e
# ============================================

API_BASE_URL="http://localhost:8080/api/v1"

# ============================================
# KONFIGURASI LOGIN
# ============================================
USERNAME="administrator@pertamina.com"
PASSWORD="Pedeve123!@#"

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
echo "🔑 Token: ${TOKEN:0:50}..."
echo ""

# ============================================
# Mendapatkan CSRF Token
# ============================================
echo "🔒 Mendapatkan CSRF token..."
CSRF_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/csrf-token" \
  -H "Authorization: Bearer ${TOKEN}" \
  -c /tmp/cookies.txt)

# Extract CSRF token dari response
CSRF_TOKEN=$(echo $CSRF_RESPONSE | grep -o '"csrf_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$CSRF_TOKEN" ]; then
  echo "❌ Gagal mendapatkan CSRF token. Response:"
  echo "$CSRF_RESPONSE"
  exit 1
fi

echo "✅ CSRF token berhasil didapat!"
echo "🔐 CSRF Token: ${CSRF_TOKEN:0:30}..."
echo ""

# ============================================
# 1. Input RKAP Tahunan - Tahun 2025
# ============================================
echo "📊 Mengirim data RKAP 2025..."
RESPONSE_RKAP=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025",
    "is_rkap": true,
    "current_assets": 300000000000,
    "non_current_assets": 500000000000,
    "short_term_liabilities": 150000000000,
    "long_term_liabilities": 200000000000,
    "equity": 450000000000,
    "revenue": 1000000000000,
    "operating_expenses": 650000000000,
    "operating_profit": 350000000000,
    "other_income": 50000000000,
    "tax": 100000000000,
    "net_profit": 300000000000,
    "operating_cashflow": 400000000000,
    "investing_cashflow": -150000000000,
    "financing_cashflow": -100000000000,
    "ending_balance": 150000000000,
    "roe": 66.67,
    "roi": 30.00,
    "current_ratio": 2.00,
    "cash_ratio": 1.00,
    "ebitda": 400000000000,
    "ebitda_margin": 40.00,
    "net_profit_margin": 30.00,
    "operating_profit_margin": 35.00,
    "debt_to_equity": 0.78,
    "remark": "RKAP tahun 2025. Target anggaran tahunan yang telah disetujui. Target revenue 1T dengan rata-rata bulanan 83.33M, namun diharapkan ada fluktuasi sesuai kondisi operasional dan musiman."
  }')

if echo "$RESPONSE_RKAP" | grep -q '"id"'; then
  echo "✅ RKAP 2025 berhasil disimpan!"
else
  echo "❌ Error RKAP 2025: $RESPONSE_RKAP"
fi

echo ""

# ============================================
# 2. Input Realisasi Bulanan - Januari sampai Desember 2025
# ============================================

# Januari 2025
echo "📊 Mengirim data Realisasi Januari 2025..."
RESPONSE_JAN=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-01",
    "is_rkap": false,
    "current_assets": 215000000000,
    "non_current_assets": 438000000000,
    "short_term_liabilities": 108000000000,
    "long_term_liabilities": 192000000000,
    "equity": 353000000000,
    "revenue": 68000000000,
    "operating_expenses": 49000000000,
    "operating_profit": 19000000000,
    "other_income": 2500000000,
    "tax": 5375000000,
    "net_profit": 16125000000,
    "operating_cashflow": 26500000000,
    "investing_cashflow": -18500000000,
    "financing_cashflow": -8000000000,
    "ending_balance": 98000000000,
    "roe": 4.57,
    "roi": 2.69,
    "current_ratio": 1.99,
    "cash_ratio": 1.08,
    "ebitda": 28000000000,
    "ebitda_margin": 41.18,
    "net_profit_margin": 23.71,
    "operating_profit_margin": 27.94,
    "debt_to_equity": 0.85,
    "remark": "Laporan realisasi bulan Januari 2025. Awal tahun dengan performa yang cukup baik, revenue sedikit di bawah target bulanan karena periode transisi."
  }')

if echo "$RESPONSE_JAN" | grep -q '"id"'; then
  echo "✅ Januari 2025 berhasil disimpan!"
else
  echo "❌ Error Januari 2025: $RESPONSE_JAN"
fi

echo ""

# Februari 2025
echo "📊 Mengirim data Realisasi Februari 2025..."
RESPONSE_FEB=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-02",
    "is_rkap": false,
    "current_assets": 222000000000,
    "non_current_assets": 440000000000,
    "short_term_liabilities": 110000000000,
    "long_term_liabilities": 190000000000,
    "equity": 362000000000,
    "revenue": 75000000000,
    "operating_expenses": 48500000000,
    "operating_profit": 26500000000,
    "other_income": 3200000000,
    "tax": 7425000000,
    "net_profit": 22275000000,
    "operating_cashflow": 30000000000,
    "investing_cashflow": -17500000000,
    "financing_cashflow": -7500000000,
    "ending_balance": 103000000000,
    "roe": 6.15,
    "roi": 3.71,
    "current_ratio": 2.02,
    "cash_ratio": 1.11,
    "ebitda": 33000000000,
    "ebitda_margin": 44.00,
    "net_profit_margin": 29.70,
    "operating_profit_margin": 35.33,
    "debt_to_equity": 0.83,
    "remark": "Laporan realisasi bulan Februari 2025. Peningkatan revenue 10.29% dari bulan sebelumnya. Performa operasional membaik dengan efisiensi biaya yang lebih baik."
  }')

if echo "$RESPONSE_FEB" | grep -q '"id"'; then
  echo "✅ Februari 2025 berhasil disimpan!"
else
  echo "❌ Error Februari 2025: $RESPONSE_FEB"
fi

echo ""

# Maret 2025
echo "📊 Mengirim data Realisasi Maret 2025..."
RESPONSE_MAR=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-03",
    "is_rkap": false,
    "current_assets": 228000000000,
    "non_current_assets": 442000000000,
    "short_term_liabilities": 113000000000,
    "long_term_liabilities": 187000000000,
    "equity": 370000000000,
    "revenue": 82000000000,
    "operating_expenses": 52000000000,
    "operating_profit": 30000000000,
    "other_income": 3800000000,
    "tax": 8450000000,
    "net_profit": 25350000000,
    "operating_cashflow": 33500000000,
    "investing_cashflow": -16500000000,
    "financing_cashflow": -8200000000,
    "ending_balance": 108000000000,
    "roe": 6.85,
    "roi": 4.23,
    "current_ratio": 2.02,
    "cash_ratio": 1.14,
    "ebitda": 36000000000,
    "ebitda_margin": 43.90,
    "net_profit_margin": 30.91,
    "operating_profit_margin": 36.59,
    "debt_to_equity": 0.81,
    "remark": "Laporan realisasi bulan Maret 2025. Peningkatan revenue 9.33% dari bulan sebelumnya. Performa sangat baik dengan margin yang meningkat."
  }')

if echo "$RESPONSE_MAR" | grep -q '"id"'; then
  echo "✅ Maret 2025 berhasil disimpan!"
else
  echo "❌ Error Maret 2025: $RESPONSE_MAR"
fi

echo ""

# April 2025
echo "📊 Mengirim data Realisasi April 2025..."
RESPONSE_APR=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-04",
    "is_rkap": false,
    "current_assets": 225000000000,
    "non_current_assets": 440000000000,
    "short_term_liabilities": 112000000000,
    "long_term_liabilities": 188000000000,
    "equity": 365000000000,
    "revenue": 72000000000,
    "operating_expenses": 50500000000,
    "operating_profit": 21500000000,
    "other_income": 3000000000,
    "tax": 6125000000,
    "net_profit": 18375000000,
    "operating_cashflow": 29000000000,
    "investing_cashflow": -17000000000,
    "financing_cashflow": -8500000000,
    "ending_balance": 105000000000,
    "roe": 5.03,
    "roi": 3.06,
    "current_ratio": 2.01,
    "cash_ratio": 1.12,
    "ebitda": 31000000000,
    "ebitda_margin": 43.06,
    "net_profit_margin": 25.52,
    "operating_profit_margin": 29.86,
    "debt_to_equity": 0.82,
    "remark": "Laporan realisasi bulan April 2025. Penurunan revenue 12.20% dari bulan sebelumnya karena periode libur dan penyesuaian operasional. Margin sedikit menurun."
  }')

if echo "$RESPONSE_APR" | grep -q '"id"'; then
  echo "✅ April 2025 berhasil disimpan!"
else
  echo "❌ Error April 2025: $RESPONSE_APR"
fi

echo ""

# Mei 2025
echo "📊 Mengirim data Realisasi Mei 2025..."
RESPONSE_MEI=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-05",
    "is_rkap": false,
    "current_assets": 232000000000,
    "non_current_assets": 444000000000,
    "short_term_liabilities": 114000000000,
    "long_term_liabilities": 186000000000,
    "equity": 376000000000,
    "revenue": 85000000000,
    "operating_expenses": 53500000000,
    "operating_profit": 31500000000,
    "other_income": 4200000000,
    "tax": 8925000000,
    "net_profit": 26775000000,
    "operating_cashflow": 35000000000,
    "investing_cashflow": -16000000000,
    "financing_cashflow": -8800000000,
    "ending_balance": 112000000000,
    "roe": 7.12,
    "roi": 4.46,
    "current_ratio": 2.04,
    "cash_ratio": 1.16,
    "ebitda": 38000000000,
    "ebitda_margin": 44.71,
    "net_profit_margin": 31.50,
    "operating_profit_margin": 37.06,
    "debt_to_equity": 0.80,
    "remark": "Laporan realisasi bulan Mei 2025. Peningkatan revenue 18.06% dari bulan sebelumnya. Recovery yang kuat setelah periode libur dengan margin yang membaik."
  }')

if echo "$RESPONSE_MEI" | grep -q '"id"'; then
  echo "✅ Mei 2025 berhasil disimpan!"
else
  echo "❌ Error Mei 2025: $RESPONSE_MEI"
fi

echo ""

# Juni 2025
echo "📊 Mengirim data Realisasi Juni 2025..."
RESPONSE_JUN=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-06",
    "is_rkap": false,
    "current_assets": 238000000000,
    "non_current_assets": 446000000000,
    "short_term_liabilities": 115000000000,
    "long_term_liabilities": 185000000000,
    "equity": 384000000000,
    "revenue": 88000000000,
    "operating_expenses": 55000000000,
    "operating_profit": 33000000000,
    "other_income": 4500000000,
    "tax": 9375000000,
    "net_profit": 28125000000,
    "operating_cashflow": 36500000000,
    "investing_cashflow": -15500000000,
    "financing_cashflow": -9000000000,
    "ending_balance": 115000000000,
    "roe": 7.33,
    "roi": 4.69,
    "current_ratio": 2.07,
    "cash_ratio": 1.18,
    "ebitda": 40000000000,
    "ebitda_margin": 45.45,
    "net_profit_margin": 31.96,
    "operating_profit_margin": 37.50,
    "debt_to_equity": 0.78,
    "remark": "Laporan realisasi bulan Juni 2025. Pertengahan tahun dengan performa yang sangat baik, revenue meningkat 3.53% dari bulan sebelumnya. Semua indikator menunjukkan tren positif."
  }')

if echo "$RESPONSE_JUN" | grep -q '"id"'; then
  echo "✅ Juni 2025 berhasil disimpan!"
else
  echo "❌ Error Juni 2025: $RESPONSE_JUN"
fi

echo ""

# Juli 2025
echo "📊 Mengirim data Realisasi Juli 2025..."
RESPONSE_JUL=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-07",
    "is_rkap": false,
    "current_assets": 235000000000,
    "non_current_assets": 444000000000,
    "short_term_liabilities": 114000000000,
    "long_term_liabilities": 186000000000,
    "equity": 379000000000,
    "revenue": 78000000000,
    "operating_expenses": 52500000000,
    "operating_profit": 25500000000,
    "other_income": 4000000000,
    "tax": 7375000000,
    "net_profit": 22125000000,
    "operating_cashflow": 32000000000,
    "investing_cashflow": -16500000000,
    "financing_cashflow": -9200000000,
    "ending_balance": 110000000000,
    "roe": 5.84,
    "roi": 3.70,
    "current_ratio": 2.06,
    "cash_ratio": 1.15,
    "ebitda": 35000000000,
    "ebitda_margin": 44.87,
    "net_profit_margin": 28.37,
    "operating_profit_margin": 32.69,
    "debt_to_equity": 0.79,
    "remark": "Laporan realisasi bulan Juli 2025. Penurunan revenue 11.36% dari bulan sebelumnya karena periode libur dan penyesuaian operasional musiman. Margin tetap terjaga dengan baik."
  }')

if echo "$RESPONSE_JUL" | grep -q '"id"'; then
  echo "✅ Juli 2025 berhasil disimpan!"
else
  echo "❌ Error Juli 2025: $RESPONSE_JUL"
fi

echo ""

# Agustus 2025
echo "📊 Mengirim data Realisasi Agustus 2025..."
RESPONSE_AGU=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-08",
    "is_rkap": false,
    "current_assets": 240000000000,
    "non_current_assets": 448000000000,
    "short_term_liabilities": 116000000000,
    "long_term_liabilities": 184000000000,
    "equity": 388000000000,
    "revenue": 90000000000,
    "operating_expenses": 56000000000,
    "operating_profit": 34000000000,
    "other_income": 4800000000,
    "tax": 9700000000,
    "net_profit": 29100000000,
    "operating_cashflow": 37500000000,
    "investing_cashflow": -15000000000,
    "financing_cashflow": -9500000000,
    "ending_balance": 118000000000,
    "roe": 7.50,
    "roi": 4.85,
    "current_ratio": 2.07,
    "cash_ratio": 1.20,
    "ebitda": 41000000000,
    "ebitda_margin": 45.56,
    "net_profit_margin": 32.33,
    "operating_profit_margin": 37.78,
    "debt_to_equity": 0.77,
    "remark": "Laporan realisasi bulan Agustus 2025. Peningkatan revenue 15.38% dari bulan sebelumnya. Recovery yang kuat setelah periode libur dengan performa yang sangat baik."
  }')

if echo "$RESPONSE_AGU" | grep -q '"id"'; then
  echo "✅ Agustus 2025 berhasil disimpan!"
else
  echo "❌ Error Agustus 2025: $RESPONSE_AGU"
fi

echo ""

# September 2025
echo "📊 Mengirim data Realisasi September 2025..."
RESPONSE_SEP=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-09",
    "is_rkap": false,
    "current_assets": 242000000000,
    "non_current_assets": 450000000000,
    "short_term_liabilities": 118000000000,
    "long_term_liabilities": 182000000000,
    "equity": 392000000000,
    "revenue": 86000000000,
    "operating_expenses": 55500000000,
    "operating_profit": 30500000000,
    "other_income": 4500000000,
    "tax": 8750000000,
    "net_profit": 26250000000,
    "operating_cashflow": 36000000000,
    "investing_cashflow": -14500000000,
    "financing_cashflow": -9800000000,
    "ending_balance": 120000000000,
    "roe": 6.70,
    "roi": 4.38,
    "current_ratio": 2.05,
    "cash_ratio": 1.19,
    "ebitda": 39500000000,
    "ebitda_margin": 45.93,
    "net_profit_margin": 30.52,
    "operating_profit_margin": 35.47,
    "debt_to_equity": 0.77,
    "remark": "Laporan realisasi bulan September 2025. Penurunan revenue 4.44% dari bulan sebelumnya karena penyesuaian operasional dan maintenance rutin. Margin tetap stabil."
  }')

if echo "$RESPONSE_SEP" | grep -q '"id"'; then
  echo "✅ September 2025 berhasil disimpan!"
else
  echo "❌ Error September 2025: $RESPONSE_SEP"
fi

echo ""

# Oktober 2025
echo "📊 Mengirim data Realisasi Oktober 2025..."
RESPONSE_OKT=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-10",
    "is_rkap": false,
    "current_assets": 245000000000,
    "non_current_assets": 452000000000,
    "short_term_liabilities": 119000000000,
    "long_term_liabilities": 181000000000,
    "equity": 397000000000,
    "revenue": 92000000000,
    "operating_expenses": 57000000000,
    "operating_profit": 35000000000,
    "other_income": 5000000000,
    "tax": 10000000000,
    "net_profit": 30000000000,
    "operating_cashflow": 38500000000,
    "investing_cashflow": -14000000000,
    "financing_cashflow": -10000000000,
    "ending_balance": 125000000000,
    "roe": 7.56,
    "roi": 5.05,
    "current_ratio": 2.06,
    "cash_ratio": 1.22,
    "ebitda": 42000000000,
    "ebitda_margin": 45.65,
    "net_profit_margin": 32.61,
    "operating_profit_margin": 38.04,
    "debt_to_equity": 0.76,
    "remark": "Laporan realisasi bulan Oktober 2025. Peningkatan revenue 6.98% dari bulan sebelumnya. Performa sangat baik dengan margin yang meningkat signifikan."
  }')

if echo "$RESPONSE_OKT" | grep -q '"id"'; then
  echo "✅ Oktober 2025 berhasil disimpan!"
else
  echo "❌ Error Oktober 2025: $RESPONSE_OKT"
fi

echo ""

# November 2025
echo "📊 Mengirim data Realisasi November 2025..."
RESPONSE_NOV=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-11",
    "is_rkap": false,
    "current_assets": 248000000000,
    "non_current_assets": 454000000000,
    "short_term_liabilities": 120000000000,
    "long_term_liabilities": 180000000000,
    "equity": 402000000000,
    "revenue": 88000000000,
    "operating_expenses": 56500000000,
    "operating_profit": 31500000000,
    "other_income": 4800000000,
    "tax": 9075000000,
    "net_profit": 27225000000,
    "operating_cashflow": 37000000000,
    "investing_cashflow": -14500000000,
    "financing_cashflow": -9500000000,
    "ending_balance": 128000000000,
    "roe": 6.77,
    "roi": 4.54,
    "current_ratio": 2.07,
    "cash_ratio": 1.24,
    "ebitda": 40500000000,
    "ebitda_margin": 46.02,
    "net_profit_margin": 30.94,
    "operating_profit_margin": 35.80,
    "debt_to_equity": 0.75,
    "remark": "Laporan realisasi bulan November 2025. Penurunan revenue 4.35% dari bulan sebelumnya karena penyesuaian operasional menjelang akhir tahun. Margin tetap terjaga dengan baik."
  }')

if echo "$RESPONSE_NOV" | grep -q '"id"'; then
  echo "✅ November 2025 berhasil disimpan!"
else
  echo "❌ Error November 2025: $RESPONSE_NOV"
fi

echo ""

# ============================================
# 3. Input Realisasi Bulanan - Desember 2025
# ============================================
echo "📊 Mengirim data Realisasi Desember 2025..."
RESPONSE_DEC=$(curl -s -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-CSRF-Token: ${CSRF_TOKEN}" \
  -b /tmp/cookies.txt \
  -d '{
    "company_id": "7ebf36d1-9541-4de5-8f80-95768fa00b8e",
    "year": "2025",
    "period": "2025-12",
    "is_rkap": false,
    "current_assets": 275000000000,
    "non_current_assets": 458000000000,
    "short_term_liabilities": 124000000000,
    "long_term_liabilities": 176000000000,
    "equity": 433000000000,
    "revenue": 102000000000,
    "operating_expenses": 60000000000,
    "operating_profit": 42000000000,
    "other_income": 6000000000,
    "tax": 12000000000,
    "net_profit": 36000000000,
    "operating_cashflow": 46000000000,
    "investing_cashflow": -12500000000,
    "financing_cashflow": -10500000000,
    "ending_balance": 145000000000,
    "roe": 8.31,
    "roi": 5.40,
    "current_ratio": 2.22,
    "cash_ratio": 1.38,
    "ebitda": 50000000000,
    "ebitda_margin": 49.02,
    "net_profit_margin": 35.29,
    "operating_profit_margin": 41.18,
    "debt_to_equity": 0.69,
    "remark": "Laporan realisasi bulan Desember 2025. Penutupan tahun dengan performa yang sangat baik, revenue meningkat 15.91% dari bulan November. Puncak performa tahunan dengan semua indikator keuangan menunjukkan tren positif yang kuat."
  }')

if echo "$RESPONSE_DEC" | grep -q '"id"'; then
  echo "✅ Desember 2025 berhasil disimpan!"
else
  echo "❌ Error Desember 2025: $RESPONSE_DEC"
fi

echo ""
echo "✅ Selesai! Data untuk tahun 2025 (Januari-Desember) sudah dikirim."
echo "💡 Refresh halaman di browser untuk melihat data."
echo "📊 Total: 1 RKAP + 12 Realisasi Bulanan = 13 data financial reports"

# Cleanup
rm -f /tmp/cookies.txt
