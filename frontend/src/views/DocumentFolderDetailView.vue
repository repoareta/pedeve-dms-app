<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import axios from 'axios'
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
  try {
    await Promise.all([loadUsers(), loadFolders(), loadDocuments(), loadActivities()])
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
  /* // margin-bottom: 16px; */
}

.detail-icon {
  text-align: center;
  /* // margin-bottom: 16px; */
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
}
</style>
