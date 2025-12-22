<template>
  <div class="editable-financial-table">
    <a-form :form="form" component="false">
      <a-table
        :columns="mergedColumns"
        :data-source="dataSource"
        :pagination="false"
        :loading="loading"
        :bordered="true"
        :scroll="{ x: 'max-content' }"
        :row-class-name="(record: any, index?: number) => index !== undefined && index % 2 === 1 ? 'editable-row table-row-striped' : 'editable-row'"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'period'">
            <strong>{{ record.period }}</strong>
          </template>
          <template v-else-if="column.key === 'month'">
            <strong>{{ record.month }}</strong>
          </template>
          <template v-else-if="column.key === 'operation'">
            <span v-if="isEditing(record)">
              <a-typography-link @click="save(record.key as string)" style="margin-right: 8px;">
                Simpan
              </a-typography-link>
              <a-popconfirm title="Yakin batalkan?" @confirm="cancel">
                <a>Batal</a>
              </a-popconfirm>
            </span>
            <span v-else>
              <a-space>
                <a-typography-link 
                  :disabled="editingKey !== ''" 
                  @click="edit(record)"
                  v-if="canEdit"
                >
                  Edit
                </a-typography-link>
                <a-popconfirm
                  title="Yakin hapus data ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="handleDelete(record.key as string)"
                  v-if="canEdit"
                >
                  <a-typography-link 
                    :disabled="editingKey !== ''" 
                    type="danger"
                  >
                    <IconifyIcon icon="mdi:trash" width="16" style="color: #ff4d4f;" />
                  </a-typography-link>
                </a-popconfirm>
              </a-space>
            </span>
          </template>
          <template v-else-if="isEditing(record) && column.dataIndex && getColumnEditable(column)">
            <a-form-item
              :name="column.dataIndex"
              :style="{ margin: 0 }"
              :rules="[
                {
                  required: true,
                  message: `Harap isi ${column.title}!`,
                  validator: (_rule: any, value: any) => {
                    // Pakai Promise reference dari component scope untuk hindari context issues
                    // Ini memastikan Promise selalu tersedia bahkan di execution contexts berbeda
                    const P = PromiseRef
                    
                    // Izinkan 0 sebagai nilai valid, hanya reject undefined, null, atau empty string
                    if (value === undefined || value === null || value === '') {
                      return P.reject(new Error(`Harap isi ${column.title}!`))
                    }
                    // For numbers, 0 is valid
                    if (typeof value === 'number' && value === 0) {
                      return P.resolve()
                    }
                    // Validasi field ratio (max 100 untuk ratio berbasis persentase)
                    const isRatioField = column.dataIndex?.includes('roe') || 
                                        column.dataIndex?.includes('roi') || 
                                        column.dataIndex?.includes('ratio') || 
                                        column.dataIndex?.includes('margin') || 
                                        column.dataIndex?.includes('debt_to_equity')
                    if (isRatioField && typeof value === 'number' && value > 100) {
                      return P.reject(new Error(`${column.title} tidak boleh lebih dari 100%`))
                    }
                    return P.resolve()
                  },
                },
              ]"
            >
              <a-input-number
                v-if="getColumnInputType(column) === 'number'"
                :style="{ width: '100%' }"
                :precision="getColumnPrecision(column)"
                :max="getColumnMax(column)"
                :min="0"
                v-model:value="record[column.dataIndex]"
              />
              <a-input
                v-else
                :style="{ width: '100%' }"
                v-model:value="record[column.dataIndex]"
              />
            </a-form-item>
          </template>
          <template v-else-if="column.dataIndex && record[column.dataIndex] !== undefined">
            {{ formatCellValue(record[column.dataIndex], getColumnInputType(column)) }}
          </template>
        </template>
      </a-table>
    </a-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, unref } from 'vue'
import { Form } from 'ant-design-vue'
import type { TableColumnType } from 'ant-design-vue'
import { Icon as IconifyIcon } from '@iconify/vue'

interface ColumnType {
  title: string
  key: string
  dataIndex?: string
  editable?: boolean
  inputType?: 'number' | 'text'
  width?: number
  align?: 'left' | 'right' | 'center'
  fixed?: 'left' | 'right'
  children?: Array<{
    title: string
    key: string
    dataIndex?: string
    editable?: boolean
    inputType?: 'number' | 'text'
    width?: number
    align?: 'left' | 'right' | 'center'
  }>
}

interface Props {
  columns: ColumnType[]
  dataSource: Array<Record<string, unknown>>
  loading?: boolean
  canEdit?: boolean
  onSave?: (key: string, record: Record<string, unknown>) => Promise<void>
  onDelete?: (key: string) => Promise<void>
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  canEdit: true,
})

const emit = defineEmits<{
  save: [key: string, record: Record<string, unknown>]
  delete: [key: string]
}>()

// Type untuk form instance - di Ant Design Vue 4.x, form instance punya struktur berbeda
interface FormInstance {
  setFieldsValue?: (values: Record<string, unknown>) => void
  validateFields?: () => Promise<Record<string, unknown>>
  resetFields?: () => void
  validate?: () => Promise<Record<string, unknown>>
  modelRef?: ReturnType<typeof ref<Record<string, unknown>>>
  rulesRef?: unknown
  initialModel?: unknown
  validateInfos?: unknown
  [key: string]: unknown
}

// Initialize form - Form.useForm() returns form instance in Ant Design Vue 4.x
// In Ant Design Vue 4.x, useForm() typically returns array [form] or form instance directly
// Handle both cases to avoid undefined watch source
// Create a fallback form instance first to ensure it's always defined
const defaultFormInstance: FormInstance = {
  modelRef: ref<Record<string, unknown>>({}),
  setFieldsValue: () => {},
  validateFields: async () => ({}),
  resetFields: () => {},
  validate: async () => ({}),
} as unknown as FormInstance

// Coba ambil form dari Form.useForm(), dengan fallback ke default
// Di Ant Design Vue 4.x, useForm() bisa dipanggil tanpa arguments
let formResultRaw: unknown
try {
  // TypeScript may require arguments, but runtime doesn't - use type assertion
  formResultRaw = (Form.useForm as () => unknown)()
} catch (error) {
  console.warn('Form.useForm() failed, using fallback:', error)
  formResultRaw = null
}

// Extract form instance dengan aman - handle return array dan object
let form: FormInstance = defaultFormInstance

if (formResultRaw) {
  if (Array.isArray(formResultRaw)) {
    // If array, take first element
    const instance = formResultRaw[0]
    if (instance && typeof instance === 'object') {
      form = instance as unknown as FormInstance
    }
  } else if (typeof formResultRaw === 'object') {
    // Kalau object, pakai langsung
    form = formResultRaw as unknown as FormInstance
  }
}

// Pastikan form selalu punya methods yang diperlukan
if (!form.setFieldsValue || !form.validateFields || !form.resetFields) {
  // Merge with default to ensure all methods exist
  form = { ...defaultFormInstance, ...form } as unknown as FormInstance
}
const editingKey = ref<string>('')

// Simpan Promise reference di level component untuk pastikan tersedia di validator
// Ini mencegah error "Cannot read properties of undefined (reading 'Promise')"
const PromiseRef = Promise

// Buat setFieldsValue wrapper pakai modelRef kalau tersedia
// Di Ant Design Vue 4.x, modelRef adalah ref yang perlu diakses dengan unref
if (!form.setFieldsValue) {
  form.setFieldsValue = (values: Record<string, unknown>) => {
    console.log('setFieldsValue called with:', values)
    if (form.modelRef) {
      // modelRef adalah ref, akses property value-nya
      const modelRefValue = form.modelRef as { value?: Record<string, unknown> }
      if (modelRefValue && 'value' in modelRefValue) {
        if (!modelRefValue.value) {
          modelRefValue.value = {}
        }
        Object.assign(modelRefValue.value, values)
        console.log('setFieldsValue - modelRef.value after assign:', modelRefValue.value)
      } else {
        // Try unref approach
        const model = unref(form.modelRef)
        if (model && typeof model === 'object') {
          Object.assign(model, values)
          console.log('setFieldsValue - model after assign:', model)
        } else {
          // Initialize kalau diperlukan
          if (form.modelRef && typeof form.modelRef === 'object') {
            (form.modelRef as { value: Record<string, unknown> }).value = { ...values }
            console.log('setFieldsValue - initialized modelRef.value:', (form.modelRef as { value: Record<string, unknown> }).value)
          }
        }
      }
    } else {
      console.error('setFieldsValue - modelRef is not available')
    }
  }
}

// Buat validateFields wrapper pakai validate kalau tersedia
if (!form.validateFields && form.validate) {
  form.validateFields = async () => {
    await form.validate?.()
    if (form.modelRef) {
      const model = unref(form.modelRef)
      return (model || {}) as Record<string, unknown>
    }
    return {}
  }
}

const isEditing = (record: Record<string, unknown>) => {
  return record.key === editingKey.value
}

const getColumnEditable = (column: ColumnType): boolean => {
  if (column.editable) return true
  if (column.children) {
    return column.children.some(c => c.editable)
  }
  return false
}

const getColumnInputType = (column: ColumnType): 'number' | 'text' => {
  return column.inputType || 'number'
}

const getColumnPrecision = (column: ColumnType): number | undefined => {
  // For ratio fields, use 2 decimal places
  if (column.key?.includes('ratio') || column.key?.includes('roe') || column.key?.includes('roi') || 
      column.key?.includes('margin') || column.key?.includes('debt')) {
    return 2
  }
  return undefined
}

const getColumnMax = (column: ColumnType): number | undefined => {
  // Untuk field ratio (berbasis persentase), limit ke 100
  const isRatioField = column.dataIndex?.includes('roe') || 
                      column.dataIndex?.includes('roi') || 
                      column.dataIndex?.includes('ratio') || 
                      column.dataIndex?.includes('margin') || 
                      column.dataIndex?.includes('debt_to_equity')
  if (isRatioField) {
    return 100
  }
  // Untuk EBITDA (nilai absolut, bukan persentase), izinkan nilai lebih besar
  // Tapi tetap limit yang wajar (misalnya, 1 triliun)
  if (column.dataIndex?.includes('ebitda') && !column.dataIndex?.includes('margin')) {
    return 1000000000000 // 1 trillion
  }
  return undefined
}

const edit = (record: Record<string, unknown>) => {
  const formValues: Record<string, unknown> = {}
  
  // Kumpulkan semua field editable dari columns dan children-nya
  const collectEditableFields = (cols: ColumnType[]) => {
    cols.forEach((col) => {
      if (col.editable && col.dataIndex) {
        const value = record[col.dataIndex]
        if (value !== undefined && value !== null) {
          formValues[col.dataIndex] = value
        }
      }
      if (col.children) {
        col.children.forEach((child) => {
          if (child.editable && child.dataIndex) {
            const value = record[child.dataIndex]
            if (value !== undefined && value !== null) {
              formValues[child.dataIndex] = value
            }
          }
        })
      }
    })
  }
  
  collectEditableFields(props.columns)
  
  // Debug: Log what we're trying to set
  console.log('Edit - Record:', record)
  console.log('Edit - Form values to set:', formValues)
  console.log('Edit - Available record keys:', Object.keys(record))
  
  // Ensure form is available
  if (!form) {
    console.error('Form instance is not available')
    return
  }
  
  // Coba set form values - di Ant Design Vue 4.x, pakai modelRef
  try {
    if (typeof form.setFieldsValue === 'function') {
      // Pakai setFieldsValue kalau tersedia (wrapper kita)
      form.setFieldsValue(formValues)
      console.log('Form values set successfully using setFieldsValue')
    } else if (form.modelRef) {
      // Use modelRef directly - it's a ref, so we need to access .value
      const modelRefValue = form.modelRef as { value?: Record<string, unknown> }
      if (modelRefValue && 'value' in modelRefValue) {
        if (!modelRefValue.value) {
          modelRefValue.value = {}
        }
        // Set values satu per satu untuk pastikan reactivity
        Object.keys(formValues).forEach((key) => {
          if (modelRefValue.value) {
            modelRefValue.value[key] = formValues[key]
          }
        })
        console.log('Form values set successfully using modelRef.value:', modelRefValue.value)
      } else {
        // Try unref approach
        const model = unref(form.modelRef)
        if (model && typeof model === 'object') {
          Object.keys(formValues).forEach((key) => {
            (model as Record<string, unknown>)[key] = formValues[key]
          })
          console.log('Form values set successfully using unref(modelRef):', model)
        } else {
          console.error('modelRef is not accessible:', form.modelRef, 'Type:', typeof form.modelRef)
          return
        }
      }
    } else {
      console.error('Cannot set form values - no setFieldsValue or modelRef available', form)
      return
    }
  } catch (error) {
    console.error('Error setting form values:', error)
    return
  }
  
  editingKey.value = record.key as string
}

const cancel = () => {
  editingKey.value = ''
}

const handleDelete = async (key: string) => {
  if (props.onDelete) {
    await props.onDelete(key)
  } else {
    emit('delete', key)
  }
}

const save = async (key: string) => {
  if (!form) {
    console.error('Form instance is not available')
    return
  }
  
  try {
    // In Ant Design Vue 4.x, validateFields might be different
    let row: Record<string, unknown>
    if (typeof form.validateFields === 'function') {
      // Pakai validateFields kalau tersedia (wrapper kita)
      row = await form.validateFields()
    } else if (form.validate) {
      // Use validate method and then get modelRef
      await form.validate()
      const model = unref(form.modelRef)
      if (model && typeof model === 'object') {
        row = model as Record<string, unknown>
      } else {
        console.error('Cannot get form values from modelRef')
        return
      }
    } else if (form.modelRef) {
      // Fallback: use modelRef directly
      const model = unref(form.modelRef)
      if (model && typeof model === 'object') {
        row = model as Record<string, unknown>
      } else {
        console.error('Cannot get form values from modelRef')
        return
      }
    } else {
      console.error('Cannot validate form - no validateFields, validate, or modelRef available')
      return
    }
    const newData = [...props.dataSource]
    const index = newData.findIndex((item) => item.key === key)
    
    if (index > -1) {
      const item = newData[index]
      const updatedRecord = {
        ...item,
        ...row,
      }
      newData.splice(index, 1, updatedRecord)
      
      if (props.onSave) {
        await props.onSave(key, updatedRecord)
      } else {
        emit('save', key, updatedRecord)
      }
      
      editingKey.value = ''
    }
  } catch (errInfo) {
    console.log('Validate Failed:', errInfo)
  }
}

// Format nilai cell untuk ditampilkan
const formatCellValue = (value: unknown, inputType?: 'number' | 'text'): string => {
  if (value === undefined || value === null) return '-'
  
  if (inputType === 'number') {
    const numValue = typeof value === 'string' ? parseFloat(value) : value
    if (typeof numValue === 'number' && !isNaN(numValue)) {
      // Format sebagai currency untuk angka besar, atau sebagai ratio untuk angka kecil
      if (numValue >= 1000000000) {
        return `Rp ${(numValue / 1000000000).toFixed(2)}M`
      } else if (numValue >= 1000000) {
        return `Rp ${(numValue / 1000000).toFixed(2)}Jt`
      } else if (numValue >= 1000) {
        return `Rp ${(numValue / 1000).toFixed(2)}Rb`
      } else if (numValue < 100 && numValue > 0) {
        // Kemungkinan ratio/persentase
        return `${numValue.toFixed(2)}%`
      }
      return numValue.toLocaleString('id-ID')
    }
  }
  
  return String(value)
}

const mergedColumns = computed(() => {
  return props.columns.map((col) => {
    // Kalau column punya children, handle children
    if (col.children) {
      return {
        ...col,
        children: col.children.map((child) => {
          if (!child.editable) {
            return child
          }
          return {
            ...child,
            onCell: (record: Record<string, unknown>) => ({
              record,
              inputType: child.inputType || 'number',
              dataIndex: child.dataIndex || child.key,
              title: child.title,
              editing: isEditing(record),
            }),
          }
        }),
      } as TableColumnType
    }
    
    // Kalau column tidak editable, return apa adanya
    if (!col.editable) {
      return col as TableColumnType
    }
    
    // Column editable
    return {
      ...col,
      onCell: (record: Record<string, unknown>) => ({
        record,
        inputType: col.inputType || 'number',
        dataIndex: col.dataIndex || col.key,
        title: col.title,
        editing: isEditing(record),
      }),
    } as TableColumnType
  })
})
</script>

<style scoped>
.editable-financial-table :deep(.editable-row) {
  cursor: pointer;
}

.editable-financial-table :deep(.editable-row:hover) {
  background: #fafafa;
}

.editable-financial-table :deep(.table-row-striped) {
  background-color: #fafafa;
}

.editable-financial-table :deep(.table-row-striped:hover) {
  background-color: #f0f0f0;
}
</style>
