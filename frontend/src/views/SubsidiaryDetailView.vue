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
        <div class="back-button-container">
          <a-button type="text" @click="handleBack" class="back-button">
            <IconifyIcon icon="mdi:arrow-left" width="20" style="margin-right: 8px;" />
            Kembali ke Daftar Subsidiary
          </a-button>
        </div>

        <div class="page-header-container" style="min-height: 350px; width: 100%;">
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

                <!-- Company Hierarchy -->
                <div v-if="companyHierarchy.length > 1" class="company-hierarchy">
                  <span class="hierarchy-label">Hirarki:</span>
                  <span class="hierarchy-path">
                    <template v-for="(hierarchyCompany, index) in companyHierarchy" :key="hierarchyCompany.id">
                      <span class="hierarchy-item">{{ hierarchyCompany.name }}</span>
                      <span v-if="index < companyHierarchy.length - 1" class="hierarchy-separator">/</span>
                    </template>
                  </span>
                </div>
                
              </div>
            </div>
            <div class="header-actions">
              <a-space>
                <a-button @click="handleExportPDF" :loading="exportLoading"  class="btn-icon-label">
                  <IconifyIcon icon="mdi:file-pdf-box" width="16" style="margin-right: 8px;" />
                  PDF
                </a-button>
                <a-button @click="handleExportExcel" :loading="exportLoading"  class="btn-icon-label">
                  <IconifyIcon icon="mdi:file-excel-box" width="16" style="margin-right: 8px;" />
                  Excel
                </a-button>
                <a-date-picker v-model:value="selectedPeriod" picker="month" placeholder="Select Periode"
                  format="YYYY-MM" style="width: 150px;" @change="handlePeriodChange" />
                <a-dropdown v-if="hasAnyMenuOption">
                  <template #overlay>
                    <a-menu @click="handleMenuClick">
                      <a-menu-item v-if="canEdit" key="edit">
                        <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                        Edit
                      </a-menu-item>
                      <a-menu-item v-if="canAssignRole" key="assign-role">
                        <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                        Assign Role
                      </a-menu-item>
                      <a-menu-divider v-if="canDelete && (canEdit || canAssignRole)" />
                      <a-menu-item v-if="canDelete" key="delete" danger>
                        <IconifyIcon icon="mdi:delete" width="16" style="margin-right: 8px;" />
                        Hapus
                      </a-menu-item>
                    </a-menu>
                  </template>
                  <a-button  style="display: flex; align-items: center;" class="btn-icon-label">
                    <IconifyIcon icon="mdi:dots-vertical" width="16" style="margin-right: 8px;" />
                    Options
                  </a-button>
                </a-dropdown>
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
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart"
                          preserveAspectRatio="none">
                          <defs>
                            <linearGradient id="rkapGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#ff9800;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#ff9800;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path v-if="rkapChartFillPath" :d="rkapChartFillPath" fill="url(#rkapGradient)"
                            class="chart-fill" />
                          <path v-if="rkapTargetChartPath" :d="rkapTargetChartPath" stroke="#1890ff" stroke-width="1.5"
                            stroke-dasharray="4,2" fill="none" class="chart-line" />
                          <path v-if="rkapChartPath" :d="rkapChartPath" stroke="#ff9800" stroke-width="2" fill="none"
                            class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span v-if="rkapChartData.periods.length > 0">{{ rkapChartData.periods[0] }}</span>
                          <span v-else>Jan</span>
                          <span v-if="rkapChartData.periods.length > 0">{{
                            rkapChartData.periods[rkapChartData.periods.length -
                            1] }}</span>
                          <span v-else>Des</span>
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
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart"
                          preserveAspectRatio="none">
                          <defs>
                            <linearGradient id="opexGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#666;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#666;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path v-if="opexChartFillPath" :d="opexChartFillPath" fill="url(#opexGradient)"
                            class="chart-fill" />
                          <path v-if="opexChartPath" :d="opexChartPath" stroke="#666" stroke-width="2" fill="none"
                            class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span v-if="opexChartData.periods.length > 0">{{ opexChartData.periods[0] }}</span>
                          <span v-else>Jan</span>
                          <span v-if="opexChartData.periods.length > 0">{{
                            opexChartData.periods[opexChartData.periods.length -
                            1] }}</span>
                          <span v-else>Des</span>
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
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart"
                          preserveAspectRatio="none">
                          <defs>
                            <linearGradient id="npatGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#52c41a;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#52c41a;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path v-if="npatChartFillPath" :d="npatChartFillPath" fill="url(#npatGradient)"
                            class="chart-fill" />
                          <path v-if="npatChartPath" :d="npatChartPath" stroke="#52c41a" stroke-width="2" fill="none"
                            class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span v-if="npatChartData.periods.length > 0">{{ npatChartData.periods[0] }}</span>
                          <span v-else>Jan</span>
                          <span v-if="npatChartData.periods.length > 0">{{
                            npatChartData.periods[npatChartData.periods.length -
                            1] }}</span>
                          <span v-else>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>
                </div>

                <!-- Recent Reports (Recent Files disembunyikan) -->
                <div class="recent-section">
                  <!-- Recent Reports -->
                  <a-card class="recent-card full-width" :bordered="false">
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
                    <a-table :columns="reportColumns" :data-source="recentReports" :pagination="false"
                      :show-header="true" size="small">
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

      <!-- Assign Role Modal -->
      <a-modal v-model:open="assignRoleModalVisible" title="Assign Role - Kelola Pengurus"
        :confirm-loading="assignRoleLoading" width="900px" :footer="null">
        <div class="assign-role-container">
          <!-- Form Assign Role Baru -->
          <div class="assign-new-section">
            <h3 class="section-header">
              <IconifyIcon icon="mdi:account-plus" width="20" style="margin-right: 8px;" />
              Assign Role Baru
            </h3>
            <a-form :model="assignRoleForm" layout="vertical">
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-form-item label="Cari User" required>
                    <a-select v-model:value="assignRoleForm.userId" show-search
                      placeholder="Cari user berdasarkan nama atau email" :filter-option="filterUserOption"
                      :loading="usersLoading" @search="handleUserSearch" allow-clear :disabled="usersLoading">
                      <a-select-option v-for="user in filteredUsers" :key="user.id" :value="user.id"
                        :disabled="companyUsers.some(u => u.id === user.id)">
                        {{ user.username }} ({{ user.email }})
                        <span v-if="companyUsers.some(u => u.id === user.id)" class="text-muted"> - Sudah menjadi
                          pengurus</span>
                      </a-select-option>
                    </a-select>
                    <small v-if="usersLoading" class="text-muted">Memuat daftar user...</small>
                    <small v-else-if="allUsers.length === 0 && !usersLoading" class="text-muted">
                      Tidak ada user yang tersedia untuk di-assign.
                    </small>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Pilih Role" required>
                    <a-select v-model:value="assignRoleForm.roleId" show-search placeholder="Cari role"
                      :filter-option="filterRoleOption" :loading="rolesLoading" @search="handleRoleSearch" allow-clear
                      :disabled="rolesLoading">
                      <a-select-option v-for="role in filteredRoles" :key="role.id" :value="role.id">
                        {{ role.name }}
                      </a-select-option>
                    </a-select>
                    <small v-if="rolesLoading" class="text-muted">Memuat daftar role...</small>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-form-item>
                <a-button type="primary" :loading="assignRoleLoading" @click="handleAssignRole"
                  :disabled="!assignRoleForm.userId || !assignRoleForm.roleId">
                  <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                  Assign Role
                </a-button>
              </a-form-item>
            </a-form>
          </div>

          <a-divider>Pengurus Saat Ini</a-divider>

          <!-- List Pengurus Saat Ini -->
          <div class="current-users-section">
            <h3 class="section-header">
              <IconifyIcon icon="mdi:account-group" width="20" style="margin-right: 8px;" />
              Pengurus Saat Ini
            </h3>
            <a-table :columns="userColumns" :data-source="companyUsers" :loading="usersLoading"
              :pagination="{ pageSize: 10 }" row-key="id" size="middle">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'role'">
                  <a-tag v-if="record.role" :color="getRoleColor(record.role)">
                    {{ record.role }}
                  </a-tag>
                  <span v-else class="text-muted">-</span>
                </template>
                <template v-if="column.key === 'status'">
                  <a-tag :color="record.is_active ? 'green' : 'red'">
                    {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
                  </a-tag>
                </template>
                <template v-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleEditUserRole(record)">
                      <IconifyIcon icon="mdi:pencil" width="16" />
                      Ubah Role
                    </a-button>
                    <a-button type="link" size="small" danger @click="handleRemoveUser(record)">
                      <IconifyIcon icon="mdi:delete" width="16" />
                      Hapus
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
          </div>
        </div>
      </a-modal>

      <!-- Edit User Role Modal -->
      <a-modal v-model:open="editingUserRoleModalVisible" title="Ubah Role Pengurus"
        :confirm-loading="editingRoleLoading" @ok="handleSaveUserRole" @cancel="handleCancelEditUserRole" width="500px">
        <a-form layout="vertical" v-if="editingUserRole">
          <a-form-item label="User">
            <a-input :value="getUserById(editingUserRole.userId)?.username" disabled />
          </a-form-item>
          <a-form-item label="Pilih Role Baru" required>
            <a-select v-model:value="editingUserRole.roleId" show-search placeholder="Cari role"
              :filter-option="filterRoleOption" :loading="rolesLoading" @search="handleRoleSearch">
              <a-select-option v-for="role in filteredRoles" :key="role.id" :value="role.id">
                {{ role.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-form>
      </a-modal>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, userApi, roleApi, type Company, type BusinessField, type User, type Role } from '../api/userManagement'
import reportsApi, { type Report } from '../api/reports'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import dayjs, { type Dayjs } from 'dayjs'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const company = ref<Company | null>(null)
const loading = ref(false)
const activeTab = ref('performance')
const selectedPeriod = ref<Dayjs | null>(null)
const exportLoading = ref(false)
const companyHierarchy = ref<Company[]>([])
const allCompanies = ref<Company[]>([])

// Assign Role Modal
const assignRoleModalVisible = ref(false)
const assignRoleLoading = ref(false)
const assignRoleForm = ref({
  userId: undefined as string | undefined,
  roleId: undefined as string | undefined,
})

// Company Users (Pengurus)
const companyUsers = ref<User[]>([])
const allUsers = ref<User[]>([])
const allRoles = ref<Role[]>([])
const usersLoading = ref(false)
const rolesLoading = ref(false)
const userSearchText = ref('')
const roleSearchText = ref('')

// Editing user role
const editingUserRole = ref<{ userId: string; roleId: string | undefined } | null>(null)
const editingUserRoleModalVisible = ref(false)
const editingRoleLoading = ref(false)

// User columns for table
const userColumns = [
  { title: 'Username', dataIndex: 'username', key: 'username' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Role', key: 'role' },
  { title: 'Status', key: 'status' },
  { title: 'Aksi', key: 'action', width: 200 },
]

// Computed: Check user roles
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperAdmin = computed(() => userRole.value === 'superadmin')
const isAdmin = computed(() => userRole.value === 'admin')
const isManager = computed(() => userRole.value === 'manager')
const isStaff = computed(() => userRole.value === 'staff')

// RBAC: Assign Role hanya untuk admin
const canAssignRole = computed(() => isAdmin.value || isSuperAdmin.value)

// RBAC: Delete hanya untuk admin
const canDelete = computed(() => isAdmin.value || isSuperAdmin.value)

// RBAC: Edit untuk semua role (staff, manager, admin, superadmin)
const canEdit = computed(() => isAdmin.value || isManager.value || isStaff.value || isSuperAdmin.value)

// Check if any menu item is available (to show/hide Options dropdown)
const hasAnyMenuOption = computed(() => canEdit.value || canAssignRole.value || canDelete.value)

// Filtered users and roles
const filteredUsers = computed(() => {
  // Filter out users that are already pengurus
  const availableUsers = allUsers.value.filter(
    user => !companyUsers.value.some(cu => cu.id === user.id)
  )

  if (!userSearchText.value) {
    return availableUsers.slice(0, 20) // Limit to 20 for performance
  }
  const search = userSearchText.value.toLowerCase()
  return availableUsers.filter(
    user =>
      user.username.toLowerCase().includes(search) ||
      user.email.toLowerCase().includes(search)
  ).slice(0, 20)
})

const filteredRoles = computed(() => {
  // Filter out superadmin role - hanya untuk developer, bukan untuk user pengguna
  const nonSuperadminRoles = allRoles.value.filter(role => role.name.toLowerCase() !== 'superadmin')

  if (!roleSearchText.value) {
    return nonSuperadminRoles
  }
  const search = roleSearchText.value.toLowerCase()
  return nonSuperadminRoles.filter(
    role => role.name.toLowerCase().includes(search)
  )
})

// Chart data computed from filtered reports
const rkapData = computed(() => {
  if (filteredReports.value.length === 0) {
    return { value: 0, year: '2025', change: 0 }
  }

  // Get latest report
  const latest = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return b.period.localeCompare(a.period)
  })[0]

  if (!latest) {
    return { value: 0, year: '2025', change: 0 }
  }

  const year = latest.period ? latest.period.split('-')[0] : '2025'
  const revenue = latest.revenue || 0

  // Calculate change from previous period
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const currentIndex = sorted.findIndex(r => r.id === latest.id)
  let change = 0
  if (currentIndex > 0) {
    const previous = sorted[currentIndex - 1]
    if (!previous) return { value: revenue, year: year, change: 0 }
    const prevRevenue = previous.revenue || 0
    if (prevRevenue > 0) {
      change = ((revenue - prevRevenue) / prevRevenue) * 100
    }
  }

  return {
    value: revenue,
    year: year,
    change: Math.round(change * 10) / 10
  }
})

const opexData = computed(() => {
  if (filteredReports.value.length === 0) {
    return { value: 0, quarter: 'Q1 2025', change: 0 }
  }

  // Get latest report
  const latest = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return b.period.localeCompare(a.period)
  })[0]

  if (!latest) {
    return { value: 0, quarter: 'Q1 2025', change: 0 }
  }

  const period = latest.period || ''
  const quarter = formatPeriod(period)
  const opex = latest.opex || 0

  // Calculate change from previous period
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const currentIndex = sorted.findIndex(r => r.id === latest.id)
  let change = 0
  if (currentIndex > 0) {
    const previous = sorted[currentIndex - 1]
    if (!previous) return { value: opex, quarter: quarter, change: 0 }
    const prevOpex = previous.opex || 0
    if (prevOpex > 0) {
      change = ((opex - prevOpex) / prevOpex) * 100
    }
  }

  return {
    value: opex,
    quarter: quarter,
    change: Math.round(Math.abs(change) * 10) / 10
  }
})

const npatData = computed(() => {
  if (filteredReports.value.length === 0) {
    return { value: 0, quarter: 'Q1 2025', change: 0 }
  }

  // Get latest report
  const latest = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return b.period.localeCompare(a.period)
  })[0]

  if (!latest) {
    return { value: 0, quarter: 'Q1 2025', change: 0 }
  }

  const period = latest.period || ''
  const quarter = formatPeriod(period)
  const npat = latest.npat || 0

  // Calculate change from previous period
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const currentIndex = sorted.findIndex(r => r.id === latest.id)
  let change = 0
  if (currentIndex > 0) {
    const previous = sorted[currentIndex - 1]
    if (!previous) return { value: npat, quarter: quarter, change: 0 }
    const prevNpat = previous.npat || 0
    if (prevNpat > 0) {
      change = ((npat - prevNpat) / prevNpat) * 100
    }
  }

  return {
    value: npat,
    quarter: quarter,
    change: Math.round(change * 10) / 10
  }
})

// Generate chart data from real reports
const generateChartDataFromReports = (reports: Report[], field: 'revenue' | 'opex' | 'npat') => {
  if (reports.length === 0) return []

  // Sort by period
  const sorted = [...reports].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })

  // Extract data for the field
  return sorted.map(report => {
    if (field === 'revenue') return report.revenue || 0
    if (field === 'opex') return report.opex || 0
    if (field === 'npat') return report.npat || 0
    return 0
  })
}

// Format period helper
const formatPeriod = (period: string | undefined): string => {
  if (!period) return 'Unknown'
  const [year, month] = period.split('-')
  if (!year || !month) return period
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  const monthIndex = parseInt(month, 10) - 1
  if (monthIndex < 0 || monthIndex >= months.length) return period
  return `${months[monthIndex]} ${year}`
}

// Calculate RKAP percentage (dummy calculation for now - can be enhanced with actual RKAP data)
const calculateRKAPPercent = (report: Report): number => {
  // Dummy calculation - in real app, this would come from RKAP data
  // For now, calculate as percentage of revenue vs a target
  const target = report.revenue * 1.1 // Assume target is 110% of revenue
  if (target === 0) return 0
  return Math.round((report.revenue / target) * 100)
}

// RKAP vs Realization chart data
// For RKAP, we'll use revenue as realization and calculate RKAP as target (110% of average revenue)
const rkapChartData = computed(() => {
  const revenueData = generateChartDataFromReports(filteredReports.value, 'revenue')
  if (revenueData.length === 0) return { rkap: [], realization: [], periods: [] }

  // Get periods for labels
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const periods = sorted.map(r => {
    if (!r.period) return ''
    const parts = r.period.split('-')
    const month = parts[1]
    if (!month) return ''
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
    const monthIndex = parseInt(month, 10) - 1
    return monthIndex >= 0 && monthIndex < monthNames.length ? monthNames[monthIndex] : month
  })

  // Calculate average revenue
  const avgRevenue = revenueData.reduce((sum, val) => sum + val, 0) / revenueData.length
  // RKAP target is 110% of average
  const rkapTarget = avgRevenue * 1.1

  // Generate RKAP line (target line)
  const rkap = revenueData.map(() => rkapTarget)

  return {
    rkap: rkap,
    realization: revenueData,
    periods: periods
  }
})

const opexChartData = computed(() => {
  const data = generateChartDataFromReports(filteredReports.value, 'opex')
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const periods = sorted.map(r => {
    if (!r.period) return ''
    const parts = r.period.split('-')
    const month = parts[1]
    if (!month) return ''
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
    const monthIndex = parseInt(month, 10) - 1
    return monthIndex >= 0 && monthIndex < monthNames.length ? monthNames[monthIndex] : month
  })
  return { data, periods }
})

const npatChartData = computed(() => {
  const data = generateChartDataFromReports(filteredReports.value, 'npat')
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return a.period.localeCompare(b.period)
  })
  const periods = sorted.map(r => {
    if (!r.period) return ''
    const parts = r.period.split('-')
    const month = parts[1]
    if (!month) return ''
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
    const monthIndex = parseInt(month, 10) - 1
    return monthIndex >= 0 && monthIndex < monthNames.length ? monthNames[monthIndex] : month
  })
  return { data, periods }
})

// Generate smooth SVG path untuk chart menggunakan cubic bezier (smooth seperti gelombang radio)
const generateChartPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  if (data.length === 1) {
    const value = data[0] ?? 0
    const max = Math.max(...data)
    const min = Math.min(...data)
    const range = max - min || 1
    const y = height - ((value - min) / range) * height
    return `M 0 ${y} L ${width} ${y}`
  }

  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)

  // Convert data points to coordinates
  const points: Array<{ x: number; y: number }> = data.map((value, i) => ({
    x: i * stepX,
    y: height - ((value - min) / range) * height
  }))

  // Generate smooth curve using cubic bezier with better control points
  let path = `M ${points[0]!.x} ${points[0]!.y}`

  // Use smooth bezier curves - calculate control points based on tangent direction
  for (let i = 0; i < points.length - 1; i++) {
    const current = points[i]!
    const next = points[i + 1]!

    // Get previous and next points for tangent calculation
    const prev = i > 0 ? points[i - 1]! : current
    const after = i < points.length - 2 ? points[i + 2]! : next

    // Calculate tangent direction (slope)
    const dx1 = (next.x - prev.x) / 2
    const dy1 = (next.y - prev.y) / 2
    const dx2 = (after.x - current.x) / 2
    const dy2 = (after.y - current.y) / 2

    // Control points for smooth curve (using 1/3 of distance for smoothness)
    const cp1x = current.x + dx1 / 3
    const cp1y = current.y + dy1 / 3
    const cp2x = next.x - dx2 / 3
    const cp2y = next.y - dy2 / 3

    path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${next.x} ${next.y}`
  }

  return path
}

const generateChartFillPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  if (data.length === 1) {
    const value = data[0] ?? 0
    const max = Math.max(...data)
    const min = Math.min(...data)
    const range = max - min || 1
    const y = height - ((value - min) / range) * height
    return `M 0 ${height} L 0 ${y} L ${width} ${y} L ${width} ${height} Z`
  }

  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)

  // Convert data points to coordinates
  const points: Array<{ x: number; y: number }> = data.map((value, i) => ({
    x: i * stepX,
    y: height - ((value - min) / range) * height
  }))

  // Start from bottom left
  let path = `M 0 ${height}`

  // Line to first point
  path += ` L ${points[0]!.x} ${points[0]!.y}`

  // Generate smooth curve using cubic bezier (same algorithm as line path)
  for (let i = 0; i < points.length - 1; i++) {
    const current = points[i]!
    const next = points[i + 1]!

    // Get previous and next points for tangent calculation
    const prev = i > 0 ? points[i - 1]! : current
    const after = i < points.length - 2 ? points[i + 2]! : next

    // Calculate tangent direction (slope)
    const dx1 = (next.x - prev.x) / 2
    const dy1 = (next.y - prev.y) / 2
    const dx2 = (after.x - current.x) / 2
    const dy2 = (after.y - current.y) / 2

    // Control points for smooth curve (using 1/3 of distance for smoothness)
    const cp1x = current.x + dx1 / 3
    const cp1y = current.y + dy1 / 3
    const cp2x = next.x - dx2 / 3
    const cp2y = next.y - dy2 / 3

    path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${next.x} ${next.y}`
  }

  // Close path to bottom right and back to start
  path += ` L ${width} ${height} Z`
  return path
}

// RKAP chart paths - need two lines (RKAP target and Realization)
const rkapChartPath = computed(() => {
  const chartData = rkapChartData.value
  if (chartData.realization.length === 0) return ''
  // Use realization data for the main line
  return generateChartPath(chartData.realization)
})

const rkapChartFillPath = computed(() => {
  const chartData = rkapChartData.value
  if (chartData.realization.length === 0) return ''
  return generateChartFillPath(chartData.realization)
})

// RKAP target line path
const rkapTargetChartPath = computed(() => {
  const chartData = rkapChartData.value
  if (chartData.rkap.length === 0) return ''
  return generateChartPath(chartData.rkap)
})

const opexChartPath = computed(() => generateChartPath(opexChartData.value.data))
const opexChartFillPath = computed(() => generateChartFillPath(opexChartData.value.data))
const npatChartPath = computed(() => generateChartPath(npatChartData.value.data))
const npatChartFillPath = computed(() => generateChartFillPath(npatChartData.value.data))

// Reports data
const companyReports = ref<Report[]>([])
const reportsLoading = ref(false)

// Filtered reports based on selected period
const filteredReports = computed(() => {
  if (!selectedPeriod.value) {
    return companyReports.value
  }
  const periodStr = selectedPeriod.value.format('YYYY-MM')
  return companyReports.value.filter(report => report.period === periodStr)
})

// Recent reports computed from filtered data
const recentReports = computed(() => {
  // Sort by period descending and take latest 5
  const sorted = [...filteredReports.value].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return b.period.localeCompare(a.period)
  }).slice(0, 5)

  return sorted.map((report, index) => ({
    key: String(index + 1),
    name: `Laporan ${formatPeriod(report.period)}`,
    rkap_percent: calculateRKAPPercent(report),
    revenue: formatCurrency(report.revenue),
    npat: formatCurrency(report.npat),
    opex: formatCurrency(report.opex),
    period: report.period
  }))
})

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
    // Load reports after company is loaded
    if (company.value) {
      await loadCompanyReports(id)
      await loadCompanyHierarchy(id)
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat data perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    loading.value = false
  }
}

const loadCompanyHierarchy = async (companyId: string) => {
  try {
    // Try to get ancestors from API first
    try {
      const ancestors = await companyApi.getAncestors(companyId)
      // Ancestors API returns from parent to root (excluding current company)
      // We need to reverse to get from root to parent, then add current company
      companyHierarchy.value = [...ancestors].reverse()
      // Add current company at the end
      if (company.value) {
        companyHierarchy.value.push(company.value)
      }
      return
    } catch (apiError) {
      // If API endpoint doesn't exist (404 or other error), build hierarchy manually
      // Don't log as error if it's just 404 (endpoint not implemented yet)
      if ((apiError as { response?: { status?: number } })?.response?.status !== 404) {
        console.warn('Ancestors API error, building hierarchy manually:', apiError)
      }
    }

    // Fallback: Build hierarchy manually by loading all companies
    if (allCompanies.value.length === 0) {
      allCompanies.value = await companyApi.getAll()
    }

    // Build hierarchy from current company to root
    const hierarchy: Company[] = []
    let currentCompany: Company | undefined = company.value || undefined

    // If company not found in allCompanies, try to find it
    if (!currentCompany) {
      currentCompany = allCompanies.value.find(c => c.id === companyId)
    }

    if (!currentCompany) {
      companyHierarchy.value = []
      return
    }

    // Build hierarchy by traversing parent_id
    const companyMap = new Map<string, Company>()
    allCompanies.value.forEach(c => companyMap.set(c.id, c))

    let current: Company | undefined = currentCompany
    while (current) {
      hierarchy.unshift(current) // Add to beginning to get root -> current order
      
      if (current.parent_id) {
        current = companyMap.get(current.parent_id)
      } else {
        break
      }
    }

    companyHierarchy.value = hierarchy
  } catch (error) {
    console.error('Error loading company hierarchy:', error)
    companyHierarchy.value = []
  }
}

const loadCompanyReports = async (companyId: string) => {
  reportsLoading.value = true
  try {
    const reports = await reportsApi.getByCompanyId(companyId)
    companyReports.value = reports
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    console.error('Gagal memuat data reports:', axiosError.response?.data?.message || axiosError.message || 'Unknown error')
    // Don't show error message to user, just log it
    companyReports.value = []
  } finally {
    reportsLoading.value = false
  }
}

const getMainBusiness = (company: Company): BusinessField | null => {
  if (company.main_business) {
    return company.main_business
  }
  if (company.business_fields && company.business_fields.length > 0) {
    // Find main business field (checking is_main property if it exists)
    const businessFieldsWithMain = company.business_fields as Array<BusinessField & { is_main?: boolean }>
    const mainField = businessFieldsWithMain.find((bf) => bf.is_main)
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
  if (level === 0) return 'Holding'
  return `Level ${String(level)}`
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

const handleDelete = async () => {
  if (!company.value) return

  try {
    await companyApi.delete(company.value.id)
    message.success('Subsidiary berhasil dihapus')
    router.push('/subsidiaries')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal menghapus subsidiary')
  }
}

const handleMenuClick = ({ key }: { key: string }) => {
  if (key === 'edit') {
    handleEdit()
  } else if (key === 'assign-role') {
    openAssignRoleModal()
  } else if (key === 'delete') {
    // Show confirmation before delete
    if (company.value && confirm(`Apakah Anda yakin ingin menghapus ${company.value.name}?`)) {
      handleDelete()
    }
  }
}

const openAssignRoleModal = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }

  assignRoleModalVisible.value = true
  assignRoleForm.value = {
    userId: undefined,
    roleId: undefined,
  }

  // Load users and roles
  // Note: For non-superadmin, users endpoint might return limited results
  // but we still try to load - error will be handled gracefully
  await Promise.all([
    loadUsers(),
    loadRoles()
  ])
}

const loadUsers = async () => {
  if (!company.value) return

  usersLoading.value = true
  try {
    // Load all users (backend will filter based on access) - for dropdown selection
    const allUsersData = await userApi.getAll()
    allUsers.value = allUsersData

    // Load company users from junction table (supports multiple company assignments)
    try {
      const companyUsersData = await companyApi.getUsers(company.value.id)
      companyUsers.value = companyUsersData
    } catch (error: unknown) {
      // Fallback: if endpoint doesn't exist yet, filter from allUsers
      console.warn('Failed to load company users from endpoint, using fallback:', error)
      companyUsers.value = allUsersData.filter(user => user.company_id === company.value?.id)
    }
  } catch (error: unknown) {
    // Better error handling - check for axios error structure
    const axiosError = error as {
      response?: {
        status?: number
        data?: {
          message?: string
          error?: string
        }
      }
      message?: string
      code?: string
    }

    const statusCode = axiosError.response?.status
    const errorMessage = axiosError.response?.data?.message ||
      axiosError.response?.data?.error ||
      axiosError.message ||
      'Unknown error'

    // Handle different error scenarios
    if (statusCode === 403 || statusCode === 401) {
      // Permission denied - silently handle, don't show error to user
      console.warn('Access denied to users endpoint (status:', statusCode, '):', errorMessage)
      allUsers.value = []
      companyUsers.value = []
      // Don't show error message for permission issues
    } else if (statusCode === 404) {
      // Not found - might be endpoint issue
      console.warn('Users endpoint not found:', errorMessage)
      allUsers.value = []
      companyUsers.value = []
    } else if (statusCode && statusCode >= 500) {
      // Server error
      console.error('Server error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Server error')
      allUsers.value = []
      companyUsers.value = []
    } else if (axiosError.code === 'ECONNABORTED' || axiosError.code === 'NETWORK_ERROR') {
      // Network/timeout error
      console.error('Network error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Masalah koneksi')
      allUsers.value = []
      companyUsers.value = []
    } else {
      // Other errors
      console.error('Error loading users:', error)
      // Only show error if it's not a silent permission issue
      if (statusCode !== 403 && statusCode !== 401) {
        message.error('Gagal memuat daftar user: ' + errorMessage)
      }
      allUsers.value = []
      companyUsers.value = []
    }
  } finally {
    usersLoading.value = false
  }
}

// Alias untuk backward compatibility
const loadCompanyUsers = loadUsers

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    allRoles.value = await roleApi.getAll()
  } catch {
    message.error('Gagal memuat daftar role')
  } finally {
    rolesLoading.value = false
  }
}

const handleUserSearch = (value: string) => {
  userSearchText.value = value
}

const handleRoleSearch = (value: string) => {
  roleSearchText.value = value
}

const filterUserOption = (input: string, option: unknown) => {
  const opt = option as { value: string }
  const user = allUsers.value.find(u => u.id === opt.value)
  if (!user) return false
  const search = input.toLowerCase()
  return user.username.toLowerCase().includes(search) ||
    user.email.toLowerCase().includes(search)
}

const filterRoleOption = (input: string, option: unknown) => {
  const opt = option as { value: string }
  const role = allRoles.value.find(r => r.id === opt.value)
  if (!role) return false
  return role.name.toLowerCase().includes(input.toLowerCase())
}

const handleAssignRole = async () => {
  if (!company.value || !assignRoleForm.value.userId || !assignRoleForm.value.roleId) {
    message.error('Harap pilih user dan role')
    return
  }

  assignRoleLoading.value = true
  try {
    await userApi.assignToCompany(
      assignRoleForm.value.userId,
      company.value.id,
      assignRoleForm.value.roleId
    )
    message.success('User berhasil diassign sebagai pengurus')
    assignRoleForm.value = {
      userId: undefined,
      roleId: undefined,
    }
    // Reload company users
    await loadCompanyUsers()
    // Reload company data
    await loadCompany()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal mengassign user')
  } finally {
    assignRoleLoading.value = false
  }
}

// Edit User Role
const handleEditUserRole = async (user: User) => {
  editingUserRole.value = {
    userId: user.id,
    roleId: user.role_id || undefined,
  }
  editingUserRoleModalVisible.value = true
  await loadRoles()
}

const handleCancelEditUserRole = () => {
  editingUserRoleModalVisible.value = false
  editingUserRole.value = null
}

const handleSaveUserRole = async () => {
  if (!editingUserRole.value || !editingUserRole.value.roleId) {
    message.error('Harap pilih role')
    return
  }

  editingRoleLoading.value = true
  try {
    await userApi.update(editingUserRole.value.userId, {
      role_id: editingUserRole.value.roleId,
    })
    message.success('Role pengurus berhasil diubah')
    editingUserRoleModalVisible.value = false
    editingUserRole.value = null
    // Reload company users
    await loadCompanyUsers()
    // Reload company data
    await loadCompany()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal mengubah role')
  } finally {
    editingRoleLoading.value = false
  }
}

// Remove User
const handleRemoveUser = async (user: User) => {
  if (!company.value) return

  // Show confirmation
  const confirmed = confirm(`Apakah Anda yakin ingin menghapus ${user.username} dari pengurus?`)
  if (!confirmed) return

  try {
    // Remove user from company using unassign endpoint (supports multiple company assignments)
    await userApi.unassignFromCompany(user.id, company.value.id)
    message.success('Pengurus berhasil dihapus')
    // Reload company users
    await loadCompanyUsers()
    // Reload company data
    await loadCompany()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal menghapus pengurus')
  }
}

const getUserById = (userId: string): User | undefined => {
  return companyUsers.value.find(u => u.id === userId) || allUsers.value.find(u => u.id === userId)
}

const getRoleColor = (role: string): string => {
  const roleLower = role.toLowerCase()
  if (roleLower.includes('admin')) return 'red'
  if (roleLower.includes('manager')) return 'blue'
  if (roleLower.includes('staff')) return 'green'
  return 'default'
}

const handleManageReports = () => {
  router.push('/reports')
}

// Handle period change
const handlePeriodChange = () => {
  // Filter is automatically applied via computed property
  // No need to reload data, just let computed properties react
}

// Export PDF
const handleExportPDF = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }

  try {
    exportLoading.value = true
    const params: { company_id?: string; period?: string } = {
      company_id: company.value.id,
    }

    if (selectedPeriod.value) {
      params.period = selectedPeriod.value.format('YYYY-MM')
    }

    const blob = await reportsApi.exportPDF(params)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')

    // Generate filename dengan filter info
    let filename = `subsidiary_${company.value.name.replace(/\s+/g, '_')}`
    if (selectedPeriod.value) {
      filename += `_${selectedPeriod.value.format('YYYY-MM')}`
    }
    filename += '.pdf'

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
  } finally {
    exportLoading.value = false
  }
}

// Export Excel
const handleExportExcel = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }

  try {
    exportLoading.value = true
    const params: { company_id?: string; period?: string } = {
      company_id: company.value.id,
    }

    if (selectedPeriod.value) {
      params.period = selectedPeriod.value.format('YYYY-MM')
    }

    const blob = await reportsApi.exportExcel(params)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')

    // Generate filename dengan filter info
    let filename = `subsidiary_${company.value.name.replace(/\s+/g, '_')}`
    if (selectedPeriod.value) {
      filename += `_${selectedPeriod.value.format('YYYY-MM')}`
    }
    filename += '.xlsx'

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
  } finally {
    exportLoading.value = false
  }
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

.back-button-container {
  margin-top: 100px;
  margin-bottom: -100px;
}

.back-button-container .back-button {
  margin-bottom: 24px;
  padding: 0;
  height: auto;
  display: flex;
  align-items: center;
  padding: 5px 8px;
  margin: 24px;
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
  font-size: 16px;
  color: #666;
  margin: -14px 0 16px 0;
}

.company-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  /* background: orange; */
  margin-top: -15px;
}

.meta-item {
  font-size: 14px;
  color: #666;
}

.company-hierarchy {
  /* background: red; */
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: -10px;
  padding-top: 0px;
  /* border-top: 1px solid #f0f0f0; */
}

.hierarchy-label {
  font-size: 12px;
  font-weight: 500;
  color: #333;
}

.hierarchy-path {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.hierarchy-item {
  font-size: 12px;
  color: #333;
}

.hierarchy-separator {
  font-size: 14px;
  color: #999;
  margin: 0 2px;
}

.header-actions {
  flex-shrink: 0;
}

/* Ensure all buttons and date picker in header-actions have consistent height */
.header-actions :deep(.ant-btn) {
  height: 40px !important;
  min-height: 40px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.header-actions :deep(.ant-picker) {
  height: 40px !important;
}

.header-actions :deep(.ant-picker-input) {
  height: 40px !important;
}

.header-actions :deep(.ant-picker-input > input) {
  height: 38px !important;
  line-height: 38px !important;
}

/* Tabs Container */
.tabs-container {
  margin-top: 50px;
  padding: 24px;
  /* background: red; */
}

.tabs-container :deep(.ant-tabs-card) {
  background: transparent;
  background: white;
  margin-top: -50px;
  padding: 24px;
  border-radius: 
  15px;
}

.tabs-container :deep(.ant-tabs-tab) {
  border-radius: 8px 8px 0 0;
}

.tabs-container :deep(.ant-tabs-tab-active) {
  background: white;
}

/* Performance Content */
.performance-content {
  padding: 0px 0 24px 0;
  /* background: orange; */
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
  width: 100%;
}

.trend-card-content :deep(.ant-card-body) {
  padding: 24px;
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
  margin: 0;
  padding: 0;
  overflow: visible;
  display: block;
}

.mini-chart {
  width: 100% !important;
  height: 60px;
  display: block;
  margin: 0;
  padding: 0;
  max-width: 100%;
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
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.recent-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.full-width {
  width: 100%;
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
    flex-direction: column;
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

/* Assign Role Modal Styles */
.assign-role-container {
  max-height: 70vh;
  overflow-y: auto;
}

.current-users-section,
.assign-new-section {
  margin-bottom: 24px;
}

.section-header {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

.text-muted {
  color: #999;
  font-size: 12px;
}
</style>
