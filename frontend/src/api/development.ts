import apiClient from './client'

export interface SeederStatusResponse {
  exists: boolean
  message: string
}

export interface ResetSubsidiaryResponse {
  message: string
  success: boolean
}

export interface RunSeederResponse {
  message: string
  success: boolean
}

const developmentApi = {
  // Check seeder status
  async checkSeederStatus(): Promise<SeederStatusResponse> {
    const response = await apiClient.get<SeederStatusResponse>(
      '/development/check-seeder-status'
    )
    return response.data
  },

  // Reset subsidiary data
  async resetSubsidiaryData(): Promise<ResetSubsidiaryResponse> {
    const response = await apiClient.post<ResetSubsidiaryResponse>(
      '/development/reset-subsidiary',
      {}
    )
    return response.data
  },

  // Run subsidiary seeder
  async runSubsidiarySeeder(): Promise<RunSeederResponse> {
    const response = await apiClient.post<RunSeederResponse>(
      '/development/run-subsidiary-seeder',
      {}
    )
    return response.data
  },

  // Check report status
  async checkReportStatus(): Promise<SeederStatusResponse> {
    const response = await apiClient.get<SeederStatusResponse>(
      '/development/check-report-status'
    )
    return response.data
  },

  // Reset report data
  async resetReportData(): Promise<ResetSubsidiaryResponse> {
    const response = await apiClient.post<ResetSubsidiaryResponse>(
      '/development/reset-reports',
      {}
    )
    return response.data
  },

  // Run report seeder
  async runReportSeeder(): Promise<RunSeederResponse> {
    const response = await apiClient.post<RunSeederResponse>(
      '/development/run-report-seeder',
      {}
    )
    return response.data
  },

  // Run all seeders (company + reports)
  async runAllSeeders(): Promise<RunSeederResponse & { details?: Record<string, string> }> {
    const response = await apiClient.post<RunSeederResponse & { details?: Record<string, string> }>(
      '/development/run-all-seeders',
      {}
    )
    return response.data
  },

  // Reset all seeded data (reports + company)
  async resetAllSeededData(): Promise<ResetSubsidiaryResponse & { details?: Record<string, string> }> {
    const response = await apiClient.post<ResetSubsidiaryResponse & { details?: Record<string, string> }>(
      '/development/reset-all-seeded-data',
      {}
    )
    return response.data
  },

  // Check all seeder status
  async checkAllSeederStatus(): Promise<{ status: Record<string, boolean>; message: string }> {
    const response = await apiClient.get<{ status: Record<string, boolean>; message: string }>(
      '/development/check-all-seeder-status'
    )
    return response.data
  },

  // Reset all financial reports
  async resetAllFinancialReports(): Promise<ResetSubsidiaryResponse> {
    const response = await apiClient.post<ResetSubsidiaryResponse>(
      '/development/reset-all-financial-reports',
      {}
    )
    return response.data
  },

  // Check expiring documents
  async checkExpiringDocuments(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; documents_found: number; notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; documents_found: number; notifications_created: number }>(
      '/development/check-expiring-documents',
      { threshold_days: thresholdDays }
    )
    return response.data
  },

  // Check expiring director terms
  async checkExpiringDirectorTerms(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; directors_found: number; notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; directors_found: number; notifications_created: number }>(
      '/development/check-expiring-director-terms',
      { threshold_days: thresholdDays }
    )
    return response.data
  },

  // Check all expiring notifications
  async checkAllExpiringNotifications(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; documents: { found: number; notifications_created: number }; directors: { found: number; notifications_created: number }; total_notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; documents: { found: number; notifications_created: number }; directors: { found: number; notifications_created: number }; total_notifications_created: number }>(
      '/development/check-all-expiring-notifications',
      { threshold_days: thresholdDays }
    )
    return response.data
  },
}

export default developmentApi

