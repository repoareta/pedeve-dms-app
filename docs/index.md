# Pedeve DMS User Guide

Selamat datang di panduan penggunaan aplikasi **Pedeve Document Management System (DMS)**.

## Tentang Aplikasi

Pedeve DMS adalah sistem manajemen dokumen yang dirancang untuk mengelola dokumen dan data perusahaan dalam struktur hierarki multi-level. Aplikasi ini menyediakan fitur lengkap untuk:

- **Manajemen Perusahaan** - Mengelola data perusahaan dengan hierarki multi-level
- **Manajemen Dokumen** - Upload, kategorisasi, dan tracking dokumen
- **Laporan Keuangan** - Input dan analisis laporan keuangan (RKAP & Realisasi)
- **Manajemen Pengguna** - Kontrol akses berbasis peran (RBAC)
- **Notifikasi** - Sistem notifikasi untuk berbagai event

## Mulai Cepat

1. **[Login & Autentikasi](/user-guideline/guide/authentication)** - Pelajari cara login dan setup 2FA
2. **[Halaman Daftar Anak Perusahaan](/user-guideline/guide/dashboard)** - Kenali halaman utama setelah login
3. **[Manajemen Perusahaan](/user-guideline/features/company-management)** - Mulai mengelola perusahaan

## Fitur Utama

<div class="feature-grid">
  <div class="feature-card">
    <h3>🏢 Manajemen Perusahaan</h3>
    <p>Kelola data perusahaan dengan hierarki multi-level, pemegang saham, dan dewan direksi</p>
    <a href="/user-guideline/features/company-management">Pelajari lebih lanjut →</a>
  </div>
  
  <div class="feature-card">
    <h3>📄 Manajemen Dokumen</h3>
    <p>Upload, kategorisasi, dan tracking dokumen dengan sistem folder</p>
    <a href="/user-guideline/features/document-management">Pelajari lebih lanjut →</a>
  </div>
  
  <div class="feature-card">
    <h3>📊 Laporan Keuangan</h3>
    <p>Input laporan keuangan, bulk upload via Excel, dan analisis RKAP vs Realisasi</p>
    <a href="/user-guideline/features/financial-reports">Pelajari lebih lanjut →</a>
  </div>
  
  <div class="feature-card">
    <h3>👥 Manajemen Pengguna</h3>
    <p>Kelola pengguna dengan kontrol akses berbasis peran (RBAC)</p>
    <a href="/user-guideline/features/user-management">Pelajari lebih lanjut →</a>
  </div>
</div>

## Peran Pengguna

Aplikasi ini mendukung beberapa peran dengan level akses yang berbeda:

- **Administrator** - Akses penuh ke semua fitur
- **Admin** - Akses terbatas pada perusahaan yang di-assign
- **Manager** - Akses terbatas untuk melihat dan mengelola data perusahaan
- **Staff** - Akses terbatas untuk input data

## Butuh Bantuan?

Jika Anda mengalami kesulitan atau memiliki pertanyaan, silakan hubungi administrator sistem.

---

**Versi Dokumentasi:** 1.0.0  
**Terakhir Diupdate:** 2025

<style>
.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}

.feature-card {
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  padding: 1.5rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.feature-card h3 {
  margin-top: 0;
  color: var(--vp-c-brand);
}

.feature-card a {
  color: var(--vp-c-brand);
  text-decoration: none;
  font-weight: 500;
}

.feature-card a:hover {
  text-decoration: underline;
}
</style>
