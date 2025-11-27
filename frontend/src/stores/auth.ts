import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type User } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  // Token sekarang di httpOnly cookie, jadi tidak disimpan di localStorage
  // Pertahankan token ref untuk kompatibilitas ke belakang (tapi tidak digunakan untuk API calls)
  const token = ref<string | null>(null)
  
  // Initialize user from localStorage
  const getInitialUser = (): User | null => {
    const stored = localStorage.getItem('auth_user')
    return stored ? JSON.parse(stored) : null
  }
  const user = ref<User | null>(getInitialUser())
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isLoggingOut = ref(false) // Flag untuk mencegah validasi saat logout

  // Cek status autentikasi - token ada di httpOnly cookie, jadi cek user saja
  const isAuthenticated = computed(() => !!user.value)

  // Login (bisa gunakan username atau email)
  const login = async (usernameOrEmail: string, password: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login({ username: usernameOrEmail, password })
      
      // Jika 2FA diperlukan, jangan simpan token/user - return response tanpa menyimpan
      if (response.requires_2fa) {
        // Hapus token/user yang ada ketika 2FA diperlukan
        token.value = null
        user.value = null
        localStorage.removeItem('auth_user')
        return response
      }
      
      // Hanya simpan info user untuk UI state (token sekarang di httpOnly cookie)
      if (response.token && response.user) {
        token.value = response.token // Simpan untuk UI state, tapi tidak digunakan untuk API calls
        user.value = response.user
        // Jangan simpan token di localStorage lagi (sudah di httpOnly cookie)
        localStorage.setItem('auth_user', JSON.stringify(response.user))
      }
      
      return response
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Login failed'
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
      // Token disimpan di httpOnly cookie oleh backend, frontend hanya simpan info user
      token.value = response.token // Simpan untuk UI state, tapi tidak digunakan untuk API calls
      user.value = response.user
      // Jangan simpan token di localStorage lagi (sudah di httpOnly cookie)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      return response
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Logout
  const logout = async () => {
    // Set flag untuk mencegah validasi token saat logout
    isLoggingOut.value = true
    try {
      // Hanya panggil logout endpoint jika user terautentikasi (punya cookie)
      // Cek apakah user ada sebelum memanggil API
      if (user.value) {
        await authApi.logout()
      }
    } catch (error) {
      // Lanjutkan dengan pembersihan lokal meskipun API call gagal
      // Jangan log error jika status 401 (user sudah logout)
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { status?: number } }
        const status = axiosError.response?.status
        if (status !== 401) {
          console.error('Logout API error:', error)
        }
      }
    } finally {
      // Hapus state lokal (token ada di cookie, jadi hapus localStorage)
      token.value = null
      user.value = null
      localStorage.removeItem('auth_user')
      // Reset flag setelah logout selesai
      isLoggingOut.value = false
    }
  }

  // Get profile
  const fetchProfile = async () => {
    // Jangan fetch profile jika sedang logout
    if (isLoggingOut.value) {
      throw new Error('Logout in progress')
    }
    
    loading.value = true
    error.value = null
    try {
      const profile = await authApi.getProfile()
      user.value = profile
      localStorage.setItem('auth_user', JSON.stringify(profile))
      return profile
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } }; code?: string; message?: string }
      error.value = axiosError.response?.data?.message || 'Failed to fetch profile'
      
      // Handle connection errors (server tidak tersedia)
      const isConnectionError = axiosError.code === 'ERR_NETWORK' || 
                                axiosError.code === 'ERR_CONNECTION_REFUSED' || 
                                axiosError.message?.includes('Network Error')
      
      if (isConnectionError) {
        // Server tidak tersedia - clear state dan throw error khusus
        clearAuthState()
        throw new Error('SERVER_UNAVAILABLE')
      }
      
      throw err
    } finally {
      loading.value = false
    }
  }

  // Clear auth state (helper function)
  const clearAuthState = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('auth_user')
  }

  // Login with 2FA code
  const loginWith2FA = async (usernameOrEmail: string, password: string, code: string) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.loginWith2FA({ username: usernameOrEmail, password, code })
      // Token disimpan di httpOnly cookie oleh backend, frontend hanya simpan info user
      token.value = response.token // Simpan untuk UI state, tapi tidak digunakan untuk API calls
      user.value = response.user
      // Jangan simpan token di localStorage lagi (sudah di httpOnly cookie)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      return response
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || '2FA verification failed'
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
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Failed to generate 2FA secret'
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
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Failed to verify 2FA code'
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
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Failed to get 2FA status'
      throw err
    } finally {
      loading.value = false
    }
  }

  const disable2FA = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.disable2FA()
      return response
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } }
      error.value = axiosError.response?.data?.message || 'Failed to disable 2FA'
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
    clearAuthState, // Export helper function
    generate2FA,
    verify2FA,
    get2FAStatus,
    disable2FA,
  }
})

