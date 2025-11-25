<template>
  <div class="subsidiary-form-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="form-content">
      <a-card class="form-card">
        <!-- Progress Steps -->
        <a-steps :current="currentStep" class="form-steps">
          <a-step title="1 Identitas Perusahaan" />
          <a-step title="2 Struktur Kepemilikan" />
          <a-step title="3 Bidang Usaha" />
          <a-step title="4 Pengurus/Dewan Direksi" />
        </a-steps>

        <!-- Step 1: Identitas Perusahaan -->
        <div v-if="currentStep === 0" class="step-content">
          <h2 class="step-title">Company Information</h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <!-- Informasi Dasar -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Informasi Dasar
            </h3>
            <a-row :gutter="[12, 1]">

              <a-col :xs="24" :md="6">
                <a-form-item label="Nama Lengkap" required>
                  <a-input v-model:value="formData.name" placeholder="Nama lengkap perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Nama Singkat">
                  <a-input v-model:value="formData.short_name" placeholder="Nama singkat perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="NPWP">
                  <a-input v-model:value="formData.npwp" placeholder="Nomor Pokok Wajib Pajak" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="NIB">
                  <a-input v-model:value="formData.nib" placeholder="Nomor Induk Berusaha" />
                </a-form-item>
              </a-col>
              <a-col :xs="16">
                <a-form-item label="Deskripsi">
                  <a-textarea v-model:value="formData.description" :rows="1" placeholder="Deskripsi perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="8">
                <a-form-item label="Status">
                  <a-select v-model:value="formData.status" placeholder="Pilih status">
                    <a-select-option value="Aktif">Aktif</a-select-option>
                    <a-select-option value="Tidak Aktif">Tidak Aktif</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Logo">
                  <a-upload
                    :file-list="logoFileList"
                    :before-upload="handleLogoUpload"
                    :remove="handleLogoRemove"
                    accept="image/png,image/jpeg,image/jpg"
                    :max-count="1"
                    list-type="picture-card"
                  >
                    <div v-if="logoFileList.length < 1">
                      <IconifyIcon icon="mdi:plus" width="24" />
                      <div style="margin-top: 8px">Upload</div>
                    </div>
                  </a-upload>
                  <div v-if="logoFileList.length > 0" style="margin-top: 8px; color: #666; font-size: 12px">
                    Format: PNG, JPG, JPEG | Maks: 5MB
                  </div>
                </a-form-item>
              </a-col>
              
            </a-row>
          </div>

          <!-- Informasi Kontak -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:phone" width="20" style="margin-right: 8px;" />
              Informasi Kontak
            </h3>
            <a-row :gutter="[12, 1]">
              <a-col :xs="24" :md="6">
                <a-form-item label="Telp">
                  <a-input v-model:value="formData.phone" placeholder="Nomor telepon" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Fax">
                  <a-input v-model:value="formData.fax" placeholder="Nomor fax" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Email">
                  <a-input v-model:value="formData.email" type="email" placeholder="Email perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Website">
                  <a-input v-model:value="formData.website" placeholder="Website perusahaan" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>

          <!-- Alamat Perusahaan -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:map-marker" width="20" style="margin-right: 8px;" />
              Alamat Perusahaan
            </h3>
            <a-row :gutter="[12, 1]">
              <a-col :xs="12">
                <a-form-item label="Alamat Perusahaan">
                  <a-textarea v-model:value="formData.address" :rows="3" placeholder="Alamat perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="12">
                <a-form-item label="Alamat Operasional">
                  <a-textarea v-model:value="formData.operational_address" :rows="3" placeholder="Alamat operasional" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>
          </a-form>
        </div>

        <!-- Step 2: Struktur Kepemilikan -->
        <div v-if="currentStep === 1" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:account-group" width="24" style="margin-right: 8px;" />
            Struktur Kepemilikan
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
            <!-- Informasi Dasar -->
            <div class="form-section">
              <h3 class="section-title">
                <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                Informasi Dasar
              </h3>
              <a-row :gutter="[12, 1]">
                <a-col v-if="!route.params.id" :xs="24" :md="12">
                  <a-form-item label="Kode Perusahaan" required>
                    <a-input v-model:value="formData.code" placeholder="Kode perusahaan (unik)" />
                  </a-form-item>
                </a-col>
                <a-col v-if="!route.params.id" :xs="24" :md="12">
                  <a-form-item label="Perusahaan Induk">
                    <a-select
                      v-model:value="formData.parent_id"
                      placeholder="Pilih perusahaan induk (opsional)"
                      allow-clear
                    >
                      <a-select-option
                        v-for="company in availableCompanies"
                        :key="company.id"
                        :value="company.id"
                      >
                        {{ company.name }} ({{ getLevelLabel(company.level) }})
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
                <a-col :xs="24" :md="12">
                  <a-form-item label="Perusahaan Induk Utama">
                    <a-select
                      v-model:value="formData.main_parent_company"
                      placeholder="Pilih perusahaan induk utama (opsional)"
                      allow-clear
                    >
                      <a-select-option
                        v-for="company in availableCompanies"
                        :key="company.id"
                        :value="company.id"
                      >
                        {{ company.name }} ({{ getLevelLabel(company.level) }})
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>
          </div>

          <!-- Pemegang Saham -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Pemegang Saham
            </h3>
            <a-table
              :columns="shareholderColumns"
              :data-source="formData.shareholders"
              :pagination="false"
              row-key="id"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'type'">
                  <a-select
                    v-model:value="record.type"
                    style="width: 100%"
                    placeholder="Jenis pemegang saham"
                  >
                    <a-select-option value="Badan Hukum (Induk)">Badan Hukum (Induk)</a-select-option>
                    <a-select-option value="Badan Hukum">Badan Hukum</a-select-option>
                    <a-select-option value="Individu">Individu</a-select-option>
                  </a-select>
                </template>
                <template v-if="column.key === 'name'">
                  <a-input v-model:value="record.name" placeholder="Nama pemegang saham" />
                </template>
                <template v-if="column.key === 'identity_number'">
                  <a-input v-model:value="record.identity_number" placeholder="KTP/NPWP" />
                </template>
                <template v-if="column.key === 'ownership_percent'">
                  <a-input-number
                    v-model:value="record.ownership_percent"
                    :min="0"
                    :max="100"
                    :precision="2"
                    style="width: 100%"
                    placeholder="%"
                  />
                </template>
                <template v-if="column.key === 'share_count'">
                  <a-input-number
                    v-model:value="record.share_count"
                    :min="0"
                    style="width: 100%"
                    placeholder="Jumlah saham"
                  />
                </template>
                <template v-if="column.key === 'actions'">
                  <a-button type="link" danger size="small" @click="removeShareholder(index)">
                    <IconifyIcon icon="mdi:delete" width="16" />
                  </a-button>
                </template>
              </template>
            </a-table>
            <a-button type="dashed" style="width: 100%; margin-top: 16px;" @click="addShareholder">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Tambah Pemegang Saham
            </a-button>
          </div>
          </a-form>
        </div>

        <!-- Step 3: Bidang Usaha -->
        <div v-if="currentStep === 2" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:briefcase" width="24" style="margin-right: 8px;" />
            Bidang Usaha
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <!-- Utama -->
          <div class="form-section">
            <h3 class="section-title">Utama</h3>
            <a-row :gutter="[12, 1]">
              <a-col :xs="24" :md="12">
                <a-form-item label="Sektor Industri">
                  <a-input v-model:value="formData.main_business.industry_sector" placeholder="Sektor industri" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="KBLI">
                  <a-input v-model:value="formData.main_business.kbli" placeholder="Klasifikasi Baku Lapangan Usaha Indonesia" />
                </a-form-item>
              </a-col>
              <a-col :xs="24">
                <a-form-item label="Uraian Kegiatan Usaha Utama">
                  <a-textarea v-model:value="formData.main_business.main_business_activity" :rows="4" placeholder="Uraian kegiatan usaha utama" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>

          <!-- Lain-lain -->
          <div class="form-section">
            <h3 class="section-title">Lain-lain</h3>
            <a-row :gutter="[12, 1]">
              <a-col :xs="24">
                <a-form-item label="Kegiatan Usaha Tambahan">
                  <a-textarea v-model:value="formData.main_business.additional_activities" :rows="3" placeholder="Kegiatan usaha tambahan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Tanggal Mulai Beroperasi">
                  <a-date-picker
                    v-model:value="formData.main_business.start_operation_date"
                    style="width: 100%"
                    format="DD MMMM YYYY"
                    placeholder="Pilih tanggal"
                  />
                </a-form-item>
              </a-col>
            </a-row>
          </div>
          </a-form>
        </div>

        <!-- Step 4: Pengurus/Dewan Direksi -->
        <div v-if="currentStep === 3" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:account-tie" width="24" style="margin-right: 8px;" />
            Pengurus/Dewan Direksi
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Data Individu
            </h3>
            <a-table
              :columns="directorColumns"
              :data-source="formData.directors"
              :pagination="false"
              row-key="id"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'position'">
                  <a-select
                    v-model:value="record.position"
                    style="width: 100%"
                    placeholder="Jabatan"
                  >
                    <a-select-option value="Direktur Utama">Direktur Utama</a-select-option>
                    <a-select-option value="Direktur Keuangan">Direktur Keuangan</a-select-option>
                    <a-select-option value="Direktur Operasional">Direktur Operasional</a-select-option>
                    <a-select-option value="Komisaris Utama">Komisaris Utama</a-select-option>
                    <a-select-option value="Komisaris">Komisaris</a-select-option>
                  </a-select>
                </template>
                <template v-if="column.key === 'full_name'">
                  <a-input v-model:value="record.full_name" placeholder="Nama lengkap" />
                </template>
                <template v-if="column.key === 'ktp'">
                  <a-input v-model:value="record.ktp" placeholder="Nomor KTP" />
                </template>
                <template v-if="column.key === 'npwp'">
                  <a-input v-model:value="record.npwp" placeholder="Nomor NPWP" />
                </template>
                <template v-if="column.key === 'start_date'">
                  <a-date-picker
                    v-model:value="record.start_date"
                    style="width: 100%"
                    format="DD MMMM YYYY"
                    placeholder="Tanggal awal jabatan"
                  />
                </template>
                <template v-if="column.key === 'domicile_address'">
                  <a-input v-model:value="record.domicile_address" placeholder="Alamat domisili" />
                </template>
                <template v-if="column.key === 'actions'">
                  <a-button type="link" danger size="small" @click="removeDirector(index)">
                    <IconifyIcon icon="mdi:delete" width="16" />
                  </a-button>
                </template>
              </template>
            </a-table>
            <a-button type="dashed" style="width: 100%; margin-top: 16px;" @click="addDirector">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Tambah +
            </a-button>
          </div>
          </a-form>
        </div>

        <!-- Navigation Buttons -->
        <div class="form-actions">
          <a-space>
            <a-button @click="handleCancel">Cancel</a-button>
            <a-button v-if="currentStep > 0" @click="prevStep">
              <IconifyIcon icon="mdi:arrow-left" width="16" style="margin-right: 4px;" />
              Previous
            </a-button>
            <a-button v-if="currentStep < 3" type="primary" @click="nextStep">
              Next
              <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
            </a-button>
            <a-button v-if="currentStep === 3" type="primary" @click="handleSubmit">
              Finish
              <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
            </a-button>
          </a-space>
        </div>
      </a-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, uploadApi, type Company } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import apiClient from '../api/client'
import dayjs from 'dayjs'
import 'dayjs/locale/id'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const currentStep = ref(0)
const loading = ref(false)
const availableCompanies = ref<Company[]>([])
const logoFileList = ref<any[]>([])
const uploadingLogo = ref(false)

const formData = ref({
  // Step 1: Identitas Perusahaan
  name: '',
  short_name: '',
  description: '',
  npwp: '',
  nib: '',
  status: 'Aktif',
  logo: '',
  phone: '',
  fax: '',
  email: '',
  website: '',
  address: '',
  operational_address: '',
  code: '',
  parent_id: undefined as string | undefined,
  main_parent_company: undefined as string | undefined,
  
  // Step 2: Struktur Kepemilikan
  shareholders: [] as Array<{
    id?: string
    type: string
    name: string
    identity_number: string
    ownership_percent: number
    share_count: number
    is_main_parent: boolean
  }>,
  
  // Step 3: Bidang Usaha
  main_business: {
    industry_sector: '',
    kbli: '',
    main_business_activity: '',
    additional_activities: '',
    start_operation_date: null as any,
  },
  
  // Step 4: Pengurus/Dewan Direksi
  directors: [] as Array<{
    id?: string
    position: string
    full_name: string
    ktp: string
    npwp: string
    start_date: any
    domicile_address: string
  }>,
})

const shareholderColumns = [
  { title: 'Jenis Pemegang Saham', key: 'type', width: 200 },
  { title: 'Nama Pemegang Saham', key: 'name', width: 200 },
  { title: 'Nomor Identitas (KTP/NPWP)', key: 'identity_number', width: 180 },
  { title: 'Persentase Kepemilikan', key: 'ownership_percent', width: 150 },
  { title: 'Jumlah Saham', key: 'share_count', width: 150 },
  { title: 'Aksi', key: 'actions', width: 80 },
]

const directorColumns = [
  { title: 'Jabatan', key: 'position', width: 150 },
  { title: 'Nama Lengkap', key: 'full_name', width: 200 },
  { title: 'Nomor Identitas (KTP)', key: 'ktp', width: 150 },
  { title: 'NPWP', key: 'npwp', width: 150 },
  { title: 'Tanggal Awal Jabatan', key: 'start_date', width: 150 },
  { title: 'Alamat Domisili', key: 'domicile_address', width: 200 },
  { title: 'Aksi', key: 'actions', width: 80 },
]

const addShareholder = () => {
  formData.value.shareholders.push({
    type: '',
    name: '',
    identity_number: '',
    ownership_percent: 0,
    share_count: 0,
    is_main_parent: false,
  })
}

const removeShareholder = (index: number) => {
  formData.value.shareholders.splice(index, 1)
}

const addDirector = () => {
  formData.value.directors.push({
    position: '',
    full_name: '',
    ktp: '',
    npwp: '',
    start_date: null,
    domicile_address: '',
  })
}

const removeDirector = (index: number) => {
  formData.value.directors.splice(index, 1)
}

const handleLogoUpload = async (file: File): Promise<boolean> => {
  // Validasi ukuran (max 5MB)
  const maxSize = 5 * 1024 * 1024
  if (file.size > maxSize) {
    message.error('Ukuran file melebihi 5MB')
    return false
  }

  // Validasi format
  const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg']
  if (!allowedTypes.includes(file.type)) {
    message.error('Format file tidak diizinkan. Hanya PNG, JPG, dan JPEG yang diperbolehkan')
    return false
  }

  uploadingLogo.value = true
  try {
    const response = await uploadApi.uploadLogo(file)
    formData.value.logo = response.url
    // Get base URL tanpa /api/v1 untuk static files
    const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const baseURL = apiURL.replace(/\/api\/v1$/, '') // Hapus /api/v1 jika ada
    logoFileList.value = [{
      uid: '-1',
      name: file.name,
      status: 'done',
      url: `${baseURL}${response.url}`,
    }]
    message.success('Logo berhasil diupload')
    return false // Prevent default upload
  } catch (error: any) {
    message.error(error.response?.data?.message || 'Gagal upload logo')
    return false
  } finally {
    uploadingLogo.value = false
  }
}

const handleLogoRemove = (): boolean => {
  formData.value.logo = ''
  logoFileList.value = []
  return true
}

const nextStep = () => {
  if (currentStep.value < 3) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const handleCancel = () => {
  router.push('/subsidiaries')
}

const handleSubmit = async () => {
  loading.value = true
  try {
    // Prepare data untuk API - menggunakan snake_case sesuai JSON tag
    const submitData = {
      name: formData.value.name,
      short_name: formData.value.short_name,
      description: formData.value.description,
      code: formData.value.code || `COMP-${Date.now()}`,
      npwp: formData.value.npwp,
      nib: formData.value.nib,
      status: formData.value.status,
      logo: formData.value.logo,
      phone: formData.value.phone,
      fax: formData.value.fax,
      email: formData.value.email,
      website: formData.value.website,
      address: formData.value.address,
      operational_address: formData.value.operational_address,
      parent_id: formData.value.parent_id || null,
      main_parent_company: formData.value.main_parent_company || null,
      shareholders: formData.value.shareholders.map(sh => ({
        type: sh.type,
        name: sh.name,
        identity_number: sh.identity_number,
        ownership_percent: sh.ownership_percent || 0,
        share_count: sh.share_count || 0,
        is_main_parent: false,
      })),
      main_business: (formData.value.main_business.industry_sector || formData.value.main_business.kbli) ? {
        industry_sector: formData.value.main_business.industry_sector,
        kbli: formData.value.main_business.kbli,
        main_business_activity: formData.value.main_business.main_business_activity,
        additional_activities: formData.value.main_business.additional_activities,
        start_operation_date: formData.value.main_business.start_operation_date?.format('YYYY-MM-DD') || null,
      } : null,
      directors: formData.value.directors.map(d => ({
        position: d.position,
        full_name: d.full_name,
        ktp: d.ktp,
        npwp: d.npwp,
        start_date: d.start_date?.format('YYYY-MM-DD') || null,
        domicile_address: d.domicile_address,
      })),
    }

    if (route.params.id) {
      // Edit mode - use full update endpoint
      await apiClient.put(`/companies/${route.params.id}/full`, submitData)
      message.success('Perusahaan berhasil diupdate')
    } else {
      // Create mode - use full create endpoint
      await apiClient.post('/companies/full', submitData)
      message.success('Perusahaan berhasil dibuat')
    }
    
    router.push('/subsidiaries')
  } catch (error: any) {
    message.error('Gagal menyimpan: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
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

const loadAvailableCompanies = async () => {
  try {
    availableCompanies.value = await companyApi.getAll()
  } catch (error) {
    console.error('Failed to load companies:', error)
  }
}

const loadCompanyData = async () => {
  if (route.params.id) {
    loading.value = true
    try {
      // Load available companies dulu untuk mencari main_parent_company
      if (availableCompanies.value.length === 0) {
        await loadAvailableCompanies()
      }
      const company = await companyApi.getById(route.params.id as string)
      // Populate form data
      formData.value.name = company.name || ''
      formData.value.short_name = company.short_name || ''
      formData.value.description = company.description || ''
      formData.value.npwp = company.npwp || ''
      formData.value.nib = company.nib || ''
      formData.value.status = company.status || 'Aktif'
      formData.value.logo = company.logo || ''
      // Set logo file list jika ada logo
      if (company.logo) {
        let logoUrl: string
        if (company.logo.startsWith('http')) {
          logoUrl = company.logo
        } else {
          // Get base URL tanpa /api/v1 untuk static files
          const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
          const baseURL = apiURL.replace(/\/api\/v1$/, '') // Hapus /api/v1 jika ada
          logoUrl = `${baseURL}${company.logo}`
        }
        logoFileList.value = [{
          uid: '-1',
          name: company.logo.split('/').pop() || 'logo',
          status: 'done',
          url: logoUrl,
        }]
      } else {
        logoFileList.value = []
      }
      formData.value.phone = company.phone || ''
      formData.value.fax = company.fax || ''
      formData.value.email = company.email || ''
      formData.value.website = company.website || ''
      formData.value.address = company.address || ''
      formData.value.operational_address = company.operational_address || ''
      formData.value.code = company.code || ''
      formData.value.parent_id = company.parent_id
      // Ambil main_parent_company langsung dari response
      formData.value.main_parent_company = company.main_parent_company || undefined
      formData.value.shareholders = (company.shareholders || []).map((sh: any) => ({
        ...sh,
        ownership_percent: sh.ownership_percent || 0,
        share_count: sh.share_count || 0,
      }))
      // Transform business_fields array to main_business (ambil yang is_main = true atau yang pertama)
      if (company.business_fields && company.business_fields.length > 0) {
        const mainBusiness = company.business_fields.find((bf: any) => bf.is_main) || company.business_fields[0]
        formData.value.main_business.industry_sector = mainBusiness.industry_sector || ''
        formData.value.main_business.kbli = mainBusiness.kbli || ''
        formData.value.main_business.main_business_activity = mainBusiness.main_business_activity || ''
        formData.value.main_business.additional_activities = mainBusiness.additional_activities || ''
        formData.value.main_business.start_operation_date = mainBusiness.start_operation_date ? dayjs(mainBusiness.start_operation_date) : null
      } else if (company.main_business) {
        // Fallback untuk kompatibilitas jika ada main_business langsung
        formData.value.main_business.industry_sector = company.main_business.industry_sector || ''
        formData.value.main_business.kbli = company.main_business.kbli || ''
        formData.value.main_business.main_business_activity = company.main_business.main_business_activity || ''
        formData.value.main_business.additional_activities = company.main_business.additional_activities || ''
        formData.value.main_business.start_operation_date = company.main_business.start_operation_date ? dayjs(company.main_business.start_operation_date) : null
      }
      formData.value.directors = (company.directors || []).map((d: any) => ({
        ...d,
        start_date: d.start_date ? dayjs(d.start_date) : null,
      }))
    } catch (error: any) {
      message.error('Gagal memuat data perusahaan')
    } finally {
      loading.value = false
    }
  }
}

onMounted(async () => {
  await loadAvailableCompanies()
  await loadCompanyData()
})
</script>

<style scoped>
.subsidiary-form-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.form-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.form-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-steps {
  margin-bottom: 32px;
}

.step-content {
  min-height: 400px;
}

.step-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.form-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #035CAB;
}

.form-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e8e8e8;
  display: flex;
  justify-content: flex-end;
}

/* Label di atas input */
.form-card :deep(.ant-form-item-label) {
  display: block;
  text-align: left;
  margin-bottom: 8px;
  padding: 0;
}

.form-card :deep(.ant-form-item-label > label) {
  height: auto;
  line-height: 1.5;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
}

.form-card :deep(.ant-form-item-label > label.ant-form-item-required:not(.ant-form-item-required-mark-optional)::before) {
  margin-right: 4px;
}

.form-card :deep(.ant-form-item-control) {
  flex: 1;
}

@media (max-width: 768px) {
  .form-content {
    padding: 16px;
  }
  
  .form-steps {
    margin-bottom: 24px;
  }
  
  .step-content {
    min-height: 300px;
  }
}
</style>

