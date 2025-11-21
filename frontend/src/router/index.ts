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

  // If route requires authentication, validate token with backend
  if (to.meta.requiresAuth) {
    // Check if token exists in store/localStorage
    if (!authStore.isAuthenticated) {
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }

    // Validate token with backend (verify it's still valid)
    // Only validate if we haven't already validated recently (to avoid too many API calls)
    try {
      await authStore.fetchProfile()
      // Token is valid, allow access
      next()
      return
    } catch (error: any) {
      // Token is invalid or expired (401/403) - clear auth and redirect to login
      console.error('Token validation failed:', error)
      authStore.logout()
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }
  }

  // Check if route requires guest (not authenticated)
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    // Validate token before redirecting away from guest routes
    try {
      await authStore.fetchProfile()
      // Token is valid, redirect to dashboard
      next({ name: 'dashboard' })
      return
    } catch (error: any) {
      // Token is invalid, allow access to guest route
      authStore.logout()
      next()
      return
    }
  }

  next()
})

// Set document title based on route
router.afterEach((to) => {
  const appName = 'Pedeve App'
  const pageTitle = (to.meta.title as string) || 'Pedeve App'
  document.title = pageTitle === appName ? appName : `${pageTitle} - ${appName}`
})

// Handle navigation errors
router.onError((error) => {
  console.error('Router error:', error)
  // Fallback to dashboard if there's an error
  if (error.message.includes('Failed to fetch dynamically imported module')) {
    router.push('/dashboard').catch(() => {
      window.location.href = '/dashboard'
    })
  }
})

export default router
