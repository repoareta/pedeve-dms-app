import apiClient from './client'
import axios from 'axios'
import { getCSRFToken } from './client'

// Types
export interface Company {
  id: string
  name: string
  short_name?: string
  code: string
  description?: string
  npwp?: string
  nib?: string
  status?: string
  logo?: string
  phone?: string
  fax?: string
  email?: string
  website?: string
  address?: string
  operational_address?: string
  authorized_capital?: number
  paid_up_capital?: number
  currency?: string // IDR (Rupiah) atau USD (Dollar), default: IDR
  parent_id?: string
  level: number
  is_active: boolean
  main_parent_company?: string
  shareholders?: Shareholder[]
  business_fields?: BusinessField[] // Array dari backend
  main_business?: BusinessField // Singular untuk kompatibilitas
  directors?: Director[]
  created_at: string
  updated_at: string
}

export interface UserCompanyResponse {
  company: Company
  role_id?: string
  role: string
  role_level: number // 0=superadmin, 1=admin, 2=manager, 3=staff
}

export interface Shareholder {
  id?: string
  shareholder_company_id?: string | null // ID perusahaan pemegang saham (nullable: jika null berarti individu/eksternal)
  type: string // Backend stores as string (comma-separated), frontend uses array
  name: string
  identity_number: string
  ownership_percent: number // 10 digit desimal
  share_sheet_count?: number
  share_value_per_sheet?: number
  is_main_parent?: boolean
  shareholder_company?: Company // Perusahaan pemegang saham (jika shareholder_company_id tidak null)
}

export interface ShareholderType {
  id: string
  name: string
  is_active: boolean
  usage_count: number
  created_by: string
  created_at?: string
  updated_at?: string
}

export interface DirectorPosition {
  id: string
  name: string
  is_active: boolean
  usage_count: number
  created_by: string
  created_at?: string
  updated_at?: string
}

export interface BusinessField {
  id?: string
  industry_sector: string
  kbli: string
  main_business_activity: string
  additional_activities?: string
  start_operation_date?: string
}

export interface Director {
  id?: string
  position: string // Comma-separated string from backend, array in frontend
  full_name: string
  ktp: string
  npwp: string
  start_date?: string
  domicile_address: string
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

export interface CompanyCreateRequest {
  name: string
  short_name?: string
  code: string
  description?: string
  npwp?: string
  nib?: string
  status?: string
  logo?: string
  phone?: string
  fax?: string
  email?: string
  website?: string
  address?: string
  operational_address?: string
  authorized_capital?: number
  paid_up_capital?: number
  currency?: string // IDR (Rupiah) atau USD (Dollar), default: IDR
  parent_id?: string
  main_parent_company?: string
  shareholders?: Shareholder[]
  main_business?: BusinessField
  directors?: Director[]
}

export interface CompanyUpdateRequest {
  name?: string
  short_name?: string
  description?: string
  npwp?: string
  nib?: string
  status?: string
  logo?: string
  phone?: string
  fax?: string
  email?: string
  website?: string
  address?: string
  operational_address?: string
  authorized_capital?: number
  paid_up_capital?: number
  currency?: string // IDR (Rupiah) atau USD (Dollar), default: IDR
  parent_id?: string
  main_parent_company?: string
  shareholders?: Shareholder[]
  main_business?: BusinessField
  directors?: Director[]
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

  getAncestors: async (id: string): Promise<Company[]> => {
    const response = await apiClient.get<Company[]>(`/companies/${id}/ancestors`)
    return response.data
  },

  create: async (data: CompanyCreateRequest): Promise<Company> => {
    const response = await apiClient.post<Company>('/companies', data)
    return response.data
  },

  update: async (id: string, data: CompanyUpdateRequest): Promise<Company> => {
    const response = await apiClient.put<Company>(`/companies/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/companies/${id}`)
  },

  getUsers: async (id: string): Promise<User[]> => {
    const response = await apiClient.get<User[]>(`/companies/${id}/users`)
    return response.data
  },
}

// Shareholder Types API
const shareholderTypesApi = {
  async getShareholderTypes(includeInactive = false): Promise<ShareholderType[]> {
    const res = await apiClient.get<ShareholderType[]>('/shareholder-types', {
      params: { include_inactive: includeInactive },
    })
    return res.data
  },

  async createShareholderType(name: string): Promise<ShareholderType> {
    const res = await apiClient.post<ShareholderType>('/shareholder-types', { name })
    return res.data
  },

  async updateShareholderType(id: string, payload: { name?: string; is_active?: boolean }): Promise<ShareholderType> {
    const res = await apiClient.put<ShareholderType>(`/shareholder-types/${id}`, payload)
    return res.data
  },

  async deleteShareholderType(id: string): Promise<void> {
    await apiClient.delete(`/shareholder-types/${id}`)
  },
}

export { shareholderTypesApi }

const directorPositionsApi = {
  async getDirectorPositions(includeInactive = false): Promise<DirectorPosition[]> {
    const res = await apiClient.get<DirectorPosition[]>('/director-positions', {
      params: { include_inactive: includeInactive },
    })
    return res.data
  },

  async createDirectorPosition(name: string): Promise<DirectorPosition> {
    const res = await apiClient.post<DirectorPosition>('/director-positions', { name })
    return res.data
  },

  async updateDirectorPosition(id: string, payload: { name?: string; is_active?: boolean }): Promise<DirectorPosition> {
    const res = await apiClient.put<DirectorPosition>(`/director-positions/${id}`, payload)
    return res.data
  },

  async deleteDirectorPosition(id: string): Promise<void> {
    await apiClient.delete(`/director-positions/${id}`)
  },
}

export { directorPositionsApi }
export default companyApi

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
    company_id?: string | null
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

  assignToCompany: async (id: string, companyId: string, roleId?: string): Promise<User> => {
    const response = await apiClient.post<User>(`/users/${id}/assign-company`, {
      company_id: companyId,
      role_id: roleId,
    })
    return response.data
  },

  unassignFromCompany: async (id: string, companyId: string): Promise<User> => {
    const response = await apiClient.post<User>(`/users/${id}/unassign-company`, {
      company_id: companyId,
    })
    return response.data
  },

  getMyCompanies: async (): Promise<UserCompanyResponse[]> => {
    const response = await apiClient.get<UserCompanyResponse[]>('/users/me/companies')
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

// Upload API
export interface UploadResponse {
  url: string
  filename: string
  size: number
}

export const uploadApi = {
  uploadLogo: async (file: File): Promise<UploadResponse> => {
    const formData = new FormData()
    formData.append('file', file)

    // Ambil CSRF token
    const csrfToken = await getCSRFToken()
    
    // Gunakan axios langsung untuk multipart/form-data
    const API_BASE_URL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
    const baseURL = API_BASE_URL.endsWith('/api/v1') 
      ? API_BASE_URL 
      : `${API_BASE_URL}/api/v1`

    const headers: Record<string, string> = {}
    
    // Tambahkan CSRF token jika tersedia
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
    
    // Jangan set Content-Type secara manual untuk multipart/form-data
    // Browser akan otomatis set dengan boundary yang benar

    const response = await axios.post<UploadResponse>(
      `${baseURL}/upload/logo`,
      formData,
      {
        withCredentials: true,
        headers,
      }
    )
    return response.data
  },
}
