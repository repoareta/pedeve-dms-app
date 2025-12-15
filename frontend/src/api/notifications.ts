import apiClient from './client'

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  message: string
  resource_type?: string
  resource_id?: string
  is_read: boolean
  created_at: string
  read_at?: string
  document?: {
    id: string
    name: string
    file_name: string
    expiry_date?: string
  }
}

export interface NotificationsResponse {
  data: Notification[]
}

export interface NotificationsInboxResponse {
  data: Notification[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface UnreadCountResponse {
  count: number
}

export interface NotificationFilters {
  unread_only?: boolean
  days_until_expiry?: number // 3, 7, 30
  page?: number
  page_size?: number
}

export const notificationApi = {
  // Get notifications (untuk dropdown bell)
  // PENTING: Endpoint ini sudah menggunakan RBAC di backend
  // - Superadmin/Administrator: melihat semua notifikasi
  // - Admin: melihat notifikasi dari company mereka + descendants  
  // - Regular users: hanya melihat notifikasi mereka sendiri
  getNotifications: async (unreadOnly = false, limit = 5): Promise<Notification[]> => {
    const response = await apiClient.get<NotificationsResponse>(
      `/notifications?unread_only=${unreadOnly}&limit=${limit}`
    )
    return response.data.data
  },

  // Get notifications dengan filters (untuk inbox page)
  getNotificationsInbox: async (filters?: NotificationFilters): Promise<NotificationsInboxResponse> => {
    const params = new URLSearchParams()
    
    if (filters?.unread_only !== undefined) {
      params.append('unread_only', filters.unread_only.toString())
    }
    if (filters?.days_until_expiry) {
      params.append('days_until_expiry', filters.days_until_expiry.toString())
    }
    if (filters?.page) {
      params.append('page', filters.page.toString())
    }
    if (filters?.page_size) {
      params.append('page_size', filters.page_size.toString())
    }

    const queryString = params.toString()
    const url = queryString ? `/notifications/inbox?${queryString}` : '/notifications/inbox'
    
    const response = await apiClient.get<NotificationsInboxResponse>(url)
    return response.data
  },

  // Mark notification as read
  markAsRead: async (notificationId: string): Promise<void> => {
    await apiClient.put(`/notifications/${notificationId}/read`, {})
  },

  // Mark all notifications as read
  markAllAsRead: async (): Promise<void> => {
    await apiClient.put('/notifications/read-all', {})
  },

  // Get unread count
  // PENTING: Endpoint ini sudah menggunakan RBAC di backend (GetUnreadCountWithRBAC)
  // - Superadmin/Administrator: melihat semua unread count
  // - Admin: melihat unread count dari company mereka + descendants
  // - Regular users: hanya melihat unread count mereka sendiri
  getUnreadCount: async (): Promise<number> => {
    const response = await apiClient.get<UnreadCountResponse>('/notifications/unread-count')
    return response.data.count
  },

  // Delete all notifications
  deleteAll: async (): Promise<void> => {
    await apiClient.delete('/notifications/delete-all')
  },
}

