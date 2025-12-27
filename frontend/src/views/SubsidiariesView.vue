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
            <a-input v-if="viewMode === 'grid'" v-model:value="searchText" placeholder="Search" class="search-input"
              allow-clear>
              <template #prefix>
                <IconifyIcon icon="mdi:account" width="16" />
              </template>
            </a-input>
            <div class="view-mode-buttons">
              <a-button :type="viewMode === 'grid' ? 'primary' : 'default'" @click="handleViewModeChange('grid')"
                class="view-mode-btn">
                <IconifyIcon icon="mdi:view-grid" width="20" />
              </a-button>
              <a-button :type="viewMode === 'list' ? 'primary' : 'default'" @click="handleViewModeChange('list')"
                class="view-mode-btn">
                <IconifyIcon icon="mdi:view-list" width="20" />
              </a-button>
            </div>
            <a-button v-if="isSuperAdmin || isAdministrator || isAdmin" type="primary" @click="handleCreateCompany" style="display: flex; align-items: center;">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Add new Subsidiary
            </a-button>
          </div>
        </div>
      </div>

      <div class="mainContentPage">
        <!-- Subsidiary Cards Grid -->
        <a-row :gutter="[24, 24]" v-if="viewMode === 'grid' && !companiesLoading && filteredCompanies.length > 0">
          <a-col v-for="company in paginatedCompanies" :key="company.id" :xs="24" :sm="12" :md="12" :lg="8" :xl="6">
            <div class="subsidiary-card" :class="{ 'inactive-company': ENABLE_ACTIVATE_DEACTIVATE_FEATURE && !company.is_active }">
            <!-- Card Actions Dropdown -->
            <div class="card-actions" v-if="hasAnyMenuOption" @click.stop>
              <a-dropdown>
                <a-button type="text" size="small" class="card-action-button">
                  <IconifyIcon icon="mdi:dots-vertical" width="20" />
                </a-button>
                <template #overlay>
                  <a-menu @click="(e: { key: string }) => handleCardMenuClick(e.key, company)">
                    <a-menu-item key="view">
                      <IconifyIcon icon="mdi:eye" width="16" style="margin-right: 8px;" />
                      Lihat Detail
                    </a-menu-item>
                    <a-menu-item v-if="canEdit" key="edit">
                      <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                      Edit
                    </a-menu-item>
                    <a-menu-item v-if="canAssignRole" key="assign-role">
                      <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                      Assign Role
                    </a-menu-item>
                    <a-menu-divider v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator) && (canEdit || canAssignRole)" />
                    <a-menu-item 
                      v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)" 
                      :key="company.is_active ? 'deactivate' : 'activate'"
                      :danger="company.is_active"
                      @click.stop="() => handleToggleCompanyStatusFromMenu(company.id, company.name, company.is_active)"
                    >
                      <IconifyIcon :icon="company.is_active ? 'mdi:power-off' : 'mdi:power-on'" width="16" style="margin-right: 8px;" />
                      {{ company.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                    </a-menu-item>
                    <a-menu-divider v-if="canDelete && (canEdit || canAssignRole || (ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)))" />
                    <a-menu-item v-if="canDelete" key="delete" danger>
                      <IconifyIcon icon="mdi:delete" width="16" style="margin-right: 8px;" />
                      Hapus
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
            
            <!-- Card Header -->
            <div class="card-header" @click="handleViewDetail(company.id)">
              <div class="company-icon">
                <img v-if="getCompanyLogo(company)" :src="getCompanyLogo(company)" :alt="company.name"
                  class="logo-image" />
                <div v-else class="icon-placeholder" :style="{ backgroundColor: getIconColor(company.name) }">
                  {{ getCompanyInitial(company.name) }}
                </div>
              </div>
              <div class="company-info">
                <h3 class="company-name">
                  {{ company.name }}
                  <a-tag v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && !company.is_active" color="red" style="margin-left: 8px; font-size: 10px;">
                    Nonaktif
                  </a-tag>
                </h3>
                <p class="company-reg">No Reg {{ company.nib || 'N/A' }}</p>
              </div>
            </div>

            <!-- Card Divider -->
            <div class="card-divider"></div>

            <!-- Card Content -->
            <div class="card-content" @click="handleViewDetail(company.id)">
              <div class="latest-month-header">
                <a-popover :title="getPopoverTitle(company.id)" placement="top" trigger="hover">
                  <template #content>
                    <div style="max-width: 350px;">
                      <div v-html="getPopoverContent(company.id)"></div>
                    </div>
                  </template>
                  <div style="display: flex; align-items: center; cursor: help;">
                    <IconifyIcon icon="mdi:information-outline" width="16" style="margin-right: 4px;" />
                    <span>Latest Month</span>
                  </div>
                </a-popover>
              </div>

              <div class="metrics-row">
                <!-- Net Profit (NPAT) -->
                <div class="metric-item">
                  <div class="metric-value">{{ formatCurrency(getNetProfitData(company.id), company.id) }}</div>
                  <div class="metric-meta">
                    <span class="metric-year">{{ getNetProfitPeriod(company.id) }}</span>
                    <span class="metric-change" :class="getNetProfitChange(company.id) >= 0 ? 'positive' : 'negative'">
                      {{ getNetProfitChange(company.id) >= 0 ? '+' : '' }}{{ getNetProfitChange(company.id) }}%
                    </span>
                  </div>
                  <div class="metric-label">Net Profit</div>
                </div>

                <!-- Financial Health Score -->
                <div class="metric-item">
                  <div class="metric-value">{{ getFinancialHealthScore(company.id) }}</div>
                  <div class="metric-meta">
                    <span class="metric-quarter">{{ getFinancialHealthPeriod(company.id) }}</span>
                    <span class="metric-change" :class="getFinancialHealthStatus(company.id).color">
                      {{ getFinancialHealthStatus(company.id).label }}
                    </span>
                  </div>
                  <div class="metric-label">Financial Health</div>
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
          </a-col>
        </a-row>

        <!-- Subsidiary Table View -->
        <div v-if="viewMode === 'list'">
          <a-card class="subsidiaries-table-card" :bordered="false">
            <!-- Table Filters and Actions -->
            <div class="table-filters-container">
              <a-input v-model:value="searchText" placeholder="Search" class="search-input" allow-clear>
                <template #prefix>
                  <IconifyIcon icon="mdi:magnify" width="16" />
                </template>
              </a-input>
            </div>

            <a-table :columns="tableColumns" :data-source="tableData" :loading="companiesLoading || tableDataLoading"
              :pagination="{
                current: tablePagination.current,
                pageSize: tablePagination.pageSize,
                showSizeChanger: true,
                showTotal: (total: number) => `Total ${total} subsidiaries`,
                pageSizeOptions: ['10', '20', '50', '100'],
              }" @change="handleTableChange" row-key="id" :scroll="{ x: 'max-content' }" class="striped-table">
              <template #emptyText>
                <div class="empty-state" style="padding: 40px 20px;">
                  <IconifyIcon icon="mdi:office-building-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
                  <p v-if="isSuperAdmin || isAdministrator">Belum ada data subsidiary</p>
                  <p v-else>Anda belum di-assign ke perusahaan manapun. Silakan hubungi administrator untuk mendapatkan akses.</p>
                  <a-button v-if="isSuperAdmin || isAdministrator" type="primary" @click="handleCreateCompany" style="margin-top: 16px;">
                    <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
                    Tambah Subsidiary Pertama
                  </a-button>
                </div>
              </template>
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'logo'">
                  <div class="table-logo-cell">
                    <img v-if="getCompanyLogo(record)" :src="getCompanyLogo(record)" :alt="record.name"
                      class="table-logo" />
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
                  <a-switch
                    v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)"
                    :checked="record.is_active"
                    :loading="statusUpdatingIds.has(record.id)"
                    @change="(checked: boolean) => handleToggleCompanyStatus(record.id, record.name, checked)"
                    checked-children="Aktif"
                    un-checked-children="Nonaktif"
                  />
                  <a-tag v-else-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE" :color="record.is_active ? 'green' : 'red'">
                    {{ record.is_active ? 'Aktif' : 'Nonaktif' }}
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
                        <a-menu-item v-if="canEdit" key="edit" @click="handleEditCompany(record.id)">
                          <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                          Edit
                        </a-menu-item>
                        <a-menu-item v-if="canAssignRole" key="assign-role" @click="handleAssignRole(record.id)">
                          <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                          Assign Role
                        </a-menu-item>
                        <a-menu-divider v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator) && (canEdit || canAssignRole)" />
                        <a-menu-item 
                          v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)" 
                          :key="record.is_active ? 'deactivate' : 'activate'"
                          :danger="record.is_active"
                          @click.stop="() => handleToggleCompanyStatusFromMenu(record.id, record.name, record.is_active)"
                        >
                          <IconifyIcon :icon="record.is_active ? 'mdi:power-off' : 'mdi:power-on'" width="16" style="margin-right: 8px;" />
                          {{ record.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                        </a-menu-item>
                        <a-menu-divider v-if="canDelete && (canEdit || canAssignRole || (ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)))" />
                        <a-menu-item v-if="canDelete" key="delete" danger @click="handleDeleteCompany(record.id)">
                          <IconifyIcon icon="mdi:delete" width="16" style="margin-right: 8px;" />
                          Hapus
                        </a-menu-item>
                      </a-menu>
                    </template>
                  </a-dropdown>
                </template>
              </template>
            </a-table>
          </a-card>
        </div>

        <!-- Loading State -->
        <div v-if="companiesLoading && viewMode === 'grid'" class="loading-container">
          <a-spin size="large" />
        </div>

        <!-- Empty State -->
        <div v-if="viewMode === 'grid' && !companiesLoading && companies.length === 0" class="empty-state">
          <IconifyIcon icon="mdi:office-building-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
          <p v-if="isSuperAdmin || isAdministrator">Belum ada data subsidiary</p>
          <p v-else>Anda belum di-assign ke perusahaan manapun. Silakan hubungi administrator untuk mendapatkan akses.</p>
          <a-button v-if="isSuperAdmin || isAdministrator" type="primary" @click="handleCreateCompany">
            <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
            Tambah Subsidiary Pertama
          </a-button>
        </div>

        <!-- No Search Results -->
        <div v-if="viewMode === 'grid' && !companiesLoading && companies.length > 0 && filteredCompanies.length === 0"
          class="empty-state">
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
                <a-button type="primary" :loading="assignRoleLoading" @click="handleAssignRoleSubmit"
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
              :pagination="{ pageSize: 10 }" row-key="id" size="middle" class="striped-table">
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'role'">
                  <a-tag v-if="record.role" :color="getRoleColor(record.role)">
                    {{ record.role }}
                  </a-tag>
                  <span v-else class="text-muted">-</span>
                </template>
                <template v-if="column.key === 'status'">
                  <a-switch
                    v-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin || isAdministrator)"
                    :checked="record.is_active"
                    :loading="statusUpdatingIds.has(record.id)"
                    @change="(checked: boolean) => handleToggleCompanyStatus(record.id, record.name, checked)"
                    checked-children="Aktif"
                    un-checked-children="Nonaktif"
                  />
                  <a-tag v-else-if="ENABLE_ACTIVATE_DEACTIVATE_FEATURE" :color="record.is_active ? 'green' : 'red'">
                    {{ record.is_active ? 'Aktif' : 'Nonaktif' }}
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
import { ref, onMounted, computed, watch, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, userApi, roleApi, type Company, type User, type Role } from '../api/userManagement'
// NOTE: reportsApi (old Reports module) is NOT used anymore - only use financialReportsApi (Input Laporan)
import { financialReportsApi, type FinancialReport } from '../api/financialReports'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

// Feature flag untuk enable/disable activate/deactivate subsidiary feature
const ENABLE_ACTIVATE_DEACTIVATE_FEATURE = false

// Computed: Check user roles
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperAdmin = computed(() => userRole.value === 'superadmin')
const isAdministrator = computed(() => userRole.value === 'administrator')
const isAdmin = computed(() => userRole.value === 'admin')
const isManager = computed(() => userRole.value === 'manager')
const isStaff = computed(() => userRole.value === 'staff')

// RBAC: Assign Role untuk admin/superadmin/administrator
const canAssignRole = computed(() => isAdmin.value || isSuperAdmin.value || isAdministrator.value)

// RBAC: Delete untuk admin/superadmin/administrator
const canDelete = computed(() => isAdmin.value || isSuperAdmin.value || isAdministrator.value)

// RBAC: Edit untuk semua role (staff, manager, admin, superadmin, administrator)
const canEdit = computed(() => isAdmin.value || isManager.value || isStaff.value || isSuperAdmin.value || isAdministrator.value)

const hasAnyMenuOption = computed(() => canEdit.value || canAssignRole.value || canDelete.value || (ENABLE_ACTIVATE_DEACTIVATE_FEATURE && (isSuperAdmin.value || isAdministrator.value)))

// Note: Actions dropdown always shown because "Lihat Detail" menu is always available

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

// Financial reports data per company (from Input Laporan feature)
// NOTE: companyReportsMap (old Reports module) is NOT used anymore
const companyFinancialReportsMap = ref<Record<string, FinancialReport[]>>({})
const financialReportsLoading = ref(false)

// Format period helper
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _formatPeriod = (period: string | undefined): string => {
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

// Get quarter from period
const getQuarterFromPeriod = (period: string | undefined): string => {
  if (!period) return 'Q1 2025'
  const [year, month] = period.split('-')
  if (!year || !month) return 'Q1 2025'
  const monthNum = parseInt(month, 10)
  const quarter = Math.floor((monthNum - 1) / 3) + 1
  return `Q${quarter} ${year}`
}

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

  // Sort: active companies first, then by updated_at (most recent first), fallback to created_at
  return filtered.sort((a, b) => {
    // Active companies come first
    if (a.is_active && !b.is_active) return -1
    if (!a.is_active && b.is_active) return 1
    
    // Then sort by date (newest first)
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

// Get financial data from financial reports (detailed with ratios)
// Net Profit (NPAT) functions - using Financial Reports (Realisasi from Input Laporan tab)
// CRITICAL: Only use data from Input Laporan (Financial Reports), NOT from old Reports module
const getNetProfitData = (companyId: string): number => {
  // ONLY use financial reports (Realisasi) from Input Laporan tab
  // These are the data entered in: Input Laporan > Realisasi (Bulanan) tabs
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    // Filter only Realisasi (monthly reports), exclude RKAP (yearly)
    // Realisasi has is_rkap = false and period format: YYYY-MM
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length > 0) {
      // Get latest Realisasi by period (newest first)
      const latest = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return b.period.localeCompare(a.period)
      })[0]
      
      return latest?.net_profit || 0
    }
  }
  
  // NO FALLBACK to old Reports module - return 0 if no data from Input Laporan
  return 0
}

const getNetProfitPeriod = (companyId: string): string => {
  // ONLY use financial reports (Realisasi) from Input Laporan
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length > 0) {
      const latest = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return b.period.localeCompare(a.period)
      })[0]
      
      if (latest?.period) {
        const [year, month] = latest.period.split('-')
        if (year && month) {
          const months = [
            'Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun',
            'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'
          ]
          const monthIndex = parseInt(month, 10) - 1
          if (monthIndex >= 0 && monthIndex < months.length) {
            return `${months[monthIndex]} ${year}`
          }
        }
        return year || '2025'
      }
    }
  }
  
  // NO FALLBACK - return default if no data
  return '2025'
}

const getNetProfitChange = (companyId: string): number => {
  // ONLY use financial reports (Realisasi) from Input Laporan
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length >= 2) {
      const sorted = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return a.period.localeCompare(b.period)
      })

      const latest = sorted[sorted.length - 1]
      const previous = sorted[sorted.length - 2]

      if (latest && previous) {
        const latestNPAT = latest.net_profit || 0
        const prevNPAT = previous.net_profit || 0

        // Handle edge cases
        if (prevNPAT === 0) {
          if (latestNPAT > 0) return 100
          if (latestNPAT < 0) return -100
          return 0
        }

        // Cap the change percentage to reasonable values (±1000%)
        const change = ((latestNPAT - prevNPAT) / Math.abs(prevNPAT)) * 100
        const cappedChange = Math.max(-1000, Math.min(1000, change))
        return Math.round(cappedChange * 10) / 10
      }
    }
  }
  
  // NO FALLBACK - return 0 if no data or less than 2 periods
  return 0
}

// Financial Health Score functions - using Financial Reports with full ratio data from Input Laporan
// CRITICAL: Only use data from Input Laporan (Financial Reports), NOT from old Reports module
const getFinancialHealthScore = (companyId: string): string => {
  // ONLY use financial reports (Realisasi) from Input Laporan tab
  // These contain full ratio data: ROE, ROI, Current Ratio, Net Profit Margin, etc.
  // Data is entered in: Input Laporan > Realisasi (Bulanan) > Rasio Keuangan tab
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    // Filter only Realisasi (monthly reports with ratio data)
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length > 0) {
      const latest = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return b.period.localeCompare(a.period)
      })[0]

      if (latest) {
        
        const score = calculateFinancialHealthScoreFromFinancialReport(latest)
        return score.grade
      }
    }
  }
  
  // NO FALLBACK to old Reports module - return 'N/A' if no data from Input Laporan
  return 'N/A'
}

const getFinancialHealthPeriod = (companyId: string): string => {
  // ONLY use financial reports (Realisasi) from Input Laporan
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length > 0) {
      const latest = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return b.period.localeCompare(a.period)
      })[0]
      return getQuarterFromPeriod(latest?.period)
    }
  }
  
  // NO FALLBACK - return default if no data
  return 'Q1 2025'
}

const getFinancialHealthStatus = (companyId: string): { label: string; color: string } => {
  // ONLY use financial reports (Realisasi) from Input Laporan
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length > 0) {
    const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
    if (realisasiReports.length > 0) {
      const latest = [...realisasiReports].sort((a, b) => {
        if (!a.period || !b.period) return 0
        return b.period.localeCompare(a.period)
      })[0]

      if (latest) {
        const score = calculateFinancialHealthScoreFromFinancialReport(latest)
        return { label: score.status, color: score.color }
      }
    }
  }
  
  // NO FALLBACK - return 'No Data' if no data from Input Laporan
  return { label: 'No Data', color: 'neutral' }
}

// Calculate Financial Health Score from Financial Report (with full ratio data)
const calculateFinancialHealthScoreFromFinancialReport = (report: FinancialReport): { grade: string; status: string; color: string } => {
  let score = 0

  // Factor 1: Profitability (Net Profit) - 30%
  const netProfit = report.net_profit || 0
  if (netProfit > 0) {
    score += 30
  } else if (netProfit === 0) {
    score += 10
  }

  // Factor 2: Revenue Growth - 20%
  const revenue = report.revenue || 0
  if (revenue > 0) {
    score += 20
  }

  // Factor 3: Financial Ratios - 30% (using actual ratio data)
  let ratioScore = 0
  
  // ROE (Return on Equity) - 10%
  const roe = report.roe || 0
  if (roe >= 15) {
    ratioScore += 10
  } else if (roe >= 10) {
    ratioScore += 7
  } else if (roe >= 5) {
    ratioScore += 5
  } else if (roe > 0) {
    ratioScore += 2
  }
  
  // Current Ratio (Liquidity) - 10%
  const currentRatio = report.current_ratio || 0
  if (currentRatio >= 2) {
    ratioScore += 10
  } else if (currentRatio >= 1.5) {
    ratioScore += 7
  } else if (currentRatio >= 1) {
    ratioScore += 5
  } else if (currentRatio > 0) {
    ratioScore += 2
  }
  
  // Net Profit Margin - 10%
  const netProfitMargin = report.net_profit_margin || 0
  if (netProfitMargin >= 20) {
    ratioScore += 10
  } else if (netProfitMargin >= 10) {
    ratioScore += 7
  } else if (netProfitMargin >= 5) {
    ratioScore += 5
  } else if (netProfitMargin > 0) {
    ratioScore += 2
  }
  
  score += ratioScore

  // Factor 4: Operating Efficiency - 20%
  const operatingExpenses = report.operating_expenses || 0
  if (revenue > 0 && operatingExpenses > 0) {
    const efficiency = ((revenue - operatingExpenses) / revenue) * 100
    if (efficiency >= 20) {
      score += 20
    } else if (efficiency >= 10) {
      score += 15
    } else if (efficiency >= 0) {
      score += 10
    } else {
      score += 5
    }
  } else {
    score += 5
  }

  // Determine grade and status
  let grade: string
  let status: string
  let color: string

  if (score >= 80) {
    grade = 'A'
    status = 'Excellent'
    color = 'positive'
  } else if (score >= 65) {
    grade = 'B'
    status = 'Good'
    color = 'positive'
  } else if (score >= 50) {
    grade = 'C'
    status = 'Fair'
    color = 'neutral'
  } else if (score >= 35) {
    grade = 'D'
    status = 'Poor'
    color = 'negative'
  } else {
    grade = 'F'
    status = 'Critical'
    color = 'negative'
  }

  return { grade, status, color }
}

// Get detailed breakdown for Financial Health Score calculation
const getFinancialHealthBreakdown = (companyId: string): {
  netProfit: number
  revenue: number
  roe: number
  currentRatio: number
  netProfitMargin: number
  operatingEfficiency: number
  score: number
  grade: string
  breakdown: Array<{ factor: string; points: number; maxPoints: number; details: string }>
} | null => {
  const financialReports = companyFinancialReportsMap.value[companyId] || []
  if (financialReports.length === 0) return null
  
  const realisasiReports = financialReports.filter(r => !r.is_rkap && r.period && r.period.includes('-'))
  if (realisasiReports.length === 0) return null
  
  const latest = [...realisasiReports].sort((a, b) => {
    if (!a.period || !b.period) return 0
    return b.period.localeCompare(a.period)
  })[0]
  
  if (!latest) return null
  
  const netProfit = latest.net_profit || 0
  const revenue = latest.revenue || 0
  const roe = latest.roe || 0
  const currentRatio = latest.current_ratio || 0
  const netProfitMargin = latest.net_profit_margin || 0
  const operatingExpenses = latest.operating_expenses || 0
  
  let score = 0
  const breakdown: Array<{ factor: string; points: number; maxPoints: number; details: string }> = []
  
  // Factor 1: Profitability
  let profitPoints = 0
  let profitDetails = ''
  if (netProfit > 0) {
    profitPoints = 30
    profitDetails = `Profit: ${formatCurrency(netProfit, companyId)}`
  } else if (netProfit === 0) {
    profitPoints = 10
    profitDetails = 'Break-even (tidak untung tidak rugi)'
  } else {
    profitPoints = 0
    profitDetails = `Rugi: ${formatCurrency(Math.abs(netProfit), companyId)}`
  }
  score += profitPoints
  breakdown.push({ factor: 'Profitabilitas (Net Profit)', points: profitPoints, maxPoints: 30, details: profitDetails })
  
  // Factor 2: Revenue
  const revenuePoints = revenue > 0 ? 20 : 0
  score += revenuePoints
  breakdown.push({ 
    factor: 'Revenue', 
    points: revenuePoints, 
    maxPoints: 20, 
    details: revenue > 0 ? `Revenue: ${formatCurrency(revenue, companyId)}` : 'Tidak ada revenue' 
  })
  
  // Factor 3: Ratios
  let ratioScore = 0
  let roePoints = 0
  if (roe >= 15) roePoints = 10
  else if (roe >= 10) roePoints = 7
  else if (roe >= 5) roePoints = 5
  else if (roe > 0) roePoints = 2
  ratioScore += roePoints
  
  let currentRatioPoints = 0
  if (currentRatio >= 2) currentRatioPoints = 10
  else if (currentRatio >= 1.5) currentRatioPoints = 7
  else if (currentRatio >= 1) currentRatioPoints = 5
  else if (currentRatio > 0) currentRatioPoints = 2
  ratioScore += currentRatioPoints
  
  let marginPoints = 0
  if (netProfitMargin >= 20) marginPoints = 10
  else if (netProfitMargin >= 10) marginPoints = 7
  else if (netProfitMargin >= 5) marginPoints = 5
  else if (netProfitMargin > 0) marginPoints = 2
  ratioScore += marginPoints
  
  score += ratioScore
  breakdown.push({ 
    factor: 'Rasio Keuangan', 
    points: ratioScore, 
    maxPoints: 30, 
    details: `ROE: ${roe.toFixed(1)}% (${roePoints}p) | Current Ratio: ${currentRatio.toFixed(2)} (${currentRatioPoints}p) | Margin: ${netProfitMargin.toFixed(1)}% (${marginPoints}p)` 
  })
  
  // Factor 4: Efficiency
  let efficiencyPoints = 5
  let efficiencyDetails = 'Tidak ada data'
  if (revenue > 0 && operatingExpenses > 0) {
    const efficiency = ((revenue - operatingExpenses) / revenue) * 100
    if (efficiency >= 20) {
      efficiencyPoints = 20
      efficiencyDetails = `Efisiensi tinggi: ${efficiency.toFixed(1)}%`
    } else if (efficiency >= 10) {
      efficiencyPoints = 15
      efficiencyDetails = `Efisiensi sedang: ${efficiency.toFixed(1)}%`
    } else if (efficiency >= 0) {
      efficiencyPoints = 10
      efficiencyDetails = `Efisiensi rendah: ${efficiency.toFixed(1)}%`
    } else {
      efficiencyPoints = 5
      efficiencyDetails = `Tidak efisien: ${efficiency.toFixed(1)}%`
    }
  }
  score += efficiencyPoints
  breakdown.push({ 
    factor: 'Efisiensi Operasional', 
    points: efficiencyPoints, 
    maxPoints: 20, 
    details: efficiencyDetails 
  })
  
  const result = calculateFinancialHealthScoreFromFinancialReport(latest)
  
  return {
    netProfit,
    revenue,
    roe,
    currentRatio,
    netProfitMargin,
    operatingEfficiency: revenue > 0 && operatingExpenses > 0 ? ((revenue - operatingExpenses) / revenue) * 100 : 0,
    score,
    grade: result.grade,
    breakdown
  }
}

// Get popover title
const getPopoverTitle = (companyId: string): string => {
  const breakdown = getFinancialHealthBreakdown(companyId)
  if (breakdown) {
    return `Penjelasan Metrik - Grade ${breakdown.grade} (Skor: ${breakdown.score}/100)`
  }
  return 'Penjelasan Metrik'
}

// Get popover content with dynamic explanation
const getPopoverContent = (companyId: string): string => {
  const breakdown = getFinancialHealthBreakdown(companyId)
  const netProfit = getNetProfitData(companyId)
  const netProfitPeriod = getNetProfitPeriod(companyId)
  const netProfitChange = getNetProfitChange(companyId)
  
  if (!breakdown) {
    return `
      <div style="font-size: 12px; color: #666;">
        <p>Data belum tersedia. Silakan input data di tab <strong>Input Laporan > Realisasi (Bulanan)</strong>.</p>
      </div>
    `
  }
  
  let html = `
    <div style="font-size: 12px; color: #666; line-height: 1.6;">
      <div style="margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #eee;">
        <strong style="color: #1890ff; display: block; margin-bottom: 4px;">Net Profit (NPAT)</strong>
        <div style="font-size: 11px;">
          <div>Nilai: <strong>${formatCurrency(netProfit, companyId)}</strong></div>
          <div>Periode: ${netProfitPeriod}</div>
          <div>Perubahan: <span style="color: ${netProfitChange >= 0 ? '#52c41a' : '#ff4d4f'}">${netProfitChange >= 0 ? '+' : ''}${netProfitChange}%</span></div>
        </div>
      </div>
      
      <div style="margin-bottom: 8px;">
        <strong style="color: #52c41a; display: block; margin-bottom: 6px;">Financial Health Score: ${breakdown.grade}</strong>
        <div style="font-size: 11px; margin-bottom: 8px;">
          <div>Total Skor: <strong>${breakdown.score}/100</strong></div>
          <div style="margin-top: 4px; color: #999;">Grade ${breakdown.grade} = Skor ${breakdown.score >= 80 ? '80-100' : breakdown.score >= 65 ? '65-79' : breakdown.score >= 50 ? '50-64' : breakdown.score >= 35 ? '35-49' : '0-34'}</div>
        </div>
      </div>
      
      <div style="font-size: 11px;">
        <strong style="display: block; margin-bottom: 6px; color: #333;">Breakdown Perhitungan:</strong>
  `
  
  breakdown.breakdown.forEach((item) => {
    const percentage = Math.round((item.points / item.maxPoints) * 100)
    const color = percentage >= 80 ? '#52c41a' : percentage >= 50 ? '#faad14' : '#ff4d4f'
    html += `
        <div style="margin-bottom: 8px; padding: 6px; background: #f5f5f5; border-radius: 4px;">
          <div style="display: flex; justify-content: space-between; margin-bottom: 4px;">
            <span><strong>${item.factor}</strong></span>
            <span style="color: ${color}; font-weight: 600;">${item.points}/${item.maxPoints}</span>
          </div>
          <div style="font-size: 10px; color: #666; margin-top: 2px;">${item.details}</div>
        </div>
    `
  })
  
  html += `
      </div>
    </div>
  `
  
  return html
}

// NOTE: This function is DEPRECATED - no longer used
// We only use calculateFinancialHealthScoreFromFinancialReport which uses data from Input Laporan
// Keeping this for reference but it should never be called
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _calculateFinancialHealthScore_DEPRECATED = (_report: unknown): { grade: string; status: string; color: string } => {
  let score = 0
  const factors: string[] = []
  
  // Type assertion untuk deprecated function
  const report = _report as { npat?: number; revenue?: number; financial_ratio?: number; opex?: number }

  // Factor 1: Profitability (NPAT) - 30%
  const npat = report.npat || 0
  if (npat > 0) {
    score += 30
    factors.push('Profitable')
  } else if (npat === 0) {
    score += 10
    factors.push('Break-even')
  } else {
    factors.push('Loss')
  }

  // Factor 2: Revenue Growth - 25%
  // We'll use financial_ratio as a proxy if available, or calculate from revenue
  const revenue = report.revenue || 0
  if (revenue > 0) {
    score += 20
    factors.push('Revenue')
  }

  // Factor 3: Financial Ratio (if available) - 25%
  const financialRatio = report.financial_ratio || 0
  if (financialRatio > 0) {
    if (financialRatio >= 80) {
      score += 25
      factors.push('Excellent Ratio')
    } else if (financialRatio >= 60) {
      score += 20
      factors.push('Good Ratio')
    } else if (financialRatio >= 40) {
      score += 15
      factors.push('Fair Ratio')
    } else {
      score += 5
      factors.push('Poor Ratio')
    }
  } else {
    score += 10 // Default if no ratio data
  }

  // Factor 4: Operating Efficiency (Revenue vs Opex) - 20%
  const opex = report.opex || 0
  if (revenue > 0 && opex > 0) {
    const efficiency = ((revenue - opex) / revenue) * 100
    if (efficiency >= 20) {
      score += 20
      factors.push('High Efficiency')
    } else if (efficiency >= 10) {
      score += 15
      factors.push('Moderate Efficiency')
    } else if (efficiency >= 0) {
      score += 10
      factors.push('Low Efficiency')
    } else {
      score += 5
      factors.push('Inefficient')
    }
  } else {
    score += 5 // Default if no data
  }

  // Determine grade and status
  let grade: string
  let status: string
  let color: string

  if (score >= 80) {
    grade = 'A'
    status = 'Excellent'
    color = 'positive'
  } else if (score >= 65) {
    grade = 'B'
    status = 'Good'
    color = 'positive'
  } else if (score >= 50) {
    grade = 'C'
    status = 'Fair'
    color = 'neutral'
  } else if (score >= 35) {
    grade = 'D'
    status = 'Poor'
    color = 'negative'
  } else {
    grade = 'F'
    status = 'Critical'
    color = 'negative'
  }

  return { grade, status, color }
}

// Format currency (handles negative values) - uses company currency setting
const formatCurrency = (value: number, companyId?: string): string => {
  const absValue = Math.abs(value)
  const sign = value < 0 ? '-' : ''
  
  // Get currency from company, default to IDR (Rupiah)
  let currency = 'IDR'
  if (companyId) {
    const company = companies.value.find(c => c.id === companyId)
    currency = company?.currency || 'IDR'
  }
  
  // Format based on currency
  if (currency === 'USD') {
    // USD format: $32B, $129M, $5K
    if (absValue >= 1000000000) {
      return `${sign}$${(absValue / 1000000000).toFixed(0)}B`
    } else if (absValue >= 1000000) {
      return `${sign}$${(absValue / 1000000).toFixed(0)}M`
    } else if (absValue >= 1000) {
      return `${sign}$${(absValue / 1000).toFixed(0)}K`
    }
    return `${sign}$${absValue.toFixed(0)}`
  } else {
    // IDR format: 129M, 32B, 5K (no currency symbol prefix, just number with suffix)
    if (absValue >= 1000000000) {
      return `${sign}${(absValue / 1000000000).toFixed(0)}B`
    } else if (absValue >= 1000000) {
      return `${sign}${(absValue / 1000000).toFixed(0)}M`
    } else if (absValue >= 1000) {
      return `${sign}${(absValue / 1000).toFixed(0)}K`
    }
    // For values less than 1000, show full number with Rp prefix
    return `${sign}Rp ${absValue.toLocaleString('id-ID')}`
  }
}

// Get company logo atau generate icon
const getCompanyLogo = (company: Company): string | undefined => {
  if (company.logo) {
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
    // Superadmin/Administrator melihat semua companies (termasuk yang nonaktif)
    // User lain hanya melihat companies yang di-assign ke mereka
    // Backend akan selalu include inactive companies in listing
    // But calculations/aggregations tetap exclude inactive companies
    if (isSuperAdmin.value || isAdministrator.value) {
      companies.value = await companyApi.getAll(true) // Always include inactive for listing
    } else {
      // Get companies assigned to current user
      const userCompanies = await userApi.getMyCompanies()
      // Convert UserCompanyResponse[] to Company[]
      companies.value = userCompanies.map(uc => uc.company)
    }
    // ONLY load financial reports from Input Laporan (new feature)
    // DO NOT load old Reports module data
    await loadAllCompanyFinancialReports()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    companiesLoading.value = false
  }
}

// Load financial reports for all companies (with full ratio data)
const loadAllCompanyFinancialReports = async () => {
  if (companies.value.length === 0) return

  financialReportsLoading.value = true
  try {
    // Load financial reports for all companies in parallel
    const financialReportPromises = companies.value.map(async (company) => {
      try {
        const financialReports = await financialReportsApi.getByCompanyId(company.id)
        // Filter and sort: prioritize Realisasi (monthly) over RKAP (yearly)
        // Sort by period descending to get latest first
        const sortedReports = [...financialReports].sort((a, b) => {
          if (!a.period || !b.period) return 0
          return b.period.localeCompare(a.period)
        })
        companyFinancialReportsMap.value[company.id] = sortedReports
        
        // Debug: log data untuk memastikan data terambil dengan benar
        const realisasiReports = sortedReports.filter(r => !r.is_rkap)
        if (realisasiReports.length > 0) {
          const latest = realisasiReports[0]
          if (latest) {
          }
        }
      } catch (error) {
        // Silently fail for individual companies - just log it
        console.warn(`Failed to load financial reports for company ${company.id}:`, error)
        companyFinancialReportsMap.value[company.id] = []
      }
    })

    await Promise.all(financialReportPromises)
  } catch (error) {
    console.error('Error loading company financial reports:', error)
  } finally {
    financialReportsLoading.value = false
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
    // If companies are already loaded, just reset pagination
    // Note: Ant Design Table will handle pagination after filtering
    tablePagination.value.current = 1
  }
}

// Computed untuk table data
// IMPORTANT: Ant Design Table akan melakukan filtering sendiri berdasarkan onFilter di columns
// Jadi kita gunakan companies.value langsung (semua data), bukan filteredCompanies
// filteredCompanies hanya untuk search text filter di grid view
// Ant Design Table akan handle pagination secara internal setelah filtering
const tableData = computed(() => {
  return companies.value
})

// Load table data (lazy loading) - hanya set loading state
const loadTableData = async () => {
  if (companies.value.length === 0) {
    await loadCompanies()
  }

  tableDataLoading.value = true
  try {
    // Reset to first page
    // Note: Ant Design Table will handle pagination and total count after filtering
    tablePagination.value.current = 1
  } catch {
    message.error('Gagal memuat data table')
  } finally {
    tableDataLoading.value = false
  }
}

// Watch for changes in companies to reset table pagination
// Note: Ant Design Table will handle pagination and total count after filtering internally
watch([companies, viewMode], () => {
  if (viewMode.value === 'list') {
    // Reset to first page when companies change
    tablePagination.value.current = 1
  }
})


// Table Columns
const tableColumns = computed<TableColumnsType>(() => {
  const baseColumns: TableColumnsType = [
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
        { text: 'Holding', value: 0 },
        { text: 'Level 1', value: 1 },
        { text: 'Level 2', value: 2 },
        { text: 'Level 3', value: 3 },
      ],
      onFilter: (value: string | number | boolean, record: Company) => {
        if (typeof value === 'number') {
          return record.level === value
        }
        return false
      },
    },
  ]

  // Add Status column only if feature is enabled
  if (ENABLE_ACTIVATE_DEACTIVATE_FEATURE) {
    baseColumns.push({
      title: 'Status',
      dataIndex: 'is_active',
      key: 'status',
      width: 150,
      filters: [
        { text: 'Aktif', value: true },
        { text: 'Nonaktif', value: false },
      ],
      onFilter: (value: string | number | boolean, record: Company) => {
        // Ant Design Table passes the filter value directly
        // Ensure value is boolean
        const filterValue = typeof value === 'boolean' ? value : Boolean(value)
        
        // Ensure record.is_active is a boolean (default to true if undefined)
        const recordIsActive = record.is_active !== undefined ? Boolean(record.is_active) : true
        
        return recordIsActive === filterValue
      },
      filterMultiple: false, // Allow only one filter at a time
    })
  }

  baseColumns.push({
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
  })

  baseColumns.push({
    title: 'Aksi',
    key: 'actions',
    width: 120,
    fixed: 'right',
  })

  return baseColumns
})

// Table Change Handler
const handleTableChange: TableProps['onChange'] = (pagination) => {
  if (pagination) {
    tablePagination.value.current = pagination.current || 1
    tablePagination.value.pageSize = pagination.pageSize || 10
  }
}

// Get Level Label
const getLevelLabel = (level: number): string => {
  if (level === 0) return 'Holding'
  return `Level ${level}`
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

// Assign Role Modal State
const assignRoleModalVisible = ref(false)
const assignRoleLoading = ref(false)
const selectedCompanyId = ref<string | null>(null)
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

const handleAssignRole = async (id: string) => {
  selectedCompanyId.value = id
  await openAssignRoleModal(id)
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

// Status updating state
const statusUpdatingIds = ref<Set<string>>(new Set())

// Handle toggle company status from switch
const handleToggleCompanyStatus = async (id: string, name: string, newStatus: boolean) => {
  const actionText = newStatus ? 'mengaktifkan' : 'menonaktifkan'
  
  Modal.confirm({
    title: `${newStatus ? 'Aktifkan' : 'Nonaktifkan'} Subsidiary?`,
    content: `Apakah Anda yakin ingin ${actionText} subsidiary "${name}"? Subsidiary yang dinonaktifkan tidak akan muncul di perhitungan, dashboard, dan aggregasi data.`,
    okText: 'Ya',
    cancelText: 'Batal',
      onOk: async () => {
      statusUpdatingIds.value.add(id)
      try {
        const updatedCompany = await companyApi.updateStatus(id, newStatus)
        // Update company in list
        const index = companies.value.findIndex(c => c.id === id)
        if (index !== -1) {
          companies.value[index] = updatedCompany
        } else {
          // If company not found in list, reload to ensure data consistency
          await loadCompanies()
        }
        message.success(`Subsidiary berhasil ${actionText}`)
      } catch (error: unknown) {
        const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
        message.error(`Gagal ${actionText} subsidiary: ${axiosError.response?.data?.message || axiosError.message || 'Unknown error'}`)
        // Reload companies to get correct state
        await loadCompanies()
      } finally {
        statusUpdatingIds.value.delete(id)
      }
    },
  })
}

// Handle toggle company status from menu
const handleToggleCompanyStatusFromMenu = async (id: string, name: string, currentStatus: boolean) => {
  await handleToggleCompanyStatus(id, name, !currentStatus)
}

// Handle card menu click
const handleCardMenuClick = (key: string, company: Company) => {
  if (key === 'view') {
    handleViewDetail(company.id)
  } else if (key === 'edit') {
    handleEditCompany(company.id)
  } else if (key === 'assign-role') {
    handleAssignRole(company.id)
  } else if (key === 'delete') {
    handleDeleteCompany(company.id)
  }
  // activate/deactivate handled directly in menu item with @click
}

// Assign Role Modal Functions
const openAssignRoleModal = async (companyId: string) => {
  if (!companyId) {
    message.error('Company ID tidak ditemukan')
    return
  }

  selectedCompanyId.value = companyId
  assignRoleModalVisible.value = true
  assignRoleForm.value = {
    userId: undefined,
    roleId: undefined,
  }

  // Load users and roles
  await Promise.all([
    loadUsers(companyId),
    loadRoles()
  ])
}

const loadUsers = async (companyId: string) => {
  if (!companyId) return

  usersLoading.value = true
  try {
    // Load all users (backend will filter based on access) - for dropdown selection
    const allUsersData = await userApi.getAll()
    allUsers.value = allUsersData

    // Load company users from junction table (supports multiple company assignments)
    try {
      const companyUsersData = await companyApi.getUsers(companyId)
      companyUsers.value = companyUsersData
    } catch (error: unknown) {
      // Fallback: if endpoint doesn't exist yet, filter from allUsers
      console.warn('Failed to load company users from endpoint, using fallback:', error)
      companyUsers.value = allUsersData.filter(user => user.company_id === companyId)
    }
  } catch (error: unknown) {
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
      console.warn('Access denied to users endpoint (status:', statusCode, '):', errorMessage)
      allUsers.value = []
      companyUsers.value = []
    } else if (statusCode === 404) {
      console.warn('Users endpoint not found:', errorMessage)
      allUsers.value = []
      companyUsers.value = []
    } else if (statusCode && statusCode >= 500) {
      console.error('Server error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Server error')
      allUsers.value = []
      companyUsers.value = []
    } else if (axiosError.code === 'ECONNABORTED' || axiosError.code === 'NETWORK_ERROR') {
      console.error('Network error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Masalah koneksi')
      allUsers.value = []
      companyUsers.value = []
    } else {
      console.error('Error loading users:', error)
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

const handleAssignRoleSubmit = async () => {
  if (!selectedCompanyId.value || !assignRoleForm.value.userId || !assignRoleForm.value.roleId) {
    message.error('Harap pilih user dan role')
    return
  }

  assignRoleLoading.value = true
  try {
    await userApi.assignToCompany(
      assignRoleForm.value.userId,
      selectedCompanyId.value,
      assignRoleForm.value.roleId
    )
    message.success('User berhasil diassign sebagai pengurus')
    assignRoleForm.value = {
      userId: undefined,
      roleId: undefined,
    }
    // Reload company users
    await loadUsers(selectedCompanyId.value)
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
  if (!editingUserRole.value || !editingUserRole.value.roleId || !selectedCompanyId.value) {
    message.error('Harap pilih role')
    return
  }

  editingRoleLoading.value = true
  try {
    await userApi.assignToCompany(
      editingUserRole.value.userId,
      selectedCompanyId.value,
      editingUserRole.value.roleId
    )
    message.success('Role pengurus berhasil diubah')
    editingUserRoleModalVisible.value = false
    editingUserRole.value = null
    // Reload company users
    await loadUsers(selectedCompanyId.value)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal mengubah role')
  } finally {
    editingRoleLoading.value = false
  }
}

// Remove User
const handleRemoveUser = async (user: User) => {
  if (!selectedCompanyId.value) return

  // Show confirmation
  Modal.confirm({
    title: 'Hapus Pengurus',
    content: `Apakah Anda yakin ingin menghapus ${user.username} dari pengurus?`,
    okText: 'Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      try {
        // Remove user from company using unassign endpoint (supports multiple company assignments)
        await userApi.unassignFromCompany(user.id, selectedCompanyId.value!)
        message.success('Pengurus berhasil dihapus')
        // Reload company users
        await loadUsers(selectedCompanyId.value!)
      } catch (error: unknown) {
        const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
        message.error(axiosError.response?.data?.message || 'Gagal menghapus pengurus')
      }
    },
  })
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

onMounted(async () => {
  await loadCompanies()

  // If view mode is 'list', load table data
  if (viewMode.value === 'list') {
    await loadTableData()
  }
})

// Refresh financial reports when component is activated (e.g., returning from detail page)
onActivated(async () => {
  if (companies.value.length > 0) {
    await loadAllCompanyFinancialReports()
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

.view-mode-buttons {
  display: flex;
  gap: 8px;
  /* border: 1px solid #d9d9d9; */
  border-radius: 8px;
  padding: 4px;
  /* background: #fafafa; */
  height: 40px;
  align-items: center;
}

.view-mode-btn {
  height: 32px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Semua button di header-right harus sama tinggi (40px default) */
.header-right :deep(.ant-btn) {
  height: 40px !important;
  min-height: 40px !important;
}

/* Subsidiaries Table Card */
.subsidiaries-table-card {
  border-radius: 12px;
  overflow: hidden;
  background: white;
}

.subsidiaries-table-card :deep(.ant-card-body) {
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

.table-filters-container .search-input {
  flex: 1;
  min-width: 200px;
  max-width: 300px;
}

/* Cards Grid - Now using Ant Design Row/Col */

.subsidiary-card.inactive-company {
  opacity: 0.6;
  border: 1px dashed #d9d9d9;
}

.subsidiary-card.inactive-company:hover {
  opacity: 0.8;
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
  position: relative;
}

.subsidiary-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.card-actions {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 10;
}

.card-action-button {
  opacity: 0.6;
  transition: opacity 0.2s;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(4px);
}

.subsidiary-card:hover .card-action-button {
  opacity: 1;
  background: rgba(255, 255, 255, 1);
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

.metric-change.neutral {
  color: #faad14;
  background: rgba(250, 173, 20, 0.1);
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

/* Striped table styles */
.striped-table :deep(.ant-table-tbody > tr:nth-child(even) > td),
.striped-table :deep(.ant-table-tbody tr:nth-child(even) td) {
  background-color: #fafafa !important;
}

.striped-table :deep(.ant-table-tbody > tr:nth-child(odd) > td),
.striped-table :deep(.ant-table-tbody tr:nth-child(odd) td) {
  background-color: #ffffff !important;
}

.striped-table :deep(.ant-table-tbody > tr:hover > td),
.striped-table :deep(.ant-table-tbody tr:hover td) {
  background-color: #e6f7ff !important;
}
</style>
