<template>
  <div class="subsidiaries-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="subsidiaries-content">
      <h1 class="page-title">Manajemen Anak Perusahaan</h1>

      <a-card class="management-card">
        <!-- Companies Tab (Main Content) -->
        <div class="companies-section">
          <div class="table-header">
            <a-button type="primary" @click="handleCreateCompany">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Tambah Perusahaan
            </a-button>
          </div>

          <!-- Search Input -->
          <div style="margin-bottom: 16px;">
            <a-input
              v-model:value="companySearchText"
              placeholder="Cari perusahaan (nama, kode, deskripsi)..."
              allow-clear
              style="width: 300px;"
            >
              <template #prefix>
                <span>🔍</span>
              </template>
            </a-input>
          </div>

          <a-table
            :columns="companyColumns"
            :data-source="filteredCompanies"
            :loading="companiesLoading"
            :pagination="{
              current: companyPagination.current,
              pageSize: companyPagination.pageSize,
              total: companyPagination.total,
              showSizeChanger: true,
              showTotal: (total: number) => `Total ${total} perusahaan`,
              pageSizeOptions: ['10', '20', '50', '100'],
            }"
            @change="(pagination: any) => { 
              companyPagination.current = pagination.current || 1
              companyPagination.pageSize = pagination.pageSize || 10
            }"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'level'">
                <a-tag :color="getLevelColor(record.level)">
                  {{ getLevelLabel(record.level) }}
                </a-tag>
              </template>
              <template v-if="column.key === 'is_active'">
                <a-tag :color="record.is_active ? 'green' : 'red'">
                  {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
                </a-tag>
              </template>
              <template v-if="column.key === 'actions'">
                <a-space>
                  <a-button 
                    type="link" 
                    size="small" 
                    @click="handleEditCompany(record)"
                  >
                    Edit
                  </a-button>
                  <a-popconfirm
                    v-if="!isUserCompany(record.id)"
                    title="Apakah Anda yakin ingin menghapus perusahaan ini?"
                    @confirm="handleDeleteCompany(record.id)"
                  >
                    <a-button type="link" size="small" danger>Hapus</a-button>
                  </a-popconfirm>
                  <span v-else class="no-action-text">Tidak dapat dihapus</span>
                </a-space>
              </template>
            </template>
          </a-table>

          <!-- Info Penjelasan Istilah Teknis -->
          <a-collapse class="info-accordion" :bordered="false">
            <a-collapse-panel key="1" header="📚 Penjelasan Istilah Teknis">
              <div class="info-content">
                <h3>Tingkat Perusahaan</h3>
                <ul>
                  <li><a-tag color="red">Holding (Induk)</a-tag>: Perusahaan utama di level tertinggi (Level 0).</li>
                  <li><a-tag color="blue">Anak Perusahaan</a-tag>: Perusahaan yang dimiliki oleh Holding (Level 1).</li>
                  <li><a-tag color="green">Cucu Perusahaan</a-tag>: Perusahaan yang dimiliki oleh Anak Perusahaan (Level 2).</li>
                  <li><a-tag color="orange">Cicit Perusahaan</a-tag>: Perusahaan yang dimiliki oleh Cucu Perusahaan (Level 3).</li>
                  <li>Level 4+: Tingkat hierarki selanjutnya.</li>
                </ul>
                <h3>Hierarki Perusahaan</h3>
                <p>Sistem ini mendukung struktur perusahaan seperti pohon, di mana setiap perusahaan bisa memiliki anak perusahaan, dan seterusnya. Ini memungkinkan pengelolaan data yang terisolasi sesuai struktur organisasi.</p>
                <h3>Status Aktif/Tidak Aktif</h3>
                <ul>
                  <li><a-tag color="green">Aktif</a-tag>: Company sedang beroperasi dan dapat digunakan.</li>
                  <li><a-tag color="red">Tidak Aktif</a-tag>: Company tidak beroperasi atau dinonaktifkan sementara.</li>
                </ul>
              </div>
            </a-collapse-panel>
          </a-collapse>
        </div>
      </a-card>

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
                {{ company.name }} ({{ getLevelLabel(company.level) }})
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
const companyModalVisible = ref(false)
const companyForm = ref<Partial<Company>>({})
const editingCompany = ref<Company | null>(null)

// Search states
const companySearchText = ref('')

// Pagination
const companyPagination = ref({
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

// Update pagination total when data changes
watch(companies, () => {
  if (!companySearchText.value) {
    companyPagination.value.total = companies.value.length
  }
}, { immediate: true })

// Watch search text changes
watch(companySearchText, () => {
  companyPagination.value.current = 1 // Reset to first page on search
})

const companyColumns = computed(() => [
  { title: 'Nama', dataIndex: 'name', key: 'name', sorter: (a: Company, b: Company) => a.name.localeCompare(b.name) },
  { title: 'Kode', dataIndex: 'code', key: 'code', sorter: (a: Company, b: Company) => a.code.localeCompare(b.code) },
  { title: 'Tingkat', dataIndex: 'level', key: 'level', sorter: (a: Company, b: Company) => a.level - b.level },
  { title: 'Deskripsi', dataIndex: 'description', key: 'description' },
  { title: 'Status', dataIndex: 'is_active', key: 'is_active' },
  { title: 'Aksi', key: 'actions', width: 150 },
])

const loadCompanies = async () => {
  companiesLoading.value = true
  try {
    companies.value = await companyApi.getAll()
  } catch (error: any) {
    message.error('Gagal memuat perusahaan: ' + (error.response?.data?.message || error.message))
  } finally {
    companiesLoading.value = false
  }
}

const handleCreateCompany = () => {
  router.push('/subsidiaries/new')
}

const handleEditCompany = (company: Company) => {
  router.push(`/subsidiaries/${company.id}/edit`)
}

const handleSaveCompany = async () => {
  try {
    if (editingCompany.value) {
      await companyApi.update(editingCompany.value.id, {
        name: companyForm.value.name!,
        description: companyForm.value.description,
      })
      message.success('Perusahaan berhasil diupdate')
    } else {
      await companyApi.create({
        name: companyForm.value.name!,
        code: companyForm.value.code!,
        description: companyForm.value.description,
        parent_id: companyForm.value.parent_id,
      })
      message.success('Perusahaan berhasil dibuat')
    }
    companyModalVisible.value = false
    loadCompanies()
  } catch (error: any) {
    message.error('Gagal menyimpan perusahaan: ' + (error.response?.data?.message || error.message))
  }
}

const isUserCompany = (companyId: string): boolean => {
  return authStore.user?.company_id === companyId
}

const handleDeleteCompany = async (id: string) => {
  try {
    await companyApi.delete(id)
    message.success('Perusahaan berhasil dihapus')
    loadCompanies()
  } catch (error: any) {
    message.error('Gagal menghapus perusahaan: ' + (error.response?.data?.message || error.message))
  }
}

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

onMounted(() => {
  loadCompanies()
})
</script>

<style scoped>
.subsidiaries-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.subsidiaries-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: #333;
}

.management-card {
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.table-header {
  margin-bottom: 16px;
  display: flex;
  justify-content: flex-end;
}

.info-accordion {
  margin-top: 24px;
  border-radius: 8px;
  background: white;
}

.info-content h3 {
  margin-top: 16px;
  margin-bottom: 8px;
  color: #333;
}

.info-content ul {
  margin-left: 20px;
  margin-bottom: 16px;
}

.info-content li {
  margin-bottom: 8px;
  line-height: 1.6;
}

.no-action-text {
  color: #999;
  font-size: 12px;
  font-style: italic;
}
</style>

