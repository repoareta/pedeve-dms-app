<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import DocumentSidebarActivityCard from '../components/DocumentSidebarActivityCard.vue'
import documentsApi, { type DocumentFolder, type DocumentItem, type DocumentFolderStat } from '../api/documents'
import { auditApi, type UserActivityLog } from '../api/audit'
import { userApi, type User } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const router = useRouter()
const authStore = useAuthStore()

// Check user role
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperAdminOrAdministrator = computed(() => {
  return userRole.value === 'superadmin' || userRole.value === 'administrator'
})

const folders = ref<DocumentFolder[]>([])
const pagedFiles = ref<DocumentItem[]>([])
const totalDocuments = ref(0)

const recentSearch = ref('')
const folderSearch = ref('')
const selectedFolderId = ref<string | undefined>(undefined)
const addFolderModalVisible = ref(false)
const newFolderName = ref('')
const renameFolderModalVisible = ref(false)
const renameFolderName = ref('')
const folderBeingEdited = ref<DocumentFolder | null>(null)
const searchModalVisible = ref(false)
const searchQuery = ref('')
const searchResults = ref<DocumentItem[]>([])
const loading = ref(false)
const tableLoading = ref(false)
const pageLoading = ref(true) // Initial page load
const tablePage = ref(1)
const tablePageSize = ref(5)
const currentSortBy = ref<string>('updated_at') // Sort by updated_at for recently files
const currentSortDir = ref<string>('desc') // Descending: newest first
const currentTypeFilter = ref<string>('')
const folderStatsMap = ref<Record<string, { count: number; size: number }>>({})
const totalStorageSize = ref(0)
const storageBreakdownSource = ref<DocumentItem[]>([])
const userMap = ref<Record<string, string>>({})

// Activity feed
const activities = ref<UserActivityLog[]>([])
const activityLoading = ref(false)

// View and filter
const viewMode = ref('grid')
const timeFilter = ref('last-month')
const foldersExpanded = ref(false)

// Computed
const selectedFolder = computed(() => {
  return folders.value.find(f => f.id === selectedFolderId.value)
})

const loadFolders = async () => {
  loading.value = true
  try {
    const data = await documentsApi.listFolders()
    folders.value = data
    // Default: tidak memilih folder mana pun (recent files = semua file)
    selectedFolderId.value = undefined
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat folder')
  } finally {
    loading.value = false
  }
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
    userMap.value = {}
  }
}

const loadDocumentSummary = async () => {
  let summaryTotal = 0
  try {
    const res = await documentsApi.getDocumentSummary()
    const map: Record<string, { count: number; size: number }> = {}
    res.folder_stats.forEach((item: DocumentFolderStat) => {
      const key = item.folder_id || 'unassigned'
      map[key] = {
        count: Number(item.file_count) || 0,
        size: Number(item.total_size) || 0,
      }
    })
    folderStatsMap.value = map
    totalStorageSize.value = res.total_size || 0
    summaryTotal = res.total_size || 0
    storageBreakdownSource.value = [] // gunakan data dari summary; legend akan fallback ke pagedFiles

  } catch (error) {
    console.error('Failed to load document summary', error)
  }

  // Hitung manual untuk breakdown; jangan timpa total jika summary sudah punya nilai >0
  await computeStatsFromAllDocuments(summaryTotal)
}

const computeStatsFromAllDocuments = async (summaryTotal = 0) => {
  try {
    let page = 1
    const pageSize = 500
    let docs: DocumentItem[] = []
    while (true) {
      const res = await documentsApi.listDocumentsPaginated({
        page,
        page_size: pageSize,
        sort_by: 'created_at',
        sort_dir: 'desc',
      })
      docs = docs.concat(res.data)
      if (docs.length >= res.total || res.data.length === 0) break
      page += 1
      if (page > 20) break // guard
    }

    const map: Record<string, { count: number; size: number }> = {}
    let totalSize = 0
    docs.forEach((doc) => {
      const key = doc.folder_id || 'unassigned'
      if (!map[key]) {
        map[key] = { count: 0, size: 0 }
      }
      map[key].count += 1
      map[key].size += doc.size || 0
      totalSize += doc.size || 0
    })
    folderStatsMap.value = map
    // Jika sudah punya total dari summary (global), pertahankan; jika belum, pakai hasil manual
    if (summaryTotal > 0) {
      totalStorageSize.value = summaryTotal
    } else {
      totalStorageSize.value = totalSize
    }
    storageBreakdownSource.value = docs
  } catch (error) {
    console.error('Fallback summary calculation failed', error)
  }
}

const loadDocuments = async (opts: { page?: number; pageSize?: number; search?: string; sortBy?: string; sortDir?: string; type?: string } = {}) => {
  tableLoading.value = true
  try {
    const page = opts.page ?? tablePage.value
    const pageSize = opts.pageSize ?? tablePageSize.value
    const search = opts.search ?? recentSearch.value
    const sortBy = opts.sortBy ?? currentSortBy.value
    const sortDir = opts.sortDir ?? currentSortDir.value
    const type = opts.type ?? currentTypeFilter.value

    const res = await documentsApi.listDocumentsPaginated({
      page,
      page_size: pageSize,
      search,
      sort_by: sortBy,
      sort_dir: sortDir,
      type,
    })
    
    // Sort by updated_at DESC (newest first) as fallback to ensure correct order
    const sortedData = [...res.data].sort((a, b) => {
      const dateA = new Date(a.updated_at || a.created_at || 0).getTime()
      const dateB = new Date(b.updated_at || b.created_at || 0).getTime()
      return dateB - dateA // Descending: newest first
    })
    
    pagedFiles.value = sortedData
    totalDocuments.value = res.total
    tablePage.value = res.page
    tablePageSize.value = res.page_size
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat dokumen')
  } finally {
    tableLoading.value = false
  }
}

const handleAddFolderClick = () => {
  addFolderModalVisible.value = true
  newFolderName.value = ''
}

const handleOpenFolder = (folderId: string) => {
  router.push(`/documents/folders/${folderId}`)
}

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
    folders.value = folders.value.map(f => f.id === updated.id ? updated : f)
    message.success('Folder berhasil diubah')
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal mengganti nama folder')
  } finally {
    renameFolderModalVisible.value = false
    folderBeingEdited.value = null
  }
}

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
        folders.value = folders.value.filter(f => f.id !== folder.id)
        if (selectedFolderId.value === folder.id) {
          selectedFolderId.value = undefined
        }
        message.success('Folder dan file di dalamnya telah dihapus')
        await loadDocumentSummary()
        await loadDocuments({ page: 1 })
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || 'Gagal menghapus folder')
      }
    },
  })
}

const handleCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    message.warning('Nama folder wajib diisi')
    return
  }
  try {
    const folder = await documentsApi.createFolder(newFolderName.value.trim())
    folders.value = [folder, ...folders.value]
    selectedFolderId.value = folder.id
    addFolderModalVisible.value = false
    message.success('Folder berhasil dibuat')
    await Promise.all([loadDocuments({ page: 1 }), loadDocumentSummary()])
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal membuat folder')
  }
}

const handleUploadClick = () => {
  router.push('/documents/upload')
}

const handleRefresh = async () => {
  pageLoading.value = true
  try {
    await Promise.all([loadFolders(), loadDocuments({ page: 1 }), loadDocumentSummary(), loadActivities()])
    tablePage.value = 1
    message.success('Data telah diperbarui')
  } catch {
    message.error('Gagal menyegarkan data')
  } finally {
    pageLoading.value = false
  }
}

watch(recentSearch, (val) => {
  tablePage.value = 1
  loadDocuments({ page: 1, search: val })
})

const openSearchModal = () => {
  searchModalVisible.value = true
  searchQuery.value = ''
}

let searchTimeout: ReturnType<typeof setTimeout> | undefined
watch(searchQuery, (val) => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(async () => {
    const q = val.trim()
    if (!q) {
      searchResults.value = []
      return
    }
    try {
      const res = await documentsApi.listDocumentsPaginated({
        search: q,
        page: 1,
        page_size: 10,
      })
      searchResults.value = res.data
    } catch (error) {
      console.error('Search error', error)
      searchResults.value = []
    }
  }, 300)
})

const getFolderName = (folderId?: string | null): string => {
  if (!folderId) return '-'
  const folder = folders.value.find(f => f.id === folderId)
  return folder?.name || '-'
}

const handleSelectSearchResult = (docId: string) => {
  searchModalVisible.value = false
  router.push(`/documents/${docId}`)
}

const selectFolder = (folderId: string | undefined) => {
  selectedFolderId.value = folderId
  // Klik tunggal hanya untuk seleksi; tidak ada notifikasi
}

// Handle row click to navigate to document detail
const handleRowClick = (record: DocumentItem) => {
  router.push(`/documents/${record.id}`)
}

// Load activity feed (5 latest)
const loadActivities = async () => {
  activityLoading.value = true
  try {
    const params: Record<string, unknown> = {
      page: 1,
      pageSize: 5, // tampilkan 5 aktivitas terbaru saja
      resource: 'document',
    }
    // Jika bukan superadmin, backend sudah otomatis filter ke user sendiri. Untuk superadmin biarkan melihat semua.
    const response = await auditApi.getUserActivityLogs(params)
    activities.value = response.data
  } catch (error: unknown) {
    console.error('Failed to load activities:', error)
    activities.value = []
  } finally {
    activityLoading.value = false
  }
}

const getDisplayName = (username: string): string => {
  if (!username) return ''
  const parts = username.trim().split(/\s+/)
  return parts[0] || username
}

// Format activity description (lebih naratif)
const getDocumentNameById = (id: string | undefined): string => {
  if (!id) return ''
  const doc = pagedFiles.value.find(f => f.id === id) || searchResults.value.find(f => f.id === id)
  if (doc) return doc.name || doc.file_name || ''
  return ''
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

  if (!fileName) {
    fileName = getDocumentNameById(activity.resource_id) || 'dokumen'
  }
  const target = fileName || 'dokumen'

  if (action.includes('update') || action.includes('edit')) {
    return `Telah mengupdate file ${target}`
  }
  if (action.includes('create') || action.includes('upload')) {
    return `Telah mengunggah file ${target}`
  }
  if (action.includes('delete')) {
    return `Telah menghapus file ${target}`
  }
  if (action.includes('view')) {
    return `Telah melihat file ${target}`
  }
  return `Telah melakukan aksi ${action} pada ${target}`
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

// Get user avatar color (dummy - untuk styling)
const getUserAvatarColor = (username: string): string => {
  const colors: string[] = ['#82a2bf', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#eb2f96']
  if (!username || username.length === 0) return colors[0] || '#82a2bf'
  const firstChar = username.charAt(0)
  if (!firstChar) return colors[0] || '#82a2bf'
  const index = firstChar.charCodeAt(0) % colors.length
  return colors[index] || colors[0] || '#82a2bf'
}

const legendColorClass = (type: string, idx: number) => {
  const classes = ['red', 'blue', 'orange', 'green']
  if (type.toLowerCase().includes('pdf')) return 'red'
  if (type.toLowerCase().includes('doc')) return 'blue'
  if (type.toLowerCase().includes('png')) return 'orange'
  if (type.toLowerCase().includes('ppt')) return 'green'
  return classes[idx % classes.length]
}

// Folder helpers
// Check if folder has subfolders
const hasSubfolders = (folderId: string): boolean => {
  return folders.value.some(f => f.parent_id === folderId)
}

// Recursively get all child folder IDs including the folder itself
const getAllChildFolderIds = (folderId: string): string[] => {
  const result: string[] = [folderId]
  const children = folders.value.filter(f => f.parent_id === folderId)
  for (const child of children) {
    result.push(...getAllChildFolderIds(child.id))
  }
  return result
}

// Recursive count: includes files from all subfolders
const getFolderFileCount = (folderId: string): number => {
  const allFolderIds = getAllChildFolderIds(folderId)
  let totalCount = 0
  for (const id of allFolderIds) {
    const key = id || 'unassigned'
    totalCount += folderStatsMap.value[key]?.count || 0
  }
  return totalCount
}

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

// Recursive size: includes files from all subfolders
const getFolderSize = (folderId: string): string => {
  const allFolderIds = getAllChildFolderIds(folderId)
  let totalSize = 0
  for (const id of allFolderIds) {
    const key = id || 'unassigned'
    totalSize += folderStatsMap.value[key]?.size || 0
  }
  return formatBytes(totalSize)
}

const getSelectedFolderSize = (): string => {
  if (!selectedFolder.value) return '0B'
  return getFolderSize(selectedFolder.value.id)
}

const getSelectedFolderLastEdited = (): string => {
  if (!selectedFolder.value) return '24/07/2025 14:32'
  if (!selectedFolder.value.created_at) return '-'
  const date = new Date(selectedFolder.value.created_at)
  return `${date.getDate().toString().padStart(2, '0')}/${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getFullYear()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

// Storage summary
const STORAGE_CAPACITY_BYTES = 5 * 1024 * 1024 * 1024 // 5 GB default
const storageUsage = computed(() => {
  const used = totalStorageSize.value
  const capacity = STORAGE_CAPACITY_BYTES
  const percent = capacity > 0 ? Math.min(100, (used / capacity) * 100) : 0
  return {
    used,
    capacity,
    percent: Number(percent.toFixed(1)),
    text: `${formatBytes(used)} of ${formatBytes(capacity)}`,
    isAlmostFull: percent >= 85,
  }
})

const storageBreakdown = computed(() => {
  const counts: Record<'pdf' | 'doc' | 'png' | 'ppt', { count: number; size: number }> = {
    pdf: { count: 0, size: 0 },
    doc: { count: 0, size: 0 },
    png: { count: 0, size: 0 },
    ppt: { count: 0, size: 0 },
  }

  const source = storageBreakdownSource.value.length ? storageBreakdownSource.value : pagedFiles.value

  source.forEach((file) => {
    const mime = (file.mime_type || '').toLowerCase()
    const size = file.size || 0
    if (mime.includes('pdf')) {
      counts.pdf.count += 1
      counts.pdf.size += size
    } else if (mime.includes('word') || mime.includes('doc')) {
      counts.doc.count += 1
      counts.doc.size += size
    } else if (mime.includes('presentation') || mime.includes('powerpoint') || mime.includes('ppt')) {
      counts.ppt.count += 1
      counts.ppt.size += size
    } else if (mime.includes('png') || mime.includes('jpeg') || mime.includes('jpg') || mime.includes('image')) {
      counts.png.count += 1
      counts.png.size += size
    }
  })

  const entries = Object.entries(counts).filter(([, val]) => val.count > 0)
  return entries.length ? entries : Object.entries(counts) // jika kosong, tampilkan nol semua
})

// Folder list filtering by time + search
const filteredFolders = computed(() => {
  const keyword = folderSearch.value.trim().toLowerCase()
  const now = dayjs()
  // Filter: hanya tampilkan root folders (yang tidak punya parent_id)
  return folders.value.filter(folder => {
    // Only show root folders (no parent_id)
    if (folder.parent_id) return false
    
    const matchSearch =
      !keyword ||
      folder.name?.toLowerCase().includes(keyword) ||
      folder.id?.toLowerCase().includes(keyword)

    const created = folder.created_at ? dayjs(folder.created_at) : null
    let matchTime = true
    if (created) {
      if (timeFilter.value === 'last-month') {
        matchTime = created.isAfter(now.subtract(1, 'month'))
      } else if (timeFilter.value === 'last-week') {
        matchTime = created.isAfter(now.subtract(1, 'week'))
      }
    }

    return matchSearch && matchTime
  }).sort((a, b) => {
    const tA = a.created_at ? dayjs(a.created_at).valueOf() : 0
    const tB = b.created_at ? dayjs(b.created_at).valueOf() : 0
    return tB - tA
  })
})

// File helpers
const getFileTypeLabel = (mimeType: string | undefined): string => {
  if (!mimeType) return 'Unknown'
  if (mimeType.includes('pdf')) return 'Pdf'
  if (mimeType.includes('excel') || mimeType.includes('spreadsheet')) return 'Excel'
  if (mimeType.includes('word') || mimeType.includes('document')) return 'Word'
  if (mimeType.includes('image')) return 'Image'
  return 'File'
}

const getFileTypeColor = (mimeType: string | undefined): string => {
  if (!mimeType) return 'default'
  if (mimeType.includes('pdf')) return 'red'
  if (mimeType.includes('excel') || mimeType.includes('spreadsheet')) return 'green'
  if (mimeType.includes('word') || mimeType.includes('document')) return 'blue'
  return 'default'
}

const getUploaderName = (uploaderId: string | undefined): string => {
  if (!uploaderId) return 'Unknown'
  const fromMap = userMap.value[uploaderId]
  if (fromMap) return fromMap
  const match = pagedFiles.value.find((f) => f.uploader_id === uploaderId)
  const metaName = (match?.metadata as Record<string, unknown> | undefined)?.['uploaded_by'] as string | undefined
  return match?.uploader_name || metaName || uploaderId
}

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

// Helper function to get metadata from document
const getMeta = (record: DocumentItem): Record<string, unknown> => {
  return (record.metadata as Record<string, unknown> | undefined) || {}
}

// Check if document has complete metadata (same validation as DocumentFolderDetailView)
const hasCompleteMetadata = (file: DocumentItem): boolean => {
  const meta = getMeta(file)
  const required = ['doc_type', 'reference', 'issued_date', 'effective_date', 'expired_date', 'is_active']
  return required.every((key) => Boolean(meta[key]))
}

const tableColumns = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name',
    sorter: (a: DocumentItem, b: DocumentItem) => (a.name || '').localeCompare(b.name || ''),
  },
  {
    title: 'Type',
    dataIndex: 'mime_type',
    key: 'mime_type',
    width: 110,
    filters: [
      { text: 'PDF', value: 'pdf' },
      { text: 'Word', value: 'word' },
      { text: 'Excel', value: 'excel' },
      { text: 'Image', value: 'image' },
    ],
    onFilter: (value: string, record: DocumentItem) => (record.mime_type || '').toLowerCase().includes(value),
    sorter: (a: DocumentItem, b: DocumentItem) => (a.mime_type || '').localeCompare(b.mime_type || ''),
  },
  {
    title: 'Uploaded by',
    dataIndex: 'uploader_id',
    key: 'uploader_id',
    width: 130,
    sorter: (a: DocumentItem, b: DocumentItem) => (a.uploader_id || '').localeCompare(b.uploader_id || ''),
  },
  {
    title: 'Meta Data status',
    dataIndex: 'updated_at',
    key: 'metadata_status',
    width: 220,
    sorter: (a: DocumentItem, b: DocumentItem) => {
      const tA = a.updated_at ? dayjs(a.updated_at).valueOf() : a.created_at ? dayjs(a.created_at).valueOf() : 0
      const tB = b.updated_at ? dayjs(b.updated_at).valueOf() : b.created_at ? dayjs(b.created_at).valueOf() : 0
      return tA - tB
    },
  },
]

const handleTableChange = (
  pag: { current?: number; pageSize?: number },
  filters: Record<string, (string | number | boolean)[] | null>,
  sorter: { field?: string; order?: string; columnKey?: string }
) => {
  const nextPage = pag.current || tablePage.value
  const nextSize = pag.pageSize || tablePageSize.value

  // Sorting
  const sortField = sorter?.field || sorter?.columnKey || 'created_at'
  const sortOrder = sorter?.order === 'ascend' ? 'asc' : sorter?.order === 'descend' ? 'desc' : currentSortDir.value
  const allowedSort = ['name', 'mime_type', 'uploader_id', 'updated_at', 'size']
  currentSortBy.value = allowedSort.includes(String(sortField)) ? String(sortField) : 'updated_at'
  currentSortDir.value = sortOrder

  // Type filter
  const typeFilterArr = (filters?.mime_type as string[]) || []
  currentTypeFilter.value = typeFilterArr[0] || ''

  tablePage.value = nextPage
  tablePageSize.value = nextSize

  loadDocuments({
    page: nextPage,
    pageSize: nextSize,
    sortBy: currentSortBy.value,
    sortDir: currentSortDir.value,
    type: currentTypeFilter.value,
  })
}

// Handle logout
const handleLogout = async () => {
  const { useAuthStore } = await import('../stores/auth')
  const authStore = useAuthStore()
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('Logout error:', error)
    // Force redirect even if logout fails
    router.push('/login')
  }
}

onMounted(async () => {
  pageLoading.value = true
  try {
    await Promise.all([
      loadFolders(),
      loadDocuments({ page: 1 }),
      loadDocumentSummary(),
      loadActivities(),
      loadUsers(),
    ])
  } finally {
    pageLoading.value = false
  }
})
</script>

<template>
  <div class="documents-layout">
    <DashboardHeader @logout="handleLogout" />
    <div class="documents-content">
      <!-- <div class="hero-card">
        <div class="hero-left">
          <h1>Documents</h1>
          <p>Kelola folder, upload file, dan lengkapi metadata dokumen.</p>
        </div>
        <div class="hero-actions">
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item key="add-folder" @click="handleAddFolderClick">
                  <IconifyIcon icon="mdi:folder-plus-outline" width="18" style="margin-right: 8px;" />
                  Add Folder
                </a-menu-item>
                <a-menu-item key="upload-file" @click="handleUploadClick">
                  <IconifyIcon icon="mdi:upload" width="18" style="margin-right: 8px;" />
                  Upload File
                </a-menu-item>
              </a-menu>
            </template>
            <a-button type="primary" size="large" class="new-btn">
              <IconifyIcon icon="mdi:plus" width="18" style="margin-right: 8px;" />
              Baru
            </a-button>
          </a-dropdown>
        </div>
      </div> -->

      <a-row :gutter="[16, 16]" style="margin-top: 70px;">
        <a-col :xs="24" :lg="5" :xl="5" class="left-column1">
          <DocumentSidebarActivityCard
            :activities="activities"
            :activity-loading="activityLoading"
            :page-loading="pageLoading"
            :hide-search="false"
            @search="openSearchModal"
            @refresh="handleRefresh"
            @add-folder="handleAddFolderClick"
            @upload-file="handleUploadClick"
            @nav-dashboard="router.push('/documents')"
            @nav-recent="message.info('Belum ada data recent')"
            @nav-trash="message.info('Belum ada data trash')"
            @see-all="message.info('Belum ada fitur see all activity')"
            :get-display-name="getDisplayName"
            :get-activity-description="getActivityDescription"
            :format-time="formatTime"
            :get-user-avatar-color="getUserAvatarColor"
          />
        </a-col>

        <a-col :xs="24" :lg="15" :xl="15" class="center-column1">
          <a-card class="folders-card all-folders" :class="{ 'folders-expanded': foldersExpanded }" :bordered="false">
            <div class="folders-scroll-container">
              <div class="folders-header">
                <div class="title">Folders</div>
                <div class="folders-actions">
                  <a-input
                    v-model:value="folderSearch"
                    placeholder="Search"
                    allow-clear
                    class="search-input"
                    size="small"
                    style="width: 150px; margin-right: 8px;"
                  >
                    <template #prefix>
                      <IconifyIcon icon="mdi:magnify" width="16" />
                    </template>
                  </a-input>
                  <a-select v-model:value="viewMode" style="width: 120px; margin-right: 8px;" size="small">
                    <a-select-option value="grid">View</a-select-option>
                    <a-select-option value="list">List</a-select-option>
                  </a-select>
                  <a-select v-model:value="timeFilter" style="width: 120px; margin-right: 8px;" size="small">
                    <a-select-option value="all">All Time</a-select-option>
                    <a-select-option value="last-month">Last Month</a-select-option>
                    <a-select-option value="last-week">Last Week</a-select-option>
                  </a-select>
                  <a-button 
                    type="text" 
                    size="small" 
                    @click="foldersExpanded = !foldersExpanded"
                    style="padding: 4px 8px; display: inline-flex; align-items: center; justify-content: center; min-width: 32px;"
                    :title="foldersExpanded ? 'Collapse' : 'Expand'"
                  >
                    <template #icon>
                      <IconifyIcon icon="material-symbols:expand-content-rounded" width="20" />
                    </template>
                  </a-button>
                </div>
              </div>
              <div class="folders-content" :class="{ 'expanded': foldersExpanded }">
            <div v-if="pageLoading || loading" class="folders-skeleton">
              <a-skeleton active :paragraph="{ rows: 0 }" :title="{ width: '100%' }" />
              <div class="folders-row-skeleton">
                <a-skeleton-button active :size="'large'" :block="false" style="width: 180px; height: 140px; border-radius: 8px;" />
                <a-skeleton-button active :size="'large'" :block="false" style="width: 180px; height: 140px; border-radius: 8px;" />
                <a-skeleton-button active :size="'large'" :block="false" style="width: 180px; height: 140px; border-radius: 8px;" />
                <a-skeleton-button active :size="'large'" :block="false" style="width: 180px; height: 140px; border-radius: 8px;" />
              </div>
            </div>
            <transition name="fade">
              <div v-if="viewMode === 'grid'" class="folders-row1">
                <a-row :gutter="[14, 14]">
                  <a-col :xs="24" :sm="12" :md="12" :lg="8" :xl="6">
                    <div class="folder-card add-folder-card" @click="handleAddFolderClick">
                      <div class="folder-icon-large">
                        <IconifyIcon icon="mdi:plus" width="35" />
                      </div>
                      <div class="folder-name">Add new folder</div>
                    </div>
                  </a-col>
                  <a-col
                    v-for="folder in filteredFolders"
                    :key="folder.id"
                    :xs="24"
                    :sm="12"
                    :md="12"
                    :lg="8"
                    :xl="6"
                  >
                    <div
                      class="folder-card"
                      :class="{ active: folder.id === selectedFolderId }"
                      @click="selectFolder(folder.id)"
                      @dblclick.stop="handleOpenFolder(folder.id)"
                    >
                      <div class="folder-card-header">
                        <div class="folder-icon">
                          <IconifyIcon 
                            :icon="hasSubfolders(folder.id) ? 'ph:folders-fill' : 'mdi:folder'" 
                            width="50" 
                            :color="hasSubfolders(folder.id) ? '#82a2bf' : '#E7EAE9'" 
                          />
                        </div>
                        <a-dropdown :trigger="['click']">
                          <IconifyIcon icon="mdi:dots-vertical" width="20" class="folder-menu-icon" @click.stop />
                          <template #overlay>
                            <a-menu>
                              <a-menu-item key="rename" @click="openRenameModal(folder)">Rename</a-menu-item>
                              <a-menu-item v-if="isSuperAdminOrAdministrator" key="delete" @click="handleDeleteFolder(folder)">Delete</a-menu-item>
                            </a-menu>
                          </template>
                        </a-dropdown>
                      </div>
                      <div class="folder-name">{{ folder.name }}</div>
                      <div class="folder-meta">
                        <span>{{ getFolderFileCount(folder.id) }} Files</span>
                        <span>{{ getFolderSize(folder.id) }}</span>
                      </div>
                    </div>
                  </a-col>
                </a-row>
              </div>
              <div v-else class="folders-list">
                <a-table
                  :data-source="filteredFolders"
                  :pagination="false"
                  size="small"
                  row-key="id"
                  :custom-row="(record: DocumentFolder) => ({
                    onClick: () => selectFolder(record.id),
                    onDblclick: () => handleOpenFolder(record.id),
                    style: { cursor: 'pointer' }
                  })"
                >
                  <a-table-column title="Name" key="name">
                    <template #default="{ record }">
                      <div style="display: flex; align-items: center; gap: 8px;">
                        <IconifyIcon 
                          :icon="hasSubfolders(record.id) ? 'ph:folders-fill' : 'mdi:folder'" 
                          width="20" 
                          :color="hasSubfolders(record.id) ? '#82a2bf' : '#E7EAE9'" 
                        />
                        <a @click.prevent="selectFolder(record.id)">{{ record.name }}</a>
                      </div>
                    </template>
                  </a-table-column>
                  <a-table-column title="Files" key="files">
                    <template #default="{ record }">
                      {{ getFolderFileCount(record.id) }}
                    </template>
                  </a-table-column>
                  <a-table-column title="Size" key="size">
                    <template #default="{ record }">
                      {{ getFolderSize(record.id) }}
                    </template>
                  </a-table-column>
                  <a-table-column title="Last Edited" key="updated_at">
                    <template #default="{ record }">
                      {{ record.updated_at ? dayjs(record.updated_at).format('YYYY-MM-DD HH:mm') : '-' }}
                    </template>
                  </a-table-column>
                  <a-table-column title="Actions" key="actions">
                    <template #default="{ record }">
                      <a @click.stop="openRenameModal(record)">Rename</a>
                      <template v-if="isSuperAdminOrAdministrator">
                        <a-divider type="vertical" />
                        <a @click.stop="handleDeleteFolder(record)">Delete</a>
                      </template>
                    </template>
                  </a-table-column>
                </a-table>
              </div>
            </transition>
              </div>
            </div>
          </a-card>

          <a-card class="table-card" :bordered="false">
            <div class="table-header">
              <div class="title">Recently Files</div>
              <a-input
                v-model:value="recentSearch"
                size="small"
                placeholder="Cari file..."
                allow-clear
                style="width: 200px;"
              >
                <template #prefix>
                  <IconifyIcon icon="mdi:magnify" width="16" />
                </template>
              </a-input>
            </div>
            <a-table
              :data-source="pagedFiles"
              :loading="tableLoading || pageLoading"
              size="small"
              :pagination="{
                current: tablePage,
                pageSize: tablePageSize,
                total: totalDocuments,
                showSizeChanger: false,
                showQuickJumper: false,
              }"
              :columns="tableColumns"
              row-key="id"
              :custom-row="(record: DocumentItem) => ({
                onClick: () => handleRowClick(record),
                style: { cursor: 'pointer' }
              })"
              @change="handleTableChange"
              :scroll="{ x: 700 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'name'">
                  <div class="file-name-cell">
                    <IconifyIcon icon="mdi:file-document-outline" width="20" style="margin-right: 8px; color: #666;" />
                    <span>{{ record.name || record.title || 'Untitled' }}</span>
                  </div>
                </template>
                <template v-else-if="column.key === 'mime_type'">
                  <a-tag :color="getFileTypeColor(record.mime_type)">
                    {{ getFileTypeLabel(record.mime_type) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'uploader_id'">
                  {{ getUploaderName(record.uploader_id) }}
                </template>
                <template v-else-if="column.key === 'metadata_status'">
                  <div class="metadata-status-cell">
                    <div class="meta-action">
                      <a-button
                        v-if="hasCompleteMetadata(record)"
                        type="default"
                        size="small"
                        style="background: #f5f5f5; border-color: #d9d9d9;"
                        @click.stop="handleRowClick(record)"
                      >
                        Meta Data ✓
                      </a-button>
                      <a-button
                        v-else
                        type="primary"
                        size="small"
                        style="background: #52c41a; border-color: #52c41a;"
                        @click.stop="handleRowClick(record)"
                      >
                        Lengkapi Meta Data →
                      </a-button>
                    </div>
                    <div class="meta-date">
                      <IconifyIcon icon="lets-icons:time" width="14"/>: {{ formatDateTime(record.updated_at || record.created_at) }}
                    </div>
                  </div>
                </template>
              </template>
            </a-table>
          </a-card>
        </a-col>

        <a-col :xs="24" :lg="4" :xl="4" class="right-column1">
          <a-card class="detail-card" :bordered="false" v-if="selectedFolder">
            <div class="detail-icon">
              <IconifyIcon icon="mdi:folder" width="100" style="color: #E7EAE9;" />
            </div>
            <div class="detail-title">Details</div>
            <div class="detail-row">
              <span>Name</span>
              <strong>{{ selectedFolder.name }}</strong>
            </div>
            <div class="detail-row">
              <span>Type</span>
              <strong>File Folder</strong>
            </div>
            <div class="detail-row">
              <span>Size</span>
              <strong>{{ getSelectedFolderSize() }}</strong>
            </div>
            <div class="detail-row">
              <span>Last Edited</span>
              <strong>{{ getSelectedFolderLastEdited() }}</strong>
            </div>
          </a-card>

          <a-card class="storage-card" :bordered="false">
            <div v-if="pageLoading || loading" class="storage-skeleton">
              <a-skeleton active :paragraph="{ rows: 4 }" />
            </div>
            <template v-else>
              <div class="storage-header">
                <div class="title">My Storage</div>
                <a-dropdown :trigger="['click']">
                  <IconifyIcon icon="mdi:dots-vertical" width="20" class="storage-menu-icon" />
                  <template #overlay>
                    <a-menu>
                      <a-menu-item key="upgrade">Upgrade Storage</a-menu-item>
                      <a-menu-item key="settings">Storage Settings</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
              <div v-if="storageUsage.isAlmostFull" class="storage-warning">
                <IconifyIcon icon="mdi:alert-circle" width="20" style="color: #ff4d4f; margin-right: 8px;" />
                <span>Your Storage is almost full</span>
              </div>
              <div class="storage-body">
                <div class="storage-chart">
                  <a-progress
                    type="circle"
                    :percent="storageUsage.percent"
                    :width="200"
                    stroke-color="#ff4d4f"
                    :format="() => ''"
                  />
                  <div class="storage-chart-center">
                    <div class="storage-percent">{{ storageUsage.percent }}%</div>
                    <div class="storage-size">{{ storageUsage.text }}</div>
                  </div>
                </div>
                <div class="storage-legend">
                  <div
                    v-for="([type, val], idx) in storageBreakdown"
                    :key="type"
                    class="legend-item"
                  >
                    <span class="dot" :class="legendColorClass(type, idx)"></span>
                    <span class="legend-label">
                      {{ type.toUpperCase() }} · {{ val.count }} file(s) · {{ formatBytes(val.size) }}
                    </span>
                  </div>
                </div>
              </div>
            </template>
          </a-card>
        </a-col>
      </a-row>

      <a-modal
        v-model:open="renameFolderModalVisible"
        title="Rename Folder"
        ok-text="Simpan"
        cancel-text="Batal"
        @ok="handleRenameFolder"
      >
        <a-input
          v-model:value="renameFolderName"
          placeholder="Nama folder baru"
          @pressEnter="handleRenameFolder"
        />
      </a-modal>
    </div>

    <a-modal
      v-model:open="addFolderModalVisible"
      title="Tambah Folder"
      ok-text="Buat"
      cancel-text="Batal"
      @ok="handleCreateFolder"
    >
      <a-form layout="vertical">
        <a-form-item label="Nama Folder" required>
          <a-input v-model:value="newFolderName" placeholder="Masukkan nama folder" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="searchModalVisible"
      title="Cari File"
      :footer="null"
      cancel-text="Tutup"
      @cancel="searchModalVisible = false"
    >
      <a-input
        v-model:value="searchQuery"
        placeholder="Ketik nama file..."
        allow-clear
      />
      <div v-if="searchResults.length" class="search-dropdown">
        <div
          v-for="item in searchResults"
          :key="item.id"
          class="search-result"
          @click="handleSelectSearchResult(item.id)"
        >
          <div class="search-result-name">{{ item.name || item.file_name }}</div>
          <div class="search-result-meta">
            <span class="muted">{{ getFolderName(item.folder_id) }}</span>
            <span class="muted">{{ getFolderSize(item.folder_id || '') }}</span>
          </div>
        </div>
      </div>
      <div v-else class="search-empty">Tidak ada hasil</div>
    </a-modal>
  </div>
</template>

<style scoped lang="scss">
.documents-layout {
  min-height: 100vh;
  // background: #f7f8fb;
}

.documents-content {
  max-width: 100%;
  margin: 0 auto;
  padding: 16px 24px 40px;
}

.hero-card {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.hero-left h1 {
  margin: 0 0 4px 0;
}

.hero-left p {
  margin: 0;
  color: #666;
}

.new-btn {
  background: #d8342c;
  border-color: #d8342c;
}

.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.search-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.search-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-icon {
  color: #666;
  cursor: pointer;
  transition: color 0.2s;
}

.action-icon:hover {
  color: #035cab;
}

.new-folder-btn {
  width: 100%;
  height: 40px;
  margin-bottom: 16px;
  background: #db241b;
  border-color: #db241b;
  font-weight: 500;
}

.new-folder-btn:hover {
  background: #c41d1a;
  border-color: #c41d1a;
}

.nav-links {
  margin-top: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-link {
  text-align: left;
  height: 40px;
  padding: 0 12px;
  border-radius: 6px;
  transition: all 0.2s;
}

.nav-link:hover {
  background: #f5f5f5;
}

.folders-card,
.table-card,
.activity-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.all-folders{
  // background-color: orange;
  max-height: 425px;
  position: relative;
  display: flex;
  flex-direction: column;
  
  &.folders-expanded {
    max-height: 100vh;
    height: 100vh;
  }
  
  :deep(.ant-card-body) {
    display: flex;
    flex-direction: column;
    padding: 24px;
    height: 100%;
    max-height: 100%;
    overflow: hidden;
  }
  
  .folders-scroll-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    max-height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    position: relative;
    
    &::-webkit-scrollbar {
      display: none;
    }
  }
  
  .folders-content {
    flex: 1;
    min-height: 0;
    padding-top: 0;
  }
  
  // Untuk collapsed state
  &:not(.folders-expanded) .folders-scroll-container {
    max-height: 425px;
  }
  
  // Untuk expanded state
  &.folders-expanded .folders-scroll-container {
    max-height: 100vh;
  }
}

.activity-card {
  background: #fff;
}

.meta-date{
  font-size: 11px;
  display: flex;
  align-items: center;
  margin-top: 4px;
  justify-content: start;
  color: #777;
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.activity-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #333;
}

.see-all-btn {
  padding: 0;
  height: auto;
  font-size: 14px;
  color: #035cab;
}

.activity-list {
  position: relative;
}

.activity-timeline {
  position: relative;
  padding-left: 24px;
}

.activity-item {
  position: relative;
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 20px;
}

.activity-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
}

.activity-line {
  position: absolute;
  left: 11px;
  top: 40px;
  bottom: -20px;
  width: 2px;
  background: repeating-linear-gradient(
    to bottom,
    #d9d9d9 0px,
    #d9d9d9 4px,
    transparent 4px,
    transparent 8px
  );
}

.activity-item:last-child .activity-line {
  display: none;
}

.activity-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 16px;
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

.activity-loading,
.activity-empty {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

.activity-skeleton {
  padding: 16px;
}

.clickable {
  cursor: pointer;
}

.folders-skeleton {
  padding: 16px;
}

.folders-row-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 10px;
  margin-top: 16px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

.detail-skeleton {
  padding: 16px;
}

.storage-skeleton {
  padding: 16px;
}

.search-dropdown {
  margin-top: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  max-height: 240px;
  overflow-y: auto;
}

.search-result {
  padding: 10px 12px;
  cursor: pointer;
}

.search-result:hover {
  background: #f5f5f5;
}

.search-result-name {
  font-weight: 600;
}

.search-result-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #888;
}

.search-empty {
  margin-top: 12px;
  color: #888;
}

.activity-card {
  margin-top: 16px;
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
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #333;
}

.see-all-btn {
  padding: 0;
  height: auto;
  font-size: 13px;
  color: #035cab;
}

.activity-list {
  position: relative;
}

.activity-timeline {
  position: relative;
  padding-left: 24px;
}

.activity-item {
  position: relative;
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 20px;
}

.activity-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
}

.activity-line {
  position: absolute;
  left: 11px;
  top: 40px;
  bottom: -20px;
  width: 2px;
  background: repeating-linear-gradient(
    to bottom,
    #d9d9d9 0px,
    #d9d9d9 4px,
    transparent 4px,
    transparent 8px
  );
}

.activity-item:last-child .activity-line {
  display: none;
}

.activity-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 16px;
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

.activity-loading,
.activity-empty {
  text-align: center;
  padding: 40px 20px;
  color: #999;
  font-size: 13px;
}

.folders-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  position: sticky;
  top: 0;
  background: #fff;
  z-index: 10;
  padding: 0 0 12px 0;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
}

.folders-actions {
  display: flex;
  gap: 8px;
}

.folders-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 16px;
}

.folder-card {
  border: 1px solid #e8e8e8;
  padding: 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
  position: relative;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
}

.folder-card.add-folder-card {
  border: 2px dashed #d9d9d9;
  background: #fafafa;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 124px;
}

.folder-card.add-folder-card:hover {
  border-color: #035cab;
  background: #f0f6ff;
}

.folder-card.active {
  border-color: #035cab;
  background: #f0f6ff;
  box-shadow: 0 2px 8px rgba(3, 92, 171, 0.1);
}

.folder-card:hover:not(.add-folder-card) {
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
  color: #035cab;
  flex-shrink: 0;
}

.folder-icon-large {
  color: #999;
  margin-bottom: 12px;
}

.folder-menu-icon {
  color: #999;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.folder-card:hover .folder-menu-icon {
  opacity: 1;
}

.folder-name {
  font-weight: 500;
  line-height: 19px;
  font-size: 13px;
  color: #333;
  // margin-bottom: 8px;
  word-break: break-word;
}

.folder-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
  gap: 8px;
}

.table-card {
  margin-top: 12px;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.table-header .title {
  font-weight: 600;
  font-size: 16px;
}

.file-name-cell {
  display: flex;
  align-items: center;
}

:deep(.ant-table) {
  .ant-table-thead > tr > th {
    font-size: 13px;
  }
  .ant-table-tbody > tr > td {
    font-size: 13px;
  }
}

/* Table row hover effect */
:deep(.ant-table-tbody) {
  tr {
    transition: background-color 0.2s;
    
    &:hover {
      background-color: #f5f5f5 !important;
    }
  }
}

.detail-icon {
  text-align: center;
  // margin-bottom: 16px;
}

.detail-title {
  font-weight: 700;
  font-size: 14px;
  margin-bottom: 6px;
  color: #333;
}

.detail-row {
  display: block;
  // justify-content: space-between;
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

.storage-card {
  margin-top: 16px;
}

.storage-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.storage-header .title {
  font-weight: 600;
  font-size: 16px;
}

.storage-menu-icon {
  color: #999;
  cursor: pointer;
}

.storage-warning {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #fff1f0;
  border: 1px solid #ffccc7;
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #ff4d4f;
}

.storage-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.storage-chart {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.storage-chart-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.storage-percent {
  font-size: 20px;
  font-weight: 700;
  color: #ff4d4f;
  margin-bottom: 4px;
}

.storage-size {
  font-size: 12px;
  color: #666;
}

.storage-legend {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 13px;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.dot.red { background: #ff4d4f; }
.dot.blue { background: #82a2bf; }
.dot.orange { background: #faad14; }


</style>
