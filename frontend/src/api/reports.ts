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
}

export default reportsApi

