import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: '/dashboard',
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true, title: 'Dashboard' },
    },
    {
      path: '/subsidiaries',
      name: 'subsidiaries',
      component: () => import('../views/SubsidiariesView.vue'),
      meta: { requiresAuth: true, title: 'Anak Perusahaan' },
    },
    {
      path: '/subsidiaries/new',
      name: 'subsidiary-new',
      component: () => import('../views/SubsidiaryFormView.vue'),
      meta: { requiresAuth: true, title: 'Tambah Anak Perusahaan' },
    },
    {
      path: '/subsidiaries/:id',
      name: 'subsidiary-detail',
      component: () => import('../views/SubsidiaryDetailView.vue'),
      meta: { requiresAuth: true, title: 'Detail Anak Perusahaan' },
    },
    {
      path: '/subsidiaries/:id/edit',
      name: 'subsidiary-edit',
      component: () => import('../views/SubsidiaryFormView.vue'),
      meta: { requiresAuth: true, title: 'Edit Anak Perusahaan' },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { requiresGuest: true, title: 'Login' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
      meta: { requiresGuest: true, title: 'Register' },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true, title: 'Pengaturan' },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/ProfileView.vue'),
      meta: { requiresAuth: true, title: 'Profil' },
    },
    {
      path: '/my-company',
      name: 'my-company',
      component: () => import('../views/MyCompanyView.vue'),
      meta: { requiresAuth: true, title: 'My Company' },
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UserManagementView.vue'),
      meta: { requiresAuth: true, title: 'Manajemen Pengguna' },
    },
    {
      path: '/reports',
      name: 'reports',
      component: () => import('../views/ReportsView.vue'),
      meta: { requiresAuth: true, title: 'Laporan' },
    },
    {
      path: '/reports/new',
      name: 'report-new',
      component: () => import('../views/ReportFormView.vue'),
      meta: { requiresAuth: true, title: 'Tambah Laporan' },
    },
    {
      path: '/reports/:id/edit',
      name: 'report-edit',
      component: () => import('../views/ReportFormView.vue'),
      meta: { requiresAuth: true, title: 'Edit Laporan' },
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
      meta: { title: 'About' },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/NotFoundView.vue'),
      meta: { title: 'Page Not Found' },
    },
  ],
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Jika route memerlukan autentikasi, validasi token dengan backend
  if (to.meta.requiresAuth) {
    // Cek apakah token ada di store/localStorage
    if (!authStore.isAuthenticated) {
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }

    // Validasi token dengan backend (verifikasi apakah masih valid)
    // Hanya validasi jika belum divalidasi baru-baru ini (untuk menghindari terlalu banyak API calls)
    try {
      await authStore.fetchProfile()
      
      // Validasi role - hanya izinkan role yang dikenali
      const userRole = authStore.user?.role?.toLowerCase() || ''
      const validRoles = ['superadmin', 'admin', 'manager', 'staff']
      const isRoleValid = validRoles.includes(userRole)
      
      // Jika role tidak dikenali dan bukan di dashboard, redirect ke dashboard
      if (!isRoleValid && to.name !== 'dashboard' && to.name !== 'profile' && to.name !== 'settings') {
        next({ name: 'dashboard' })
        return
      }
      
      // Token valid, izinkan akses
      next()
      return
    } catch (error: unknown) {
      // Handle connection errors (server tidak tersedia)
      const err = error as { message?: string; code?: string }
      const isConnectionError = err.message === 'SERVER_UNAVAILABLE' ||
                                err.code === 'ERR_NETWORK' ||
                                err.code === 'ERR_CONNECTION_REFUSED' ||
                                err.message?.includes('Network Error')
      
      // Token tidak valid atau expired (401/403) atau server tidak tersedia
      // Hapus auth dan redirect ke login
      const isNavigatingToLogin = to.name === 'login'
      
      if (isConnectionError) {
        // Server tidak tersedia - clear state dan redirect ke login
        authStore.clearAuthState()
        if (!isNavigatingToLogin) {
          next({ name: 'login', query: { redirect: to.fullPath } })
        } else {
          next()
        }
        return
      }
      
      // Error lainnya (401/403) - hapus auth dan redirect ke login
      if (!isNavigatingToLogin) {
        console.error('Token validation failed:', error)
      }
      
      // Hapus state lokal tanpa memanggil logout API (untuk menghindari loop)
      // karena cookie mungkin sudah dihapus atau tidak valid
      authStore.clearAuthState()
      
      if (!isNavigatingToLogin) {
        next({ name: 'login', query: { redirect: to.fullPath } })
      } else {
        next()
      }
      return
    }
  }

  // Cek apakah route memerlukan guest (tidak terautentikasi)
  if (to.meta.requiresGuest) {
    // Cek apakah info user ada di localStorage
    const hasUserInStorage = localStorage.getItem('auth_user') !== null
    
    if (hasUserInStorage) {
      // Info user ada - validasi apakah cookie masih valid
      // Tapi jangan blokir akses jika validasi gagal (cookie mungkin sudah expired)
      try {
        // Coba validasi dengan backend (diam-diam)
        await authStore.fetchProfile()
        // Cookie valid, redirect ke dashboard
        next({ name: 'dashboard' })
        return
      } catch {
        // Cookie tidak valid atau hilang atau server tidak tersedia
        // Hapus state lokal secara diam-diam
        // Jangan panggil logout API karena cookie tidak ada atau sudah dihapus
        authStore.clearAuthState()
        
        // Izinkan akses ke route guest (user sudah logout atau server tidak tersedia)
        next()
        return
      }
    }
    
    // Tidak ada info user di storage, izinkan akses ke route guest tanpa API calls
    next()
    return
  }

  next()
})

// Set judul dokumen berdasarkan route
router.afterEach((to) => {
  const appName = 'Pedeve App'
  const pageTitle = (to.meta.title as string) || 'Pedeve App'
  document.title = pageTitle === appName ? appName : `${pageTitle} - ${appName}`
})

// Tangani error navigasi
router.onError((error) => {
  console.error('Router error:', error)
  // Fallback ke dashboard jika ada error
  if (error.message.includes('Failed to fetch dynamically imported module')) {
    router.push('/dashboard').catch(() => {
      window.location.href = '/dashboard'
    })
  }
})

export default router
