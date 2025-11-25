<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, userApi, roleApi, permissionApi, type Company, type User, type Role, type Permission } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// Active tab
const activeTab = ref('companies')

// Companies
const companies = ref<Company[]>([])
const companiesLoading = ref(false)
const companyModalVisible = ref(false)
const companyForm = ref<Partial<Company>>({})
const editingCompany = ref<Company | null>(null)

// Users
const users = ref<User[]>([])
const usersLoading = ref(false)
const userModalVisible = ref(false)
const userForm = ref<Partial<User & { password: string }>>({})
const editingUser = ref<User | null>(null)
const resetPasswordModalVisible = ref(false)
const resetPasswordForm = ref<{ user_id: string; username: string; new_password: string; confirm_password: string }>({
  user_id: '',
  username: '',
  new_password: '',
  confirm_password: '',
})

// Search states
const companySearchText = ref('')
const userSearchText = ref('')

// Pagination
const companyPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

const userPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

// Computed untuk filtered data dengan search
const filteredCompanies = computed(() => {
  let filtered = [...companies.value]
  
  // Search filter
  if (companySearchText.value) {
    const search = companySearchText.value.toLowerCase()
    filtered = filtered.filter(c => 
      c.name.toLowerCase().includes(search) ||
      c.code.toLowerCase().includes(search) ||
      (c.description && c.description.toLowerCase().includes(search))
    )
  }
  
  // Update pagination total
  companyPagination.value.total = filtered.length
  
  return filtered
})

const filteredUsers = computed(() => {
  let filtered = [...users.value]
  
  // Search filter
  if (userSearchText.value) {
    const search = userSearchText.value.toLowerCase()
    filtered = filtered.filter(u => 
      u.username.toLowerCase().includes(search) ||
      u.email.toLowerCase().includes(search) ||
      u.role.toLowerCase().includes(search)
    )
  }
  
  // Update pagination total
  userPagination.value.total = filtered.length
  
  return filtered
})

// Update pagination total when data changes
watch(companies, () => {
  if (!companySearchText.value) {
    companyPagination.value.total = companies.value.length
  }
}, { immediate: true })

watch(users, () => {
  if (!userSearchText.value) {
    userPagination.value.total = users.value.length
  }
}, { immediate: true })

// Watch search text changes
watch(companySearchText, () => {
  companyPagination.value.current = 1 // Reset to first page on search
})

watch(userSearchText, () => {
  userPagination.value.current = 1 // Reset to first page on search
})

// Filter roles untuk exclude superadmin
const availableRoles = computed(() => {
  return roles.value.filter(r => r.name !== 'superadmin')
})

// Check if current user is superadmin
const isCurrentUserSuperadmin = computed(() => {
  return authStore.user?.role === 'superadmin'
})

// Check if user is superadmin (for edit/delete protection)
const isUserSuperadmin = (user: User) => {
  return user.role === 'superadmin' || user.role_id === roles.value.find(r => r.name === 'superadmin')?.id
}

// Check if user is current logged in user
const isCurrentUser = (user: User) => {
  return user.id === authStore.user?.id
}

// Table columns
const companyColumns = computed(() => [
  { 
    title: 'Nama Perusahaan', 
    dataIndex: 'name', 
    key: 'name', 
    sorter: (a: Company, b: Company) => a.name.localeCompare(b.name),
  },
  { title: 'Kode', dataIndex: 'code', key: 'code', sorter: (a: Company, b: Company) => a.code.localeCompare(b.code) },
  { title: 'Tingkat', dataIndex: 'level', key: 'level', sorter: (a: Company, b: Company) => a.level - b.level },
  { title: 'Deskripsi', dataIndex: 'description', key: 'description' },
  { 
    title: 'Status', 
    dataIndex: 'is_active', 
    key: 'is_active', 
    filters: [
      { text: 'Aktif', value: true },
      { text: 'Tidak Aktif', value: false }
    ], 
    onFilter: (value: boolean, record: Company) => record.is_active === value 
  },
  { title: 'Aksi', key: 'actions', width: 150 },
])

const userColumns = computed(() => [
  { 
    title: 'Nama Pengguna', 
    dataIndex: 'username', 
    key: 'username', 
    sorter: (a: User, b: User) => a.username.localeCompare(b.username),
  },
  { title: 'Email', dataIndex: 'email', key: 'email', sorter: (a: User, b: User) => a.email.localeCompare(b.email) },
  { 
    title: 'Peran', 
    dataIndex: 'role', 
    key: 'role', 
    filters: roles.value.filter(r => r.name !== 'superadmin').map(r => ({ text: r.name, value: r.name })), 
    onFilter: (value: string, record: User) => record.role === value 
  },
  { title: 'Perusahaan', dataIndex: 'company_id', key: 'company_id' },
  { 
    title: 'Status', 
    dataIndex: 'is_active', 
    key: 'is_active', 
    filters: [
      { text: 'Aktif', value: true },
      { text: 'Tidak Aktif', value: false }
    ], 
    onFilter: (value: boolean, record: User) => record.is_active === value 
  },
  { title: 'Aksi', key: 'actions', width: 150 },
])

// Roles
const roles = ref<Role[]>([])
const rolesLoading = ref(false)

// Permissions
const permissions = ref<Permission[]>([])
const permissionsLoading = ref(false)

// Load data
const loadCompanies = async () => {
  companiesLoading.value = true
  try {
    companies.value = await companyApi.getAll()
  } catch (error: any) {
    message.error('Gagal memuat companies: ' + (error.response?.data?.message || error.message))
  } finally {
    companiesLoading.value = false
  }
}

const loadUsers = async () => {
  usersLoading.value = true
  try {
    users.value = await userApi.getAll()
  } catch (error: any) {
    message.error('Gagal memuat users: ' + (error.response?.data?.message || error.message))
  } finally {
    usersLoading.value = false
  }
}

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    roles.value = await roleApi.getAll()
  } catch (error: any) {
    message.error('Gagal memuat roles: ' + (error.response?.data?.message || error.message))
  } finally {
    rolesLoading.value = false
  }
}

const loadPermissions = async () => {
  permissionsLoading.value = true
  try {
    permissions.value = await permissionApi.getAll()
  } catch (error: any) {
    message.error('Gagal memuat permissions: ' + (error.response?.data?.message || error.message))
  } finally {
    permissionsLoading.value = false
  }
}

// Company handlers
const handleCreateCompany = () => {
  editingCompany.value = null
  companyForm.value = { level: 0 }
  companyModalVisible.value = true
}

const handleEditCompany = (company: Company) => {
  editingCompany.value = company
  companyForm.value = { ...company }
  companyModalVisible.value = true
}

const handleSaveCompany = async () => {
  try {
    if (editingCompany.value) {
      await companyApi.update(editingCompany.value.id, {
        name: companyForm.value.name!,
        description: companyForm.value.description,
      })
      message.success('Company berhasil diupdate')
    } else {
      await companyApi.create({
        name: companyForm.value.name!,
        code: companyForm.value.code!,
        description: companyForm.value.description,
        parent_id: companyForm.value.parent_id,
      })
      message.success('Company berhasil dibuat')
    }
    companyModalVisible.value = false
    loadCompanies()
  } catch (error: any) {
    message.error('Gagal menyimpan company: ' + (error.response?.data?.message || error.message))
  }
}

const handleDeleteCompany = async (id: string) => {
  try {
    await companyApi.delete(id)
    message.success('Company berhasil dihapus')
    loadCompanies()
  } catch (error: any) {
    message.error('Gagal menghapus company: ' + (error.response?.data?.message || error.message))
  }
}

// User handlers
const handleCreateUser = () => {
  editingUser.value = null
  userForm.value = {}
  userModalVisible.value = true
}

const handleEditUser = (user: User) => {
  editingUser.value = user
  userForm.value = { ...user }
  userModalVisible.value = true
}

const handleSaveUser = async () => {
  try {
    if (editingUser.value) {
      await userApi.update(editingUser.value.id, {
        username: userForm.value.username,
        email: userForm.value.email,
        company_id: userForm.value.company_id,
        role_id: userForm.value.role_id,
      })
      message.success('User berhasil diupdate')
    } else {
      if (!userForm.value.password) {
        message.error('Password wajib diisi')
        return
      }
      await userApi.create({
        username: userForm.value.username!,
        email: userForm.value.email!,
        password: userForm.value.password!,
        company_id: userForm.value.company_id,
        role_id: userForm.value.role_id,
      })
      message.success('User berhasil dibuat')
    }
    userModalVisible.value = false
    loadUsers()
  } catch (error: any) {
    message.error('Gagal menyimpan user: ' + (error.response?.data?.message || error.message))
  }
}

const handleDeleteUser = async (id: string) => {
  try {
    await userApi.delete(id)
    message.success('User berhasil dihapus')
    loadUsers()
  } catch (error: any) {
    message.error('Gagal menghapus user: ' + (error.response?.data?.message || error.message))
  }
}

const handleToggleUserStatus = async (user: User) => {
  try {
    const updatedUser = await userApi.toggleStatus(user.id)
    message.success(`User berhasil ${updatedUser.is_active ? 'diaktifkan' : 'dinonaktifkan'}`)
    loadUsers()
  } catch (error: any) {
    message.error('Gagal mengubah status user: ' + (error.response?.data?.message || error.message))
  }
}

const handleResetPassword = (user: User) => {
  resetPasswordForm.value = {
    user_id: user.id,
    username: user.username,
    new_password: '',
    confirm_password: '',
  }
  resetPasswordModalVisible.value = true
}

const handleSaveResetPassword = async () => {
  if (!resetPasswordForm.value.new_password || resetPasswordForm.value.new_password.length < 8) {
    message.error('Password harus minimal 8 karakter')
    return
  }
  if (resetPasswordForm.value.new_password !== resetPasswordForm.value.confirm_password) {
    message.error('Password dan konfirmasi password tidak cocok')
    return
  }
  try {
    await userApi.resetPassword(resetPasswordForm.value.user_id, resetPasswordForm.value.new_password)
    message.success('Password berhasil direset')
    resetPasswordModalVisible.value = false
    resetPasswordForm.value = {
      user_id: '',
      username: '',
      new_password: '',
      confirm_password: '',
    }
  } catch (error: any) {
    message.error('Gagal reset password: ' + (error.response?.data?.message || error.message))
  }
}

// Load data on mount
onMounted(() => {
  loadCompanies()
  loadUsers()
  loadRoles()
  loadPermissions()
})

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// Helper functions untuk level label
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

// Helper untuk mendapatkan nama company dari ID
const getCompanyName = (companyId: string): string => {
  const company = companies.value.find(c => c.id === companyId)
  return company?.name || ''
}

// Helper untuk scope label
const getScopeLabel = (scope: string): string => {
  switch (scope) {
    case 'global':
      return 'Global (Seluruh Sistem)'
    case 'company':
      return 'Perusahaan'
    case 'sub_company':
      return 'Sub Perusahaan'
    default:
      return scope
  }
}

const getScopeColor = (scope: string): string => {
  switch (scope) {
    case 'global':
      return 'red'
    case 'company':
      return 'blue'
    case 'sub_company':
      return 'green'
    default:
      return 'default'
  }
}
</script>

<template>
  <div class="user-management-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="user-management-content">
      <h1 class="page-title">Manajemen Pengguna</h1>

      <a-tabs v-model:activeKey="activeTab" class="management-tabs">
        <!-- Users Tab -->
        <a-tab-pane key="users" tab="Pengguna">
          <div class="table-header">
            <a-button type="primary" @click="handleCreateUser">
              <template #icon>
                <span>+</span>
              </template>
              Tambah Pengguna
            </a-button>
          </div>

          <div style="margin-bottom: 16px;">
            <a-input
              v-model:value="userSearchText"
              placeholder="Cari pengguna (username, email, role)..."
              allow-clear
              style="width: 300px;"
            >
              <template #prefix>
                <span>🔍</span>
              </template>
            </a-input>
          </div>

          <a-table
            :columns="userColumns"
            :data-source="filteredUsers"
            :loading="usersLoading"
            :pagination="{
              current: userPagination.current,
              pageSize: userPagination.pageSize,
              total: userPagination.total,
              showSizeChanger: true,
              showTotal: (total: number) => `Total ${total} user`,
              pageSizeOptions: ['10', '20', '50', '100'],
            }"
            @change="(pagination: any) => { 
              userPagination.current = pagination.current || 1
              userPagination.pageSize = pagination.pageSize || 10
            }"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'company_id'">
                <span v-if="record.company_id">
                  {{ getCompanyName(record.company_id) || record.company_id }}
                </span>
                <a-tag v-else color="purple">Superadmin (Global)</a-tag>
              </template>
              <template v-if="column.key === 'is_active'">
                <a-switch
                  :checked="record.is_active"
                  :disabled="isCurrentUser(record) && isUserSuperadmin(record)"
                  @change="() => handleToggleUserStatus(record)"
                  :checked-children="'Aktif'"
                  :un-checked-children="'Nonaktif'"
                />
              </template>
              <template v-if="column.key === 'actions'">
                <a-space v-if="!isCurrentUser(record) || !isUserSuperadmin(record)">
                  <a-button 
                    type="link" 
                    size="small" 
                    @click="handleEditUser(record)"
                  >
                    Edit
                  </a-button>
                  <a-button 
                    v-if="isCurrentUserSuperadmin && !isCurrentUser(record)"
                    type="link" 
                    size="small" 
                    @click="handleResetPassword(record)"
                  >
                    Atur Ulang Password
                  </a-button>
                  <a-popconfirm
                    title="Apakah Anda yakin ingin menghapus user ini?"
                    @confirm="handleDeleteUser(record.id)"
                  >
                    <a-button type="link" size="small" danger>Hapus</a-button>
                  </a-popconfirm>
                </a-space>
                <span v-else class="no-action-text">
                  Tidak dapat diubah
                </span>
              </template>
            </template>
          </a-table>

          <!-- Info Penjelasan Istilah Teknis - Users -->
          <a-collapse class="info-accordion" :bordered="false">
            <a-collapse-panel key="1" header="📚 Penjelasan Istilah Teknis">
              <div class="info-content">
                <div class="info-item">
                  <strong>Pengguna (User):</strong>
                  <p>Akun pengguna dalam sistem yang memiliki akses untuk menggunakan aplikasi sesuai dengan role dan permission yang diberikan.</p>
                </div>
                <div class="info-item">
                  <strong>Status Aktif/Nonaktif:</strong>
                  <ul>
                    <li><strong>Aktif:</strong> Pengguna dapat login dan menggunakan sistem</li>
                    <li><strong>Nonaktif:</strong> Pengguna tidak dapat login, akun dinonaktifkan sementara</li>
                  </ul>
                </div>
                <div class="info-item">
                  <strong>Role dan Permission:</strong>
                  <p>Setiap pengguna memiliki role yang menentukan permission (izin) mereka dalam sistem. Role dapat di-assign ke perusahaan tertentu untuk data isolation.</p>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </a-tab-pane>

        <!-- Roles Tab -->
        <a-tab-pane key="roles" tab="Peran">
          <a-table
            :columns="[
              { title: 'Nama Peran', dataIndex: 'name', key: 'name' },
              { title: 'Deskripsi', dataIndex: 'description', key: 'description' },
              { title: 'Tingkat', dataIndex: 'level', key: 'level' },
              { title: 'Tipe Peran', dataIndex: 'is_system', key: 'is_system' },
            ]"
            :data-source="roles"
            :loading="rolesLoading"
            :scroll="{ x: 'max-content' }"
            :pagination="{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total: number) => `Total ${total} peran`,
              pageSizeOptions: ['10', '20', '50', '100'],
            }"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'level'">
                <a-tag :color="getLevelColor(record.level)">
                  Level {{ record.level }}
                </a-tag>
              </template>
              <template v-if="column.key === 'is_system'">
                <a-tag :color="record.is_system ? 'blue' : 'default'">
                  {{ record.is_system ? 'Sistem' : 'Kustom' }}
                </a-tag>
              </template>
            </template>
          </a-table>

          <!-- Info Penjelasan Istilah Teknis - Roles -->
          <a-collapse class="info-accordion" :bordered="false">
            <a-collapse-panel key="1" header="📚 Penjelasan Istilah Teknis">
              <div class="info-content">
                <div class="info-item">
                  <strong>Role (Peran):</strong>
                  <p>Role menentukan hak akses dan wewenang user dalam sistem. Setiap role memiliki permissions (izin) tertentu yang menentukan apa yang bisa dilakukan user.</p>
                </div>
                <div class="info-item">
                  <strong>Tipe Role:</strong>
                  <ul>
                    <li><strong>Sistem:</strong> Role bawaan sistem yang tidak bisa dihapus atau diubah (superadmin, admin, manager, staff)</li>
                    <li><strong>Kustom:</strong> Role yang dibuat khusus oleh administrator, bisa diubah atau dihapus</li>
                  </ul>
                </div>
                <div class="info-item">
                  <strong>Tingkat Role:</strong>
                  <p>Angka yang menunjukkan hierarki role. Semakin kecil angkanya, semakin tinggi wewenangnya (0 = superadmin, 1 = admin, dst).</p>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </a-tab-pane>

        <!-- Permissions Tab -->
        <a-tab-pane key="permissions" tab="Izin">
          <a-table
            :columns="[
              { title: 'Nama Izin', dataIndex: 'name', key: 'name' },
              { title: 'Deskripsi', dataIndex: 'description', key: 'description' },
              { title: 'Resource', dataIndex: 'resource', key: 'resource' },
              { title: 'Aksi', dataIndex: 'action', key: 'action' },
              { title: 'Cakupan', dataIndex: 'scope', key: 'scope' },
            ]"
            :data-source="permissions"
            :loading="permissionsLoading"
            :scroll="{ x: 'max-content' }"
            :pagination="{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total: number) => `Total ${total} izin`,
              pageSizeOptions: ['10', '20', '50', '100'],
            }"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'scope'">
                <a-tag :color="getScopeColor(record.scope)">
                  {{ getScopeLabel(record.scope) }}
                </a-tag>
              </template>
            </template>
          </a-table>

          <!-- Info Penjelasan Istilah Teknis - Permissions -->
          <a-collapse class="info-accordion" :bordered="false">
            <a-collapse-panel key="1" header="📚 Penjelasan Istilah Teknis">
              <div class="info-content">
                <div class="info-item">
                  <strong>Permission (Izin):</strong>
                  <p>Permission menentukan aksi spesifik yang bisa dilakukan user terhadap resource tertentu (misalnya: melihat dokumen, membuat user, menghapus company).</p>
                </div>
                <div class="info-item">
                  <strong>Resource (Sumber Data):</strong>
                  <p>Jenis data atau fitur dalam sistem (misalnya: user, company, document, dashboard).</p>
                </div>
                <div class="info-item">
                  <strong>Aksi (Action):</strong>
                  <p>Operasi yang bisa dilakukan terhadap resource (misalnya: view, create, update, delete, manage).</p>
                </div>
                <div class="info-item">
                  <strong>Cakupan (Scope):</strong>
                  <ul>
                    <li><strong>Global:</strong> Akses ke seluruh sistem, hanya untuk superadmin</li>
                    <li><strong>Company:</strong> Akses terbatas pada level perusahaan (holding atau anak perusahaan)</li>
                    <li><strong>Sub Company:</strong> Akses terbatas pada sub-perusahaan dan di bawahnya</li>
                  </ul>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </a-tab-pane>
      </a-tabs>

      <!-- Company Modal -->
      <a-modal
        v-model:open="companyModalVisible"
        :title="editingCompany ? 'Edit Perusahaan' : 'Tambah Perusahaan'"
        @ok="handleSaveCompany"
      >
        <a-form :model="companyForm" layout="vertical">
          <a-form-item label="Nama" required>
            <a-input v-model:value="companyForm.name" placeholder="Nama perusahaan" />
          </a-form-item>
          <a-form-item v-if="!editingCompany" label="Kode" required>
            <a-input v-model:value="companyForm.code" placeholder="Kode perusahaan" />
          </a-form-item>
          <a-form-item label="Deskripsi">
            <a-textarea v-model:value="companyForm.description" placeholder="Deskripsi" />
          </a-form-item>
          <a-form-item v-if="!editingCompany" label="Perusahaan Induk">
            <a-select
              v-model:value="companyForm.parent_id"
              placeholder="Pilih perusahaan induk (opsional)"
              allow-clear
            >
              <a-select-option
                v-for="company in companies"
                :key="company.id"
                :value="company.id"
              >
                {{ company.name }} (Level {{ company.level }})
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-form>
      </a-modal>

      <!-- User Modal -->
      <a-modal
        v-model:open="userModalVisible"
        :title="editingUser ? 'Edit Pengguna' : 'Tambah Pengguna'"
        @ok="handleSaveUser"
      >
        <a-form :model="userForm" layout="vertical">
          <a-form-item label="Nama Pengguna" required>
            <a-input v-model:value="userForm.username" placeholder="Nama pengguna" />
          </a-form-item>
          <a-form-item label="Email" required>
            <a-input v-model:value="userForm.email" type="email" placeholder="Email" />
          </a-form-item>
          <a-form-item v-if="!editingUser" label="Password" required>
            <a-input-password v-model:value="userForm.password" placeholder="Password" />
          </a-form-item>
          <a-form-item label="Perusahaan">
            <a-select
              v-model:value="userForm.company_id"
              placeholder="Select company (optional)"
              allow-clear
            >
              <a-select-option
                v-for="company in companies"
                :key="company.id"
                :value="company.id"
              >
                {{ company.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="Peran">
            <a-select
              v-model:value="userForm.role_id"
              placeholder="Pilih peran (opsional)"
              allow-clear
            >
              <a-select-option
                v-for="role in availableRoles"
                :key="role.id"
                :value="role.id"
              >
                {{ role.name }}
              </a-select-option>
            </a-select>
            <div class="form-help-text">
              <small>Role Superadmin tidak tersedia untuk dibuat dari antarmuka ini</small>
            </div>
          </a-form-item>
        </a-form>
      </a-modal>

      <!-- Reset Password Modal -->
      <a-modal
        v-model:open="resetPasswordModalVisible"
        title="Atur Ulang Password"
        @ok="handleSaveResetPassword"
        ok-text="Atur Ulang Password"
        cancel-text="Batal"
      >
        <a-form :model="resetPasswordForm" layout="vertical">
          <a-form-item label="Username">
            <a-input v-model:value="resetPasswordForm.username" disabled />
          </a-form-item>
          <a-form-item label="Password Baru" required>
            <a-input-password 
              v-model:value="resetPasswordForm.new_password" 
              placeholder="Masukkan password baru (min 8 karakter)"
            />
          </a-form-item>
          <a-form-item label="Konfirmasi Password" required>
            <a-input-password 
              v-model:value="resetPasswordForm.confirm_password" 
              placeholder="Konfirmasi password baru"
            />
          </a-form-item>
          <a-alert
            message="Peringatan"
            description="Password akan diatur ulang dan pengguna harus login dengan password baru."
            type="warning"
            show-icon
            style="margin-top: 16px;"
          />
        </a-form>
      </a-modal>
    </div>
  </div>
</template>

<style scoped>
.user-management-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.user-management-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 16px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #333;
}

@media (min-width: 768px) {
  .user-management-content {
    padding: 24px;
  }
  
  .page-title {
    font-size: 24px;
    margin-bottom: 24px;
  }
}

.management-tabs {
  background: white;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

@media (min-width: 768px) {
  .management-tabs {
    padding: 24px;
  }
}

.table-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}

.search-container {
  margin-bottom: 16px;
}

.search-input {
  width: 100%;
  max-width: 300px;
}


.no-action-text {
  color: #999;
  font-size: 12px;
  font-style: italic;
}

.form-help-text {
  margin-top: 4px;
  color: #999;
  font-size: 12px;
}

.info-accordion {
  margin-top: 24px;
  border-radius: 8px;
  background: white;
}

/* Responsive table adjustments */
:deep(.ant-table-wrapper) {
  overflow-x: auto;
}

:deep(.ant-table) {
  min-width: 600px;
}

@media (max-width: 768px) {
  .table-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .table-header .ant-btn {
    width: 100%;
    margin-bottom: 8px;
  }
  
  :deep(.ant-table-pagination) {
    margin: 16px 0 !important;
  }
  
  :deep(.ant-space) {
    flex-wrap: wrap;
  }
}

.info-content {
  line-height: 1.8;
}

.info-item {
  margin-bottom: 20px;
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-item strong {
  color: #035CAB;
  display: block;
  margin-bottom: 8px;
  font-size: 15px;
}

.info-item ul {
  margin: 8px 0 0 20px;
  padding-left: 0;
}

.info-item ul li {
  margin-bottom: 6px;
  color: #666;
}

.info-item p {
  margin: 8px 0 0 0;
  color: #666;
}

/* Responsive table adjustments */
:deep(.ant-table-wrapper) {
  overflow-x: auto;
}

:deep(.ant-table) {
  min-width: 600px;
}

@media (max-width: 768px) {
  .table-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .table-header .ant-btn {
    width: 100%;
    margin-bottom: 8px;
  }
  
  :deep(.ant-table-pagination) {
    margin: 16px 0 !important;
  }
  
  :deep(.ant-space) {
    flex-wrap: wrap;
  }
  
  .search-input {
    max-width: 100%;
  }
}
</style>

