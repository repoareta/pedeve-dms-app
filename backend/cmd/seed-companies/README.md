# Company Seeder

Seeder untuk membuat sample data companies dengan hierarki 3 layer dan user admin untuk setiap company.

## Struktur Data

- **1 Holding Company**: Pedeve Pertamina (Level 0)
- **5 Level 1 Companies**: Anak perusahaan langsung dari holding
- **3 Level 2 Companies**: Cucu perusahaan (anak dari level 1)
- **2 Level 3 Companies**: Cicit perusahaan (anak dari level 2)

**Total: 11 Companies (1 holding + 10 subsidiaries)**

Setiap company memiliki 1 user admin dengan password: `admin123`

## Cara Menjalankan

### Prerequisites
- Database PostgreSQL sudah running
- Role "admin" sudah ada di database (biasanya sudah ada dari seed roles)

### Menjalankan Seeder

**Opsi 1: Menggunakan Makefile (Paling Mudah)**
```bash
# Dari root project
make seed-companies
```

**Opsi 2: Menggunakan Script**
```bash
# Dari root project
cd backend/cmd/seed-companies
./seed.sh
```

**Opsi 3: Manual**
```bash
cd backend
DATABASE_URL="postgres://postgres:dms_password@localhost:5432/db_dms_pedeve?sslmode=disable" go run ./cmd/seed-companies
```

Atau jika DATABASE_URL sudah di-set di environment:

```bash
cd backend
go run ./cmd/seed-companies
```

### Catatan

- Seeder akan skip jika holding company sudah ada
- Jika ingin re-seed, hapus semua companies terlebih dahulu atau gunakan fresh database
- Semua user memiliki password default: `admin123`

## Struktur Hierarki

```
Pedeve Pertamina (Holding - Level 0)
├── PT Energi Nusantara (Level 1)
│   ├── PT ENU Exploration (Level 2)
│   │   └── PT ENU-EXP Drilling (Level 3)
│   └── PT ENU Production (Level 2)
│       └── PT ENU-PRO Refinery (Level 3)
├── PT Pertamina Gas (Level 1)
│   └── PT PTG Distribution (Level 2)
├── PT Pertamina Lubricants (Level 1)
├── PT Pertamina Retail (Level 1)
└── PT Pertamina Shipping (Level 1)
```

## Users yang Dibuat

1. `admin.pedeve` - Admin untuk Pedeve Pertamina (Holding)
2. `admin.enu` - Admin untuk PT Energi Nusantara
3. `admin.ptg` - Admin untuk PT Pertamina Gas
4. `admin.plb` - Admin untuk PT Pertamina Lubricants
5. `admin.prt` - Admin untuk PT Pertamina Retail
6. `admin.psh` - Admin untuk PT Pertamina Shipping
7. `admin.enu.exp` - Admin untuk PT ENU Exploration
8. `admin.enu.pro` - Admin untuk PT ENU Production
9. `admin.ptg.dist` - Admin untuk PT PTG Distribution
10. `admin.enu.exp.drl` - Admin untuk PT ENU-EXP Drilling
11. `admin.enu.pro.ref` - Admin untuk PT ENU-PRO Refinery

**Password untuk semua user: `admin123`**

