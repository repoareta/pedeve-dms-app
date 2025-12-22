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
  // Cek status seeder
  async checkSeederStatus(): Promise<SeederStatusResponse> {
    const response = await apiClient.get<SeederStatusResponse>(
      '/development/check-seeder-status'
    )
    return response.data
  },

  // Reset data subsidiary
  async resetSubsidiaryData(): Promise<ResetSubsidiaryResponse> {
    const response = await apiClient.post<ResetSubsidiaryResponse>(
      '/development/reset-subsidiary',
      {}
    )
    return response.data
  },

  // Jalankan subsidiary seeder
  async runSubsidiarySeeder(): Promise<RunSeederResponse> {
    const response = await apiClient.post<RunSeederResponse>(
      '/development/run-subsidiary-seeder',
      {}
    )
    return response.data
  },

  // Cek status report
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

  // Jalankan report seeder
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

  // Reset semua seeded data (reports + company)
  async resetAllSeededData(): Promise<ResetSubsidiaryResponse & { details?: Record<string, string> }> {
    const response = await apiClient.post<ResetSubsidiaryResponse & { details?: Record<string, string> }>(
      '/development/reset-all-seeded-data',
      {}
    )
    return response.data
  },

  // Cek status semua seeder
  async checkAllSeederStatus(): Promise<{ status: Record<string, boolean>; message: string }> {
    const response = await apiClient.get<{ status: Record<string, boolean>; message: string }>(
      '/development/check-all-seeder-status'
    )
    return response.data
  },

  // Reset semua financial reports
  async resetAllFinancialReports(): Promise<ResetSubsidiaryResponse> {
    const response = await apiClient.post<ResetSubsidiaryResponse>(
      '/development/reset-all-financial-reports',
      {}
    )
    return response.data
  },

  // Cek dokumen yang akan expired
  async checkExpiringDocuments(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; documents_found: number; notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; documents_found: number; notifications_created: number }>(
      '/development/check-expiring-documents',
      { threshold_days: thresholdDays }
    )
    return response.data
  },

  // Cek masa jabatan direktur yang akan expired
  async checkExpiringDirectorTerms(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; directors_found: number; notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; directors_found: number; notifications_created: number }>(
      '/development/check-expiring-director-terms',
      { threshold_days: thresholdDays }
    )
    return response.data
  },

  // Cek semua notifikasi yang akan expired
  async checkAllExpiringNotifications(thresholdDays: number = 30): Promise<{ message: string; threshold_days: number; documents: { found: number; notifications_created: number }; directors: { found: number; notifications_created: number }; total_notifications_created: number }> {
    const response = await apiClient.post<{ message: string; threshold_days: number; documents: { found: number; notifications_created: number }; directors: { found: number; notifications_created: number }; total_notifications_created: number }>(
      '/development/check-all-expiring-notifications',
      { threshold_days: thresholdDays }
    )
    return response.data
  },
}

export default developmentApi

