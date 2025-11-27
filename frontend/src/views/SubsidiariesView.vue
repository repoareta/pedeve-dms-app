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
            <a-button type="primary" size="large" @click="handleCreateCompany" class="add-button">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Add new Subsidiary
            </a-button>
          </div>
        </div>
      </div>

      <div class="mainContentPage">
        <!-- Subsidiary Cards Grid -->
        <div class="subsidiary-cards-grid" v-if="!companiesLoading && filteredCompanies.length > 0">
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

        <!-- Loading State -->
        <div v-if="companiesLoading" class="loading-container">
          <a-spin size="large" />
        </div>

        <!-- Empty State -->
        <div v-if="!companiesLoading && companies.length === 0" class="empty-state">
          <IconifyIcon icon="mdi:office-building-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
          <p>Belum ada data subsidiary</p>
          <a-button type="primary" @click="handleCreateCompany">
            <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
            Tambah Subsidiary Pertama
          </a-button>
        </div>

        <!-- No Search Results -->
        <div v-if="!companiesLoading && companies.length > 0 && filteredCompanies.length === 0" class="empty-state">
          <IconifyIcon icon="mdi:magnify" width="64" style="color: #ccc; margin-bottom: 16px;" />
          <p>Tidak ada hasil untuk "{{ searchText }}"</p>
          <a-button type="default" @click="searchText = ''">Hapus Filter</a-button>
        </div>

        <!-- Pagination -->
        <div v-if="filteredCompanies.length > 0" class="pagination-container">
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
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, type Company } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'

const router = useRouter()
const authStore = useAuthStore()

// Companies
const companies = ref<Company[]>([])
const companiesLoading = ref(false)
const searchText = ref('')

// Pagination
const currentPage = ref(1)
const pageSize = ref(8)

// Sample financial data (RKAP & Opex) - akan diganti dengan data real jika ada
const financialData = ref<Record<string, {
  rkap: { value: number; year: string; change: number }
  opex: { value: number; quarter: string; change: number }
}>>({})

// Computed untuk filtered companies berdasarkan search
const filteredCompanies = computed(() => {
  if (!searchText.value.trim()) {
    return companies.value
  }

  const search = searchText.value.toLowerCase().trim()
  return companies.value.filter(company =>
    company.name.toLowerCase().includes(search) ||
    company.code.toLowerCase().includes(search) ||
    (company.short_name && company.short_name.toLowerCase().includes(search)) ||
    (company.nib && company.nib.toLowerCase().includes(search)) ||
    (company.description && company.description.toLowerCase().includes(search))
  )
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
    message.error('Gagal memuat perusahaan: ' + (error.response?.data?.message || error.message))
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

onMounted(() => {
  loadCompanies()
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
</style>
