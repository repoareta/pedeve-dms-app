import apiClient from './client'

export interface Report {
  id: string
  period: string
  report_date: string
  company_id: string
  inputter_id?: string
  revenue: number
  opex: number
  npat: number
  dividend: number
  financial_ratio: number
  attachment?: string
  remark?: string
  created_at: string
  updated_at: string
  company?: {
    id: string
    name: string
    code: string
  }
  inputter?: {
    id: string
    username: string
    email: string
  }
}

export interface CreateReportRequest {
  period: string
  company_id: string
  inputter_id?: string
  revenue: number
  opex: number
  npat: number
  dividend: number
  financial_ratio: number
  attachment?: string
  remark?: string
}

export interface UpdateReportRequest {
  period?: string
  company_id?: string
  inputter_id?: string
  revenue?: number
  opex?: number
  npat?: number
  dividend?: number
  financial_ratio?: number
  attachment?: string
  remark?: string
}

export interface ReportsResponse {
  data: Report[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type ValidationRow = {
  period?: string
  company_code?: string
  revenue?: number | string
  opex?: number | string
  npat?: number | string
  dividend?: number | string
  financial_ratio?: number | string
  remark?: string
  [key: string]: unknown
}

const reportsApi = {
  // Get all reports with pagination and filters
  // company_id can be a single ID or comma-separated string for multiple IDs
  async getAll(params?: {
    company_id?: string
    period?: string
    page?: number
    page_size?: number
  }): Promise<ReportsResponse> {
    const response = await apiClient.get<ReportsResponse>('/reports', { params })
    return response.data
  },

  // Get report by ID
  async getById(id: string): Promise<Report> {
    const response = await apiClient.get<Report>(`/reports/${id}`)
    return response.data
  },

  // Get reports by company ID
  async getByCompanyId(companyId: string): Promise<Report[]> {
    const response = await apiClient.get<Report[]>(`/reports/company/${companyId}`)
    return response.data
  },

  // Create report
  async create(data: CreateReportRequest): Promise<Report> {
    const response = await apiClient.post<Report>('/reports', data)
    return response.data
  },

  // Update report
  async update(id: string, data: UpdateReportRequest): Promise<Report> {
    const response = await apiClient.put<Report>(`/reports/${id}`, data)
    return response.data
  },

  // Delete report
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/reports/${id}`)
  },

  // Export reports to Excel
  async exportExcel(params?: {
    company_id?: string
    period?: string
  }): Promise<Blob> {
    const response = await apiClient.get('/reports/export/excel', {
      params,
      responseType: 'blob',
    })
    return response.data
  },

  // Export reports to PDF
  async exportPDF(params?: {
    company_id?: string
    period?: string
  }): Promise<Blob> {
    const response = await apiClient.get('/reports/export/pdf', {
      params,
      responseType: 'blob',
    })
    return response.data
  },

  // Upload reports from Excel file
  async uploadReports(file: File, onProgress?: (progress: number) => void): Promise<{ success: number; failed: number; errors: Array<{ row: number; message: string }> }> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post('/reports/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        } else if (onProgress && progressEvent.loaded) {
          // Fallback jika total tidak tersedia
          onProgress(Math.min(99, Math.round(progressEvent.loaded / 1024))) // Estimate based on KB
        }
      },
    })
    return response.data
  },

  // Validate Excel file before upload
  async validateExcelFile(file: File): Promise<{ valid: boolean; errors: Array<{ row: number; column: string; message: string }>; data: ValidationRow[] }> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post('/reports/validate', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  // Download template Excel
  async downloadTemplate(): Promise<Blob> {
    const response = await apiClient.get('/reports/template', {
      responseType: 'blob',
      validateStatus: (status) => status < 500, // Don't throw on 4xx
    })
    
    if (response.status === 404) {
      throw new Error('Template endpoint not found')
    }
    
    if (response.status >= 400) {
      throw new Error(`Failed to download template: ${response.status}`)
    }
    
    return response.data
  },
}

export default reportsApi
