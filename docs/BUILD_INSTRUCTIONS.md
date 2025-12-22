# Build Instructions - User Guideline

## Status Setup

✅ **Dependencies sudah diinstall**
✅ **Build sudah berhasil**
✅ **File sudah di-copy ke `frontend/public/user-guideline/`**

## Testing

### Saat `make dev` Berjalan

Karena Anda sedang menjalankan `make dev`, dokumentasi sudah bisa diakses langsung:

1. **Buka Browser**
2. **Akses**: http://localhost:5173/user-guideline/
3. **Tidak perlu login** - dokumentasi public access

### Verifikasi

- ✅ File sudah di `frontend/public/user-guideline/`
- ✅ Vite akan serve file dari `public/` otomatis
- ✅ Tidak perlu restart `make dev`
- ✅ Cukup refresh browser

## Update Dokumentasi

Jika ingin update dokumentasi:

1. **Edit file Markdown** di `docs/`
2. **Build ulang**:
   ```bash
   make docs-build
   ```
3. **Refresh browser** untuk melihat perubahan

## Catatan

- **Tidak perlu run dev terpisah** untuk docs saat menggunakan `make dev`
- File di `public/` akan di-serve otomatis oleh Vite
- Build output tidak perlu di-commit (akan di-build saat deployment)
