<template>
  <div class="financial-report-form">
    <a-spin :spinning="loading">
      <a-form :label-col="{ span: 8 }" :wrapper-col="{ span: 16 }" layout="horizontal">
        <!-- Header: Tahun/Periode -->
        <a-card :bordered="false" style="margin-bottom: 24px;">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Tahun" required>
                <a-space direction="vertical" style="width: 100%;">
                  <a-select 
                    v-model:value="formData.year" 
                    style="width: 100%;" 
                    show-search
                    :filter-option="filterYear"
                    placeholder="Pilih atau ketik tahun (contoh: 2025, 2030)"
                    @change="handleYearChange"
                    @search="handleYearSearch"
                    :dropdown-match-select-width="false"
                    allow-clear
                    :not-found-content="null"
                  >
                    <a-select-option 
                      v-for="yearOption in availableYearsWithStatus" 
                      :key="yearOption.value" 
                      :value="yearOption.value"
                      :disabled="yearOption.disabled"
                    >
                      <span :style="{ color: getYearStatusColor(yearOption.status) }">
                        {{ yearOption.value }}
                        <span v-if="yearOption.status === 'exists'" style="margin-left: 8px;">
                          ✓ (Sudah ada)
                        </span>
                        <span v-else-if="yearOption.status === 'future'" style="margin-left: 8px;">
                          (Tahun Depan)
                        </span>
                        <span v-else-if="yearOption.status === 'missing'" style="margin-left: 8px;">
                          ⚠ (Belum ada)
                        </span>
                      </span>
                    </a-select-option>
                  </a-select>
                  <div v-if="isRKAP" style="font-size: 12px; color: #8c8c8c;">
                    <span>💡 Tip: Ketik tahun 4 digit di search box untuk tahun yang tidak ada di list (contoh: 2030, 2035, 2040)</span>
                  </div>
                  <div v-if="isRKAP && formData.year && getYearStatus(formData.year)" style="font-size: 12px; margin-top: 4px;">
                    <a-tag :color="getYearStatus(formData.year) === 'exists' ? 'success' : (getYearStatus(formData.year) === 'future' ? 'processing' : 'warning')">
                      <span v-if="getYearStatus(formData.year) === 'exists'">✓ RKAP untuk tahun {{ formData.year }} sudah ada</span>
                      <span v-else-if="getYearStatus(formData.year) === 'future'">🔮 Tahun {{ formData.year }} adalah tahun depan</span>
                      <span v-else>⚠ RKAP untuk tahun {{ formData.year }} belum ada</span>
                    </a-tag>
                  </div>
                </a-space>
              </a-form-item>
            </a-col>
            <a-col v-if="!isRKAP" :span="12">
              <a-form-item label="Bulan" required>
                <a-select v-model:value="month" style="width: 100%;" :disabled="!!existingReport">
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
            </a-col>
          </a-row>
          <a-alert
            v-if="isRKAP"
            message="RKAP hanya boleh diinput 1x per tahun per perusahaan"
            type="info"
            show-icon
            style="margin-top: 16px;"
          />
          <a-alert
            v-else
            message="Realisasi diinput setiap bulan. Pastikan periode yang dipilih belum ada datanya."
            type="info"
            show-icon
            style="margin-top: 16px;"
          />
          
          <!-- Warning untuk tahun depan -->
          <a-alert
            v-if="isRKAP && isFutureYear(formData.year)"
            message="Anda sedang membuat RKAP untuk tahun depan. Pastikan data yang diinput adalah perencanaan untuk tahun tersebut."
            type="warning"
            show-icon
            style="margin-top: 16px;"
          />
        </a-card>

        <!-- Neraca (Balance Sheet) -->
        <a-card title="A. Neraca (Balance Sheet)" :bordered="false" style="margin-bottom: 24px;">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Aset Lancar">
                <a-input-number
                  v-model:value="formData.current_assets"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Aset Tidak Lancar">
                <a-input-number
                  v-model:value="formData.non_current_assets"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Liabilitas Jangka Pendek">
                <a-input-number
                  v-model:value="formData.short_term_liabilities"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Liabilitas Jangka Panjang">
                <a-input-number
                  v-model:value="formData.long_term_liabilities"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Ekuitas">
                <a-input-number
                  v-model:value="formData.equity"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Laba Rugi (Profit & Loss) -->
        <a-card title="B. Laba Rugi (Profit & Loss)" :bordered="false" style="margin-bottom: 24px;">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Revenue">
                <a-input-number
                  v-model:value="formData.revenue"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Beban Usaha">
                <a-input-number
                  v-model:value="formData.operating_expenses"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Laba Usaha">
                <a-input-number
                  v-model:value="formData.operating_profit"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Pendapatan Lain-Lain">
                <a-input-number
                  v-model:value="formData.other_income"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Tax">
                <a-input-number
                  v-model:value="formData.tax"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Laba Bersih">
                <a-input-number
                  v-model:value="formData.net_profit"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Cashflow -->
        <a-card title="C. Cashflow" :bordered="false" style="margin-bottom: 24px;">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Arus kas bersih dari operasi">
                <a-input-number
                  v-model:value="formData.operating_cashflow"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Arus kas bersih dari investasi">
                <a-input-number
                  v-model:value="formData.investing_cashflow"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Arus kas bersih dari pendanaan">
                <a-input-number
                  v-model:value="formData.financing_cashflow"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Saldo Akhir">
                <a-input-number
                  v-model:value="formData.ending_balance"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Rasio Keuangan (%) -->
        <a-card title="D. Rasio Keuangan (%)" :bordered="false" style="margin-bottom: 24px;">
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="ROE (Return on Equity)">
                <a-input-number
                  v-model:value="formData.roe"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="ROI (Return on Investment)">
                <a-input-number
                  v-model:value="formData.roi"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Rasio Lancar">
                <a-input-number
                  v-model:value="formData.current_ratio"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Rasio Kas">
                <a-input-number
                  v-model:value="formData.cash_ratio"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="EBITDA">
                <a-input-number
                  v-model:value="formData.ebitda"
                  style="width: 100%;"
                  :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                  :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="EBITDA Margin">
                <a-input-number
                  v-model:value="formData.ebitda_margin"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Net Profit Margin">
                <a-input-number
                  v-model:value="formData.net_profit_margin"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Operating Profit Margin">
                <a-input-number
                  v-model:value="formData.operating_profit_margin"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Debt to Equity">
                <a-input-number
                  v-model:value="formData.debt_to_equity"
                  style="width: 100%;"
                  :precision="2"
                  :min="0"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Remark -->
        <a-card title="Keterangan" :bordered="false" style="margin-bottom: 24px;">
          <a-form-item label="Remark">
            <a-textarea v-model:value="formData.remark" :rows="3" placeholder="Keterangan tambahan (opsional)" />
          </a-form-item>
        </a-card>

        <!-- Action Buttons -->
        <div style="text-align: right; margin-top: 24px;">
          <a-space>
            <a-button @click="handleReset">Reset</a-button>
            <a-button type="primary" @click="handleSubmit" :loading="loading">
              {{ existingReport ? 'Update' : 'Simpan' }}
            </a-button>
          </a-space>
        </div>
      </a-form>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { financialReportsApi, type CreateFinancialReportRequest, type FinancialReport } from '../api/financialReports'
import dayjs from 'dayjs'

const props = defineProps<{
  companyId: string
  isRKAP: boolean
}>()

const emit = defineEmits<{
  saved: []
}>()

const loading = ref(false)
const existingReport = ref<FinancialReport | null>(null)

const formData = ref<CreateFinancialReportRequest>({
  company_id: props.companyId,
  year: dayjs().format('YYYY'),
  period: '',
  is_rkap: props.isRKAP,
  current_assets: 0,
  non_current_assets: 0,
  short_term_liabilities: 0,
  long_term_liabilities: 0,
  equity: 0,
  revenue: 0,
  operating_expenses: 0,
  operating_profit: 0,
  other_income: 0,
  tax: 0,
  net_profit: 0,
  operating_cashflow: 0,
  investing_cashflow: 0,
  financing_cashflow: 0,
  ending_balance: 0,
  roe: 0,
  roi: 0,
  current_ratio: 0,
  cash_ratio: 0,
  ebitda: 0,
  ebitda_margin: 0,
  net_profit_margin: 0,
  operating_profit_margin: 0,
  debt_to_equity: 0,
  remark: undefined,
})

const month = ref<string>(dayjs().format('MM'))

// RKAP years yang sudah ada (untuk status indicator)
const existingRKAPYears = ref<string[]>([])
const loadingRKAPYears = ref(false)

// Available years: 5 tahun lalu, tahun ini, 10 tahun depan (total 16 tahun)
// Range ini cukup untuk kebutuhan perencanaan jangka menengah
const availableYears = computed(() => {
  const currentYear = parseInt(dayjs().format('YYYY'))
  const years: string[] = []
  // 10 tahun ke depan
  for (let i = 10; i > 0; i--) {
    years.push(String(currentYear + i))
  }
  // Tahun ini
  years.push(String(currentYear))
  // 5 tahun ke belakang
  for (let i = 1; i <= 5; i++) {
    years.push(String(currentYear - i))
  }
  return years
})

// Available years dengan status indicator
interface YearOption {
  value: string
  status: 'exists' | 'missing' | 'future'
  disabled: boolean
}

const availableYearsWithStatus = computed<YearOption[]>(() => {
  if (!props.isRKAP) {
    // Untuk Realisasi, tidak perlu status indicator
    return availableYears.value.map(year => ({
      value: year,
      status: 'missing' as const,
      disabled: false,
    }))
  }
  
  const currentYear = parseInt(dayjs().format('YYYY'))
  const result: YearOption[] = availableYears.value.map(year => {
    const yearNum = parseInt(year)
    const hasRKAP = existingRKAPYears.value.includes(year)
    const isFuture = yearNum > currentYear
    
    return {
      value: year,
      status: hasRKAP ? 'exists' : (isFuture ? 'future' : 'missing'),
      disabled: false,
    }
  })
  
  // Jika user sudah input tahun manual yang tidak ada di list, tambahkan ke result
  if (formData.value.year && !availableYears.value.includes(formData.value.year)) {
    const yearNum = parseInt(formData.value.year)
    const hasRKAP = existingRKAPYears.value.includes(formData.value.year)
    const isFuture = yearNum > currentYear
    
    result.push({
      value: formData.value.year,
      status: hasRKAP ? 'exists' : (isFuture ? 'future' : 'missing'),
      disabled: false,
    })
  }
  
  return result
})

// Helper functions
const isFutureYear = (year: string): boolean => {
  const currentYear = parseInt(dayjs().format('YYYY'))
  return parseInt(year) > currentYear
}

// Get status untuk tahun tertentu
const getYearStatus = (year: string): 'exists' | 'missing' | 'future' | null => {
  if (!year || !props.isRKAP) return null
  const currentYear = parseInt(dayjs().format('YYYY'))
  const yearNum = parseInt(year)
  const hasRKAP = existingRKAPYears.value.includes(year)
  const isFuture = yearNum > currentYear
  
  if (hasRKAP) return 'exists'
  if (isFuture) return 'future'
  return 'missing'
}

const getYearStatusColor = (status: string): string => {
  switch (status) {
    case 'exists':
      return '#52c41a' // Green
    case 'future':
      return '#1890ff' // Blue
    case 'missing':
      return '#faad14' // Orange
    default:
      return '#000000'
  }
}

const filterYear = (input: string, option: { value: string }) => {
  // Jika tidak ada input, tampilkan semua
  if (!input) {
    return true
  }
  // Jika user mengetik tahun yang valid (4 digit angka), izinkan semua opsi
  if (/^\d{4}$/.test(input)) {
    return true
  }
  return option.value.toLowerCase().includes(input.toLowerCase())
}

// Handler untuk search tahun - jika user mengetik tahun yang valid, set sebagai value
const handleYearSearch = (value: string) => {
  // Jika user mengetik tahun yang valid (4 digit angka), set sebagai value
  if (value && /^\d{4}$/.test(value)) {
    const yearNum = parseInt(value)
    // Validasi: tahun harus masuk akal (misalnya antara 2000-2100)
    if (yearNum >= 2000 && yearNum <= 2100) {
      // Set tahun jika berbeda dengan yang sekarang
      // Note: a-select akan menerima value ini meskipun tidak ada di list
      // karena kita sudah menambahkan logika di availableYearsWithStatus
      if (formData.value.year !== value) {
        formData.value.year = value
        // Trigger change handler
        handleYearChange(value)
      }
    } else {
      message.warning('Tahun harus antara 2000-2100')
    }
  }
}


// Load existing RKAP years
const loadRKAPYears = async () => {
  if (!props.companyId || !props.isRKAP) return
  
  loadingRKAPYears.value = true
  try {
    existingRKAPYears.value = await financialReportsApi.getRKAPYears(props.companyId)
  } catch (error) {
    console.error('Failed to load RKAP years:', error)
    existingRKAPYears.value = []
  } finally {
    loadingRKAPYears.value = false
  }
}

// Watch year and month to update period
watch([() => formData.value.year, month], ([year, monthValue]) => {
  if (props.isRKAP) {
    formData.value.period = year
  } else {
    formData.value.period = `${year}-${monthValue}`
  }
}, { immediate: true })

// Handler untuk perubahan tahun - reset existing report dan load data baru
const handleYearChange = (value: string) => {
  // Validasi tahun: harus 4 digit angka
  if (value && !/^\d{4}$/.test(value)) {
    message.warning('Tahun harus berupa 4 digit angka (contoh: 2025)')
    return
  }
  
  // Reset existing report ketika tahun berubah
  existingReport.value = null
  // Reset form data ke default values (tapi tetap pertahankan tahun yang dipilih)
  const selectedYear = value || formData.value.year
  const selectedMonth = month.value
  formData.value = {
    company_id: props.companyId,
    year: selectedYear,
    period: props.isRKAP ? selectedYear : `${selectedYear}-${selectedMonth}`,
    is_rkap: props.isRKAP,
    current_assets: 0,
    non_current_assets: 0,
    short_term_liabilities: 0,
    long_term_liabilities: 0,
    equity: 0,
    revenue: 0,
    operating_expenses: 0,
    operating_profit: 0,
    other_income: 0,
    tax: 0,
    net_profit: 0,
    operating_cashflow: 0,
    investing_cashflow: 0,
    financing_cashflow: 0,
    ending_balance: 0,
    roe: 0,
    roi: 0,
    current_ratio: 0,
    cash_ratio: 0,
    ebitda: 0,
    ebitda_margin: 0,
    net_profit_margin: 0,
    operating_profit_margin: 0,
    debt_to_equity: 0,
    remark: undefined,
  }
  // Load existing report untuk tahun yang baru dipilih
  loadExistingReport()
}

// Load existing report if exists
const loadExistingReport = async () => {
  if (!props.companyId || !formData.value.year) return
  
  try {
    if (props.isRKAP) {
      existingReport.value = await financialReportsApi.getRKAP(props.companyId, formData.value.year)
    } else {
      existingReport.value = await financialReportsApi.getRealisasi(props.companyId, formData.value.period)
    }
    
    if (existingReport.value) {
      // Populate form with existing data
      formData.value.current_assets = existingReport.value.current_assets
      formData.value.non_current_assets = existingReport.value.non_current_assets
      formData.value.short_term_liabilities = existingReport.value.short_term_liabilities
      formData.value.long_term_liabilities = existingReport.value.long_term_liabilities
      formData.value.equity = existingReport.value.equity
      formData.value.revenue = existingReport.value.revenue
      formData.value.operating_expenses = existingReport.value.operating_expenses
      formData.value.operating_profit = existingReport.value.operating_profit
      formData.value.other_income = existingReport.value.other_income
      formData.value.tax = existingReport.value.tax
      formData.value.net_profit = existingReport.value.net_profit
      formData.value.operating_cashflow = existingReport.value.operating_cashflow
      formData.value.investing_cashflow = existingReport.value.investing_cashflow
      formData.value.financing_cashflow = existingReport.value.financing_cashflow
      formData.value.ending_balance = existingReport.value.ending_balance
      formData.value.roe = existingReport.value.roe
      formData.value.roi = existingReport.value.roi
      formData.value.current_ratio = existingReport.value.current_ratio
      formData.value.cash_ratio = existingReport.value.cash_ratio
      formData.value.ebitda = existingReport.value.ebitda
      formData.value.ebitda_margin = existingReport.value.ebitda_margin
      formData.value.net_profit_margin = existingReport.value.net_profit_margin
      formData.value.operating_profit_margin = existingReport.value.operating_profit_margin
      formData.value.debt_to_equity = existingReport.value.debt_to_equity
      formData.value.remark = existingReport.value.remark || undefined
      
      if (!props.isRKAP && existingReport.value.period) {
        const periodParts = existingReport.value.period.split('-')
        if (periodParts.length === 2) {
          const monthPart = periodParts[1]
          if (monthPart) {
            month.value = monthPart
          }
        }
      }
    }
  } catch {
    // Report doesn't exist yet, that's okay
    existingReport.value = null
  }
}


const handleSubmit = async () => {
  if (!formData.value.year) {
    message.error('Tahun harus diisi')
    return
  }
  
  if (!props.isRKAP && !month.value) {
    message.error('Bulan harus diisi')
    return
  }
  
  // Konfirmasi sebelum update RKAP
  if (existingReport.value && props.isRKAP) {
    return new Promise<void>((resolve) => {
      Modal.confirm({
        title: 'Konfirmasi Update RKAP',
        content: `Apakah Anda yakin ingin mengupdate RKAP untuk tahun ${formData.value.year}? Perubahan ini akan dicatat dalam history perubahan data.`,
        okText: 'Ya, Update',
        cancelText: 'Batal',
        onOk: async () => {
          await performSubmit()
          resolve()
        },
        onCancel: () => {
          resolve()
        },
      })
    })
  }
  
  // Untuk create atau update Realisasi, langsung submit tanpa konfirmasi
  await performSubmit()
}

const performSubmit = async () => {
  loading.value = true
  try {
    if (existingReport.value) {
      // Update existing report
      await financialReportsApi.update(existingReport.value.id, formData.value)
      message.success('Laporan keuangan berhasil diupdate')
    } else {
      // Create new report
      await financialReportsApi.create(formData.value)
      message.success('Laporan keuangan berhasil disimpan')
    }
    
    // Reload RKAP years setelah save/update
    if (props.isRKAP) {
      await loadRKAPYears()
    }
    
    emit('saved')
    await loadExistingReport()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(err.response?.data?.message || err.message || 'Gagal menyimpan laporan keuangan')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  formData.value = {
    company_id: props.companyId,
    year: dayjs().format('YYYY'),
    period: props.isRKAP ? dayjs().format('YYYY') : `${dayjs().format('YYYY')}-${dayjs().format('MM')}`,
    is_rkap: props.isRKAP,
    current_assets: 0,
    non_current_assets: 0,
    short_term_liabilities: 0,
    long_term_liabilities: 0,
    equity: 0,
    revenue: 0,
    operating_expenses: 0,
    operating_profit: 0,
    other_income: 0,
    tax: 0,
    net_profit: 0,
    operating_cashflow: 0,
    investing_cashflow: 0,
    financing_cashflow: 0,
    ending_balance: 0,
    roe: 0,
    roi: 0,
    current_ratio: 0,
    cash_ratio: 0,
    ebitda: 0,
    ebitda_margin: 0,
    net_profit_margin: 0,
    operating_profit_margin: 0,
    debt_to_equity: 0,
    remark: undefined,
  }
  month.value = dayjs().format('MM')
  existingReport.value = null
}

onMounted(() => {
  formData.value.company_id = props.companyId
  loadRKAPYears()
  loadExistingReport()
})

watch(() => props.companyId, (newId) => {
  if (newId) {
    formData.value.company_id = newId
    loadRKAPYears()
    loadExistingReport()
  }
})

</script>

<style scoped>
.financial-report-form {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}
</style>
