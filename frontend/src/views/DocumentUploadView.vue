<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import axios from 'axios'
import dayjs, { type Dayjs } from 'dayjs'
import DashboardHeader from '../components/DashboardHeader.vue'
import { Icon as IconifyIcon } from '@iconify/vue'
import documentsApi, { type DocumentFolder, type DocumentItem, type DocumentType } from '../api/documents'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// Check if this is edit mode
const isEditMode = computed(() => !!route.params.id)
const documentId = computed(() => route.params.id as string | undefined)
const loading = ref(false)

// Document types
const documentTypes = ref<DocumentType[]>([])
const loadingDocumentTypes = ref(false)
const documentTypeSearchValue = ref('')

// Check if user can manage document types (superadmin/administrator only)
const canManageDocumentTypes = computed(() => {
  const role = authStore.user?.role?.toLowerCase() || ''
  return role === 'superadmin' || role === 'administrator'
})

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
const isGeneratingReference = ref(false) // Flag to prevent multiple simultaneous generations

const formState = ref({
  title: '',
  documentId: '',
  docType: [] as string[], // Changed to array for multiple selection
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

// Normalize document type for reference number format
const normalizeDocType = (docType: string | undefined): string => {
  if (!docType || typeof docType !== 'string') return 'DOC'
  
  return docType
    .trim()
    .toUpperCase()
    .replace(/\s+/g, '-')      // Multiple spaces → single dash
    .replace(/[\/\\]/g, '-')   // Slashes → dash
    .replace(/[^A-Z0-9\-]/g, '') // Remove invalid chars (keep only A-Z, 0-9, dash)
    .replace(/-+/g, '-')        // Multiple dashes → single dash
    .replace(/^-|-$/g, '')      // Remove leading/trailing dash
    .substring(0, 20) || 'DOC'  // Max 20 chars, fallback to DOC
}

// Generate reference number automatically
const generateReferenceNumber = async (): Promise<string> => {
  // Get first doc type or use default
  const docType = formState.value.docType && formState.value.docType.length > 0 
    ? formState.value.docType[0] 
    : 'DOC'
  
  const normalizedType = normalizeDocType(docType)
  const year = dayjs().format('YYYY')
  const month = dayjs().format('MM')
  const prefix = `${normalizedType}/${year}/${month}/`
  
  try {
    // Fetch all documents to check existing reference numbers
    const allDocuments = await documentsApi.listDocuments()
    
    // Extract reference numbers from metadata and filter by prefix
    const existingReferences: string[] = []
    
    for (const doc of allDocuments) {
      if (doc.metadata) {
        let meta: Record<string, unknown> = {}
        if (typeof doc.metadata === 'string') {
          try {
            meta = JSON.parse(doc.metadata) as Record<string, unknown>
          } catch {
            continue
          }
        } else {
          meta = doc.metadata as Record<string, unknown>
        }
        
        const ref = meta.reference as string
        if (ref && typeof ref === 'string') {
          // Normalize reference untuk comparison
          const normalizedRef = normalizeReferenceForComparison(ref)
          const normalizedPrefix = normalizeReferenceForComparison(prefix)
          
          // Check if reference starts with our prefix (case-insensitive, normalized)
          if (normalizedRef.startsWith(normalizedPrefix)) {
            existingReferences.push(ref)
          }
        }
      }
    }
    
    // Find the highest sequence number
    let maxSequence = 0
    for (const ref of existingReferences) {
      // Extract sequence number (last 3 digits after last slash)
      const parts = ref.split('/')
      if (parts.length > 0) {
        const lastPart = parts[parts.length - 1]
        if (lastPart && typeof lastPart === 'string') {
          const sequence = parseInt(lastPart, 10)
          if (!isNaN(sequence) && sequence > maxSequence) {
            maxSequence = sequence
          }
        }
      }
    }
    
    // Generate next sequence number (3 digits, padded with zeros)
    const nextSequence = String(maxSequence + 1).padStart(3, '0')
    return `${prefix}${nextSequence}`
    
  } catch (error) {
    console.error('Error generating reference number:', error)
    // Fallback: return with sequence 001 if query fails
    return `${prefix}001`
  }
}

// Normalize reference for comparison (case-insensitive, trim spaces)
const normalizeReferenceForComparison = (ref: string): string => {
  return ref
    .trim()
    .toUpperCase()
    .replace(/\s+/g, '') // Remove all spaces for comparison
}

// Check if reference number already exists in other documents
const checkReferenceExists = async (reference: string, excludeDocumentId?: string): Promise<boolean> => {
  if (!reference || !reference.trim()) {
    return false // Empty reference is considered not existing
  }
  
  try {
    // Fetch all documents to check existing reference numbers
    const allDocuments = await documentsApi.listDocuments()
    
    // Normalize input reference for comparison
    const normalizedInput = normalizeReferenceForComparison(reference)
    
    // Check if any document has the same reference (excluding current document if in edit mode)
    for (const doc of allDocuments) {
      // Skip current document if in edit mode
      if (excludeDocumentId && doc.id === excludeDocumentId) {
        continue
      }
      
      if (doc.metadata) {
        let meta: Record<string, unknown> = {}
        if (typeof doc.metadata === 'string') {
          try {
            meta = JSON.parse(doc.metadata) as Record<string, unknown>
          } catch {
            continue
          }
        } else {
          meta = doc.metadata as Record<string, unknown>
        }
        
        const existingRef = meta.reference as string
        if (existingRef && typeof existingRef === 'string') {
          // Normalize existing reference for comparison
          const normalizedExisting = normalizeReferenceForComparison(existingRef)
          
          // Check if they match (case-insensitive, space-insensitive)
          if (normalizedInput === normalizedExisting) {
            return true // Reference already exists
          }
        }
      }
    }
    
    return false // Reference is unique
  } catch (error) {
    console.error('Error checking reference uniqueness:', error)
    // If check fails, we'll let backend handle it, but return false for now
    return false
  }
}

const handleSubmit = async () => {
  // Wait for any ongoing document type creation to complete
  if (loadingDocumentTypes.value) {
    message.warning('Sedang membuat jenis dokumen, harap tunggu...')
    return
  }

  // Reload document types to ensure we have the latest data from database
  await loadDocumentTypes(false)
  
  // Validate: ensure all docType values exist in document_types table (case-insensitive)
  // If not found and user has permission, try to create them automatically
  const invalidDocTypes: string[] = []
  const docTypesToCreate: string[] = []
  
  for (const docTypeName of formState.value.docType) {
    const exists = documentTypes.value.find((dt: DocumentType) => 
      dt.name === docTypeName || dt.name.toLowerCase() === docTypeName.toLowerCase()
    )
    if (!exists) {
      // Try to create if user has permission
      if (canManageDocumentTypes.value && docTypeName.trim()) {
        docTypesToCreate.push(docTypeName.trim())
      } else {
        invalidDocTypes.push(docTypeName)
      }
    } else {
      // Update formState to use the exact name from database (in case of casing differences)
      const index = formState.value.docType.indexOf(docTypeName)
      if (index !== -1 && exists.name !== docTypeName) {
        formState.value.docType[index] = exists.name
      }
    }
  }
  
  // Try to create missing document types
  if (docTypesToCreate.length > 0) {
    message.loading('Membuat jenis dokumen yang belum ada...', 0)
    const failedCreates: string[] = []
    
    for (const docTypeName of docTypesToCreate) {
      try {
        await handleDocumentTypeCreate(docTypeName)
        // Reload after each create to ensure latest data
        await loadDocumentTypes(false)
      } catch (error) {
        console.error(`Failed to create document type "${docTypeName}":`, error)
        failedCreates.push(docTypeName)
      }
    }
    
    message.destroy() // Close loading message
    
    if (failedCreates.length > 0) {
      invalidDocTypes.push(...failedCreates)
    }
  }
  
  // Final validation after attempting to create missing types
  if (invalidDocTypes.length > 0) {
    message.error(`Jenis dokumen berikut tidak ditemukan di database: ${invalidDocTypes.join(', ')}. Silakan hapus atau buat jenis dokumen tersebut terlebih dahulu.`)
    return
  }

  // Auto-generate reference number if empty (fallback if not generated yet)
  if (!formState.value.reference?.trim()) {
    try {
      const generatedReference = await generateReferenceNumber()
      formState.value.reference = generatedReference
      console.log('Auto-generated reference number:', generatedReference)
    } catch (error) {
      console.error('Failed to generate reference number:', error)
      message.warning('Gagal generate nomor referensi otomatis, mohon isi manual')
      return
    }
  }

  // Validate reference number uniqueness
  if (formState.value.reference?.trim()) {
    const referenceExists = await checkReferenceExists(
      formState.value.reference.trim(),
      isEditMode.value ? documentId.value : undefined
    )
    
    if (referenceExists) {
      message.error(`Nomor referensi "${formState.value.reference}" sudah digunakan di dokumen lain. Silakan gunakan nomor referensi yang berbeda.`)
      return
    }
  }

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
      // Handle doc_type as string or array
      const docTypeValue = meta.doc_type
      if (Array.isArray(docTypeValue)) {
        formState.value.docType = docTypeValue as string[]
      } else if (typeof docTypeValue === 'string' && docTypeValue) {
        formState.value.docType = [docTypeValue]
      } else {
        formState.value.docType = []
      }
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

const loadDocumentTypes = async (includeInactive = false) => {
  loadingDocumentTypes.value = true
  try {
    documentTypes.value = await documentsApi.getDocumentTypes(includeInactive)
  } catch (error) {
    console.error('Failed to load document types:', error)
    message.error('Gagal memuat jenis dokumen')
  } finally {
    loadingDocumentTypes.value = false
  }
}

const handleDocumentTypeSearch = (value: string) => {
  documentTypeSearchValue.value = value
}

const handleDocumentTypeChange = async (values: string[]) => {
  // Filter out __CREATE__ values (they should not appear in the input)
  const filteredValues = values.filter((v: string) => !v.startsWith('__CREATE__'))
  
  // Validate: check if any of the selected values are inactive (and not already in document)
  const invalidValues: string[] = []
  const currentValues = formState.value.docType || []
  
  for (const value of filteredValues) {
    const docType = documentTypes.value.find((dt: DocumentType) => dt.name === value)
    if (docType && !docType.is_active) {
      // Only allow inactive types if they were already selected (edit mode)
      const wasAlreadySelected = currentValues.includes(value)
      if (!wasAlreadySelected) {
        invalidValues.push(value)
      }
    }
  }
  
  // Remove invalid (inactive) values that weren't already selected
  if (invalidValues.length > 0) {
    message.warning(`Jenis dokumen berikut tidak aktif dan tidak dapat dipilih: ${invalidValues.join(', ')}`)
    filteredValues.splice(0, filteredValues.length, ...filteredValues.filter(v => !invalidValues.includes(v)))
  }
  
  // Handle when new tag is added (user typed and pressed Enter)
  const newValues = filteredValues.filter((v: string) => !currentValues.includes(v))
  
  // Track which values failed to create
  const failedValues: string[] = []
  
  for (const newValue of newValues) {
    // Check if this is a new value that doesn't exist in documentTypes (case-insensitive)
    const exists = documentTypes.value.find((dt: DocumentType) => 
      dt.name === newValue || dt.name.toLowerCase() === newValue.toLowerCase()
    )
    if (!exists && canManageDocumentTypes.value && newValue.trim()) {
      // User typed a new value, create it
      try {
        await handleDocumentTypeCreate(newValue.trim())
        // After successful create, reload to get latest data
        await loadDocumentTypes(false)
        // Verify it was added to documentTypes (case-insensitive)
        const created = documentTypes.value.find((dt: DocumentType) => 
          dt.name.toLowerCase() === newValue.trim().toLowerCase()
        )
        if (!created) {
          // Creation failed silently, mark as failed
          console.error(`[DocumentUploadView] Document type "${newValue}" was not found after creation`)
          failedValues.push(newValue)
        } else {
          // Use the actual name from database (might have different casing)
          const index = filteredValues.indexOf(newValue)
          if (index !== -1) {
            filteredValues[index] = created.name
          }
        }
      } catch (error) {
        // Creation failed, mark as failed
        console.error(`[DocumentUploadView] Failed to create document type "${newValue}":`, error)
        failedValues.push(newValue)
      }
    } else if (!exists) {
      // Value doesn't exist and user can't create it, mark as failed
      failedValues.push(newValue)
    } else {
      // Value exists, use the exact name from database
      const index = filteredValues.indexOf(newValue)
      if (index !== -1 && exists.name !== newValue) {
        filteredValues[index] = exists.name
      }
    }
  }
  
  // Remove failed values from filteredValues before assigning
  const finalValues = filteredValues.filter((v: string) => !failedValues.includes(v))
  
  // IMPORTANT: Always update formState to remove failed values
  // This ensures that if create fails, the value is removed from formState
  formState.value.docType = finalValues
  
  // Show warning if any values were removed
  if (failedValues.length > 0) {
    message.warning(`Jenis dokumen berikut gagal dibuat atau tidak ditemukan: ${failedValues.join(', ')}`)
    console.warn(`[DocumentUploadView] Removed failed document types from formState:`, failedValues)
  }
  
  // Log final state for debugging
  console.log(`[DocumentUploadView] handleDocumentTypeChange completed. Final docTypes:`, formState.value.docType)
  
  // Auto-generate reference number if:
  // - Reference is empty or only whitespace
  // - At least one docType is selected
  // - Not currently generating (to avoid multiple simultaneous calls)
  if ((!formState.value.reference || !formState.value.reference.trim()) &&
      finalValues.length > 0 &&
      !isGeneratingReference.value) {
    // Trigger auto-generation in next tick to ensure docType is fully updated
    isGeneratingReference.value = true
    try {
      const generatedReference = await generateReferenceNumber()
      formState.value.reference = generatedReference
      console.log('Auto-generated reference number:', generatedReference)
    } catch (error) {
      console.error('Failed to auto-generate reference number:', error)
      // Don't show error message here, as it's auto-generation and user can still fill manually
    } finally {
      isGeneratingReference.value = false
    }
  }
}

const handleDocumentTypeSelect = async (value: string) => {
  // Check if it's a create action (using __CREATE__ prefix to distinguish from user input)
  if (value.startsWith('__CREATE__')) {
    const newName = value.replace('__CREATE__', '')
    await handleDocumentTypeCreate(newName)
    // handleDocumentTypeCreate will add the actual name to the array
    return
  }
  
  // Validate: check if the selected document type is active
  const selectedDocType = documentTypes.value.find((dt: DocumentType) => dt.name === value)
  if (selectedDocType && !selectedDocType.is_active) {
    // Only allow selecting inactive types if they are already in the document (edit mode)
    const isAlreadySelected = formState.value.docType.includes(value)
    if (!isAlreadySelected) {
      message.warning(`Jenis dokumen "${value}" tidak aktif dan tidak dapat dipilih. Silakan pilih jenis dokumen aktif lainnya.`)
      return
    }
  }
  
  // Add to array if not already exists
  if (!formState.value.docType.includes(value)) {
    formState.value.docType.push(value)
    
    // Auto-generate reference number if reference is empty (similar to handleDocumentTypeChange)
    // Note: handleDocumentTypeChange will also be triggered by the @change event,
    // but we call it here for immediate feedback
    if ((!formState.value.reference || !formState.value.reference.trim()) &&
        !isGeneratingReference.value) {
      isGeneratingReference.value = true
      // Use nextTick to ensure the docType array is updated first
      setTimeout(async () => {
        try {
          const generatedReference = await generateReferenceNumber()
          formState.value.reference = generatedReference
          console.log('Auto-generated reference number (from select):', generatedReference)
        } catch (error) {
          console.error('Failed to auto-generate reference number:', error)
        } finally {
          isGeneratingReference.value = false
        }
      }, 50) // Reduced delay for better UX
    }
  }
  documentTypeSearchValue.value = ''
}

const handleDocumentTypeCreate = async (value: string) => {
  if (!canManageDocumentTypes.value) {
    message.warning('Hanya superadmin dan administrator yang dapat membuat jenis dokumen baru')
    return
  }

  const trimmedValue = value.trim()
  if (!trimmedValue) {
    message.warning('Nama jenis dokumen tidak boleh kosong')
    return
  }

  // Check if already exists
  const existing = documentTypes.value.find(
    (dt: DocumentType) => dt.name.toLowerCase() === trimmedValue.toLowerCase()
  )
  if (existing) {
    message.warning(`Jenis dokumen "${trimmedValue}" sudah ada`)
    // Add to array if not already exists
    if (!formState.value.docType.includes(existing.name)) {
      formState.value.docType.push(existing.name)
    }
    return
  }

  try {
    loadingDocumentTypes.value = true
    console.log(`[DocumentUploadView] Creating document type: "${trimmedValue}"`)
    const newDocType = await documentsApi.createDocumentType(trimmedValue)
    console.log(`[DocumentUploadView] Document type created:`, newDocType)
    
    // Verify response is valid - MUST have id to confirm it's saved to document_types table
    if (!newDocType || !newDocType.id) {
      console.error(`[DocumentUploadView] Invalid response: document type not created in database`, newDocType)
      throw new Error('Invalid response from server: document type not created in database')
    }
    
    // Reload document types to ensure we have the latest from database
    // This ensures the new document type is available for validation
    await loadDocumentTypes(false)
    
    // Verify it was actually saved by checking if it exists in documentTypes
    const verified = documentTypes.value.find((dt: DocumentType) => dt.id === newDocType.id || dt.name.toLowerCase() === trimmedValue.toLowerCase())
    if (!verified) {
      console.error(`[DocumentUploadView] Document type not found after creation:`, newDocType.id)
      throw new Error('Document type was not saved to database')
    }
    
    // Use the verified document type (might have different name casing)
    const finalDocType = verified
    
    // Add to array if not already exists
    if (!formState.value.docType.includes(finalDocType.name)) {
      formState.value.docType.push(finalDocType.name)
    }
    message.success(`Jenis dokumen "${trimmedValue}" berhasil dibuat dan disimpan ke database`)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal membuat jenis dokumen'
    console.error(`[DocumentUploadView] Failed to create document type:`, error)
    message.error(errorMessage)
    // Re-throw error so caller knows creation failed
    throw error
  } finally {
    loadingDocumentTypes.value = false
    documentTypeSearchValue.value = ''
  }
}

const handleDocumentTypeDelete = async (id: string, name: string) => {
  if (!canManageDocumentTypes.value) {
    message.warning('Hanya superadmin dan administrator yang dapat menghapus jenis dokumen')
    return
  }

  try {
    await documentsApi.deleteDocumentType(id)
    documentTypes.value = documentTypes.value.filter((dt: DocumentType) => dt.id !== id)
    // Remove from selected array if exists
    formState.value.docType = formState.value.docType.filter((dt: string) => dt !== name)
    message.success(`Jenis dokumen "${name}" berhasil dihapus`)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal menghapus jenis dokumen'
    message.error(errorMessage)
  }
}

// Filter document types based on search and active status
const filteredDocumentTypes = computed(() => {
  // Get currently selected document types (for edit mode, these might be inactive)
  const selectedDocTypeNames = formState.value.docType || []
  
  // Filter: show active types OR inactive types that are already selected (for edit mode)
  let filtered = documentTypes.value.filter((dt: DocumentType) => {
    // Always show active types
    if (dt.is_active) return true
    // Show inactive types only if they are already selected (for edit mode)
    return selectedDocTypeNames.includes(dt.name)
  })
  
  // Apply search filter if there's a search value
  if (documentTypeSearchValue.value) {
    const searchLower = documentTypeSearchValue.value.toLowerCase()
    filtered = filtered.filter((dt: DocumentType) => 
      dt.name.toLowerCase().includes(searchLower)
    )
  }
  
  return filtered
})

onMounted(async () => {
  await Promise.all([
    loadFolders(),
    loadDocumentTypes(false), // Load active types first
  ])
  if (isEditMode.value) {
    await loadDocument()
    // After loading document, check if any docType exists in active types
    // If not, reload with inactive types included to show the document's types
    const hasInactiveType = formState.value.docType.some((dt: string) => 
      !documentTypes.value.find((d: DocumentType) => d.name === dt)
    )
    if (hasInactiveType) {
      await loadDocumentTypes(true) // Include inactive to show the document's types
    }
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
                  <a-select 
                    v-model:value="formState.folder_id" 
                    placeholder="Pilih folder (opsional)"
                    :loading="loadingFolders" 
                    :disabled="isEditMode"
                    allow-clear>
                    <a-select-option v-for="folder in folders" :key="folder.id" :value="folder.id">
                      {{ folder.name }}
                    </a-select-option>
                  </a-select>
                  <div v-if="isEditMode" style="margin-top: 4px; font-size: 12px; color: #999;">
                    Folder tidak dapat diubah saat edit
                  </div>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Judul Dokumen" required>
                  <a-input v-model:value="formState.title" placeholder="Judul dokumen" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Nomor Dokumen / Referensi">
                  <a-input 
                    v-model:value="formState.reference" 
                    placeholder="No. referensi (otomatis generate jika kosong)" 
                  />
                  <div v-if="!isEditMode" style="margin-top: 4px; font-size: 12px; color: #666;">
                    <IconifyIcon icon="mdi:information-outline" style="margin-right: 4px;" />
                    Jika dikosongkan, nomor referensi akan otomatis di-generate berdasarkan jenis dokumen
                  </div>
                </a-form-item>
              </a-col>
              <a-col v-if="isEditMode" :xs="24" :md="12">
                <a-form-item label="Document ID">
                  <a-input v-model:value="formState.documentId" disabled />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Jenis Dokumen">
                  <a-select
                    v-model:value="formState.docType"
                    mode="tags"
                    placeholder="Pilih atau ketik jenis dokumen baru (tekan Enter untuk membuat)"
                    show-search
                    allow-clear
                    :filter-option="false"
                    :loading="loadingDocumentTypes"
                    :search-value="documentTypeSearchValue"
                    @search="handleDocumentTypeSearch"
                    @change="handleDocumentTypeChange"
                    @select="handleDocumentTypeSelect"
                    @deselect="(value: string) => {
                      formState.docType = formState.docType.filter((dt: string) => dt !== value)
                    }"
                  >
                    <a-select-option
                      v-for="docType in filteredDocumentTypes"
                      :key="docType.id"
                      :value="docType.name"
                      :disabled="!docType.is_active && !formState.docType.includes(docType.name)"
                    >
                      <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
                        <span :style="{ opacity: docType.is_active ? 1 : 0.6 }">
                          {{ docType.name }}
                          <a-tag v-if="!docType.is_active" color="default" size="small" style="margin-left: 8px;">
                            Tidak Aktif
                          </a-tag>
                        </span>
                        <IconifyIcon
                          v-if="canManageDocumentTypes && docType.is_active"
                          icon="mdi:delete-outline"
                          style="margin-left: 8px; cursor: pointer; color: #ff4d4f;"
                          @click.stop="handleDocumentTypeDelete(docType.id, docType.name)"
                          :title="docType.usage_count > 0 ? `Digunakan oleh ${docType.usage_count} dokumen (soft delete)` : 'Hapus jenis dokumen'"
                        />
                      </div>
                    </a-select-option>
                    <a-select-option
                      v-if="documentTypeSearchValue && !filteredDocumentTypes.find((dt: DocumentType) => dt.name.toLowerCase() === documentTypeSearchValue.toLowerCase()) && canManageDocumentTypes"
                      :value="`__CREATE__${documentTypeSearchValue}`"
                      style="color: #1890ff;"
                    >
                      <IconifyIcon icon="mdi:plus-circle" style="margin-right: 4px;" />
                      Buat "{{ documentTypeSearchValue }}"
                    </a-select-option>
                  </a-select>
                  <div v-if="canManageDocumentTypes" style="margin-top: 4px; font-size: 12px; color: #1890ff;">
                    <IconifyIcon icon="mdi:information-outline" style="margin-right: 4px;" />
                    Ketik nama baru dan tekan Enter untuk membuat jenis dokumen baru. Klik icon hapus di dropdown untuk menghapus jenis dokumen.
                  </div>
                </a-form-item>
              </a-col>
              <!-- <a-col :xs="24" :md="12">
                <a-form-item label="Unit / Divisi">
                  <a-input v-model:value="formState.unit" />
                </a-form-item>
              </a-col> -->
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

