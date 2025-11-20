import axios from 'axios'

// Ensure baseURL always ends with /api/v1
const getBaseURL = () => {
  const envURL = import.meta.env.VITE_API_URL
  if (envURL) {
    // Remove trailing slash if exists
    const cleanURL = envURL.replace(/\/$/, '')
    // Ensure /api/v1 is appended
    return cleanURL.endsWith('/api/v1') ? cleanURL : `${cleanURL}/api/v1`
  }
  return 'http://localhost:8080/api/v1'
}

const API_BASE_URL = getBaseURL()

// Debug log (remove in production)
if (import.meta.env.DEV) {
  console.log('[API Client] Base URL:', API_BASE_URL)
}

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add JWT token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Check if this is a login/register endpoint - don't redirect in that case
      const url = error.config?.url || ''
      const isAuthEndpoint = url.includes('/auth/login') || url.includes('/auth/register')
      
      if (!isAuthEndpoint) {
        // Unauthorized - clear token and redirect to login (only for protected endpoints)
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_user')
        // Only redirect if not already on login page
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
      }
      // For auth endpoints, let the error pass through so LoginView can handle it
    }
    return Promise.reject(error)
  }
)

export default apiClient

