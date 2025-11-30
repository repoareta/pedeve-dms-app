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
}

export default developmentApi

