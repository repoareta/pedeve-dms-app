<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import KPICard from '../components/KPICard.vue'
import RevenueChart from '../components/RevenueChart.vue'
import SubsidiariesList from '../components/SubsidiariesList.vue'
import AdminDashboard from '../components/AdminDashboard.vue'
import ManagerDashboard from '../components/ManagerDashboard.vue'
import StaffDashboard from '../components/StaffDashboard.vue'
import reportsApi, { type Report } from '../api/reports'
import { companyApi, type Company } from '../api/userManagement'
import dayjs from 'dayjs'

const router = useRouter()
const authStore = useAuthStore()

// Period filter - format YYYY-MM
const selectedPeriod = ref<string | null>(null)
const loading = ref(false)

// Initialize selectedPeriod to current month
const initializePeriod = () => {
  const now = dayjs()
  selectedPeriod.value = now.format('YYYY-MM')
}

// Available periods (last 12 months)
const availablePeriods = computed(() => {
  const periods: string[] = []
  const now = dayjs()
  for (let i = 11; i >= 0; i--) {
    const period = now.subtract(i, 'month')
    periods.push(period.format('YYYY-MM'))
  }
  return periods
})

// Format period for display
const formatPeriodDisplay = (period: string | null): string => {
  if (!period) return 'Pilih Periode'
  const parts = period.split('-')
  if (parts.length < 2) return period
  const [year, month] = parts
  if (!year || !month) return period
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  const monthIndex = parseInt(month, 10) - 1
  if (monthIndex >= 0 && monthIndex < months.length) {
    return `${months[monthIndex]} ${year}`
  }
  return period
}

const currentDate = computed(() => {
  if (selectedPeriod.value) {
    return formatPeriodDisplay(selectedPeriod.value)
  }
  const date = new Date()
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  return `${months[date.getMonth()]} ${date.getFullYear()}`
})

// Reports data
const allReports = ref<Report[]>([])
const previousPeriodReports = ref<Report[]>([])
const companies = ref<Company[]>([])

// Load reports based on period filter
const loadReports = async () => {
  loading.value = true
  try {
    // Load current period reports
    const params: any = {
      page: 1,
      page_size: 9999, // Load all reports
    }
    
    if (selectedPeriod.value) {
      params.period = selectedPeriod.value
    }

    const response = await reportsApi.getAll(params)
    allReports.value = response.data

    // Load previous period reports for comparison
    if (selectedPeriod.value) {
      const currentPeriod = dayjs(selectedPeriod.value)
      const prevPeriod = currentPeriod.subtract(1, 'month').format('YYYY-MM')
      
      try {
        const prevResponse = await reportsApi.getAll({
          page: 1,
          page_size: 9999,
          period: prevPeriod,
        })
        previousPeriodReports.value = prevResponse.data
      } catch (error) {
        // Silently fail if previous period doesn't exist
        previousPeriodReports.value = []
      }
    } else {
      previousPeriodReports.value = []
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    console.error('Failed to load reports:', axiosError.response?.data?.message || axiosError.message)
    allReports.value = []
    previousPeriodReports.value = []
  } finally {
    loading.value = false
  }
}

// Load companies
const loadCompanies = async () => {
  try {
    companies.value = await companyApi.getAll()
  } catch (error: unknown) {
    console.error('Failed to load companies:', error)
    companies.value = []
  }
}

// KPI Metrics computed from real data
const kpiMetrics = computed(() => {
  const reports = allReports.value
  
  if (reports.length === 0) {
    return {
      revenue: { value: 0, change: 0 },
      opex: { value: 0, change: 0 },
      npat: { value: 0, change: 0 },
      financialRatio: { value: 0, change: 0 },
      dividend: { value: 0, change: 0 },
    }
  }

  // Calculate totals for current period
  const currentRevenue = reports.reduce((sum, r) => sum + (r.revenue || 0), 0)
  const currentOpex = reports.reduce((sum, r) => sum + (r.opex || 0), 0)
  const currentNpat = reports.reduce((sum, r) => sum + (r.npat || 0), 0)
  const currentDividend = reports.reduce((sum, r) => sum + (r.dividend || 0), 0)
  const currentFinancialRatio = reports.length > 0 
    ? reports.reduce((sum, r) => sum + (r.financial_ratio || 0), 0) / reports.length 
    : 0

  // Calculate previous period for comparison
  const prevReports = previousPeriodReports.value
  const previousRevenue = prevReports.reduce((sum, r) => sum + (r.revenue || 0), 0)
  const previousOpex = prevReports.reduce((sum, r) => sum + (r.opex || 0), 0)
  const previousNpat = prevReports.reduce((sum, r) => sum + (r.npat || 0), 0)
  const previousDividend = prevReports.reduce((sum, r) => sum + (r.dividend || 0), 0)
  const previousFinancialRatio = prevReports.length > 0
    ? prevReports.reduce((sum, r) => sum + (r.financial_ratio || 0), 0) / prevReports.length
    : 0

  // Calculate percentage changes
  const calculateChange = (current: number, previous: number): number => {
    if (previous === 0) return 0
    return ((current - previous) / previous) * 100
  }

  return {
    revenue: {
      value: currentRevenue,
      change: calculateChange(currentRevenue, previousRevenue),
    },
    opex: {
      value: currentOpex,
      change: calculateChange(currentOpex, previousOpex),
    },
    npat: {
      value: currentNpat,
      change: calculateChange(currentNpat, previousNpat),
    },
    financialRatio: {
      value: currentFinancialRatio,
      change: calculateChange(currentFinancialRatio, previousFinancialRatio),
    },
    dividend: {
      value: currentDividend,
      change: calculateChange(currentDividend, previousDividend),
    },
  }
})

// Format currency
const formatCurrency = (value: number): string => {
  if (value >= 1000000000) {
    return `$${(value / 1000000000).toFixed(0)}B`
  } else if (value >= 1000000) {
    return `$${(value / 1000000).toFixed(0)}M`
  } else if (value >= 1000) {
    return `$${(value / 1000).toFixed(0)}K`
  }
  return `$${value.toFixed(0)}`
}

// Format change percentage
const formatChange = (change: number): string => {
  const sign = change >= 0 ? '+' : ''
  return `${sign}${change.toFixed(0)}%`
}

// Chart data for RevenueChart - show last 12 months
const chartData = computed(() => {
  // Combine all reports (current + previous periods) for chart
  const allReportsForChart = [...allReports.value, ...previousPeriodReports.value]
  
  // Group reports by period and calculate totals
  const periodMap = new Map<string, { revenue: number; npat: number; count: number }>()
  
  allReportsForChart.forEach(report => {
    if (!report.period) return
    
    const existing = periodMap.get(report.period) || { revenue: 0, npat: 0, count: 0 }
    existing.revenue += report.revenue || 0
    existing.npat += report.npat || 0
    existing.count += 1
    periodMap.set(report.period, existing)
  })

  // Get last 12 months
  const now = selectedPeriod.value ? dayjs(selectedPeriod.value) : dayjs()
  const periods: string[] = []
  for (let i = 11; i >= 0; i--) {
    const period = now.subtract(i, 'month').format('YYYY-MM')
    periods.push(period)
  }
  
  // Generate labels and data
  const labels = periods.map(period => {
    const parts = period.split('-')
    if (parts.length < 2) return period
    const [year, month] = parts
    if (!year || !month) return period
    const months = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ]
    const monthIndex = parseInt(month, 10) - 1
    return monthIndex >= 0 && monthIndex < months.length 
      ? `${months[monthIndex]} ${year}`
      : period
  })

  const revenueData = periods.map(period => {
    const data = periodMap.get(period)
    return data ? data.revenue / 1000000 : 0 // Convert to millions
  })

  const npatData = periods.map(period => {
    const data = periodMap.get(period)
    return data ? data.npat / 1000000 : 0 // Convert to millions
  })

  // Calculate RKAP (110% of average revenue)
  const avgRevenue = revenueData.length > 0 
    ? revenueData.reduce((sum, val) => sum + val, 0) / revenueData.length 
    : 0
  const rkapData = revenueData.map(() => avgRevenue * 1.1)

  return {
    labels,
    revenueData,
    npatData,
    rkapData,
  }
})

// Underperforming subsidiaries - Show top 5 worst performers based on financial ratio
const underperformingSubsidiaries = computed(() => {
  // Calculate performance metrics for each company
  const companyMetrics: Array<{
    company: Company
    reports: Report[]
    avgFinancialRatio: number
    totalDividend: number
    totalRevenue: number
    rkapPercent: number
    performanceScore: number // Combined score for sorting (lower is worse)
  }> = []

  companies.value.forEach(company => {
    const companyReports = allReports.value.filter(r => r.company_id === company.id)
    if (companyReports.length === 0) return

    const avgFinancialRatio = companyReports.reduce((sum, r) => sum + (r.financial_ratio || 0), 0) / companyReports.length
    const totalDividend = companyReports.reduce((sum, r) => sum + (r.dividend || 0), 0)
    const totalRevenue = companyReports.reduce((sum, r) => sum + (r.revenue || 0), 0)
    const avgRevenue = totalRevenue / companyReports.length
    const rkapTarget = avgRevenue * 1.1
    const rkapPercent = rkapTarget > 0 ? (totalRevenue / rkapTarget) * 100 : 0

    // Calculate performance score (lower is worse)
    // Combine financial ratio (weight 70%) and RKAP percent (weight 30%)
    // Normalize RKAP percent (100% = good, so invert it: 100 - rkapPercent)
    const normalizedRKAP = Math.max(0, 100 - rkapPercent) / 100 // 0-1 scale, higher is worse
    const normalizedFinancialRatio = Math.max(0, 2.0 - avgFinancialRatio) / 2.0 // 0-1 scale, higher is worse
    const performanceScore = (normalizedFinancialRatio * 0.7) + (normalizedRKAP * 0.3)

    companyMetrics.push({
      company,
      reports: companyReports,
      avgFinancialRatio,
      totalDividend,
      totalRevenue,
      rkapPercent,
      performanceScore,
    })
  })

  // Sort by worst performance first (highest performanceScore = worst)
  const sorted = companyMetrics
    .sort((a, b) => b.performanceScore - a.performanceScore) // Sort by worst first
    .slice(0, 5) // Top 5 worst

  return sorted.map(metric => {
    // Calculate dividend percentage
    const dividendPercent = metric.totalRevenue > 0 
      ? (metric.totalDividend / metric.totalRevenue) * 100 
      : 0

    // Determine financial score
    const score = metric.avgFinancialRatio >= 2.0 ? 'A' : 
                  metric.avgFinancialRatio >= 1.5 ? 'B' : 
                  metric.avgFinancialRatio >= 1.0 ? 'C' : 
                  metric.avgFinancialRatio >= 0.5 ? 'D+' : 'D'

    return {
      id: metric.company.id,
      name: metric.company.name,
      rkap: `${Math.round(metric.rkapPercent)}%`,
      dividen: `${Math.round(dividendPercent)}%`,
      score: score,
      financialRatio: metric.avgFinancialRatio,
      company: metric.company, // Include full company object for logo
    }
  })
})

// Determine which dashboard to show based on user role
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperadmin = computed(() => userRole.value === 'superadmin')
const isAdmin = computed(() => userRole.value === 'admin')
const isManager = computed(() => userRole.value === 'manager')
const isStaff = computed(() => userRole.value === 'staff')

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

const handleExportPDF = async () => {
  try {
    const params: any = {}
    if (selectedPeriod.value) {
      params.period = selectedPeriod.value
    }
    
    const blob = await reportsApi.exportPDF(params)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    
    let filename = 'dashboard_report'
    if (selectedPeriod.value) {
      filename += `_${selectedPeriod.value}`
    }
    filename += `_${dayjs().format('YYYYMMDD')}.pdf`
    
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    message.success('Export PDF berhasil')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal export PDF: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  }
}

const handleExportExcel = async () => {
  try {
    const params: any = {}
    if (selectedPeriod.value) {
      params.period = selectedPeriod.value
    }
    
    const blob = await reportsApi.exportExcel(params)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    
    let filename = 'dashboard_report'
    if (selectedPeriod.value) {
      filename += `_${selectedPeriod.value}`
    }
    filename += `_${dayjs().format('YYYYMMDD')}.xlsx`
    
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    message.success('Export Excel berhasil')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal export Excel: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  }
}

// Watch period changes
watch(selectedPeriod, () => {
  loadReports()
})

// Load data on mount
onMounted(async () => {
  initializePeriod()
  await Promise.all([
    loadCompanies(),
    loadReports(),
  ])
})
</script>

<template>
  <div class="dashboard-layout">
    <DashboardHeader @logout="handleLogout" />

    <!-- Superadmin Dashboard (Holding Company) -->
    <div v-if="isSuperadmin" class="dashboard-content">

      <div class="bg-header-general">

        <div class="dashboard-header-container">
          <div class="dashboard-header-row">
            <div class="dashboard-title-section">
              <h1 class="dashboard-title">Dashboard</h1>
              <span class="dashboard-date-label">{{ currentDate }}</span>
            </div>
            <div class="dashboard-actions">
              <a-select 
                v-model:value="selectedPeriod" 
                placeholder="Select Periode" 
                class="period-selector"
                size="large"
                :loading="loading"
              >
                <a-select-option 
                  v-for="period in availablePeriods" 
                  :key="period" 
                  :value="period"
                >
                  {{ formatPeriodDisplay(period) }}
                </a-select-option>
              </a-select>
              <a-button 
                type="text" 
                class="export-btn export-excel" 
                @click="handleExportExcel"
                :loading="loading"
              >
                <IconifyIcon icon="mdi:file-excel" width="20" height="20" />
              </a-button>
              <a-button 
                type="text" 
                class="export-btn export-pdf" 
                @click="handleExportPDF"
                :loading="loading"
              >
                <IconifyIcon icon="mdi:file-pdf-box" width="20" height="20" />
              </a-button>
            </div>
          </div>

          <!-- KPI Cards -->
          <div class="kpi-row">
            <KPICard 
              title="Revenue" 
              :value="formatCurrency(kpiMetrics.revenue.value)" 
              :change="formatChange(kpiMetrics.revenue.change)" 
              :change-type="kpiMetrics.revenue.change >= 0 ? 'increase' : 'decrease'" 
              icon="mdi:currency-usd" 
            />
            <KPICard 
              title="Opex" 
              :value="formatCurrency(kpiMetrics.opex.value)" 
              :change="formatChange(kpiMetrics.opex.change)" 
              :change-type="kpiMetrics.opex.change >= 0 ? 'decrease' : 'increase'" 
              icon="mdi:chart-line" 
            />
            <KPICard 
              title="NPAT" 
              :value="formatCurrency(kpiMetrics.npat.value)" 
              :change="formatChange(kpiMetrics.npat.change)" 
              :change-type="kpiMetrics.npat.change >= 0 ? 'increase' : 'decrease'" 
              icon="mdi:chart-bar" 
            />
            <KPICard 
              title="Financial Ratios" 
              :value="`${kpiMetrics.financialRatio.value.toFixed(1)}x`" 
              :change="formatChange(kpiMetrics.financialRatio.change)" 
              :change-type="kpiMetrics.financialRatio.change >= 0 ? 'increase' : 'decrease'" 
              icon="mdi:chart-pie" 
            />
            <KPICard 
              title="Dividend" 
              :value="formatCurrency(kpiMetrics.dividend.value)" 
              :change="formatChange(kpiMetrics.dividend.change)" 
              :change-type="kpiMetrics.dividend.change >= 0 ? 'increase' : 'decrease'" 
              icon="mdi:cash-multiple" 
            />
          </div>
        </div>
      </div>







      <!-- Charts and Lists Row -->
      <div class="mainContent">
        <a-row :gutter="[16, 16]" class="content-row">
          <a-col :xs="24" :lg="16" :xl="16">
            <RevenueChart 
              :chart-data="chartData"
              :loading="loading"
            />
          </a-col>
          <a-col :xs="24" :lg="8" :xl="8">
            <SubsidiariesList 
              :subsidiaries="underperformingSubsidiaries"
              :loading="loading"
            />
          </a-col>
        </a-row>
      </div>



    </div>

    <!-- Admin Dashboard -->
    <div v-else-if="isAdmin" class="dashboard-content">
      <AdminDashboard />
    </div>

    <!-- Manager Dashboard -->
    <div v-else-if="isManager" class="dashboard-content">
      <ManagerDashboard />
    </div>

    <!-- Staff Dashboard -->
    <div v-else-if="isStaff" class="dashboard-content">
      <StaffDashboard />
    </div>

    <!-- Fallback for unknown roles -->
    <div v-else class="dashboard-content">
      <a-card>
        <a-result status="warning" title="Role tidak dikenali"
          sub-title="Dashboard untuk role Anda belum tersedia. Silakan hubungi administrator.">
          <template #extra>
            <a-button type="primary" @click="handleLogout">
              Logout
            </a-button>
          </template>
        </a-result>
      </a-card>
    </div>
  </div>
</template>
