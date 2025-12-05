<template>
  <div class="report-form-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="report-form-wrapper">
      <!-- Page Header Section -->
      <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">{{ isEditMode ? 'Edit Report Monthly' : 'Add Report Monthly' }}</h1>
            <p class="page-description">
              Formulir pelaporan kinerja bulanan untuk setiap anak perusahaan.
            </p>
          </div>
        </div>
      </div>

      <!-- Loading Overlay -->
      <a-spin :spinning="loading" tip="Menyimpan data laporan, harap tunggu..." size="large" style="min-height: 400px;">
        <div class="form-content">
          <a-card class="form-card">
            <a-form :model="formData" layout="vertical" @finish="handleSubmit">
              <!-- Informasi Dasar -->
              <div class="form-section">
                <h3 class="section-title">
                  <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                  Informasi dasar
                </h3>
                <a-row :gutter="[16, 16]">
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Periode Laporan" required>
                      <a-date-picker
                        v-model:value="formData.period"
                        picker="month"
                        format="MMMM YYYY"
                        placeholder="Pilih periode laporan"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Tanggal Laporan">
                      <a-date-picker
                        v-model:value="formData.report_date"
                        format="DD MMMM YYYY"
                        placeholder="Pilih tanggal laporan"
                        style="width: 100%"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Nama Anak Perusahaan" required>
                      <a-select
                        v-model:value="formData.company_id"
                        placeholder="Pilih anak perusahaan"
                        show-search
                        :filter-option="filterCompanyOption"
                        :loading="companiesLoading"
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
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Nama Penginput / PIC / Petugas" required>
                      <a-select
                        v-model:value="formData.inputter_id"
                        placeholder="Pilih penginput"
                        show-search
                        :filter-option="filterUserOption"
                        :loading="usersLoading"
                        allow-clear
                      >
                        <a-select-option
                          v-for="user in users"
                          :key="user.id"
                          :value="user.id"
                        >
                          {{ user.username }} ({{ user.email }})
                        </a-select-option>
                      </a-select>
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>

              <!-- Performance This Month -->
              <div class="form-section">
                <h3 class="section-title">
                  <IconifyIcon icon="mdi:chart-line" width="20" style="margin-right: 8px;" />
                  Performance This Month
                </h3>
                <a-row :gutter="[16, 16]">
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Revenue (pendapatan)" required>
                      <a-input-number
                        v-model:value="formData.revenue"
                        placeholder="Masukkan revenue"
                        style="width: 100%"
                        :min="0"
                        :formatter="(value: string | number | undefined) => `${value || ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                        :parser="(value: string | undefined) => (value || '').replace(/\$\s?|(,*)/g, '')"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Opex (Operational Expenditure)" required>
                      <a-input-number
                        v-model:value="formData.opex"
                        placeholder="Masukkan opex"
                        style="width: 100%"
                        :min="0"
                        :formatter="(value: string | number | undefined) => `${value || ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                        :parser="(value: string | undefined) => (value || '').replace(/\$\s?|(,*)/g, '')"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="NPAT (Net Profit After Tax)" required>
                      <a-input-number
                        v-model:value="formData.npat"
                        placeholder="Masukkan NPAT"
                        style="width: 100%"
                        :formatter="(value: string | number | undefined) => `${value || ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                        :parser="(value: string | undefined) => (value || '').replace(/\$\s?|(,*)/g, '')"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Dividend (Dividen)" required>
                      <a-input-number
                        v-model:value="formData.dividend"
                        placeholder="Masukkan dividend"
                        style="width: 100%"
                        :min="0"
                        :formatter="(value: string | number | undefined) => `${value || ''}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                        :parser="(value: string | undefined) => (value || '').replace(/\$\s?|(,*)/g, '')"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Financial Ratio">
                      <a-input-number
                        v-model:value="formData.financial_ratio"
                        placeholder="Masukkan financial ratio"
                        style="width: 100%"
                        :min="0"
                        :precision="2"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :md="12">
                    <a-form-item label="Attachment (opsional)">
                      <a-upload
                        :file-list="attachmentFileList"
                        :before-upload="handleAttachmentUpload"
                        @remove="handleAttachmentRemove"
                        :max-count="1"
                      >
                        <a-button>
                          <IconifyIcon icon="mdi:paperclip" width="16" style="margin-right: 8px;" />
                          Upload File
                        </a-button>
                      </a-upload>
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24">
                    <a-form-item label="Catatan / Remark (opsional)">
                      <a-textarea
                        v-model:value="formData.remark"
                        placeholder="Masukkan catatan atau remark"
                        :rows="4"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>

              <!-- Action Buttons -->
              <div class="form-actions">
                <a-button type="default" size="large" @click="handleCancel" :disabled="loading">
                  Cancel
                </a-button>
                <a-button type="primary" size="large" html-type="submit" :loading="loading" class="save-button">
                  Save
                </a-button>
              </div>
            </a-form>
          </a-card>
        </div>
      </a-spin>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import 'dayjs/locale/id'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import { companyApi, userApi, type Company, type User } from '../api/userManagement'
import apiClient from '../api/client'
import type { UploadFile } from 'ant-design-vue'
import reportsApi from '../api/reports'

dayjs.locale('id')

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const companiesLoading = ref(false)
const usersLoading = ref(false)

const isEditMode = computed(() => !!route.params.id)

// Form data
const formData = ref({
  period: null as Dayjs | null,
  report_date: null as Dayjs | null,
  company_id: undefined as string | undefined,
  inputter_id: undefined as string | undefined,
  revenue: undefined as number | undefined,
  opex: undefined as number | undefined,
  npat: undefined as number | undefined,
  dividend: undefined as number | undefined,
  financial_ratio: undefined as number | undefined,
  attachment: undefined as string | undefined,
  remark: undefined as string | undefined,
})

// Companies and Users
const companies = ref<Company[]>([])
const users = ref<User[]>([])

// File upload
const attachmentFileList = ref<UploadFile[]>([])

// Load data
const loadCompanies = async () => {
  companiesLoading.value = true
  try {
    companies.value = await companyApi.getAll()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat daftar perusahaan: ' + (axiosError.response?.data?.message || axiosError.message))
  } finally {
    companiesLoading.value = false
  }
}

const loadUsers = async () => {
  usersLoading.value = true
  try {
    users.value = await userApi.getAll()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat daftar pengguna: ' + (axiosError.response?.data?.message || axiosError.message))
  } finally {
    usersLoading.value = false
  }
}

// Filter options
interface FilterOption {
  value: string
}

const filterCompanyOption = (input: string, option: FilterOption) => {
  const company = companies.value.find(c => c.id === option.value)
  if (!company) return false
  return company.name.toLowerCase().includes(input.toLowerCase())
}

const filterUserOption = (input: string, option: FilterOption) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  const searchText = input.toLowerCase()
  return user.username.toLowerCase().includes(searchText) || 
         user.email.toLowerCase().includes(searchText)
}

// File upload handlers
const handleAttachmentUpload = (file: File): boolean => {
  // TODO: Implement file upload to server
  // For now, just store filename
  formData.value.attachment = file.name
  attachmentFileList.value = [{
    uid: '-1',
    name: file.name,
    status: 'done',
  }]
  message.info('File attachment akan diupload saat save')
  return false // Prevent default upload
}

const handleAttachmentRemove = (): void => {
  formData.value.attachment = undefined
  attachmentFileList.value = []
}

// Form handlers
const handleCancel = () => {
  router.push('/reports')
}

const handleSubmit = async () => {
  // Prevent multiple submissions
  if (loading.value) {
    return
  }

  // Validation
  if (!formData.value.period) {
    message.error('Periode laporan harus diisi')
    return
  }

  if (!formData.value.company_id) {
    message.error('Nama anak perusahaan harus dipilih')
    return
  }

  if (!formData.value.inputter_id) {
    message.error('Nama penginput harus dipilih')
    return
  }

  if (formData.value.revenue === undefined || formData.value.revenue === null) {
    message.error('Revenue harus diisi')
    return
  }

  if (formData.value.opex === undefined || formData.value.opex === null) {
    message.error('Opex harus diisi')
    return
  }

  if (formData.value.npat === undefined || formData.value.npat === null) {
    message.error('NPAT harus diisi')
    return
  }

  if (formData.value.dividend === undefined || formData.value.dividend === null) {
    message.error('Dividend harus diisi')
    return
  }

  // Set loading
  loading.value = true

  try {
    // Prepare data untuk API
    const submitData = {
      period: formData.value.period.format('YYYY-MM'),
      report_date: formData.value.report_date ? formData.value.report_date.format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
      company_id: formData.value.company_id,
      inputter_id: formData.value.inputter_id,
      revenue: formData.value.revenue,
      opex: formData.value.opex,
      npat: formData.value.npat,
      dividend: formData.value.dividend,
      financial_ratio: formData.value.financial_ratio || null,
      attachment: formData.value.attachment || null,
      remark: formData.value.remark || null,
    }

    if (isEditMode.value) {
      // Edit mode
      await apiClient.put(`/reports/${route.params.id}`, submitData)
      message.success('Laporan berhasil diperbarui')
    } else {
      // Create mode
      await apiClient.post('/reports', submitData)
      message.success('Laporan berhasil disimpan')
    }

    // Success - redirect setelah delay kecil
    await new Promise(resolve => setTimeout(resolve, 500))
    router.push('/reports')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal menyimpan laporan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    loading.value = false
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// Load data on mount
onMounted(async () => {
  await Promise.all([
    loadCompanies(),
    loadUsers(),
  ])

  // Set default values for add mode
  if (!isEditMode.value) {
    // Set default inputter to current user if available
    if (authStore.user?.id) {
      formData.value.inputter_id = authStore.user.id
    }
    // Set default report date to today
    formData.value.report_date = dayjs()
  }

  // Load report data if edit mode
  if (isEditMode.value) {
    await loadReportData(route.params.id as string)
  }
})

// Load report data for editing
const loadReportData = async (reportId: string) => {
  loading.value = true
  try {
    // Ensure users are loaded first so select can display username correctly
    if (users.value.length === 0) {
      await loadUsers()
    }
    
    const report = await reportsApi.getById(reportId)
    
    // Populate form data
    if (report.period) {
      // Parse period (YYYY-MM) to dayjs
      const [year, month] = report.period.split('-')
      formData.value.period = dayjs(`${year}-${month}-01`)
    }
    // Parse report_date - check if report_date exists in response, otherwise use created_at or today
    if (report.report_date) {
      formData.value.report_date = dayjs(report.report_date)
    } else if (report.created_at) {
      formData.value.report_date = dayjs(report.created_at)
    } else {
      // Fallback to today if report_date is not available
      formData.value.report_date = dayjs()
    }
    formData.value.company_id = report.company_id
    formData.value.inputter_id = report.inputter_id
    formData.value.revenue = report.revenue
    formData.value.opex = report.opex
    formData.value.npat = report.npat
    formData.value.dividend = report.dividend
    formData.value.financial_ratio = report.financial_ratio
    formData.value.attachment = report.attachment || undefined
    formData.value.remark = report.remark || undefined
    
    // Set attachment file list if exists
    if (report.attachment) {
      attachmentFileList.value = [{
        uid: '-1',
        name: report.attachment.split('/').pop() || 'attachment',
        status: 'done',
      }]
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat data report: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    // Redirect back to reports list on error
    router.push('/reports')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.report-form-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.report-form-wrapper {
  margin: 0 auto;
}

.form-content {
  padding: 24px;
}

.form-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.form-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.save-button {
  min-width: 120px;
}

/* Responsive */
@media (max-width: 768px) {
  .form-content {
    padding: 16px;
  }

  .form-actions {
    flex-direction: column-reverse;
  }

  .form-actions .ant-btn {
    width: 100%;
  }
}
</style>

