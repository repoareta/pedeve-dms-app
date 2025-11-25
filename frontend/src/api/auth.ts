import apiClient from './client'

export interface LoginRequest {
  username: string // Bisa username atau email
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface User {
  id: string
  username: string
  email: string
  role: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
  requires_2fa?: boolean
  message?: string
}

export interface TwoFAResponse {
  secret: string
  qr_code: string
  url: string
  message: string
}

export interface TwoFAVerifyResponse {
  message: string
  backup_codes?: string[]
}

export interface TwoFAStatus {
  enabled: boolean
}

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  loginWith2FA: async (data: LoginRequest & { code: string }): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  getProfile: async (): Promise<User> => {
    const response = await apiClient.get<User>('/auth/profile')
    return response.data
  },

  logout: async (): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/auth/logout', {})
    return response.data
  },

  // 2FA endpoints
  generate2FA: async (): Promise<TwoFAResponse> => {
    const response = await apiClient.post<TwoFAResponse>('/auth/2fa/generate', {})
    return response.data
  },

  verify2FA: async (code: string): Promise<TwoFAVerifyResponse> => {
    const response = await apiClient.post<TwoFAVerifyResponse>('/auth/2fa/verify', { code })
    return response.data
  },

  get2FAStatus: async (): Promise<TwoFAStatus> => {
    const response = await apiClient.get<TwoFAStatus>('/auth/2fa/status')
    return response.data
  },

  disable2FA: async (): Promise<TwoFAVerifyResponse> => {
    const response = await apiClient.post<TwoFAVerifyResponse>('/auth/2fa/disable', {})
    return response.data
  },

  // Profile management
  updateEmail: async (email: string): Promise<User> => {
    const response = await apiClient.put<User>('/auth/profile/email', { email })
    return response.data
  },

  changePassword: async (oldPassword: string, newPassword: string): Promise<{ message: string }> => {
    const response = await apiClient.put<{ message: string }>('/auth/profile/password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
    return response.data
  },
}

