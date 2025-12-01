<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import { auditApi, type AuditLog, type AuditLogsParams, type AuditLogStats, type UserActivityLog, type UserActivityLogsParams } from '../api/audit'
import developmentApi from '../api/development'
import { sonarqubeApi, type SonarQubeIssue, type SonarQubeIssuesParams } from '../api/sonarqube'
import type { TableColumnsType } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

// Check if user is superadmin
const isSuperadmin = computed(() => {
  return authStore.user?.role?.toLowerCase() === 'superadmin'
})

// Check if user is admin
const isAdmin = computed(() => {
  return authStore.user?.role?.toLowerCase() === 'admin'
})

const loading = ref(false)
const is2FAEnabled = ref(false)
const setupStep = ref<'idle' | 'generate' | 'verify' | 'success'>('idle')
const qrCode = ref<string>('')
const secret = ref<string>('')
const twoFACode = ref('')
const backupCodes = ref<string[]>([])

// Development features - Combined
const allSeederStatusLoading = ref(false)
const allSeederStatus = ref<{ status: Record<string, boolean>; message: string } | null>(null)
const resetAllLoading = ref(false)
const runAllSeedersLoading = ref(false)

// Audit logs
const auditLogs = ref<AuditLog[]>([])
const auditLoading = ref(false)
const auditStats = ref<AuditLogStats | null>(null)
const auditStatsLoading = ref(false)
const auditPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const auditFilters = ref({
  action: undefined as string | undefined,
  resource: undefined as string | undefined,
  status: undefined as string | undefined,
  logType: undefined as string | undefined, // "user_action" or "technical_error"
})
const selectedAuditLog = ref<AuditLog | null>(null)
const detailModalVisible = ref(false)
let auditStatsInterval: number | null = null

// User Activity Logs (permanent logs untuk data penting)
const userActivityLogs = ref<UserActivityLog[]>([])
const userActivityLoading = ref(false)
const userActivityPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const userActivityFilters = ref({
  action: undefined as string | undefined,
  resource: undefined as string | undefined,
  status: undefined as string | undefined,
})
const selectedUserActivityLog = ref<UserActivityLog | null>(null)
const userActivityDetailModalVisible = ref(false)

// Tab state untuk audit logs
const auditLogActiveTab = ref<string>('audit') // 'audit' atau 'user-activity'

// SonarQube state
const sonarqubeEnabled = ref(false) // Track if feature is enabled
const sonarqubeIssues = ref<SonarQubeIssue[]>([])
const sonarqubeLoading = ref(false)
const sonarqubeExporting = ref(false)
const sonarqubeTotal = ref(0)
const sonarqubeFilters = ref<SonarQubeIssuesParams>({
  severities: ['BLOCKER', 'CRITICAL', 'MAJOR'],
  types: ['BUG', 'VULNERABILITY', 'CODE_SMELL'], // Include CODE_SMELL untuk melihat semua issues
  statuses: ['OPEN', 'CONFIRMED', 'REOPENED'],
})
const sonarqubeComponents = ref<Record<string, string>>({})

// Navigation state
const selectedMenuKey = ref<string>('2fa') // Default: 2FA Setting
const selectedKeys = ref<string[]>(['2fa']) // Array untuk v-model:selectedKeys
const openKeys = ref<string[]>(['2fa']) // Default: 2FA Setting open
const isMaximized = ref(false) // State untuk maximize/minimize

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const handleToggleMaximize = (value: boolean) => {
  isMaximized.value = value
}

const check2FAStatus = async () => {
  try {
    loading.value = true
    const status = await authStore.get2FAStatus()
    is2FAEnabled.value = status.enabled
    if (!status.enabled) {
      setupStep.value = 'idle'
    }
  } catch (error: unknown) {
    console.error('Failed to get 2FA status:', error)
  } finally {
    loading.value = false
  }
}

const handleEnable2FA = async () => {
  try {
    loading.value = true
    const response = await authStore.generate2FA()
    qrCode.value = response.qr_code
    secret.value = response.secret
    setupStep.value = 'generate'
    message.success('QR Code berhasil di-generate. Silakan scan dengan authenticator app Anda.')
  } catch (error: unknown) {
    console.error('Error generating 2FA:', error)
    const axiosError = error as { response?: { data?: { message?: string; Message?: string } }; message?: string }
    const errorMessage = 
      axiosError.response?.data?.message || 
      axiosError.response?.data?.Message ||
      axiosError.message ||
      authStore.error ||
      'Gagal generate QR Code. Pastikan Anda sudah login dan coba lagi.'
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    loading.value = false
  }
}

const handleVerify2FA = async () => {
  if (!twoFACode.value || twoFACode.value.length !== 6) {
    message.error('Kode harus 6 digit')
    return
  }

  try {
    loading.value = true
    const response = await authStore.verify2FA(twoFACode.value)
    backupCodes.value = response.backup_codes || []
    is2FAEnabled.value = true
    setupStep.value = 'success'
    message.success('2FA berhasil diaktifkan!')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || 'Kode verifikasi tidak valid'
    message.error(errorMessage)
  } finally {
    loading.value = false
  }
}

const handleCancelSetup = () => {
  setupStep.value = 'idle'
  qrCode.value = ''
  secret.value = ''
  twoFACode.value = ''
}

const copySecret = () => {
  navigator.clipboard.writeText(secret.value)
  message.success('Secret berhasil di-copy!')
}

const copyBackupCodes = () => {
  const codesText = backupCodes.value.join('\n')
  navigator.clipboard.writeText(codesText)
  message.success('Backup codes berhasil di-copy!')
}

const downloadBackupCodes = () => {
  const codesText = backupCodes.value.join('\n')
  const blob = new Blob([codesText], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'backup-codes.txt'
  a.click()
  URL.revokeObjectURL(url)
  message.success('Backup codes berhasil di-download!')
}

const handleDone = () => {
  setupStep.value = 'idle'
  twoFACode.value = ''
  backupCodes.value = []
}

const handleDisable2FA = async () => {
  try {
    loading.value = true
    await authStore.disable2FA()
    is2FAEnabled.value = false
    setupStep.value = 'idle'
    message.success('2FA berhasil dinonaktifkan')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string; Message?: string } }; message?: string }
    const errorMessage = 
      axiosError.response?.data?.message || 
      axiosError.response?.data?.Message ||
      axiosError.message ||
      authStore.error ||
      'Gagal menonaktifkan 2FA'
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    loading.value = false
  }
}

// Fungsi fetch audit stats
const fetchAuditStats = async () => {
  try {
    auditStatsLoading.value = true
    const stats = await auditApi.getAuditLogStats()
    auditStats.value = stats
  } catch (error: unknown) {
    console.error('Failed to fetch audit stats:', error)
    // Set default values jika error
    if (!auditStats.value) {
      auditStats.value = {
        total_records: 0,
        user_action_count: 0,
        technical_error_count: 0,
        estimated_size_mb: 0,
      }
    }
  } finally {
    auditStatsLoading.value = false
  }
}

// Fungsi audit logs
const fetchAuditLogs = async (page: number = 1, pageSize: number = 10) => {
  try {
    auditLoading.value = true
    const params: AuditLogsParams = {
      page,
      pageSize,
      action: auditFilters.value.action,
      resource: auditFilters.value.resource,
      status: auditFilters.value.status,
      logType: auditFilters.value.logType,
    }
    
    const response = await auditApi.getAuditLogs(params)
    auditLogs.value = response.data
    auditPagination.value = {
      current: response.page,
      pageSize: response.pageSize,
      total: response.total,
    }
    // Tidak refresh stats otomatis saat fetch logs, hanya saat user klik refresh atau auto-refresh interval
  } catch (error: unknown) {
    console.error('Failed to fetch audit logs:', error)
    message.error('Gagal mengambil audit logs')
  } finally {
    auditLoading.value = false
  }
}

const handleAuditTableChange = (pag: { current?: number; pageSize?: number }) => {
  if (pag.current) {
    auditPagination.value.current = pag.current
  }
  if (pag.pageSize) {
    auditPagination.value.pageSize = pag.pageSize
  }
  fetchAuditLogs(auditPagination.value.current, auditPagination.value.pageSize)
}

const handleFilterChange = () => {
  auditPagination.value.current = 1
  fetchAuditLogs(1, auditPagination.value.pageSize)
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString('id-ID', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'success':
      return 'success'
    case 'failure':
      return 'error'
    case 'error':
      return 'error'
    default:
      return 'default'
  }
}

const getActionColor = (action: string) => {
  if (action.includes('create') || action.includes('generate')) {
    return 'green'
  } else if (action.includes('update')) {
    return 'blue'
  } else if (action.includes('delete')) {
    return 'red'
  } else if (action.includes('view') || action.includes('export')) {
    return 'cyan'
  }
  return 'default'
}

const getActionLabel = (action: string) => {
  const labels: Record<string, string> = {
    // Authentication actions
    login: 'Login',
    logout: 'Logout',
    register: 'Register',
    failed_login: 'Failed Login',
    password_reset: 'Password Reset',
    
    // Generic CRUD actions
    create: 'Create',
    update: 'Update',
    delete: 'Delete',
    view: 'View',
    
    // User Management actions
    create_user: 'Create User',
    update_user: 'Update User',
    delete_user: 'Delete User',
    
    // Company/Subsidiary actions
    create_company: 'Create Company',
    update_company: 'Update Company',
    delete_company: 'Delete Company',
    
    // Document actions
    create_document: 'Create Document',
    update_document: 'Update Document',
    delete_document: 'Delete Document',
    view_document: 'View Document',
    
    // File Management actions
    create_file: 'Create File',
    update_file: 'Update File',
    delete_file: 'Delete File',
    download_file: 'Download File',
    view_file: 'View File',
    upload_file: 'Upload File',
    
    // Report Management actions
    generate_report: 'Generate Report',
    view_report: 'View Report',
    export_report: 'Export Report',
    delete_report: 'Delete Report',
    
    // 2FA actions
    enable_2fa: 'Enable 2FA',
    disable_2fa: 'Disable 2FA',
    
    // Other actions
    assign_user_to_company: 'Assign User to Company',
    unassign_user_from_company: 'Unassign User from Company',
    reset_user_password: 'Reset User Password',
    update_email: 'Update Email',
    change_password: 'Change Password',
    change_password_failed: 'Change Password Failed',
    assign_permission: 'Assign Permission',
    revoke_permission: 'Revoke Permission',
    
    // Technical errors
    system_error: 'System Error',
    database_error: 'Database Error',
    validation_error: 'Validation Error',
    panic: 'Panic',
  }
  return labels[action] || action
}

const auditColumns: TableColumnsType = [
  {
    title: 'Waktu',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 180,
    customRender: ({ text }: { text: string }) => formatDate(text),
  },
  {
    title: 'User',
    dataIndex: 'username',
    key: 'username',
    width: 150,
  },
  {
    title: 'Action',
    dataIndex: 'action',
    key: 'action',
    width: 150,
    customRender: ({ text }: { text: string }) => getActionLabel(text),
  },
  {
    title: 'Resource',
    dataIndex: 'resource',
    key: 'resource',
    width: 120,
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: string }) => {
      return h('a-tag', { color: getStatusColor(text) }, { default: () => text.toUpperCase() })
    },
  },
  {
    title: 'Type',
    dataIndex: 'log_type',
    key: 'log_type',
    width: 120,
    customRender: ({ text }: { text: string }) => {
      const typeLabels: Record<string, string> = {
        user_action: 'User Action',
        technical_error: 'Technical Error',
      }
      const typeColors: Record<string, string> = {
        user_action: 'blue',
        technical_error: 'red',
      }
      return h('a-tag', { color: typeColors[text] || 'default' }, { default: () => typeLabels[text] || text })
    },
  },
  {
    title: 'Actions',
    key: 'action_buttons',
    width: 100,
    fixed: 'right',
  },
]

const showDetailModal = (log: AuditLog) => {
  selectedAuditLog.value = log
  detailModalVisible.value = true
}

const closeDetailModal = () => {
  detailModalVisible.value = false
  selectedAuditLog.value = null
}

// Fungsi user activity logs (permanent logs)
const fetchUserActivityLogs = async (page: number = 1, pageSize: number = 10) => {
  try {
    userActivityLoading.value = true
    const params: UserActivityLogsParams = {
      page,
      pageSize,
      action: userActivityFilters.value.action,
      resource: userActivityFilters.value.resource,
      status: userActivityFilters.value.status,
    }
    
    const response = await auditApi.getUserActivityLogs(params)
    userActivityLogs.value = response.data
    userActivityPagination.value = {
      current: response.page,
      pageSize: response.pageSize,
      total: response.total,
    }
  } catch (error: unknown) {
    console.error('Failed to fetch user activity logs:', error)
    message.error('Gagal mengambil user activity logs')
  } finally {
    userActivityLoading.value = false
  }
}

const handleUserActivityTableChange = (pag: { current?: number; pageSize?: number }) => {
  if (pag.current) {
    userActivityPagination.value.current = pag.current
  }
  if (pag.pageSize) {
    userActivityPagination.value.pageSize = pag.pageSize
  }
  fetchUserActivityLogs(userActivityPagination.value.current, userActivityPagination.value.pageSize)
}

const handleUserActivityFilterChange = () => {
  userActivityPagination.value.current = 1
  fetchUserActivityLogs(1, userActivityPagination.value.pageSize)
}

const showUserActivityDetailModal = (log: UserActivityLog) => {
  selectedUserActivityLog.value = log
  userActivityDetailModalVisible.value = true
}

const closeUserActivityDetailModal = () => {
  userActivityDetailModalVisible.value = false
  selectedUserActivityLog.value = null
}

const parseDetails = (detailsJson: string) => {
  if (!detailsJson) return null
  try {
    return JSON.parse(detailsJson)
  } catch {
    return null
  }
}

// SonarQube functions
const checkSonarQubeStatus = async () => {
  try {
    const status = await sonarqubeApi.getStatus()
    sonarqubeEnabled.value = status.enabled
  } catch (error) {
    // If check fails, assume feature is disabled
    sonarqubeEnabled.value = false
    console.warn('SonarQube Monitor status check failed:', error)
  }
}

const fetchSonarQubeIssues = async () => {
  try {
    sonarqubeLoading.value = true
    const response = await sonarqubeApi.getIssues(sonarqubeFilters.value)
    
    // Debug logging
    console.log('SonarQube API Response:', {
      total: response.total,
      issuesCount: response.issues?.length || 0,
      componentsCount: response.components?.length || 0,
      issues: response.issues,
      components: response.components,
    })
    
    sonarqubeIssues.value = response.issues || []
    sonarqubeTotal.value = response.total || 0
    
    // Build component map for display
    const componentMap: Record<string, string> = {}
    if (response.components) {
      response.components.forEach(comp => {
        componentMap[comp.key] = comp.name || comp.key
      })
    }
    sonarqubeComponents.value = componentMap
    
    if (response.total > 0) {
      message.success(`Berhasil mengambil ${response.total} issues dari SonarCloud (${response.issues?.length || 0} ditampilkan)`)
    } else {
      message.warning('Tidak ada issues yang ditemukan dengan filter yang dipilih')
    }
  } catch (error: unknown) {
    console.error('Failed to fetch SonarQube issues:', error)
    const errorMessage = error instanceof Error ? error.message : 'Gagal mengambil issues dari SonarCloud'
    message.error({
      content: errorMessage,
      duration: 5,
    })
    sonarqubeIssues.value = []
  } finally {
    sonarqubeLoading.value = false
  }
}

const handleSonarQubeFilterChange = () => {
  // Filter change handler - bisa digunakan untuk auto-refresh jika diperlukan
  // Untuk sekarang, user harus klik Refresh button
}

const exportSonarQubeIssues = async () => {
  try {
    sonarqubeExporting.value = true
    const blob = await sonarqubeApi.exportIssues(sonarqubeFilters.value)
    
    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `sonarqube-issues-${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    message.success('Berhasil export issues ke JSON')
  } catch (error: unknown) {
    console.error('Failed to export SonarQube issues:', error)
    const errorMessage = error instanceof Error ? error.message : 'Gagal export issues'
    message.error({
      content: errorMessage,
      duration: 5,
    })
  } finally {
    sonarqubeExporting.value = false
  }
}

const getSeverityCount = (severity: string) => {
  return sonarqubeIssues.value.filter(issue => issue.severity === severity).length
}

const getSeverityColor = (severity: string) => {
  switch (severity) {
    case 'BLOCKER':
      return 'red'
    case 'CRITICAL':
      return 'volcano'
    case 'MAJOR':
      return 'orange'
    case 'MINOR':
      return 'blue'
    case 'INFO':
      return 'default'
    default:
      return 'default'
  }
}

const getTypeColor = (type: string) => {
  switch (type) {
    case 'BUG':
      return 'red'
    case 'VULNERABILITY':
      return 'volcano'
    case 'CODE_SMELL':
      return 'orange'
    default:
      return 'default'
  }
}

const getComponentName = (componentKey: string) => {
  return sonarqubeComponents.value[componentKey] || componentKey.split(':').pop() || componentKey
}

const sonarqubeColumns: TableColumnsType = [
  {
    title: 'Severity',
    dataIndex: 'severity',
    key: 'severity',
    width: 100,
    fixed: 'left',
  },
  {
    title: 'Type',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: 'Component',
    dataIndex: 'component',
    key: 'component',
    width: 200,
    ellipsis: true,
  },
  {
    title: 'Message',
    dataIndex: 'message',
    key: 'message',
    width: 300,
    ellipsis: true,
  },
  {
    title: 'Line',
    dataIndex: 'line',
    key: 'line',
    width: 80,
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    width: 100,
  },
  {
    title: 'Rule',
    dataIndex: 'rule',
    key: 'rule',
    width: 150,
    ellipsis: true,
  },
]

// Combined Development functions
const checkAllSeederStatus = async () => {
  try {
    allSeederStatusLoading.value = true
    const status = await developmentApi.checkAllSeederStatus()
    allSeederStatus.value = status
  } catch (error: unknown) {
    console.error('Failed to check all seeder status:', error)
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal memeriksa status seeder'
    message.error(errorMessage)
  } finally {
    allSeederStatusLoading.value = false
  }
}

const handleResetAllSeededData = () => {
  Modal.confirm({
    title: 'Reset Semua Data Seeder',
    content: 'Apakah Anda yakin ingin menghapus semua data yang sudah di-seed (Reports, Companies, Users)? Tindakan ini tidak dapat dibatalkan dan akan menghapus semua relasi data.',
    okText: 'Ya, Reset Semua',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      try {
        resetAllLoading.value = true
        const result = await developmentApi.resetAllSeededData()
        message.success(result.message || 'Semua data seeder berhasil di-reset')
        // Refresh status
        await checkAllSeederStatus()
      } catch (error: unknown) {
        console.error('Failed to reset all seeded data:', error)
        const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
        const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal reset semua data seeder'
        message.error(errorMessage)
      } finally {
        resetAllLoading.value = false
      }
    },
  })
}

const handleRunAllSeeders = async () => {
  Modal.confirm({
    title: 'Jalankan Semua Seeder Data',
    content: 'Ini akan menjalankan semua seeder secara berurutan: Company → Reports. Memastikan relasi data terjaga dengan benar. Lanjutkan?',
    okText: 'Ya, Jalankan',
    cancelText: 'Batal',
    onOk: async () => {
      try {
        runAllSeedersLoading.value = true
        const result = await developmentApi.runAllSeeders()
        message.success(result.message || 'Semua seeder berhasil dijalankan')
        if (result.details) {
          // Show details if available
          const detailsText = Object.entries(result.details).map(([key, value]) => `• ${key}: ${value}`).join('\n')
          Modal.info({
            title: 'Seeder Details',
            content: detailsText,
            okText: 'OK',
          })
        }
        // Refresh status
        await checkAllSeederStatus()
      } catch (error: unknown) {
        console.error('Failed to run all seeders:', error)
        const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
        const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal menjalankan semua seeder'
        message.error(errorMessage)
      } finally {
        runAllSeedersLoading.value = false
      }
    },
  })
}

const handleMenuClick = (e: { key: string }) => {
  selectedMenuKey.value = e.key
  selectedKeys.value = [e.key]
}

const handleOpenChange = (keys: string[]) => {
  openKeys.value = keys
}

onMounted(() => {
  check2FAStatus()
  // Only fetch audit logs if user is superadmin
  if (isSuperadmin.value) {
    fetchAuditLogs()
    fetchAuditStats()
    checkAllSeederStatus()
    
    // Check SonarQube Monitor status (only for superadmin/admin)
    if (isSuperadmin.value || isAdmin.value) {
      checkSonarQubeStatus()
    }
    
    // Auto-refresh stats cards saja setiap 30 detik (hanya untuk audit stats, bukan seluruh halaman)
    // Interval akan dihentikan otomatis saat user keluar dari halaman (onUnmounted)
    auditStatsInterval = window.setInterval(() => {
      fetchAuditStats()
    }, 30000) // 30 detik - hanya refresh stats cards
  }
})

onUnmounted(() => {
  // Clear interval saat komponen di-unmount
  if (auditStatsInterval !== null) {
    clearInterval(auditStatsInterval)
    auditStatsInterval = null
  }
})
</script>

<template>
  <div class="settings-page">
    <DashboardHeader @logout="handleLogout" @toggleMaximize="handleToggleMaximize" />

    <div class="settings-wrapper-layout">
      <!-- Page Header Section -->
      <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">Settings</h1>
            <p class="page-description">
              Kelola pengaturan akun, keamanan, dan konfigurasi sistem Anda.
            </p>
          </div>
        </div>
      </div>

      <!-- Main Content -->
      <div class="mainContentPage">
        <a-layout-content class="settings-content">
          <div class="settings-container">
            <a-card class="settings-wrapper-card">
          <div class="settings-wrapper">
            <!-- Sidebar Navigation -->
            <div class="settings-sidebar" :class="{ 'sidebar-hidden': isMaximized }">
              <a-menu
                v-model:selectedKeys="selectedKeys"
                v-model:openKeys="openKeys"
                mode="inline"
                @click="handleMenuClick"
                @openChange="handleOpenChange"
                class="settings-menu"
              >
                <a-menu-item key="2fa">
                  <template #icon>
                    <IconifyIcon icon="mdi:shield-lock" width="20" />
                  </template>
                  <span>2FA Setting</span>
                </a-menu-item>
                
                <a-menu-item v-if="isSuperadmin" key="development">
                  <template #icon>
                    <IconifyIcon icon="mdi:code-tags" width="20" />
                  </template>
                  <span>Fitur untuk Development</span>
                </a-menu-item>
                
                <a-menu-item v-if="isSuperadmin" key="audit">
                  <template #icon>
                    <IconifyIcon icon="mdi:file-document-outline" width="20" />
                  </template>
                  <span>Audit Log</span>
                </a-menu-item>
                
                <a-menu-item v-if="(isSuperadmin || isAdmin) && sonarqubeEnabled" key="sonarqube">
                  <template #icon>
                    <IconifyIcon icon="mdi:code-tags-check" width="20" />
                  </template>
                  <span>SonarQube Monitor</span>
                </a-menu-item>
              </a-menu>
            </div>

            <!-- Main Content Area -->
            <div class="settings-main-content" :class="{ 'content-maximized': isMaximized }">
          <!-- 2FA Setting Content -->
          <div v-if="selectedMenuKey === '2fa'" class="settings-section">
            <a-card class="security-card" :loading="loading">
              <template #title>
                <div class="card-title">
                  <IconifyIcon icon="mdi:shield-lock" width="24" height="24" />
                  <span>Security</span>
                </div>
              </template>

              <!-- Two-Factor Authentication -->
              <div class="security-section">
            <div class="section-header">
              <div>
                <h3 class="section-title">Two-Factor Authentication (2FA)</h3>
                <p class="section-description">
                  Tambahkan lapisan keamanan ekstra dengan authenticator app
                </p>
              </div>
              <a-tag :color="is2FAEnabled ? 'success' : 'default'">
                {{ is2FAEnabled ? 'Enabled' : 'Disabled' }}
              </a-tag>
            </div>

            <!-- Step 1: Idle / Enable Button -->
            <div v-if="setupStep === 'idle' && !is2FAEnabled" class="setup-content">
              <a-button type="primary" size="large" @click="handleEnable2FA" :loading="loading">
                <IconifyIcon icon="mdi:shield-check" width="18" style="margin-right: 8px;" />
                Enable 2FA
              </a-button>
            </div>

            <!-- Step 2: Generate QR Code -->
            <div v-if="setupStep === 'generate'" class="setup-content">
              <div class="qr-container">
                <h4 class="step-title">Scan QR Code</h4>
                <p class="step-description">
                  Buka aplikasi authenticator (Google Authenticator, Authy, dll) dan scan QR code berikut:
                </p>
                
                <div class="qr-code-wrapper">
                  <img :src="`data:image/png;base64,${qrCode}`" alt="QR Code" class="qr-code-image" />
                </div>

                <div class="manual-entry">
                  <p class="manual-entry-label">Atau masukkan secret secara manual:</p>
                  <div class="secret-display">
                    <code class="secret-code">{{ secret }}</code>
                    <a-button type="text" @click="copySecret" size="small">
                      <IconifyIcon icon="mdi:content-copy" width="16" />
                    </a-button>
                  </div>
                </div>

                <a-divider />

                <h4 class="step-title">Masukkan Kode Verifikasi</h4>
                <p class="step-description">
                  Setelah scan QR code, masukkan kode 6 digit dari authenticator app:
                </p>

                <a-input
                  v-model:value="twoFACode"
                  placeholder="Masukkan kode 6 digit"
                  size="large"
                  :maxlength="6"
                  class="code-input"
                  @keyup.enter="handleVerify2FA"
                >
                  <template #prefix>
                    <IconifyIcon icon="mdi:key" width="18" />
                  </template>
                </a-input>

                <div class="setup-actions">
                  <a-button @click="handleCancelSetup">Cancel</a-button>
                  <a-button type="primary" @click="handleVerify2FA" :loading="loading">
                    Verify & Enable
                  </a-button>
                </div>
              </div>
            </div>

            <!-- Step 3: Success with Backup Codes -->
            <div v-if="setupStep === 'success'" class="setup-content">
              <a-result
                status="success"
                title="2FA Berhasil Diaktifkan!"
                sub-title="Silakan simpan backup codes berikut untuk akses darurat"
              >
                <template #extra>
                  <div class="backup-codes-container">
                    <div class="backup-codes-header">
                      <h4>Backup Codes</h4>
                      <div class="backup-codes-actions">
                        <a-button @click="copyBackupCodes" size="small">
                          <IconifyIcon icon="mdi:content-copy" width="16" style="margin-right: 4px;" />
                          Copy
                        </a-button>
                        <a-button @click="downloadBackupCodes" size="small">
                          <IconifyIcon icon="mdi:download" width="16" style="margin-right: 4px;" />
                          Download
                        </a-button>
                      </div>
                    </div>
                    <div class="backup-codes-list">
                      <div v-for="(code, index) in backupCodes" :key="index" class="backup-code-item">
                        {{ code }}
                      </div>
                    </div>
                    <a-alert
                      message="Penting!"
                      description="Simpan backup codes ini di tempat yang aman. Anda akan membutuhkannya jika kehilangan akses ke authenticator app."
                      type="warning"
                      show-icon
                      style="margin-top: 16px;"
                    />
                    <a-button type="primary" block @click="handleDone" style="margin-top: 16px;">
                      Done
                    </a-button>
                  </div>
                </template>
              </a-result>
            </div>

            <!-- 2FA Enabled Info -->
            <div v-if="is2FAEnabled && setupStep === 'idle'" class="setup-content">
              <a-alert
                message="2FA Aktif"
                description="Two-Factor Authentication sudah aktif untuk akun Anda. Pastikan Anda memiliki akses ke authenticator app saat login."
                type="success"
                show-icon
                style="margin-bottom: 16px;"
              />
              <a-popconfirm
                title="Apakah Anda yakin ingin menonaktifkan 2FA?"
                description="Akun Anda akan menjadi kurang aman setelah 2FA dinonaktifkan."
                ok-text="Ya, Nonaktifkan"
                cancel-text="Batal"
                @confirm="handleDisable2FA"
              >
                <a-button type="default" danger size="large" :loading="loading">
                  <IconifyIcon icon="mdi:shield-off" width="18" style="margin-right: 8px;" />
                  Disable 2FA
                </a-button>
              </a-popconfirm>
            </div>
              </div>
            </a-card>
          </div>

          <!-- Development Features Content -->
          <div v-if="selectedMenuKey === 'development' && isSuperadmin" class="settings-section">
            <a-card class="development-card">
              <template #title>
                <div class="card-title">
                  <IconifyIcon icon="mdi:code-tags" width="24" height="24" />
                  <span>Fitur untuk Development</span>
                </div>
              </template>

              <div class="development-section">
                <div class="section-header">
                  <div>
                    <h3 class="section-title">Manajemen Data Seeder</h3>
                    <p class="section-description">
                      Reset dan seed semua data sample (Company, User, Reports) secara terpusat. Memastikan relasi data terjaga dengan benar.
                    </p>
                  </div>
                </div>

                <!-- All Seeder Status -->
                <div class="seeder-status" style="margin-bottom: 24px;">
                  <a-space size="middle" align="center" wrap>
                    <a-spin v-if="allSeederStatusLoading" size="small" />
                    <template v-else-if="allSeederStatus">
                      <a-tag :color="allSeederStatus.status.company ? 'success' : 'default'">
                        <IconifyIcon :icon="allSeederStatus.status.company ? 'mdi:check-circle' : 'mdi:alert-circle'" width="16" style="margin-right: 4px;" />
                        Company: {{ allSeederStatus.status.company ? 'Tersedia' : 'Belum Tersedia' }}
                      </a-tag>
                      <a-tag :color="allSeederStatus.status.report ? 'success' : 'default'">
                        <IconifyIcon :icon="allSeederStatus.status.report ? 'mdi:check-circle' : 'mdi:alert-circle'" width="16" style="margin-right: 4px;" />
                        Report: {{ allSeederStatus.status.report ? 'Tersedia' : 'Belum Tersedia' }}
                      </a-tag>
                    </template>
                    <a-button size="small" @click="checkAllSeederStatus" :loading="allSeederStatusLoading">
                      <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                      Refresh Status
                    </a-button>
                  </a-space>
                </div>

                <!-- Action Buttons -->
                <div class="development-actions">
                  <a-space size="large" direction="vertical" style="width: 100%;">
                    <a-button
                      type="primary"
                      danger
                      size="large"
                      block
                      @click="handleResetAllSeededData"
                      :loading="resetAllLoading"
                    >
                      <IconifyIcon icon="mdi:delete-sweep" width="18" style="margin-right: 8px;" />
                      Reset Semua Data Seeder
                    </a-button>
                    <a-button
                      type="primary"
                      size="large"
                      block
                      @click="handleRunAllSeeders"
                      :loading="runAllSeedersLoading"
                    >
                      <IconifyIcon icon="mdi:database-plus" width="18" style="margin-right: 8px;" />
                      Seeder Data
                    </a-button>
                  </a-space>
                </div>

                <!-- Info Alert -->
                <a-alert
                  message="Informasi"
                  description="Seeder Data akan menjalankan semua seeder secara berurutan: Company → Reports. Ini memastikan relasi data terjaga dengan benar. Reset Semua Data Seeder akan menghapus semua data yang sudah di-seed (Reports → Companies) untuk memastikan relasi dihapus dengan benar."
                  type="info"
                  show-icon
                  style="margin-top: 24px;"
                />
              </div>
            </a-card>
          </div>

          <!-- Audit Logs Content -->
          <div v-if="selectedMenuKey === 'audit' && isSuperadmin" class="settings-section">
            <a-card class="audit-logs-card">
              <template #title>
                <div class="card-title">
                  <IconifyIcon icon="mdi:file-document-outline" width="24" height="24" />
                  <span>Audit Logs</span>
                </div>
              </template>

              <!-- Tabs untuk Audit Logs dan User Activity -->
              <a-tabs v-model:activeKey="auditLogActiveTab" class="audit-tabs">
                <a-tab-pane key="audit" tab="Audit Logs">
                  <div class="audit-logs-section">
                <!-- Fixed Header Section -->
                <div class="section-header-fixed">
                  <div class="section-header-content">
                    <div>
                      <h3 class="section-title">Activity Log</h3>
                      <p class="section-description">
                        Riwayat aktivitas dan akses ke sistem
                      </p>
                    </div>
                  </div>
                </div>

                <!-- Scrollable Content -->
                <div class="audit-logs-content">

                  <!-- Filters -->
                  <div class="audit-filters">
              <a-space size="middle" wrap>
                <a-select
                  v-model:value="auditFilters.action"
                  placeholder="Filter by Action"
                  allow-clear
                  style="width: 220px"
                  @change="handleFilterChange"
                  show-search
                  :filter-option="(input: string, option: any) => 
                    option.children.toLowerCase().includes(input.toLowerCase())"
                >
                  <!-- Authentication Actions -->
                  <a-select-option value="login">Login</a-select-option>
                  <a-select-option value="logout">Logout</a-select-option>
                  <a-select-option value="register">Register</a-select-option>
                  <a-select-option value="failed_login">Failed Login</a-select-option>
                  <a-select-option value="password_reset">Password Reset</a-select-option>
                  
                  <!-- User Management -->
                  <a-select-option value="create_user">Create User</a-select-option>
                  <a-select-option value="update_user">Update User</a-select-option>
                  <a-select-option value="delete_user">Delete User</a-select-option>
                  <a-select-option value="reset_user_password">Reset User Password</a-select-option>
                  <a-select-option value="assign_user_to_company">Assign User to Company</a-select-option>
                  <a-select-option value="unassign_user_from_company">Unassign User from Company</a-select-option>
                  <a-select-option value="update_email">Update Email</a-select-option>
                  <a-select-option value="change_password">Change Password</a-select-option>
                  
                  <!-- Company/Subsidiary -->
                  <a-select-option value="create_company">Create Company</a-select-option>
                  <a-select-option value="update_company">Update Company</a-select-option>
                  <a-select-option value="delete_company">Delete Company</a-select-option>
                  
                  <!-- Document -->
                  <a-select-option value="create_document">Create Document</a-select-option>
                  <a-select-option value="update_document">Update Document</a-select-option>
                  <a-select-option value="delete_document">Delete Document</a-select-option>
                  <a-select-option value="view_document">View Document</a-select-option>
                  
                  <!-- File Management -->
                  <a-select-option value="create_file">Create File</a-select-option>
                  <a-select-option value="update_file">Update File</a-select-option>
                  <a-select-option value="delete_file">Delete File</a-select-option>
                  <a-select-option value="upload_file">Upload File</a-select-option>
                  <a-select-option value="download_file">Download File</a-select-option>
                  <a-select-option value="view_file">View File</a-select-option>
                  
                  <!-- Report Management -->
                  <a-select-option value="generate_report">Generate Report</a-select-option>
                  <a-select-option value="view_report">View Report</a-select-option>
                  <a-select-option value="export_report">Export Report</a-select-option>
                  <a-select-option value="delete_report">Delete Report</a-select-option>
                  
                  <!-- Role & Permission -->
                  <a-select-option value="create">Create Role/Permission</a-select-option>
                  <a-select-option value="update">Update Role/Permission</a-select-option>
                  <a-select-option value="delete">Delete Role/Permission</a-select-option>
                  <a-select-option value="assign_permission">Assign Permission</a-select-option>
                  <a-select-option value="revoke_permission">Revoke Permission</a-select-option>
                  
                  <!-- 2FA -->
                  <a-select-option value="enable_2fa">Enable 2FA</a-select-option>
                  <a-select-option value="disable_2fa">Disable 2FA</a-select-option>
                </a-select>

                <a-select
                  v-model:value="auditFilters.resource"
                  placeholder="Filter by Resource"
                  allow-clear
                  style="width: 150px"
                  @change="handleFilterChange"
                >
                  <a-select-option value="auth">Auth</a-select-option>
                  <a-select-option value="user">User</a-select-option>
                  <a-select-option value="company">Company</a-select-option>
                  <a-select-option value="document">Document</a-select-option>
                  <a-select-option value="file">File</a-select-option>
                  <a-select-option value="report">Report</a-select-option>
                  <a-select-option value="role">Role</a-select-option>
                  <a-select-option value="permission">Permission</a-select-option>
                </a-select>

                <a-select
                  v-model:value="auditFilters.status"
                  placeholder="Filter by Status"
                  allow-clear
                  style="width: 150px"
                  @change="handleFilterChange"
                >
                  <a-select-option value="success">Success</a-select-option>
                  <a-select-option value="failure">Failure</a-select-option>
                  <a-select-option value="error">Error</a-select-option>
                </a-select>

                <a-select
                  v-model:value="auditFilters.logType"
                  placeholder="Filter by Type"
                  allow-clear
                  style="width: 180px"
                  @change="handleFilterChange"
                >
                  <a-select-option value="user_action">User Actions</a-select-option>
                  <a-select-option value="technical_error">Technical Errors</a-select-option>
                </a-select>

                <a-button @click="handleFilterChange">
                  <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                  Refresh
                </a-button>
              </a-space>
            </div>

                  <!-- Summary Stats Cards -->
                  <div class="audit-stats-cards">
                    <div class="stats-header">
                      <h4>Audit Log Statistics</h4>
                      <div class="stats-header-actions">
                        <a-spin v-if="auditStatsLoading" size="small" />
                        <span v-if="auditStatsLoading" class="refreshing-text">Refreshing...</span>
                        <a-button 
                          size="small" 
                          type="text" 
                          @click="fetchAuditStats"
                          :loading="auditStatsLoading"
                          class="refresh-stats-btn"
                        >
                          <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                          Refresh Stats
                        </a-button>
                      </div>
                    </div>
              <a-row :gutter="[16, 16]">
                <a-col :xs="24" :sm="12" :md="6">
                  <a-card :loading="auditStatsLoading" size="small">
                    <a-statistic
                      title="Total Records"
                      :value="auditStats?.total_records || 0"
                      :value-style="{ color: '#1890ff' }"
                    >
                      <template #prefix>
                        <IconifyIcon icon="mdi:database" width="20" />
                      </template>
                    </a-statistic>
                  </a-card>
                </a-col>
                <a-col :xs="24" :sm="12" :md="6">
                  <a-card :loading="auditStatsLoading" size="small">
                    <a-statistic
                      title="User Actions"
                      :value="auditStats?.user_action_count || 0"
                      :value-style="{ color: '#52c41a' }"
                    >
                      <template #prefix>
                        <IconifyIcon icon="mdi:account-check" width="20" />
                      </template>
                    </a-statistic>
                  </a-card>
                </a-col>
                <a-col :xs="24" :sm="12" :md="6">
                  <a-card :loading="auditStatsLoading" size="small">
                    <a-statistic
                      title="Technical Errors"
                      :value="auditStats?.technical_error_count || 0"
                      :value-style="{ color: '#ff4d4f' }"
                    >
                      <template #prefix>
                        <IconifyIcon icon="mdi:alert-circle" width="20" />
                      </template>
                    </a-statistic>
                  </a-card>
                </a-col>
                <a-col :xs="24" :sm="12" :md="6">
                  <a-card :loading="auditStatsLoading" size="small">
                    <a-statistic
                      title="Estimated Size"
                      :value="auditStats ? auditStats.estimated_size_mb.toFixed(2) : '0.00'"
                      suffix="MB"
                      :value-style="{ color: '#722ed1' }"
                    >
                      <template #prefix>
                        <IconifyIcon icon="mdi:harddisk" width="20" />
                      </template>
                    </a-statistic>
                    <div style="margin-top: 8px; font-size: 12px; color: #8c8c8c;">
                      <span v-if="auditStats?.retention_policy">
                        Retention: {{ auditStats.retention_policy.user_action_days }}d / {{ auditStats.retention_policy.technical_error_days }}d
                      </span>
                    </div>
                  </a-card>
                </a-col>
              </a-row>
                    <div class="stats-footer">
                      <span class="auto-refresh-note">Stats akan auto-refresh setiap 30 detik</span>
                      <span class="retention-info">Retention: 90d/30d</span>
                    </div>
                  </div>

            <!-- Table -->
                  <!-- Audit Table -->
                  <div class="audit-table">
              <a-table
                :columns="auditColumns"
                :data-source="auditLogs"
                :loading="auditLoading"
                :pagination="{
                  current: auditPagination.current,
                  pageSize: auditPagination.pageSize,
                  total: auditPagination.total,
                  showSizeChanger: true,
                  showTotal: (total: number) => `Total ${total} logs`,
                  pageSizeOptions: ['10', '20', '50', '100'],
                }"
                :scroll="{ x: 800 }"
                @change="handleAuditTableChange"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'action_buttons'">
                    <a-button type="link" size="small" @click="showDetailModal(record)">
                      Detail
                    </a-button>
                  </template>
                </template>
              </a-table>
                  </div>
                </div>
              </div>
                </a-tab-pane>

                <!-- Tab User Activity -->
                <a-tab-pane key="user-activity" tab="User Activity">
                  <div class="audit-logs-section">
                    <!-- Fixed Header Section -->
                    <div class="section-header-fixed">
                      <div class="section-header-content">
                        <div>
                          <h3 class="section-title">User Activity Logs</h3>
                          <p class="section-description">
                            Riwayat aktivitas permanen untuk data penting: Report, Document, Subsidiary, dan User Management
                          </p>
                        </div>
                      </div>
                    </div>

                    <!-- Scrollable Content -->
                    <div class="audit-logs-content">
                      <!-- Filters -->
                      <div class="audit-filters">
                        <a-space size="middle" wrap>
                          <a-select
                            v-model:value="userActivityFilters.action"
                            placeholder="Filter by Action"
                            allow-clear
                            style="width: 220px"
                            @change="handleUserActivityFilterChange"
                            show-search
                            :filter-option="(input: string, option: any) => 
                              option.children.toLowerCase().includes(input.toLowerCase())"
                          >
                            <!-- Report Actions -->
                            <a-select-option value="generate_report">Generate Report</a-select-option>
                            <a-select-option value="view_report">View Report</a-select-option>
                            <a-select-option value="export_report">Export Report</a-select-option>
                            <a-select-option value="delete_report">Delete Report</a-select-option>
                            
                            <!-- Document Actions -->
                            <a-select-option value="create_document">Create Document</a-select-option>
                            <a-select-option value="update_document">Update Document</a-select-option>
                            <a-select-option value="delete_document">Delete Document</a-select-option>
                            <a-select-option value="view_document">View Document</a-select-option>
                            
                            <!-- Company/Subsidiary Actions -->
                            <a-select-option value="create_company">Create Company</a-select-option>
                            <a-select-option value="update_company">Update Company</a-select-option>
                            <a-select-option value="delete_company">Delete Company</a-select-option>
                            
                            <!-- User Management Actions -->
                            <a-select-option value="create_user">Create User</a-select-option>
                            <a-select-option value="update_user">Update User</a-select-option>
                            <a-select-option value="delete_user">Delete User</a-select-option>
                          </a-select>

                          <a-select
                            v-model:value="userActivityFilters.resource"
                            placeholder="Filter by Resource"
                            allow-clear
                            style="width: 180px"
                            @change="handleUserActivityFilterChange"
                          >
                            <a-select-option value="report">Report</a-select-option>
                            <a-select-option value="document">Document</a-select-option>
                            <a-select-option value="company">Company</a-select-option>
                            <a-select-option value="user">User</a-select-option>
                          </a-select>

                          <a-select
                            v-model:value="userActivityFilters.status"
                            placeholder="Filter by Status"
                            allow-clear
                            style="width: 150px"
                            @change="handleUserActivityFilterChange"
                          >
                            <a-select-option value="success">Success</a-select-option>
                            <a-select-option value="failure">Failure</a-select-option>
                            <a-select-option value="error">Error</a-select-option>
                          </a-select>

                          <a-button @click="handleUserActivityFilterChange">
                            <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                            Refresh
                          </a-button>
                        </a-space>
                      </div>

                      <!-- Info Alert -->
                      <a-alert
                        message="Permanent Storage"
                        description="Data ini disimpan secara permanen (tidak ada retention policy) untuk keperluan compliance dan legal. Hanya menampilkan aktivitas untuk resource: Report, Document, Company, dan User Management."
                        type="info"
                        show-icon
                        style="margin-top: 16px; margin-bottom: 16px;"
                      />

                      <!-- User Activity Table -->
                      <div class="audit-table">
                        <a-table
                          :columns="auditColumns.filter(col => col.key !== 'log_type')"
                          :data-source="userActivityLogs"
                          :loading="userActivityLoading"
                          :pagination="{
                            current: userActivityPagination.current,
                            pageSize: userActivityPagination.pageSize,
                            total: userActivityPagination.total,
                            showSizeChanger: true,
                            showTotal: (total: number) => `Total ${total} logs`,
                            pageSizeOptions: ['10', '20', '50', '100'],
                          }"
                          :scroll="{ x: 800 }"
                          @change="handleUserActivityTableChange"
                        >
                          <template #bodyCell="{ column, record }">
                            <template v-if="column.key === 'action_buttons'">
                              <a-button type="link" size="small" @click="showUserActivityDetailModal(record)">
                                Detail
                              </a-button>
                            </template>
                          </template>
                        </a-table>
                      </div>
                    </div>
                  </div>
                </a-tab-pane>
              </a-tabs>
            </a-card>
          </div>

          <!-- SonarQube Monitor Content -->
          <div v-if="selectedMenuKey === 'sonarqube' && (isSuperadmin || isAdmin) && sonarqubeEnabled" class="settings-section">
            <a-card class="sonarqube-card">
              <template #title>
                <div class="card-title">
                  <IconifyIcon icon="mdi:code-tags-check" width="24" height="24" />
                  <span>SonarQube Monitor</span>
                </div>
              </template>

              <div class="sonarqube-section">
                <!-- Fixed Header Section -->
                <div class="section-header-fixed">
                  <div class="section-header-content">
                    <div>
                      <h3 class="section-title">Code Quality Issues</h3>
                      <p class="section-description">
                        Monitor dan analisis issues dari SonarCloud untuk kebutuhan VAPT
                      </p>
                    </div>
                  </div>
                </div>

                <!-- Scrollable Content -->
                <div class="sonarqube-content">
                  <!-- Filters -->
                  <div class="sonarqube-filters">
                    <a-space size="middle" wrap>
                      <a-select
                        v-model:value="sonarqubeFilters.severities"
                        placeholder="Filter by Severity"
                        allow-clear
                        mode="multiple"
                        style="width: 200px"
                        @change="handleSonarQubeFilterChange"
                      >
                        <a-select-option value="BLOCKER">BLOCKER</a-select-option>
                        <a-select-option value="CRITICAL">CRITICAL</a-select-option>
                        <a-select-option value="MAJOR">MAJOR</a-select-option>
                        <a-select-option value="MINOR">MINOR</a-select-option>
                        <a-select-option value="INFO">INFO</a-select-option>
                      </a-select>

                      <a-select
                        v-model:value="sonarqubeFilters.types"
                        placeholder="Filter by Type"
                        allow-clear
                        mode="multiple"
                        style="width: 200px"
                        @change="handleSonarQubeFilterChange"
                      >
                        <a-select-option value="BUG">BUG</a-select-option>
                        <a-select-option value="VULNERABILITY">VULNERABILITY</a-select-option>
                        <a-select-option value="CODE_SMELL">CODE_SMELL</a-select-option>
                      </a-select>

                      <a-select
                        v-model:value="sonarqubeFilters.statuses"
                        placeholder="Filter by Status"
                        allow-clear
                        mode="multiple"
                        style="width: 200px"
                        @change="handleSonarQubeFilterChange"
                      >
                        <a-select-option value="OPEN">OPEN</a-select-option>
                        <a-select-option value="CONFIRMED">CONFIRMED</a-select-option>
                        <a-select-option value="REOPENED">REOPENED</a-select-option>
                        <a-select-option value="RESOLVED">RESOLVED</a-select-option>
                      </a-select>

                      <a-button type="primary" @click="fetchSonarQubeIssues" :loading="sonarqubeLoading">
                        <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                        Refresh
                      </a-button>

                      <a-button @click="exportSonarQubeIssues" :loading="sonarqubeExporting" :disabled="sonarqubeIssues.length === 0">
                        <IconifyIcon icon="mdi:download" width="16" style="margin-right: 4px;" />
                        Export JSON
                      </a-button>
                    </a-space>
                  </div>

                  <!-- Summary Stats -->
                  <div v-if="sonarqubeIssues.length > 0" class="sonarqube-stats" style="margin-bottom: 20px;">
                    <a-row :gutter="[16, 16]">
                      <a-col :xs="24" :sm="12" :md="6">
                        <a-card size="small">
                          <a-statistic
                            title="Total Issues"
                            :value="sonarqubeIssues.length"
                            :value-style="{ color: '#1890ff' }"
                          >
                            <template #prefix>
                              <IconifyIcon icon="mdi:alert-circle" width="20" />
                            </template>
                          </a-statistic>
                        </a-card>
                      </a-col>
                      <a-col :xs="24" :sm="12" :md="6">
                        <a-card size="small">
                          <a-statistic
                            title="BLOCKER"
                            :value="getSeverityCount('BLOCKER')"
                            :value-style="{ color: '#ff4d4f' }"
                          />
                        </a-card>
                      </a-col>
                      <a-col :xs="24" :sm="12" :md="6">
                        <a-card size="small">
                          <a-statistic
                            title="CRITICAL"
                            :value="getSeverityCount('CRITICAL')"
                            :value-style="{ color: '#ff7875' }"
                          />
                        </a-card>
                      </a-col>
                      <a-col :xs="24" :sm="12" :md="6">
                        <a-card size="small">
                          <a-statistic
                            title="MAJOR"
                            :value="getSeverityCount('MAJOR')"
                            :value-style="{ color: '#faad14' }"
                          />
                        </a-card>
                      </a-col>
                    </a-row>
                  </div>

                  <!-- Issues Table -->
                  <div class="sonarqube-table">
                    <a-table
                      :columns="sonarqubeColumns"
                      :data-source="sonarqubeIssues"
                      :loading="sonarqubeLoading"
                      :pagination="{
                        current: 1,
                        pageSize: 20,
                        total: sonarqubeTotal,
                        showSizeChanger: true,
                        showTotal: (total: number) => `Total ${total} issues`,
                        pageSizeOptions: ['10', '20', '50', '100'],
                      }"
                      :scroll="{ x: 1000 }"
                    >
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'severity'">
                          <a-tag :color="getSeverityColor(record.severity)">
                            {{ record.severity }}
                          </a-tag>
                        </template>
                        <template v-if="column.key === 'type'">
                          <a-tag :color="getTypeColor(record.type)">
                            {{ record.type }}
                          </a-tag>
                        </template>
                        <template v-if="column.key === 'status'">
                          <a-tag :color="getStatusColor(record.status)">
                            {{ record.status }}
                          </a-tag>
                        </template>
                        <template v-if="column.key === 'component'">
                          <code>{{ getComponentName(record.component) }}</code>
                        </template>
                      </template>
                    </a-table>
                  </div>
                </div>
              </div>
            </a-card>
          </div>
            </div>
          </div>
        </a-card>
          </div>
        </a-layout-content>
      </div>
    </div>

    <!-- Audit Log Detail Modal (Superadmin Only) -->
    <a-modal
      v-if="isSuperadmin"
      v-model:open="detailModalVisible"
      title="Detail Audit Log"
      width="800px"
      :footer="null"
      @cancel="closeDetailModal"
    >
      <div v-if="selectedAuditLog" class="audit-log-detail">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="ID" :span="2">
            <code>{{ selectedAuditLog.id }}</code>
          </a-descriptions-item>
          <a-descriptions-item label="Waktu" :span="2">
            {{ formatDate(selectedAuditLog.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="User">
            {{ selectedAuditLog.username }}
          </a-descriptions-item>
          <a-descriptions-item label="User ID">
            <code>{{ selectedAuditLog.user_id }}</code>
          </a-descriptions-item>
          <a-descriptions-item label="Action">
            {{ getActionLabel(selectedAuditLog.action) }}
          </a-descriptions-item>
          <a-descriptions-item label="Resource">
            {{ selectedAuditLog.resource || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Resource ID" :span="2">
            {{ selectedAuditLog.resource_id || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Status" :span="2">
            <a-tag :color="getStatusColor(selectedAuditLog.status)">
              {{ selectedAuditLog.status.toUpperCase() }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="IP Address" :span="2">
            <code>{{ selectedAuditLog.ip_address || '-' }}</code>
          </a-descriptions-item>
          <a-descriptions-item label="User Agent" :span="2">
            <div class="user-agent-text">{{ selectedAuditLog.user_agent || '-' }}</div>
          </a-descriptions-item>
          <a-descriptions-item label="Details" :span="2">
            <div v-if="selectedAuditLog.details" class="details-container">
              <pre class="details-json">{{ JSON.stringify(parseDetails(selectedAuditLog.details), null, 2) }}</pre>
            </div>
            <span v-else>-</span>
          </a-descriptions-item>
        </a-descriptions>

        <div class="modal-footer" style="margin-top: 24px; text-align: right;">
          <a-button @click="closeDetailModal">Close</a-button>
        </div>
      </div>
    </a-modal>

    <!-- User Activity Log Detail Modal (Superadmin Only) -->
    <a-modal
      v-if="isSuperadmin"
      v-model:open="userActivityDetailModalVisible"
      title="Detail User Activity Log"
      width="800px"
      :footer="null"
      @cancel="closeUserActivityDetailModal"
    >
      <div v-if="selectedUserActivityLog" class="audit-log-detail">
        <a-descriptions bordered :column="1" size="small">
          <a-descriptions-item label="ID">
            {{ selectedUserActivityLog.id }}
          </a-descriptions-item>
          <a-descriptions-item label="User ID">
            {{ selectedUserActivityLog.user_id }}
          </a-descriptions-item>
          <a-descriptions-item label="Username">
            {{ selectedUserActivityLog.username }}
          </a-descriptions-item>
          <a-descriptions-item label="Action">
            <a-tag :color="getActionColor(selectedUserActivityLog.action)">
              {{ getActionLabel(selectedUserActivityLog.action) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Resource">
            <a-tag color="blue">{{ selectedUserActivityLog.resource }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Resource ID">
            {{ selectedUserActivityLog.resource_id || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="getStatusColor(selectedUserActivityLog.status)">
              {{ selectedUserActivityLog.status.toUpperCase() }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="IP Address">
            {{ selectedUserActivityLog.ip_address || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="User Agent">
            {{ selectedUserActivityLog.user_agent || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Created At">
            {{ formatDate(selectedUserActivityLog.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="Details" v-if="selectedUserActivityLog.details">
            <pre class="details-json">{{ parseDetails(selectedUserActivityLog.details) }}</pre>
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="scss">
.settings-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.settings-wrapper-layout {
  width: 100%;
}

.settings-content {
  padding: 0;
  background: #f5f5f5;
  overflow-y: auto;
}

.settings-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

.settings-wrapper-card {
background: white;
  :deep(.ant-card-body) {
    padding: 0;
  }
}

.settings-wrapper {
  display: flex;
  min-height: 600px;
  gap: 0;
}

.settings-sidebar {
  width: 300px;
  background: #fff;
  border-right: 1px solid #e8e8e8;
  padding: 16px;
  flex-shrink: 0;
  border-top-left-radius: 20px;
  border-bottom-left-radius: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  min-width: 300px;

  &.sidebar-hidden {
    width: 0 !important;
    min-width: 0;
    padding: 0;
    margin: 0;
    border-right: none;
    opacity: 0;
    pointer-events: none;
  }

  .settings-menu {
    border-right: none;
    background: transparent;
    height: 100%;

    :deep(.ant-menu-item) {
      margin: 4px 8px;
      border-radius: 6px;
      height: 40px;
      line-height: 40px;
      
      &:hover {
        background: #f0f7ff;
      }

      &.ant-menu-item-selected {
        background: #e6f4ff;
        color: #035CAB;
        font-weight: 500;

        &::after {
          display: none;
        }
      }
    }

    :deep(.ant-menu-item-icon) {
      margin-right: 12px;
    }
  }
}

.settings-main-content {
  flex: 1;
  padding: 24px;
  background: transparent;
  overflow-y: auto;
  -ms-overflow-style: none;
  scrollbar-width: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  
  &.content-maximized {
    flex: 1 1 100%;
    width: 100%;
    max-width: 100%;
  }
  
  &::-webkit-scrollbar {
    display: none;
    width: 0;
    height: 0;
  }
  
  &::-webkit-scrollbar-track {
    display: none;
  }
}

.settings-section {
  width: 100%;

  :deep(.ant-card) {
    border-radius: 12px;
    // box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    border: 1px solid #f0f0f0;
  }

  :deep(.ant-card-head) {
    padding: 20px 24px;
    border-bottom: 1px solid #f0f0f0;
    background: #fafafa;
    border-radius: 12px 12px 0 0;
  }

  :deep(.ant-card-body) {
    padding: 24px;
  }
}

.security-card {
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
  }
}

.security-section {
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;
  }

  .section-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 8px 0;
    color: #1a1a1a;
    line-height: 1.4;
  }

  .section-description {
    margin: 0;
    color: #666;
    font-size: 14px;
    line-height: 1.5;
  }
}

.setup-content {
  padding: 16px 0;
}

.qr-container {
  max-width: 500px;

  .step-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 8px 0;
    color: #1a1a1a;
  }

  .step-description {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
  }

  .qr-code-wrapper {
    display: flex;
    justify-content: center;
    padding: 24px;
    background: #fff;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    margin-bottom: 24px;
  }

  .qr-code-image {
    width: 200px;
    height: 200px;
  }

  .manual-entry {
    margin-bottom: 24px;

    .manual-entry-label {
      font-size: 14px;
      color: #666;
      margin-bottom: 8px;
    }

    .secret-display {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 12px;
      background: #f5f5f5;
      border: 1px solid #e8e8e8;
      border-radius: 6px;

      .secret-code {
        flex: 1;
        font-family: monospace;
        font-size: 14px;
        color: #1a1a1a;
        word-break: break-all;
      }
    }
  }

  .code-input {
    margin-bottom: 16px;
  }

  .setup-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
}

.backup-codes-container {
  max-width: 500px;
  margin: 0 auto;

  .backup-codes-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h4 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
    }

    .backup-codes-actions {
      display: flex;
      gap: 8px;
    }
  }

  .backup-codes-list {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    padding: 16px;
    background: #f5f5f5;
    border: 1px solid #e8e8e8;
    border-radius: 6px;
    margin-bottom: 16px;

    .backup-code-item {
      font-family: monospace;
      font-size: 14px;
      padding: 8px;
      background: #fff;
      border: 1px solid #e8e8e8;
      border-radius: 4px;
      text-align: center;
    }
  }
}

.development-card {
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
  }
}

.development-section {
  .section-header {
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;
  }

  .section-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 8px 0;
    color: #1a1a1a;
    line-height: 1.4;
  }

  .section-description {
    margin: 0;
    color: #666;
    font-size: 14px;
    line-height: 1.5;
  }

  .seeder-status {
    padding: 16px;
    background: #fafafa;
    border-radius: 8px;
    border: 1px solid #f0f0f0;
    margin-bottom: 20px;
  }

  .development-actions {
    margin-top: 20px;
    margin-bottom: 0;

    :deep(.ant-space) {
      width: 100%;
    }

    :deep(.ant-btn) {
      height: 44px;
      font-size: 15px;
      border-radius: 8px;
    }
  }
}

.audit-logs-card {
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
  }
}

.audit-logs-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;

  // Fixed Header Section
  .section-header-fixed {
    position: sticky;
    top: 0;
    z-index: 10;
    background: #fff;
    padding: 20px 0 16px 0;
    margin-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;

    .section-header-content {
      .section-title {
        font-size: 20px;
        font-weight: 600;
        margin: 0 0 6px 0;
        color: #1a1a1a;
        line-height: 1.4;
      }

      .section-description {
        margin: 0;
        color: #666;
        font-size: 14px;
        line-height: 1.5;
      }
    }
  }

  // Scrollable Content
  .audit-logs-content {
    flex: 1;
    overflow-y: auto;
    padding-right: 4px;
    
    // Hide scrollbar but keep functionality
    -ms-overflow-style: none;
    scrollbar-width: none;
    &::-webkit-scrollbar {
      display: none;
      width: 0;
      height: 0;
    }
  }

  .audit-filters {
    margin-bottom: 20px;
    padding: 16px;
    background: #fafafa;
    border-radius: 8px;
    border: 1px solid #f0f0f0;
  }

  .audit-stats-cards {
    margin-bottom: 24px;

    .stats-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      flex-wrap: wrap;
      gap: 12px;

      h4 {
        font-size: 16px;
        font-weight: 600;
        margin: 0;
        color: #1a1a1a;
      }

      .stats-header-actions {
        display: flex;
        align-items: center;
        gap: 8px;

        .refreshing-text {
          font-size: 12px;
          color: #8c8c8c;
        }

        .refresh-stats-btn {
          padding: 0 8px;
          height: auto;
        }
      }
    }

    .stats-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-top: 12px;
      padding-top: 12px;
      border-top: 1px solid #f0f0f0;
      flex-wrap: wrap;
      gap: 8px;

      .auto-refresh-note {
        font-size: 12px;
        color: #8c8c8c;
      }

      .retention-info {
        font-size: 12px;
        color: #8c8c8c;
        font-weight: 500;
      }
    }

    :deep(.ant-row) {
      margin: 0 -8px;
    }

    :deep(.ant-col) {
      padding: 0 8px;
      margin-bottom: 16px;
    }

    :deep(.ant-card) {
      border-radius: 8px;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
      transition: all 0.3s ease;
      border: 1px solid #f0f0f0;

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
        transform: translateY(-2px);
      }
    }
  }

  .audit-table {
    margin-top: 0;

    .text-ellipsis {
      display: block;
      max-width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    :deep(.ant-table) {
      border-radius: 8px;
      overflow: hidden;
    }

    :deep(.ant-table-thead > tr > th) {
      background: #fafafa;
      font-weight: 600;
      border-bottom: 2px solid #f0f0f0;
    }
  }
}

.sonarqube-card {
  .card-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 18px;
    font-weight: 600;
  }
}

.sonarqube-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;

  // Fixed Header Section
  .section-header-fixed {
    position: sticky;
    top: 0;
    z-index: 10;
    background: #fff;
    padding: 20px 0 16px 0;
    margin-bottom: 20px;
    border-bottom: 1px solid #f0f0f0;

    .section-header-content {
      .section-title {
        font-size: 20px;
        font-weight: 600;
        margin: 0 0 6px 0;
        color: #1a1a1a;
        line-height: 1.4;
      }

      .section-description {
        margin: 0;
        color: #666;
        font-size: 14px;
        line-height: 1.5;
      }
    }
  }

  // Scrollable Content
  .sonarqube-content {
    flex: 1;
    overflow-y: auto;
    padding-right: 4px;
    
    // Hide scrollbar but keep functionality
    -ms-overflow-style: none;
    scrollbar-width: none;
    &::-webkit-scrollbar {
      display: none;
      width: 0;
      height: 0;
    }
  }

  .sonarqube-filters {
    margin-bottom: 20px;
    padding: 16px;
    background: #fafafa;
    border-radius: 8px;
    border: 1px solid #f0f0f0;
  }

  .sonarqube-stats {
    :deep(.ant-row) {
      margin: 0 -8px;
    }

    :deep(.ant-col) {
      padding: 0 8px;
      margin-bottom: 16px;
    }

    :deep(.ant-card) {
      border-radius: 8px;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
      transition: all 0.3s ease;
      border: 1px solid #f0f0f0;

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
        transform: translateY(-2px);
      }
    }
  }

  .sonarqube-table {
    margin-top: 0;

    :deep(.ant-table) {
      border-radius: 8px;
      overflow: hidden;
    }

    :deep(.ant-table-thead > tr > th) {
      background: #fafafa;
      font-weight: 600;
      border-bottom: 2px solid #f0f0f0;
    }
  }
}

.audit-log-detail {
  .user-agent-text {
    word-break: break-all;
    font-size: 12px;
    color: #666;
  }

  .details-container {
    max-height: 300px;
    overflow-y: auto;
    background: #f5f5f5;
    padding: 12px;
    border-radius: 4px;
    border: 1px solid #e8e8e8;

    .details-json {
      margin: 0;
      font-family: 'Courier New', monospace;
      font-size: 12px;
      color: #1a1a1a;
      white-space: pre-wrap;
      word-wrap: break-word;
    }
  }

  code {
    background: #f5f5f5;
    padding: 2px 6px;
    border-radius: 3px;
    font-family: 'Courier New', monospace;
    font-size: 12px;
  }
}

@media (max-width: 768px) {
  .settings-wrapper {
    flex-direction: column;
  }

  .settings-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid #e8e8e8;
    border-top-left-radius: 12px;
    border-top-right-radius: 12px;
    border-bottom-left-radius: 0;
    padding: 12px;
  }

  .settings-main-content {
    padding: 16px;
  }

  .settings-content {
    padding: 0;
  }

  .settings-container {
    padding: 0;
  }

  .audit-logs-section {
    .section-header-fixed {
      padding: 16px 0 12px 0;
      margin-bottom: 16px;
    }

    .audit-filters {
      padding: 12px;
      margin-bottom: 16px;

      :deep(.ant-space) {
        width: 100%;
      }

      :deep(.ant-select),
      :deep(.ant-btn) {
        width: 100%;
      }
    }

    .audit-stats-cards {
      margin-bottom: 20px;

      .stats-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 12px;
      }

      :deep(.ant-col) {
        margin-bottom: 12px;
      }
    }
  }

  .security-section,
  .development-section {
    .section-header {
      flex-direction: column;
      gap: 12px;
      margin-bottom: 20px;
      padding-bottom: 16px;
    }
  }

  .backup-codes-list {
    grid-template-columns: 1fr !important;
  }
}
</style>

