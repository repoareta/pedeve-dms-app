import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Pedeve DMS User Guide',
  description: 'Panduan lengkap penggunaan aplikasi Pedeve Document Management System',
  base: '/user-guideline/',
  ignoreDeadLinks: true, // Ignore dead links during build
  
  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'Pedeve DMS Guide',
    
    nav: [
      { text: 'Beranda', link: '/' },
      { text: 'Memulai', link: '/guide/getting-started' },
      { text: 'Fitur', link: '/features/company-management' }
    ],
    
    sidebar: {
      '/guide/': [
        {
          text: 'Memulai',
          items: [
            { text: 'Pengenalan', link: '/guide/getting-started' },
            { text: 'Login & Autentikasi', link: '/guide/authentication' },
            { text: 'Halaman Daftar Anak Perusahaan', link: '/guide/dashboard' }
          ]
        }
      ],
      '/features/': [
        {
          text: 'Manajemen Perusahaan',
          items: [
            { text: 'Daftar Perusahaan', link: '/features/company-management' },
            { text: 'Tambah Perusahaan', link: '/features/add-company' },
            { text: 'Detail Perusahaan', link: '/features/company-detail' },
            { text: 'Edit Perusahaan', link: '/features/edit-company' },
            { text: 'My Company', link: '/features/my-company' }
          ]
        },
        {
          text: 'Manajemen Dokumen',
          items: [
            { text: 'Daftar Dokumen', link: '/features/document-management' },
            { text: 'Upload Dokumen', link: '/features/upload-document' },
            { text: 'Detail Dokumen', link: '/features/document-detail' },
            { text: 'Folder & Kategori', link: '/features/document-folders' }
          ]
        },
        {
          text: 'Laporan Keuangan',
          items: [
            { text: 'Daftar Laporan', link: '/features/financial-reports' },
            { text: 'Tambah Laporan', link: '/features/add-report' },
            { text: 'Edit Laporan', link: '/features/edit-report' },
            { text: 'Bulk Upload', link: '/features/bulk-upload-reports' }
          ]
        },
        {
          text: 'Manajemen Pengguna',
          items: [
            { text: 'Daftar Pengguna', link: '/features/user-management' },
            { text: 'Tambah Pengguna', link: '/features/add-user' },
            { text: 'Edit Pengguna', link: '/features/edit-user' }
          ]
        },
        {
          text: 'Notifikasi',
          items: [
            { text: 'Daftar Notifikasi', link: '/features/notifications' },
            { text: 'Mark as Read', link: '/features/mark-notifications-read' }
          ]
        },
        {
          text: 'Pengaturan',
          items: [
            { text: 'Profil', link: '/features/profile' },
            { text: 'Settings', link: '/features/settings' },
            { text: 'Two-Factor Authentication', link: '/features/2fa' }
          ]
        }
      ]
    },
    
    socialLinks: [
      // Add social links if needed
    ],
    
    footer: {
      message: 'Pedeve DMS App - Document Management System',
      copyright: 'Copyright © 2025'
    },
    
    search: {
      provider: 'local'
    }
  }
})
