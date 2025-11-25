# Cara Edit Rate Limit Config di Vault UI

## Overview

Di HashiCorp Vault, secret tidak bisa langsung di-edit. Vault menggunakan **versioning system**, jadi untuk update secret, kita perlu membuat **version baru**.

---

## Cara Edit Rate Limit Config via Vault UI

### Method 1: Create New Version (Recommended)

1. **Buka Vault UI**: http://127.0.0.1:8200/ui
2. **Login** dengan token: `dev-root-token-12345`
3. **Navigate** ke: `Secrets` → `secret` → `kv` → `dms-app`
4. **Klik tab "Secret"** (jika belum aktif)
5. **Klik tombol "Create new version +"** (di kanan atas)
6. **Edit value** di field `rate_limit`:
   ```json
   {
     "general": {
       "rps": 100,
       "burst": 100
     },
     "auth": {
       "rpm": 5,
       "burst": 5
     },
     "strict": {
       "rpm": 50,
       "burst": 50
     }
   }
   ```
7. **Klik "Save"** atau "Create version"
8. **Restart backend** untuk load config baru

---

### Method 2: Via Vault CLI

```bash
# Edit config
docker exec dms-vault-dev vault kv put secret/dms-app \
  rate_limit='{"general":{"rps":100,"burst":100},"auth":{"rpm":5,"burst":5},"strict":{"rpm":50,"burst":50}}'

# Verify
docker exec dms-vault-dev vault kv get secret/dms-app
```

---

### Method 3: Via Script

```bash
# Edit dengan custom values
RATE_LIMIT_GENERAL_RPS=100 \
RATE_LIMIT_GENERAL_BURST=100 \
./backend/scripts/store-rate-limit-config.sh
```

---

## Catatan Penting

### 1. Versioning System
- Vault menyimpan **semua versi** secret
- Setiap update membuat **version baru**
- Bisa rollback ke version sebelumnya via "Version History" tab

### 2. Backend Auto-Reload
- **Saat ini**: Backend perlu di-restart untuk load config baru
- **Future enhancement**: Bisa ditambahkan hot-reload dengan polling atau webhook

### 3. Format JSON
- Pastikan JSON **valid** saat edit
- Gunakan JSON formatter untuk memastikan format benar
- Contoh valid:
  ```json
  {"general":{"rps":500,"burst":500},"auth":{"rpm":5,"burst":5},"strict":{"rpm":50,"burst":50}}
  ```

---

## Troubleshooting

### Config tidak ter-load setelah update
1. **Cek Vault path**: Pastikan `VAULT_SECRET_PATH` benar
2. **Cek Vault connection**: Pastikan backend bisa connect ke Vault
3. **Restart backend**: Config di-load saat startup
4. **Cek logs**: Lihat log backend untuk error messages

### JSON invalid error
- Gunakan JSON validator online
- Pastikan semua quotes menggunakan double quotes (`"`)
- Pastikan tidak ada trailing comma

---

## Best Practices

1. **Test di development** sebelum update production
2. **Backup config** sebelum update (via Version History)
3. **Document changes** di commit message atau changelog
4. **Monitor logs** setelah update untuk memastikan config ter-load

---

## Future Enhancement Ideas

1. **Hot-reload**: Auto-reload config tanpa restart
2. **Validation**: Validate config sebelum save
3. **Notifications**: Notify backend saat config berubah
4. **Rollback UI**: Easy rollback dari Vault UI

