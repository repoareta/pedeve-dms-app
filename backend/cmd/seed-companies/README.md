# Company Seeder

Seeder untuk membuat sample data companies dengan hierarki 3 layer dan user admin untuk setiap company.

## Struktur Data

- **1 Holding Company**: Pedeve Pertamina (Level 0)
- **5 Level 1 Companies**: Anak perusahaan langsung dari holding
- **3 Level 2 Companies**: Cucu perusahaan (anak dari level 1)
- **2 Level 3 Companies**: Cicit perusahaan (anak dari level 2)

**Total: 11 Companies (1 holding + 10 subsidiaries)**

Setiap company memiliki 1 user admin dengan password: `admin123`

## Implementasi

Seeder ini menggunakan sistem **junction table** (`user_company_assignments`) untuk mengelola hubungan many-to-many antara users dan companies. Setiap user yang dibuat akan memiliki:
- Entry di tabel `users` dengan role "admin"
- Entry di tabel `user_company_assignments` yang menghubungkan user dengan company dan role-nya

Ini memungkinkan:
- Satu user dapat di-assign ke multiple companies dengan role berbeda
- Fleksibilitas dalam manajemen user-company assignments
- Kompatibilitas dengan sistem RBAC yang sudah ada

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

- Seeder akan skip jika holding company sudah ada (dicek berdasarkan code "PDV")
- Jika ingin re-seed, gunakan fitur "Reset Data Subsidiary" di Settings (hanya superadmin) atau hapus semua companies terlebih dahulu
- Semua user memiliki password default: `admin123`
- Seeder menggunakan junction table `user_company_assignments` untuk user-company relationships
- Seeder akan otomatis membuat entry di junction table untuk setiap user yang dibuat

## Struktur Hierarki Lengkap

```
Pedeve Pertamina (Holding - Level 0)
│   Code: PDV
│   User: admin.pedeve
│
├── PT Energi Nusantara (Level 1)
│   │   Code: ENU
│   │   User: admin.enu
│   │
│   ├── PT ENU Exploration (Level 2)
│   │   │   Code: ENU-EXP
│   │   │   User: admin.enu.exp
│   │   │
│   │   └── PT ENU-EXP Drilling (Level 3)
│   │       │   Code: ENU-EXP-DRL
│   │       │   User: admin.enu.exp.drl
│   │
│   └── PT ENU Production (Level 2)
│       │   Code: ENU-PRO
│       │   User: admin.enu.pro
│       │
│       └── PT ENU-PRO Refinery (Level 3)
│           │   Code: ENU-PRO-REF
│           │   User: admin.enu.pro.ref
│
├── PT Pertamina Gas (Level 1)
│   │   Code: PTG
│   │   User: admin.ptg
│   │
│   └── PT PTG Distribution (Level 2)
│       │   Code: PTG-DIST
│       │   User: admin.ptg.dist
│
├── PT Pertamina Lubricants (Level 1)
│   │   Code: PLB
│   │   User: admin.plb
│
├── PT Pertamina Retail (Level 1)
│   │   Code: PRT
│   │   User: admin.prt
│
└── PT Pertamina Shipping (Level 1)
    │   Code: PSH
    │   User: admin.psh
```

### Detail Per Level

**Level 0 (Holding):**
- 1 company: Pedeve Pertamina (PDV)
- 1 user: admin.pedeve

**Level 1 (Anak Perusahaan):**
- 5 companies: ENU, PTG, PLB, PRT, PSH
- 5 users: admin.enu, admin.ptg, admin.plb, admin.prt, admin.psh

**Level 2 (Cucu Perusahaan):**
- 3 companies: ENU-EXP, ENU-PRO, PTG-DIST
- 3 users: admin.enu.exp, admin.enu.pro, admin.ptg.dist

**Level 3 (Cicit Perusahaan):**
- 2 companies: ENU-EXP-DRL, ENU-PRO-REF
- 2 users: admin.enu.exp.drl, admin.enu.pro.ref

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

## Cara Menggunakan via UI (Recommended)

Seeder sekarang dapat dijalankan melalui UI di halaman **Settings** (hanya superadmin):

1. Login sebagai superadmin
2. Buka halaman **Settings**
3. Scroll ke card **"Fitur untuk Development"**
4. Klik **"Jalankan Seeder Data Subsidiary"** untuk membuat sample data
5. Klik **"Reset Data Subsidiary"** untuk menghapus semua data subsidiary (jika perlu re-seed)

**Catatan:**
- Seeder akan otomatis mengecek apakah data sudah ada
- Jika data sudah ada, proses akan dibatalkan untuk mencegah duplikasi
- Reset data akan menghapus semua subsidiary dan user terkait (kecuali superadmin)

## Database Schema

Seeder membuat data di tabel berikut:
- `companies` - Data perusahaan dengan hierarki
- `users` - Data user admin untuk setiap company
- `user_company_assignments` - Junction table untuk user-company relationships

**Junction Table Structure:**
- Setiap user yang dibuat akan memiliki entry di `user_company_assignments`
- Entry ini menghubungkan user dengan company dan role-nya
- Memungkinkan satu user di-assign ke multiple companies dengan role berbeda

