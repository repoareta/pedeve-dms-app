<template>
  <div class="notifications-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="notifications-content">
      <!-- Page Header Section -->
      <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">Notifikasi</h1>
            <p class="page-description">
              Kelola semua notifikasi dan peringatan sistem.
            </p>
          </div>
        </div>
      </div>

      <!-- Main Content -->
      <div class="mainContentPage">
        <div class="notifications-container">
          <!-- Left Sidebar - Filter -->
          <div class="notifications-sidebar">
            <a-card :bordered="false" class="filter-card">
              <template #title>
                <span>Filter By Status</span>
              </template>
              <div class="sidebar-filters">
                <div
                  v-for="filter in filters"
                  :key="filter.key"
                  class="filter-item"
                  :class="{ active: activeFilter === filter.key }"
                  @click="handleFilterChange(filter.key, filter.unreadOnly, filter.daysUntilExpiry)"
                >
                  {{ filter.label }}
                </div>
              </div>
            </a-card>
          </div>

          <!-- Right Content - Table -->
          <div class="notifications-table-wrapper">
            <a-card :bordered="false" class="table-card">
              <template #extra>
                <a-button 
                  v-if="hasUnread" 
                  type="primary" 
                  @click="handleMarkAllAsRead"
                  :loading="markingAllAsRead"
                >
                  <IconifyIcon icon="mdi:check-all" width="16" style="margin-right: 8px;" />
                  Tandai Semua Ditindak Lanjuti
                </a-button>
              </template>
              
              <!-- Search and Actions -->
              <div class="table-filters-container">
                <a-input 
                  v-model:value="searchText" 
                  placeholder="Cari notifikasi..." 
                  class="search-input"
                  allow-clear
                  @pressEnter="handleSearch"
                  @clear="handleSearch"
                >
                  <template #prefix>
                    <IconifyIcon icon="mdi:magnify" width="16" />
                  </template>
                </a-input>
                <a-button 
                  type="default" 
                  danger
                  @click="handleDeleteAll"
                  :loading="deletingAll"
                >
                  <IconifyIcon icon="mdi:delete-sweep" width="16" style="margin-right: 8px;" />
                  Hapus Semua
                </a-button>
              </div>
              
              <a-table
            :columns="columns"
            :data-source="notifications"
            :loading="loading"
            :pagination="pagination"
            @change="handleTableChange"
            :row-class-name="getRowClassName"
            row-key="id"
            :scroll="{ x: 'max-content' }"
            class="striped-table"
            :locale="{ emptyText: 'Tidak ada notifikasi' }"
            :customRow="(record: Notification) => ({ onClick: () => handleRowClick(record) })"
          >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              <a-tag :color="getTypeColor(record.type)">
                {{ getTypeLabel(record.type) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'status'">
              <div v-if="!record.is_read" @click.stop>
                <a-button 
                  type="primary" 
                  size="small"
                  @click.stop="handleMarkAsResolved(record)"
                  :loading="markingResolvedIds.has(record.id)"
                >
                  <IconifyIcon icon="mdi:check-circle" width="16" style="margin-right: 4px;" />
                  Tandai sudah ditindak lanjuti
                </a-button>
              </div>
              <a-tag v-else color="default">
                Sudah ditindak lanjuti
              </a-tag>
            </template>
            <template v-else-if="column.key === 'created_at'">
              <span>{{ formatTime(record.created_at) }}</span>
            </template>
            <template v-else-if="column.key === 'action'">
              <div class="action-cell" @click.stop="handleRowClick(record, $event)">
                <IconifyIcon 
                  icon="mdi:chevron-right" 
                  width="20" 
                  style="cursor: pointer; color: #1890ff;"
                />
              </div>
            </template>
          </template>
              </a-table>
            </a-card>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Icon as IconifyIcon } from '@iconify/vue'
import { Modal, message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { notificationApi, type Notification, type NotificationFilters } from '../api/notifications'
import { useAuthStore } from '../stores/auth'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import type { TableColumnsType, TableProps } from 'ant-design-vue'

dayjs.extend(relativeTime)

const router = useRouter()
const authStore = useAuthStore()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

// Filters
const filters = [
  { key: 'all', label: 'Tampilkan semua', unreadOnly: undefined, daysUntilExpiry: undefined },
  { key: 'unread', label: 'Belum ditindak lanjuti', unreadOnly: true, daysUntilExpiry: undefined },
  { key: 'read', label: 'Sudah ditindak lanjuti', unreadOnly: false, daysUntilExpiry: undefined },
  { key: 'expiry_3', label: 'Kurang dari 3 Hari Expired', unreadOnly: undefined, daysUntilExpiry: 3 },
  { key: 'expiry_7', label: 'Kurang dari 1 Minggu Expired', unreadOnly: undefined, daysUntilExpiry: 7 },
  { key: 'expiry_30', label: 'Kurang dari 1 Bulan Expired', unreadOnly: undefined, daysUntilExpiry: 30 },
]

const activeFilter = ref('all')
const currentUnreadOnly = ref<boolean | undefined>(undefined)
const currentDaysUntilExpiry = ref<number | undefined>(undefined)

// Data
const notifications = ref<Notification[]>([])
const loading = ref(false)
const markingAllAsRead = ref(false)
const deletingAll = ref(false)
const markingResolvedIds = ref<Set<string>>(new Set())
const searchText = ref('')
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// Table columns
const columns: TableColumnsType = [
  {
    title: 'Judul',
    key: 'title',
    dataIndex: 'title',
    width: '300px',
    sorter: (a: Notification, b: Notification) => a.title.localeCompare(b.title),
    sortDirections: ['ascend', 'descend'],
  },
  {
    title: 'Pesan',
    key: 'message',
    dataIndex: 'message',
    width: '200px',
    ellipsis: true,
  },
  {
    title: 'Tipe',
    key: 'type',
    dataIndex: 'type',
    width: '100px',
    filters: [
      { text: 'Info', value: 'info' },
      { text: 'Success', value: 'success' },
      { text: 'Warning', value: 'warning' },
      { text: 'Error', value: 'error' },
    ],
    onFilter: (value: string | number | boolean, record: Notification) => {
      return String(record.type).toLowerCase() === String(value).toLowerCase()
    },
  },
  {
    title: 'Waktu',
    key: 'created_at',
    dataIndex: 'created_at',
    width: '10%',
    sorter: (a: Notification, b: Notification) => {
      return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    },
    sortDirections: ['ascend', 'descend'],
  },
  {
    title: '',
    key: 'action',
    width: '5%',
  },
  {
    title: 'Status',
    key: 'status',
    width: 200, // Fixed width untuk button "Tandai sudah ditindak lanjuti"
    fixed: 'right' as const, // Sticky di kanan
    filters: [
      { text: 'Belum ditindak lanjuti', value: false },
      { text: 'Sudah ditindak lanjuti', value: true },
    ],
    onFilter: (value: string | number | boolean, record: Notification) => {
      return record.is_read === Boolean(value)
    },
  },
]

// Computed
const hasUnread = computed(() => {
  return notifications.value.some(n => !n.is_read)
})

const unreadCount = ref(0)

const pagination = computed(() => ({
  current: currentPage.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (total: number) => `Total ${total} notifikasi`,
}))

// Methods
const handleFilterChange = (key: string, unreadOnly?: boolean, daysUntilExpiry?: number) => {
  activeFilter.value = key
  currentUnreadOnly.value = unreadOnly
  currentDaysUntilExpiry.value = daysUntilExpiry
  currentPage.value = 1
  loadNotifications()
}

const loadNotifications = async () => {
  loading.value = true
  try {
    const filters: NotificationFilters = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    
    // Filter by read status
    if (currentUnreadOnly.value !== undefined) {
      filters.unread_only = currentUnreadOnly.value
    }
    
    // Filter by expiry date
    if (currentDaysUntilExpiry.value !== undefined) {
      filters.days_until_expiry = currentDaysUntilExpiry.value
    }
    
    const response = await notificationApi.getNotificationsInbox(filters)
    
    // Apply search filter (client-side)
    let filteredData = response.data
    if (searchText.value.trim()) {
      const searchLower = searchText.value.toLowerCase()
      filteredData = filteredData.filter(notif => 
        notif.title.toLowerCase().includes(searchLower) ||
        notif.message.toLowerCase().includes(searchLower) ||
        notif.type.toLowerCase().includes(searchLower)
      )
    }
    
    notifications.value = filteredData
    total.value = searchText.value.trim() ? filteredData.length : response.total
  } catch (error) {
    console.error('Failed to load notifications:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadNotifications()
}

const handleTableChange: TableProps['onChange'] = (pag) => {
  if (pag) {
    currentPage.value = pag.current || 1
    pageSize.value = pag.pageSize || 10
  }
  loadNotifications()
}

// Handle mark as resolved (sudah ditindak lanjuti)
const handleMarkAsResolved = async (notification: Notification) => {
  if (notification.is_read) {
    return // Sudah ditindak lanjuti
  }
  
  markingResolvedIds.value.add(notification.id)
  
  try {
    await notificationApi.markAsRead(notification.id)
    
    // Update local state
    notification.is_read = true
    unreadCount.value = Math.max(0, (unreadCount.value || 0) - 1)
    
    // Update di array notifications juga
    const index = notifications.value.findIndex(n => n.id === notification.id)
    if (index !== -1 && notifications.value[index]) {
      notifications.value[index].is_read = true
    }
    
    message.success('Notifikasi telah ditandai sebagai sudah ditindak lanjuti')
    
    // Reload notifications untuk memastikan state ter-update
    await loadNotifications()
    
    // Trigger event untuk refresh notifications di header
    window.dispatchEvent(new CustomEvent('notification-read', { 
      detail: { notificationId: notification.id } 
    }))
  } catch (error) {
    console.error('❌ [NotificationsView] Failed to mark notification as resolved:', error)
    message.error('Gagal menandai notifikasi sebagai sudah ditindak lanjuti')
  } finally {
    markingResolvedIds.value.delete(notification.id)
  }
}

const handleRowClick = async (notification: Notification, event?: Event) => {
  console.log('👆 [NotificationsView] Row clicked:', notification.id)
  
  // Prevent navigation jika klik di action column
  if (event && (event.target as HTMLElement).closest('.action-cell')) {
    console.log('🚫 [NotificationsView] Click on action cell, ignoring')
    return
  }
  
  // Navigate to resource if available (TIDAK mark as read otomatis)
  if (notification.resource_type === 'document' && notification.resource_id) {
    console.log('📄 [NotificationsView] Navigating to document:', notification.resource_id)
    router.push(`/documents/${notification.resource_id}`)
  } else {
    console.log('📋 [NotificationsView] No resource to navigate to')
  }
}

const handleMarkAllAsRead = async () => {
  markingAllAsRead.value = true
  try {
    await notificationApi.markAllAsRead()
    message.success('Semua notifikasi telah ditandai sebagai sudah ditindak lanjuti')
    await loadNotifications()
  } catch (error) {
    console.error('Failed to mark all as read:', error)
  } finally {
    markingAllAsRead.value = false
  }
}

const handleDeleteAll = async () => {
  const userRole = authStore.user?.role?.toLowerCase() || ''
  let confirmMessage = 'Apakah Anda yakin ingin menghapus semua notifikasi? Tindakan ini tidak dapat dibatalkan.'
  
  // Sesuaikan pesan berdasarkan RBAC
  if (userRole === 'superadmin' || userRole === 'administrator') {
    confirmMessage = 'Apakah Anda yakin ingin menghapus SEMUA notifikasi dari SEMUA user? Tindakan ini tidak dapat dibatalkan.'
  } else if (userRole === 'admin') {
    confirmMessage = 'Apakah Anda yakin ingin menghapus semua notifikasi dari company Anda dan semua anak perusahaan? Tindakan ini tidak dapat dibatalkan.'
  } else {
    confirmMessage = 'Apakah Anda yakin ingin menghapus semua notifikasi Anda? Tindakan ini tidak dapat dibatalkan.'
  }
  
  Modal.confirm({
    title: 'Hapus Semua Notifikasi?',
    content: confirmMessage,
    okText: 'Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      deletingAll.value = true
      try {
        await notificationApi.deleteAll()
        await loadNotifications()
        message.success('Semua notifikasi berhasil dihapus')
      } catch (error) {
        console.error('Failed to delete all notifications:', error)
        message.error('Gagal menghapus notifikasi')
      } finally {
        deletingAll.value = false
      }
    },
  })
}

const formatTime = (date: string) => {
  const d = dayjs(date)
  const now = dayjs()
  const diffMinutes = now.diff(d, 'minute')
  const diffHours = now.diff(d, 'hour')
  const diffDays = now.diff(d, 'day')
  
  // Format yang lebih mudah dibaca
  if (diffMinutes < 1) {
    return 'Baru saja'
  } else if (diffMinutes < 60) {
    return `${diffMinutes} menit yang lalu`
  } else if (diffHours < 24) {
    return `${diffHours} jam yang lalu`
  } else if (diffDays < 7) {
    return `${diffDays} hari yang lalu`
  } else {
    // Untuk lebih dari 7 hari, tampilkan tanggal lengkap
    return d.format('DD MMM YYYY, HH:mm')
  }
}

const getTypeColor = (type?: string) => {
  switch (type?.toLowerCase()) {
    case 'success':
      return 'green'
    case 'warning':
      return 'orange'
    case 'error':
      return 'red'
    case 'info':
    default:
      return 'blue'
  }
}

const getTypeLabel = (type?: string) => {
  switch (type?.toLowerCase()) {
    case 'success':
      return 'Success'
    case 'warning':
      return 'Warning'
    case 'error':
      return 'Error'
    case 'info':
    default:
      return 'Info'
  }
}

const getRowClassName = (record: Notification) => {
  return record.is_read ? '' : 'unread-row'
}

const loadUnreadCount = async () => {
  try {
    unreadCount.value = await notificationApi.getUnreadCount()
  } catch (error) {
    console.error('Failed to load unread count:', error)
  }
}

onMounted(() => {
  loadNotifications()
  loadUnreadCount()
})
</script>

<style lang="scss" scoped>
.notifications-layout {
  min-height: 100vh;
  background: #f0f2f5;
}

.notifications-content {
  padding: 24px;
}

// Page Header
.page-header-container {
  margin-bottom: 24px;
}

.page-header {
  padding: 24px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

  .header-left {
    .page-title {
      margin: 0 0 8px 0;
      font-size: 24px;
      font-weight: 600;
      color: #333;
    }

    .page-description {
      margin: 0;
      font-size: 14px;
      color: #666;
    }
  }
}

// Main Content
.mainContentPage {
  .notifications-container {
    display: flex;
    gap: 24px;
  }

  // Left Sidebar - Filter
  .notifications-sidebar {
    width: 280px;
    flex-shrink: 0;

    .filter-card {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      border-radius: 8px;

      :deep(.ant-card-body) {
        padding: 16px;
      }
    }

    .sidebar-filters {
      .filter-item {
        padding: 12px 16px;
        margin-bottom: 8px;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.2s;
        color: #666;
        font-size: 14px;
        background: #fafafa;
        
        &:hover {
          background-color: #f0f0f0;
          color: #1890ff;
        }
        
        &.active {
          background-color: #e6f7ff;
          color: #1890ff;
          font-weight: 500;
          border-left: 3px solid #1890ff;
        }
      }
    }
  }

  // Right Content - Table
  .notifications-table-wrapper {
    flex: 1;
    min-width: 0; // Prevent flex item from overflowing

    .table-card {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      border-radius: 8px;

      :deep(.ant-card-body) {
        padding: 24px;
      }
    }
  }
}

// Table filters
.table-filters-container {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  align-items: center;

  .search-input {
    flex: 1;
    max-width: 400px;
  }
}

.action-cell {
  display: inline-block;
}

:deep(.unread-row) {
  background-color: #e6f7ff !important;
  
  &:hover {
    background-color: #bae7ff !important;
  }
}

:deep(.ant-table-tbody > tr) {
  cursor: pointer;
  
  &:hover {
    background-color: #f5f5f5;
  }
}

// Sticky column untuk Status (paling akhir)
:deep(.ant-table) {
  .ant-table-thead > tr > th:last-child,
  .ant-table-tbody > tr > td:last-child {
    position: sticky;
    right: 0;
    z-index: 10;
    background-color: #fff;
    box-shadow: -2px 0 4px rgba(0, 0, 0, 0.1);
  }
  
  .ant-table-thead > tr > th:last-child {
    z-index: 11; // Header lebih tinggi dari body
  }
  
  // Hover state untuk sticky column
  .ant-table-tbody > tr:hover > td:last-child {
    background-color: #f5f5f5;
  }
  
  // Unread row sticky column background
  .ant-table-tbody > tr.unread-row > td:last-child {
    background-color: #e6f7ff;
    
    &:hover {
      background-color: #bae7ff;
    }
  }
}
</style>


