import apiClient from './client'

export interface SonarQubeIssue {
  key: string
  rule: string
  severity: string
  component: string
  project: string
  line: number
  message: string
  type: string
  status: string
  effort?: string
  debt?: string
  author?: string
  creationDate?: string
  updateDate?: string
}

export interface SonarQubeComponent {
  key: string
  name: string
  qualifier: string
  path?: string
}

export interface SonarQubeIssuesResponse {
  total: number
  p: number
  ps: number
  paging: {
    pageIndex: number
    pageSize: number
    total: number
  }
  issues: SonarQubeIssue[]
  components: SonarQubeComponent[]
}

export interface SonarQubeIssuesParams {
  severities?: string[]
  types?: string[]
  statuses?: string[]
}

export interface SonarQubeStatus {
  enabled: boolean
}

export const sonarqubeApi = {
  /**
   * Check if SonarQube Monitor feature is enabled
   */
  async getStatus(): Promise<SonarQubeStatus> {
    try {
      const response = await apiClient.get<SonarQubeStatus>('/sonarqube/status')
      return response.data
    } catch {
      // If endpoint doesn't exist or returns error, feature is disabled
      return { enabled: false }
    }
  },

  /**
   * Get SonarCloud issues
   */
  async getIssues(params?: SonarQubeIssuesParams): Promise<SonarQubeIssuesResponse> {
    const queryParams = new URLSearchParams()
    
    if (params?.severities && params.severities.length > 0) {
      params.severities.forEach(s => queryParams.append('severities', s))
    }
    
    if (params?.types && params.types.length > 0) {
      params.types.forEach(t => queryParams.append('types', t))
    }
    
    if (params?.statuses && params.statuses.length > 0) {
      params.statuses.forEach(s => queryParams.append('statuses', s))
    }

    const queryString = queryParams.toString()
    const url = `/sonarqube/issues${queryString ? `?${queryString}` : ''}`
    
    const response = await apiClient.get<SonarQubeIssuesResponse>(url)
    return response.data
  },

  /**
   * Export SonarCloud issues as JSON
   */
  async exportIssues(params?: SonarQubeIssuesParams): Promise<Blob> {
    const queryParams = new URLSearchParams()
    
    if (params?.severities && params.severities.length > 0) {
      params.severities.forEach(s => queryParams.append('severities', s))
    }
    
    if (params?.types && params.types.length > 0) {
      params.types.forEach(t => queryParams.append('types', t))
    }
    
    if (params?.statuses && params.statuses.length > 0) {
      params.statuses.forEach(s => queryParams.append('statuses', s))
    }

    const queryString = queryParams.toString()
    const url = `/sonarqube/issues/export${queryString ? `?${queryString}` : ''}`
    
    const response = await apiClient.get<Blob>(url, {
      responseType: 'blob',
    })
    
    return response.data
  },
}

