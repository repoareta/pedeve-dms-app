import apiClient from './client'

export interface FinancialReport {
  id: string
  company_id: string
  year: string
  period: string
  is_rkap: boolean
  inputter_id?: string
  
  // Neraca
  current_assets: number
  non_current_assets: number
  short_term_liabilities: number
  long_term_liabilities: number
  equity: number
  
  // Laba Rugi
  revenue: number
  operating_expenses: number
  operating_profit: number
  other_income: number
  tax: number
  net_profit: number
  
  // Cashflow
  operating_cashflow: number
  investing_cashflow: number
  financing_cashflow: number
  ending_balance: number
  
  // Rasio
  roe: number
  roi: number
  current_ratio: number
  cash_ratio: number
  ebitda: number
  ebitda_margin: number
  net_profit_margin: number
  operating_profit_margin: number
  debt_to_equity: number
  
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

export interface CreateFinancialReportRequest {
  company_id: string
  year: string
  period: string
  is_rkap: boolean
  
  // Neraca
  current_assets: number
  non_current_assets: number
  short_term_liabilities: number
  long_term_liabilities: number
  equity: number
  
  // Laba Rugi
  revenue: number
  operating_expenses: number
  operating_profit: number
  other_income: number
  tax: number
  net_profit: number
  
  // Cashflow
  operating_cashflow: number
  investing_cashflow: number
  financing_cashflow: number
  ending_balance: number
  
  // Rasio
  roe: number
  roi: number
  current_ratio: number
  cash_ratio: number
  ebitda: number
  ebitda_margin: number
  net_profit_margin: number
  operating_profit_margin: number
  debt_to_equity: number
  
  remark?: string
}

export interface UpdateFinancialReportRequest {
  year?: string
  period?: string
  is_rkap?: boolean
  
  // Neraca
  current_assets?: number
  non_current_assets?: number
  short_term_liabilities?: number
  long_term_liabilities?: number
  equity?: number
  
  // Laba Rugi
  revenue?: number
  operating_expenses?: number
  operating_profit?: number
  other_income?: number
  tax?: number
  net_profit?: number
  
  // Cashflow
  operating_cashflow?: number
  investing_cashflow?: number
  financing_cashflow?: number
  ending_balance?: number
  
  // Rasio
  roe?: number
  roi?: number
  current_ratio?: number
  cash_ratio?: number
  ebitda?: number
  ebitda_margin?: number
  net_profit_margin?: number
  operating_profit_margin?: number
  debt_to_equity?: number
  
  remark?: string
}

export interface ComparisonItem {
  rkap: number | string
  realisasi_ytd: number | string
  difference: number | string
  percentage: number
}

export interface FinancialReportComparison {
  company_id: string
  year: string
  month: string
  rkap?: FinancialReport
  realisasi_ytd?: FinancialReport
  comparison: Record<string, ComparisonItem>
}

export const financialReportsApi = {
  // Create financial report (RKAP or Realisasi)
  async create(data: CreateFinancialReportRequest): Promise<FinancialReport> {
    const response = await apiClient.post<FinancialReport>('/financial-reports', data)
    return response.data
  },

  // Get financial report by ID
  async getById(id: string): Promise<FinancialReport> {
    const response = await apiClient.get<FinancialReport>(`/financial-reports/${id}`)
    return response.data
  },

  // Get all financial reports for a company
  async getByCompanyId(companyId: string): Promise<FinancialReport[]> {
    const response = await apiClient.get<FinancialReport[]>(`/financial-reports/company/${companyId}`)
    return response.data
  },

  // Get list of years that have RKAP for a company
  async getRKAPYears(companyId: string): Promise<string[]> {
    const response = await apiClient.get<string[]>(`/financial-reports/rkap-years/${companyId}`)
    return response.data
  },

  // Get RKAP for a company and year
  async getRKAP(companyId: string, year: string): Promise<FinancialReport | null> {
    try {
      const reports = await this.getByCompanyId(companyId)
      const rkap = reports.find(r => r.is_rkap && r.year === year)
      return rkap || null
    } catch {
      return null
    }
  },

  // Get Realisasi for a company and period
  async getRealisasi(companyId: string, period: string): Promise<FinancialReport | null> {
    try {
      const reports = await this.getByCompanyId(companyId)
      const realisasi = reports.find(r => !r.is_rkap && r.period === period)
      return realisasi || null
    } catch {
      return null
    }
  },

  // Get comparison between RKAP and Realisasi YTD
  async getComparison(companyId: string, year: string, month: string): Promise<FinancialReportComparison> {
    const response = await apiClient.get<FinancialReportComparison>('/financial-reports/compare', {
      params: {
        company_id: companyId,
        year,
        month,
      },
    })
    return response.data
  },

  // Update financial report
  async update(id: string, data: UpdateFinancialReportRequest): Promise<FinancialReport> {
    const response = await apiClient.put<FinancialReport>(`/financial-reports/${id}`, data)
    return response.data
  },

  // Delete financial report
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/financial-reports/${id}`)
  },

  // Export performance data to Excel (with charts and tables)
  async exportPerformanceExcel(
    companyId: string,
    startPeriod: string,
    endPeriod: string
  ): Promise<Blob> {
    const response = await apiClient.get(
      `/companies/${companyId}/performance/export/excel`,
      {
        params: {
          start_period: startPeriod,
          end_period: endPeriod,
        },
        responseType: 'blob',
      }
    )
    return response.data
  },

  // Download bulk upload template Excel
  async downloadBulkUploadTemplate(params?: {
    period?: string
    is_rkap?: boolean
  }): Promise<Blob> {
    try {
      const response = await apiClient.get('/financial-reports/bulk-upload/template', {
        params,
        responseType: 'blob',
      })
      
      // Check if response is actually a blob (Excel file)
      if (!(response.data instanceof Blob)) {
        throw new Error('Response is not a valid file')
      }
      
      // Check content type - if it's JSON, it means there's an error
      const contentType = response.headers['content-type'] || ''
      if (contentType.includes('application/json')) {
        // Response is JSON error, parse it
        const text = await (response.data as Blob).text()
        const json = JSON.parse(text)
        throw new Error(json.message || json.error || 'Failed to download template')
      }
      
      return response.data
    } catch (error: unknown) {
      // If it's already an Error with message, re-throw it
      if (error instanceof Error) {
        throw error
      }
      
      // Otherwise, wrap it
      const axiosError = error as {
        response?: {
          status?: number
          data?: unknown
          headers?: { 'content-type'?: string }
        }
        message?: string
      }
      
      if (axiosError.response?.status === 404) {
        throw new Error('Template endpoint not found. Pastikan server sudah di-restart setelah perubahan route.')
      }
      
      throw new Error(axiosError.message || 'Failed to download template')
    }
  },

  // Validate bulk upload Excel file before upload
  async validateBulkExcelFile(file: File): Promise<{
    valid: boolean
    errors: Array<{ row: number; column: string; message: string }>
    data: Array<Record<string, unknown>>
  }> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post('/financial-reports/bulk-upload/validate', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  // Upload bulk financial reports from Excel file
  async uploadBulkFinancialReports(
    file: File,
    onProgress?: (progress: number) => void
  ): Promise<{
    success: number
    failed: number
    created: number
    updated: number
    errors: Array<{ row: number; column?: string; message: string }>
    message?: string
  }> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post('/financial-reports/bulk-upload', formData, {
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
}
