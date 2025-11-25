# Mengubah Password Superadmin

Password superadmin dapat dikonfigurasi melalui 3 cara (dalam urutan prioritas):

## 1. Via Vault (Recommended untuk Production)

### Menyimpan Password di Vault

**Opsi A: Menggunakan script**
```bash
cd backend
SUPERADMIN_PASSWORD="PasswordBaru123!" ./scripts/store-all-secrets.sh
```

**Opsi B: Langsung via Vault CLI**
```bash
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-root-token-12345 dms-vault-dev vault kv put secret/dms-app superadmin_password="PasswordBaru123!"
```

**Opsi C: Via Vault Web UI**
1. Buka http://127.0.0.1:8200/ui
2. Login dengan token: `dev-root-token-12345`
3. Navigate ke: **Secrets Engines** → **secret** → **dms-app**
4. Klik **"Create new version +"**
5. Tambahkan/update field `superadmin_password` dengan password baru
6. Klik **Save**

### Verifikasi Password di Vault
```bash
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-root-token-12345 dms-vault-dev vault kv get secret/dms-app | grep superadmin_password
```

## 2. Via Environment Variable

Set environment variable sebelum menjalankan backend:
```bash
export SUPERADMIN_PASSWORD="PasswordBaru123!"
make dev
```

Atau di `docker-compose.dev.yml`:
```yaml
services:
  backend:
    environment:
      - SUPERADMIN_PASSWORD=PasswordBaru123!
```

## 3. Default (Hardcoded - Development Only)

Jika tidak ada di Vault dan tidak ada environment variable, sistem akan menggunakan password default:
- **Default password**: `Pedeve123`
- ⚠️ **WARNING**: Jangan gunakan default password di production!

## Update Password Superadmin yang Sudah Ada

Jika superadmin sudah ada di database, ada beberapa cara untuk mengupdate password:

### Opsi 1: Auto-Sync dari Vault (Recommended)

Enable auto-sync saat startup dengan menambahkan environment variable:

**Via docker-compose.dev.yml:**
```yaml
services:
  backend:
    environment:
      - SUPERADMIN_AUTO_SYNC_PASSWORD=true
```

**Via environment variable:**
```bash
export SUPERADMIN_AUTO_SYNC_PASSWORD=true
make dev
```

Dengan ini, setiap kali backend start, password superadmin akan otomatis di-sync dari Vault jika berbeda.

### Opsi 2: Script Update Manual

Jalankan script untuk update password dari Vault:

```bash
cd backend
./scripts/update-superadmin-password.sh
```

Script ini akan:
1. Mengambil password dari Vault
2. Update password superadmin di database
3. Password langsung bisa digunakan untuk login

### Opsi 3: Update via Database (Manual)

Jika perlu update langsung via database:

```bash
# 1. Hash password baru terlebih dahulu (gunakan backend)
docker exec dms-backend-dev go run -c 'package main; import ("fmt"; "github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/password"); func main() { hash, _ := password.HashPassword("PasswordBaru123!"); fmt.Println(hash) }'

# 2. Update di database
docker exec -it dms-postgres-dev psql -U postgres -d db_dms_pedeve -c "UPDATE users SET password = '<hashed_password>' WHERE username = 'superadmin';"
```

## Catatan Penting

1. **Auto-Sync**: Jika `SUPERADMIN_AUTO_SYNC_PASSWORD=true`, password akan otomatis di-sync dari Vault setiap startup jika berbeda. Ini berguna untuk production.
2. **Manual Update**: Gunakan script `update-superadmin-password.sh` untuk update manual tanpa restart.
3. **Security**: Pastikan Vault token dan akses database aman.

2. **Password Requirements**: 
   - Minimum 8 karakter
   - Disarankan menggunakan kombinasi huruf besar, huruf kecil, angka, dan simbol

3. **Security Best Practice**:
   - Gunakan Vault untuk production
   - Jangan commit password ke git
   - Rotate password secara berkala
   - Gunakan password yang kuat dan unik

## Troubleshooting

**Password tidak berubah setelah update di Vault?**
- Pastikan backend sudah di-restart setelah update Vault
- Pastikan superadmin belum ada di database (hapus dulu jika perlu)
- Cek log backend untuk melihat dari mana password di-load

**Cek log untuk melihat source password:**
```bash
# Log akan menunjukkan:
# "Superadmin password loaded from Vault" - jika dari Vault
# "Superadmin password loaded from environment variable" - jika dari env var
# "Using default superadmin password..." - jika menggunakan default
```

