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
  log_type: string // "user_action" atau "technical_error"
  created_at: string
}

// UserActivityLog - permanent logs untuk data penting (report, document, company, user)
export interface UserActivityLog {
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
  // Note: Tidak ada log_type karena semua adalah user_action permanent
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
  logType?: string // "user_action" or "technical_error"
}

export interface AuditLogStats {
  total_records: number
  user_action_count: number
  technical_error_count: number
  oldest_record_date?: string
  newest_record_date?: string
  estimated_size_mb: number
  retention_policy?: {
    user_action_days: number
    technical_error_days: number
  }
}

export interface UserActivityLogsResponse {
  data: UserActivityLog[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface UserActivityLogsParams {
  page?: number
  pageSize?: number
  action?: string
  resource?: string
  resource_id?: string
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
    if (params?.logType) {
      queryParams.append('logType', params.logType)
    }

    const queryString = queryParams.toString()
    const url = queryString ? `/audit-logs?${queryString}` : '/audit-logs'
    
    const response = await apiClient.get<AuditLogsResponse>(url)
    return response.data
  },
  getUserActivityLogs: async (params?: UserActivityLogsParams): Promise<UserActivityLogsResponse> => {
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
    if (params?.resource_id) {
      queryParams.append('resource_id', params.resource_id)
    }
    if (params?.status) {
      queryParams.append('status', params.status)
    }

    const queryString = queryParams.toString()
    const url = queryString ? `/user-activity-logs?${queryString}` : '/user-activity-logs'
    
    const response = await apiClient.get<UserActivityLogsResponse>(url)
    return response.data
  },
  getAuditLogStats: async (): Promise<AuditLogStats> => {
    const response = await apiClient.get<AuditLogStats>('/audit-logs/stats')
    return response.data
  },
}

