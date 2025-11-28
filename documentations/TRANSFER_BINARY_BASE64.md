# 📤 Transfer Binary via Base64

## Masalah
- SSH key tidak terkonfigurasi
- `scp` tidak bisa digunakan

## Solusi: Base64 Encode/Decode

### Step 1: Encode Binary di Local

**Di local machine (masih di directory backend):**

```bash
# Encode binary ke base64
base64 seed-companies > seed-companies.b64

# Cek ukuran file (untuk reference)
ls -lh seed-companies.b64
```

### Step 2: Copy Base64 Text ke VM

**Opsi A: Copy-paste manual**
1. Buka file `seed-companies.b64` di text editor
2. Copy semua isinya
3. SSH ke VM dan paste ke file

**Opsi B: Split file jika terlalu besar**
Jika file terlalu besar untuk copy-paste, split menjadi beberapa bagian:

```bash
# Split file menjadi chunks 1MB
split -b 1M seed-companies.b64 seed-companies.b64.part
```

### Step 3: Decode di VM

**SSH ke VM dan jalankan:**

```bash
# Opsi A: Jika copy-paste manual
cat > ~/seed-companies.b64 << 'EOF'
# Paste isi file seed-companies.b64 di sini
EOF

# Decode
base64 -d ~/seed-companies.b64 > ~/seed-companies
chmod +x ~/seed-companies

# Verifikasi
ls -lh ~/seed-companies
file ~/seed-companies
```

## Quick Script untuk Copy-Paste

**Di local machine, generate script untuk di-copy-paste:**

```bash
# Generate script dengan base64 embedded
echo "cat > ~/seed-companies.b64 << 'SEEDEOF'" > transfer-seeder.sh
base64 seed-companies >> transfer-seeder.sh
echo "SEEDEOF" >> transfer-seeder.sh
echo "base64 -d ~/seed-companies.b64 > ~/seed-companies" >> transfer-seeder.sh
echo "chmod +x ~/seed-companies" >> transfer-seeder.sh
echo "ls -lh ~/seed-companies" >> transfer-seeder.sh

# File transfer-seeder.sh bisa di-copy ke VM dan dijalankan
```

