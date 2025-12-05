<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import axios from 'axios'
import dayjs, { type Dayjs } from 'dayjs'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import documentsApi, { type DocumentFolder, type DocumentItem } from '../api/documents'

const router = useRouter()
const route = useRoute()

// Check if this is edit mode
const isEditMode = computed(() => !!route.params.id)
const documentId = computed(() => route.params.id as string | undefined)
const loading = ref(false)

type UploadItem = {
  uid: string
  name: string
  status?: string
  size?: number
  originFileObj?: File
  url?: string
  [key: string]: unknown
}

const fileList = ref<UploadItem[]>([])
const folders = ref<DocumentFolder[]>([])
const loadingFolders = ref(false)
const document = ref<DocumentItem | null>(null)
// Fiber default body limit ~4MB, naikkan guard ke 5MB (sesuaikan backend/proxy)
const MAX_FILE_SIZE = 5 * 1024 * 1024 // 5MB

const formState = ref({
  title: '',
  documentId: '',
  docType: '',
  reference: '',
  unit: '',
  uploader: '',
  status: 'active',
  issuedDate: null as Dayjs | null,
  effectiveDate: null as Dayjs | null,
  expiredDate: null as Dayjs | null,
  isActive: '',
  folder_id: undefined as string | undefined,
})

const toDayjs = (value: unknown): Dayjs | null => {
  if (!value) return null
  if (dayjs.isDayjs(value)) return value as Dayjs
  if (value instanceof Date) {
    const parsedDate = dayjs(value)
    return parsedDate.isValid() ? parsedDate : null
  }
  if (typeof value === 'string') {
    const parsedString = dayjs(value)
    return parsedString.isValid() ? parsedString : null
  }
  return null
}

const toIsoString = (value: Dayjs | null) => {
  if (!value) return undefined
  return value.toISOString()
}

const handleUploadChange = ({ fileList: newList }: { fileList: UploadItem[] }) => {
  fileList.value = newList
}

const handleSubmit = async () => {
  if (isEditMode.value) {
    // Update document metadata (dan opsional ganti file)
    if (!documentId.value) {
      message.error('Document ID tidak ditemukan')
      return
    }

    const maybeFile = fileList.value[0]?.originFileObj as File | undefined

    loading.value = true
    try {
      const payload = {
        folder_id: formState.value.folder_id,
        title: formState.value.title,
        status: formState.value.status,
        metadata: {
          doc_type: formState.value.docType,
          reference: formState.value.reference,
          unit: formState.value.unit,
          issued_date: toIsoString(formState.value.issuedDate),
          effective_date: toIsoString(formState.value.effectiveDate),
          expired_date: toIsoString(formState.value.expiredDate),
          is_active: formState.value.isActive,
        },
      }

      if (maybeFile) {
        await documentsApi.updateDocument(documentId.value, { ...payload, file: maybeFile })
      } else {
        await documentsApi.updateDocument(documentId.value, payload)
      }

      message.success('Dokumen berhasil diperbarui')
      router.push(`/documents/${documentId.value}`)
    } catch (error: unknown) {
      const err = error as { message?: string }
      message.error(err.message || 'Gagal update metadata')
    } finally {
      loading.value = false
    }
  } else {
    // Upload new document
    if (!fileList.value.length) {
      message.warning('Silakan pilih file untuk diupload')
      return
    }
    const file = fileList.value[0]?.originFileObj as File | undefined
    if (!file) {
      message.error('File tidak valid')
      return
    }
    if (file.size > MAX_FILE_SIZE) {
      message.error('File terlalu besar, batas maksimal 5MB')
      return
    }

    loading.value = true
    try {
      await documentsApi.uploadDocument({
        file,
        folder_id: formState.value.folder_id,
        title: formState.value.title,
        status: formState.value.status,
        metadata: {
          doc_type: formState.value.docType,
          reference: formState.value.reference,
          unit: formState.value.unit,
          issued_date: toIsoString(formState.value.issuedDate),
          effective_date: toIsoString(formState.value.effectiveDate),
          expired_date: toIsoString(formState.value.expiredDate),
          is_active: formState.value.isActive,
        },
      })
      message.success('File berhasil diupload')
      router.push('/documents')
    } catch (error: unknown) {
      if (axios.isAxiosError(error) && error.response?.status === 413) {
        message.error('File terlalu besar, server menolak upload (413). Silakan pilih file yang lebih kecil atau hubungi admin.')
      } else {
        const err = error as { message?: string }
        message.error(err.message || 'Gagal upload dokumen')
      }
    } finally {
      loading.value = false
    }
  }
}

const handleCancel = () => {
  if (isEditMode.value && documentId.value) {
    router.push('/documents')
  } else {
    router.push('/documents')
  }
}

const loadFolders = async () => {
  loadingFolders.value = true
  try {
    folders.value = await documentsApi.listFolders()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat folder')
  } finally {
    loadingFolders.value = false
  }
}

// Load document data for edit mode
const loadDocument = async () => {
  if (!documentId.value) return

  loading.value = true
  try {
    const doc = await documentsApi.getDocument(documentId.value)
    document.value = doc

    // Populate form with existing data
    formState.value.title = doc.name
    formState.value.folder_id = doc.folder_id || undefined
    formState.value.status = doc.status || 'active'
    formState.value.documentId = doc.id

    // Load metadata
    if (doc.metadata) {
      let meta: Record<string, unknown> = {}
      if (typeof doc.metadata === 'string') {
        try {
          meta = JSON.parse(doc.metadata) as Record<string, unknown>
        } catch {
          meta = {}
        }
      } else {
        meta = doc.metadata as Record<string, unknown>
      }
      formState.value.docType = (meta.doc_type as string) || ''
      formState.value.reference = (meta.reference as string) || ''
      formState.value.unit = (meta.unit as string) || ''
      formState.value.issuedDate = toDayjs(meta.issued_date)
      formState.value.effectiveDate = toDayjs(meta.effective_date)
      formState.value.expiredDate = toDayjs(meta.expired_date)
      formState.value.isActive = (meta.is_active as string) || ''
    }

    // Set file list to show existing file (read-only)
    fileList.value = [{
      uid: doc.id,
      name: doc.file_name,
      status: 'done',
      size: doc.size,
      url: doc.file_path,
    }]
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || 'Gagal memuat dokumen')
    router.push('/documents')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadFolders()
  if (isEditMode.value) {
    await loadDocument()
  }
})
</script>

<template>
  <a-layout class="upload-layout">
    <DashboardHeader />
     <!-- Page Header Section -->
     <div class="page-header-container">
        <div class="page-header">
          <div class="header-left">
            <h1 class="page-title">Upload Dokumen</h1>
          </div>
        </div>
      </div>
    <a-layout-content class="upload-content">

     


      <a-card class="upload-card" :bordered="false">
        <div class="card-header">
          <div>
            <h2>{{ isEditMode ? 'Edit Metadata' : 'Upload file' }}</h2>
            <p>{{ isEditMode ? 'Edit metadata dokumen' : 'Upload dokumen dan lengkapi metadata' }}</p>
          </div>
          <div class="actions">
            <a-button @click="handleCancel">Cancel</a-button>
            <a-button type="primary" :loading="loading" @click="handleSubmit">
              {{ isEditMode ? 'Update' : 'Upload' }}
            </a-button>
          </div>
        </div>

        <div v-if="!isEditMode" class="form-section">
          <h4>Upload file</h4>
          <a-upload-dragger name="file" :file-list="fileList" :before-upload="() => false" :max-count="1"
            @change="handleUploadChange">
            <p class="ant-upload-drag-icon">
              <IconifyIcon icon="mdi:cloud-upload-outline" width="40" />
            </p>
            <p class="ant-upload-text">Browse file to upload</p>
            <p class="ant-upload-hint">Format: PDF, DOCX, XLSX</p>
          </a-upload-dragger>
        </div>

        <div v-else-if="document" class="form-section">
          <h4>File</h4>
          <a-upload-dragger :file-list="fileList" name="file" :before-upload="() => false" :max-count="1"
            @change="handleUploadChange">
            <p class="ant-upload-drag-icon">
              <IconifyIcon icon="mdi:file-document-outline" width="40" />
            </p>
            <p class="ant-upload-text">{{ document.file_name }}</p>
            <p class="ant-upload-hint">Pilih file baru jika ingin mengganti lampiran</p>
          </a-upload-dragger>
        </div>

        <div class="form-section">
          <h4>Meta data</h4>
          <a-form layout="vertical">
            <a-row :gutter="[16, 8]">
              <a-col :xs="24" :md="12">
                <a-form-item label="Folder">
                  <a-select v-model:value="formState.folder_id" placeholder="Pilih folder (opsional)"
                    :loading="loadingFolders" allow-clear>
                    <a-select-option v-for="folder in folders" :key="folder.id" :value="folder.id">
                      {{ folder.name }}
                    </a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Judul Dokumen" required>
                  <a-input v-model:value="formState.title" placeholder="Judul dokumen" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Nomor Dokumen / Referensi">
                  <a-input v-model:value="formState.reference" placeholder="No. referensi" />
                </a-form-item>
              </a-col>
              <a-col v-if="isEditMode" :xs="24" :md="12">
                <a-form-item label="Document ID">
                  <a-input v-model:value="formState.documentId" disabled />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Jenis Dokumen">
                  <a-select v-model:value="formState.docType" placeholder="Pilih jenis dokumen">
                    <a-select-option value="RUPS">RUPS</a-select-option>
                    <a-select-option value="Legal">Legal</a-select-option>
                    <a-select-option value="Keuangan">Keuangan</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Unit / Divisi">
                  <a-input v-model:value="formState.unit" />
                </a-form-item>
              </a-col>
              <a-col v-if="!isEditMode" :xs="24" :md="12">
                <a-form-item label="Uploaded By / PIC">
                  <a-input v-model:value="formState.uploader" disabled />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Status Dokumen">
                  <a-select v-model:value="formState.status">
                    <a-select-option value="active">Active</a-select-option>
                    <a-select-option value="draft">Draft</a-select-option>
                    <a-select-option value="archived">Archived</a-select-option>
                    <a-select-option value="Disetujui">Disetujui</a-select-option>
                    <a-select-option value="Draft">Draft</a-select-option>
                    <a-select-option value="Ditolak">Ditolak</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Tanggal Dokumen (Diterbitkan)">
                  <a-date-picker v-model:value="formState.issuedDate" style="width: 100%;" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Tanggal Berlaku (Effective Date)">
                  <a-date-picker v-model:value="formState.effectiveDate" style="width: 100%;" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Tanggal Berakhir (Expired Date)">
                  <a-date-picker v-model:value="formState.expiredDate" style="width: 100%;" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Is Active">
                  <a-select v-model:value="formState.isActive">
                    <a-select-option value="Aktif">Aktif</a-select-option>
                    <a-select-option value="Tidak Aktif">Tidak Aktif</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
        </div>
      </a-card>
    </a-layout-content>
  </a-layout>
</template>

<style scoped>
.upload-layout {
  min-height: 100vh;
  /* background: #f7f8fb; */
}

.page-header{
  /* background: orange; */
  max-width: 1150px;
  margin: 0 auto;
  width: 100%;
}

.upload-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px 40px;
}

.upload-card {
  border-radius: 14px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.card-header h2 {
  margin: 0;
}

.card-header p {
  margin: 0;
  color: #666;
}

.actions {
  display: flex;
  gap: 8px;
}

.form-section {
  margin-top: 16px;
}

.form-section h4 {
  margin-bottom: 12px;
}
</style>
