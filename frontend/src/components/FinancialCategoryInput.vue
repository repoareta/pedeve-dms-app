<template>
  <div class="financial-category-input">
    <a-button 
      type="primary" 
      @click="showAddModal"
      style="margin-bottom: 24px;"
      :disabled="!canEdit"
    >
      <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 4px;" />
      Add Report {{ categoryLabel }}
    </a-button>

    <!-- Editable Table -->
    <EditableFinancialTable
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :can-edit="canEdit"
      @save="handleSave"
      @delete="handleDelete"
    />

    <!-- Modal Add/Edit -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="800"
      @ok="handleSubmit"
      @cancel="handleCancel"
      :confirm-loading="submitting"
    >
      <a-form
        :model="formData"
        :label-col="{ span: 8 }"
        :wrapper-col="{ span: 16 }"
        layout="horizontal"
      >
        <a-form-item label="Tahun" required>
          <a-select v-model:value="formData.year" style="width: 100%;" :disabled="!!editingRecord">
            <a-select-option v-for="year in availableYears" :key="year" :value="year">
              {{ year }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Bulan" required>
          <a-select v-model:value="formData.month" style="width: 100%;" :disabled="!!editingRecord">
            <a-select-option value="01">Januari</a-select-option>
            <a-select-option value="02">Februari</a-select-option>
            <a-select-option value="03">Maret</a-select-option>
            <a-select-option value="04">April</a-select-option>
            <a-select-option value="05">Mei</a-select-option>
            <a-select-option value="06">Juni</a-select-option>
            <a-select-option value="07">Juli</a-select-option>
            <a-select-option value="08">Agustus</a-select-option>
            <a-select-option value="09">September</a-select-option>
            <a-select-option value="10">Oktober</a-select-option>
            <a-select-option value="11">November</a-select-option>
            <a-select-option value="12">Desember</a-select-option>
          </a-select>
        </a-form-item>

        <a-divider>{{ categoryLabel }}</a-divider>

        <a-row :gutter="16">
          <a-col :span="12" v-for="item in items" :key="item.key">
            <a-form-item :label="item.label">
              <a-input-number
                v-if="item.isRatio"
                v-model:value="formData[item.field]"
                style="width: 100%;"
                :precision="2"
                :min="0"
              />
              <a-input-number
                v-else
                v-model:value="formData[item.field]"
                style="width: 100%;"
                :formatter="(value: number | string | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="Keterangan">
          <a-textarea v-model:value="formData.remark" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { financialReportsApi, type FinancialReport, type CreateFinancialReportRequest } from '../api/financialReports'
import EditableFinancialTable from './EditableFinancialTable.vue'
import dayjs from 'dayjs'
import { Icon as IconifyIcon } from '@iconify/vue'

interface FinancialItem {
  key: string
  label: string
  field: string
  isRatio: boolean
}

interface Props {
  companyId: string
  category: 'neraca' | 'laba-rugi' | 'cashflow' | 'rasio'
  items: FinancialItem[]
  canEdit?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  canEdit: true,
})

const emit = defineEmits<{
  saved: []
}>()

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const editingRecord = ref<FinancialReport | null>(null)
const reports = ref<FinancialReport[]>([])

const categoryLabel = computed(() => {
  const labels = {
    'neraca': 'Neraca',
    'laba-rugi': 'Laba Rugi',
    'cashflow': 'Cashflow',
    'rasio': 'Rasio Keuangan',
  }
  return labels[props.category]
})

const availableYears = computed(() => {
  const currentYear = parseInt(dayjs().format('YYYY'))
  const years: string[] = []
  for (let i = 0; i < 6; i++) {
    years.push(String(currentYear - i))
  }
  return years
})

const formData = ref<Record<string, unknown>>({
  year: dayjs().format('YYYY'),
  month: dayjs().format('MM'),
  remark: '',
})

// Initialize form data dengan semua field dari items
watch(() => props.items, (newItems) => {
  newItems.forEach((item) => {
    if (!(item.field in formData.value)) {
      formData.value[item.field] = 0
    }
  })
}, { immediate: true })

const modalTitle = computed(() => {
  return editingRecord.value ? `Edit ${categoryLabel.value}` : `Add Report ${categoryLabel.value}`
})

const showAddModal = () => {
  editingRecord.value = null
  // Reset form
  formData.value = {
    year: dayjs().format('YYYY'),
    month: dayjs().format('MM'),
    remark: '',
  }
  props.items.forEach((item) => {
    formData.value[item.field] = 0
  })
  modalVisible.value = true
}

const handleCancel = () => {
  modalVisible.value = false
  editingRecord.value = null
}

const handleSubmit = async () => {
  if (!formData.value.year || !formData.value.month) {
    message.warning('Harap isi Tahun dan Bulan')
    return
  }

  submitting.value = true
  try {
    const period = `${formData.value.year}-${formData.value.month}`
    const requestData: Record<string, unknown> = {
      company_id: props.companyId,
      year: formData.value.year,
      period,
      is_rkap: false,
    }

    // Add all fields from items
    props.items.forEach((item) => {
      requestData[item.field] = formData.value[item.field] || 0
    })

    // Add other required fields with default 0 if not in items
    if (props.category === 'neraca') {
      // Only include neraca fields
    } else if (props.category === 'laba-rugi') {
      // Only include laba rugi fields
    } else if (props.category === 'cashflow') {
      // Only include cashflow fields
    } else if (props.category === 'rasio') {
      // Only include rasio fields
    }

    // Set other fields to 0 if not in current category
    const allFields = [
      'current_assets', 'non_current_assets', 'short_term_liabilities', 'long_term_liabilities', 'equity',
      'revenue', 'operating_expenses', 'operating_profit', 'other_income', 'tax', 'net_profit',
      'operating_cashflow', 'investing_cashflow', 'financing_cashflow', 'ending_balance',
      'roe', 'roi', 'current_ratio', 'cash_ratio', 'ebitda', 'ebitda_margin', 'net_profit_margin', 'operating_profit_margin', 'debt_to_equity',
    ]

    allFields.forEach((field) => {
      if (!(field in requestData)) {
        requestData[field] = 0
      }
    })

    if (formData.value.remark) {
      requestData.remark = formData.value.remark
    }

    if (editingRecord.value) {
      // Update existing
      await financialReportsApi.update(editingRecord.value.id, requestData)
      message.success(`${categoryLabel.value} berhasil diupdate`)
    } else {
      // Create new - use double cast to satisfy TypeScript
      await financialReportsApi.create(requestData as unknown as CreateFinancialReportRequest)
      message.success(`${categoryLabel.value} berhasil ditambahkan`)
    }

    modalVisible.value = false
    editingRecord.value = null
    await loadReports()
    emit('saved')
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(`Gagal menyimpan: ${err.response?.data?.message || err.message || 'Unknown error'}`)
  } finally {
    submitting.value = false
  }
}

const loadReports = async () => {
  if (!props.companyId) {
    reports.value = []
    return
  }
  
  loading.value = true
  try {
    const allReports = await financialReportsApi.getByCompanyId(props.companyId)
    // Filter hanya realisasi (bukan RKAP)
    const realisasiReports = allReports.filter(r => !r.is_rkap)
    reports.value = realisasiReports
    console.log(`[${props.category}] Loaded ${realisasiReports.length} reports for company ${props.companyId}`)
  } catch (error: unknown) {
    console.error(`[${props.category}] Failed to load reports:`, error)
    reports.value = []
  } finally {
    loading.value = false
  }
}

const getMonthName = (month: string): string => {
  const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  const monthIndex = parseInt(month, 10) - 1
  return months[monthIndex] || month
}

const tableData = computed(() => {
  if (!reports.value || reports.value.length === 0) {
    return []
  }
  
  const data = reports.value.map((report) => {
    const row: Record<string, unknown> = {
      key: report.id,
      period: report.year,
      month: getMonthName(report.period.split('-')[1] || ''),
    }

    props.items.forEach((item) => {
      const value = (report as unknown as Record<string, unknown>)[item.field]
      row[`${item.key}_realisasi`] = value ?? 0
    })

    return row
  }).sort((a, b) => {
    // Sort by month: December (12) first, January (1) last
    // First compare by year (newest first)
    const aYear = parseInt(String(a.period), 10)
    const bYear = parseInt(String(b.period), 10)
    if (aYear !== bYear) {
      return bYear - aYear // Newest year first
    }
    
    // If same year, sort by month: December (12) to January (1)
    const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
    const aMonthIndex = monthNames.indexOf(String(a.month))
    const bMonthIndex = monthNames.indexOf(String(b.month))
    
    // December (index 11) should be first, January (index 0) should be last
    // So we reverse the order: higher index (December) comes first
    return bMonthIndex - aMonthIndex
  })
  
  console.log(`[${props.category}] Table data:`, data.length, 'rows')
  return data
})

const tableColumns = computed(() => {
  const baseColumns = [
    {
      title: 'Periode',
      key: 'period',
      dataIndex: 'period',
      width: 100,
      fixed: 'left' as const,
      align: 'left' as const,
      editable: false,
    },
    {
      title: 'Bulan',
      key: 'month',
      dataIndex: 'month',
      width: 120,
      fixed: 'left' as const,
      align: 'left' as const,
      editable: false,
    },
  ]

  const itemColumns = props.items.map((item) => ({
    title: item.label,
    key: item.key,
    align: 'center' as const,
    children: [
      {
        title: 'Realisasi',
        key: `${item.key}_realisasi`,
        dataIndex: `${item.key}_realisasi`,
        align: 'right' as const,
        width: 150,
        editable: props.canEdit,
        inputType: item.isRatio ? 'number' as const : 'number' as const,
      },
    ],
  }))

  const operationColumn = {
    title: 'Aksi',
    key: 'operation',
    width: 150,
    fixed: 'right' as const,
    editable: false,
  }

  return [...baseColumns, ...itemColumns, operationColumn]
})

const handleSave = async (key: string, record: Record<string, unknown>) => {
  if (!props.canEdit) {
    message.warning('Anda tidak memiliki izin untuk mengedit data')
    return
  }

  const report = reports.value.find(r => r.id === key)
  if (!report) return

  submitting.value = true
  try {
    const updateData: Record<string, unknown> = {}

    props.items.forEach((item) => {
      const fieldKey = `${item.key}_realisasi`
      if (fieldKey in record) {
        updateData[item.field] = record[fieldKey] ?? 0
      }
    })

    await financialReportsApi.update(key, updateData)
    message.success('Data berhasil diupdate')
    await loadReports()
    emit('saved')
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(`Gagal mengupdate: ${err.response?.data?.message || err.message || 'Unknown error'}`)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (key: string) => {
  if (!props.canEdit) {
    message.warning('Anda tidak memiliki izin untuk menghapus data')
    return
  }

  submitting.value = true
  try {
    await financialReportsApi.delete(key)
    message.success('Data berhasil dihapus')
    await loadReports()
    emit('saved')
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(`Gagal menghapus: ${err.response?.data?.message || err.message || 'Unknown error'}`)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadReports()
})

watch(() => props.companyId, () => {
  loadReports()
})
</script>

<style scoped>
.financial-category-input {
  padding: 16px;
}
</style>
