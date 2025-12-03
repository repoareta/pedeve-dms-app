<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import documentsApi, { type DocumentItem } from '../api/documents'
import { auditApi, type UserActivityLog } from '../api/audit'
import DocumentSidebarActivityCard from '../components/DocumentSidebarActivityCard.vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import mammoth from 'mammoth'
import * as XLSX from 'xlsx'

dayjs.extend(relativeTime)

const router = useRouter()
const route = useRoute()

const documentId = computed(() => route.params.id as string)
const document = ref<DocumentItem | null>(null)
const loading = ref(true)
const pageLoading = ref(true)

// Activity feed
const activities = ref<UserActivityLog[]>([])
const activityLoading = ref(false)

// Document viewer
const zoomLevel = ref(100)

// Metadata
const metadata = ref<Record<string, unknown>>({})
const editMode = ref(false)

// Load document
const loadDocument = async () => {
  loading.value = true
  try {
    const data = await documentsApi.getDocument(documentId.value)
    document.value = data
    metadata.value = data.metadata || {}
    
    // Debug: log document data
    console.log('Document loaded:', {
      id: data.id,
      name: data.name,
      file_path: data.file_path,
      file_name: data.file_name,
      mime_type: data.mime_type
    })
    
    // Load file with authentication after document data is loaded
    await loadDocumentFile()
  } catch (error: unknown) {
    const err = error as { message?: string }
    console.error('Error loading document:', error)
    message.error(err.message || 'Gagal memuat dokumen')
    router.push('/documents')
  } finally {
    loading.value = false
  }
}

// Load activity feed (5 latest)
const loadActivities = async () => {
  activityLoading.value = true
  try {
    // Ambil aktivitas terbaru untuk resource document, lalu filter berdasarkan documentId ini
    const response = await auditApi.getUserActivityLogs({
      page: 1,
      pageSize: 50, // ambil lebih banyak lalu filter per dokumen
      resource: 'document',
    })
    const filtered = response.data
      .filter((item) => item.resource_id === documentId.value)
      .sort((a, b) => dayjs(b.created_at).valueOf() - dayjs(a.created_at).valueOf())
      .slice(0, 5)
    activities.value = filtered
  } catch (error: unknown) {
    console.error('Failed to load activities:', error)
    activities.value = []
  } finally {
    activityLoading.value = false
  }
}

// Format activity description
const getActivityDescription = (activity: UserActivityLog): string => {
  const action = activity.action.toLowerCase()
  const docName = document.value?.name || document.value?.file_name || 'dokumen ini'

  if (action.includes('view')) {
    return `Baru saja melihat dokumen ${docName}`
  }
  if (action.includes('create') || action.includes('upload')) {
    return `Baru saja mengunggah dokumen ${docName}`
  }
  if (action.includes('update')) {
    return `Baru saja memperbarui dokumen ${docName}`
  }
  if (action.includes('delete')) {
    return `Baru saja menghapus dokumen ${docName}`
  }
  return `Baru saja melakukan aksi ${action} pada dokumen ${docName}`
}

// Format timestamp to relative time
const formatTime = (timestamp: string): string => {
  const time = dayjs(timestamp)
  const now = dayjs()
  const diffSeconds = now.diff(time, 'second')
  
  if (diffSeconds < 60) {
    return `${diffSeconds} sec`
  } else if (diffSeconds < 3600) {
    const minutes = Math.floor(diffSeconds / 60)
    return `${minutes} min`
  } else if (diffSeconds < 86400) {
    const hours = Math.floor(diffSeconds / 3600)
    return `${hours} hour${hours > 1 ? 's' : ''}`
  } else {
    const days = Math.floor(diffSeconds / 86400)
    return `${days} day${days > 1 ? 's' : ''}`
  }
}

// Get user avatar/initial
const getDisplayName = (username: string): string => {
  if (!username) return ''
  // Ambil kata pertama jika ada spasi, jika tidak gunakan username apa adanya
  const parts = username.trim().split(/\s+/)
  return parts.length > 0 ? parts[0] : username
}

// Get user avatar color
const getUserAvatarColor = (username: string): string => {
  const colors: string[] = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#eb2f96']
  if (!username || username.length === 0) return colors[0] || '#1890ff'
  const firstChar = username.charAt(0)
  if (!firstChar) return colors[0] || '#1890ff'
  const index = firstChar.charCodeAt(0) % colors.length
  return colors[index] || '#1890ff'
}

// Document viewer controls (inactive for now)
// const handleZoomIn = () => {}
// const handleZoomOut = () => {}
// const handlePreviousPage = () => {}
// const handleNextPage = () => {}

const handleDownload = async () => {
  await loadDocumentFile()
  if (!documentBlobUrl.value || !document.value) {
    message.error('File tidak tersedia untuk diunduh')
    return
  }
  const link = window.document.createElement('a')
  link.href = documentBlobUrl.value
  link.download = document.value.file_name || document.value.name || 'document'
  link.click()
}


const openRenameModal = () => {
  renameInput.value = document.value?.name || document.value?.file_name || ''
  renameModalVisible.value = true
}

const submitRename = async () => {
  if (!document.value) return
  const newName = renameInput.value.trim()
  if (!newName) {
    message.warning('Nama dokumen wajib diisi')
    return
  }
  try {
    await documentsApi.updateDocument(document.value.id, { title: newName })
    document.value.name = newName
    message.success('Nama dokumen diperbarui')
    renameModalVisible.value = false
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal mengganti nama dokumen')
  }
}

const handleDelete = () => {
  if (!document.value) return
  Modal.confirm({
    title: 'Hapus dokumen?',
    content: 'Menghapus dokumen akan menghilangkan file ini secara permanen.',
    okType: 'danger',
    okText: 'Hapus',
    cancelText: 'Batal',
    async onOk() {
      try {
        await documentsApi.deleteDocument(document.value!.id)
        message.success('Dokumen dihapus')
        router.push('/documents')
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || 'Gagal menghapus dokumen')
      }
    },
  })
}

// Blob URL for secure file preview (downloaded with authentication)
const documentBlobUrl = ref<string | null>(null)
// HTML content for Office documents preview
const officeHtmlContent = ref<string | null>(null)
const officePreviewLoading = ref(false)
const renameModalVisible = ref(false)
const renameInput = ref('')

// Download file with authentication and create blob URL
const loadDocumentFile = async () => {
  if (!document.value || !document.value.file_path) {
    console.warn('Document or file_path is missing')
    return
  }

  try {
    // Construct file URL
    const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    let filePath = document.value.file_path.trim()
    
    // If file_path already starts with /api/v1/files/, use it directly
    if (!filePath.startsWith('/api/v1/files/')) {
      // Legacy support: convert to /api/v1/files/documents/...
      if (filePath.startsWith('/')) {
        filePath = filePath.substring(1)
      }
      filePath = `/api/v1/files/${filePath}`
    }
    
    const fileUrl = `${baseUrl}${filePath}`
    console.log('Downloading file with authentication:', fileUrl)
    
    // Download file with authentication (cookies will be sent automatically)
    const response = await fetch(fileUrl, {
      method: 'GET',
      credentials: 'include', // Include cookies for authentication
      headers: {
        'Accept': document.value.mime_type || 'application/pdf',
      },
    })
    
    if (!response.ok) {
      throw new Error(`Failed to download file: ${response.status} ${response.statusText}`)
    }
    
    // Create blob from response
    const blob = await response.blob()
    
    // Always create blob URL for download/print
    if (documentBlobUrl.value) {
      URL.revokeObjectURL(documentBlobUrl.value)
    }
    documentBlobUrl.value = URL.createObjectURL(blob)
    console.log('File downloaded and blob URL created:', documentBlobUrl.value)

    // For Office documents, convert to HTML for preview
    if (isOfficeDoc.value) {
      await convertOfficeToHtml(blob)
    }
  } catch (error) {
    console.error('Error loading document file:', error)
    message.error('Gagal memuat file dokumen')
  }
}

// Convert Office documents to HTML for preview
const convertOfficeToHtml = async (blob: Blob) => {
  if (!document.value) return
  
  officePreviewLoading.value = true
  officeHtmlContent.value = null
  
  try {
    const fileName = document.value.file_name?.toLowerCase() || ''
    const mimeType = document.value.mime_type || ''
    
    // Word documents (.docx)
    if (fileName.endsWith('.docx') || mimeType.includes('wordprocessingml')) {
      const arrayBuffer = await blob.arrayBuffer()
      const result = await mammoth.convertToHtml({ arrayBuffer })
      officeHtmlContent.value = result.value
      if (result.messages.length > 0) {
        console.warn('Mammoth conversion warnings:', result.messages)
      }
    }
    // Excel documents (.xlsx, .xls)
    else if (fileName.endsWith('.xlsx') || fileName.endsWith('.xls') || mimeType.includes('spreadsheetml')) {
      const arrayBuffer = await blob.arrayBuffer()
      const workbook = XLSX.read(arrayBuffer, { type: 'array' })
      
      // Convert first sheet to HTML table
      const firstSheetName = workbook.SheetNames[0]
      const worksheet = workbook.Sheets[firstSheetName]
      const html = XLSX.utils.sheet_to_html(worksheet)
      officeHtmlContent.value = html
    }
    // PowerPoint documents (.pptx, .ppt)
    else if (fileName.endsWith('.pptx') || fileName.endsWith('.ppt') || mimeType.includes('presentationml')) {
      // PowerPoint is more complex, show message to download
      officeHtmlContent.value = null
      message.info('Preview PowerPoint tidak tersedia. Silakan download file untuk membuka dengan aplikasi Office.')
    }
    // Legacy Word (.doc) - not supported by mammoth
    else if (fileName.endsWith('.doc') || mimeType.includes('msword')) {
      officeHtmlContent.value = null
      message.info('File .doc (legacy) tidak dapat di-preview. Silakan download file untuk membuka dengan aplikasi Office.')
    }
  } catch (error) {
    console.error('Error converting Office document to HTML:', error)
    officeHtmlContent.value = null
    message.warning('Gagal mengkonversi file Office ke preview. Silakan download file untuk membuka dengan aplikasi Office.')
  } finally {
    officePreviewLoading.value = false
  }
}

// Get document file URL (now returns blob URL for security)
const getDocumentUrl = computed(() => {
  return documentBlobUrl.value
})

// Check file type
const isPDF = computed(() => {
  return document.value?.mime_type === 'application/pdf' || 
         document.value?.file_name?.toLowerCase().endsWith('.pdf')
})

const isImage = computed(() => {
  const mimeType = document.value?.mime_type || ''
  const fileName = document.value?.file_name?.toLowerCase() || ''
  return mimeType.startsWith('image/') || 
         ['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp'].some(ext => fileName.endsWith(ext))
})

const isOfficeDoc = computed(() => {
  const mimeType = document.value?.mime_type || ''
  const fileName = document.value?.file_name?.toLowerCase() || ''
  
  // Office documents
  const officeMimeTypes = [
    'application/msword', // .doc
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document', // .docx
    'application/vnd.ms-excel', // .xls
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // .xlsx
    'application/vnd.ms-powerpoint', // .ppt
    'application/vnd.openxmlformats-officedocument.presentationml.presentation', // .pptx
  ]
  
  const officeExtensions = ['.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx']
  
  return officeMimeTypes.includes(mimeType) || 
         officeExtensions.some(ext => fileName.endsWith(ext))
})

// Get file type icon
const getFileTypeIcon = computed(() => {
  if (!document.value) return 'mdi:file-document-outline'
  const mimeType = document.value.mime_type.toLowerCase()
  if (mimeType.includes('pdf')) return 'mdi:file-pdf-box'
  if (mimeType.includes('word') || mimeType.includes('doc')) return 'mdi:file-word-box'
  if (mimeType.includes('excel') || mimeType.includes('spreadsheet') || mimeType.includes('xls')) return 'mdi:file-excel-box'
  if (mimeType.includes('image')) return 'mdi:file-image-box'
  return 'mdi:file-document-outline'
})

// Handle edit metadata - navigate to edit form
const handleEditMetadata = () => {
  if (document.value?.id) {
    router.push(`/documents/${document.value.id}/edit`)
  }
}

const handleSaveMetadata = async () => {
  // TODO: Implement save metadata API call
  message.success('Metadata berhasil disimpan')
  editMode.value = false
}

const handleCancelEdit = () => {
  editMode.value = false
  // Reload document to reset metadata
  loadDocument()
}

// Handle navigation
const handleAddFolderClick = () => {
  router.push('/documents')
}

const handleUploadClick = () => {
  router.push('/documents/upload')
}

const handleLogout = async () => {
  const { useAuthStore } = await import('../stores/auth')
  const authStore = useAuthStore()
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('Logout error:', error)
    router.push('/login')
  }
}

// Format date
const formatDate = (dateString?: string): string => {
  if (!dateString) return '-'
  return dayjs(dateString).format('YYYY-MM-DD')
}

onMounted(async () => {
  pageLoading.value = true
  try {
    await Promise.all([
      loadDocument(),
      loadActivities()
    ])
  } finally {
    pageLoading.value = false
  }
})

// Cleanup blob URL when component is unmounted
onBeforeUnmount(() => {
  if (documentBlobUrl.value) {
    URL.revokeObjectURL(documentBlobUrl.value)
    documentBlobUrl.value = null
  }
})
</script>

<template>
  <div class="document-detail-layout">
    <DashboardHeader @logout="handleLogout" />
    <div class="document-detail-content">
      <a-row :gutter="[20, 20]">
        <!-- Left Sidebar -->
        <a-col :xs="24" :lg="6" :xl="6" class="left-column">
          <DocumentSidebarActivityCard
            :activities="activities"
            :activity-loading="activityLoading"
            :page-loading="pageLoading"
            :hide-search="true"
            :show-see-all="false"
            @refresh="() => { loadDocument(); loadActivities(); message.success('Data diperbarui') }"
            @add-folder="handleAddFolderClick"
            @upload-file="handleUploadClick"
            @nav-dashboard="router.push('/documents')"
            @nav-recent="message.info('Belum ada data recent')"
            @nav-trash="message.info('Belum ada data trash')"
            @see-all="() => {}"
            :get-display-name="getDisplayName"
            :get-activity-description="getActivityDescription"
            :format-time="formatTime"
            :get-user-avatar-color="getUserAvatarColor"
          />
        </a-col>

        <!-- Main Content - Document Viewer -->
        <a-col :xs="24" :lg="12" :xl="12" class="center-column">
          <a-card class="document-viewer-card" :bordered="false" v-if="!loading && document">
            <!-- Document Header -->
            <div class="document-header">
              <div class="document-header-left">
                <IconifyIcon :icon="getFileTypeIcon" width="24" style="color: #035cab; margin-right: 12px;" />
                <div>
                  <div class="document-title">{{ document.name || document.file_name }}</div>
                  <div class="document-subtitle">Reports</div>
                </div>
              </div>
              <div class="document-header-right">
                <div class="action-controls">
                  <a-button type="text" size="small" title="Download" @click="handleDownload">
                    <IconifyIcon icon="mdi:download" width="20" />
                  </a-button>
                  <a-dropdown :trigger="['click']">
                    <a-button type="text" size="small" title="More">
                      <IconifyIcon icon="mdi:dots-vertical" width="20" />
                    </a-button>
                    <template #overlay>
                      <a-menu>
                        <a-menu-item key="rename" @click="openRenameModal">Rename</a-menu-item>
                        <a-menu-item key="delete" @click="handleDelete">Delete</a-menu-item>
                      </a-menu>
                    </template>
                  </a-dropdown>
                </div>
              </div>
            </div>

            <!-- Document Content -->
            <div class="document-content" :style="{ zoom: `${zoomLevel}%` }">
              <div v-if="isPDF && getDocumentUrl && getDocumentUrl !== 'http://localhost:8080/' && getDocumentUrl !== 'http://localhost:8080'" class="pdf-viewer">
                <iframe
                  :src="getDocumentUrl"
                  class="pdf-iframe"
                  frameborder="0"
                  @error="(e) => console.error('PDF iframe error:', e)"
                  @load="() => console.log('PDF iframe loaded:', getDocumentUrl)"
                ></iframe>
              </div>
              <div v-else-if="isPDF" class="unsupported-viewer">
                <div class="unsupported-message">
                  <IconifyIcon icon="mdi:alert-circle" width="64" style="color: #ff4d4f; margin-bottom: 16px;" />
                  <p v-if="!getDocumentUrl || getDocumentUrl === 'http://localhost:8080/' || getDocumentUrl === 'http://localhost:8080'">File path tidak ditemukan atau URL tidak valid</p>
                  <p v-else>Preview tidak tersedia</p>
                  <p style="font-size: 12px; color: #999; margin-top: 8px;">File path: {{ document?.file_path || 'N/A' }}</p>
                  <p style="font-size: 12px; color: #999;">Constructed URL: {{ getDocumentUrl || 'N/A' }}</p>
                  <a-button type="primary" @click="loadDocument" style="margin-top: 16px;">
                    <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 8px;" />
                    Reload Document
                  </a-button>
                </div>
              </div>
              <div v-else-if="isImage && getDocumentUrl" class="image-viewer">
                <img :src="getDocumentUrl" :alt="document.name" class="document-image" />
              </div>
              <div v-else-if="isOfficeDoc" class="office-viewer">
                <div v-if="officePreviewLoading" class="office-loading">
                  <a-spin size="large" />
                  <p style="margin-top: 16px; color: #666;">Mengkonversi file Office...</p>
                </div>
                <div v-else-if="officeHtmlContent" class="office-html-wrapper">
                  <div class="office-html-preview" v-html="officeHtmlContent"></div>
                </div>
                <div v-else class="office-fallback">
                  <IconifyIcon icon="mdi:file-document-outline" width="64" style="color: #999; margin-bottom: 16px;" />
                  <p>Preview tidak tersedia untuk file Office ini.</p>
                  <p style="font-size: 12px; color: #999; margin-top: 8px;">
                    File: {{ document?.file_name || 'N/A' }}
                  </p>
                  <a-button type="primary" @click="handleDownload" style="margin-top: 16px;">
                    <IconifyIcon icon="mdi:download" width="16" style="margin-right: 8px;" />
                    Download File
                  </a-button>
                </div>
              </div>
              <div v-else-if="!loading && document" class="unsupported-viewer">
                <div class="unsupported-message">
                  <IconifyIcon icon="mdi:alert-circle" width="64" style="color: #ff4d4f; margin-bottom: 16px;" />
                  <p>Format file tidak didukung untuk preview</p>
                  <p style="font-size: 12px; color: #999; margin-top: 8px;">
                    File: {{ document.file_name || 'N/A' }}
                  </p>
                  <p style="font-size: 12px; color: #999;">
                    Type: {{ document.mime_type || 'N/A' }}
                  </p>
                  <a-button type="primary" @click="handleDownload" style="margin-top: 16px;">
                    <IconifyIcon icon="mdi:download" width="16" style="margin-right: 8px;" />
                    Download File
                  </a-button>
                </div>
              </div>
              <div v-else class="unsupported-viewer">
                <div class="unsupported-message">
                  <IconifyIcon :icon="getFileTypeIcon" width="64" style="color: #999; margin-bottom: 16px;" />
                  <p>Preview tidak tersedia untuk tipe file ini</p>
                  <a-button type="primary" @click="window.open(getDocumentUrl, '_blank')">
                    <IconifyIcon icon="mdi:download" width="16" style="margin-right: 8px;" />
                    Download untuk melihat
                  </a-button>
                </div>
              </div>
            </div>
          </a-card>

          <div v-else-if="loading || pageLoading" class="document-loading">
            <a-spin size="large" />
            <p style="margin-top: 16px; color: #999;">Memuat dokumen...</p>
          </div>
        </a-col>

        <!-- Right Sidebar - Metadata Panel -->
        <a-col :xs="24" :lg="6" :xl="6" class="right-column">
          <a-card class="metadata-card" :bordered="false" v-if="!loading && document">
            <div class="metadata-header">
              <h3 class="metadata-title">Meta Data</h3>
            </div>
            <div class="metadata-content">
              <div class="metadata-field">
                <label>Judul Dokumen</label>
                <div class="metadata-value">{{ metadata.document_title || document.name || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Document ID</label>
                <div class="metadata-value">{{ metadata.document_id || document.id || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Jenis Dokumen</label>
                <div class="metadata-value">{{ metadata.document_type || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Nomor Dokumen / Referensi</label>
                <div class="metadata-value">{{ metadata.document_number || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Unit / Divisi</label>
                <div class="metadata-value">{{ metadata.unit || metadata.division || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Uploaded By / PIC</label>
                <div class="metadata-value">{{ metadata.uploaded_by || metadata.pic || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Status Dokumen</label>
                <div class="metadata-value">
                  <a-tag :color="metadata.status === 'Disetujui' || metadata.status === 'Approved' ? 'success' : 'default'">
                    {{ metadata.status || '-' }}
                  </a-tag>
                </div>
              </div>
              <div class="metadata-field">
                <label>Tanggal Dokumen (Diterbitkan)</label>
                <div class="metadata-value">{{ formatDate(metadata.document_date as string) || formatDate(metadata.published_date as string) || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Tanggal Berlaku</label>
                <div class="metadata-value">{{ formatDate(metadata.effective_date as string) || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Tanggal Berakhir</label>
                <div class="metadata-value">{{ formatDate(metadata.expired_date as string) || '-' }}</div>
              </div>
              <div class="metadata-field">
                <label>Is Active</label>
                <div class="metadata-value">
                  <a-tag :color="metadata.is_active ? 'success' : 'default'">
                    {{ metadata.is_active ? 'Aktif' : 'Tidak Aktif' }}
                  </a-tag>
                </div>
              </div>
            </div>
            <div class="metadata-actions">
              <a-button type="primary" block @click="handleEditMetadata" v-if="!editMode">
                <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                Edit Meta Data
              </a-button>
              <div v-else class="edit-actions">
                <a-button block @click="handleCancelEdit" style="margin-bottom: 8px;">
                  Batal
                </a-button>
                <a-button type="primary" block @click="handleSaveMetadata">
                  Simpan
                </a-button>
              </div>
            </div>
          </a-card>

          <div v-else-if="loading || pageLoading" class="metadata-loading">
            <a-skeleton active :paragraph="{ rows: 10 }" />
          </div>
        </a-col>
      </a-row>
    </div>
  </div>

  <a-modal
    v-model:open="renameModalVisible"
    title="Rename Document"
    ok-text="Simpan"
    cancel-text="Batal"
    @ok="submitRename"
  >
    <a-input v-model:value="renameInput" placeholder="Nama dokumen baru" />
  </a-modal>
</template>

<style scoped lang="scss">
.document-detail-layout {
  min-height: 100vh;
  // background: #f5f5f5;
  padding-top: 70px;
  // background: orange !important;
}

.document-detail-content {
  padding: 24px;
  max-width: 1440px;
  margin: 0 auto;
}

// Left Sidebar (same as DocumentManagementView)
.left-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.search-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.search-actions {
  display: flex;
  gap: 8px;
}

.action-icon {
  color: #666;
  cursor: pointer;
  transition: color 0.2s;

  &:hover {
    color: #035cab;
  }
}

.new-folder-btn {
  width: 100%;
  height: 40px;
  margin-top: 16px;
}

.nav-links {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-link {
  text-align: left;
  height: 40px;
  padding: 0 12px;
}

.activity-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.activity-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.see-all-btn {
  padding: 0;
  height: auto;
  font-size: 13px;
}

.activity-list {
  max-height: 400px;
  overflow-y: auto;
}

.activity-timeline {
  position: relative;
}

.activity-item {
  display: flex;
  gap: 12px;
  padding-bottom: 16px;
  position: relative;
}

.activity-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-user {
  font-weight: 600;
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.activity-description {
  font-size: 13px;
  color: #666;
  margin-bottom: 4px;
  line-height: 1.4;
}

.activity-time {
  font-size: 12px;
  color: #999;
}

.activity-line {
  position: absolute;
  left: 18px;
  top: 36px;
  bottom: 0;
  width: 2px;
  background: #e8e8e8;
}

.activity-skeleton,
.activity-empty {
  padding: 16px;
}

.activity-empty {
  text-align: center;
  color: #999;
  font-size: 13px;
}

// Center Column - Document Viewer
.center-column {
  display: flex;
  flex-direction: column;
}

.document-viewer-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 600px;
}

.document-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e8e8e8;
}

.document-header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.document-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.document-subtitle {
  font-size: 13px;
  color: #999;
}

.document-header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.pagination-controls,
.zoom-controls,
.action-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-info,
.zoom-info {
  font-size: 13px;
  color: #666;
  min-width: 60px;
  text-align: center;
}

.document-content {
  flex: 1;
  overflow: auto;
  padding: 24px;
  background: #f5f5f5;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.pdf-viewer,
.office-viewer {
  width: 100%;
  height: 100%;
  min-height: 600px;
  position: relative;
}

.office-preview-container {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  flex-direction: column;
}

.office-html-wrapper {
  width: 100%;
  display: flex;
  justify-content: center;
}

.office-iframe {
  width: 100%;
  height: 100%;
  min-height: 600px;
  border: none;
  flex: 1;
}

.office-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
}

.office-html-preview {
  width: 100%;
  max-width: 1100px;
  overflow: auto;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  margin: 0 auto;
}

.office-html-preview :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}

.office-html-preview :deep(table th),
.office-html-preview :deep(table td) {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

.office-html-preview :deep(table th) {
  background-color: #f5f5f5;
  font-weight: 600;
}

.office-html-preview :deep(p) {
  margin: 12px 0;
  line-height: 1.6;
}

.office-html-preview :deep(h1),
.office-html-preview :deep(h2),
.office-html-preview :deep(h3) {
  margin: 16px 0 8px 0;
  font-weight: 600;
}

.office-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 400px;
  text-align: center;
  padding: 24px;
}

.office-fallback p {
  margin: 8px 0;
  color: #666;
}

.pdf-iframe {
  width: 100%;
  height: 100%;
  min-height: 600px;
  border: none;
}

.image-viewer {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.document-image {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
}

.unsupported-viewer {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.unsupported-message {
  text-align: center;
  color: #999;
}

.document-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 600px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

// Right Column - Metadata Panel
.right-column {
  display: flex;
  flex-direction: column;
}

.metadata-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 120px);
}

.metadata-header {
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
  margin-bottom: 16px;
}

.metadata-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.metadata-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metadata-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metadata-field label {
  font-size: 12px;
  color: #999;
  font-weight: 500;
}

.metadata-value {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.metadata-actions {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #e8e8e8;
}

.edit-actions {
  display: flex;
  flex-direction: column;
}

.metadata-loading {
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}
</style>
