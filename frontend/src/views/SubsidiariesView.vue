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
                      <a-menu-item v-if="canEdit" key="edit" @click="handleEditCompany(record.id)">
                        <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                        Edit
                      </a-menu-item>
                      <a-menu-item v-if="canAssignRole" key="assign-role" @click="handleAssignRole(record.id)">
                        <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                        Assign Role
                      </a-menu-item>
                      <a-menu-divider v-if="canDelete && (canEdit || canAssignRole)" />
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

      <!-- Assign Role Modal -->
      <a-modal
        v-model:open="assignRoleModalVisible"
        title="Assign Role - Kelola Pengurus"
        :confirm-loading="assignRoleLoading"
        width="900px"
        :footer="null"
      >
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
                    <a-select
                      v-model:value="assignRoleForm.userId"
                      show-search
                      placeholder="Cari user berdasarkan nama atau email"
                      :filter-option="filterUserOption"
                      :loading="usersLoading"
                      @search="handleUserSearch"
                      allow-clear
                      :disabled="usersLoading"
                    >
                      <a-select-option
                        v-for="user in filteredUsers"
                        :key="user.id"
                        :value="user.id"
                        :disabled="companyUsers.some(u => u.id === user.id)"
                      >
                        {{ user.username }} ({{ user.email }})
                        <span v-if="companyUsers.some(u => u.id === user.id)" class="text-muted"> - Sudah menjadi pengurus</span>
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
                    <a-select
                      v-model:value="assignRoleForm.roleId"
                      show-search
                      placeholder="Cari role"
                      :filter-option="filterRoleOption"
                      :loading="rolesLoading"
                      @search="handleRoleSearch"
                      allow-clear
                      :disabled="rolesLoading"
                    >
                      <a-select-option
                        v-for="role in filteredRoles"
                        :key="role.id"
                        :value="role.id"
                      >
                        {{ role.name }}
                      </a-select-option>
                    </a-select>
                    <small v-if="rolesLoading" class="text-muted">Memuat daftar role...</small>
                  </a-form-item>
                </a-col>
              </a-row>
              
              <a-form-item>
                <a-button 
                  type="primary" 
                  :loading="assignRoleLoading" 
                  @click="handleAssignRoleSubmit" 
                  :disabled="!assignRoleForm.userId || !assignRoleForm.roleId"
                >
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
            <a-table
              :columns="userColumns"
              :data-source="companyUsers"
              :loading="usersLoading"
              :pagination="{ pageSize: 10 }"
              row-key="id"
              size="middle"
            >
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
      <a-modal
        v-model:open="editingUserRoleModalVisible"
        title="Ubah Role Pengurus"
        :confirm-loading="editingRoleLoading"
        @ok="handleSaveUserRole"
        @cancel="handleCancelEditUserRole"
        width="500px"
      >
        <a-form layout="vertical" v-if="editingUserRole">
          <a-form-item label="User">
            <a-input :value="getUserById(editingUserRole.userId)?.username" disabled />
          </a-form-item>
          <a-form-item label="Pilih Role Baru" required>
            <a-select
              v-model:value="editingUserRole.roleId"
              show-search
              placeholder="Cari role"
              :filter-option="filterRoleOption"
              :loading="rolesLoading"
              @search="handleRoleSearch"
            >
              <a-select-option
                v-for="role in filteredRoles"
                :key="role.id"
                :value="role.id"
              >
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
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, userApi, roleApi, type Company, type User, type Role } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import type { TableColumnsType, TableProps } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

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

// Check if any menu item is available (to show/hide Actions dropdown)
const hasAnyMenuOption = computed(() => canEdit.value || canAssignRole.value || canDelete.value)

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
