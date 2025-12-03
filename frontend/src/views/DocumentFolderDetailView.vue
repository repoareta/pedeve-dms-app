<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import axios from 'axios'
import DashboardHeader from '../components/DashboardHeader.vue'
import DocumentSidebarActivityCard from '../components/DocumentSidebarActivityCard.vue'
import documentsApi, { type DocumentFolder, type DocumentItem } from '../api/documents'
import { auditApi, type UserActivityLog } from '../api/audit'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const folderId = computed(() => route.params.id as string)

const folders = ref<DocumentFolder[]>([])
const files = ref<DocumentItem[]>([])
const loading = ref(false)
const pageLoading = ref(true)

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

const loadFolders = async () => {
  try {
    folders.value = await documentsApi.listFolders()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat folder')
  }
}

const loadDocuments = async () => {
  loading.value = true
  try {
    files.value = await documentsApi.listDocuments({ folder_id: folderId.value })
    tablePagination.value.total = files.value.length
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat dokumen')
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
  if (!username) return colors[0]
  return colors[username.charCodeAt(0) % colors.length]
}

const handleRefresh = async () => {
  pageLoading.value = true
  await Promise.all([loadFolders(), loadDocuments(), loadActivities()])
  pageLoading.value = false
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

onMounted(async () => {
  pageLoading.value = true
  await Promise.all([loadFolders(), loadDocuments(), loadActivities()])
  pageLoading.value = false
})
</script>

<template>
  <div class="folder-detail-layout">
    <DashboardHeader @logout="handleLogout" />
    <div class="folder-detail-content">
      <div class="grid-layout">
        <div class="left-column">
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
        </div>

        <div class="center-column">
          <a-card class="upload-card" :bordered="false">
            <div class="card-header">
              <div>
                <h3>Folders / {{ currentFolder?.name || 'Folder' }}</h3>
                <p>Upload files</p>
              </div>
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
          </a-card>

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
            :custom-row="(record) => ({
              onClick: () => router.push(`/documents/${record.id}`),
              style: { cursor: 'pointer' }
            })"
            >
              <a-table-column
                title="Name"
                key="name"
                :sorter="(a, b) => (a.name || a.file_name || '').localeCompare(b.name || b.file_name || '')"
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
                :sorter="(a, b) => simplifyMime(a.mime_type).localeCompare(simplifyMime(b.mime_type))"
              >
                <template #default="{ record }">
                  <a-tag>{{ simplifyMime(record.mime_type) }}</a-tag>
                </template>
              </a-table-column>
              <a-table-column
                title="Uploaded by"
                key="uploader"
                :sorter="(a, b) => (a.uploader_id || '').localeCompare(b.uploader_id || '')"
              >
                <template #default="{ record }">
                  {{ record.uploader_id || '-' }}
                </template>
              </a-table-column>
              <a-table-column
                title="Last modified"
                key="last"
                :sorter="(a, b) => (dayjs(a.updated_at || 0).valueOf() - dayjs(b.updated_at || 0).valueOf())"
              >
                <template #default="{ record }">
                  {{ record.updated_at ? dayjs(record.updated_at).format('YYYY-MM-DD HH:mm') : '-' }}
                </template>
              </a-table-column>
            </a-table>
          </a-card>
        </div>

        <div class="right-column" v-if="currentFolder">
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
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.folder-detail-layout {
  min-height: 100vh;
  /* background: #f7f8fb; */
  margin-top: 70px;
}

.folder-detail-content {
  max-width: 1440px;
  margin: 0 auto;
  padding: 16px 24px 40px;
}

.grid-layout {
  display: grid;
  grid-template-columns: 280px 1fr 280px;
  gap: 20px;
}

.upload-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  margin-bottom: 12px;
  overflow: hidden;
}

.card-header h3 {
  margin: 0;
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

.detail-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  text-align: left;
  min-width: 260px;
}

.detail-icon {
  display: flex;
  justify-content: center;
  margin-bottom: 8px;
}

.detail-title {
  font-weight: 600;
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid #f0f0f0;
}

.detail-row:last-child {
  border-bottom: none;
}

@media (max-width: 1100px) {
  .grid-layout {
    grid-template-columns: 1fr;
  }
  .center-column {
    order: 2;
  }
  .right-column {
    order: 3;
  }
}
</style>
