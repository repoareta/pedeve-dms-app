<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import { userApi, type UserCompanyResponse } from '../api/userManagement'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

// User companies count for badge
const userCompaniesCount = ref(0)
const loadingCompaniesCount = ref(false)

const showUserMenu = ref(false)
const showMobileMenu = ref(false)
const isScrolled = ref(false)

// Valid roles that can access the application
const validRoles = ['superadmin', 'admin', 'manager', 'staff']

// Check if user role is valid
const isRoleValid = computed(() => {
  const userRole = user.value?.role?.toLowerCase() || ''
  return validRoles.includes(userRole)
})

// Menu items - only show for valid roles
const menuItems = computed(() => {
  // If role is not recognized, hide all menus except dashboard (which will show error)
  if (!isRoleValid.value) {
    return []
  }
  
  return [
    { label: 'Dashboard', key: 'dashboard', path: '/dashboard', icon: 'mdi:view-dashboard' },
    { label: 'Anak Perusahaan', key: 'subsidiaries', path: '/subsidiaries', icon: 'mdi:office-building' },
    { label: 'Dokumen', key: 'documents', path: '/documents', icon: 'mdi:file-document' },
    { label: 'Laporan', key: 'reports', path: '/reports', icon: 'mdi:chart-box' },
    { label: 'Manajemen Pengguna', key: 'users', path: '/users', icon: 'mdi:account-group' },
  ]
})

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

// Load user companies count
const loadUserCompaniesCount = async () => {
  if (!authStore.isAuthenticated) return
  
  loadingCompaniesCount.value = true
  try {
    const companies = await userApi.getMyCompanies()
    userCompaniesCount.value = companies.length
  } catch (error) {
    // Silently fail - badge is not critical
    console.warn('Failed to load user companies count:', error)
    userCompaniesCount.value = 0
  } finally {
    loadingCompaniesCount.value = false
  }
}

onMounted(() => {
  loadUserCompaniesCount()
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
  interface WindowWithScrollHandler extends Window {
    __dashboardHeaderScrollHandler?: () => void
  }
  ;(window as WindowWithScrollHandler).__dashboardHeaderScrollHandler = scrollHandler
  
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
  interface WindowWithScrollHandler extends Window {
    __dashboardHeaderScrollHandler?: () => void
  }
  const handler = (window as WindowWithScrollHandler).__dashboardHeaderScrollHandler
  if (handler) {
    window.removeEventListener('scroll', handler)
    document.removeEventListener('scroll', handler)
    if (document.body) {
      document.body.removeEventListener('scroll', handler)
    }
    delete (window as WindowWithScrollHandler).__dashboardHeaderScrollHandler
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
        <!-- Show message if role is not recognized -->
        <div v-if="!isRoleValid" class="role-warning-message">
          <IconifyIcon icon="mdi:alert" width="18" style="margin-right: 8px; color: #faad14;" />
          <span style="color: #faad14;">Role tidak dikenali</span>
        </div>
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
              <a-menu-item key="profile" @click="handleMenuItemClick('/profile')">
                <IconifyIcon icon="mdi:account" width="16" style="margin-right: 8px;" />
                Profil
              </a-menu-item>
              <a-menu-item key="my-company" @click="handleMenuItemClick('/my-company')">
                <IconifyIcon icon="mdi:office-building" width="16" style="margin-right: 8px;" />
                My Company
                <a-badge v-if="userCompaniesCount > 1" :count="userCompaniesCount" :number-style="{ backgroundColor: '#52c41a' }" style="margin-left: 8px;" />
              </a-menu-item>
              <a-menu-item key="settings" @click="handleMenuItemClick('/settings')">
                <IconifyIcon icon="mdi:cog" width="16" style="margin-right: 8px;" />
                Pengaturan
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item key="logout" @click="handleLogout">
                <IconifyIcon icon="mdi:logout" width="16" style="margin-right: 8px;" />
                Keluar
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
            <!-- Show message if role is not recognized -->
            <div v-if="!isRoleValid" class="role-warning-message-mobile" style="padding: 12px; color: #faad14;">
              <IconifyIcon icon="mdi:alert" width="20" style="margin-right: 8px;" />
              <span>Role tidak dikenali</span>
            </div>
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

