<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import { userApi } from '../api/userManagement'
import { notificationApi, notificationSettingsApi, type Notification } from '../api/notifications'
import { notification } from 'ant-design-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { logger } from '../utils/logger'

dayjs.extend(relativeTime)

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

// Jumlah companies user untuk badge
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
const inAppNotificationsEnabled = ref(true) // Default: enabled, akan di-load dari settings
const expiryThresholdDays = ref<number>(14) // Default: 14 hari, akan di-load dari settings

// CATATAN: Ant Design Vue 4.2.6 tidak mendukung useNotification
// Menggunakan notification API default tanpa config tambahan

// Valid roles that can access the application
const validRoles = ['superadmin', 'administrator', 'admin', 'manager', 'staff']

// Cek apakah role user valid
const isRoleValid = computed(() => {
  const userRole = user.value?.role?.toLowerCase() || ''
  return validRoles.includes(userRole)
})

// Menu items - only show for valid roles
const menuItems = computed(() => {
  // Kalau role tidak dikenali, sembunyikan semua menu kecuali dashboard (yang akan tampilkan error)
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
  // Cek apakah browser support fullscreen API
  if (document.fullscreenElement) {
    // Keluar dari fullscreen
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
    // Masuk ke fullscreen
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
  
  // Simpan handler untuk cleanup
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
    // Fail secara silent - badge tidak kritis
    logger.warn('Failed to load user companies count:', error)
    userCompaniesCount.value = 0
  } finally {
    loadingCompaniesCount.value = false
  }
}

// Open notification box
const openNotificationBox = (notif: Notification) => {
  // Tentukan type berdasarkan notif.type
  // PENTING: Semua notifikasi document_expiry (baik sudah expired maupun akan expired) 
  // harus ditampilkan sebagai warning/error untuk menarik perhatian
  let type: 'info' | 'success' | 'warning' | 'error' = 'info'
  if (notif.type === 'document_expiry') {
    // Gunakan 'warning' untuk semua notifikasi document expiry (baik sudah expired maupun akan expired)
    // Ini akan membuat semua notifikasi document expiry ditampilkan sebagai push notification
    type = 'warning'
  } else {
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
  }
  
  try {
    // Pastikan notification API tersedia
    if (!notification) {
      return
    }
    
    // Pastikan method untuk type tersedia
    const notificationMethod = notification[type]
    if (typeof notificationMethod !== 'function') {
      return
    }
    
    // Deteksi apakah notification untuk expired document
    // PENTING: Hanya dokumen yang SUDAH expired yang mendapat styling merah
    // Dokumen yang AKAN expired menggunakan styling default (putih)
    const isExpiredDocument = notif.type === 'document_expiry'
    const isAlreadyExpired = isExpiredDocument && notif.title.includes('Sudah Expired')
    
    // Siapkan notification config
    const notificationConfig: {
      message: string
      description: string
      duration: number
      placement: 'topRight'
      onClick: () => void
      className?: string
      style?: {
        backgroundColor: string
        border: string
        borderRadius: string
      }
    } = {
      message: notif.title,
      description: formatDynamicMessage(notif),
      duration: 2.5, // Auto hide setelah 2.5 detik
      placement: 'topRight',
      onClick: () => {
        handleNotificationClick(notif)
      },
    }
    
    // Custom styling HANYA untuk dokumen yang SUDAH expired (warna merah)
    // Dokumen yang AKAN expired menggunakan styling default (putih)
    if (isAlreadyExpired) {
      // Gunakan className dan style untuk custom styling
      notificationConfig.className = 'expired-document-notification'
      // Tambahkan inline style sebagai fallback
      notificationConfig.style = {
        backgroundColor: '#bf4e4e',
        border: '2px dashed #ccc',
        borderRadius: '8px',
      }
    }
    
    // Tampilkan notification dengan Ant Design Vue
    notificationMethod(notificationConfig)
    
    // Jika dokumen sudah expired, tambahkan class dan style setelah notification di-render
    if (isAlreadyExpired) {
      // Pakai multiple setTimeout untuk pastikan element sudah di-render
      setTimeout(() => {
        const notices = document.querySelectorAll('.ant-notification-topRight .ant-notification-notice')
        if (notices.length > 0) {
          // Ambil notice terakhir (yang baru saja dibuat)
          const lastNotice = notices[notices.length - 1] as HTMLElement
          if (lastNotice) {
            lastNotice.classList.add('expired-document-notification')
            // Tambahkan inline style langsung sebagai fallback
            lastNotice.style.backgroundColor = '#bf4e4e'
            lastNotice.style.border = '2px dashed #ccc'
            lastNotice.style.borderRadius = '8px'
            lastNotice.style.backdropFilter = 'blur(10px)'
            lastNotice.style.setProperty('-webkit-backdrop-filter', 'blur(10px)')
            lastNotice.style.color = '#fff'
            
            // Pastikan semua text element di dalam notification berwarna putih
            const textElements = lastNotice.querySelectorAll('span, div, p, strong, em, .ant-notification-notice-message, .ant-notification-notice-description')
            textElements.forEach((textEl) => {
              (textEl as HTMLElement).style.color = '#fff'
            })
          }
        }
      }, 100)
      
      // Double check setelah 200ms dan pastikan semua text berwarna putih
      setTimeout(() => {
        const notices = document.querySelectorAll('.ant-notification-topRight .ant-notification-notice')
        notices.forEach((notice) => {
          if (notice.classList.contains('expired-document-notification')) {
            const el = notice as HTMLElement
            el.style.backgroundColor = '#bf4e4e'
            el.style.border = '2px dashed #ccc'
            el.style.borderRadius = '8px'
            el.style.color = '#fff'
            
            // Pastikan semua text element di dalam notification berwarna putih
            const textElements = el.querySelectorAll('span, div, p, strong, em, .ant-notification-notice-message, .ant-notification-notice-description')
            textElements.forEach((textEl) => {
              (textEl as HTMLElement).style.color = '#fff'
            })
          }
        })
      }, 200)
    }
  } catch {
    // Fallback: coba tampilkan dengan method langsung
    try {
      if (notification && typeof notification[type] === 'function') {
        notification[type]({
          message: notif.title,
          description: formatDynamicMessage(notif),
          duration: 1.5,
        })
      }
    } catch {
      // Silent fail - notification mungkin tidak bisa ditampilkan
    }
  }
}

// Show push notification
const showPushNotification = (notif: Notification) => {
  try {
    // Validasi: jangan tampilkan jika notification sudah dibaca
    if (notif.is_read) {
      return
    }
    
    // Gunakan openNotificationBox untuk menampilkan notification
    openNotificationBox(notif)
  } catch {
    // Silent fail - notification mungkin tidak bisa ditampilkan
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
    // PENTING: Ambil hanya notifikasi yang BELUM DIBACA (unread_only: true) untuk push notification
    // Push notification hanya ditampilkan untuk notifikasi yang belum ditindak lanjuti
    const [notifsInbox, count] = await Promise.all([
      notificationApi.getNotificationsInbox({
        unread_only: true, // PENTING: Hanya ambil notifikasi yang belum dibaca untuk push notification
        page: 1,
        page_size: 50, // Ambil 50 notifikasi untuk memastikan semua document_expiry masuk
      }),
      notificationApi.getUnreadCount(), // Unread count - sudah filtered by RBAC
    ])
    
    // Extract notifications dari inbox response
    const notifs = notifsInbox.data || []
    
    
    // Notifikasi yang diterima dari API sudah filtered unread_only: true, jadi semua seharusnya unread
    // Tapi tetap filter untuk safety
    const unreadNotifs = notifs.filter(n => !n.is_read)
    notifications.value = unreadNotifs.slice(0, 5) // Maksimal 5 notifikasi unread
    
    // PENTING: Pastikan unreadCount sesuai dengan jumlah notifikasi yang belum ditindak lanjuti
    // Gunakan count dari API (sudah filtered by RBAC) sebagai sumber kebenaran
    unreadCount.value = count
    
    // PENTING: Tampilkan push notification hanya jika in-app notifications enabled
    // Icon, dropdown, dan halaman notifikasi tetap berjalan normal meskipun in-app disabled
    if (inAppNotificationsEnabled.value) {
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
      // Notifikasi yang diterima dari API sudah filtered unread_only: true, jadi semua seharusnya unread
      // Tapi tetap filter untuk safety
      const unresolvedNotifications = notifs.filter(notif => {
        // Filter hanya notifikasi yang belum dibaca
        if (notif.is_read) {
          return false
        }
        
        // PENTING: Untuk notifikasi document_expiry, pastikan semua (baik sudah expired maupun akan expired) ditampilkan
        // Backend sudah membuat notifikasi berdasarkan threshold, jadi kita hanya perlu memastikan semua document_expiry ditampilkan
        if (notif.type === 'document_expiry') {
          // Tampilkan semua notifikasi document_expiry (baik "Sudah Expired" maupun "Akan Expired")
          // Backend sudah filter berdasarkan threshold, jadi kita hanya perlu memastikan semua ditampilkan
          return true
        }
        
        // Untuk notifikasi type lain, tampilkan juga
        return true
      })
      
      
      // Tampilkan push notification untuk notifikasi yang belum ditindak lanjuti
      // Jangan skip notifikasi yang sudah pernah ditampilkan - tampilkan lagi jika masih belum ditindak lanjuti
      if (unresolvedNotifications.length > 0) {
        // Prioritaskan notifikasi document_expiry (baik sudah expired maupun akan expired)
        const documentExpiryNotifs = unresolvedNotifications.filter(n => n.type === 'document_expiry')
        const otherNotifs = unresolvedNotifications.filter(n => n.type !== 'document_expiry')
        
        
        // PENTING: Tampilkan SEMUA notifikasi document_expiry (baik sudah expired maupun akan expired)
        // Jangan batasi hanya 5, karena kita perlu menampilkan semua document_expiry sebagai push notification
        const allDocumentExpiryToShow = documentExpiryNotifs // Tampilkan semua, tidak dibatasi
        const otherNotifsToShow = otherNotifs.slice(0, Math.max(0, 10 - allDocumentExpiryToShow.length))
        
        // Gabungkan: semua document_expiry dulu, baru yang lain
        const notificationsToShow = [
          ...allDocumentExpiryToShow,
          ...otherNotifsToShow
        ].slice(0, 10) // Maksimal 10 notifikasi untuk memastikan semua document_expiry masuk
        
        notificationsToShow.forEach((notif, index) => {
          // Jangan tambahkan ke shownNotificationIds - biarkan muncul berulang sampai ditindak lanjuti
          
          // Tampilkan dengan delay berurutan (setiap 800ms untuk balance antara visibility dan performance)
          setTimeout(() => {
            showPushNotification(notif)
          }, index * 800) // 800ms delay
        })
      }
    } else {
      // Jika in-app notifications disabled, skip push notifications tapi tetap update icon dan dropdown
      // Reset isFirstLoad flag meskipun tidak menampilkan push notification
      if (isFirstLoad.value) {
        hasShownInitialNotifications.value = true
        setTimeout(() => {
          isFirstLoad.value = false
        }, 1000)
      }
    }
  } catch (error) {
    logger.error('❌ [Notifications] Failed to load notifications:', error)
  } finally {
    loadingNotifications.value = false
  }
}

// Load notification settings untuk cek in-app notifications enabled dan threshold
const loadNotificationSettings = async (): Promise<void> => {
  try {
    const settings = await notificationSettingsApi.getSettings()
    inAppNotificationsEnabled.value = settings.in_app_enabled
    expiryThresholdDays.value = settings.expiry_threshold_days || 14 // Default 14 hari jika tidak ada
  } catch (error) {
    logger.error('❌ [Notifications] Failed to load notification settings:', error)
    // Default to enabled jika gagal load
    inAppNotificationsEnabled.value = true
    expiryThresholdDays.value = 14 // Default 14 hari
  }
}

// Start polling for notifications
const startNotificationPolling = () => {
  // Prevent multiple polling instances
  if (notificationPollingInterval.value) {
    return // Already running
  }
  
  // Load notification settings first (async, tidak blocking)
  loadNotificationSettings().then(() => {
    // Load notifications setelah settings loaded
    loadNotifications()
  }).catch(() => {
    // Jika gagal load settings, tetap load notifications dengan default (enabled)
    loadNotifications()
  })
  
  // Poll every 2 minutes untuk push notification berulang
  notificationPollingInterval.value = setInterval(() => {
    // Reload settings setiap polling untuk memastikan perubahan settings langsung terdeteksi
    loadNotificationSettings()
    loadNotifications()
  }, 120000) // 2 menit (120000 ms)
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

// Format dynamic message berdasarkan waktu real-time untuk document expiry notifications
const formatDynamicMessage = (notif: Notification): string => {
  // Hanya untuk document_expiry notification yang memiliki document dengan expiry_date
  if (notif.type === 'document_expiry' && notif.document?.expiry_date) {
    const expiryDate = dayjs(notif.document.expiry_date)
    const now = dayjs()
    const diffDays = expiryDate.diff(now, 'day')
    
    // Extract document name dari original message atau dari document.name
    const docName = notif.document.name || notif.title.replace("Dokumen '", '').replace("' Akan Expired", '')
    
    if (diffDays < 0) {
      // Sudah expired
      const daysAgo = Math.abs(diffDays)
      if (daysAgo === 0) {
        return `Dokumen '${docName}' sudah expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.`
      } else if (daysAgo === 1) {
        return `Dokumen '${docName}' sudah expired 1 hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.`
      } else if (daysAgo < 7) {
        return `Dokumen '${docName}' sudah expired ${daysAgo} hari yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.`
      } else if (daysAgo < 30) {
        const weeksAgo = Math.floor(daysAgo / 7)
        return `Dokumen '${docName}' sudah expired ${weeksAgo} minggu yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.`
      } else {
        const monthsAgo = Math.floor(daysAgo / 30)
        return `Dokumen '${docName}' sudah expired lebih dari ${monthsAgo} bulan yang lalu. Silakan perbarui atau perpanjang dokumen tersebut.`
      }
    } else if (diffDays === 0) {
      // Expired hari ini
      return `Dokumen '${docName}' akan expired hari ini. Silakan perbarui atau perpanjang dokumen tersebut.`
    } else if (diffDays === 1) {
      // Expired besok
      return `Dokumen '${docName}' akan expired dalam 1 hari. Silakan perbarui atau perpanjang dokumen tersebut.`
    } else {
      // Masih beberapa hari lagi
      return `Dokumen '${docName}' akan expired dalam ${diffDays} hari. Silakan perbarui atau perpanjang dokumen tersebut.`
    }
  }
  
  // Untuk notification type lain, gunakan message as-is
  return notif.message
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
      // Jika polling sudah berjalan, reload settings dan load notifications
      loadNotificationSettings().then(() => {
        loadNotifications()
      }).catch(() => {
        loadNotifications()
      })
    }
  }
  // Update previous state
  previousAuthState = isAuthenticated
})

onMounted(() => {
  // Inject CSS untuk custom styling expired document notification
  const style = document.createElement('style')
  style.id = 'expired-document-notification-style'
  style.textContent = `
    /* Custom styling untuk expired document notification */
    /* Gunakan selector yang lebih luas untuk memastikan styling diterapkan */
    .expired-document-notification,
    .ant-notification-notice.expired-document-notification,
    .ant-notification-topRight .ant-notification-notice.expired-document-notification,
    .ant-notification .ant-notification-notice.expired-document-notification,
    li.ant-notification-notice.expired-document-notification {
      background: #bf4e4e !important;
      background-color: #bf4e4e !important;
      border: 2px dashed #ccc !important;
      border-radius: 8px !important;
      backdrop-filter: blur(10px) !important;
      -webkit-backdrop-filter: blur(10px) !important;
    }
    
    /* Override untuk inner content agar tetap readable dengan blur effect */
    .expired-document-notification .ant-notification-notice-content,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-content {
      position: relative;
      z-index: 1;
    }
    
    /* Pastikan icon dan close button tetap visible dengan warna putih */
    .expired-document-notification .ant-notification-notice-icon,
    .expired-document-notification .ant-notification-close-icon,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-icon,
    .ant-notification-notice.expired-document-notification .ant-notification-close-icon {
      color: #fff !important;
    }
    
    /* Pastikan semua text tetap readable dengan warna putih */
    .expired-document-notification,
    .expired-document-notification *,
    .ant-notification-notice.expired-document-notification,
    .ant-notification-notice.expired-document-notification * {
      color: #fff !important;
    }
    
    /* Override khusus untuk message dan description */
    .expired-document-notification .ant-notification-notice-message,
    .expired-document-notification .ant-notification-notice-description,
    .expired-document-notification .ant-notification-notice-message *,
    .expired-document-notification .ant-notification-notice-description *,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-message,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-description,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-message *,
    .ant-notification-notice.expired-document-notification .ant-notification-notice-description * {
      color: #fff !important;
    }
    
    /* Pastikan semua span, div, p, dan elemen text lainnya berwarna putih */
    .expired-document-notification span,
    .expired-document-notification div,
    .expired-document-notification p,
    .expired-document-notification strong,
    .expired-document-notification em,
    .ant-notification-notice.expired-document-notification span,
    .ant-notification-notice.expired-document-notification div,
    .ant-notification-notice.expired-document-notification p,
    .ant-notification-notice.expired-document-notification strong,
    .ant-notification-notice.expired-document-notification em {
      color: #fff !important;
    }
  `
  document.head.appendChild(style)
  
  loadUserCompaniesCount()
  startNotificationPolling()
  
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
  
  // Remove custom CSS untuk expired document notification
  const expiredStyle = document.getElementById('expired-document-notification-style')
  if (expiredStyle) {
    expiredStyle.remove()
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
          <a-badge :count="unreadCount > 99 ? '99+' : unreadCount" :offset="[-10, 10]" style="font-size: 11px !important; padding: 0 4px !important;">
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
                      <div class="notification-message">{{ formatDynamicMessage(notif) }}</div>
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
