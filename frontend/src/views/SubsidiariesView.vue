<template>
  <div class="subsidiaries-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="subsidiaries-content">
      <!-- Header Section -->

      <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">Subsidiary</h1>
            <p class="page-description">
              Overview of key financial metrics and performance indicators for all subsidiaries.
            </p>
          </div>
          <div class="header-right">
            <a-input v-model:value="searchText" placeholder="Search" allow-clear class="search-input" size="large">
              <template #prefix>
                <IconifyIcon icon="mdi:magnify" width="20" />
              </template>
            </a-input>
            <div class="view-mode-buttons">
              <a-button 
                :type="viewMode === 'grid' ? 'primary' : 'default'"
                size="large"
                @click="handleViewModeChange('grid')"
                class="view-mode-btn"
              >
                <IconifyIcon icon="mdi:view-grid" width="20" />
              </a-button>
              <a-button 
                :type="viewMode === 'list' ? 'primary' : 'default'"
                size="large"
                @click="handleViewModeChange('list')"
                class="view-mode-btn"
              >
                <IconifyIcon icon="mdi:view-list" width="20" />
              </a-button>
            </div>
            <a-button type="primary" size="large" @click="handleCreateCompany" class="add-button">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Add new Subsidiary
            </a-button>
          </div>
        </div>
      </div>

      <div class="mainContentPage">
        <!-- Subsidiary Cards Grid -->
        <div class="subsidiary-cards-grid" v-if="viewMode === 'grid' && !companiesLoading && filteredCompanies.length > 0">
          <div v-for="company in paginatedCompanies" :key="company.id" class="subsidiary-card"
            @click="handleViewDetail(company.id)">
            <!-- Card Header -->
            <div class="card-header">
              <div class="company-icon">
                <img v-if="getCompanyLogo(company)" :src="getCompanyLogo(company)" :alt="company.name"
                  class="logo-image" />
                <div v-else class="icon-placeholder" :style="{ backgroundColor: getIconColor(company.name) }">
                  {{ getCompanyInitial(company.name) }}
                </div>
              </div>
              <div class="company-info">
                <h3 class="company-name">{{ company.name }}</h3>
                <p class="company-reg">No Reg {{ company.nib || 'N/A' }}</p>
              </div>
            </div>

            <!-- Card Divider -->
            <div class="card-divider"></div>

            <!-- Card Content -->
            <div class="card-content">
              <div class="latest-month-header">
                <IconifyIcon icon="mdi:information-outline" width="16" style="margin-right: 4px;" />
                <span>Latest Month</span>
              </div>

              <div class="metrics-row">
                <!-- RKAP vs Realization -->
                <div class="metric-item">
                  <div class="metric-value">{{ formatCurrency(getRKAPData(company.id)) }}</div>
                  <div class="metric-meta">
                    <span class="metric-year">{{ getRKAPYear(company.id) }}</span>
                    <span class="metric-change positive">+{{ getRKAPChange(company.id) }}%</span>
                  </div>
                  <div class="metric-label">RKAP vs Realization</div>
                </div>

                <!-- Opex Trend -->
                <div class="metric-item">
                  <div class="metric-value">{{ formatCurrency(getOpexData(company.id)) }}</div>
                  <div class="metric-meta">
                    <span class="metric-quarter">{{ getOpexQuarter(company.id) }}</span>
                    <span class="metric-change negative">-{{ getOpexChange(company.id) }}%</span>
                  </div>
                  <div class="metric-label">Opex Trend</div>
                </div>
              </div>
            </div>

            <!-- Card Footer -->
            <div class="card-footer">
              <a-button type="link" class="learn-more-btn">
                Learn more
                <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
              </a-button>
            </div>
          </div>
        </div>

        <!-- Subsidiary Table View -->
        <div v-if="viewMode === 'list'">
          <a-table
            :columns="tableColumns"
            :data-source="tableData"
            :loading="companiesLoading || tableDataLoading"
            :pagination="{
              current: tablePagination.current,
              pageSize: tablePagination.pageSize,
              total: tablePagination.total,
              showSizeChanger: true,
              showTotal: (total: number) => `Total ${total} subsidiaries`,
              pageSizeOptions: ['10', '20', '50', '100'],
            }"
            @change="handleTableChange"
            row-key="id"
            :scroll="{ x: 'max-content' }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'logo'">
                <div class="table-logo-cell">
                  <img v-if="getCompanyLogo(record)" :src="getCompanyLogo(record)" :alt="record.name" class="table-logo" />
                  <div v-else class="table-logo-placeholder" :style="{ backgroundColor: getIconColor(record.name) }">
                    {{ getCompanyInitial(record.name) }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'level'">
                <a-tag :color="getLevelColor(record.level)">
                  {{ getLevelLabel(record.level) }}
                </a-tag>
              </template>
              <template v-if="column.key === 'status'">
                <a-tag :color="record.is_active ? 'green' : 'red'">
                  {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
                </a-tag>
              </template>
              <template v-if="column.key === 'actions'">
                <a-dropdown>
                  <a-button type="link" size="small">
                    Aksi
                    <IconifyIcon icon="mdi:chevron-down" width="16" style="margin-left: 4px;" />
                  </a-button>
                  <template #overlay>
                    <a-menu>
                      <a-menu-item key="view" @click="handleViewDetail(record.id)">
                        <IconifyIcon icon="mdi:eye" width="16" style="margin-right: 8px;" />
                        Lihat Detail
                      </a-menu-item>
                      <a-menu-item key="edit" @click="handleEditCompany(record.id)">
                        <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                        Edit
                      </a-menu-item>
                      <a-menu-item key="assign-role" @click="handleAssignRole(record.id)">
                        <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                        Assign Role
                      </a-menu-item>
                      <a-menu-divider />
                      <a-menu-item key="delete" danger @click="handleDeleteCompany(record.id)">
                        <IconifyIcon icon="mdi:delete" width="16" style="margin-right: 8px;" />
                        Hapus
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </template>
            </template>
          </a-table>
        </div>

        <!-- Loading State -->
        <div v-if="companiesLoading && viewMode === 'grid'" class="loading-container">
          <a-spin size="large" />
        </div>

        <!-- Empty State -->
        <div v-if="viewMode === 'grid' && !companiesLoading && companies.length === 0" class="empty-state">
          <IconifyIcon icon="mdi:office-building-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
          <p>Belum ada data subsidiary</p>
          <a-button type="primary" @click="handleCreateCompany">
            <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
            Tambah Subsidiary Pertama
          </a-button>
        </div>

        <!-- No Search Results -->
        <div v-if="viewMode === 'grid' && !companiesLoading && companies.length > 0 && filteredCompanies.length === 0" class="empty-state">
          <IconifyIcon icon="mdi:magnify" width="64" style="color: #ccc; margin-bottom: 16px;" />
          <p>Tidak ada hasil untuk "{{ searchText }}"</p>
          <a-button type="default" @click="searchText = ''">Hapus Filter</a-button>
        </div>

        <!-- Pagination for Grid View -->
        <div v-if="viewMode === 'grid' && filteredCompanies.length > 0" class="pagination-container">
          <a-pagination v-model:current="currentPage" v-model:page-size="pageSize" :total="filteredCompanies.length"
            :show-total="(total: number) => `Total ${total} subsidiaries`" :page-size-options="['8', '16', '24', '32']"
            show-size-changer @change="handlePageChange" />
        </div>
      </div>



    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, type Company } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

// View Mode: 'grid' or 'list' - load from localStorage
const getStoredViewMode = (): 'grid' | 'list' => {
  const stored = localStorage.getItem('subsidiaries-view-mode')
  return (stored === 'grid' || stored === 'list') ? stored : 'grid'
}

const viewMode = ref<'grid' | 'list'>(getStoredViewMode())

// Companies
const companies = ref<Company[]>([])
const companiesLoading = ref(false)
const searchText = ref('')

// Table data loading state
const tableDataLoading = ref(false)

// Pagination
const currentPage = ref(1)
const pageSize = ref(8)

// Table Pagination
const tablePagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

// Sample financial data (RKAP & Opex) - akan diganti dengan data real jika ada
const financialData = ref<Record<string, {
  rkap: { value: number; year: string; change: number }
  opex: { value: number; quarter: string; change: number }
}>>({})

// Computed untuk filtered companies berdasarkan search, diurutkan berdasarkan waktu (paling baru di atas)
const filteredCompanies = computed(() => {
  let filtered = companies.value

  // Apply search filter
  if (searchText.value.trim()) {
  const search = searchText.value.toLowerCase().trim()
    filtered = companies.value.filter(company =>
    company.name.toLowerCase().includes(search) ||
    company.code.toLowerCase().includes(search) ||
    (company.short_name && company.short_name.toLowerCase().includes(search)) ||
    (company.nib && company.nib.toLowerCase().includes(search)) ||
    (company.description && company.description.toLowerCase().includes(search))
  )
  }

  // Sort by updated_at (most recent first), fallback to created_at
  return filtered.sort((a, b) => {
    const dateA = new Date(a.updated_at || a.created_at || 0).getTime()
    const dateB = new Date(b.updated_at || b.created_at || 0).getTime()
    return dateB - dateA // Descending order (newest first)
  })
})

// Computed untuk paginated companies dari filtered companies
const paginatedCompanies = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredCompanies.value.slice(start, end)
})

// Watch search text untuk reset pagination
watch(searchText, () => {
  currentPage.value = 1
})

// Generate financial data untuk company
const generateFinancialData = (companyId: string) => {
  if (!financialData.value[companyId]) {
    // Generate random but consistent data berdasarkan company ID
    const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
    const baseValue = 100 + (hash % 100)
    const baseOpex = 50 + (hash % 50)

    financialData.value[companyId] = {
      rkap: {
        value: baseValue * 1000000, // dalam juta
        year: '2025',
        change: 10 + (hash % 10)
      },
      opex: {
        value: baseOpex * 1000000,
        quarter: `Q${1 + (hash % 4)} 2024`,
        change: 3 + (hash % 5)
      }
    }
  }
  return financialData.value[companyId]
}

const getRKAPData = (companyId: string): number => {
  return generateFinancialData(companyId).rkap.value
}

const getRKAPYear = (companyId: string): string => {
  return generateFinancialData(companyId).rkap.year
}

const getRKAPChange = (companyId: string): number => {
  return generateFinancialData(companyId).rkap.change
}

const getOpexData = (companyId: string): number => {
  return generateFinancialData(companyId).opex.value
}

const getOpexQuarter = (companyId: string): string => {
  return generateFinancialData(companyId).opex.quarter
}

const getOpexChange = (companyId: string): number => {
  return generateFinancialData(companyId).opex.change
}

// Format currency
const formatCurrency = (value: number): string => {
  if (value >= 1000000000) {
    return `$${(value / 1000000000).toFixed(0)}B`
  } else if (value >= 1000000) {
    return `${(value / 1000000).toFixed(0)}M`
  } else if (value >= 1000) {
    return `${(value / 1000).toFixed(0)}K`
  }
  return `$${value.toFixed(0)}`
}

// Get company logo atau generate icon
const getCompanyLogo = (company: Company): string | undefined => {
  if (company.logo) {
    const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const baseURL = apiURL.replace(/\/api\/v1$/, '')
    return company.logo.startsWith('http') ? company.logo : `${baseURL}${company.logo}`
  }
  return undefined
}

// Get company initial untuk icon placeholder
const getCompanyInitial = (name: string): string => {
  const trimmed = name.trim()
  if (!trimmed) return '??'

  const words = trimmed.split(/\s+/).filter(w => w.length > 0)
  if (words.length >= 2) {
    const first = words[0]?.[0]
    const second = words[1]?.[0]
    if (first && second) {
      return (first + second).toUpperCase()
    }
  }
  const firstTwo = trimmed.substring(0, 2)
  return firstTwo ? firstTwo.toUpperCase() : '??'
}

// Get icon color berdasarkan nama company
const getIconColor = (name: string): string => {
  const colors: string[] = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
    '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80'
  ]
  if (!name) return colors[0]!
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[hash % colors.length]!
}

const loadCompanies = async () => {
  companiesLoading.value = true
  try {
    companies.value = await companyApi.getAll()
    // Generate financial data untuk semua companies
    companies.value.forEach(company => {
      generateFinancialData(company.id)
    })
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    companiesLoading.value = false
  }
}

const handleCreateCompany = () => {
  router.push('/subsidiaries/new')
}

const handleViewDetail = (id: string) => {
  router.push(`/subsidiaries/${id}`)
}

const handlePageChange = () => {
  // Scroll to top saat ganti page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// View Mode Handler
const handleViewModeChange = async (mode: 'grid' | 'list') => {
  viewMode.value = mode
  // Save to localStorage
  localStorage.setItem('subsidiaries-view-mode', mode)
  
  // Lazy load table data only when switching to list view
  // Check if companies are already loaded, if not load them
  if (mode === 'list' && companies.value.length === 0) {
    await loadTableData()
  } else if (mode === 'list') {
    // If companies are already loaded, just update pagination
    tablePagination.value.total = filteredCompanies.value.length
    tablePagination.value.current = 1
  }
}

// Computed untuk table data dengan pagination
const tableData = computed(() => {
  const start = (tablePagination.value.current - 1) * tablePagination.value.pageSize
  const end = start + tablePagination.value.pageSize
  return filteredCompanies.value.slice(start, end)
})

// Load table data (lazy loading) - hanya set loading state
const loadTableData = async () => {
  if (companies.value.length === 0) {
    await loadCompanies()
  }
  
  tableDataLoading.value = true
  try {
    // Update pagination total
    tablePagination.value.total = filteredCompanies.value.length
    // Reset to first page
    tablePagination.value.current = 1
  } catch (error) {
    message.error('Gagal memuat data table')
  } finally {
    tableDataLoading.value = false
  }
}

// Watch for changes in filtered companies to update table pagination
watch([filteredCompanies, viewMode], () => {
  if (viewMode.value === 'list') {
    tablePagination.value.total = filteredCompanies.value.length
    // Reset to first page if current page is out of bounds
    const maxPage = Math.ceil(filteredCompanies.value.length / tablePagination.value.pageSize)
    if (tablePagination.value.current > maxPage && maxPage > 0) {
      tablePagination.value.current = 1
    }
  }
})

// Table Columns
const tableColumns: TableColumnsType = [
  {
    title: 'Logo',
    key: 'logo',
    width: 80,
    fixed: 'left',
  },
  {
    title: 'Nama Perusahaan',
    dataIndex: 'name',
    key: 'name',
    sorter: (a: Company, b: Company) => a.name.localeCompare(b.name),
    width: 250,
  },
  {
    title: 'Kode',
    dataIndex: 'code',
    key: 'code',
    sorter: (a: Company, b: Company) => a.code.localeCompare(b.code),
    width: 120,
  },
  {
    title: 'NIB',
    dataIndex: 'nib',
    key: 'nib',
    sorter: (a: Company, b: Company) => (a.nib || '').localeCompare(b.nib || ''),
    width: 150,
  },
  {
    title: 'Tingkat',
    dataIndex: 'level',
    key: 'level',
    sorter: (a: Company, b: Company) => a.level - b.level,
    width: 150,
    filters: [
      { text: 'Holding (Induk)', value: 0 },
      { text: 'Anak Perusahaan', value: 1 },
      { text: 'Cucu Perusahaan', value: 2 },
      { text: 'Cicit Perusahaan', value: 3 },
    ],
    onFilter: (value: number, record: Company) => record.level === value,
  },
  {
    title: 'Status',
    dataIndex: 'is_active',
    key: 'status',
    width: 120,
    filters: [
      { text: 'Aktif', value: true },
      { text: 'Tidak Aktif', value: false },
    ],
    onFilter: (value: boolean, record: Company) => record.is_active === value,
  },
  {
    title: 'Tanggal Dibuat',
    dataIndex: 'created_at',
    key: 'created_at',
    sorter: (a: Company, b: Company) => {
      const dateA = new Date(a.created_at || 0).getTime()
      const dateB = new Date(b.created_at || 0).getTime()
      return dateA - dateB
    },
    width: 180,
    customRender: ({ text }: { text: string }) => {
      if (!text) return '-'
      const date = new Date(text)
      return date.toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    },
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 120,
    fixed: 'right',
  },
]

// Table Change Handler
const handleTableChange: TableProps['onChange'] = (pagination) => {
  if (pagination) {
    tablePagination.value.current = pagination.current || 1
    tablePagination.value.pageSize = pagination.pageSize || 10
  }
}

// Get Level Label
const getLevelLabel = (level: number): string => {
  switch (level) {
    case 0:
      return 'Holding (Induk)'
    case 1:
      return 'Anak Perusahaan'
    case 2:
      return 'Cucu Perusahaan'
    case 3:
      return 'Cicit Perusahaan'
    default:
      return `Level ${level}`
  }
}

// Get Level Color
const getLevelColor = (level: number): string => {
  switch (level) {
    case 0:
      return 'purple'
    case 1:
      return 'blue'
    case 2:
      return 'cyan'
    case 3:
      return 'green'
    default:
      return 'default'
  }
}

// Action Handlers
const handleEditCompany = (id: string) => {
  router.push(`/subsidiaries/${id}/edit`)
}

const handleAssignRole = (id: string) => {
  router.push(`/subsidiaries/${id}`)
  // TODO: Open assign role modal in detail page
}

const handleDeleteCompany = (id: string) => {
  Modal.confirm({
    title: 'Hapus Subsidiary',
    content: 'Apakah Anda yakin ingin menghapus subsidiary ini? Tindakan ini tidak dapat dibatalkan.',
    okText: 'Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      try {
        await companyApi.delete(id)
        message.success('Subsidiary berhasil dihapus')
        await loadCompanies()
        if (viewMode.value === 'list') {
          await loadTableData()
        }
      } catch (error: unknown) {
        const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
        message.error('Gagal menghapus subsidiary: ' + (axiosError.response?.data?.message || axiosError.message))
      }
    },
  })
}

onMounted(async () => {
  await loadCompanies()
  
  // If view mode is 'list', load table data
  if (viewMode.value === 'list') {
    await loadTableData()
  }
})
</script>

<style scoped>
.subsidiaries-layout {
  min-height: 100vh;
  /* background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%); */
  /* background-image: 
    radial-gradient(circle at 20% 50%, rgba(120, 119, 198, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 80%, rgba(255, 119, 198, 0.1) 0%, transparent 50%); */
}

.subsidiaries-content {
  /* max-width: 1400px; */
  margin: 0 auto;
  /* padding: 32px 24px; */
}

/* Header Section */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
  flex-wrap: wrap;
  gap: 16px;
  width: 100%;
}

.header-left {
  flex: 1;
  min-width: 300px;
}

.page-title {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1a1a1a;
  line-height: 1.2;
}

.page-description {
  font-size: 16px;
  color: #666;
  margin: 0;
  line-height: 1.5;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.search-input {
  width: 300px;
}

.search-input :deep(.ant-input) {
  border-radius: 8px;
}

.view-mode-buttons {
  display: flex;
  gap: 8px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  padding: 4px;
  background: #fafafa;
}

.view-mode-btn {
  height: 36px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.add-button {
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(3, 92, 171, 0.2);
}

/* Cards Grid */
.subsidiary-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
  margin-bottom: 32px;
}

.subsidiary-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.subsidiary-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

/* Card Header */
.card-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.company-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.icon-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 24px;
  font-weight: 700;
  border-radius: 12px;
}

.company-info {
  flex: 1;
  min-width: 0;
}

.company-name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #1a1a1a;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.company-reg {
  font-size: 13px;
  color: #999;
  margin: 0;
}

/* Card Divider */
.card-divider {
  height: 1px;
  background: #e8e8e8;
  margin: 16px 0;
}

/* Card Content */
.card-content {
  flex: 1;
}

.latest-month-header {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #666;
  margin-bottom: 16px;
}

.metrics-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.metric-item {
  display: flex;
  flex-direction: column;
}

.metric-value {
  font-size: 28px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 8px;
  line-height: 1.2;
}

.metric-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.metric-year,
.metric-quarter {
  color: #666;
}

.metric-change {
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.metric-change.positive {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.1);
}

.metric-change.negative {
  color: #ff4d4f;
  background: rgba(255, 77, 79, 0.1);
}

.metric-label {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* Card Footer */
.card-footer {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.learn-more-btn {
  padding: 0;
  height: auto;
  font-weight: 500;
  color: #035CAB;
}

.learn-more-btn:hover {
  color: #024a8f;
}

/* Loading & Empty States */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.empty-state {
  text-align: center;
  padding: 64px 24px;
  color: #999;
}

.empty-state p {
  font-size: 16px;
  margin-bottom: 16px;
}

/* Pagination */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 32px;
  padding: 24px 0;
}

/* Responsive */
@media (max-width: 768px) {
  .subsidiary-cards-grid {
    grid-template-columns: 1fr;
  }

  .page-header {
    flex-direction: column;
  }

  .page-title {
    font-size: 28px;
  }

  .metrics-row {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

@media (min-width: 769px) and (max-width: 1024px) {
  .subsidiary-cards-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1025px) and (max-width: 1440px) {
  .subsidiary-cards-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1441px) {
  .subsidiary-cards-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* Table View Styles */
.table-logo-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.table-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.table-logo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
  font-weight: 600;
}
</style>
