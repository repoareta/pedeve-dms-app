import apiClient from './client'

// Types
export interface Company {
  id: string
  name: string
  code: string
  description?: string
  parent_id?: string
  level: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface User {
  id: string
  username: string
  email: string
  role: string
  role_id?: string
  company_id?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  description?: string
  level: number
  is_system: boolean
  created_at: string
  updated_at: string
}

export interface Permission {
  id: string
  name: string
  description?: string
  resource: string
  action: string
  scope: 'global' | 'company' | 'sub_company'
  created_at: string
  updated_at: string
}

// Company API
export const companyApi = {
  getAll: async (): Promise<Company[]> => {
    const response = await apiClient.get<Company[]>('/companies')
    return response.data
  },

  getById: async (id: string): Promise<Company> => {
    const response = await apiClient.get<Company>(`/companies/${id}`)
    return response.data
  },

  getChildren: async (id: string): Promise<Company[]> => {
    const response = await apiClient.get<Company[]>(`/companies/${id}/children`)
    return response.data
  },

  create: async (data: {
    name: string
    code: string
    description?: string
    parent_id?: string
  }): Promise<Company> => {
    const response = await apiClient.post<Company>('/companies', data)
    return response.data
  },

  update: async (id: string, data: {
    name: string
    description?: string
  }): Promise<Company> => {
    const response = await apiClient.put<Company>(`/companies/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/companies/${id}`)
  },
}

// User API
export const userApi = {
  getAll: async (): Promise<User[]> => {
    const response = await apiClient.get<User[]>('/users')
    return response.data
  },

  getById: async (id: string): Promise<User> => {
    const response = await apiClient.get<User>(`/users/${id}`)
    return response.data
  },

  create: async (data: {
    username: string
    email: string
    password: string
    company_id?: string
    role_id?: string
  }): Promise<User> => {
    const response = await apiClient.post<User>('/users', data)
    return response.data
  },

  update: async (id: string, data: {
    username?: string
    email?: string
    company_id?: string
    role_id?: string
  }): Promise<User> => {
    const response = await apiClient.put<User>(`/users/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/users/${id}`)
  },

  toggleStatus: async (id: string): Promise<User> => {
    const response = await apiClient.patch<User>(`/users/${id}/toggle-status`)
    return response.data
  },

  resetPassword: async (id: string, newPassword: string): Promise<{ message: string; user_id: string }> => {
    const response = await apiClient.post<{ message: string; user_id: string }>(`/users/${id}/reset-password`, {
      new_password: newPassword,
    })
    return response.data
  },
}

// Role API
export const roleApi = {
  getAll: async (): Promise<Role[]> => {
    const response = await apiClient.get<Role[]>('/roles')
    return response.data
  },

  getById: async (id: string): Promise<Role> => {
    const response = await apiClient.get<Role>(`/roles/${id}`)
    return response.data
  },

  getPermissions: async (id: string): Promise<Permission[]> => {
    const response = await apiClient.get<Permission[]>(`/roles/${id}/permissions`)
    return response.data
  },

  create: async (data: {
    name: string
    description?: string
    level: number
  }): Promise<Role> => {
    const response = await apiClient.post<Role>('/roles', data)
    return response.data
  },

  update: async (id: string, data: {
    name?: string
    description?: string
    level?: number
  }): Promise<Role> => {
    const response = await apiClient.put<Role>(`/roles/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/roles/${id}`)
  },

  assignPermission: async (roleId: string, permissionId: string): Promise<void> => {
    await apiClient.post(`/roles/${roleId}/permissions`, {
      permission_id: permissionId,
    })
  },

  revokePermission: async (roleId: string, permissionId: string): Promise<void> => {
    await apiClient.delete(`/roles/${roleId}/permissions`, {
      data: { permission_id: permissionId },
    })
  },
}

// Permission API
export const permissionApi = {
  getAll: async (params?: {
    resource?: string
    scope?: string
  }): Promise<Permission[]> => {
    const response = await apiClient.get<Permission[]>('/permissions', { params })
    return response.data
  },

  getById: async (id: string): Promise<Permission> => {
    const response = await apiClient.get<Permission>(`/permissions/${id}`)
    return response.data
  },

  create: async (data: {
    name: string
    description?: string
    resource: string
    action: string
    scope: 'global' | 'company' | 'sub_company'
  }): Promise<Permission> => {
    const response = await apiClient.post<Permission>('/permissions', data)
    return response.data
  },

  update: async (id: string, data: {
    name?: string
    description?: string
  }): Promise<Permission> => {
    const response = await apiClient.put<Permission>(`/permissions/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/permissions/${id}`)
  },
}

