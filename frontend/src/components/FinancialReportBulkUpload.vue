<template>
  <div class="financial-report-bulk-upload">
    <a-card :bordered="false">
      <template #title>
        <div style="display: flex; align-items: center; justify-content: space-between;">
          <div style="display: flex; align-items: center; gap: 8px;">
            <IconifyIcon icon="mdi:file-excel" width="24" />
            <span>Upload Excel - Bulk Input Laporan</span>
          </div>
          <a-button type="primary" @click="handleDownloadTemplate" :loading="templateLoading">
            <IconifyIcon icon="mdi:download" width="16" style="margin-right: 8px;" />
            Download Template
          </a-button>
        </div>
      </template>

      <div class="upload-content">
        <a-alert
          message="Panduan Upload Excel"
          description="Download template Excel, isi data laporan keuangan untuk semua perusahaan yang ingin diinput dalam satu periode. Setelah selesai, upload file Excel tersebut ke sini. Sistem akan memvalidasi dan memproses semua data sekaligus."
          type="info"
          show-icon
          style="margin-bottom: 24px;"
        />


        <!-- File Upload Area -->
        <a-upload-dragger
          v-model:fileList="fileList"
          :before-upload="handleBeforeUpload"
          accept=".xlsx,.xls"
          :multiple="false"
          :show-upload-list="true"
          @remove="handleRemoveFile"
        >
          <p class="ant-upload-drag-icon">
            <IconifyIcon icon="mdi:file-excel-outline" width="64" />
          </p>
          <p class="ant-upload-text">Klik atau drag file Excel ke sini untuk upload</p>
          <p class="ant-upload-hint">Hanya file .xlsx atau .xls yang diperbolehkan</p>
        </a-upload-dragger>

        <!-- Validation Result -->
        <div v-if="validationResult" class="validation-result" style="margin-top: 24px;">
          <a-card size="small">
            <template #title>
              <div style="display: flex; align-items: center; gap: 8px;">
                <a-tag :color="validationResult.valid ? 'success' : 'error'">
                  {{ validationResult.valid ? 'Valid' : 'Ada Error' }}
                </a-tag>
                <span>Hasil Validasi</span>
              </div>
            </template>

            <div v-if="validationResult.valid" style="color: #52c41a; margin-bottom: 16px;">
              <IconifyIcon icon="mdi:check-circle" width="20" style="margin-right: 8px;" />
              File valid! Siap untuk diupload. Total {{ validationResult.data?.length || 0 }} baris data.
            </div>

            <div v-else style="color: #ff4d4f; margin-bottom: 16px;">
              <IconifyIcon icon="mdi:alert-circle" width="20" style="margin-right: 8px;" />
              Ditemukan {{ validationResult.errors?.length || 0 }} error. Harap perbaiki sebelum upload.
            </div>

            <!-- Error Details -->
            <div v-if="validationResult.errors && validationResult.errors.length > 0" class="error-details" style="margin-top: 16px;">
              <a-collapse>
                <a-collapse-panel key="errors" header="Lihat Detail Error">
                  <a-list
                    :data-source="validationResult.errors"
                    size="small"
                    :pagination="{ pageSize: 10, size: 'small' }"
                  >
                    <template #renderItem="{ item }">
                      <a-list-item>
                        <a-list-item-meta>
                          <template #title>
                            <a-tag color="error">Baris {{ item.row }}</a-tag>
                            {{ item.column }}
                          </template>
                          <template #description>
                            {{ item.message }}
                          </template>
                        </a-list-item-meta>
                      </a-list-item>
                    </template>
                  </a-list>
                </a-collapse-panel>
              </a-collapse>
            </div>

            <!-- Upload Button -->
            <div style="margin-top: 16px; text-align: right;">
              <a-button
                type="primary"
                size="large"
                :disabled="!validationResult.valid || uploading"
                :loading="uploading"
                @click="handleUpload"
              >
                <IconifyIcon icon="mdi:upload" width="16" style="margin-right: 8px;" />
                Upload Data ({{ validationResult.data?.length || 0 }} baris)
              </a-button>
            </div>

            <!-- Upload Progress -->
            <div v-if="uploading" style="margin-top: 16px;">
              <a-progress :percent="uploadProgress" :status="uploadProgress === 100 ? 'success' : 'active'" />
              <p style="text-align: center; margin-top: 8px; color: #666;">
                Mengupload data... {{ uploadProgress }}%
              </p>
            </div>
          </a-card>
        </div>
      </div>
    </a-card>

    <!-- Upload Error Modal -->
    <a-modal
      v-model:open="uploadErrorModalVisible"
      title="Detail Error Upload"
      width="800"
      :footer="null"
    >
      <div style="max-height: 500px; overflow-y: auto;">
        <a-alert
          :message="`Total ${uploadErrors.length} error ditemukan`"
          type="error"
          show-icon
          style="margin-bottom: 16px;"
        />
        <a-list
          :data-source="uploadErrors"
          size="small"
          :pagination="{ pageSize: 20, size: 'small', showSizeChanger: true }"
        >
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta>
                <template #title>
                  <a-tag color="error" style="margin-right: 8px;">
                    Baris {{ item.row || '?' }}
                  </a-tag>
                  <span style="font-weight: 500;">{{ item.column || 'general' }}</span>
                </template>
                <template #description>
                  <span style="color: #ff4d4f;">{{ item.message || 'Unknown error' }}</span>
                </template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>
      </div>
      <div style="margin-top: 16px; text-align: right;">
        <a-button @click="uploadErrorModalVisible = false">Tutup</a-button>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import type { UploadFile } from 'ant-design-vue'
import { financialReportsApi } from '../api/financialReports'
import { logger } from '../utils/logger'

// Props
const props = defineProps<{
  onUploadSuccess?: () => void
}>()

// Emits
const emit = defineEmits<{
  (e: 'upload-success'): void
}>()


// File upload
const fileList = ref<UploadFile[]>([])
const templateLoading = ref(false)
const validating = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)

// Validation result
const validationResult = ref<{
  valid: boolean
  errors: Array<{ row: number; column: string; message: string }>
  data: Array<Record<string, unknown>>
} | null>(null)

// Upload error modal
const uploadErrorModalVisible = ref(false)
const uploadErrors = ref<Array<{ row?: number; column?: string; message?: string }>>([])


// Download template
const handleDownloadTemplate = async () => {
  templateLoading.value = true
  try {
    // Download template tanpa parameter (default tahun ini)
    const blob = await financialReportsApi.downloadBulkUploadTemplate()
    
    // Generate filename berdasarkan tahun saat ini
    const currentYear = new Date().getFullYear().toString()
    const filename = `financial_report_template_${currentYear}.xlsx`
    
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    message.success('Template berhasil didownload')
  } catch (error: unknown) {
    const axiosError = error as {
      response?: {
        status?: number
        data?: { message?: string; error?: string }
      }
      message?: string
      code?: string
    }
    
    if (axiosError.response?.status === 404 || axiosError.code === 'ERR_BAD_REQUEST') {
      message.warning('Endpoint template belum tersedia. Silakan hubungi administrator.')
    } else {
      const errorMessage = axiosError.response?.data?.message ||
                          axiosError.response?.data?.error ||
                          axiosError.message ||
                          'Unknown error'
      message.error('Gagal download template: ' + errorMessage)
    }
  } finally {
    templateLoading.value = false
  }
}

// Handle before upload - validate file
const handleBeforeUpload = async (file: File): Promise<boolean> => {
  // Validate file extension
  const validExtensions = ['.xlsx', '.xls']
  const fileExtension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!validExtensions.includes(fileExtension)) {
    message.error('Format file tidak valid. Hanya file Excel (.xlsx, .xls) yang diperbolehkan.')
    return false
  }

  // Validate file
  validating.value = true
  uploadProgress.value = 0

  try {
    const result = await financialReportsApi.validateBulkExcelFile(file)
    
    // Ensure result has required structure
    if (!result.errors) {
      result.errors = []
    }
    if (!result.data) {
      result.data = []
    }
    
    validationResult.value = result

    if (result.valid) {
      message.success(`File valid. Siap untuk diupload. Total ${result.data.length} baris data.`)
    } else {
      message.warning(`Ditemukan ${result.errors.length} error. Harap perbaiki sebelum upload.`)
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memvalidasi file: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    fileList.value = []
    validationResult.value = null
    return false
  } finally {
    validating.value = false
  }

  // Don't auto upload, wait for user to click upload button
  return false
}

// Handle remove file
const handleRemoveFile = () => {
  fileList.value = []
  validationResult.value = null
  uploadProgress.value = 0
}

// Handle upload
const handleUpload = async () => {
  if (fileList.value.length === 0 || !validationResult.value || !validationResult.value.valid) {
    message.error('Tidak dapat upload. Harap perbaiki semua error terlebih dahulu.')
    return
  }

  const file = fileList.value[0]?.originFileObj
  if (!file) {
    message.error('File tidak ditemukan')
    return
  }

  uploading.value = true
  uploadProgress.value = 0

  try {
    const result = await financialReportsApi.uploadBulkFinancialReports(file, (progress) => {
      uploadProgress.value = progress
    })

    // Build success message with created/updated info
    let successMessage = ''
    if (result.created && result.updated) {
      successMessage = `Upload berhasil! ${result.created} data dibuat, ${result.updated} data diupdate.`
    } else if (result.created) {
      successMessage = `Upload berhasil! ${result.created} data berhasil dibuat.`
    } else if (result.updated) {
      successMessage = `Upload berhasil! ${result.updated} data berhasil diupdate.`
    } else {
      successMessage = `Upload berhasil! ${result.success} data berhasil diproses.`
    }

    if (result.errors && result.errors.length > 0) {
      // Log errors untuk debugging
      logger.error('Upload errors:', result.errors)
      logger.error('Upload result:', result)
      
      // Simpan error untuk ditampilkan di modal
      uploadErrors.value = result.errors
      
      // Tampilkan warning dan buka modal error
      message.warning({
        content: `${successMessage} Namun ada ${result.failed} baris yang gagal. Klik untuk melihat detail.`,
        duration: 8,
        onClick: () => {
          uploadErrorModalVisible.value = true
        },
      })
      
      // Auto buka modal error
      uploadErrorModalVisible.value = true
    } else {
      message.success(successMessage)
    }

    // Reset form
    fileList.value = []
    validationResult.value = null
    uploadProgress.value = 0

    // Emit success event
    emit('upload-success')
    if (props.onUploadSuccess) {
      props.onUploadSuccess()
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal upload: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    uploading.value = false
    uploadProgress.value = 0
  }
}
</script>

<style scoped>
.financial-report-bulk-upload {
  width: 100%;
}

.upload-content {
  padding: 8px 0;
}

.validation-result {
  margin-top: 24px;
}

.error-details {
  margin-top: 16px;
}

.ant-upload-drag-icon {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
}

.ant-upload-text {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
}

.ant-upload-hint {
  font-size: 14px;
  color: #999;
}
</style>

