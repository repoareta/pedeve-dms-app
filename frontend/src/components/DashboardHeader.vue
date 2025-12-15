<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import { userApi } from '../api/userManagement'
import { notificationApi, type Notification } from '../api/notifications'
import { notification } from 'ant-design-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

// User companies count for badge
const userCompaniesCount = ref(0)
const loadingCompaniesCount = ref(false)

const showUserMenu = ref(false)
const showNotificationMenu = ref(false)
const showMobileMenu = ref(false)
const isScrolled = ref(false)
const isMaximized = ref(false)

// Notifications
const notifications = ref<Notification[]>([])
const unreadCount = ref(0)
const loadingNotifications = ref(false)
const notificationPollingInterval = ref<ReturnType<typeof setInterval> | null>(null)
const shownNotificationIds = ref<Set<string>>(new Set()) // Track notifikasi yang sudah ditampilkan
const isFirstLoad = ref(true) // Flag untuk menandai load pertama
const hasShownInitialNotifications = ref(false) // Flag untuk track apakah sudah menampilkan notifikasi saat login

// Stack configuration untuk notification
// Ant Design Vue automatically stacks notifications when there are multiple
// Stack behavior is built-in and doesn't require explicit configuration

// Valid roles that can access the application
const validRoles = ['superadmin', 'administrator', 'admin', 'manager', 'staff']

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
    // { label: 'Dashboard', key: 'dashboard', path: '/dashboard', icon: 'mdi:view-dashboard' },
    { label: 'Daftar Perusahaan', key: 'subsidiaries', path: '/subsidiaries', icon: 'mdi:office-building' },
    { label: 'Documents', key: 'documents', path: '/documents', icon: 'mdi:file-document' },
    { label: 'Laporan', key: 'reports', path: '/reports', icon: 'mdi:chart-box' },
    { label: 'Manajemen Pengguna', key: 'users', path: '/users', icon: 'mdi:account-group' },
  ]
})

const emit = defineEmits<{
  logout: []
  toggleMaximize: [value: boolean]
}>()

const handleLogout = () => {
  emit('logout')
}

const handleToggleMaximize = () => {
  // Check if browser supports fullscreen API
  if (document.fullscreenElement) {
    // Exit fullscreen
    document.exitFullscreen().then(() => {
      isMaximized.value = false
      emit('toggleMaximize', false)
    }).catch(() => {
      // Fallback: try to minimize window (if in Electron or similar)
      interface WindowWithElectron extends Window {
        electron?: {
          minimize?: () => void
          maximize?: () => void
        }
      }
      const win = window as WindowWithElectron
      if (win.electron?.minimize) {
        win.electron.minimize()
      }
    })
  } else {
    // Enter fullscreen
    const element = document.documentElement
    if (element.requestFullscreen) {
      element.requestFullscreen().then(() => {
        isMaximized.value = true
        emit('toggleMaximize', true)
      }).catch(() => {
        // Fallback: try to maximize window (if in Electron or similar)
        interface WindowWithElectron extends Window {
          electron?: {
            minimize?: () => void
            maximize?: () => void
          }
        }
        const win = window as WindowWithElectron
        if (win.electron?.maximize) {
          win.electron.maximize()
          isMaximized.value = true
          emit('toggleMaximize', true)
        }
      })
    }
  }
}

// Listen for fullscreen changes
onMounted(() => {
  const handleFullscreenChange = () => {
    isMaximized.value = !!document.fullscreenElement
  }
  
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('mozfullscreenchange', handleFullscreenChange)
  document.addEventListener('MSFullscreenChange', handleFullscreenChange)
  
  // Store handler for cleanup
  interface WindowWithFullscreenHandler extends Window {
    __fullscreenHandler?: () => void
  }
  ;(window as WindowWithFullscreenHandler).__fullscreenHandler = handleFullscreenChange
})

const handleMenuItemClick = (path: string) => {
  router.push(path)
  showMobileMenu.value = false
}

const handleMenuClick = (e: { key: string }) => {
  // Handle menu item clicks safely
  switch (e.key) {
    case 'profile':
      handleMenuItemClick('/profile')
      showUserMenu.value = false
      break
    case 'my-company':
      handleMenuItemClick('/my-company')
      showUserMenu.value = false
      break
    case 'settings':
      handleMenuItemClick('/settings')
      showUserMenu.value = false
      break
    case 'logout':
      handleLogout()
      showUserMenu.value = false
      break
  }
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

// Open notification box
const openNotificationBox = (notif: Notification) => {
  // Tentukan type berdasarkan notif.type
  let type: 'info' | 'success' | 'warning' | 'error' = 'info'
  switch (notif.type) {
    case 'success':
      type = 'success'
      break
    case 'warning':
      type = 'warning'
      break
    case 'error':
      type = 'error'
      break
    default:
      type = 'info'
  }
  
  try {
    // Tampilkan notification dengan Ant Design Vue
    // Note: showProgress dan pauseOnHover tidak didukung di Ant Design Vue
    notification[type]({
      message: notif.title,
      description: notif.message,
      duration: 4.5, // Auto hide setelah 4.5 detik
      placement: 'topRight',
      onClick: () => {
        handleNotificationClick(notif)
      },
    })
  } catch (error) {
    console.error('❌ [PushNotification] Failed to show notification:', error)
  }
}

// Show push notification
const showPushNotification = (notif: Notification) => {
  try {
    // Gunakan openNotificationBox untuk menampilkan notification
    openNotificationBox(notif)
  } catch (error) {
    console.error('❌ [PushNotification] Failed to show notification:', error)
  }
}

// Load notifications
// PENTING: RBAC sudah di-handle di backend melalui GetNotificationsWithRBAC
// - Superadmin/Administrator: melihat semua notifikasi
// - Admin: melihat notifikasi dari company mereka + descendants
// - Regular users: hanya melihat notifikasi mereka sendiri
// Frontend tidak perlu melakukan filtering tambahan, cukup menggunakan endpoint yang sudah ada
const loadNotifications = async () => {
  if (!authStore.isAuthenticated) {
    // Stop polling jika user tidak authenticated
    stopNotificationPolling()
    return
  }
  
  // Jangan restart polling di sini - biarkan hanya di startNotificationPolling
  // untuk menghindari infinite loop
  
  loadingNotifications.value = true
  try {
    // Endpoint ini sudah menggunakan RBAC di backend (GetNotificationsWithRBAC)
    // Tidak perlu filtering tambahan di frontend
    const [notifs, count] = await Promise.all([
      notificationApi.getNotifications(false, 5), // Ambil 5 notifikasi terakhir (read + unread) - sudah filtered by RBAC
      notificationApi.getUnreadCount(), // Unread count - sudah filtered by RBAC
    ])
    
    // Hanya tampilkan notifikasi yang belum dibaca di dropdown
    const unreadNotifs = notifs.filter(n => !n.is_read)
    notifications.value = unreadNotifs.slice(0, 5) // Maksimal 5 notifikasi unread
    
    unreadCount.value = count
    
    // PENTING: Saat first load setelah login (session baru), tampilkan SEMUA notifikasi unread sebagai push notification
    if (isFirstLoad.value && !hasShownInitialNotifications.value) {
      // Tampilkan semua notifikasi unread sebagai push notification (PENTING untuk reminder expired document)
      if (unreadNotifs.length > 0) {
        // Tampilkan maksimal 5 notifikasi unread (untuk menghindari spam berlebihan)
        const notificationsToShow = unreadNotifs.slice(0, 5)
        notificationsToShow.forEach((notif, index) => {
          // Jangan tambahkan ke shownNotificationIds - biarkan muncul berulang sampai ditindak lanjuti
          
          // Tampilkan dengan delay berurutan (setiap 1000ms untuk visibility yang baik)
          setTimeout(() => {
            showPushNotification(notif)
          }, index * 1000) // 1 detik delay
        })
        
        // Tandai bahwa sudah menampilkan notifikasi awal
        hasShownInitialNotifications.value = true
      }
      
      // Reset isFirstLoad setelah beberapa detik
      setTimeout(() => {
        isFirstLoad.value = false
      }, 3000)
      return
    }
    
    // PENTING: Tampilkan push notification untuk notifikasi yang BELUM ditindak lanjuti (is_read = false)
    // Push notification akan muncul berulang-ulang sampai user klik "Sudah ditindak lanjuti"
    // Bahkan setelah expired date lewat, push notification tetap muncul sampai ditindak lanjuti
    const unresolvedNotifications = notifs.filter(notif => !notif.is_read)
    
    // Tampilkan push notification untuk notifikasi yang belum ditindak lanjuti
    // Jangan skip notifikasi yang sudah pernah ditampilkan - tampilkan lagi jika masih belum ditindak lanjuti
    if (unresolvedNotifications.length > 0) {
      const notificationsToShow = unresolvedNotifications.slice(0, 3) // Maksimal 3 notifikasi
      notificationsToShow.forEach((notif, index) => {
        // Jangan tambahkan ke shownNotificationIds - biarkan muncul berulang sampai ditindak lanjuti
        
        // Tampilkan dengan delay berurutan (setiap 800ms untuk balance antara visibility dan performance)
        setTimeout(() => {
          showPushNotification(notif)
        }, index * 800) // 800ms delay
      })
    }
  } catch (error) {
    console.error('❌ [Notifications] Failed to load notifications:', error)
  } finally {
    loadingNotifications.value = false
  }
}

// Start polling for notifications
const startNotificationPolling = () => {
  // Prevent multiple polling instances
  if (notificationPollingInterval.value) {
    return // Already running
  }
  
  // Load immediately - akan menampilkan semua unread notifications jika first load
  loadNotifications()
  
  // Poll every 30 seconds (reduced frequency untuk mengurangi load)
  notificationPollingInterval.value = setInterval(() => {
    loadNotifications()
  }, 30000) // 30 detik
}

// Stop polling
const stopNotificationPolling = () => {
  if (notificationPollingInterval.value) {
    clearInterval(notificationPollingInterval.value)
    notificationPollingInterval.value = null
  }
}

// Handle notification click
// PENTING: Hanya navigate, TIDAK mark as read
// Notifikasi hanya selesai setelah user klik button "Sudah ditindak lanjuti" di halaman notifikasi
const handleNotificationClick = async (notification: Notification) => {
  // Navigate to resource if available (TIDAK mark as read)
  if (notification.resource_type === 'document' && notification.resource_id) {
    router.push(`/documents/${notification.resource_id}`)
    showNotificationMenu.value = false
  } else {
    // Navigate to notifications inbox
    router.push('/notifications')
    showNotificationMenu.value = false
  }
}

// Format time
const formatTime = (date: string) => {
  return dayjs(date).fromNow()
}

// Watch untuk detect login (session baru)
// Saat user login, reset state untuk menampilkan semua unread notifications
let previousAuthState = authStore.isAuthenticated
watch(() => authStore.isAuthenticated, (isAuthenticated) => {
  // Jika user baru login (berubah dari false ke true)
  if (!previousAuthState && isAuthenticated) {
    // Reset state untuk menampilkan semua unread notifications saat login
    isFirstLoad.value = true
    hasShownInitialNotifications.value = false
    shownNotificationIds.value.clear()
    
    // Restart polling jika belum berjalan
    if (!notificationPollingInterval.value) {
      startNotificationPolling()
    } else {
      // Jika polling sudah berjalan, load notifications untuk menampilkan push notification
      loadNotifications()
    }
  }
  // Update previous state
  previousAuthState = isAuthenticated
})

onMounted(() => {
  // Setup notification configuration
  // Note: Stack configuration is handled automatically by Ant Design Vue
  // Multiple notifications will be stacked automatically
  notification.config({
    placement: 'topRight',
    top: 24,
    bottom: 24,
    duration: 4.5,
    rtl: false,
  })
  
  loadUserCompaniesCount()
  startNotificationPolling()
//   console.log('🚀 DashboardHeader mounted')
  
  // Listen untuk refresh notifications setelah navigate
  const handleNotificationRead = () => {
    // Refresh notifications setelah beberapa detik (untuk memberi waktu backend update)
    setTimeout(() => {
      loadNotifications()
    }, 1000)
  }
  
  // Store handler reference untuk cleanup
  interface WindowWithNotificationHandler extends Window {
    __notificationReadHandler?: EventListener
  }
  ;(window as WindowWithNotificationHandler).__notificationReadHandler = handleNotificationRead as EventListener
  window.addEventListener('notification-read', handleNotificationRead as EventListener)
  
  // Check initial scroll position
  updateScrollState()
  
  // Create scroll handler function
  const scrollHandler = () => {
    // console.log('📜 Scroll event fired!')
    updateScrollState()
  }
  
  // Hanya gunakan 1 scroll listener untuk menghindari ribuan event fires
  if (window.addEventListener) {
    window.addEventListener('scroll', scrollHandler, { passive: true })
  }
  
  // Store handler reference for cleanup
  interface WindowWithScrollHandler extends Window {
    __dashboardHeaderScrollHandler?: () => void
  }
  ;(window as WindowWithScrollHandler).__dashboardHeaderScrollHandler = scrollHandler
  
  // Scroll detection sudah menggunakan event listeners, tidak perlu polling
})

onUnmounted(() => {
//   console.log('🛑 DashboardHeader unmounting, removing listeners')
  interface WindowWithScrollHandler extends Window {
    __dashboardHeaderScrollHandler?: () => void
    __fullscreenHandler?: () => void
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
  
  // Remove fullscreen listeners
  const fullscreenHandler = (window as WindowWithScrollHandler).__fullscreenHandler
  if (fullscreenHandler) {
    document.removeEventListener('fullscreenchange', fullscreenHandler)
    document.removeEventListener('webkitfullscreenchange', fullscreenHandler)
    document.removeEventListener('mozfullscreenchange', fullscreenHandler)
    document.removeEventListener('MSFullscreenChange', fullscreenHandler)
    delete (window as WindowWithScrollHandler).__fullscreenHandler
  }
  
  // Stop notification polling
  stopNotificationPolling()
  
  // Clear shown notification IDs saat unmount
  shownNotificationIds.value.clear()
  
  // Remove notification-read event listener
  interface WindowWithNotificationHandler extends Window {
    __notificationReadHandler?: EventListener
  }
  const notificationHandler = (window as WindowWithNotificationHandler).__notificationReadHandler
  if (notificationHandler) {
    window.removeEventListener('notification-read', notificationHandler)
    delete (window as WindowWithNotificationHandler).__notificationReadHandler
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
        <nav class="custom-nav-menu">
          <button
            v-for="item in menuItems" 
            :key="item.key"
            @click="handleMenuItemClick(item.path)"
            :class="['nav-item', { 'nav-item-active': route.name === item.key }]"
          >
            <IconifyIcon :icon="item.icon" width="18" style="margin-right: 8px;" />
            {{ item.label }}
          </button>
        </nav>
        <!-- Show message if role is not recognized -->
        <div v-if="!isRoleValid" class="role-warning-message">
          <IconifyIcon icon="mdi:alert" width="18" style="margin-right: 8px; color: #faad14;" />
          <span style="color: #faad14;">Role tidak dikenali</span>
        </div>
      </div>

      <div class="header-right">
        <a-button 
          type="text" 
          class="icon-btn desktop-icon"
          @click="handleToggleMaximize"
          :title="isMaximized ? 'Exit Fullscreen' : 'Fullscreen'"
        >
          <IconifyIcon 
            :icon="isMaximized ? 'ant-design:fullscreen-exit-outlined' : 'ant-design:fullscreen-outlined'" 
            width="20" 
            height="20" 
          />
        </a-button>

        <a-dropdown 
          v-model:open="showNotificationMenu" 
          placement="bottomRight"
          :z-index="1001"
          :trigger="['click']"
        >
          <a-badge :count="unreadCount" :offset="[10, 0]">
            <a-button type="text" class="icon-btn desktop-icon">
              <IconifyIcon icon="mdi:bell-outline" width="20" height="20" />
            </a-button>
          </a-badge>
          <template #overlay>
            <div class="notification-dropdown">
              <div class="notification-header">
                <span class="notification-title">Notifikasi</span>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="router.push('/notifications'); showNotificationMenu = false"
                >
                  Lihat Semua
                </a-button>
              </div>
              <a-spin :spinning="loadingNotifications">
                <div class="notification-list">
                  <div 
                    v-if="notifications.length === 0" 
                    class="notification-empty"
                  >
                    Tidak ada notifikasi baru
                  </div>
                  <div
                    v-for="notif in notifications"
                    :key="notif.id"
                    class="notification-item"
                    :class="{ 'unread': !notif.is_read }"
                    @click="handleNotificationClick(notif)"
                  >
                    <div class="notification-content">
                      <div class="notification-title-text">{{ notif.title }}</div>
                      <div class="notification-message">{{ notif.message }}</div>
                      <div class="notification-time">{{ formatTime(notif.created_at) }}</div>
                    </div>
                  </div>
                </div>
              </a-spin>
            </div>
          </template>
        </a-dropdown>

        <a-dropdown 
          v-model:open="showUserMenu" 
          placement="bottomRight"
          :z-index="1002"
        >
          <div class="user-profile">
            <div class="user-avatar">
              {{ user?.username?.charAt(0).toUpperCase() || 'U' }}
            </div>
            <span class="user-name desktop-username">{{ user?.username || 'User' }}</span>
            <IconifyIcon icon="mdi:chevron-down" width="16" class="desktop-icon" />
          </div>
          <template #overlay>
            <a-menu style="z-index: 1002;" @click="handleMenuClick">
              <a-menu-item key="profile">
                <IconifyIcon icon="mdi:account" width="16" style="margin-right: 8px;" />
                Profil
              </a-menu-item>
              <a-menu-item key="my-company">
                <IconifyIcon icon="mdi:office-building" width="16" style="margin-right: 8px;" />
                My Company
                <a-badge v-if="userCompaniesCount > 1" :count="userCompaniesCount" :number-style="{ backgroundColor: '#52c41a' }" style="margin-left: 8px;" />
              </a-menu-item>
              <a-menu-item key="settings">
                <IconifyIcon icon="mdi:cog" width="16" style="margin-right: 8px;" />
                Pengaturan
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item key="logout">
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

<style lang="scss" scoped>
.notification-dropdown {
  width: 360px;
  max-height: 480px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  
  .notification-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #f0f0f0;
    
    .notification-title {
      font-weight: 600;
      font-size: 16px;
      color: #333;
    }
  }
  
  .notification-list {
    max-height: 400px;
    overflow-y: auto;
    
    .notification-empty {
      padding: 40px 20px;
      text-align: center;
      color: #8c8c8c;
    }
    
    .notification-item {
      padding: 12px 16px;
      border-bottom: 1px solid #f0f0f0;
      cursor: pointer;
      transition: background-color 0.2s;
      
      &:hover {
        background-color: #f5f5f5;
      }
      
      &.unread {
        background-color: #e6f7ff;
        border-left: 3px solid #1890ff;
      }
      
      .notification-content {
        .notification-title-text {
          font-weight: 500;
          color: #333;
          margin-bottom: 4px;
          font-size: 14px;
        }
        
        .notification-message {
          color: #666;
          font-size: 13px;
          margin-bottom: 4px;
          line-height: 1.4;
        }
        
        .notification-time {
          color: #8c8c8c;
          font-size: 12px;
        }
      }
    }
  }
}
</style>
