<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import { auditApi, type AuditLog, type AuditLogsParams, type AuditLogStats } from '../api/audit'
import type { TableColumnsType } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const is2FAEnabled = ref(false)
const setupStep = ref<'idle' | 'generate' | 'verify' | 'success'>('idle')
const qrCode = ref<string>('')
const secret = ref<string>('')
const twoFACode = ref('')
const backupCodes = ref<string[]>([])

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

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const check2FAStatus = async () => {
  try {
    loading.value = true
    const status = await authStore.get2FAStatus()
    is2FAEnabled.value = status.enabled
    if (!status.enabled) {
      setupStep.value = 'idle'
    }
  } catch (error: any) {
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
  } catch (error: any) {
    console.error('Error generating 2FA:', error)
    const errorMessage = 
      error?.response?.data?.message || 
      error?.response?.data?.Message ||
      error?.message ||
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
  } catch (error: any) {
    const errorMessage = error?.response?.data?.message || 'Kode verifikasi tidak valid'
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
  } catch (error: any) {
    const errorMessage = 
      error?.response?.data?.message || 
      error?.response?.data?.Message ||
      error?.message ||
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
  } catch (error: any) {
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
  } catch (error: any) {
    console.error('Failed to fetch audit logs:', error)
    message.error('Gagal mengambil audit logs')
  } finally {
    auditLoading.value = false
  }
}

const handleAuditTableChange = (pag: any) => {
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

const getActionLabel = (action: string) => {
  const labels: Record<string, string> = {
    login: 'Login',
    logout: 'Logout',
    register: 'Register',
    create_user: 'Create User',
    update_user: 'Update User',
    delete_user: 'Delete User',
    create_document: 'Create Document',
    update_document: 'Update Document',
    delete_document: 'Delete Document',
    view_document: 'View Document',
    enable_2fa: 'Enable 2FA',
    disable_2fa: 'Disable 2FA',
    failed_login: 'Failed Login',
    password_reset: 'Password Reset',
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

const parseDetails = (detailsJson: string) => {
  if (!detailsJson) return null
  try {
    return JSON.parse(detailsJson)
  } catch {
    return null
  }
}

onMounted(() => {
  check2FAStatus()
  fetchAuditLogs()
  fetchAuditStats()
  
  // Auto-refresh stats cards saja setiap 30 detik (hanya untuk audit stats, bukan seluruh halaman)
  // Interval akan dihentikan otomatis saat user keluar dari halaman (onUnmounted)
  auditStatsInterval = window.setInterval(() => {
    fetchAuditStats()
  }, 30000) // 30 detik - hanya refresh stats cards
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
    <DashboardHeader @logout="handleLogout" />

    <div class="settings-content">
      <div class="settings-container">
        <h1 class="settings-title">Settings</h1>

        <!-- Security Section -->
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

        <!-- Audit Logs Section -->
        <a-card class="audit-logs-card" style="margin-top: 24px;">
          <template #title>
            <div class="card-title">
              <IconifyIcon icon="mdi:file-document-outline" width="24" height="24" />
              <span>Audit Logs</span>
            </div>
          </template>

          <div class="audit-logs-section">
            <div class="section-header">
              <div>
                <h3 class="section-title">Activity Log</h3>
                <p class="section-description">
                  Riwayat aktivitas dan akses ke sistem
                </p>
              </div>
            </div>

            <!-- Filters -->
            <div class="audit-filters">
              <a-space size="middle" wrap>
                <a-select
                  v-model:value="auditFilters.action"
                  placeholder="Filter by Action"
                  allow-clear
                  style="width: 180px"
                  @change="handleFilterChange"
                >
                  <a-select-option value="login">Login</a-select-option>
                  <a-select-option value="logout">Logout</a-select-option>
                  <a-select-option value="enable_2fa">Enable 2FA</a-select-option>
                  <a-select-option value="disable_2fa">Disable 2FA</a-select-option>
                  <a-select-option value="failed_login">Failed Login</a-select-option>
                  <a-select-option value="create_document">Create Document</a-select-option>
                  <a-select-option value="update_document">Update Document</a-select-option>
                  <a-select-option value="delete_document">Delete Document</a-select-option>
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
                  <a-select-option value="document">Document</a-select-option>
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
            <div class="audit-stats-cards" style="margin-bottom: 24px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
                <h4 style="margin: 0;">Audit Log Statistics</h4>
                <div style="display: flex; align-items: center; gap: 8px;">
                  <a-spin v-if="auditStatsLoading" size="small" />
                  <span v-if="auditStatsLoading" style="font-size: 12px; color: #8c8c8c;">Refreshing...</span>
                  <a-button 
                    size="small" 
                    type="text" 
                    @click="fetchAuditStats"
                    :loading="auditStatsLoading"
                    style="padding: 0 8px;"
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
              <div style="margin-top: 12px; font-size: 12px; color: #8c8c8c; text-align: center;">
                Stats akan auto-refresh setiap 30 detik
              </div>
            </div>

            <!-- Table -->
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
        </a-card>
      </div>
    </div>

    <!-- Audit Log Detail Modal -->
    <a-modal
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
  </div>
</template>

<style scoped lang="scss">
.settings-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.settings-content {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.settings-container {
  .settings-title {
    font-size: 28px;
    font-weight: 600;
    margin-bottom: 24px;
    color: #1a1a1a;
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
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #1a1a1a;
  }

  .section-description {
    margin: 0;
    color: #666;
    font-size: 14px;
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
  .section-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #1a1a1a;
  }

  .section-description {
    margin: 0;
    color: #666;
    font-size: 14px;
  }

  .audit-filters {
    margin-bottom: 16px;
    padding: 16px;
    background: #f5f5f5;
    border-radius: 6px;
  }

  .audit-table {
    .text-ellipsis {
      display: block;
      max-width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
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
  .settings-content {
    padding: 16px;
  }

  .backup-codes-list {
    grid-template-columns: 1fr !important;
  }

  .audit-filters {
    :deep(.ant-space) {
      width: 100%;
    }

    :deep(.ant-select) {
      width: 100% !important;
    }
  }
}
</style>

