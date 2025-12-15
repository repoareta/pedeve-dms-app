# Cara Menggunakan Sample Data Financial Report

File `sample-financial-reports-nov-dec.json` berisi sample data untuk:
- **November 2024** (Realisasi Bulanan)
- **Desember 2024** (Realisasi Bulanan)
- **RKAP 2024** (untuk referensi perbandingan)

## Cara 1: Via Frontend UI

1. Buka aplikasi dan navigasi ke halaman **Subsidiary Detail**
2. Pilih perusahaan yang ingin diisi datanya
3. Klik tab **"Input Laporan"**
4. Pilih sub-tab **"Input Realisasi (Bulanan)"**
5. Pilih:
   - **Tahun**: 2024
   - **Bulan**: November (untuk data November) atau Desember (untuk data Desember)
6. Isi semua field sesuai dengan data dari file JSON
7. Klik **"Simpan"**

## Cara 2: Via API (cURL)

### Untuk November 2024:
```bash
curl -X POST http://localhost:8080/api/v1/financial-reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "company_id": "YOUR_COMPANY_ID",
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
```

### Untuk Desember 2024:
```bash
curl -X POST http://localhost:8080/api/v1/financial-reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "company_id": "YOUR_COMPANY_ID",
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
```

## Catatan Penting

1. **Ganti `YOUR_COMPANY_ID`** dengan ID perusahaan yang sebenarnya dari database
2. **Ganti `YOUR_TOKEN`** dengan token autentikasi yang valid
3. **Pastikan periode belum ada datanya** - sistem tidak mengizinkan duplikasi periode untuk perusahaan yang sama
4. **Format nilai**: Semua nilai dalam Rupiah (integer, tanpa desimal kecuali untuk rasio)
5. **Format rasio**: Nilai rasio dalam persentase (contoh: 6.56 berarti 6.56%)

## Struktur Data

Setiap data financial report terdiri dari:

### A. Neraca (Balance Sheet)
- `current_assets`: Aset Lancar
- `non_current_assets`: Aset Tidak Lancar
- `short_term_liabilities`: Liabilitas Jangka Pendek
- `long_term_liabilities`: Liabilitas Jangka Panjang
- `equity`: Ekuitas

### B. Laba Rugi (Profit & Loss)
- `revenue`: Revenue
- `operating_expenses`: Beban Usaha
- `operating_profit`: Laba Usaha
- `other_income`: Pendapatan Lain-Lain
- `tax`: Tax
- `net_profit`: Laba Bersih

### C. Cashflow
- `operating_cashflow`: Arus kas bersih dari operasi
- `investing_cashflow`: Arus kas bersih dari investasi
- `financing_cashflow`: Arus kas bersih dari pendanaan
- `ending_balance`: Saldo Akhir

### D. Rasio Keuangan (%)
- `roe`: Return on Equity
- `roi`: Return on Investment
- `current_ratio`: Rasio Lancar
- `cash_ratio`: Rasio Kas
- `ebitda`: EBITDA
- `ebitda_margin`: EBITDA Margin
- `net_profit_margin`: Net Profit Margin
- `operating_profit_margin`: Operating Profit Margin
- `debt_to_equity`: Debt to Equity

## Validasi Data

Data sample ini sudah dirancang dengan:
- ✅ Nilai yang realistis dan konsisten
- ✅ Rasio keuangan yang masuk akal
- ✅ Tren peningkatan dari November ke Desember
- ✅ Semua field yang required sudah terisi

Setelah data diinput, Anda dapat melihat perbandingan RKAP vs Realisasi di tab **"Performance"** pada halaman Subsidiary Detail.
