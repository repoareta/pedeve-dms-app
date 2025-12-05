import apiClient from './client'
import axios from 'axios'

export interface DocumentFolder {
  id: string
  name: string
  parent_id?: string | null
  created_at?: string
  updated_at?: string
}

export interface DocumentItem {
  id: string
  folder_id?: string | null
  name: string
  file_name: string
  file_path: string
  mime_type: string
  size: number
  status: string
  metadata?: Record<string, unknown>
  uploader_id?: string
  uploader_name?: string
  created_at?: string
  updated_at?: string
}

export interface DocumentListResponse {
  data: DocumentItem[]
  total: number
  page: number
  page_size: number
}

export interface DocumentFolderStat {
  folder_id: string | null
  file_count: number
  total_size: number
}

export interface DocumentSummary {
  folder_stats: DocumentFolderStat[]
  total_size: number
}

export interface DocumentType {
  id: string
  name: string
  is_active: boolean
  usage_count: number
  created_by: string
  created_at?: string
  updated_at?: string
}

const documentsApi = {
  async listFolders(): Promise<DocumentFolder[]> {
    const res = await apiClient.get<DocumentFolder[]>('/documents/folders')
    return res.data
  },

  async createFolder(name: string, parent_id?: string): Promise<DocumentFolder> {
    const res = await apiClient.post<DocumentFolder>('/documents/folders', {
      name,
      parent_id,
    })
    return res.data
  },

  async renameFolder(id: string, name: string): Promise<DocumentFolder> {
    const res = await apiClient.put<DocumentFolder>(`/documents/folders/${id}`, { name })
    return res.data
  },

  async deleteFolder(id: string): Promise<void> {
    await apiClient.delete(`/documents/folders/${id}`)
  },

  async listDocumentsPaginated(params: {
    page?: number
    page_size?: number
    search?: string
    sort_by?: string
    sort_dir?: string
    type?: string
    folder_id?: string
  } = {}): Promise<DocumentListResponse> {
    const res = await apiClient.get<DocumentListResponse>('/documents', { params })
    return res.data
  },

  async listDocuments(params?: { folder_id?: string }): Promise<DocumentItem[]> {
    const res = await apiClient.get<DocumentItem[]>('/documents', { params })
    return res.data
  },

  async getDocument(id: string): Promise<DocumentItem> {
    const res = await apiClient.get<DocumentItem>(`/documents/${id}`)
    return res.data
  },

  async uploadDocument(payload: {
    file: File
    folder_id?: string
    title?: string
    status?: string
    metadata?: Record<string, unknown>
  }): Promise<DocumentItem> {
    const formData = new FormData()
    formData.append('file', payload.file)
    if (payload.folder_id) formData.append('folder_id', payload.folder_id)
    if (payload.title) formData.append('title', payload.title)
    if (payload.status) formData.append('status', payload.status)
    if (payload.metadata) formData.append('metadata', JSON.stringify(payload.metadata))

    try {
      const res = await apiClient.post<DocumentItem>('/documents/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return res.data
    } catch (error: unknown) {
      if (axios.isAxiosError(error) && error.response?.status === 413) {
        throw new Error('Ukuran file terlalu besar untuk diupload. Silakan pilih file yang lebih kecil.')
      }
      throw error
    }
  },

  async updateDocument(
    id: string,
    payload:
      | {
          folder_id?: string
          title?: string
          status?: string
          metadata?: Record<string, unknown>
        }
      | {
          file: File
          folder_id?: string
          title?: string
          status?: string
          metadata?: Record<string, unknown>
        }
  ): Promise<DocumentItem> {
    // Jika ada file, kirim multipart
    if ('file' in payload) {
      const formData = new FormData()
      formData.append('file', payload.file)
      if (payload.folder_id) formData.append('folder_id', payload.folder_id)
      if (payload.title) formData.append('title', payload.title)
      if (payload.status) formData.append('status', payload.status)
      if (payload.metadata) formData.append('metadata', JSON.stringify(payload.metadata))
      const res = await apiClient.put<DocumentItem>(`/documents/${id}`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return res.data
    }

    // JSON update (metadata saja)
    const res = await apiClient.put<DocumentItem>(`/documents/${id}`, payload)
    return res.data
  },

  async deleteDocument(id: string): Promise<void> {
    await apiClient.delete(`/documents/${id}`)
  },

  async getDocumentSummary(): Promise<DocumentSummary> {
    const res = await apiClient.get<DocumentSummary>('/documents/summary')
    return res.data
  },

  // Document Types API
  async getDocumentTypes(includeInactive = false): Promise<DocumentType[]> {
    const res = await apiClient.get<DocumentType[]>('/document-types', {
      params: { include_inactive: includeInactive },
    })
    return res.data
  },

  async createDocumentType(name: string): Promise<DocumentType> {
    const res = await apiClient.post<DocumentType>('/document-types', { name })
    return res.data
  },

  async updateDocumentType(id: string, payload: { name?: string; is_active?: boolean }): Promise<DocumentType> {
    const res = await apiClient.put<DocumentType>(`/document-types/${id}`, payload)
    return res.data
  },

  async deleteDocumentType(id: string): Promise<void> {
    await apiClient.delete(`/document-types/${id}`)
  },
}

export default documentsApi
