# 🌐 Frontend Production Domain Configuration

## Domain Baru
- **Domain:** `dms.pertamina-pedeve.co.id`
- **VM Name:** `frontend-prod-2`
- **Zone:** `asia-southeast2-a`

## 🔍 Cara Cek IP Address Frontend Production VM

### Option 1: Via GCP Console
1. Buka [GCP Console](https://console.cloud.google.com)
2. Pilih project production
3. Navigate ke **Compute Engine** → **VM instances**
4. Cari VM `frontend-prod-2`
5. Lihat kolom **External IP**

### Option 2: Via gcloud CLI
```bash
gcloud compute instances describe frontend-prod-2 \
  --zone=asia-southeast2-a \
  --project=<PROJECT_ID_PROD> \
  --format="get(networkInterfaces[0].accessConfigs[0].natIP)"
```

### Option 3: List semua VM dengan IP
```bash
gcloud compute instances list \
  --filter="name:frontend-prod-2" \
  --format="table(name,zone,EXTERNAL_IP)" \
  --project=<PROJECT_ID_PROD>
```

## 📝 DNS Configuration

Setelah mendapatkan IP address, buat A record di DNS provider:

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | `dms` | `<IP_ADDRESS>` | 3600 |

**Contoh:**
- **Domain:** `dms.pertamina-pedeve.co.id`
- **Type:** `A`
- **Value:** `<IP_ADDRESS_FROM_GCP>`
- **TTL:** `3600` (1 hour)

## ✅ Checklist

- [ ] Dapatkan IP address dari GCP Console atau gcloud CLI
- [ ] Buat A record di DNS provider untuk `dms.pertamina-pedeve.co.id`
- [ ] Tunggu DNS propagation (biasanya 5-15 menit)
- [ ] Verifikasi DNS: `nslookup dms.pertamina-pedeve.co.id` atau `dig dms.pertamina-pedeve.co.id`
- [ ] Deploy frontend dengan domain baru (akan otomatis setup SSL via Let's Encrypt)
- [ ] Test akses: `https://dms.pertamina-pedeve.co.id`

## 🔄 Files yang Sudah Diupdate

1. ✅ `.github/workflows/ci-cd.yml` - Domain di production deployment
2. ✅ `scripts/setup-nginx-frontend.sh` - Comment updated

## 📌 Catatan

- Domain lama: `reports.pertamina-pedeve.co.id` (sudah diganti)
- Domain baru: `dms.pertamina-pedeve.co.id`
- SSL certificate akan otomatis dibuat saat deployment pertama kali dengan domain baru
- Pastikan DNS sudah pointing ke IP yang benar sebelum deployment
