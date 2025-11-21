import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import HomeView from '../views/HomeView.vue'

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
      meta: { requiresAuth: true, title: 'Settings' },
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
      // Token valid, izinkan akses
      next()
      return
    } catch (error: any) {
      // Token tidak valid atau expired (401/403) - hapus auth dan redirect ke login
      console.error('Token validation failed:', error)
      authStore.logout()
      next({ name: 'login', query: { redirect: to.fullPath } })
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
      } catch (error: any) {
        // Cookie tidak valid atau hilang - hapus state lokal secara diam-diam
        // Jangan panggil logout API karena cookie tidak ada atau sudah dihapus
        // Manipulasi store langsung untuk menghindari memicu side effects
        const store = useAuthStore()
        store.$patch({
          user: null,
          token: null,
        })
        localStorage.removeItem('auth_user')
        // Izinkan akses ke route guest (user sudah logout)
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
