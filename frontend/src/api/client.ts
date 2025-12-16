import axios from 'axios'
import { useAuthStore } from '../stores/auth'
import { logger } from '../utils/logger'

// Pastikan baseURL selalu diakhiri dengan /api/v1
const getBaseURL = () => {
  const envURL = import.meta.env.VITE_API_BASE_URL || import.meta.env.VITE_API_URL
  const fallbackHost = import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com'

  const base = (envURL || fallbackHost).replace(/\/$/, '')
  return base.endsWith('/api/v1') ? base : `${base}/api/v1`
}

const API_BASE_URL = getBaseURL()

// Debug log (hanya muncul di development)
logger.api('Base URL:', API_BASE_URL)

// Penyimpanan token CSRF
let csrfToken: string | null = null

// Fungsi untuk mengambil token CSRF dari backend
export const getCSRFToken = async (): Promise<string | null> => {
  try {
    const response = await axios.get<{ csrf_token: string }>(`${API_BASE_URL}/csrf-token`)
    csrfToken = response.data.csrf_token
    return csrfToken
  } catch (error: unknown) {
    // Handle connection errors (server tidak tersedia)
    const axiosError = error as { code?: string; message?: string }
    if (axiosError.code === 'ERR_NETWORK' || axiosError.code === 'ERR_CONNECTION_REFUSED' || axiosError.message?.includes('Network Error')) {
      // Server tidak tersedia - jangan log error, hanya return null
      // Token CSRF akan di-fetch lagi saat diperlukan
      csrfToken = null
      return null
    }
    // Error lainnya - log untuk debugging
    logger.error('Failed to get CSRF token:', error)
    csrfToken = null
    return null
  }
}

// Inisialisasi token CSRF saat module load (opsional)
// getCSRFToken()

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Aktifkan cookies
})

// Request interceptor untuk menambahkan token JWT dan token CSRF
apiClient.interceptors.request.use(
  async (config) => {
    // Tambahkan Authorization header jika token tersedia di store/localStorage (fallback jika cookie tidak terkirim)
    const authStore = useAuthStore()
    const bearerToken = authStore?.token || localStorage.getItem('auth_token')
    if (bearerToken) {
      config.headers.Authorization = `Bearer ${bearerToken}`
    }
    
    // Tambahkan token CSRF untuk method yang mengubah state (POST, PUT, DELETE, PATCH)
    const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH']
    if (config.method && stateChangingMethods.includes(config.method.toUpperCase())) {
      // Ambil token CSRF jika belum tersedia
      if (!csrfToken) {
        await getCSRFToken()
      }
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor untuk menangani error
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    // Tangani error token CSRF
    if (error.response?.status === 403) {
      const errorCode = error.response?.data?.error
      if (errorCode === 'csrf_token_missing' || errorCode === 'csrf_token_invalid') {
        // Ambil token CSRF baru dan coba request lagi
        const newToken = await getCSRFToken()
        if (newToken && error.config) {
          error.config.headers['X-CSRF-Token'] = newToken
          return apiClient.request(error.config)
        }
      }
    }

    if (error.response?.status === 401) {
      // Cek apakah ini endpoint login/register/logout - jangan redirect dalam kasus ini
      const url = error.config?.url || ''
      const isAuthEndpoint = url.includes('/auth/login') || 
                             url.includes('/auth/register') ||
                             url.includes('/auth/logout')
      
      // Cek apakah kita di halaman guest (login/register)
      const isGuestPage = window.location.pathname === '/login' || 
                          window.location.pathname === '/register'
      
      if (!isAuthEndpoint && !isGuestPage) {
        // Unauthorized di halaman yang dilindungi - hapus state lokal dan redirect
        localStorage.removeItem('auth_user')
        csrfToken = null // Hapus token CSRF
        window.location.href = '/login'
      } else if (isGuestPage && url.includes('/auth/profile')) {
        // Di halaman guest dan pengecekan profile gagal - hapus state secara diam-diam
        // Jangan redirect atau panggil logout (user sudah di halaman login)
        localStorage.removeItem('auth_user')
        csrfToken = null
      }
      // Untuk endpoint auth atau halaman guest, biarkan error diteruskan
    }
    return Promise.reject(error)
  }
)

export default apiClient
