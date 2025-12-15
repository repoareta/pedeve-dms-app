<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import axios from 'axios'
import { Icon as IconifyIcon } from '@iconify/vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import DocumentSidebarActivityCard from '../components/DocumentSidebarActivityCard.vue'
import documentsApi, { type DocumentFolder, type DocumentItem } from '../api/documents'
import { auditApi, type UserActivityLog } from '../api/audit'
import { userApi, type User } from '../api/userManagement'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const folderId = computed(() => (route.params.id as string) || '')

const folders = ref<DocumentFolder[]>([])
const files = ref<DocumentItem[]>([])
const loading = ref(false)
const pageLoading = ref(true)
const userMap = ref<Record<string, string>>({})

// Subfolder state
const subfolders = ref<DocumentFolder[]>([])
const subfolderModalVisible = ref(false)
const subfolderName = ref('')
const creatingSubfolder = ref(false)

// Rename folder state
const renameFolderModalVisible = ref(false)
const renameFolderName = ref('')
const folderBeingEdited = ref<DocumentFolder | null>(null)

// Auth state for permission check
const userRole = ref<string>('')
const isSuperAdminOrAdministratorSync = computed(() => {
  const role = userRole.value
  return role === 'superadmin' || role === 'administrator'
})

// Upload state
type UploadItem = {
  uid: string
  name: string
  status?: string
  originFileObj?: File
  size?: number
}
const uploadList = ref<UploadItem[]>([])
const uploading = ref(false)
// Fiber default body limit ~4MB, naikkan guard ke 5MB (sesuaikan dengan backend/proxy)
const MAX_FILE_SIZE = 5 * 1024 * 1024 // 5MB
const searchText = ref('')
const tablePagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['5', '10', '20', '50'],
})

// Activity
const activities = ref<UserActivityLog[]>([])
const activityLoading = ref(false)

const currentFolder = computed(() => {
  return folders.value.find(f => f.id === folderId.value)
})

// Build breadcrumb path from current folder to root
const breadcrumbPath = computed(() => {
  const path: Array<{ id: string; name: string }> = []
  if (!folderId.value || folders.value.length === 0) {
    return path
  }

  // Start from current folder and traverse up to root
  const findFolderById = (id: string): DocumentFolder | undefined => {
    return folders.value.find(f => f.id === id)
  }

  let currentFolderId: string | null = folderId.value
  const visited = new Set<string>() // Prevent infinite loops

  while (currentFolderId && !visited.has(currentFolderId)) {
    visited.add(currentFolderId)
    const folder = findFolderById(currentFolderId)
    if (folder) {
      path.unshift({ id: folder.id, name: folder.name })
      currentFolderId = folder.parent_id || null
    } else {
      break
    }
  }

  return path
})

const formatBytes = (bytes: number): string => {
  if (bytes <= 0) return '0B'
  const kb = bytes / 1024
  if (kb < 1) return `${bytes}B`
  const mb = kb / 1024
  if (mb < 1) return `${kb.toFixed(0)}KB`
  const gb = mb / 1024
  if (gb < 1) return `${mb.toFixed(0)}MB`
  return `${gb.toFixed(1)}GB`
}

const totalFolderSize = computed(() => {
  const total = files.value.reduce((sum, f) => sum + (f.size || 0), 0)
  return formatBytes(total)
})

const formatDateTime = (dateString: string | undefined): string => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  const year = date.getFullYear()
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const day = date.getDate().toString().padStart(2, '0')
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${year}-${month}-${day} : ${hours}:${minutes}`
}

const filteredFiles = computed(() => {
  const q = searchText.value.trim().toLowerCase()
  const data = !q
    ? files.value
    : files.value.filter(f =>
        (f.name || f.file_name || '').toLowerCase().includes(q) ||
        (f.mime_type || '').toLowerCase().includes(q) ||
        (f.uploader_id || '').toLowerCase().includes(q)
      )
  return data
})

const hasCompleteMetadata = (file: DocumentItem): boolean => {
  const meta = getMeta(file)
  const required = ['doc_type', 'reference', 'issued_date', 'effective_date', 'expired_date', 'is_active']
  return required.every((key) => Boolean(meta[key]))
}

const loadUsers = async () => {
  try {
    const users: User[] = await userApi.getAll()
    const map: Record<string, string> = {}
    users.forEach((u) => {
      map[u.id] = u.username || u.email || u.id
    })
    userMap.value = map
  } catch (error) {
    console.warn('Gagal memuat data pengguna untuk uploader name:', error)
  }
}

const getMeta = (record: DocumentItem): Record<string, unknown> => {
  return (record.metadata as Record<string, unknown> | undefined) || {}
}

const getUploaderName = (record: DocumentItem): string => {
  const meta = getMeta(record)
  return (
    (meta['uploaded_by'] as string) ||
    (meta['uploader'] as string) ||
    (record as unknown as { uploader_name?: string }).uploader_name ||
    (record.uploader_id ? userMap.value[record.uploader_id] : undefined) ||
    record.uploader_id ||
    '-'
  )
}

const loadFolders = async () => {
  try {
    folders.value = await documentsApi.listFolders()
    // Load subfolders after folders are loaded
    await loadSubfolders()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat folder')
  }
}

const loadSubfolders = async () => {
  if (!folderId.value) {
    subfolders.value = []
    return
  }
  try {
    // Filter folders where parent_id matches current folderId
    subfolders.value = folders.value.filter(f => f.parent_id === folderId.value)
    // Load files for all subfolders to calculate stats
    // Note: This loads all files, which might be inefficient for many subfolders
    // Consider optimizing by loading summary stats per folder if available
  } catch (error: unknown) {
    console.error('Failed to load subfolders:', error)
    subfolders.value = []
  }
}

// Load files for all subfolders to get accurate file counts and sizes
const allFiles = ref<DocumentItem[]>([])
const loadAllFilesForStats = async () => {
  try {
    // Load all files without folder filter to get stats for subfolders
    allFiles.value = await documentsApi.listDocuments()
  } catch (error: unknown) {
    console.error('Failed to load all files for stats:', error)
    allFiles.value = []
  }
}

const loadDocuments = async () => {
  if (!folderId.value) {
    files.value = []
    tablePagination.value.total = 0
    return
  }

  loading.value = true
  try {
    // Ensure we only load documents for the current folder_id
    console.log('Loading documents for folder_id:', folderId.value)
    files.value = await documentsApi.listDocuments({ folder_id: folderId.value })
    console.log('Loaded documents:', files.value.length, 'files for folder', folderId.value)
    tablePagination.value.total = files.value.length
  } catch (error: unknown) {
    const err = error as { message?: string }
    console.error('Error loading documents:', error)
    message.error(err.message || 'Gagal memuat dokumen')
    files.value = []
    tablePagination.value.total = 0
  } finally {
    loading.value = false
  }
}

const handleUploadChange = ({ fileList }: { fileList: UploadItem[] }) => {
  uploadList.value = fileList
}

const handleBatchUpload = async () => {
  if (!uploadList.value.length) {
    message.warning('Pilih file terlebih dahulu')
    return
  }
  const tooBig = uploadList.value.find(item => (item.originFileObj?.size || 0) > MAX_FILE_SIZE)
  if (tooBig) {
    message.error(`File ${tooBig.name} terlalu besar, batas maksimal 5MB`)
    return
  }
  uploading.value = true
  try {
    for (const item of uploadList.value) {
      const file = item.originFileObj as File | undefined
      if (!file) continue
      await documentsApi.uploadDocument({
        file,
        folder_id: folderId.value,
        title: file.name,
        status: 'active',
      })
    }
    message.success('File berhasil diupload')
    uploadList.value = []
    await loadDocuments()
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 413) {
      message.error('File terlalu besar, server menolak upload (413). Silakan pilih file yang lebih kecil atau hubungi admin.')
    } else {
      const err = error as { message?: string }
      message.error(err.message || 'Gagal upload file')
    }
  } finally {
    uploading.value = false
  }
}

const loadActivities = async () => {
  activityLoading.value = true
  try {
    const res = await auditApi.getUserActivityLogs({
      page: 1,
      pageSize: 5,
      resource: 'document',
    })
    activities.value = res.data
  } catch (error) {
    console.error('Failed to load activities', error)
    activities.value = []
  } finally {
    activityLoading.value = false
  }
}

const simplifyMime = (mime?: string): string => {
  if (!mime) return 'File'
  const lower = mime.toLowerCase()
  if (lower.includes('pdf')) return 'pdf'
  if (lower.includes('word') || lower.includes('msword')) return 'doc'
  if (lower.includes('spreadsheet') || lower.includes('excel')) return 'xlsx'
  if (lower.includes('ppt')) return 'ppt'
  if (lower.includes('image')) return 'image'
  return lower.split('/').pop() || 'file'
}

const getFileTypeLabel = (mimeType: string | undefined): string => {
  if (!mimeType) return 'File'
  const lower = mimeType.toLowerCase()
  if (lower.includes('pdf')) return 'Pdf'
  if (lower.includes('excel') || lower.includes('spreadsheet')) return 'Excel'
  if (lower.includes('word') || lower.includes('document')) return 'Word'
  if (lower.includes('image')) return 'Image'
  return 'File'
}

const getFileTypeColor = (mimeType: string | undefined): string => {
  if (!mimeType) return 'default'
  const lower = mimeType.toLowerCase()
  if (lower.includes('pdf')) return 'red'
  if (lower.includes('excel') || lower.includes('spreadsheet')) return 'green'
  if (lower.includes('word') || lower.includes('document')) return 'blue'
  if (lower.includes('image')) return 'orange'
  return 'default'
}

const getDisplayName = (username: string): string => {
  if (!username) return ''
  const parts = username.trim().split(/\s+/)
  return parts[0] || username
}

const getActivityDescription = (activity: UserActivityLog): string => {
  const action = activity.action.toLowerCase()
  let fileName = ''
  try {
    const details = activity.details ? JSON.parse(activity.details) : {}
    fileName = details.title || details.name || details.file_name || ''
  } catch {
    fileName = ''
  }
  const target = fileName || 'dokumen'
  if (action.includes('update') || action.includes('edit')) return `Telah mengupdate file ${target}`
  if (action.includes('create') || action.includes('upload')) return `Telah mengunggah file ${target}`
  if (action.includes('delete')) return `Telah menghapus file ${target}`
  if (action.includes('view')) return `Telah melihat file ${target}`
  return `Telah melakukan aksi ${action} pada ${target}`
}

const formatTime = (timestamp: string): string => {
  const diff = dayjs().diff(dayjs(timestamp), 'second')
  if (diff < 60) return `${diff} sec`
  if (diff < 3600) return `${Math.floor(diff / 60)} min`
  if (diff < 86400) return `${Math.floor(diff / 3600)} hour`
  return `${Math.floor(diff / 86400)} day`
}

const getUserAvatarColor = (username: string): string => {
  const colors = ['#52c41a', '#1890ff', '#faad14', '#f5222d', '#722ed1', '#eb2f96']
  const safeName = username || ''
  if (!safeName) return '#52c41a'
  return (colors[safeName.charCodeAt(0) % colors.length] ?? '#52c41a') as string
}

const handleRefresh = async () => {
  pageLoading.value = true
  await Promise.all([loadFolders(), loadDocuments(), loadActivities(), loadAllFilesForStats()])
  pageLoading.value = false
}

const handleAddSubfolder = () => {
  subfolderModalVisible.value = true
  subfolderName.value = ''
}

const handleCreateSubfolder = async () => {
  if (!subfolderName.value.trim()) {
    message.warning('Nama subfolder tidak boleh kosong')
    return
  }

  if (!folderId.value) {
    message.error('Folder ID tidak ditemukan')
    return
  }

  creatingSubfolder.value = true
  try {
    await documentsApi.createFolder(subfolderName.value.trim(), folderId.value)
    message.success('Subfolder berhasil dibuat')
    subfolderModalVisible.value = false
    subfolderName.value = ''
    // Reload folders to get the new subfolder and reload files for stats
    await loadFolders()
    await loadAllFilesForStats()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal membuat subfolder')
  } finally {
    creatingSubfolder.value = false
  }
}

// Removed handleSubfolderClick - using handleSubfolderDoubleClick instead

const handleSubfolderDoubleClick = (subfolder: DocumentFolder) => {
  // Add transition effect
  pageLoading.value = true
  setTimeout(() => {
    router.push(`/documents/folders/${subfolder.id}`)
  }, 150) // Short delay for transition effect
}

const handleBreadcrumbClick = (folderId: string) => {
  pageLoading.value = true
  setTimeout(() => {
    router.push(`/documents/folders/${folderId}`)
  }, 150)
}

// Helper functions for folder stats
const getSubfolderFileCount = (subfolderId: string): number => {
  // Use allFiles for more accurate count (includes files from all folders)
  return allFiles.value.filter(f => f.folder_id === subfolderId).length
}

const getSubfolderSize = (subfolderId: string): string => {
  // Use allFiles for more accurate size (includes files from all folders)
  const subfolderFiles = allFiles.value.filter(f => f.folder_id === subfolderId)
  const totalSize = subfolderFiles.reduce((sum, f) => sum + (f.size || 0), 0)
  return formatBytes(totalSize)
}

// Rename folder functions
const openRenameModal = (folder: DocumentFolder) => {
  folderBeingEdited.value = folder
  renameFolderName.value = folder.name
  renameFolderModalVisible.value = true
}

const handleRenameFolder = async () => {
  if (!folderBeingEdited.value) return
  if (!renameFolderName.value.trim()) {
    message.warning('Nama folder wajib diisi')
    return
  }
  try {
    const updated = await documentsApi.renameFolder(folderBeingEdited.value.id, renameFolderName.value.trim())
    // Update in folders array
    folders.value = folders.value.map(f => f.id === updated.id ? updated : f)
    // Update in subfolders if it's a subfolder
    subfolders.value = subfolders.value.map(f => f.id === updated.id ? updated : f)
    message.success('Folder berhasil diubah')
    renameFolderModalVisible.value = false
    folderBeingEdited.value = null
    await loadFolders() // Reload to refresh breadcrumb
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal mengganti nama folder')
  }
}

// Delete folder functions
const handleDeleteFolder = (folder: DocumentFolder) => {
  Modal.confirm({
    title: 'Hapus folder?',
    content: `Menghapus folder "${folder.name}" akan menghapus seluruh file di dalamnya. Tindakan ini tidak dapat dibatalkan. Lanjutkan?`,
    okText: 'Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      try {
        await documentsApi.deleteFolder(folder.id)
        // Remove from subfolders
        subfolders.value = subfolders.value.filter(f => f.id !== folder.id)
        // Reload folders to update list
        await loadFolders()
        // If deleted folder is current folder, navigate to parent or root
        if (folder.id === folderId.value) {
          const parentFolder = folders.value.find(f => f.id === (currentFolder.value?.parent_id || ''))
          if (parentFolder) {
            router.push(`/documents/folders/${parentFolder.id}`)
          } else {
            router.push('/documents')
          }
        } else {
          // Reload documents and stats to refresh the list
          await loadDocuments()
          await loadAllFilesForStats()
        }
        message.success('Folder dan file di dalamnya telah dihapus')
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || 'Gagal menghapus folder')
      }
    },
  })
}

const handleTableChange = (pag: { current?: number; pageSize?: number }) => {
  if (pag.current) tablePagination.value.current = pag.current
  if (pag.pageSize) {
    tablePagination.value.pageSize = pag.pageSize
    tablePagination.value.current = 1
  }
}

const handleSeeAll = () => {
  message.info('Belum ada fitur See All activity')
}

const handleAddFolderClick = () => router.push('/documents')
const handleUploadClick = () => router.push('/documents/upload')

const handleLogout = async () => {
  const { useAuthStore } = await import('../stores/auth')
  const authStore = useAuthStore()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  } finally {
    router.push('/login')
  }
}

// Watch for route changes to reload data when navigating between folders
watch(
  () => folderId.value,
  async (newFolderId, oldFolderId) => {
    if (newFolderId !== oldFolderId) {
      // Reload data when folder ID changes (including when going to subfolder)
      console.log('Folder ID changed from', oldFolderId, 'to', newFolderId)
      pageLoading.value = true
      try {
        // Load folders first, then load documents for the new folder_id
        await loadFolders()
        await loadDocuments()
        await loadActivities()
        await loadAllFilesForStats() // Reload stats when folder changes
      } catch (error) {
        console.error('Failed to reload folder detail page data:', error)
        message.error('Gagal memuat data. Silakan coba lagi.')
      } finally {
        pageLoading.value = false
      }
    }
  },
  { immediate: false }
)

onMounted(async () => {
  // Load user role for permission check
  try {
    const { useAuthStore } = await import('../stores/auth')
    const authStore = useAuthStore()
    userRole.value = authStore.user?.role?.toLowerCase() || ''
  } catch (error) {
    console.error('Failed to load auth store:', error)
  }

  pageLoading.value = true
  try {
    await Promise.all([
      loadUsers(),
      loadFolders(),
      loadDocuments(),
      loadActivities(),
      loadAllFilesForStats() // Load all files for subfolder stats
    ])
  } catch (error) {
    console.error('Failed to load folder detail page data:', error)
    message.error('Gagal memuat data. Silakan coba lagi.')
  } finally {
    pageLoading.value = false
  }
})
</script>

<template>
  <div class="folder-detail-layout">
    <DashboardHeader @logout="handleLogout" />
    <div class="folder-detail-content">
      <a-row :gutter="[16, 16]">
        <a-col :xs="24" :lg="5" :xl="5" class="left-column1">
          <DocumentSidebarActivityCard
            :activities="activities"
            :activity-loading="activityLoading"
            :page-loading="pageLoading"
            :hide-search="true"
            @search="() => message.info('Pencarian tersedia di halaman Documents')"
            @refresh="handleRefresh"
            @add-folder="handleAddFolderClick"
            @upload-file="handleUploadClick"
            @nav-dashboard="router.push('/documents')"
            @nav-recent="message.info('Belum ada data recent')"
            @nav-trash="message.info('Belum ada data trash')"
            @see-all="handleSeeAll"
            :get-display-name="getDisplayName"
            :get-activity-description="getActivityDescription"
            :format-time="formatTime"
            :get-user-avatar-color="getUserAvatarColor"
          />
        </a-col>

        <a-col :xs="24" :lg="15" :xl="15" class="center-column1">
          <a-card class="upload-card" :bordered="false">
            <div class="card-header">
              <div>
                <div class="breadcrumb-container">
                  <span class="breadcrumb-item" @click="router.push('/documents')">Folders</span>
                  <template v-for="(item, index) in breadcrumbPath" :key="item.id">
                    <span class="breadcrumb-separator">/</span>
                    <span
                      class="breadcrumb-item"
                      :class="{ 'breadcrumb-active': index === breadcrumbPath.length - 1 }"
                      @click="index < breadcrumbPath.length - 1 ? handleBreadcrumbClick(item.id) : null"
                    >
                      {{ item.name }}
                    </span>
                  </template>
                </div>
                <p>Upload files</p>
              </div>
              <a-button type="primary" @click="handleAddSubfolder">
                <IconifyIcon icon="mdi:folder-plus" width="16" style="margin-right: 8px;" />
                Add Subfolder
              </a-button>
            </div>
            <a-upload-dragger
              multiple
              name="file"
              :before-upload="() => false"
              :file-list="uploadList"
              @change="handleUploadChange"
              class="upload-dragger"
            >
              <p class="ant-upload-drag-icon">
                <IconifyIcon icon="mdi:cloud-upload-outline" width="40" />
              </p>
              <p class="ant-upload-text">Browse file to upload</p>
              <p class="ant-upload-hint">Multiple files supported</p>
            </a-upload-dragger>
            <div class="upload-actions">
              <a-button type="primary" :loading="uploading" @click="handleBatchUpload">Upload</a-button>
              <a-button @click="uploadList = []" :disabled="!uploadList.length">Clear</a-button>
            </div>

            <div v-if="uploadList.length" class="upload-preview">
              <div class="upload-item" v-for="item in uploadList" :key="item.uid">
                <div class="upload-item-name">{{ item.name }}</div>
                <div class="upload-item-size">{{ formatBytes(item.size || item.originFileObj?.size || 0) }}</div>
              </div>
            </div>

            <!-- Subfolders List -->
            <div v-if="subfolders.length > 0" class="subfolders-section">
              <a-divider style="margin: 16px 0;" />
              <div class="subfolders-header">
                <h4 style="margin: 0;">Subfolders ({{ subfolders.length }})</h4>
              </div>
              <div class="subfolders-grid">
                <div
                  v-for="subfolder in subfolders"
                  :key="subfolder.id"
                  class="folder-card"
                  @dblclick.stop="handleSubfolderDoubleClick(subfolder)"
                  title="Double click to open"
                >
                  <div class="folder-card-header">
                    <div class="folder-icon">
                      <IconifyIcon icon="mdi:folder" width="50" color="#E7EAE9" />
                    </div>
                    <a-dropdown :trigger="['click']">
                      <IconifyIcon icon="mdi:dots-vertical" width="20" class="folder-menu-icon" @click.stop />
                      <template #overlay>
                        <a-menu>
                          <a-menu-item key="rename" @click="openRenameModal(subfolder)">Rename</a-menu-item>
                          <a-menu-item v-if="isSuperAdminOrAdministratorSync" key="delete" @click="handleDeleteFolder(subfolder)">Delete</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                  </div>
                  <div class="folder-name">{{ subfolder.name }}</div>
                  <div class="folder-meta">
                    <span>{{ getSubfolderFileCount(subfolder.id) }} Files</span>
                    <span>{{ getSubfolderSize(subfolder.id) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </a-card>

          <!-- Create Subfolder Modal -->
          <a-modal
            v-model:open="subfolderModalVisible"
            title="Create Subfolder"
            :confirm-loading="creatingSubfolder"
            @ok="handleCreateSubfolder"
            @cancel="() => { subfolderModalVisible = false; subfolderName = '' }"
          >
            <a-form-item label="Subfolder Name" required>
              <a-input
                v-model:value="subfolderName"
                placeholder="Enter subfolder name"
                :maxlength="100"
                @pressEnter="handleCreateSubfolder"
              />
            </a-form-item>
          </a-modal>

          <!-- Rename Folder Modal -->
          <a-modal
            v-model:open="renameFolderModalVisible"
            title="Rename Folder"
            ok-text="Simpan"
            cancel-text="Batal"
            @ok="handleRenameFolder"
            @cancel="() => { renameFolderModalVisible = false; folderBeingEdited = null; renameFolderName = '' }"
          >
            <a-input
              v-model:value="renameFolderName"
              placeholder="Nama folder baru"
              @pressEnter="handleRenameFolder"
            />
          </a-modal>

          <a-card class="table-card" :bordered="false">
            <div class="table-header">
              <div class="title">Recently Files</div>
              <a-input
                v-model:value="searchText"
                placeholder="Cari file..."
                style="max-width: 240px"
                allow-clear
              />
            </div>
            <a-table
            :data-source="filteredFiles"
            :loading="loading || pageLoading"
            :pagination="tablePagination"
            row-key="id"
            size="small"
            @change="handleTableChange"
            :custom-row="(record: DocumentItem) => ({
              onClick: () => router.push(`/documents/${record.id}`),
              style: { cursor: 'pointer' }
            })"
            >
              <a-table-column
                title="Name"
                key="name"
                :sorter="(a: DocumentItem, b: DocumentItem) => (a.name || a.file_name || '').localeCompare(b.name || b.file_name || '')"
              >
                <template #default="{ record }">
                  <div class="file-name-cell">
                    <IconifyIcon icon="mdi:file-document-outline" width="18" style="margin-right: 8px; color: #666;" />
                    <span>{{ record.name || record.file_name }}</span>
                  </div>
                </template>
              </a-table-column>
              <a-table-column
                title="Type"
                key="type"
                :sorter="(a: DocumentItem, b: DocumentItem) => simplifyMime(a.mime_type).localeCompare(simplifyMime(b.mime_type))"
              >
                <template #default="{ record }">
                  <a-tag :color="getFileTypeColor(record.mime_type)">{{ getFileTypeLabel(record.mime_type) }}</a-tag>
                </template>
              </a-table-column>
              <a-table-column
                title="Uploaded by"
                key="uploader"
                :sorter="(a: DocumentItem, b: DocumentItem) => (a.uploader_id || '').localeCompare(b.uploader_id || '')"
              >
                <template #default="{ record }">
                  {{ getUploaderName(record) }}
                </template>
              </a-table-column>
              <a-table-column
                title="Status & Last modified"
                key="metadata_status"
                :sorter="(a: DocumentItem, b: DocumentItem) => (dayjs(a.updated_at || a.created_at || 0).valueOf() - dayjs(b.updated_at || b.created_at || 0).valueOf())"
              >
                <template #default="{ record }">
                  <div class="metadata-status-cell">
                    <div class="meta-action">
                      <a-button
                        v-if="hasCompleteMetadata(record)"
                        type="default"
                        size="small"
                        style="background: #f5f5f5; border-color: #d9d9d9;"
                        @click.stop="router.push(`/documents/${record.id}`)"
                      >
                        Meta Data ✓
                      </a-button>
                      <a-button
                        v-else
                        type="primary"
                        size="small"
                        style="background: #52c41a; border-color: #52c41a;"
                        @click.stop="router.push(`/documents/${record.id}`)"
                      >
                        Lengkapi Meta Data →
                      </a-button>
                    </div>
                    <div class="meta-date">
                      <IconifyIcon icon="lets-icons:time" width="14" />: {{ formatDateTime(record.updated_at || record.created_at) }}
                    </div>
                  </div>
                </template>
              </a-table-column>
            </a-table>
          </a-card>
        </a-col>

        <a-col :xs="24" :lg="4" :xl="4" class="right-column1" v-if="currentFolder">
          <a-card class="detail-card" :bordered="false">
            <div class="detail-icon">
              <IconifyIcon icon="mdi:folder" width="64" style="color: #035cab;" />
            </div>
            <div class="detail-title">Details</div>
            <div class="detail-row">
              <span>Name</span>
              <strong>{{ currentFolder.name }}</strong>
            </div>
            <div class="detail-row">
              <span>Type</span>
              <strong>File Folder</strong>
            </div>
            <div class="detail-row">
              <span>Size</span>
              <strong>{{ totalFolderSize }}</strong>
            </div>
            <div class="detail-row">
              <span>Location</span>
              <strong>Documents</strong>
            </div>
            <div class="detail-row">
              <span>Last Edited</span>
              <strong>{{ currentFolder.updated_at ? dayjs(currentFolder.updated_at).format('YYYY-MM-DD HH:mm') : '-' }}</strong>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>
  </div>
</template>

<style scoped>
.folder-detail-layout {
  /* min-height: 100vh; */
  /* background: #f7f8fb; */
  margin-top: 70px;
}

.folder-detail-content {
  max-width: 100%;
  margin: 0 auto;
  padding: 16px 24px 40px;
}

.grid-layout {
  display: block;
}

.upload-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  margin-bottom: 12px;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
}

.breadcrumb-container {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.breadcrumb-item {
  color: #1890ff;
  cursor: pointer;
  transition: color 0.2s;
  user-select: none;
}

.breadcrumb-item:hover {
  color: #40a9ff;
  text-decoration: underline;
}

.breadcrumb-item.breadcrumb-active {
  color: #333;
  cursor: default;
}

.breadcrumb-item.breadcrumb-active:hover {
  color: #333;
  text-decoration: none;
}

.breadcrumb-separator {
  color: #999;
  margin: 0 2px;
}

.upload-dragger {
  margin-top: 12px;
}

.upload-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}

.upload-preview {
  margin-top: 12px;
  background: #f7f8fb;
  border-radius: 8px;
  padding: 8px 12px;
}

.upload-item {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
}

.table-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.file-name-cell {
  display: flex;
  align-items: center;
}


.detail-icon {
  text-align: center;
}

.detail-title {
  font-weight: 700;
  font-size: 14px;
  margin-bottom: 6px;
  color: #333;
}

.detail-row {
  display: block;
  /* // justify-content: space-between; */
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-row:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.detail-row span {
  color: #666;
  font-size: 12px;
  display: block;
}

.detail-row strong {
  color: #333;
  font-size: 12px;
  font-weight: 500;
}

.left-column,
.center-column,
.right-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.subfolders-section {
  margin-top: 16px;
}

.subfolders-header {
  margin-bottom: 12px;
}

.subfolders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

/* Folder card styles (same as DocumentManagementView) */
.folder-card {
  border: 1px solid #e8e8e8;
  padding: 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
}

.folder-card:hover {
  border-color: #035cab;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.folder-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0px;
}

.folder-icon {
  flex: 1;
}

.folder-menu-icon {
  opacity: 0;
  transition: opacity 0.2s;
  cursor: pointer;
  color: #666;
}

.folder-card:hover .folder-menu-icon {
  opacity: 1;
}

.folder-name {
  font-weight: 500;
  line-height: 19px;
  font-size: 13px;
  color: #333;
  margin-top: 12px;
  margin-bottom: 8px;
}

.folder-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
  gap: 8px;
}

@media (max-width: 1100px) {
  .folder-detail-content {
    padding: 12px 12px 24px;
  }
  .left-column,
  .center-column,
  .right-column {
    width: 100%;
  }
  .right-column {
    margin-top: 8px;
  }
  .subfolders-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 12px;
  }
}
</style>
