<template>
  <div class="subsidiary-detail-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="detail-content">
      <!-- Loading State -->
      <div v-if="loading" class="loading-container">
        <a-spin size="large" />
      </div>

      <!-- Company Detail -->
      <div v-else-if="company" class="detail-card">
        <div class="page-header-container container"
          style="display:flex; flex-direction: column; flex-wrap: wrap;justify-content: center; align-items: start;">
          <div>
            <!-- Back Button -->
            <a-button type="text" @click="handleBack" class="back-button">
              <IconifyIcon icon="mdi:arrow-left" width="20" style="margin-right: 8px;" />
              Kembali ke Daftar Subsidiary
            </a-button>
          </div>

          <!-- Header Section -->
          <div class="detail-header">
            <div class="company-icon-large">
              <img v-if="getCompanyLogo(company)" :src="getCompanyLogo(company)" :alt="company.name"
                class="logo-image-large" />
              <div v-else class="icon-placeholder-large" :style="{ backgroundColor: getIconColor(company.name) }">
                {{ getCompanyInitial(company.name) }}
              </div>
            </div>
            <div class="header-info">
              <h1 class="company-title">{{ company.name }}</h1>
              <p class="company-subtitle">{{ company.short_name || company.name }}</p>
              <div class="company-meta">
                <a-tag :color="company.is_active ? 'green' : 'red'">
                  {{ company.is_active ? 'Aktif' : 'Tidak Aktif' }}
                </a-tag>
                <a-tag :color="getLevelColor(company.level)">
                  {{ getLevelLabel(company.level) }}
                </a-tag>
                <span v-if="company.code" class="meta-item">Kode: {{ company.code }}</span>
                <span v-if="company.nib" class="meta-item">No Reg {{ company.nib }}</span>
              </div>
            </div>
            <div class="header-actions">
              <a-space>
                <a-button>
                  <IconifyIcon icon="mdi:file-pdf-box" width="16" style="margin-right: 8px;" />
                  PDF
                </a-button>
                <a-button>
                  <IconifyIcon icon="mdi:chart-box" width="16" style="margin-right: 8px;" />
                </a-button>
                <a-date-picker v-model:value="selectedPeriod" picker="month" placeholder="Select Periode"
                  style="width: 150px;" />
                <a-button>
                  <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                  Options
                </a-button>
              </a-space>
            </div>
          </div>
        </div>

        <!-- Tabs -->
        <div class="tabs-container">
          <a-tabs v-model:activeKey="activeTab" type="card" size="large">
            <a-tab-pane key="performance" tab="Performance">
              <!-- Performance Tab Content -->
              <div class="performance-content">
                <!-- Financial Trend Cards -->
                <div class="trend-cards-row">
                  <!-- RKAP vs Realization Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">RKAP vs Realization</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(rkapData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ rkapData.year }}</span>
                          <span class="trend-change positive">+{{ rkapData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="rkapGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#ff9800;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#ff9800;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="rkapChartFillPath" fill="url(#rkapGradient)" class="chart-fill" />
                          <path :d="rkapChartPath" stroke="#ff9800" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>

                  <!-- Opex Trend Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">Opex Trend</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(opexData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ opexData.quarter }}</span>
                          <span class="trend-change negative">-{{ opexData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="opexGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#666;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#666;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="opexChartFillPath" fill="url(#opexGradient)" class="chart-fill" />
                          <path :d="opexChartPath" stroke="#666" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>

                  <!-- NPAT Trend Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">NPAT Trend</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(npatData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ npatData.quarter }}</span>
                          <span class="trend-change positive">+{{ npatData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="npatGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#666;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#666;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="npatChartFillPath" fill="url(#npatGradient)" class="chart-fill" />
                          <path :d="npatChartPath" stroke="#666" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>
                </div>

                <!-- Recent Files and Reports -->
                <div class="recent-section">
                  <!-- Recent Files -->
                  <a-card class="recent-card" :bordered="false">
                    <template #title>
                      <div class="card-header-title">
                        <IconifyIcon icon="mdi:clock-outline" width="20" style="margin-right: 8px;" />
                        <span>Recently Files</span>
                      </div>
                    </template>
                    <template #extra>
                      <a-button type="link" @click="handleManageFiles">
                        Manage file upload
                        <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
                      </a-button>
                    </template>
                    <a-table :columns="fileColumns" :data-source="recentFiles" :pagination="false" :show-header="true"
                      size="small">
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'type'">
                          <a-tag :color="record.type === 'Pdf' ? 'red' : 'green'">{{ record.type }}</a-tag>
                        </template>
                        <template v-if="column.key === 'status'">
                          <a-tag v-if="record.status === 'complete'" color="green">Meta Data ✓</a-tag>
                          <a-button v-else type="link" size="small">
                            Lengkapi Meta Data
                            <IconifyIcon icon="mdi:arrow-right" width="14" style="margin-left: 4px;" />
                          </a-button>
                        </template>
                        <template v-if="column.key === 'action'">
                          <IconifyIcon icon="mdi:chevron-right" width="20" style="color: #999; cursor: pointer;" />
                        </template>
                      </template>
                    </a-table>
                  </a-card>

                  <!-- Recent Reports -->
                  <a-card class="recent-card" :bordered="false">
                    <template #title>
                      <div class="card-header-title">
                        <IconifyIcon icon="mdi:clock-outline" width="20" style="margin-right: 8px;" />
                        <span>Recently Reports</span>
                      </div>
                    </template>
                    <template #extra>
                      <a-button type="link" @click="handleManageReports">
                        Manage Reports
                        <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
                      </a-button>
                    </template>
                    <a-table :columns="reportColumns" :data-source="recentReports" :pagination="false" :show-header="true"
                      size="small">
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'rkap_percent'">
                          {{ record.rkap_percent }}%
                        </template>
                        <template v-if="column.key === 'action'">
                          <IconifyIcon icon="mdi:chevron-right" width="20" style="color: #999; cursor: pointer;" />
                        </template>
                      </template>
                    </a-table>
                  </a-card>
                </div>
              </div>
            </a-tab-pane>

            <a-tab-pane key="profile" tab="Profile">
              <!-- Profile Tab Content -->
              <div class="profile-content">
                <!-- Informasi Dasar -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                    Informasi Dasar
                  </h2>
                  <a-descriptions :column="2" bordered>
                    <a-descriptions-item label="Nama Lengkap">{{ company.name }}</a-descriptions-item>
                    <a-descriptions-item label="Nama Singkat">{{ company.short_name || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Kode Perusahaan">{{ company.code }}</a-descriptions-item>
                    <a-descriptions-item label="Status">{{ company.status || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="NPWP">{{ company.npwp || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="NIB">{{ company.nib || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Deskripsi" :span="2">
                      {{ company.description || '-' }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Informasi Kontak -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:phone" width="20" style="margin-right: 8px;" />
                    Informasi Kontak
                  </h2>
                  <a-descriptions :column="2" bordered>
                    <a-descriptions-item label="Telepon">{{ company.phone || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Fax">{{ company.fax || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Email">{{ company.email || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Website">{{ company.website || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Alamat Perusahaan" :span="2">
                      {{ company.address || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Alamat Operasional" :span="2">
                      {{ company.operational_address || '-' }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Struktur Kepemilikan -->
                <div v-if="company.shareholders && company.shareholders.length > 0" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-group" width="20" style="margin-right: 8px;" />
                    Struktur Kepemilikan
                  </h2>
                  <a-table :columns="shareholderColumns" :data-source="company.shareholders" :pagination="false"
                    row-key="id">
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'ownership_percent'">
                        {{ record.ownership_percent }}%
                      </template>
                      <template v-if="column.key === 'share_count'">
                        {{ record.share_count?.toLocaleString() || '-' }}
                      </template>
                      <template v-if="column.key === 'is_main_parent'">
                        <a-tag v-if="record.is_main_parent" color="blue">Ya</a-tag>
                        <span v-else>-</span>
                      </template>
                    </template>
                  </a-table>
                </div>

                <!-- Bidang Usaha -->
                <div v-if="company.main_business || (company.business_fields && company.business_fields.length > 0)"
                  class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:briefcase" width="20" style="margin-right: 8px;" />
                    Bidang Usaha
                  </h2>
                  <a-descriptions :column="1" bordered>
                    <a-descriptions-item label="Sektor Industri">
                      {{ getMainBusiness(company)?.industry_sector || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="KBLI">
                      {{ getMainBusiness(company)?.kbli || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Uraian Kegiatan Usaha Utama">
                      {{ getMainBusiness(company)?.main_business_activity || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Kegiatan Usaha Tambahan">
                      {{ getMainBusiness(company)?.additional_activities || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Tanggal Mulai Beroperasi">
                      {{ formatDate(getMainBusiness(company)?.start_operation_date) }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Pengurus/Dewan Direksi -->
                <div v-if="company.directors && company.directors.length > 0" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-tie" width="20" style="margin-right: 8px;" />
                    Pengurus/Dewan Direksi
                  </h2>
                  <a-table :columns="directorColumns" :data-source="company.directors" :pagination="false" row-key="id">
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'start_date'">
                        {{ record.start_date ? formatDate(record.start_date) : '-' }}
                      </template>
                    </template>
                  </a-table>
                </div>
              </div>
            </a-tab-pane>
          </a-tabs>
        </div>
      </div>

      <!-- Not Found -->
      <div v-else class="not-found">
        <IconifyIcon icon="mdi:alert-circle-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
        <p>Subsidiary tidak ditemukan</p>
        <a-button type="primary" @click="handleBack">Kembali ke Daftar</a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, type Company, type BusinessField } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import dayjs from 'dayjs'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const company = ref<Company | null>(null)
const loading = ref(false)
const activeTab = ref('performance')
const selectedPeriod = ref<any>(null)

// Dummy data untuk charts
const rkapData = ref({
  value: 120000000,
  year: '2025',
  change: 15
})

const opexData = ref({
  value: 80000000,
  quarter: 'Q1 2024',
  change: 5
})

const npatData = ref({
  value: 25000000,
  quarter: 'Q1 2024',
  change: 15
})

// Generate chart data
const generateChartData = (baseValue: number, variance: number = 0.2) => {
  const points = 12
  const data: number[] = []
  for (let i = 0; i < points; i++) {
    const random = (Math.random() - 0.5) * variance
    data.push(baseValue * (1 + random))
  }
  return data
}

const rkapChartData = computed(() => generateChartData(100, 0.3))
const opexChartData = computed(() => generateChartData(80, 0.2))
const npatChartData = computed(() => generateChartData(25, 0.25))

// Generate SVG path untuk chart
const generateChartPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)
  
  const firstValue = data[0] ?? 0
  let path = `M 0 ${height - ((firstValue - min) / range) * height}`
  for (let i = 1; i < data.length; i++) {
    const value = data[i] ?? 0
    const x = i * stepX
    const y = height - ((value - min) / range) * height
    path += ` L ${x} ${y}`
  }
  return path
}

const generateChartFillPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)
  
  let path = `M 0 ${height}`
  const firstValue = data[0] ?? 0
  path += ` L 0 ${height - ((firstValue - min) / range) * height}`
  for (let i = 1; i < data.length; i++) {
    const value = data[i] ?? 0
    const x = i * stepX
    const y = height - ((value - min) / range) * height
    path += ` L ${x} ${y}`
  }
  path += ` L ${width} ${height} Z`
  return path
}

const rkapChartPath = computed(() => generateChartPath(rkapChartData.value))
const rkapChartFillPath = computed(() => generateChartFillPath(rkapChartData.value))
const opexChartPath = computed(() => generateChartPath(opexChartData.value))
const opexChartFillPath = computed(() => generateChartFillPath(opexChartData.value))
const npatChartPath = computed(() => generateChartPath(npatChartData.value))
const npatChartFillPath = computed(() => generateChartFillPath(npatChartData.value))

// Dummy data untuk recent files
const recentFiles = ref([
  {
    key: '1',
    name: 'RUPS_Tahunan_2025',
    type: 'Pdf',
    lastModified: '2 hours ago',
    status: 'complete'
  },
  {
    key: '2',
    name: 'Laporan_Keuangan_Q1_2024',
    type: 'Excel',
    lastModified: '1 day ago',
    status: 'incomplete'
  },
  {
    key: '3',
    name: 'Dokumen_Legal_2024',
    type: 'Pdf',
    lastModified: '3 days ago',
    status: 'complete'
  }
])

// Dummy data untuk recent reports
const recentReports = ref([
  {
    key: '1',
    name: 'Laporan September',
    rkap_percent: 85,
    revenue: '$120M',
    npat: '$25M',
    opex: '$80M'
  },
  {
    key: '2',
    name: 'Laporan Agustus',
    rkap_percent: 82,
    revenue: '$115M',
    npat: '$23M',
    opex: '$78M'
  },
  {
    key: '3',
    name: 'Laporan Juli',
    rkap_percent: 88,
    revenue: '$125M',
    npat: '$27M',
    opex: '$82M'
  }
])

const fileColumns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', key: 'type' },
  { title: 'Last modified', dataIndex: 'lastModified', key: 'lastModified' },
  { title: '', key: 'status' },
  { title: '', key: 'action', width: 30 }
]

const reportColumns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'RKAP (%)', key: 'rkap_percent' },
  { title: 'Revenue', dataIndex: 'revenue', key: 'revenue' },
  { title: 'NPAT', dataIndex: 'npat', key: 'npat' },
  { title: 'Opex', dataIndex: 'opex', key: 'opex' },
  { title: '', key: 'action', width: 30 }
]

const shareholderColumns = [
  { title: 'Jenis', dataIndex: 'type', key: 'type' },
  { title: 'Nama', dataIndex: 'name', key: 'name' },
  { title: 'Nomor Identitas', dataIndex: 'identity_number', key: 'identity_number' },
  { title: 'Persentase', key: 'ownership_percent' },
  { title: 'Jumlah Saham', key: 'share_count' },
  { title: 'Induk Utama', key: 'is_main_parent' },
]

const directorColumns = [
  { title: 'Jabatan', dataIndex: 'position', key: 'position' },
  { title: 'Nama Lengkap', dataIndex: 'full_name', key: 'full_name' },
  { title: 'KTP', dataIndex: 'ktp', key: 'ktp' },
  { title: 'NPWP', dataIndex: 'npwp', key: 'npwp' },
  { title: 'Tanggal Mulai', key: 'start_date' },
  { title: 'Alamat Domisili', dataIndex: 'domicile_address', key: 'domicile_address' },
]

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

const loadCompany = async () => {
  const id = route.params.id as string
  if (!id) {
    message.error('ID perusahaan tidak valid')
    return
  }

  loading.value = true
  try {
    company.value = await companyApi.getById(id)
    // Generate financial data berdasarkan company ID
    if (company.value) {
      const hash = company.value.id.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
      rkapData.value.value = (100 + (hash % 100)) * 1000000
      rkapData.value.change = 10 + (hash % 10)
      opexData.value.value = (50 + (hash % 50)) * 1000000
      opexData.value.change = 3 + (hash % 5)
      npatData.value.value = (20 + (hash % 30)) * 1000000
      npatData.value.change = 10 + (hash % 10)
    }
  } catch (error: any) {
    message.error('Gagal memuat data perusahaan: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

const getMainBusiness = (company: Company): BusinessField | null => {
  if (company.main_business) {
    return company.main_business
  }
  if (company.business_fields && company.business_fields.length > 0) {
    const mainField = company.business_fields.find((bf: BusinessField) => (bf as any).is_main)
    return mainField || company.business_fields[0] || null
  }
  return null
}

const formatDate = (date: string | undefined): string => {
  if (!date) return '-'
  return dayjs(date).format('DD MMMM YYYY')
}

const getCompanyLogo = (company: Company): string | undefined => {
  if (company.logo) {
    const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const baseURL = apiURL.replace(/\/api\/v1$/, '')
    return company.logo.startsWith('http') ? company.logo : `${baseURL}${company.logo}`
  }
  return undefined
}

const getCompanyInitial = (name: string | undefined): string => {
  if (!name) return '??'
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

const getIconColor = (name: string): string => {
  const colors: string[] = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
    '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80'
  ]
  if (!name) return colors[0]!
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[hash % colors.length]!
}

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

const getLevelColor = (level: number): string => {
  switch (level) {
    case 0:
      return 'red'
    case 1:
      return 'blue'
    case 2:
      return 'green'
    case 3:
      return 'orange'
    default:
      return 'default'
  }
}

const handleBack = () => {
  router.push('/subsidiaries')
}

const handleEdit = () => {
  if (company.value) {
    router.push(`/subsidiaries/${company.value.id}/edit`)
  }
}

const handleManageFiles = () => {
  message.info('Manage files feature coming soon')
}

const handleManageReports = () => {
  message.info('Manage reports feature coming soon')
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

onMounted(() => {
  loadCompany()
})
</script>

<style scoped>
.subsidiary-detail-layout {
  min-height: 100vh;
}

.detail-content {
  margin: 0 auto;
}

.back-button {
  margin-bottom: 24px;
  padding: 0;
  height: auto;
}

.loading-container,
.not-found {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  text-align: center;
}

.not-found p {
  font-size: 16px;
  color: #999;
  margin-bottom: 16px;
}

/* Detail card styles moved to child components */

.detail-header {
  display: flex;
  align-items: flex-start;
  gap: 24px;
  width: 100%;
}

.company-icon-large {
  width: 120px;
  height: 120px;
  border-radius: 16px;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-image-large {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.icon-placeholder-large {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 48px;
  font-weight: 700;
  border-radius: 16px;
}

.header-info {
  flex: 1;
  min-width: 0;
}

.company-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1a1a1a;
}

.company-subtitle {
  font-size: 18px;
  color: #666;
  margin: 0 0 16px 0;
}

.company-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.meta-item {
  font-size: 14px;
  color: #666;
}

.header-actions {
  flex-shrink: 0;
}

/* Tabs Container */
.tabs-container {
  margin-top: 24px;
}

.tabs-container :deep(.ant-tabs-card) {
  background: transparent;
}

.tabs-container :deep(.ant-tabs-tab) {
  border-radius: 8px 8px 0 0;
}

.tabs-container :deep(.ant-tabs-tab-active) {
  background: white;
}

/* Performance Content */
.performance-content {
  padding: 24px;
  background: white;
  border-radius: 0 8px 8px 8px;
}

.trend-cards-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 32px;
}

.trend-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.trend-card-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.trend-metric {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.metric-value {
  font-size: 32px;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.2;
}

.trend-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.trend-period {
  color: #666;
}

.trend-change {
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
}

.trend-change.positive {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.1);
}

.trend-change.negative {
  color: #ff4d4f;
  background: rgba(255, 77, 79, 0.1);
}

.mini-chart-container {
  position: relative;
  width: 100%;
}

.mini-chart {
  width: 100%;
  height: 60px;
}

.chart-fill {
  opacity: 0.6;
}

.chart-line {
  stroke-linecap: round;
  stroke-linejoin: round;
}

.chart-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #999;
  margin-top: 4px;
}

/* Recent Section */
.recent-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.recent-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.card-header-title {
  display: flex;
  align-items: center;
}

/* Profile Content */
.profile-content {
  padding: 24px;
  background: white;
  border-radius: 0 8px 8px 8px;
}

.detail-sections {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.detail-section {
  width: 100%;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

@media (max-width: 1024px) {
  .trend-cards-row {
    grid-template-columns: 1fr;
  }

  .recent-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .ant-btn {
    width: 100%;
  }

  .performance-content,
  .profile-content {
    padding: 16px;
  }
}
</style>
