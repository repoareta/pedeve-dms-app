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

// CSRF token storage
let csrfToken: string | null = null

// Function to get CSRF token from backend
export const getCSRFToken = async (): Promise<string | null> => {
  try {
    const response = await axios.get<{ csrf_token: string }>(`${API_BASE_URL}/csrf-token`)
    csrfToken = response.data.csrf_token
    return csrfToken
  } catch (error) {
    console.error('Failed to get CSRF token:', error)
    return null
  }
}

// Initialize CSRF token on module load (optional)
// getCSRFToken()

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Enable cookies
})

// Request interceptor to add JWT token and CSRF token
apiClient.interceptors.request.use(
  async (config) => {
    // Add JWT token
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Add CSRF token for state-changing methods (POST, PUT, DELETE, PATCH)
    const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH']
    if (config.method && stateChangingMethods.includes(config.method.toUpperCase())) {
      // Get CSRF token if not available
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

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    // Handle CSRF token errors
    if (error.response?.status === 403) {
      const errorCode = error.response?.data?.error
      if (errorCode === 'csrf_token_missing' || errorCode === 'csrf_token_invalid') {
        // Get new CSRF token and retry request
        const newToken = await getCSRFToken()
        if (newToken && error.config) {
          error.config.headers['X-CSRF-Token'] = newToken
          return apiClient.request(error.config)
        }
      }
    }

    if (error.response?.status === 401) {
      // Check if this is a login/register endpoint - don't redirect in that case
      const url = error.config?.url || ''
      const isAuthEndpoint = url.includes('/auth/login') || url.includes('/auth/register')
      
      if (!isAuthEndpoint) {
        // Unauthorized - clear token and redirect to login (only for protected endpoints)
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_user')
        csrfToken = null // Clear CSRF token
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

