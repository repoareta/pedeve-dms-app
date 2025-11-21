import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type User } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const user = ref<User | null>(() => {
    const stored = localStorage.getItem('auth_user')
    return stored ? JSON.parse(stored) : null
  })
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // Login (can use username or email)
  const login = async (usernameOrEmail: string, password: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login({ username: usernameOrEmail, password })
      token.value = response.token
      user.value = response.user
      localStorage.setItem('auth_token', response.token)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Register
  const register = async (username: string, email: string, password: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register({ username, email, password })
      token.value = response.token
      user.value = response.user
      localStorage.setItem('auth_token', response.token)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Logout
  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
  }

  // Get profile
  const fetchProfile = async () => {
    loading.value = true
    error.value = null
    try {
      const profile = await authApi.getProfile()
      user.value = profile
      localStorage.setItem('auth_user', JSON.stringify(profile))
      return profile
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to fetch profile'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Login with 2FA code
  const loginWith2FA = async (usernameOrEmail: string, password: string, code: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.loginWith2FA({ username: usernameOrEmail, password, code })
      token.value = response.token
      user.value = response.user
      localStorage.setItem('auth_token', response.token)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || '2FA verification failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  // 2FA methods
  const generate2FA = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.generate2FA()
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to generate 2FA secret'
      throw err
    } finally {
      loading.value = false
    }
  }

  const verify2FA = async (code: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.verify2FA(code)
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to verify 2FA code'
      throw err
    } finally {
      loading.value = false
    }
  }

  const get2FAStatus = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.get2FAStatus()
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || 'Failed to get 2FA status'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    login,
    loginWith2FA,
    register,
    logout,
    fetchProfile,
    generate2FA,
    verify2FA,
    get2FAStatus,
  }
})

