import apiClient from './client'

export interface AuditLog {
  id: string
  user_id: string
  username: string
  action: string
  resource: string
  resource_id: string
  ip_address: string
  user_agent: string
  details: string
  status: string
  created_at: string
}

export interface AuditLogsResponse {
  data: AuditLog[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface AuditLogsParams {
  page?: number
  pageSize?: number
  action?: string
  resource?: string
  status?: string
}

export const auditApi = {
  getAuditLogs: async (params?: AuditLogsParams): Promise<AuditLogsResponse> => {
    const queryParams = new URLSearchParams()
    
    if (params?.page) {
      queryParams.append('page', params.page.toString())
    }
    if (params?.pageSize) {
      queryParams.append('pageSize', params.pageSize.toString())
    }
    if (params?.action) {
      queryParams.append('action', params.action)
    }
    if (params?.resource) {
      queryParams.append('resource', params.resource)
    }
    if (params?.status) {
      queryParams.append('status', params.status)
    }

    const queryString = queryParams.toString()
    const url = queryString ? `/audit-logs?${queryString}` : '/audit-logs'
    
    const response = await apiClient.get<AuditLogsResponse>(url)
    return response.data
  },
}

