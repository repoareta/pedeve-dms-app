<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const showUserMenu = ref(false)
const showMobileMenu = ref(false)
const isScrolled = ref(false)

const menuItems = [
  { label: 'Dashboard', key: 'dashboard', path: '/dashboard', icon: 'mdi:view-dashboard' },
  { label: 'Subsidiaries', key: 'subsidiaries', path: '/subsidiaries', icon: 'mdi:office-building' },
  { label: 'Documents', key: 'documents', path: '/documents', icon: 'mdi:file-document' },
  { label: 'Reports', key: 'reports', path: '/reports', icon: 'mdi:chart-box' },
  { label: 'User Management', key: 'users', path: '/users', icon: 'mdi:account-group' },
]

const emit = defineEmits<{
  logout: []
}>()

const handleLogout = () => {
  emit('logout')
}

const handleMenuItemClick = (path: string) => {
  router.push(path)
  showMobileMenu.value = false
}

const toggleMobileMenu = () => {
  showMobileMenu.value = !showMobileMenu.value
}

const updateScrollState = () => {
  const scrollTop = window.scrollY || 
                    window.pageYOffset || 
                    document.documentElement.scrollTop || 
                    document.body.scrollTop || 
                    0
  const newValue = scrollTop > 10
  
  if (isScrolled.value !== newValue) {
    isScrolled.value = newValue
    // console.log('📜 Scroll state changed:', scrollTop, '-> isScrolled:', newValue)
    // console.log('🔍 DOM element classes:', document.querySelector('.header-container')?.className)
  } else {
    // console.log('📊 Current scroll:', scrollTop, 'isScrolled:', newValue)
  }
}

onMounted(() => {
//   console.log('🚀 DashboardHeader mounted')
  
  // Check initial scroll position
  updateScrollState()
  
  // Create scroll handler function
  const scrollHandler = () => {
    // console.log('📜 Scroll event fired!')
    updateScrollState()
  }
  
  // Method 1: window scroll (passive)
  if (window.addEventListener) {
    window.addEventListener('scroll', scrollHandler, { passive: true, capture: false })
    // console.log('✅ Added scroll listener to window (passive)')
  }
  
  // Method 2: window scroll without passive (for testing)
  window.addEventListener('scroll', () => {
    // console.log('🔄 Direct scroll handler called')
    updateScrollState()
  })
  
  // Method 3: document scroll
  if (document.addEventListener) {
    document.addEventListener('scroll', scrollHandler, { passive: true, capture: true })
    // console.log('✅ Added scroll listener to document (passive)')
  }
  
  // Method 4: document.body scroll (for some cases)
  if (document.body) {
    document.body.addEventListener('scroll', scrollHandler, { passive: true })
    // console.log('✅ Added scroll listener to document.body')
  }
  
  // Store handler reference for cleanup
  ;(window as any).__dashboardHeaderScrollHandler = scrollHandler
  
  // Test scroll detection with polling (temporary for debugging)
  let pollCount = 0
  const pollInterval = setInterval(() => {
    pollCount++
    const currentScroll = window.pageYOffset || document.documentElement.scrollTop || 0
    if (currentScroll > 0 || pollCount > 10) {
    //   console.log(`⏱️ Poll ${pollCount}: scroll = ${currentScroll}`)
      updateScrollState()
      if (pollCount > 50) {
        clearInterval(pollInterval)
        // console.log('⏹️ Stopped polling after 50 checks')
      }
    }
  }, 200)
  
  // Cleanup polling after 10 seconds
  setTimeout(() => {
    clearInterval(pollInterval)
  }, 10000)
})

onUnmounted(() => {
//   console.log('🛑 DashboardHeader unmounting, removing listeners')
  const handler = (window as any).__dashboardHeaderScrollHandler
  if (handler) {
    window.removeEventListener('scroll', handler)
    document.removeEventListener('scroll', handler)
    if (document.body) {
      document.body.removeEventListener('scroll', handler)
    }
    delete (window as any).__dashboardHeaderScrollHandler
  }
})
</script>

<template>
  <div class="dashboard-header">
    <div class="header-container" :class="{ 'onscrollnav': isScrolled }">
      <div class="header-left">
        <img src="/logo.png" alt="Pertamina Logo" class="logo" />
        <button class="mobile-menu-toggle" @click="toggleMobileMenu" type="button">
          <IconifyIcon icon="mdi:menu" width="24" height="24" />
        </button>
      </div>

      <div class="header-center">
        <a-menu 
          mode="horizontal" 
          :selected-keys="[route.name as string]"
          class="nav-menu desktop-menu"
        >
          <a-menu-item 
            v-for="item in menuItems" 
            :key="item.key"
            @click="handleMenuItemClick(item.path)"
          >
            <IconifyIcon :icon="item.icon" width="18" style="margin-right: 8px;" />
            {{ item.label }}
          </a-menu-item>
        </a-menu>
      </div>

      <div class="header-right">
        <a-button type="text" class="icon-btn desktop-icon">
          <IconifyIcon icon="mdi:bell-outline" width="20" height="20" />
        </a-button>

        <a-dropdown v-model:open="showUserMenu" placement="bottomRight">
          <div class="user-profile">
            <div class="user-avatar">
              {{ user?.username?.charAt(0).toUpperCase() || 'U' }}
            </div>
            <span class="user-name desktop-username">{{ user?.username || 'User' }}</span>
            <IconifyIcon icon="mdi:chevron-down" width="16" class="desktop-icon" />
          </div>
          <template #overlay>
            <a-menu>
              <a-menu-item key="profile">
                <IconifyIcon icon="mdi:account" width="16" style="margin-right: 8px;" />
                Profile
              </a-menu-item>
              <a-menu-item key="settings">
                <IconifyIcon icon="mdi:cog" width="16" style="margin-right: 8px;" />
                Settings
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item key="logout" @click="handleLogout">
                <IconifyIcon icon="mdi:logout" width="16" style="margin-right: 8px;" />
                Logout
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </div>

    <!-- Mobile Menu -->
    <transition name="slide-down">
      <div v-if="showMobileMenu" class="mobile-menu">
        <div class="mobile-menu-header">
          <div class="mobile-user-info">
            <div class="user-avatar">
              {{ user?.username?.charAt(0).toUpperCase() || 'U' }}
            </div>
            <div class="mobile-user-details">
              <div class="user-name">{{ user?.username || 'User' }}</div>
              <div class="user-email">{{ user?.email || '' }}</div>
            </div>
          </div>
        </div>
        <div class="mobile-menu-items">
          <a-menu mode="vertical" :selected-keys="[route.name as string]">
            <a-menu-item 
              v-for="item in menuItems" 
              :key="item.key"
              @click="handleMenuItemClick(item.path)"
            >
              <IconifyIcon :icon="item.icon" width="20" style="margin-right: 12px;" />
              {{ item.label }}
            </a-menu-item>
          </a-menu>
          <div class="mobile-menu-footer">
            <a-button type="text" block @click="handleLogout" class="mobile-logout-btn">
              <IconifyIcon icon="mdi:logout" width="18" style="margin-right: 8px;" />
              Logout
            </a-button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

