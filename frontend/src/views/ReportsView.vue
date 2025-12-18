<template>
  <div class="reports-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="reports-content">
      <!-- Page Header Section -->
      <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">Reports</h1>
            <p class="page-description">
              Manage all reports and financial data.
            </p>
          </div>
          <!-- <div class="header-right">
            <a-button type="default" class="upload-btn" @click="handleUploadReport">
              <IconifyIcon icon="mdi:file-excel" width="16" style="margin-right: 8px;" />
              Upload Report
            </a-button>
            <a-button type="primary" class="add-report-btn" @click="handleAddReport">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Add report
            </a-button>
          </div> -->
        </div>
      </div>

      <!-- Main Content -->
      <div class="mainContentPage">
        <!-- Bulk Upload Section -->
        <a-card class="bulk-upload-card" :bordered="false" style="margin-bottom: 24px;">
          <FinancialReportBulkUpload @upload-success="handleBulkUploadSuccess" />
        </a-card>

        <a-card class="reports-table-card" :bordered="false">
          <!-- Table Filters and Actions -->
          <div class="table-filters-container">
            <a-input v-model:value="searchText" placeholder="Search subsidiary" class="search-input"
              allow-clear>
              <template #prefix>
                <IconifyIcon icon="mdi:magnify" width="16" />
              </template>
            </a-input>
            <a-select
              v-model:value="currentYear"
              style="width: 120px;"
              @change="handleYearChange"
            >
              <a-select-option v-for="year in availableYears" :key="year" :value="year">
                {{ year }}
              </a-select-option>
            </a-select>
          </div>

          <a-table :columns="columns" :data-source="filteredSubsidiaries" :pagination="{
            current: currentPage,
            pageSize: pageSize,
            total: filteredSubsidiaries.length,
            showSizeChanger: true,
            showTotal: (total: number) => `Total ${total} subsidiaries`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }" :loading="loading" row-key="company.id" :scroll="{ x: 'max-content' }" class="striped-table"
            @change="handleTableChange" :locale="{ emptyText: 'Tidak ada data subsidiaries' }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'company_name'">
                <div class="name-cell">
                  <div class="table-logo-cell">
                    <img v-if="getCompanyLogo(record.company)" :src="getCompanyLogo(record.company)"
                      :alt="record.company?.name || ''" class="table-logo" />
                    <div v-else class="table-logo-placeholder"
                      :style="{ backgroundColor: getIconColor(record.company?.name || '') }">
                      {{ getCompanyInitial(record.company?.name || '') }}
                    </div>
                  </div>
                  <div class="company-name-wrapper">
                    <span class="company-name">{{ record.company?.name || 'Unknown Company' }}</span>
                  </div>
                </div>
              </template>
              <template v-else-if="column.key && column.key.startsWith('month_')">
                <a-tag :color="record.monthlyStatus[parseInt(column.key.split('_')[1]!)] ? 'success' : 'default'">
                  {{ record.monthlyStatus[parseInt(column.key.split('_')[1]!)] ? 'Ada' : 'Belum' }}
                </a-tag>
              </template>
              <template v-if="column.key === 'actions'">
                <a-button type="link" size="small" @click="handleViewSubsidiary(record)">
                  <IconifyIcon icon="mdi:eye" width="16" style="margin-right: 4px;" />
                  View
                </a-button>
              </template>
            </template>
          </a-table>

        </a-card>
      </div>
    </div>

    <!-- Upload Report Modal -->
    <a-modal
      v-model:open="uploadModalVisible"
      title="Upload Report"
      :width="900"
      :mask-closable="false"
    >
      <div class="upload-modal-content">
        <!-- File Input (Hidden) -->
        <input
          ref="fileInputRef"
          type="file"
          accept=".xlsx,.xls"
          style="display: none"
          @change="handleFileSelect"
        />

        <!-- File Selection -->
        <div v-if="!selectedFile && !uploading" class="file-selection">
          <a-button type="primary" @click="triggerFileSelect" :loading="validating">
            <IconifyIcon icon="mdi:file-upload" width="16" style="margin-right: 8px;" />
            Pilih File Excel
          </a-button>
          <p class="file-hint">Format: .xlsx atau .xls</p>
        </div>

        <!-- Preview and Validation -->
        <div v-if="selectedFile && !uploading && validationResult" class="preview-section">
          <div class="file-info">
            <IconifyIcon icon="mdi:file-excel" width="24" style="color: #52c41a; margin-right: 8px;" />
            <span class="file-name">{{ selectedFile.name }}</span>
            <a-button type="link" size="small" @click="clearFile">
              <IconifyIcon icon="mdi:close" width="16" />
            </a-button>
          </div>

          <!-- Validation Summary -->
          <div class="validation-summary" v-if="validationResult">
            <a-alert
              v-if="validationResult.errors && validationResult.errors.length === 0"
              type="success"
              message="Semua data valid. Siap untuk diupload."
              show-icon
              style="margin-bottom: 16px;"
            />
            <a-alert
              v-else-if="validationResult.errors && validationResult.errors.length > 0"
              type="error"
              :message="`Ditemukan ${validationResult.errors.length} error. Harap perbaiki sebelum upload.`"
              show-icon
              style="margin-bottom: 16px;"
            />
          </div>

          <!-- Data Preview Table -->
          <div class="data-preview" v-if="validationResult && validationResult.data">
            <h4>Preview Data ({{ validationResult.data.length }} baris)</h4>
            <a-table
              :columns="previewColumns"
              :data-source="validationResult.data"
              :pagination="{ pageSize: 10 }"
              :scroll="{ x: 'max-content', y: 300 }"
              size="small"
              class="preview-table"
            >
              <template #bodyCell="{ column, index }">
                <template v-if="column.key === 'row_number'">
                  {{ index + 1 }}
                </template>
                <template v-if="column.key === 'status'">
                  <a-tag v-if="getRowErrors(index + 1).length === 0" color="success">Valid</a-tag>
                  <a-tag v-else color="error">Error</a-tag>
                </template>
              </template>
            </a-table>
          </div>

          <!-- Error Details -->
          <div v-if="validationResult && validationResult.errors && validationResult.errors.length > 0" class="error-details">
            <h4>Detail Error:</h4>
            <a-list :data-source="validationResult.errors" size="small">
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta>
                    <template #title>
                      Baris {{ item.row }} - {{ item.column }}
                    </template>
                    <template #description>
                      {{ item.message }}
                    </template>
                  </a-list-item-meta>
                </a-list-item>
              </template>
            </a-list>
          </div>
        </div>

        <!-- Upload Progress -->
        <div v-if="uploading" class="upload-progress">
          <a-spin size="large" />
          <div class="progress-info">
            <p>Mengupload data...</p>
            <a-progress :percent="uploadProgress" :status="uploadProgress === 100 ? 'success' : 'active'" />
            <p class="progress-text">{{ uploadProgress }}%</p>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <template #footer>
        <div class="modal-footer">
          <a-button @click="closeUploadModal">Batal</a-button>
          <a-button
            v-if="selectedFile && !uploading && !validating && validationResult"
            type="primary"
            :disabled="!validationResult.valid || (validationResult.errors && validationResult.errors.length > 0)"
            @click="handleUpload"
            :loading="uploading"
          >
            <IconifyIcon icon="mdi:upload" width="16" style="margin-right: 8px;" />
            Upload
          </a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import FinancialReportBulkUpload from '../components/FinancialReportBulkUpload.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import type { TableColumnsType } from 'ant-design-vue'
import reportsApi, { type Report, type ValidationRow } from '../api/reports'
import { companyApi, type Company } from '../api/userManagement'
import { financialReportsApi, type FinancialReport } from '../api/financialReports'
import { logger } from '../utils/logger'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const totalReports = ref(0)
const totalPages = ref(0)

// Reports data from backend (keeping for compatibility)
const reportsData = ref<Report[]>([])
const companies = ref<Company[]>([])

// Subsidiaries data with monthly report status
interface SubsidiaryReportStatus {
  company: Company
  monthlyStatus: Record<number, boolean> // month (1-12) -> has report
  reports: FinancialReport[] // cached reports for this company
}

const subsidiariesWithStatus = ref<SubsidiaryReportStatus[]>([])
const currentYear = ref<string>(new Date().getFullYear().toString())

// Available years (current year and 5 years before)
const availableYears = computed(() => {
  const current = new Date().getFullYear()
  const years: string[] = []
  for (let i = 0; i <= 5; i++) {
    years.push((current - i).toString())
  }
  return years
})

// Filters
const filterCompanyIds = ref<string[]>([])
const filterPeriod = ref<string | undefined>(undefined)
const searchText = ref<string>('')

// Available periods removed - no longer used with new table structure

// Icon colors for companies
const iconColors = [
  '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
  '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80'
]

// Get company logo atau generate icon
const getCompanyLogo = (company: Company | undefined): string | undefined => {
  if (company?.logo) {
    const apiURL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
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
const getIconColor = (name: string | undefined): string => {
  if (!name) return iconColors[0]!
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return iconColors[hash % iconColors.length]!
}

// Format period to readable format (2025-09 -> September 2025)
const formatPeriod = (period: string | undefined): string => {
  if (!period) return 'Unknown'
  const [year, month] = period.split('-')
  if (!year || !month) return period
  const months = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ]
  const monthIndex = parseInt(month, 10) - 1
  if (monthIndex < 0 || monthIndex >= months.length) return period
  return `${months[monthIndex]} ${year}`
}

// Calculate RKAP percentage (dummy calculation for now)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const calculateRKAPPercent = (_report: Report): number => {
  // Dummy calculation - in real app, this would come from RKAP data
  return Math.floor(Math.random() * 100)
}

// calculateFinancialScore removed - not used in new table structure

// Load reports from backend
const loadReports = async () => {
  loading.value = true
  try {
    const params: { page: number; page_size: number; company_id?: string; period?: string } = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (filterCompanyIds.value && filterCompanyIds.value.length > 0) {
      // Support multiple company IDs - send as array or comma-separated string
      params.company_id = filterCompanyIds.value.join(',')
    }
    if (filterPeriod.value) {
      params.period = filterPeriod.value
    }

    const response = await reportsApi.getAll(params)
    reportsData.value = response.data
    totalReports.value = response.total
    totalPages.value = response.total_pages

    // Check if no data found and show appropriate message
    if (response.data.length === 0 && (filterPeriod.value || (filterCompanyIds.value && filterCompanyIds.value.length > 0))) {
      let filterMessage = ''
      if (filterPeriod.value && filterCompanyIds.value.length > 0) {
        filterMessage = `Data di periode ${formatPeriod(filterPeriod.value)} untuk perusahaan yang dipilih tidak ditemukan`
      } else if (filterPeriod.value) {
        filterMessage = `Data di periode ${formatPeriod(filterPeriod.value)} tidak ditemukan`
      } else if (filterCompanyIds.value.length > 0) {
        filterMessage = 'Data untuk perusahaan yang dipilih tidak ditemukan'
      }
      if (filterMessage) {
        message.info(filterMessage)
      }
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number; data?: { message?: string } }; message?: string }
    const status = axiosError.response?.status

    // Check if it's a 404 (no data) or empty response
    if (status === 404) {
      // Clear table data - this is normal when no reports exist
      reportsData.value = []
      totalReports.value = 0
      totalPages.value = 0
      // Don't show error message for empty data, let table show empty state
      return
    }

    // Check if it's a network error
    if (axiosError.message && axiosError.message.includes('Network Error')) {
      // Clear table data so previous results are not shown when nothing matches filters
      reportsData.value = []
      totalReports.value = 0
      totalPages.value = 0

      // Check if we have filters applied
      if (filterPeriod.value || (filterCompanyIds.value && filterCompanyIds.value.length > 0)) {
        let filterMessage = ''
        if (filterPeriod.value && filterCompanyIds.value.length > 0) {
          filterMessage = `Data di periode ${formatPeriod(filterPeriod.value)} untuk perusahaan yang dipilih tidak ditemukan`
        } else if (filterPeriod.value) {
          filterMessage = `Data di periode ${formatPeriod(filterPeriod.value)} tidak ditemukan`
        } else if (filterCompanyIds.value.length > 0) {
          filterMessage = 'Data untuk perusahaan yang dipilih tidak ditemukan'
        }
        message.info(filterMessage)
      } else {
        message.error('Gagal memuat reports: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
      }
    } else {
      message.error('Gagal memuat reports: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    }
  } finally {
    loading.value = false
  }
}

// Load companies for filter
const loadCompanies = async () => {
  try {
    companies.value = await companyApi.getAll()
  } catch (error: unknown) {
    logger.error('Failed to load companies:', error)
  }
}

// Load subsidiaries with monthly report status
const loadSubsidiariesWithStatus = async () => {
  loading.value = true
  // Clear existing data first to ensure fresh data
  subsidiariesWithStatus.value = []
  try {
    // Get all active companies
    const allCompanies = await companyApi.getAll()
    const activeCompanies = allCompanies.filter(c => c.is_active)
    
    // For each company, get financial reports for current year
    const subsidiariesData: SubsidiaryReportStatus[] = await Promise.all(
      activeCompanies.map(async (company) => {
        try {
          // Get all financial reports for this company
          const reports = await financialReportsApi.getByCompanyId(company.id)
          
          // Filter only realisasi reports (not RKAP) for current year
          // Use currentYear.value to ensure reactive value is used
          const selectedYear = currentYear.value
          const yearReports = reports.filter(r => 
            !r.is_rkap && 
            r.year === selectedYear &&
            r.period && 
            r.period.startsWith(selectedYear)
          )
          
          // Build monthly status map (1-12) using selectedYear
          const monthlyStatus: Record<number, boolean> = {}
          for (let month = 1; month <= 12; month++) {
            const period = `${selectedYear}-${month.toString().padStart(2, '0')}`
            monthlyStatus[month] = yearReports.some(r => r.period === period)
          }
          
          return {
            company,
            monthlyStatus,
            reports: yearReports,
          }
        } catch (error) {
          logger.error(`Failed to load reports for company ${company.id}:`, error)
          // Return with all months as false if error
          const monthlyStatus: Record<number, boolean> = {}
          for (let month = 1; month <= 12; month++) {
            monthlyStatus[month] = false
          }
          return {
            company,
            monthlyStatus,
            reports: [],
          }
        }
      })
    )
    
    subsidiariesWithStatus.value = subsidiariesData
  } catch (error: unknown) {
    const axiosError = error as { message?: string }
    message.error('Gagal memuat data subsidiaries: ' + (axiosError.message || 'Unknown error'))
    subsidiariesWithStatus.value = []
  } finally {
    loading.value = false
  }
}

// Watch for filter changes
watch([filterCompanyIds, filterPeriod], () => {
  currentPage.value = 1
  loadReports()
})

// Watch searchText is handled by computed filteredSubsidiaries (no need for async handler)

// Note: Pagination changes are handled by handleTableChange
// Watch is removed to prevent double loading

// Old dummy data (removed, now using backend data)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _oldDummyData = [
  {
    id: '1',
    name: 'PT Petro Nusantara Laporan Bulan Januari',
    rkap_percent: 10,
    revenue: 426,
    npat: -294,
    opex: 110000,
    dividend: 19,
    financial_score: 'D',
    periode: 'January',
    icon: 'mdi:factory',
    iconColor: '#FFA07A', // orange/yellow
  },
  {
    id: '2',
    name: 'PT Pertamina Marine Laporan Bulan Februari',
    rkap_percent: 61,
    revenue: 583,
    npat: -185,
    opex: 110000,
    dividend: 12,
    financial_score: 'D+',
    periode: 'February',
    icon: 'mdi:ship',
    iconColor: '#BB8FCE', // purple
  },
  {
    id: '3',
    name: 'PT Energi Abadi Raya Laporan Bulan Maret',
    rkap_percent: 12,
    revenue: 647,
    npat: -783,
    opex: 110000,
    dividend: 22,
    financial_score: 'D',
    periode: 'January',
    icon: 'mdi:lightning-bolt',
    iconColor: '#BB8FCE', // purple
  },
  {
    id: '4',
    name: 'PT Geo Minyak Bumi Laporan Bulan April',
    rkap_percent: 19,
    revenue: 883,
    npat: -12,
    opex: 110000,
    dividend: 61,
    financial_score: 'D',
    periode: 'February',
    icon: 'mdi:oil',
    iconColor: '#FF6B6B', // red
  },
  {
    id: '5',
    name: 'PT Mitra Fuel Logistik Laporan Bulan Mei',
    rkap_percent: 34,
    revenue: 816,
    npat: -311,
    opex: 110000,
    dividend: 13,
    financial_score: 'D+',
    periode: 'February',
    icon: 'mdi:truck',
    iconColor: '#52BE80', // green
  },
  {
    id: '6',
    name: 'PT Gas Surya Mandala Laporan Bulan Juni',
    rkap_percent: 43,
    revenue: 600,
    npat: -8,
    opex: 110000,
    dividend: 34,
    financial_score: 'D',
    periode: 'January',
    icon: 'mdi:gas-station',
    iconColor: '#52BE80', // green
  },
  {
    id: '7',
    name: 'PT Kilang Nusantara Laporan Bulan Juli',
    rkap_percent: 13,
    revenue: 177,
    npat: -3,
    opex: 110000,
    dividend: 43,
    financial_score: 'D',
    periode: 'February',
    icon: 'mdi:factory',
    iconColor: '#FFA07A', // orange/yellow
  },
  {
    id: '8',
    name: 'PT Transport Energi Laporan Bulan Agustus',
    rkap_percent: 21,
    revenue: 196,
    npat: -820,
    opex: 110000,
    dividend: 10,
    financial_score: 'D',
    periode: 'April',
    icon: 'mdi:train',
    iconColor: '#BB8FCE', // purple
  },
] // Old dummy data - not used anymore

// allReportsForSearch and filteredReports removed - using filteredSubsidiaries instead

// Computed for filtered subsidiaries (with search)
const filteredSubsidiaries = computed(() => {
  let filtered = subsidiariesWithStatus.value

  // Apply search filter
  if (searchText.value && searchText.value.trim()) {
    const search = searchText.value.toLowerCase().trim()
    filtered = filtered.filter(item => {
      const companyName = item.company.name || ''
      return companyName.toLowerCase().includes(search)
    })
  }

  return filtered
})

// Handle year change
const handleYearChange = (value: string) => {
  // Update currentYear value
  currentYear.value = value
  // Reset page to 1 when year changes
  currentPage.value = 1
  // Reload data with new year (no need to await, let it run in background)
  loadSubsidiariesWithStatus()
}

// Table columns with filters and sorters
// Month names for column headers
const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

// Table columns with monthly status
const columns: TableColumnsType = [
  {
    title: 'Company Name',
    key: 'company_name',
    dataIndex: 'company',
    width: 250,
    fixed: 'left',
    sorter: (a: SubsidiaryReportStatus, b: SubsidiaryReportStatus) => {
      const nameA = a.company.name || ''
      const nameB = b.company.name || ''
      return nameA.localeCompare(nameB)
    },
  },
  // Add columns for each month (1-12)
  ...monthNames.map((monthName, index) => {
    const month = index + 1
    return {
      title: monthName,
      key: `month_${month}`,
      dataIndex: `month_${month}`,
      width: 80,
      align: 'center' as const,
      sorter: (a: SubsidiaryReportStatus, b: SubsidiaryReportStatus) => {
        const statusA = a.monthlyStatus[month] ? 1 : 0
        const statusB = b.monthlyStatus[month] ? 1 : 0
        return statusA - statusB
      },
    }
  }),
  {
    title: 'Actions',
    key: 'actions',
    width: 100,
    fixed: 'right',
  },
]

// Visible pages - maksimal 4 halaman sesuai gambar
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _visiblePages = computed(() => {
  const maxVisible = 4
  const pages: number[] = []

  if (totalPages.value <= maxVisible) {
    // Jika total halaman <= 4, tampilkan semua
    for (let i = 1; i <= totalPages.value; i++) {
      pages.push(i)
    }
  } else {
    // Jika lebih dari 4, tampilkan 4 halaman pertama
    for (let i = 1; i <= maxVisible; i++) {
      pages.push(i)
    }
  }

  return pages
})

// formatNumber and getScoreClass removed - not used in new table structure

// Handlers
const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// handleExportPDF and handleExportExcel removed - not used in new table structure

// Upload states
const uploadModalVisible = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const validating = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
// templateLoading removed - now handled by FinancialReportBulkUpload component
const validationResult = ref<{
  valid: boolean
  errors: Array<{ row: number; column: string; message: string }>
  data: ValidationRow[]
} | null>(null)

// Preview columns for upload modal
const previewColumns = [
  { title: 'No', key: 'row_number', width: 60 },
  { title: 'Periode', dataIndex: 'period', key: 'period' },
  { title: 'Company Code', dataIndex: 'company_code', key: 'company_code' },
  { title: 'Revenue', dataIndex: 'revenue', key: 'revenue' },
  { title: 'OPEX', dataIndex: 'opex', key: 'opex' },
  { title: 'NPAT', dataIndex: 'npat', key: 'npat' },
  { title: 'Dividend', dataIndex: 'dividend', key: 'dividend' },
  { title: 'Financial Ratio', dataIndex: 'financial_ratio', key: 'financial_ratio' },
  { title: 'Status', key: 'status', width: 80 },
]

// handleUploadReport removed - not used (replaced by FinancialReportBulkUpload component)

const triggerFileSelect = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Validate file extension
  const validExtensions = ['.xlsx', '.xls']
  const fileExtension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!validExtensions.includes(fileExtension)) {
    message.error('Format file tidak valid. Hanya file Excel (.xlsx, .xls) yang diperbolehkan.')
    return
  }

  selectedFile.value = file
  validating.value = true
  uploadProgress.value = 0

  try {
    // Validate file
    const result = await reportsApi.validateExcelFile(file)
    
    // Ensure result has required structure
    if (!result.errors) {
      result.errors = []
    }
    if (!result.data) {
      result.data = []
    }
    
    validationResult.value = result

    if (result.valid) {
      message.success('File valid. Siap untuk diupload.')
    } else {
      message.warning(`Ditemukan ${result.errors.length} error. Harap perbaiki sebelum upload.`)
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memvalidasi file: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    clearFile()
  } finally {
    validating.value = false
  }
}

const getRowErrors = (rowNumber: number): Array<{ row: number; column: string; message: string }> => {
  if (!validationResult.value) return []
  return validationResult.value.errors.filter(e => e.row === rowNumber)
}

const handleUpload = async () => {
  if (!selectedFile.value || !validationResult.value || validationResult.value.errors.length > 0) {
    message.error('Tidak dapat upload. Harap perbaiki semua error terlebih dahulu.')
    return
  }

  uploading.value = true
  uploadProgress.value = 0

  try {
    const result = await reportsApi.uploadReports(selectedFile.value, (progress) => {
      uploadProgress.value = progress
    })

    if (result.errors.length > 0) {
      message.warning(`Upload selesai dengan ${result.failed} error. ${result.success} data berhasil diupload.`)
    } else {
      message.success('Upload selesai.')
    }

    // Reload reports
    await loadReports()
    
    // Close modal after delay
    setTimeout(() => {
      closeUploadModal()
    }, 1500)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal upload: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    uploading.value = false
    uploadProgress.value = 0
  }
}

// handleDownloadTemplate removed - now handled by FinancialReportBulkUpload component

const clearFile = () => {
  selectedFile.value = null
  validationResult.value = null
  uploadProgress.value = 0
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

const closeUploadModal = () => {
  uploadModalVisible.value = false
  clearFile()
  uploading.value = false
}

// handleAddReport removed - not used in new table structure

const handleBulkUploadSuccess = () => {
  message.success('Bulk upload berhasil! Data akan diperbarui.')
  // Reload subsidiaries with status
  loadSubsidiariesWithStatus()
}

// handleView, handleEdit, handleDelete removed - not used in new table structure

const handleViewSubsidiary = (record: SubsidiaryReportStatus) => {
  // Redirect ke halaman detail subsidiary
  router.push(`/subsidiaries/${record.company.id}`)
}

// Navigation functions (kept for potential future use)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _goToFirstPage = () => {
  currentPage.value = 1
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _goToLastPage = () => {
  currentPage.value = totalPages.value
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _goToPage = (page: number) => {
  currentPage.value = page
}

// Handle table change (pagination, sorting, filtering)
const handleTableChange = (pag: { current?: number; pageSize?: number }) => {
  if (pag.current !== undefined && pag.current !== currentPage.value) {
    currentPage.value = pag.current
  }

  if (pag.pageSize !== undefined && pag.pageSize !== pageSize.value) {
    pageSize.value = pag.pageSize
    currentPage.value = 1 // Reset to first page when page size changes
  }
  // Pagination is handled client-side by filteredSubsidiaries computed
}

// Load data on mount
onMounted(async () => {
  await Promise.all([
    loadSubsidiariesWithStatus(),
    loadCompanies(),
  ])
})
</script>

<style scoped>
.reports-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.search-input {
  width: 300px;
}

.reports-content {
  background: #f5f5f5;
  margin: 0 auto;
}

/* Page Header Layout */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  width: 100%;
  gap: 16px;
}

.header-left {
  flex: 1;
  min-width: 0;
}

/* Header Right - untuk tombol action */
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.export-btn {
  width: 40px;
  height: 40px;
  min-width: 40px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  color: #666;
  transition: all 0.3s ease;
}

.export-btn:hover {
  background: #f5f5f5;
  color: #035CAB;
}

.export-btn:active {
  transform: scale(0.95);
}

/* Ensure all filter elements have same height */
/* .table-filters-container :deep(.ant-input),
.table-filters-container :deep(.ant-select-selector) {
  height: 40px !important;
} */

/* .table-filters-container :deep(.ant-select-selection-item),
.table-filters-container :deep(.ant-select-selection-placeholder) {
  line-height: 38px !important;
} */

/* Styling untuk selection overflow item - perlu :deep() karena scoped */
/* .table-filters-container :deep(.ant-select-selection-overflow-item){
  background: red !important;
  line-height: 0 !important;
  padding: 0 !important;
  height: 20px;
  margin: 0 !important;
} */

.action-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  color: #666;
}

.action-btn:hover {
  background: #f5f5f5;
  color: #035CAB;
}

.upload-btn {
  height: 40px;
  padding: 0 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
}

.add-report-btn {
  height: 40px;
  padding: 0 15px;
  font-weight: 500;
  border-radius: 8px;
  /* box-shadow: 0 2px 8px rgba(3, 92, 171, 0.2); */
  display: flex;
  justify-content: center;
  align-items: center;
}

/* Main Content */
.mainContentPage {
  padding: 24px 32px;
  max-width: 100%;
}

.reports-table-card {
  border-radius: 12px;
  /* box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08); */
  overflow: hidden;
  background: white;
}

.reports-table-card :deep(.ant-card-body) {
  padding: 24px;
}

/* Table Filters Container */
.table-filters-container {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.search-input {
  flex: 1;
  min-width: 200px;
  max-width: 300px;
  height: 40px;
}

.filter-select {
  min-width: 180px;
}

.filter-select:first-of-type {
  min-width: 200px;
}

/* Ensure filter select has consistent height */
.table-filters-container :deep(.filter-select .ant-select-selector) {
  height: 40px !important;
}

.table-filters-container :deep(.filter-select .ant-select-selection-item),
.table-filters-container :deep(.filter-select .ant-select-selection-placeholder) {
  line-height: 38px !important;
}

/* Styling untuk selection item di select multiple - perlu :deep() karena scoped */
/* .filter-select :deep(.ant-select-selection-overflow-item) {
  background: red !important;
} */

.filter-select :deep(.ant-select-selection-overflow-item .ant-select-selection-item) {
  position: relative;
  display: flex;
  flex: none;
  justify-content: center;
  align-items: center;
  border-radius: 5px;

}

/* Atau dengan selector yang lebih spesifik */
/* .table-filters-container :deep(.ant-select-selection-overflow-item) {
  background: red !important;
}

.table-filters-container :deep(.ant-select-selection-overflow-item .ant-select-selection-item) {
  background: red !important;
} */

.export-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

/* Ensure all filter elements have same height */
/* .table-filters-container :deep(.ant-input),
.table-filters-container :deep(.ant-select-selector) {
  height: 40px !important;
} */

/* .table-filters-container :deep(.ant-select-selection-item),
.table-filters-container :deep(.ant-select-selection-placeholder) {
  line-height: 38px !important;
} */

/* Table Styles */
.name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-logo-cell {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.table-logo {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  object-fit: cover;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.table-logo-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 18px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.company-name {
  font-weight: 500;
  color: #1a1a1a;
}

.financial-score-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-weight: 600;
  font-size: 16px;
  background: #ffebee;
  color: #c62828;
  border: none;
}

.financial-score-badge.score-d {
  background: #ffebee;
  color: #c62828;
}

.financial-score-badge.score-dplus {
  background: #ffebee;
  color: #c62828;
}

.actions-cell {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.action-dropdown-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  background: transparent;
  border-radius: 4px;
}

.action-dropdown-btn:hover {
  background: #f5f5f5;
}

/* Table responsive */
:deep(.ant-table) {
  font-size: 14px;
}

:deep(.ant-table-thead > tr > th) {
  background: #fafafa;
  font-weight: 600;
  color: #1a1a1a;
  border-bottom: 2px solid #e8e8e8;
}

:deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid #f0f0f0;
  padding: 16px;
}

/* Striped table styles */
.reports-table-card .striped-table :deep(.ant-table-tbody > tr:nth-child(even) > td) {
  background-color: #fafafa !important;
}

.reports-table-card .striped-table :deep(.ant-table-tbody > tr:nth-child(odd) > td) {
  background-color: #ffffff !important;
}

.reports-table-card .striped-table :deep(.ant-table-tbody > tr:hover > td) {
  background-color: #e6f7ff !important;
}

:deep(.ant-table-thead > tr > th) {
  padding: 16px;
}

/* Custom Pagination - sesuai gambar dengan First/Last */
.custom-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  padding: 16px 0;
}

.pagination-link {
  padding: 0 12px;
  height: auto;
  color: #035CAB;
  font-weight: 500;
}

.pagination-link:hover:not(:disabled) {
  color: #024a8f;
}

.pagination-link:disabled {
  color: #d9d9d9;
  cursor: not-allowed;
}

.pagination-number {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  font-weight: 500;
  border: 1px solid transparent;
  transition: all 0.3s;
}

.pagination-number:hover {
  color: #035CAB;
  border-color: #035CAB;
}

.pagination-number.active {
  background: #035CAB;
  color: white;
  border-color: #035CAB;
}

/* Responsive */
@media (max-width: 768px) {
  .mainContentPage {
    padding: 16px 20px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-right {
    width: 100%;
    justify-content: flex-start;
    margin-top: 16px;
  }

  .table-filters-container {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input {
    max-width: 100%;
    width: 100%;
  }

  .filter-select {
    width: 100%;
    min-width: 100%;
  }

  .export-buttons {
    margin-left: 0;
    width: 100%;
    justify-content: flex-end;
  }

  /* Mobile: same height for all filter elements */
  /* .table-filters-container :deep(.ant-input),
  .table-filters-container :deep(.ant-select-selector) {
    height: 40px !important;
  } */



  .export-btn {
    height: 40px;
    padding: 0 12px;
  }

  .action-btn {
    width: 36px;
    height: 36px;
    min-width: 36px;
  }

  .upload-btn,
  .add-report-btn {
    height: 40px;
    font-size: 14px;
    padding: 0 16px;
  }

  .reports-table-card :deep(.ant-card-body) {
    padding: 16px;
  }

  :deep(.ant-table) {
    font-size: 12px;
  }

  .name-cell {
    gap: 8px;
  }

  .table-logo-cell,
  .table-logo,
  .table-logo-placeholder {
    width: 32px;
    height: 32px;
  }

  .table-logo-placeholder {
    font-size: 14px;
  }

  .company-name {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .header-right {
    flex-direction: column;
    align-items: stretch;
  }

  .table-filters-container {
    gap: 8px;
  }

  .export-buttons {
    justify-content: space-between;
  }

  .action-btn {
    width: 100%;
    justify-content: center;
  }

  .upload-btn,
  .add-report-btn {
    width: 100%;
  }
}

/* Upload Modal Styles */
.upload-modal-content {
  min-height: 400px;
}

.file-selection {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.file-hint {
  margin-top: 16px;
  color: #999;
  font-size: 14px;
}

.preview-section {
  margin-top: 16px;
}

.file-info {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 8px;
  margin-bottom: 16px;
}

.file-name {
  flex: 1;
  font-weight: 500;
  margin-left: 8px;
}

.validation-summary {
  margin-bottom: 16px;
}

.data-preview {
  margin-bottom: 24px;
}

.data-preview h4 {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 600;
}

.preview-table {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.error-details {
  margin-top: 24px;
  padding: 16px;
  background: #fff1f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
}

.error-details h4 {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 600;
  color: #cf1322;
}

.upload-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.progress-info {
  margin-top: 24px;
  width: 100%;
  max-width: 400px;
}

.progress-info p {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 500;
}

.progress-text {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
