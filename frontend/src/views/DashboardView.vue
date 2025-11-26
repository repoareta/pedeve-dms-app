<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import KPICard from '../components/KPICard.vue'
import RevenueChart from '../components/RevenueChart.vue'
import SubsidiariesList from '../components/SubsidiariesList.vue'
import AdminDashboard from '../components/AdminDashboard.vue'
import ManagerDashboard from '../components/ManagerDashboard.vue'
import StaffDashboard from '../components/StaffDashboard.vue'

const router = useRouter()
const authStore = useAuthStore()

const selectedPeriod = ref('juni-2025')

const currentDate = computed(() => {
  const date = new Date()
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  return `${months[date.getMonth()]} ${date.getFullYear()}`
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

const handleExportPDF = () => {
  // TODO: Implementasi export PDF
  console.log('Export PDF')
}

const handleExportExcel = () => {
  // TODO: Implementasi export Excel
  console.log('Export Excel')
}
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
              <a-select v-model:value="selectedPeriod" placeholder="Select Periode" class="period-selector"
                size="large">
                <a-select-option value="juni-2025">Juni 2025</a-select-option>
                <a-select-option value="mei-2025">Mei 2025</a-select-option>
                <a-select-option value="april-2025">April 2025</a-select-option>
              </a-select>
              <a-button type="text" class="export-btn export-excel" @click="handleExportExcel">
                <IconifyIcon icon="mdi:file-excel" width="20" height="20" />
              </a-button>
              <a-button type="text" class="export-btn export-pdf" @click="handleExportPDF">
                <IconifyIcon icon="mdi:file-pdf-box" width="20" height="20" />
              </a-button>
            </div>
          </div>

          <!-- KPI Cards -->
          <div class="kpi-row">
            <KPICard title="Revenue" value="$120M" change="+10%" change-type="increase" icon="mdi:currency-usd" />
            <KPICard title="Opex" value="$80M" change="-5%" change-type="decrease" icon="mdi:chart-line" />
            <KPICard title="NPAT" value="$25M" change="+15%" change-type="increase" icon="mdi:chart-bar" />
            <KPICard title="Financial Ratios" value="1.5x" change="+5%" change-type="increase" icon="mdi:chart-pie" />
            <KPICard title="Dividend" value="$10M" change="+20%" change-type="increase" icon="mdi:cash-multiple" />
          </div>
        </div>
      </div>







      <!-- Charts and Lists Row -->
      <div class="mainContent">
        <a-row :gutter="[16, 16]" class="content-row">
          <a-col :xs="24" :lg="16" :xl="16">
            <RevenueChart />
          </a-col>
          <a-col :xs="24" :lg="8" :xl="8">
            <SubsidiariesList />
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
