#!/bin/bash

# ============================================
# Curl Commands untuk Financial Reports
# Company ID: bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d
# ============================================

# Ganti dengan base URL API Anda
API_BASE_URL="http://localhost:8080/api/v1"

# ============================================
# PENTING: SET TOKEN TERLEBIH DAHULU!
# ============================================
# Cara mendapatkan token:
# 1. Login: curl -X POST http://localhost:8080/api/v1/auth/login \
#      -H "Content-Type: application/json" \
#      -d '{"username": "admin", "password": "admin"}'
# 2. Copy token dari response (field "token")
# 3. Ganti YOUR_AUTH_TOKEN_HERE di bawah dengan token yang didapat
#
# ATAU gunakan script otomatis: ./get-token-and-send-data.sh
# ============================================
AUTH_TOKEN="YOUR_AUTH_TOKEN_HERE"

# Cek apakah token sudah di-set
if [ "$AUTH_TOKEN" = "YOUR_AUTH_TOKEN_HERE" ]; then
  echo "❌ ERROR: Token belum di-set!"
  echo ""
  echo "💡 Cara mendapatkan token:"
  echo "   1. Login: curl -X POST ${API_BASE_URL}/auth/login -H 'Content-Type: application/json' -d '{\"username\": \"admin\", \"password\": \"admin\"}'"
  echo "   2. Copy token dari response"
  echo "   3. Edit file ini dan ganti YOUR_AUTH_TOKEN_HERE dengan token Anda"
  echo ""
  echo "   ATAU gunakan script otomatis: ./get-token-and-send-data.sh"
  exit 1
fi

# ============================================
# 1. Input Realisasi Bulanan - November 2024
# ============================================
echo "📊 Mengirim data Realisasi November 2024..."
curl -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -d '{
    "company_id": "bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d",
    "year": "2024",
    "period": "2024-11",
    "is_rkap": false,
    "current_assets": 250000000000,
    "non_current_assets": 450000000000,
    "short_term_liabilities": 120000000000,
    "long_term_liabilities": 180000000000,
    "equity": 400000000000,
    "revenue": 85000000000,
    "operating_expenses": 55000000000,
    "operating_profit": 30000000000,
    "other_income": 5000000000,
    "tax": 8750000000,
    "net_profit": 26250000000,
    "operating_cashflow": 35000000000,
    "investing_cashflow": -15000000000,
    "financing_cashflow": -8000000000,
    "ending_balance": 120000000000,
    "roe": 6.56,
    "roi": 4.38,
    "current_ratio": 2.08,
    "cash_ratio": 1.25,
    "ebitda": 38000000000,
    "ebitda_margin": 44.71,
    "net_profit_margin": 30.88,
    "operating_profit_margin": 35.29,
    "debt_to_equity": 0.75,
    "remark": "Laporan realisasi bulan November 2024. Performa operasional stabil dengan peningkatan revenue sebesar 5% dari bulan sebelumnya."
  }'

echo -e "\n\n"

# ============================================
# 2. Input Realisasi Bulanan - Desember 2024
# ============================================
echo "📊 Mengirim data Realisasi Desember 2024..."
curl -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -d '{
    "company_id": "bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d",
    "year": "2024",
    "period": "2024-12",
    "is_rkap": false,
    "current_assets": 280000000000,
    "non_current_assets": 460000000000,
    "short_term_liabilities": 125000000000,
    "long_term_liabilities": 175000000000,
    "equity": 440000000000,
    "revenue": 95000000000,
    "operating_expenses": 58000000000,
    "operating_profit": 37000000000,
    "other_income": 6000000000,
    "tax": 10750000000,
    "net_profit": 32250000000,
    "operating_cashflow": 42000000000,
    "investing_cashflow": -12000000000,
    "financing_cashflow": -10000000000,
    "ending_balance": 140000000000,
    "roe": 7.33,
    "roi": 4.78,
    "current_ratio": 2.24,
    "cash_ratio": 1.40,
    "ebitda": 45000000000,
    "ebitda_margin": 47.37,
    "net_profit_margin": 33.95,
    "operating_profit_margin": 38.95,
    "debt_to_equity": 0.68,
    "remark": "Laporan realisasi bulan Desember 2024. Penutupan tahun dengan performa yang sangat baik, revenue meningkat 11.76% dari bulan November. Semua indikator keuangan menunjukkan tren positif."
  }'

echo -e "\n\n"

# ============================================
# 3. Input RKAP Tahunan - Tahun 2024 (Opsional)
# ============================================
echo "📊 Mengirim data RKAP 2024 (opsional)..."
curl -X POST "${API_BASE_URL}/financial-reports" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -d '{
    "company_id": "bfcccf29-08d9-4b6c-9f88-dfe836ab1c1d",
    "year": "2024",
    "period": "2024",
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
    "remark": "RKAP tahun 2024. Target anggaran tahunan yang telah disetujui."
  }'

echo -e "\n\n"
echo "✅ Selesai!"
